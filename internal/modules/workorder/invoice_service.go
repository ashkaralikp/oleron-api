package workorder

import (
	"context"
	"errors"
	"strings"
	"time"

	"rmp-api/internal/dbscope"
	"rmp-api/internal/models"
)

const (
	defaultInvoiceTitle   = "EXPORT INVOICE"
	defaultInvoiceSeller  = "Oleron India"
	defaultSellerAddress  = "32/100A, Haseen, Moozhikkal,Kozhikode, Kerala:673571"
	defaultSellerEmail    = "info@oleron.in"
	defaultSellerPhone    = "9895427557"
	defaultSellerGSTIN    = "32AADFO2943D1ZQ"
	defaultInvoiceSACCode = "998399"
)

func (s *Service) CreateInvoice(ctx context.Context, workOrderID string, branchIDs []string, userID string, req CreateInvoiceRequest) (*models.WorkOrderInvoice, error) {
	wo, err := s.GetByID(ctx, workOrderID, branchIDs)
	if err != nil {
		return nil, err
	}
	if err := normalizeCreateInvoiceRequest(&req); err != nil {
		return nil, err
	}
	return s.repo.CreateInvoiceFromWorkOrder(ctx, wo, userID, req)
}

func (s *Service) ListInvoicesByWorkOrder(ctx context.Context, workOrderID string, branchIDs []string) ([]models.WorkOrderInvoice, error) {
	wo, err := s.GetByID(ctx, workOrderID, branchIDs)
	if err != nil {
		return nil, err
	}
	return s.repo.FindInvoicesByWorkOrderID(ctx, wo.ID)
}

func (s *Service) GetInvoiceByID(ctx context.Context, invoiceID string, branchIDs []string) (*models.WorkOrderInvoice, error) {
	invoice, err := s.repo.FindInvoiceByID(ctx, invoiceID)
	if err != nil {
		return nil, err
	}
	if invoice == nil {
		return nil, errors.New("invoice not found")
	}
	if !dbscope.ContainsBranch(branchIDs, invoice.BranchID) {
		return nil, errors.New("forbidden")
	}
	return invoice, nil
}

func (s *Service) UpdateInvoiceStatus(ctx context.Context, invoiceID string, branchIDs []string, req UpdateInvoiceStatusRequest) (*models.WorkOrderInvoice, error) {
	if _, err := s.GetInvoiceByID(ctx, invoiceID, branchIDs); err != nil {
		return nil, err
	}
	req.PDFURL = trimOptional(req.PDFURL)
	invoice, err := s.repo.UpdateInvoiceStatus(ctx, invoiceID, req.Status, req.PDFURL)
	if err != nil {
		return nil, err
	}
	if invoice == nil {
		return nil, errors.New("invoice not found")
	}
	return invoice, nil
}

func (s *Service) AddInvoicePayment(ctx context.Context, invoiceID string, branchIDs []string, userID string, req CreateInvoicePaymentRequest) (*models.WorkOrderInvoicePayment, error) {
	if _, err := s.GetInvoiceByID(ctx, invoiceID, branchIDs); err != nil {
		return nil, err
	}
	if err := normalizeCreateInvoicePaymentRequest(&req); err != nil {
		return nil, err
	}
	return s.repo.CreateInvoicePayment(ctx, invoiceID, userID, req)
}

func (s *Service) ListInvoicePayments(ctx context.Context, invoiceID string, branchIDs []string) ([]models.WorkOrderInvoicePayment, error) {
	if _, err := s.GetInvoiceByID(ctx, invoiceID, branchIDs); err != nil {
		return nil, err
	}
	return s.repo.FindPaymentsByInvoiceID(ctx, invoiceID)
}

func (s *Service) GetInvoicePaymentByID(ctx context.Context, paymentID string, branchIDs []string) (*models.WorkOrderInvoicePayment, error) {
	payment, err := s.repo.FindPaymentByID(ctx, paymentID)
	if err != nil {
		return nil, err
	}
	if payment == nil {
		return nil, errors.New("payment not found")
	}
	if _, err := s.GetInvoiceByID(ctx, payment.InvoiceID, branchIDs); err != nil {
		return nil, err
	}
	return payment, nil
}

func (s *Service) UpdateInvoicePaymentStatus(ctx context.Context, paymentID string, branchIDs []string, userID string, req UpdateInvoicePaymentStatusRequest) (*models.WorkOrderInvoicePayment, error) {
	if _, err := s.GetInvoicePaymentByID(ctx, paymentID, branchIDs); err != nil {
		return nil, err
	}
	req.Notes = trimOptional(req.Notes)
	return s.repo.UpdateInvoicePaymentStatus(ctx, paymentID, req.Status, req.Notes, userID)
}

func (s *Service) AddInvoicePaymentStatement(ctx context.Context, paymentID string, branchIDs []string, uploadedBy string, statementURL string, originalFilename, fileMIMEType *string, fileSizeBytes *int64, notes *string) (*models.WorkOrderInvoicePaymentStatement, error) {
	if _, err := s.GetInvoicePaymentByID(ctx, paymentID, branchIDs); err != nil {
		return nil, err
	}
	return s.repo.CreateInvoicePaymentStatement(ctx, paymentID, uploadedBy, statementURL, originalFilename, fileMIMEType, fileSizeBytes, trimOptional(notes))
}

func normalizeCreateInvoiceRequest(req *CreateInvoiceRequest) error {
	req.InvoiceDate = strings.TrimSpace(req.InvoiceDate)
	if req.InvoiceDate == "" {
		req.InvoiceDate = time.Now().Format("2006-01-02")
	}
	if _, err := time.Parse("2006-01-02", req.InvoiceDate); err != nil {
		return errors.New("invalid invoice_date: use YYYY-MM-DD")
	}
	if req.DueDate != nil {
		*req.DueDate = strings.TrimSpace(*req.DueDate)
		if *req.DueDate == "" {
			req.DueDate = nil
		} else if _, err := time.Parse("2006-01-02", *req.DueDate); err != nil {
			return errors.New("invalid due_date: use YYYY-MM-DD")
		}
	}

	req.InvoiceTitle = trimOptional(req.InvoiceTitle)
	req.SupplyNote = trimOptional(req.SupplyNote)
	req.SellerName = trimOptional(req.SellerName)
	req.SellerAddress = trimOptional(req.SellerAddress)
	req.SellerEmail = trimOptionalLower(req.SellerEmail)
	req.SellerPhone = trimOptional(req.SellerPhone)
	req.SellerGSTIN = trimOptional(req.SellerGSTIN)
	req.SellerLogoURL = trimOptional(req.SellerLogoURL)
	req.BillToName = trimOptional(req.BillToName)
	req.BillToAddress = trimOptional(req.BillToAddress)
	req.BillToEmail = trimOptionalLower(req.BillToEmail)
	req.BillToPhone = trimOptional(req.BillToPhone)
	req.BillToWebsite = trimOptional(req.BillToWebsite)
	req.LUTOrderNumber = trimOptional(req.LUTOrderNumber)
	req.ARNNumber = trimOptional(req.ARNNumber)
	req.Notes = trimOptional(req.Notes)
	req.SignerName = trimOptional(req.SignerName)
	req.SignerTitle = trimOptional(req.SignerTitle)
	req.SignatureURL = trimOptional(req.SignatureURL)
	req.SealURL = trimOptional(req.SealURL)
	req.PDFURL = trimOptional(req.PDFURL)

	if req.Currency != nil {
		currency := strings.ToUpper(strings.TrimSpace(*req.Currency))
		if currency == "" {
			req.Currency = nil
		} else {
			req.Currency = &currency
		}
	}
	if req.Status == "" {
		req.Status = "draft"
	}

	seen := map[int]bool{}
	for i := range req.Items {
		item := &req.Items[i]
		item.Description = strings.TrimSpace(item.Description)
		item.WorkOrderItemID = trimOptional(item.WorkOrderItemID)
		item.SACCode = trimOptional(item.SACCode)
		if item.Description == "" {
			return errors.New("invoice item description is required")
		}
		lineNo := item.LineNo
		if lineNo == 0 {
			lineNo = i + 1
		}
		if seen[lineNo] {
			return errors.New("duplicate invoice item line_no")
		}
		seen[lineNo] = true
	}

	return nil
}

func normalizeCreateInvoicePaymentRequest(req *CreateInvoicePaymentRequest) error {
	req.PaymentDate = strings.TrimSpace(req.PaymentDate)
	if req.PaymentDate == "" {
		req.PaymentDate = time.Now().Format("2006-01-02")
	}
	if _, err := time.Parse("2006-01-02", req.PaymentDate); err != nil {
		return errors.New("invalid payment_date: use YYYY-MM-DD")
	}
	req.Currency = strings.ToUpper(strings.TrimSpace(req.Currency))
	if req.Currency == "" {
		req.Currency = "USD"
	}
	req.Method = strings.ToLower(strings.TrimSpace(req.Method))
	req.OtherMethod = trimOptional(req.OtherMethod)
	req.ReferenceNo = trimOptional(req.ReferenceNo)
	req.PayerName = trimOptional(req.PayerName)
	req.PayerAccountLast4 = trimOptional(req.PayerAccountLast4)
	req.BankName = trimOptional(req.BankName)
	req.Notes = trimOptional(req.Notes)
	if req.Status == "" {
		req.Status = "pending"
	}
	if req.Method == "other" && req.OtherMethod == nil {
		return errors.New("other_method is required when method is other")
	}
	return nil
}
