package workorder

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log"
	"mime/multipart"
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

// POST /work-orders/assets
func (h *Handler) UpsertAsset(w http.ResponseWriter, r *http.Request) {
	_, _, userID := ctxVars(r)

	const maxSize = 10 << 20 // signature + seal, 10 MB total
	r.Body = http.MaxBytesReader(w, r.Body, maxSize)
	if err := r.ParseMultipartForm(maxSize); err != nil {
		response.Error(w, http.StatusUnprocessableEntity, "files too large: max 10MB total")
		return
	}

	signerName := strings.TrimSpace(r.FormValue("signer_name"))
	signerTitle := optionalFormValue(r, "signer_title")
	req := UpsertAssetRequest{SignerName: signerName, SignerTitle: signerTitle}
	if err := validator.Validate(req); err != nil {
		response.Error(w, http.StatusUnprocessableEntity, err.Error())
		return
	}

	signatureURL, err := h.storeUpload(r, "signature", userID)
	if err != nil {
		writeUploadError(w, err)
		return
	}
	sealURL, err := h.storeUpload(r, "seal", userID)
	if err != nil {
		writeUploadError(w, err)
		return
	}

	asset, err := h.service.UpsertAsset(r.Context(), userID, req.SignerName, req.SignerTitle, signatureURL, sealURL, &userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to save work order assets")
		return
	}
	response.Created(w, asset)
}

// GET /work-orders/assets/me
func (h *Handler) GetMyAsset(w http.ResponseWriter, r *http.Request) {
	_, _, userID := ctxVars(r)

	asset, err := h.service.GetAsset(r.Context(), userID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch work order assets")
		return
	}
	if asset == nil {
		response.Error(w, http.StatusNotFound, "work order assets not found")
		return
	}
	response.Success(w, asset)
}

// POST /work-orders
func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	role, branchID, userID := ctxVars(r)

	var req CreateWorkOrderRequest
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

	wo, err := h.service.Create(r.Context(), branchIDs, branchID, userID, role, req)
	if err != nil {
		switch err.Error() {
		case "forbidden":
			response.Error(w, http.StatusForbidden, "not authorized for this branch")
		case "branch_id is required", "invalid work_order_date: use YYYY-MM-DD", "work order signature and seal must be uploaded first", "item description is required", "duplicate item line_no":
			response.Error(w, http.StatusUnprocessableEntity, err.Error())
		default:
			log.Printf("work order create failed: %v", err)
			response.Error(w, http.StatusInternalServerError, "failed to create work order")
		}
		return
	}
	response.Created(w, wo)
}

// GET /work-orders
func (h *Handler) GetAll(w http.ResponseWriter, r *http.Request) {
	role, branchID, userID := ctxVars(r)

	branchIDs, err := dbscope.GetEffectiveBranchIDs(r.Context(), h.db, role, userID, branchID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to resolve branch access")
		return
	}

	workOrders, err := h.service.GetAll(r.Context(), branchIDs)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to fetch work orders")
		return
	}
	response.Success(w, workOrders)
}

// GET /work-orders/{id}
func (h *Handler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	role, branchID, userID := ctxVars(r)

	branchIDs, err := dbscope.GetEffectiveBranchIDs(r.Context(), h.db, role, userID, branchID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to resolve branch access")
		return
	}

	wo, err := h.service.GetByID(r.Context(), id, branchIDs)
	if err != nil {
		switch err.Error() {
		case "forbidden":
			response.Error(w, http.StatusForbidden, "insufficient permissions")
		default:
			response.Error(w, http.StatusNotFound, "work order not found")
		}
		return
	}
	response.Success(w, wo)
}

// PUT /work-orders/{id}
func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	role, branchID, userID := ctxVars(r)

	var req UpdateWorkOrderRequest
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

	wo, err := h.service.Update(r.Context(), id, branchIDs, req)
	if err != nil {
		writeWorkOrderError(w, err, "failed to update work order")
		return
	}
	response.Success(w, wo)
}

// PATCH /work-orders/{id}/status
func (h *Handler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	role, branchID, userID := ctxVars(r)

	var req UpdateWorkOrderStatusRequest
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

	wo, err := h.service.UpdateStatus(r.Context(), id, branchIDs, req)
	if err != nil {
		writeWorkOrderError(w, err, "failed to update work order status")
		return
	}
	response.Success(w, wo)
}

// DELETE /work-orders/{id}
func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	role, branchID, userID := ctxVars(r)

	branchIDs, err := dbscope.GetEffectiveBranchIDs(r.Context(), h.db, role, userID, branchID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to resolve branch access")
		return
	}

	if err := h.service.Delete(r.Context(), id, branchIDs); err != nil {
		writeWorkOrderError(w, err, "failed to delete work order")
		return
	}
	response.Success(w, map[string]string{"message": "work order deleted"})
}

func (h *Handler) storeUpload(r *http.Request, fieldName, userID string) (string, error) {
	file, header, err := r.FormFile(fieldName)
	if err != nil {
		return "", errRequiredFile(fieldName)
	}
	defer file.Close()

	filename, err := randomImageFilename(fieldName, header)
	if err != nil {
		return "", err
	}
	if err := validateImageContent(file); err != nil {
		return "", err
	}

	destDir := filepath.Join(h.uploadDir, "work-order-assets", userID)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", err
	}

	dst, err := os.Create(filepath.Join(destDir, filename))
	if err != nil {
		return "", err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return "", err
	}

	return strings.TrimRight(h.baseURL, "/") + "/uploads/work-order-assets/" + userID + "/" + filename, nil
}

func randomImageFilename(prefix string, header *multipart.FileHeader) (string, error) {
	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowed := map[string]bool{".png": true, ".jpg": true, ".jpeg": true, ".webp": true}
	if !allowed[ext] {
		return "", errInvalidImage()
	}

	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return prefix + "-" + hex.EncodeToString(b) + ext, nil
}

func validateImageContent(file multipart.File) error {
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return err
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return err
	}

	contentType := http.DetectContentType(buffer[:n])
	allowed := map[string]bool{
		"image/png":  true,
		"image/jpeg": true,
		"image/webp": true,
	}
	if !allowed[contentType] {
		return errInvalidImage()
	}
	return nil
}

func optionalFormValue(r *http.Request, key string) *string {
	value := strings.TrimSpace(r.FormValue(key))
	if value == "" {
		return nil
	}
	return &value
}

func writeWorkOrderError(w http.ResponseWriter, err error, fallback string) {
	switch err.Error() {
	case "forbidden":
		response.Error(w, http.StatusForbidden, "insufficient permissions")
	case "work order not found":
		response.Error(w, http.StatusNotFound, err.Error())
	case "only draft work orders can be deleted":
		response.Error(w, http.StatusUnprocessableEntity, err.Error())
	case "invalid work_order_date: use YYYY-MM-DD", "company_name cannot be empty", "bill_to_name cannot be empty", "job_details cannot be empty", "item description is required", "duplicate item line_no":
		response.Error(w, http.StatusUnprocessableEntity, err.Error())
	default:
		response.Error(w, http.StatusInternalServerError, fallback)
	}
}

func writeUploadError(w http.ResponseWriter, err error) {
	var reqErr requestError
	if errors.As(err, &reqErr) {
		response.Error(w, http.StatusUnprocessableEntity, err.Error())
		return
	}
	response.Error(w, http.StatusInternalServerError, "failed to store file")
}

type requestError string

func (e requestError) Error() string {
	return string(e)
}

func errRequiredFile(fieldName string) error {
	return requestError(fieldName + " file is required")
}

func errInvalidImage() error {
	return requestError("only PNG, JPG, JPEG, and WEBP image files are allowed")
}
