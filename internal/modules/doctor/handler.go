package doctor

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
	doctors, err := h.service.GetAll(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch doctors")
		return
	}

	response.Success(w, doctors)
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	doctor, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusNotFound, "doctor not found")
		return
	}

	response.Success(w, doctor)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateDoctorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	doctor, err := h.service.Create(r.Context(), req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to create doctor")
		return
	}

	response.Created(w, doctor)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req UpdateDoctorRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.service.Update(r.Context(), id, req); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to update doctor")
		return
	}

	response.Success(w, map[string]string{"message": "doctor updated"})
}
