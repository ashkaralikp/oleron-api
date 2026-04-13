package patient

import (
	"encoding/json"
	"net/http"

	"rmp-api/pkg/response"

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
	patients, err := h.service.GetAll(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch patients")
		return
	}

	response.Success(w, patients)
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	patient, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusNotFound, "patient not found")
		return
	}

	response.Success(w, patient)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreatePatientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	patient, err := h.service.Create(r.Context(), req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to create patient")
		return
	}

	response.Created(w, patient)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req UpdatePatientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.service.Update(r.Context(), id, req); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to update patient")
		return
	}

	response.Success(w, map[string]string{"message": "patient updated"})
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.service.Delete(r.Context(), id); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to delete patient")
		return
	}

	response.Success(w, map[string]string{"message": "patient deleted"})
}
