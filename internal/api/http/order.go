package http

import (
	"net/http"

	"github.com/RuslanHabibullin/KorpISCourseWork/internal/domain"
	ordersvc "github.com/RuslanHabibullin/KorpISCourseWork/internal/service/order"

	"github.com/google/uuid"
)

// --- Orders ---

func (h *Handler) createOrder(w http.ResponseWriter, r *http.Request) {
	var req struct {
		VehicleID uuid.UUID `json:"vehicle_id"`
		ClientID  uuid.UUID `json:"client_id"`
		Complaint string    `json:"complaint"`
	}
	if err := decode(r, &req); err != nil {
		respond(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	o, err := h.orders.CreateOrder(r.Context(), ordersvc.CreateOrderInput{
		VehicleID: req.VehicleID,
		ClientID:  req.ClientID,
		Complaint: req.Complaint,
	})
	if err != nil {
		respondErr(w, err)
		return
	}
	respond(w, http.StatusCreated, o)
}

func (h *Handler) listOrders(w http.ResponseWriter, r *http.Request) {
	limit, offset := parsePagination(r)
	list, err := h.orders.ListOrders(r.Context(), limit, offset)
	if err != nil {
		respondErr(w, err)
		return
	}
	respond(w, http.StatusOK, list)
}

func (h *Handler) getOrder(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(r, "id")
	if err != nil {
		respond(w, http.StatusBadRequest, map[string]string{"error": "invalid uuid"})
		return
	}
	o, err := h.orders.GetOrder(r.Context(), id)
	if err != nil {
		respondErr(w, err)
		return
	}
	respond(w, http.StatusOK, o)
}

// transitionOrder — FSM переход статуса
// POST /orders/{id}/transition  body: {"status": "approved"}
func (h *Handler) transitionOrder(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(r, "id")
	if err != nil {
		respond(w, http.StatusBadRequest, map[string]string{"error": "invalid uuid"})
		return
	}
	var req struct {
		Status domain.OrderStatus `json:"status"`
	}
	if err := decode(r, &req); err != nil {
		respond(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	o, err := h.orders.TransitionStatus(r.Context(), id, req.Status)
	if err != nil {
		respondErr(w, err)
		return
	}
	respond(w, http.StatusOK, o)
}

func (h *Handler) deleteOrder(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(r, "id")
	if err != nil {
		respond(w, http.StatusBadRequest, map[string]string{"error": "invalid uuid"})
		return
	}
	o, err := h.orders.GetOrder(r.Context(), id)
	if err != nil {
		respondErr(w, err)
		return
	}
	if o.Status != domain.OrderStatusDraft {
		respond(w, http.StatusConflict, map[string]string{"error": "only draft orders can be deleted"})
		return
	}
	respond(w, http.StatusNoContent, nil)
}

// --- Order services ---

func (h *Handler) addService(w http.ResponseWriter, r *http.Request) {
	orderID, err := parseUUID(r, "id")
	if err != nil {
		respond(w, http.StatusBadRequest, map[string]string{"error": "invalid uuid"})
		return
	}
	var req struct {
		ServiceID uuid.UUID `json:"service_id"`
		Quantity  int       `json:"quantity"`
	}
	if err := decode(r, &req); err != nil {
		respond(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	ws, err := h.orders.AddService(r.Context(), ordersvc.AddServiceInput{
		OrderID:   orderID,
		ServiceID: req.ServiceID,
		Quantity:  req.Quantity,
	})
	if err != nil {
		respondErr(w, err)
		return
	}
	respond(w, http.StatusCreated, ws)
}

func (h *Handler) listOrderServices(w http.ResponseWriter, r *http.Request) {
	orderID, err := parseUUID(r, "id")
	if err != nil {
		respond(w, http.StatusBadRequest, map[string]string{"error": "invalid uuid"})
		return
	}
	list, err := h.orders.ListServices(r.Context(), orderID)
	if err != nil {
		respondErr(w, err)
		return
	}
	respond(w, http.StatusOK, list)
}

func (h *Handler) removeService(w http.ResponseWriter, r *http.Request) {
	orderID, _ := parseUUID(r, "id")
	lineID, err := parseUUID(r, "lineId")
	if err != nil {
		respond(w, http.StatusBadRequest, map[string]string{"error": "invalid uuid"})
		return
	}
	if err := h.orders.RemoveService(r.Context(), orderID, lineID); err != nil {
		respondErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Order parts ---

func (h *Handler) addPart(w http.ResponseWriter, r *http.Request) {
	orderID, err := parseUUID(r, "id")
	if err != nil {
		respond(w, http.StatusBadRequest, map[string]string{"error": "invalid uuid"})
		return
	}
	var req struct {
		PartID   uuid.UUID `json:"part_id"`
		Quantity int       `json:"quantity"`
	}
	if err := decode(r, &req); err != nil {
		respond(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	wp, err := h.orders.AddPart(r.Context(), ordersvc.AddPartInput{
		OrderID:  orderID,
		PartID:   req.PartID,
		Quantity: req.Quantity,
	})
	if err != nil {
		respondErr(w, err)
		return
	}
	respond(w, http.StatusCreated, wp)
}

func (h *Handler) listOrderParts(w http.ResponseWriter, r *http.Request) {
	orderID, err := parseUUID(r, "id")
	if err != nil {
		respond(w, http.StatusBadRequest, map[string]string{"error": "invalid uuid"})
		return
	}
	list, err := h.orders.ListParts(r.Context(), orderID)
	if err != nil {
		respondErr(w, err)
		return
	}
	respond(w, http.StatusOK, list)
}

func (h *Handler) removePart(w http.ResponseWriter, r *http.Request) {
	orderID, _ := parseUUID(r, "id")
	lineID, err := parseUUID(r, "lineId")
	if err != nil {
		respond(w, http.StatusBadRequest, map[string]string{"error": "invalid uuid"})
		return
	}
	// Получаем строку для знания PartID и Quantity (для возврата на склад)
	parts, err := h.orders.ListParts(r.Context(), orderID)
	if err != nil {
		respondErr(w, err)
		return
	}
	var target *domain.WorkOrderPart
	for _, p := range parts {
		if p.ID == lineID {
			target = p
			break
		}
	}
	if target == nil {
		respond(w, http.StatusNotFound, map[string]string{"error": "part line not found"})
		return
	}
	if err := h.orders.RemovePart(r.Context(), orderID, target); err != nil {
		respondErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Payments ---

func (h *Handler) addPayment(w http.ResponseWriter, r *http.Request) {
	orderID, err := parseUUID(r, "id")
	if err != nil {
		respond(w, http.StatusBadRequest, map[string]string{"error": "invalid uuid"})
		return
	}
	var req struct {
		Amount float64 `json:"amount"`
	}
	if err := decode(r, &req); err != nil {
		respond(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	p, err := h.orders.AddPayment(r.Context(), ordersvc.AddPaymentInput{
		OrderID: orderID,
		Amount:  req.Amount,
	})
	if err != nil {
		respondErr(w, err)
		return
	}
	respond(w, http.StatusCreated, p)
}

func (h *Handler) listPayments(w http.ResponseWriter, r *http.Request) {
	orderID, err := parseUUID(r, "id")
	if err != nil {
		respond(w, http.StatusBadRequest, map[string]string{"error": "invalid uuid"})
		return
	}
	list, err := h.orders.ListPayments(r.Context(), orderID)
	if err != nil {
		respondErr(w, err)
		return
	}
	respond(w, http.StatusOK, list)
}

// --- Service catalog ---

func (h *Handler) createServiceCatalog(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name      string  `json:"name"`
		BasePrice float64 `json:"base_price"`
	}
	if err := decode(r, &req); err != nil {
		respond(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	svc, err := h.orders.CreateServiceCatalog(r.Context(), req.Name, req.BasePrice)
	if err != nil {
		respondErr(w, err)
		return
	}
	respond(w, http.StatusCreated, svc)
}

func (h *Handler) listServiceCatalog(w http.ResponseWriter, r *http.Request) {
	list, err := h.orders.ListServiceCatalog(r.Context())
	if err != nil {
		respondErr(w, err)
		return
	}
	respond(w, http.StatusOK, list)
}

// --- Parts catalog ---

func (h *Handler) createPart(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name  string  `json:"name"`
		Price float64 `json:"price"`
	}
	if err := decode(r, &req); err != nil {
		respond(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	p, err := h.stock.CreatePart(r.Context(), req.Name, req.Price)
	if err != nil {
		respondErr(w, err)
		return
	}
	respond(w, http.StatusCreated, p)
}

func (h *Handler) listParts(w http.ResponseWriter, r *http.Request) {
	list, err := h.stock.ListParts(r.Context())
	if err != nil {
		respondErr(w, err)
		return
	}
	respond(w, http.StatusOK, list)
}

func (h *Handler) getPart(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(r, "id")
	if err != nil {
		respond(w, http.StatusBadRequest, map[string]string{"error": "invalid uuid"})
		return
	}
	p, err := h.stock.GetPart(r.Context(), id)
	if err != nil {
		respondErr(w, err)
		return
	}
	respond(w, http.StatusOK, p)
}

// --- Stock ---

func (h *Handler) listStock(w http.ResponseWriter, r *http.Request) {
	list, err := h.stock.ListStock(r.Context())
	if err != nil {
		respondErr(w, err)
		return
	}
	respond(w, http.StatusOK, list)
}

func (h *Handler) getStock(w http.ResponseWriter, r *http.Request) {
	partID, err := parseUUID(r, "partId")
	if err != nil {
		respond(w, http.StatusBadRequest, map[string]string{"error": "invalid uuid"})
		return
	}
	s, err := h.stock.GetStock(r.Context(), partID)
	if err != nil {
		respondErr(w, err)
		return
	}
	respond(w, http.StatusOK, s)
}

func (h *Handler) replenishStock(w http.ResponseWriter, r *http.Request) {
	partID, err := parseUUID(r, "partId")
	if err != nil {
		respond(w, http.StatusBadRequest, map[string]string{"error": "invalid uuid"})
		return
	}
	var req struct {
		Qty int `json:"qty"`
	}
	if err := decode(r, &req); err != nil {
		respond(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	if err := h.stock.Replenish(r.Context(), partID, req.Qty); err != nil {
		respondErr(w, err)
		return
	}
	respond(w, http.StatusOK, map[string]string{"status": "ok"})
}
