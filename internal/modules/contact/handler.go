package contact

import (
	"encoding/json"
	"net/http"

	"rmp-api/internal/middleware"
	"rmp-api/pkg/response"
	"rmp-api/pkg/validator"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	service *Service
}

func NewHandler(db *pgxpool.Pool) *Handler {
	return &Handler{service: NewService(NewRepository(db))}
}

func (h *Handler) CreateSubmission(w http.ResponseWriter, r *http.Request) {
	var req CreateSubmissionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validator.Validate(req); err != nil {
		response.Error(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	submission, err := h.service.CreateSubmission(
		r.Context(),
		req,
		middleware.ClientIP(r),
		r.UserAgent(),
	)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to submit contact form")
		return
	}

	response.Created(w, submission)
}

func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	statusFilter := r.URL.Query().Get("status")

	submissions, err := h.service.GetAll(r.Context(), statusFilter)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch submissions")
		return
	}

	response.Success(w, submissions)
}

func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	submission, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch submission")
		return
	}
	if submission == nil {
		response.Error(w, http.StatusNotFound, "submission not found")
		return
	}

	response.Success(w, submission)
}

func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req UpdateSubmissionStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if err := validator.Validate(req); err != nil {
		response.Error(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	submission, err := h.service.UpdateStatus(r.Context(), id, req.Status)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to update submission")
		return
	}
	if submission == nil {
		response.Error(w, http.StatusNotFound, "submission not found")
		return
	}

	response.Success(w, submission)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	deleted, err := h.service.Delete(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to delete submission")
		return
	}
	if !deleted {
		response.Error(w, http.StatusNotFound, "submission not found")
		return
	}

	response.Success(w, map[string]string{"message": "submission deleted"})
}
