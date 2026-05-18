package calendar

import (
	"encoding/json"
	"net/http"

	"rmp-api/internal/dbscope"
	"rmp-api/internal/middleware"
	"rmp-api/pkg/response"
	"rmp-api/pkg/validator"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	db      *pgxpool.Pool
	service *Service
}

func NewHandler(db *pgxpool.Pool) *Handler {
	return &Handler{db: db, service: NewService(NewRepository(db))}
}

func ctxVars(r *http.Request) (role, branchID, userID string) {
	role, _ = r.Context().Value(middleware.UserRoleKey).(string)
	branchID, _ = r.Context().Value(middleware.UserBranchIDKey).(string)
	userID, _ = r.Context().Value(middleware.UserIDKey).(string)
	return
}

// GET /calendar/branch-calendar
func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	role, branchID, userID := ctxVars(r)

	branchIDs, err := dbscope.GetEffectiveBranchIDs(r.Context(), h.db, role, userID, branchID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to resolve branch access")
		return
	}

	q := r.URL.Query()
	filter := CalendarRangeFilter{
		From: q.Get("from"),
		To:   q.Get("to"),
		Type: q.Get("type"),
	}

	entries, err := h.service.GetAll(r.Context(), branchIDs, filter)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch calendar entries")
		return
	}
	response.Success(w, entries)
}

// GET /calendar/branch-calendar/{id}
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	role, branchID, userID := ctxVars(r)

	branchIDs, err := dbscope.GetEffectiveBranchIDs(r.Context(), h.db, role, userID, branchID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to resolve branch access")
		return
	}

	entry, err := h.service.GetByID(r.Context(), id, branchIDs)
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
	role, branchID, userID := ctxVars(r)

	branchIDs, err := dbscope.GetEffectiveBranchIDs(r.Context(), h.db, role, userID, branchID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to resolve branch access")
		return
	}

	var req CreateCalendarEntryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := validator.Validate(req); err != nil {
		response.Error(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	// Use home branch by default; validate against allowed branches.
	targetBranch := branchID
	if !dbscope.ContainsBranch(branchIDs, targetBranch) {
		response.Error(w, http.StatusForbidden, "not authorized for this branch")
		return
	}

	entry, err := h.service.Create(r.Context(), targetBranch, req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to create calendar entry")
		return
	}
	response.Created(w, entry)
}

// PUT /calendar/branch-calendar/{id}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	role, branchID, userID := ctxVars(r)

	branchIDs, err := dbscope.GetEffectiveBranchIDs(r.Context(), h.db, role, userID, branchID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to resolve branch access")
		return
	}

	var req UpdateCalendarEntryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := validator.Validate(req); err != nil {
		response.Error(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	entry, err := h.service.Update(r.Context(), id, branchIDs, req)
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
	role, branchID, userID := ctxVars(r)

	branchIDs, err := dbscope.GetEffectiveBranchIDs(r.Context(), h.db, role, userID, branchID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to resolve branch access")
		return
	}

	if err := h.service.Delete(r.Context(), id, branchIDs); err != nil {
		if err.Error() == "forbidden" {
			response.Error(w, http.StatusForbidden, "insufficient permissions")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to delete calendar entry")
		return
	}
	response.Success(w, nil)
}
