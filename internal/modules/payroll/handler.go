package payroll

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

// POST /payroll/generate
func (h *Handler) Generate(w http.ResponseWriter, r *http.Request) {
	role, branchID, userID := ctxVars(r)

	var req GeneratePayrollRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := validator.Validate(req); err != nil {
		response.Error(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	branchIDs, err := dbscope.GetEffectiveBranchIDs(r.Context(), h.db, role, userID, branchID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to resolve branch access")
		return
	}

	// Regional managers specify the target branch in the request body;
	// other roles use their home branch.
	targetBranch := branchID
	if req.BranchID != "" {
		targetBranch = req.BranchID
	}
	if !dbscope.ContainsBranch(branchIDs, targetBranch) {
		response.Error(w, http.StatusForbidden, "not authorized for this branch")
		return
	}

	run, err := h.service.Generate(r.Context(), targetBranch, userID, req)
	if err != nil {
		switch err.Error() {
		case "no active employees found in this branch":
			response.Error(w, http.StatusUnprocessableEntity, err.Error())
		default:
			response.Error(w, http.StatusInternalServerError, "failed to generate payroll")
		}
		return
	}
	response.Created(w, run)
}

// GET /payroll
func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	role, branchID, userID := ctxVars(r)

	branchIDs, err := dbscope.GetEffectiveBranchIDs(r.Context(), h.db, role, userID, branchID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to resolve branch access")
		return
	}

	runs, err := h.service.GetAll(r.Context(), branchIDs)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch payroll runs")
		return
	}
	response.Success(w, runs)
}

// GET /payroll/{id}
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	role, branchID, userID := ctxVars(r)

	branchIDs, err := dbscope.GetEffectiveBranchIDs(r.Context(), h.db, role, userID, branchID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to resolve branch access")
		return
	}

	run, err := h.service.GetByID(r.Context(), id, branchIDs)
	if err != nil {
		if err.Error() == "forbidden" {
			response.Error(w, http.StatusForbidden, "insufficient permissions")
			return
		}
		response.Error(w, http.StatusNotFound, "payroll run not found")
		return
	}
	response.Success(w, run)
}

// PATCH /payroll/{id}/status
func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	role, branchID, userID := ctxVars(r)

	var req UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := validator.Validate(req); err != nil {
		response.Error(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	branchIDs, err := dbscope.GetEffectiveBranchIDs(r.Context(), h.db, role, userID, branchID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to resolve branch access")
		return
	}

	run, err := h.service.UpdateStatus(r.Context(), id, branchIDs, req)
	if err != nil {
		switch err.Error() {
		case "forbidden":
			response.Error(w, http.StatusForbidden, "insufficient permissions")
		case "payroll run not found":
			response.Error(w, http.StatusNotFound, err.Error())
		default:
			response.Error(w, http.StatusUnprocessableEntity, err.Error())
		}
		return
	}
	response.Success(w, run)
}

// DELETE /payroll/{id}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	role, branchID, userID := ctxVars(r)

	branchIDs, err := dbscope.GetEffectiveBranchIDs(r.Context(), h.db, role, userID, branchID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to resolve branch access")
		return
	}

	if err := h.service.Delete(r.Context(), id, branchIDs); err != nil {
		switch err.Error() {
		case "forbidden":
			response.Error(w, http.StatusForbidden, "insufficient permissions")
		case "payroll run not found":
			response.Error(w, http.StatusNotFound, err.Error())
		case "only draft payroll runs can be deleted":
			response.Error(w, http.StatusUnprocessableEntity, err.Error())
		default:
			response.Error(w, http.StatusInternalServerError, "failed to delete payroll run")
		}
		return
	}
	response.Success(w, nil)
}
