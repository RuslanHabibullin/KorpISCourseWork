package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/RuslanHabibullin/KorpISCourseWork/internal/domain"

	clientsvc "github.com/RuslanHabibullin/KorpISCourseWork/internal/service/client"

	ordersvc "github.com/RuslanHabibullin/KorpISCourseWork/internal/service/order"

	stocksvc "github.com/RuslanHabibullin/KorpISCourseWork/internal/service/stock"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// Handler — корневой HTTP-обработчик
type Handler struct {
	log     *zap.Logger
	clients *clientsvc.Service
	orders  *ordersvc.Service
	stock   *stocksvc.Service
}

func NewHandler(
	log *zap.Logger,
	clients *clientsvc.Service,
	orders *ordersvc.Service,
	stock *stocksvc.Service,
) *Handler {
	return &Handler{
		log:     log,
		clients: clients,
		orders:  orders,
		stock:   stock,
	}
}

// Routes регистрирует все маршруты
func (h *Handler) Routes() http.Handler {
	r := chi.NewRouter()

	r.Route("/api/v1", func(r chi.Router) {
		// Clients
		r.Post("/clients", h.createClient)
		r.Get("/clients", h.listClients)
		r.Get("/clients/{id}", h.getClient)
		r.Put("/clients/{id}", h.updateClient)
		r.Delete("/clients/{id}", h.deleteClient)

		// Vehicles
		r.Post("/clients/{id}/vehicles", h.createVehicle)
		r.Get("/clients/{id}/vehicles", h.listVehicles)
		r.Get("/vehicles/{id}", h.getVehicle)
		r.Put("/vehicles/{id}", h.updateVehicle)
		r.Delete("/vehicles/{id}", h.deleteVehicle)

		// Orders
		r.Post("/orders", h.createOrder)
		r.Get("/orders", h.listOrders)
		r.Get("/orders/{id}", h.getOrder)
		r.Post("/orders/{id}/transition", h.transitionOrder)
		r.Delete("/orders/{id}", h.deleteOrder)

		// Order services (работы)
		r.Post("/orders/{id}/services", h.addService)
		r.Get("/orders/{id}/services", h.listOrderServices)
		r.Delete("/orders/{id}/services/{lineId}", h.removeService)

		// Order parts (запчасти)
		r.Post("/orders/{id}/parts", h.addPart)
		r.Get("/orders/{id}/parts", h.listOrderParts)
		r.Delete("/orders/{id}/parts/{lineId}", h.removePart)

		// Payments
		r.Post("/orders/{id}/payments", h.addPayment)
		r.Get("/orders/{id}/payments", h.listPayments)

		// Service catalog
		r.Post("/services", h.createServiceCatalog)
		r.Get("/services", h.listServiceCatalog)

		// Parts catalog
		r.Post("/parts", h.createPart)
		r.Get("/parts", h.listParts)
		r.Get("/parts/{id}", h.getPart)

		// Stock
		r.Get("/stock", h.listStock)
		r.Get("/stock/{partId}", h.getStock)
		r.Post("/stock/{partId}/replenish", h.replenishStock)
	})

	return r
}

// --- Helpers ---

func respond(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func respondErr(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	msg := "internal server error"

	switch {
	case errors.Is(err, domain.ErrNotFound):
		status = http.StatusNotFound
		msg = err.Error()
	case errors.Is(err, domain.ErrNoStock):
		status = http.StatusConflict
		msg = err.Error()
	case errors.Is(err, domain.ErrInvalidTransition):
		status = http.StatusConflict
		msg = err.Error()
	case errors.Is(err, domain.ErrOrderClosed):
		status = http.StatusConflict
		msg = err.Error()
	case errors.Is(err, domain.ErrAlreadyExists):
		status = http.StatusConflict
		msg = err.Error()
	}

	respond(w, status, map[string]string{"error": msg})
}

func parseUUID(r *http.Request, param string) (uuid.UUID, error) {
	return uuid.Parse(chi.URLParam(r, param))
}

func parsePagination(r *http.Request) (limit, offset int) {
	limit, _ = strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ = strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 {
		limit = 20
	}
	return
}

func decode(r *http.Request, v any) error {
	return json.NewDecoder(r.Body).Decode(v)
}
