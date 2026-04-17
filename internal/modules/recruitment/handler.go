package recruitment

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

func ctx(r *http.Request) (role, branchID, userID string) {
	role, _ = r.Context().Value(middleware.UserRoleKey).(string)
	branchID, _ = r.Context().Value(middleware.UserBranchIDKey).(string)
	userID, _ = r.Context().Value(middleware.UserIDKey).(string)
	return
}

// ─────────────────────────────────────────────
// PUBLIC — candidates apply (no JWT)
// POST /recruitment/vacancies/{id}/apply
// ─────────────────────────────────────────────

func (h *Handler) Apply(w http.ResponseWriter, r *http.Request) {
	vacancyID := chi.URLParam(r, "id")

	var req ApplyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := validator.Validate(req); err != nil {
		response.Error(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	app, err := h.service.Apply(r.Context(), vacancyID, req)
	if err != nil {
		switch err.Error() {
		case "vacancy not found":
			response.Error(w, http.StatusNotFound, err.Error())
		case "vacancy is not open for applications":
			response.Error(w, http.StatusUnprocessableEntity, err.Error())
		default:
			response.Error(w, http.StatusInternalServerError, "failed to submit application")
		}
		return
	}
	response.Created(w, app)
}

// ─────────────────────────────────────────────
// VACANCIES
// ─────────────────────────────────────────────

// GET /recruitment/vacancies
func (h *Handler) GetAllVacancies(w http.ResponseWriter, r *http.Request) {
	role, branchID, _ := ctx(r)
	vacancies, err := h.service.GetAllVacancies(r.Context(), role, branchID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch vacancies")
		return
	}
	response.Success(w, vacancies)
}

// GET /recruitment/vacancies/{id}
func (h *Handler) GetVacancyByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	role, branchID, _ := ctx(r)

	v, err := h.service.GetVacancyByID(r.Context(), id, role, branchID)
	if err != nil {
		if err.Error() == "forbidden" {
			response.Error(w, http.StatusForbidden, "insufficient permissions")
			return
		}
		response.Error(w, http.StatusNotFound, "vacancy not found")
		return
	}
	response.Success(w, v)
}

// POST /recruitment/vacancies
func (h *Handler) CreateVacancy(w http.ResponseWriter, r *http.Request) {
	_, branchID, userID := ctx(r)

	var req CreateVacancyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := validator.Validate(req); err != nil {
		response.Error(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	v, err := h.service.CreateVacancy(r.Context(), branchID, userID, req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to create vacancy")
		return
	}
	response.Created(w, v)
}

// PUT /recruitment/vacancies/{id}
func (h *Handler) UpdateVacancy(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	role, branchID, _ := ctx(r)

	var req UpdateVacancyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := validator.Validate(req); err != nil {
		response.Error(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	v, err := h.service.UpdateVacancy(r.Context(), id, role, branchID, req)
	if err != nil {
		switch err.Error() {
		case "forbidden":
			response.Error(w, http.StatusForbidden, "insufficient permissions")
		default:
			response.Error(w, http.StatusNotFound, "vacancy not found")
		}
		return
	}
	response.Success(w, v)
}

// PATCH /recruitment/vacancies/{id}/status
func (h *Handler) UpdateVacancyStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	role, branchID, _ := ctx(r)

	var req UpdateVacancyStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := validator.Validate(req); err != nil {
		response.Error(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	v, err := h.service.UpdateVacancyStatus(r.Context(), id, role, branchID, req)
	if err != nil {
		switch err.Error() {
		case "forbidden":
			response.Error(w, http.StatusForbidden, "insufficient permissions")
		default:
			response.Error(w, http.StatusNotFound, "vacancy not found")
		}
		return
	}
	response.Success(w, v)
}

// DELETE /recruitment/vacancies/{id}
func (h *Handler) DeleteVacancy(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	role, branchID, _ := ctx(r)

	if err := h.service.DeleteVacancy(r.Context(), id, role, branchID); err != nil {
		switch err.Error() {
		case "forbidden":
			response.Error(w, http.StatusForbidden, "insufficient permissions")
		case "only draft vacancies can be deleted":
			response.Error(w, http.StatusUnprocessableEntity, err.Error())
		default:
			response.Error(w, http.StatusNotFound, "vacancy not found")
		}
		return
	}
	response.Success(w, nil)
}

// ─────────────────────────────────────────────
// APPLICATIONS
// ─────────────────────────────────────────────

// GET /recruitment/vacancies/{id}/applications?status=
func (h *Handler) GetApplicationsByVacancy(w http.ResponseWriter, r *http.Request) {
	vacancyID := chi.URLParam(r, "id")
	role, branchID, _ := ctx(r)
	statusFilter := r.URL.Query().Get("status")

	apps, err := h.service.GetApplicationsByVacancy(r.Context(), vacancyID, role, branchID, statusFilter)
	if err != nil {
		switch err.Error() {
		case "forbidden":
			response.Error(w, http.StatusForbidden, "insufficient permissions")
		default:
			response.Error(w, http.StatusNotFound, "vacancy not found")
		}
		return
	}
	response.Success(w, apps)
}

// GET /recruitment/applications/{id}
func (h *Handler) GetApplicationByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	role, branchID, _ := ctx(r)

	app, err := h.service.GetApplicationByID(r.Context(), id, role, branchID)
	if err != nil {
		if err.Error() == "forbidden" {
			response.Error(w, http.StatusForbidden, "insufficient permissions")
			return
		}
		response.Error(w, http.StatusNotFound, "application not found")
		return
	}
	response.Success(w, app)
}

// PATCH /recruitment/applications/{id}/status
func (h *Handler) UpdateApplicationStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	role, branchID, _ := ctx(r)

	var req UpdateApplicationStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := validator.Validate(req); err != nil {
		response.Error(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	app, err := h.service.UpdateApplicationStatus(r.Context(), id, role, branchID, req)
	if err != nil {
		switch err.Error() {
		case "forbidden":
			response.Error(w, http.StatusForbidden, "insufficient permissions")
		default:
			response.Error(w, http.StatusNotFound, "application not found")
		}
		return
	}
	response.Success(w, app)
}

// DELETE /recruitment/applications/{id}
func (h *Handler) DeleteApplication(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	role, branchID, _ := ctx(r)

	if err := h.service.DeleteApplication(r.Context(), id, role, branchID); err != nil {
		switch err.Error() {
		case "forbidden":
			response.Error(w, http.StatusForbidden, "insufficient permissions")
		default:
			response.Error(w, http.StatusNotFound, "application not found")
		}
		return
	}
	response.Success(w, nil)
}

// ─────────────────────────────────────────────
// INTERVIEWS
// ─────────────────────────────────────────────

// POST /recruitment/applications/{id}/interviews
func (h *Handler) CreateInterview(w http.ResponseWriter, r *http.Request) {
	applicationID := chi.URLParam(r, "id")
	role, branchID, _ := ctx(r)

	var req CreateInterviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := validator.Validate(req); err != nil {
		response.Error(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	interview, err := h.service.CreateInterview(r.Context(), applicationID, role, branchID, req)
	if err != nil {
		switch err.Error() {
		case "forbidden":
			response.Error(w, http.StatusForbidden, "insufficient permissions")
		case "invalid scheduled_at: use RFC3339 format (e.g. 2026-04-17T10:00:00Z)":
			response.Error(w, http.StatusUnprocessableEntity, err.Error())
		default:
			response.Error(w, http.StatusNotFound, "application not found")
		}
		return
	}
	response.Created(w, interview)
}

// PUT /recruitment/interviews/{id}
func (h *Handler) UpdateInterview(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	role, branchID, _ := ctx(r)

	var req UpdateInterviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := validator.Validate(req); err != nil {
		response.Error(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	interview, err := h.service.UpdateInterview(r.Context(), id, role, branchID, req)
	if err != nil {
		switch err.Error() {
		case "forbidden":
			response.Error(w, http.StatusForbidden, "insufficient permissions")
		default:
			response.Error(w, http.StatusNotFound, "interview not found")
		}
		return
	}
	response.Success(w, interview)
}

// DELETE /recruitment/interviews/{id}
func (h *Handler) DeleteInterview(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	role, branchID, _ := ctx(r)

	if err := h.service.DeleteInterview(r.Context(), id, role, branchID); err != nil {
		switch err.Error() {
		case "forbidden":
			response.Error(w, http.StatusForbidden, "insufficient permissions")
		default:
			response.Error(w, http.StatusNotFound, "interview not found")
		}
		return
	}
	response.Success(w, nil)
}

// ─────────────────────────────────────────────
// HIRE
// ─────────────────────────────────────────────

// POST /recruitment/applications/{id}/hire
func (h *Handler) Hire(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	role, branchID, _ := ctx(r)

	var req HireRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := validator.Validate(req); err != nil {
		response.Error(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	result, err := h.service.Hire(r.Context(), id, role, branchID, req)
	if err != nil {
		switch err.Error() {
		case "forbidden":
			response.Error(w, http.StatusForbidden, "insufficient permissions")
		case "application not found":
			response.Error(w, http.StatusNotFound, err.Error())
		default:
			response.Error(w, http.StatusUnprocessableEntity, err.Error())
		}
		return
	}
	response.Created(w, result)
}
