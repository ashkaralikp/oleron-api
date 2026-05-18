package schedule

import (
	"encoding/json"
	"errors"
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

// GET /schedule/office-timings
func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	role, branchID, userID := ctxVars(r)

	branchIDs, err := dbscope.GetEffectiveBranchIDs(r.Context(), h.db, role, userID, branchID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to resolve branch access")
		return
	}

	timings, err := h.service.GetAll(r.Context(), branchIDs)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch office timings")
		return
	}
	response.Success(w, timings)
}

// GET /schedule/office-timings/{id}
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	role, branchID, userID := ctxVars(r)

	branchIDs, err := dbscope.GetEffectiveBranchIDs(r.Context(), h.db, role, userID, branchID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to resolve branch access")
		return
	}

	t, err := h.service.GetByID(r.Context(), id, branchIDs)
	if err != nil {
		if errors.Is(err, errors.New("forbidden")) || err.Error() == "forbidden" {
			response.Error(w, http.StatusForbidden, "insufficient permissions")
			return
		}
		response.Error(w, http.StatusNotFound, "office timing not found")
		return
	}
	response.Success(w, t)
}

// POST /schedule/office-timings
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	role, branchID, userID := ctxVars(r)

	branchIDs, err := dbscope.GetEffectiveBranchIDs(r.Context(), h.db, role, userID, branchID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to resolve branch access")
		return
	}

	var req CreateOfficeTimingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := validator.Validate(req); err != nil {
		response.Error(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	targetBranch := branchID
	if !dbscope.ContainsBranch(branchIDs, targetBranch) {
		response.Error(w, http.StatusForbidden, "not authorized for this branch")
		return
	}

	t, err := h.service.Create(r.Context(), targetBranch, req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to create office timing")
		return
	}
	response.Created(w, t)
}

// PUT /schedule/office-timings/{id}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	role, branchID, userID := ctxVars(r)

	branchIDs, err := dbscope.GetEffectiveBranchIDs(r.Context(), h.db, role, userID, branchID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to resolve branch access")
		return
	}

	var req UpdateOfficeTimingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := validator.Validate(req); err != nil {
		response.Error(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	t, err := h.service.Update(r.Context(), id, branchIDs, req)
	if err != nil {
		if err.Error() == "forbidden" {
			response.Error(w, http.StatusForbidden, "insufficient permissions")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to update office timing")
		return
	}
	response.Success(w, t)
}

// DELETE /schedule/office-timings/{id}
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
		response.Error(w, http.StatusInternalServerError, "failed to delete office timing")
		return
	}
	response.Success(w, nil)
}

// PUT /schedule/office-timings/{id}/activate
func (h *Handler) Activate(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	role, branchID, userID := ctxVars(r)

	branchIDs, err := dbscope.GetEffectiveBranchIDs(r.Context(), h.db, role, userID, branchID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to resolve branch access")
		return
	}

	if err := h.service.Activate(r.Context(), id, branchIDs); err != nil {
		if err.Error() == "forbidden" {
			response.Error(w, http.StatusForbidden, "insufficient permissions")
			return
		}
		if err.Error() == "timing not found" {
			response.Error(w, http.StatusNotFound, "office timing not found")
			return
		}
		response.Error(w, http.StatusInternalServerError, "failed to activate office timing")
		return
	}
	response.Success(w, nil)
}
