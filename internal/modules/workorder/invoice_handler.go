package workorder

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"rmp-api/internal/dbscope"
	"rmp-api/pkg/response"
	"rmp-api/pkg/validator"

	"github.com/go-chi/chi/v5"
)

// POST /work-orders/{id}/invoices
func (h *Handler) CreateInvoice(w http.ResponseWriter, r *http.Request) {
	workOrderID := chi.URLParam(r, "id")
	role, branchID, userID := ctxVars(r)

	var req CreateInvoiceRequest
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

	invoice, err := h.service.CreateInvoice(r.Context(), workOrderID, branchIDs, userID, req)
	if err != nil {
		writeInvoiceError(w, err, "failed to create invoice")
		return
	}
	response.Created(w, invoice)
}

// GET /work-orders/{id}/invoices
func (h *Handler) ListInvoices(w http.ResponseWriter, r *http.Request) {
	workOrderID := chi.URLParam(r, "id")
	role, branchID, userID := ctxVars(r)

	branchIDs, err := dbscope.GetEffectiveBranchIDs(r.Context(), h.db, role, userID, branchID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to resolve branch access")
		return
	}

	invoices, err := h.service.ListInvoicesByWorkOrder(r.Context(), workOrderID, branchIDs)
	if err != nil {
		writeInvoiceError(w, err, "failed to fetch invoices")
		return
	}
	response.Success(w, invoices)
}

// GET /work-orders/invoices/{invoiceID}
func (h *Handler) GetInvoice(w http.ResponseWriter, r *http.Request) {
	invoiceID := chi.URLParam(r, "invoiceID")
	role, branchID, userID := ctxVars(r)

	branchIDs, err := dbscope.GetEffectiveBranchIDs(r.Context(), h.db, role, userID, branchID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to resolve branch access")
		return
	}

	invoice, err := h.service.GetInvoiceByID(r.Context(), invoiceID, branchIDs)
	if err != nil {
		writeInvoiceError(w, err, "failed to fetch invoice")
		return
	}
	response.Success(w, invoice)
}

// PATCH /work-orders/invoices/{invoiceID}/status
func (h *Handler) UpdateInvoiceStatus(w http.ResponseWriter, r *http.Request) {
	invoiceID := chi.URLParam(r, "invoiceID")
	role, branchID, userID := ctxVars(r)

	var req UpdateInvoiceStatusRequest
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

	invoice, err := h.service.UpdateInvoiceStatus(r.Context(), invoiceID, branchIDs, req)
	if err != nil {
		writeInvoiceError(w, err, "failed to update invoice status")
		return
	}
	response.Success(w, invoice)
}

// POST /work-orders/invoices/{invoiceID}/payments
func (h *Handler) AddInvoicePayment(w http.ResponseWriter, r *http.Request) {
	invoiceID := chi.URLParam(r, "invoiceID")
	role, branchID, userID := ctxVars(r)

	var req CreateInvoicePaymentRequest
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

	payment, err := h.service.AddInvoicePayment(r.Context(), invoiceID, branchIDs, userID, req)
	if err != nil {
		writeInvoiceError(w, err, "failed to record invoice payment")
		return
	}
	response.Created(w, payment)
}

// GET /work-orders/invoices/{invoiceID}/payments
func (h *Handler) ListInvoicePayments(w http.ResponseWriter, r *http.Request) {
	invoiceID := chi.URLParam(r, "invoiceID")
	role, branchID, userID := ctxVars(r)

	branchIDs, err := dbscope.GetEffectiveBranchIDs(r.Context(), h.db, role, userID, branchID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to resolve branch access")
		return
	}

	payments, err := h.service.ListInvoicePayments(r.Context(), invoiceID, branchIDs)
	if err != nil {
		writeInvoiceError(w, err, "failed to fetch invoice payments")
		return
	}
	response.Success(w, payments)
}

// PATCH /work-orders/invoice-payments/{paymentID}/status
func (h *Handler) UpdateInvoicePaymentStatus(w http.ResponseWriter, r *http.Request) {
	paymentID := chi.URLParam(r, "paymentID")
	role, branchID, userID := ctxVars(r)

	var req UpdateInvoicePaymentStatusRequest
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

	payment, err := h.service.UpdateInvoicePaymentStatus(r.Context(), paymentID, branchIDs, userID, req)
	if err != nil {
		writeInvoiceError(w, err, "failed to update invoice payment status")
		return
	}
	response.Success(w, payment)
}

// POST /work-orders/invoice-payments/{paymentID}/statements
func (h *Handler) UploadInvoicePaymentStatement(w http.ResponseWriter, r *http.Request) {
	paymentID := chi.URLParam(r, "paymentID")
	role, branchID, userID := ctxVars(r)

	const maxSize = 10 << 20
	r.Body = http.MaxBytesReader(w, r.Body, maxSize)
	if err := r.ParseMultipartForm(maxSize); err != nil {
		response.Error(w, http.StatusUnprocessableEntity, "file too large: max 10MB")
		return
	}

	branchIDs, err := dbscope.GetEffectiveBranchIDs(r.Context(), h.db, role, userID, branchID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "failed to resolve branch access")
		return
	}
	if _, err := h.service.GetInvoicePaymentByID(r.Context(), paymentID, branchIDs); err != nil {
		writeInvoiceError(w, err, "failed to fetch invoice payment")
		return
	}

	statementURL, originalFilename, mimeType, sizeBytes, err := h.storeStatementUpload(r, paymentID)
	if err != nil {
		writeUploadError(w, err)
		return
	}
	notes := optionalFormValue(r, "notes")

	statement, err := h.service.AddInvoicePaymentStatement(r.Context(), paymentID, branchIDs, userID, statementURL, originalFilename, mimeType, sizeBytes, notes)
	if err != nil {
		writeInvoiceError(w, err, "failed to save payment statement")
		return
	}
	response.Created(w, statement)
}

func (h *Handler) storeStatementUpload(r *http.Request, paymentID string) (string, *string, *string, *int64, error) {
	file, header, err := r.FormFile("statement")
	if err != nil {
		return "", nil, nil, nil, errRequiredFile("statement")
	}
	defer file.Close()

	filename, err := randomStatementFilename("statement", header)
	if err != nil {
		return "", nil, nil, nil, err
	}
	mimeType, err := validateStatementContent(file)
	if err != nil {
		return "", nil, nil, nil, err
	}

	destDir := filepath.Join(h.uploadDir, "invoice-payment-statements", paymentID)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", nil, nil, nil, err
	}

	dst, err := os.Create(filepath.Join(destDir, filename))
	if err != nil {
		return "", nil, nil, nil, err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		return "", nil, nil, nil, err
	}

	originalFilename := header.Filename
	sizeBytes := header.Size
	url := strings.TrimRight(h.baseURL, "/") + "/uploads/invoice-payment-statements/" + paymentID + "/" + filename
	return url, &originalFilename, &mimeType, &sizeBytes, nil
}

func randomStatementFilename(prefix string, header *multipart.FileHeader) (string, error) {
	ext := strings.ToLower(filepath.Ext(header.Filename))
	allowed := map[string]bool{".pdf": true, ".png": true, ".jpg": true, ".jpeg": true, ".webp": true}
	if !allowed[ext] {
		return "", errInvalidStatement()
	}

	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return prefix + "-" + hex.EncodeToString(b) + ext, nil
}

func validateStatementContent(file multipart.File) (string, error) {
	buffer := make([]byte, 512)
	n, err := file.Read(buffer)
	if err != nil && err != io.EOF {
		return "", err
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return "", err
	}

	contentType := http.DetectContentType(buffer[:n])
	allowed := map[string]bool{
		"application/pdf": true,
		"image/png":       true,
		"image/jpeg":      true,
		"image/webp":      true,
	}
	if !allowed[contentType] {
		return "", errInvalidStatement()
	}
	return contentType, nil
}

func writeInvoiceError(w http.ResponseWriter, err error, fallback string) {
	switch err.Error() {
	case "forbidden":
		response.Error(w, http.StatusForbidden, "insufficient permissions")
	case "work order not found", "invoice not found", "payment not found":
		response.Error(w, http.StatusNotFound, err.Error())
	case "invalid invoice_date: use YYYY-MM-DD", "invalid due_date: use YYYY-MM-DD", "invalid payment_date: use YYYY-MM-DD",
		"invoice item description is required", "duplicate invoice item line_no", "other_method is required when method is other":
		response.Error(w, http.StatusUnprocessableEntity, err.Error())
	default:
		response.Error(w, http.StatusInternalServerError, fallback)
	}
}

func errInvalidStatement() error {
	return requestError("only PDF, PNG, JPG, JPEG, and WEBP statement files are allowed")
}
