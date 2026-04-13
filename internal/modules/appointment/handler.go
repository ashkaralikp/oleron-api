package appointment

import (
	"encoding/json"
	"net/http"

	"clinic-api/pkg/response"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	service *Service
}

func NewHandler(db *pgxpool.Pool) *Handler {
	repo := NewRepository(db)
	svc := NewService(repo)
	return &Handler{service: svc}
}

func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	appointments, err := h.service.GetAll(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch appointments")
		return
	}

	response.Success(w, appointments)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateAppointmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	appt, err := h.service.Create(r.Context(), req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to create appointment")
		return
	}

	response.Created(w, appt)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req UpdateAppointmentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.service.Update(r.Context(), id, req); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to update appointment")
		return
	}

	response.Success(w, map[string]string{"message": "appointment updated"})
}
