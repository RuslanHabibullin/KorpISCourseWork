package http

import (
	"net/http"

	"github.com/RuslanHabibullin/KorpISCourseWork/internal/domain"
	clientsvc "github.com/RuslanHabibullin/KorpISCourseWork/internal/service/client"

	"github.com/google/uuid"
)

// --- Clients ---

func (h *Handler) createClient(w http.ResponseWriter, r *http.Request) {
	var req struct {
		FullName string `json:"full_name"`
		Phone    string `json:"phone"`
	}
	if err := decode(r, &req); err != nil {
		respond(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	c, err := h.clients.CreateClient(r.Context(), clientsvc.CreateClientInput{
		FullName: req.FullName,
		Phone:    req.Phone,
	})
	if err != nil {
		respondErr(w, err)
		return
	}
	respond(w, http.StatusCreated, c)
}

func (h *Handler) listClients(w http.ResponseWriter, r *http.Request) {
	limit, offset := parsePagination(r)
	list, err := h.clients.ListClients(r.Context(), limit, offset)
	if err != nil {
		respondErr(w, err)
		return
	}
	respond(w, http.StatusOK, list)
}

func (h *Handler) getClient(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(r, "id")
	if err != nil {
		respond(w, http.StatusBadRequest, map[string]string{"error": "invalid uuid"})
		return
	}
	c, err := h.clients.GetClient(r.Context(), id)
	if err != nil {
		respondErr(w, err)
		return
	}
	respond(w, http.StatusOK, c)
}

func (h *Handler) updateClient(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(r, "id")
	if err != nil {
		respond(w, http.StatusBadRequest, map[string]string{"error": "invalid uuid"})
		return
	}
	var req struct {
		FullName string `json:"full_name"`
		Phone    string `json:"phone"`
	}
	if err := decode(r, &req); err != nil {
		respond(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	c, err := h.clients.UpdateClient(r.Context(), clientsvc.UpdateClientInput{
		ClientID: id,
		FullName: req.FullName,
		Phone:    req.Phone,
	})
	if err != nil {
		respondErr(w, err)
		return
	}
	respond(w, http.StatusOK, c)
}

func (h *Handler) deleteClient(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(r, "id")
	if err != nil {
		respond(w, http.StatusBadRequest, map[string]string{"error": "invalid uuid"})
		return
	}
	if err := h.clients.DeleteClient(r.Context(), id); err != nil {
		respondErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Vehicles ---

func (h *Handler) createVehicle(w http.ResponseWriter, r *http.Request) {
	clientID, err := parseUUID(r, "id")
	if err != nil {
		respond(w, http.StatusBadRequest, map[string]string{"error": "invalid client uuid"})
		return
	}
	var req struct {
		Brand string `json:"brand"`
		Model string `json:"model"`
		Plate string `json:"plate"`
	}
	if err := decode(r, &req); err != nil {
		respond(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	v, err := h.clients.CreateVehicle(r.Context(), clientsvc.CreateVehicleInput{
		ClientID: clientID,
		Brand:    req.Brand,
		Model:    req.Model,
		Plate:    req.Plate,
	})
	if err != nil {
		respondErr(w, err)
		return
	}
	respond(w, http.StatusCreated, v)
}

func (h *Handler) listVehicles(w http.ResponseWriter, r *http.Request) {
	clientID, err := parseUUID(r, "id")
	if err != nil {
		respond(w, http.StatusBadRequest, map[string]string{"error": "invalid client uuid"})
		return
	}
	list, err := h.clients.ListVehiclesByClient(r.Context(), clientID)
	if err != nil {
		respondErr(w, err)
		return
	}
	respond(w, http.StatusOK, list)
}

func (h *Handler) getVehicle(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(r, "id")
	if err != nil {
		respond(w, http.StatusBadRequest, map[string]string{"error": "invalid uuid"})
		return
	}
	v, err := h.clients.GetVehicle(r.Context(), id)
	if err != nil {
		respondErr(w, err)
		return
	}
	respond(w, http.StatusOK, v)
}

func (h *Handler) updateVehicle(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(r, "id")
	if err != nil {
		respond(w, http.StatusBadRequest, map[string]string{"error": "invalid uuid"})
		return
	}
	var req struct {
		ClientID uuid.UUID `json:"client_id"`
		Brand    string    `json:"brand"`
		Model    string    `json:"model"`
		Plate    string    `json:"plate"`
	}
	if err := decode(r, &req); err != nil {
		respond(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}
	v, err := h.clients.UpdateVehicle(r.Context(), domainVehicle(id, req.ClientID, req.Brand, req.Model, req.Plate))
	if err != nil {
		respondErr(w, err)
		return
	}
	respond(w, http.StatusOK, v)
}

func domainVehicle(id, clientID uuid.UUID, brand, model, plate string) *domain.Vehicle {
	return &domain.Vehicle{
		VehicleID: id,
		ClientID:  clientID,
		Brand:     brand,
		Model:     model,
		Plate:     plate,
	}
}

func (h *Handler) deleteVehicle(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(r, "id")
	if err != nil {
		respond(w, http.StatusBadRequest, map[string]string{"error": "invalid uuid"})
		return
	}
	if err := h.clients.DeleteVehicle(r.Context(), id); err != nil {
		respondErr(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
