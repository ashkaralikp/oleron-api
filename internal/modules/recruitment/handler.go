package recruitment

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"rmp-api/internal/dbscope"
	"rmp-api/internal/middleware"
	"rmp-api/pkg/response"
	"rmp-api/pkg/validator"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	db        *pgxpool.Pool
	service   *Service
	uploadDir string
	baseURL   string
}

func NewHandler(db *pgxpool.Pool, uploadDir, baseURL string) *Handler {
	return &Handler{
		db:        db,
		service:   NewService(NewRepository(db)),
		uploadDir: uploadDir,
		baseURL:   baseURL,
	}
}

func ctxVars(r *http.Request) (role, branchID, userID string) {
	role, _ = r.Context().Value(middleware.UserRoleKey).(string)
	branchID, _ = r.Context().Value(middleware.UserBranchIDKey).(string)
	userID, _ = r.Context().Value(middleware.UserIDKey).(string)
	return
}

// ─────────────────────────────────────────────
// PUBLIC — no JWT required
// ─────────────────────────────────────────────

// POST /recruitment/upload/cv
func (h *Handler) UploadCV(w http.ResponseWriter, r *http.Request) {
	const maxSize = 5 << 20 // 5 MB
	r.Body = http.MaxBytesReader(w, r.Body, maxSize)

	if err := r.ParseMultipartForm(maxSize); err != nil {
		response.Error(w, http.StatusUnprocessableEntity, "file too large: max 5MB")
		return
	}

	file, header, err := r.FormFile("cv")
	if err != nil {
		response.Error(w, http.StatusBadRequest, "cv file is required")
		return
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowed := map[string]bool{".pdf": true, ".doc": true, ".docx": true}
	if !allowed[ext] {
		response.Error(w, http.StatusUnprocessableEntity, "only PDF, DOC, and DOCX files are allowed")
		return
	}

	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to generate filename")
		return
	}
	filename := hex.EncodeToString(b) + ext
	destDir := filepath.Join(h.uploadDir, "cvs")

	if err := os.MkdirAll(destDir, 0755); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to store file")
		return
	}

	dst, err := os.Create(filepath.Join(destDir, filename))
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to store file")
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to store file")
		return
	}

	cvURL := strings.TrimRight(h.baseURL, "/") + "/uploads/cvs/" + filename
	response.Created(w, map[string]string{"cv_url": cvURL})
}

// GET /recruitment/vacancies/public
func (h *Handler) GetPublicVacancies(w http.ResponseWriter, r *http.Request) {
	vacancies, err := h.service.GetPublicVacancies(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch vacancies")
		return
	}
	response.Success(w, vacancies)
}

// POST /recruitment/vacancies/{id}/apply
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
	role, branchID, userID := ctxVars(r)

	branchIDs, err := dbscope.GetEffectiveBranchIDs(r.Context(), h.db, role, userID, branchID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to resolve branch access")
		return
	}

	vacancies, err := h.service.GetAllVacancies(r.Context(), branchIDs)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch vacancies")
		return
	}
	response.Success(w, vacancies)
}

// GET /recruitment/vacancies/{id}
func (h *Handler) GetVacancyByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	role, branchID, userID := ctxVars(r)

	branchIDs, err := dbscope.GetEffectiveBranchIDs(r.Context(), h.db, role, userID, branchID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to resolve branch access")
		return
	}

	v, err := h.service.GetVacancyByID(r.Context(), id, branchIDs)
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
	role, branchID, userID := ctxVars(r)

	branchIDs, err := dbscope.GetEffectiveBranchIDs(r.Context(), h.db, role, userID, branchID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to resolve branch access")
		return
	}

	var req CreateVacancyRequest
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

	v, err := h.service.CreateVacancy(r.Context(), targetBranch, userID, req)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to create vacancy")
		return
	}
	response.Created(w, v)
}

// PUT /recruitment/vacancies/{id}
func (h *Handler) UpdateVacancy(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	role, branchID, userID := ctxVars(r)

	branchIDs, err := dbscope.GetEffectiveBranchIDs(r.Context(), h.db, role, userID, branchID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to resolve branch access")
		return
	}

	var req UpdateVacancyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := validator.Validate(req); err != nil {
		response.Error(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	v, err := h.service.UpdateVacancy(r.Context(), id, branchIDs, req)
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
	role, branchID, userID := ctxVars(r)

	branchIDs, err := dbscope.GetEffectiveBranchIDs(r.Context(), h.db, role, userID, branchID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to resolve branch access")
		return
	}

	var req UpdateVacancyStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := validator.Validate(req); err != nil {
		response.Error(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	v, err := h.service.UpdateVacancyStatus(r.Context(), id, branchIDs, req)
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
	role, branchID, userID := ctxVars(r)

	branchIDs, err := dbscope.GetEffectiveBranchIDs(r.Context(), h.db, role, userID, branchID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to resolve branch access")
		return
	}

	if err := h.service.DeleteVacancy(r.Context(), id, branchIDs); err != nil {
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

// POST /recruitment/vacancies/{id}/apply/bulk
func (h *Handler) BulkApply(w http.ResponseWriter, r *http.Request) {
	vacancyID := chi.URLParam(r, "id")
	role, branchID, userID := ctxVars(r)

	branchIDs, err := dbscope.GetEffectiveBranchIDs(r.Context(), h.db, role, userID, branchID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to resolve branch access")
		return
	}

	var req BulkApplyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := validator.Validate(req); err != nil {
		response.Error(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	result, err := h.service.BulkApply(r.Context(), vacancyID, branchIDs, req)
	if err != nil {
		switch err.Error() {
		case "forbidden":
			response.Error(w, http.StatusForbidden, "insufficient permissions")
		case "vacancy not found":
			response.Error(w, http.StatusNotFound, err.Error())
		default:
			response.Error(w, http.StatusInternalServerError, "failed to process applications")
		}
		return
	}
	response.Created(w, result)
}

// GET /recruitment/vacancies/{id}/applications?status=
func (h *Handler) GetApplicationsByVacancy(w http.ResponseWriter, r *http.Request) {
	vacancyID := chi.URLParam(r, "id")
	role, branchID, userID := ctxVars(r)
	statusFilter := r.URL.Query().Get("status")

	branchIDs, err := dbscope.GetEffectiveBranchIDs(r.Context(), h.db, role, userID, branchID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to resolve branch access")
		return
	}

	apps, err := h.service.GetApplicationsByVacancy(r.Context(), vacancyID, branchIDs, statusFilter)
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
	role, branchID, userID := ctxVars(r)

	branchIDs, err := dbscope.GetEffectiveBranchIDs(r.Context(), h.db, role, userID, branchID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to resolve branch access")
		return
	}

	app, err := h.service.GetApplicationByID(r.Context(), id, branchIDs)
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
	role, branchID, userID := ctxVars(r)

	branchIDs, err := dbscope.GetEffectiveBranchIDs(r.Context(), h.db, role, userID, branchID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to resolve branch access")
		return
	}

	var req UpdateApplicationStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := validator.Validate(req); err != nil {
		response.Error(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	app, err := h.service.UpdateApplicationStatus(r.Context(), id, branchIDs, req)
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
	role, branchID, userID := ctxVars(r)

	branchIDs, err := dbscope.GetEffectiveBranchIDs(r.Context(), h.db, role, userID, branchID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to resolve branch access")
		return
	}

	if err := h.service.DeleteApplication(r.Context(), id, branchIDs); err != nil {
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
	role, branchID, userID := ctxVars(r)

	branchIDs, err := dbscope.GetEffectiveBranchIDs(r.Context(), h.db, role, userID, branchID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to resolve branch access")
		return
	}

	var req CreateInterviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := validator.Validate(req); err != nil {
		response.Error(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	interview, err := h.service.CreateInterview(r.Context(), applicationID, branchIDs, req)
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
	role, branchID, userID := ctxVars(r)

	branchIDs, err := dbscope.GetEffectiveBranchIDs(r.Context(), h.db, role, userID, branchID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to resolve branch access")
		return
	}

	var req UpdateInterviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := validator.Validate(req); err != nil {
		response.Error(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	interview, err := h.service.UpdateInterview(r.Context(), id, branchIDs, req)
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
	role, branchID, userID := ctxVars(r)

	branchIDs, err := dbscope.GetEffectiveBranchIDs(r.Context(), h.db, role, userID, branchID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to resolve branch access")
		return
	}

	if err := h.service.DeleteInterview(r.Context(), id, branchIDs); err != nil {
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
	role, branchID, userID := ctxVars(r)

	branchIDs, err := dbscope.GetEffectiveBranchIDs(r.Context(), h.db, role, userID, branchID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to resolve branch access")
		return
	}

	var req HireRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if err := validator.Validate(req); err != nil {
		response.Error(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	result, err := h.service.Hire(r.Context(), id, branchIDs, req)
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
