package calendar

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
	repo := NewRepository(db)
	svc := NewService(repo)
	return &Handler{service: svc}
}

func roleAndBranch(r *http.Request) (string, string) {
	role, _ := r.Context().Value(middleware.UserRoleKey).(string)
	branchID, _ := r.Context().Value(middleware.UserBranchIDKey).(string)
	return role, branchID
}

// GET /calendar/branch-calendar
// Query params: from, to, type
func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	role, branchID := roleAndBranch(r)
	q := r.URL.Query()

	filter := CalendarRangeFilter{
		From: q.Get("from"),
		To:   q.Get("to"),
		Type: q.Get("type"),
	}

	entries, err := h.service.GetAll(r.Context(), role, branchID, filter)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch calendar entries")
		return
	}
	response.Success(w, entries)
}

// GET /calendar/branch-calendar/{id}
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	role, branchID := roleAndBranch(r)

	entry, err := h.service.GetByID(r.Context(), id, role, branchID)
	if err != nil {
		if err.Error() == "forbidden" {
			response.Error(w, http.StatusForbidden, "insufficient permissions")
			return
		}
		response.Error(w, http.StatusNotFound, "calendar entry not found")
		return
	}
	response.Success(w, entry)
}

// POST /calendar/branch-calendar
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	role, branchID := roleAndBranch(r)

	var req CreateCalendarEntryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := validator.Validate(req); err != nil {
		response.Error(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	entry, err := h.service.Create(r.Context(), role, branchID, req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to create calendar entry")
		return
	}
	response.Created(w, entry)
}

// PUT /calendar/branch-calendar/{id}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	role, branchID := roleAndBranch(r)

	var req UpdateCalendarEntryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := validator.Validate(req); err != nil {
		response.Error(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	entry, err := h.service.Update(r.Context(), id, role, branchID, req)
	if err != nil {
		if err.Error() == "forbidden" {
			response.Error(w, http.StatusForbidden, "insufficient permissions")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to update calendar entry")
		return
	}
	response.Success(w, entry)
}

// DELETE /calendar/branch-calendar/{id}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	role, branchID := roleAndBranch(r)

	if err := h.service.Delete(r.Context(), id, role, branchID); err != nil {
		if err.Error() == "forbidden" {
			response.Error(w, http.StatusForbidden, "insufficient permissions")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to delete calendar entry")
		return
	}
	response.Success(w, nil)
}
