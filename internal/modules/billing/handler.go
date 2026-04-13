package billing

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
	billings, err := h.service.GetAll(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch billing records")
		return
	}

	response.Success(w, billings)
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	billing, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusNotFound, "billing record not found")
		return
	}

	response.Success(w, billing)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var req CreateBillingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	billing, err := h.service.Create(r.Context(), req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to create billing record")
		return
	}

	response.Created(w, billing)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req UpdateBillingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := h.service.Update(r.Context(), id, req); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to update billing record")
		return
	}

	response.Success(w, map[string]string{"message": "billing record updated"})
}
