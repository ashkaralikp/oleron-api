package workorder

import (
	"context"
	"errors"
	"fmt"
	"time"

	"rmp-api/internal/models"

	"github.com/jackc/pgx/v5"
)

func (r *Repository) CreateInvoiceFromWorkOrder(ctx context.Context, wo *models.WorkOrder, createdBy string, req CreateInvoiceRequest) (*models.WorkOrderInvoice, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	invoiceNo, err := generateInvoiceNo(ctx, tx, wo.BranchID, req.InvoiceDate)
	if err != nil {
		return nil, err
	}

	items := buildInvoiceItems(wo, req.Items)
	grossAmount, itemTaxAmount := calculateInvoiceItemTotals(items)
	taxAmount := itemTaxAmount
	if req.TaxAmount != nil {
		taxAmount = *req.TaxAmount
	}
	additionalAmount := 0.0
	if req.AdditionalAmount != nil {
		additionalAmount = *req.AdditionalAmount
	}
	totalAmount := grossAmount + taxAmount + additionalAmount

	invoiceTitle := stringValue(req.InvoiceTitle, defaultInvoiceTitle)
	sellerName := stringValue(req.SellerName, defaultInvoiceSeller)
	sellerAddress := stringPtrValue(req.SellerAddress, defaultSellerAddress)
	sellerEmail := stringPtrValue(req.SellerEmail, defaultSellerEmail)
	sellerPhone := stringPtrValue(req.SellerPhone, defaultSellerPhone)
	sellerGSTIN := stringPtrValue(req.SellerGSTIN, defaultSellerGSTIN)
	billToName := stringValue(req.BillToName, wo.CompanyName)
	billToAddress := req.BillToAddress
	if billToAddress == nil {
		billToAddress = wo.CompanyAddress
	}
	billToEmail := req.BillToEmail
	if billToEmail == nil {
		billToEmail = wo.CompanyEmail
	}
	billToPhone := req.BillToPhone
	if billToPhone == nil {
		billToPhone = wo.CompanyPhone
	}
	billToWebsite := req.BillToWebsite
	if billToWebsite == nil {
		billToWebsite = wo.CompanyWebsite
	}
	currency := stringValue(req.Currency, wo.Currency)
	if currency == "" {
		currency = "USD"
	}

	var invoice models.WorkOrderInvoice
	err = tx.QueryRow(ctx,
		`INSERT INTO work_order_invoices
		 (work_order_id, branch_id, created_by, invoice_no, invoice_date, due_date,
		  invoice_title, supply_note,
		  seller_name, seller_address, seller_email, seller_phone, seller_gstin, seller_logo_url,
		  bill_to_name, bill_to_address, bill_to_email, bill_to_phone, bill_to_website,
		  currency, gross_amount, tax_amount, additional_amount, total_amount,
		  status, lut_order_number, arn_number, notes,
		  signer_name, signer_title, signature_url, seal_url, pdf_url, issued_at)
		 VALUES
		 ($1, $2, $3, $4, $5, $6,
		  $7, $8,
		  $9, $10, $11, $12, $13, $14,
		  $15, $16, $17, $18, $19,
		  $20, $21, $22, $23, $24,
		  $25::invoice_status, $26, $27, $28,
		  $29, $30, $31, $32, $33,
		  CASE WHEN $25::invoice_status = 'issued' THEN NOW() ELSE NULL END)
		 RETURNING id, work_order_id, branch_id, created_by, invoice_no, invoice_date, due_date,
		           invoice_title, supply_note,
		           seller_name, seller_address, seller_email, seller_phone, seller_gstin, seller_logo_url,
		           bill_to_name, bill_to_address, bill_to_email, bill_to_phone, bill_to_website,
		           currency, gross_amount, tax_amount, additional_amount, total_amount,
		           paid_amount, balance_amount, status, payment_status,
		           lut_order_number, arn_number, notes,
		           signer_name, signer_title, signature_url, seal_url, pdf_url,
		           issued_at, cancelled_at, created_at, updated_at`,
		wo.ID, wo.BranchID, createdBy, invoiceNo, req.InvoiceDate, req.DueDate,
		invoiceTitle, req.SupplyNote,
		sellerName, sellerAddress, sellerEmail, sellerPhone, sellerGSTIN, req.SellerLogoURL,
		billToName, billToAddress, billToEmail, billToPhone, billToWebsite,
		currency, grossAmount, taxAmount, additionalAmount, totalAmount,
		req.Status, req.LUTOrderNumber, req.ARNNumber, req.Notes,
		req.SignerName, req.SignerTitle, req.SignatureURL, req.SealURL, req.PDFURL,
	).Scan(invoiceScanDest(&invoice)...)
	if err != nil {
		return nil, err
	}

	invoice.Items, err = insertInvoiceItems(ctx, tx, invoice.ID, items)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &invoice, nil
}

func (r *Repository) FindInvoicesByWorkOrderID(ctx context.Context, workOrderID string) ([]models.WorkOrderInvoice, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, work_order_id, branch_id, created_by, invoice_no, invoice_date, due_date,
		        invoice_title, supply_note,
		        seller_name, seller_address, seller_email, seller_phone, seller_gstin, seller_logo_url,
		        bill_to_name, bill_to_address, bill_to_email, bill_to_phone, bill_to_website,
		        currency, gross_amount, tax_amount, additional_amount, total_amount,
		        paid_amount, balance_amount, status, payment_status,
		        lut_order_number, arn_number, notes,
		        signer_name, signer_title, signature_url, seal_url, pdf_url,
		        issued_at, cancelled_at, created_at, updated_at
		 FROM work_order_invoices
		 WHERE work_order_id = $1
		 ORDER BY invoice_date DESC, created_at DESC`, workOrderID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var invoices []models.WorkOrderInvoice
	for rows.Next() {
		var invoice models.WorkOrderInvoice
		if err := rows.Scan(invoiceScanDest(&invoice)...); err != nil {
			return nil, err
		}
		invoices = append(invoices, invoice)
	}
	return invoices, rows.Err()
}

func (r *Repository) FindInvoiceByID(ctx context.Context, invoiceID string) (*models.WorkOrderInvoice, error) {
	var invoice models.WorkOrderInvoice
	err := r.db.QueryRow(ctx,
		`SELECT id, work_order_id, branch_id, created_by, invoice_no, invoice_date, due_date,
		        invoice_title, supply_note,
		        seller_name, seller_address, seller_email, seller_phone, seller_gstin, seller_logo_url,
		        bill_to_name, bill_to_address, bill_to_email, bill_to_phone, bill_to_website,
		        currency, gross_amount, tax_amount, additional_amount, total_amount,
		        paid_amount, balance_amount, status, payment_status,
		        lut_order_number, arn_number, notes,
		        signer_name, signer_title, signature_url, seal_url, pdf_url,
		        issued_at, cancelled_at, created_at, updated_at
		 FROM work_order_invoices
		 WHERE id = $1`, invoiceID,
	).Scan(invoiceScanDest(&invoice)...)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	items, err := r.FindInvoiceItemsByInvoiceID(ctx, invoiceID)
	if err != nil {
		return nil, err
	}
	invoice.Items = items
	return &invoice, nil
}

func (r *Repository) FindInvoiceItemsByInvoiceID(ctx context.Context, invoiceID string) ([]models.WorkOrderInvoiceItem, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, invoice_id, work_order_item_id, line_no, description, quantity, sac_code,
		        unit_amount, tax_amount, total_amount, created_at, updated_at
		 FROM work_order_invoice_items
		 WHERE invoice_id = $1
		 ORDER BY line_no ASC`, invoiceID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.WorkOrderInvoiceItem
	for rows.Next() {
		var item models.WorkOrderInvoiceItem
		if err := rows.Scan(invoiceItemScanDest(&item)...); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) UpdateInvoiceStatus(ctx context.Context, invoiceID, status string, pdfURL *string) (*models.WorkOrderInvoice, error) {
	var invoice models.WorkOrderInvoice
	err := r.db.QueryRow(ctx,
		`UPDATE work_order_invoices SET
		   status = $2::invoice_status,
		   pdf_url = COALESCE($3, pdf_url),
		   issued_at = CASE
		     WHEN $2::invoice_status = 'issued' AND issued_at IS NULL THEN NOW()
		     WHEN $2::invoice_status <> 'issued' THEN NULL
		     ELSE issued_at
		   END,
		   cancelled_at = CASE
		     WHEN $2::invoice_status IN ('cancelled', 'void') AND cancelled_at IS NULL THEN NOW()
		     WHEN $2::invoice_status NOT IN ('cancelled', 'void') THEN NULL
		     ELSE cancelled_at
		   END,
		   updated_at = NOW()
		 WHERE id = $1
		 RETURNING id, work_order_id, branch_id, created_by, invoice_no, invoice_date, due_date,
		           invoice_title, supply_note,
		           seller_name, seller_address, seller_email, seller_phone, seller_gstin, seller_logo_url,
		           bill_to_name, bill_to_address, bill_to_email, bill_to_phone, bill_to_website,
		           currency, gross_amount, tax_amount, additional_amount, total_amount,
		           paid_amount, balance_amount, status, payment_status,
		           lut_order_number, arn_number, notes,
		           signer_name, signer_title, signature_url, seal_url, pdf_url,
		           issued_at, cancelled_at, created_at, updated_at`,
		invoiceID, status, pdfURL,
	).Scan(invoiceScanDest(&invoice)...)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	invoice.Items, err = r.FindInvoiceItemsByInvoiceID(ctx, invoiceID)
	if err != nil {
		return nil, err
	}
	return &invoice, nil
}

func (r *Repository) CreateInvoicePayment(ctx context.Context, invoiceID, recordedBy string, req CreateInvoicePaymentRequest) (*models.WorkOrderInvoicePayment, error) {
	var payment models.WorkOrderInvoicePayment
	err := r.db.QueryRow(ctx,
		`INSERT INTO work_order_invoice_payments
		 (invoice_id, recorded_by, payment_date, amount, currency, method, other_method,
		  reference_no, payer_name, payer_account_last4, bank_name, status, notes,
		  verified_by, verified_at)
		 VALUES
		 ($1, $2, $3, $4, $5, $6::invoice_payment_method, $7,
		  $8, $9, $10, $11, $12::invoice_payment_record_status, $13,
		  CASE WHEN $12::invoice_payment_record_status IN ('confirmed', 'rejected') THEN $2 ELSE NULL END,
		  CASE WHEN $12::invoice_payment_record_status IN ('confirmed', 'rejected') THEN NOW() ELSE NULL END)
		 RETURNING id, invoice_id, recorded_by, payment_date, amount, currency, method, other_method,
		           reference_no, payer_name, payer_account_last4, bank_name, status, notes,
		           verified_by, verified_at, created_at, updated_at`,
		invoiceID, recordedBy, req.PaymentDate, req.Amount, req.Currency, req.Method, req.OtherMethod,
		req.ReferenceNo, req.PayerName, req.PayerAccountLast4, req.BankName, req.Status, req.Notes,
	).Scan(paymentScanDest(&payment)...)
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *Repository) FindPaymentsByInvoiceID(ctx context.Context, invoiceID string) ([]models.WorkOrderInvoicePayment, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, invoice_id, recorded_by, payment_date, amount, currency, method, other_method,
		        reference_no, payer_name, payer_account_last4, bank_name, status, notes,
		        verified_by, verified_at, created_at, updated_at
		 FROM work_order_invoice_payments
		 WHERE invoice_id = $1
		 ORDER BY payment_date DESC, created_at DESC`, invoiceID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []models.WorkOrderInvoicePayment
	for rows.Next() {
		var payment models.WorkOrderInvoicePayment
		if err := rows.Scan(paymentScanDest(&payment)...); err != nil {
			return nil, err
		}
		payment.Statements, err = r.FindStatementsByPaymentID(ctx, payment.ID)
		if err != nil {
			return nil, err
		}
		payments = append(payments, payment)
	}
	return payments, rows.Err()
}

func (r *Repository) FindPaymentByID(ctx context.Context, paymentID string) (*models.WorkOrderInvoicePayment, error) {
	var payment models.WorkOrderInvoicePayment
	err := r.db.QueryRow(ctx,
		`SELECT id, invoice_id, recorded_by, payment_date, amount, currency, method, other_method,
		        reference_no, payer_name, payer_account_last4, bank_name, status, notes,
		        verified_by, verified_at, created_at, updated_at
		 FROM work_order_invoice_payments
		 WHERE id = $1`, paymentID,
	).Scan(paymentScanDest(&payment)...)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	payment.Statements, err = r.FindStatementsByPaymentID(ctx, paymentID)
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *Repository) UpdateInvoicePaymentStatus(ctx context.Context, paymentID, status string, notes *string, verifiedBy string) (*models.WorkOrderInvoicePayment, error) {
	var payment models.WorkOrderInvoicePayment
	err := r.db.QueryRow(ctx,
		`UPDATE work_order_invoice_payments SET
		   status = $2::invoice_payment_record_status,
		   notes = COALESCE($3, notes),
		   verified_by = CASE
		     WHEN $2::invoice_payment_record_status IN ('confirmed', 'failed', 'reversed', 'rejected') THEN $4
		     WHEN $2::invoice_payment_record_status = 'pending' THEN NULL
		     ELSE verified_by
		   END,
		   verified_at = CASE
		     WHEN $2::invoice_payment_record_status IN ('confirmed', 'failed', 'reversed', 'rejected') THEN NOW()
		     WHEN $2::invoice_payment_record_status = 'pending' THEN NULL
		     ELSE verified_at
		   END,
		   updated_at = NOW()
		 WHERE id = $1
		 RETURNING id, invoice_id, recorded_by, payment_date, amount, currency, method, other_method,
		           reference_no, payer_name, payer_account_last4, bank_name, status, notes,
		           verified_by, verified_at, created_at, updated_at`,
		paymentID, status, notes, verifiedBy,
	).Scan(paymentScanDest(&payment)...)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, errors.New("payment not found")
	}
	if err != nil {
		return nil, err
	}
	payment.Statements, err = r.FindStatementsByPaymentID(ctx, paymentID)
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *Repository) CreateInvoicePaymentStatement(ctx context.Context, paymentID, uploadedBy, statementURL string, originalFilename, fileMIMEType *string, fileSizeBytes *int64, notes *string) (*models.WorkOrderInvoicePaymentStatement, error) {
	var statement models.WorkOrderInvoicePaymentStatement
	err := r.db.QueryRow(ctx,
		`INSERT INTO work_order_invoice_payment_statements
		 (payment_id, uploaded_by, statement_url, original_filename, file_mime_type, file_size_bytes, notes)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, payment_id, uploaded_by, statement_url, original_filename, file_mime_type, file_size_bytes, notes, created_at`,
		paymentID, uploadedBy, statementURL, originalFilename, fileMIMEType, fileSizeBytes, notes,
	).Scan(statementScanDest(&statement)...)
	if err != nil {
		return nil, err
	}
	return &statement, nil
}

func (r *Repository) FindStatementsByPaymentID(ctx context.Context, paymentID string) ([]models.WorkOrderInvoicePaymentStatement, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, payment_id, uploaded_by, statement_url, original_filename, file_mime_type, file_size_bytes, notes, created_at
		 FROM work_order_invoice_payment_statements
		 WHERE payment_id = $1
		 ORDER BY created_at DESC`, paymentID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var statements []models.WorkOrderInvoicePaymentStatement
	for rows.Next() {
		var statement models.WorkOrderInvoicePaymentStatement
		if err := rows.Scan(statementScanDest(&statement)...); err != nil {
			return nil, err
		}
		statements = append(statements, statement)
	}
	return statements, rows.Err()
}

type invoiceItemInsert struct {
	WorkOrderItemID *string
	LineNo          int
	Description     string
	Quantity        float64
	SACCode         *string
	UnitAmount      float64
	TaxAmount       float64
	TotalAmount     float64
}

func buildInvoiceItems(wo *models.WorkOrder, reqItems []CreateInvoiceItemRequest) []invoiceItemInsert {
	if len(reqItems) == 0 {
		items := make([]invoiceItemInsert, 0, len(wo.Items))
		sacCode := defaultInvoiceSACCode
		for _, item := range wo.Items {
			workOrderItemID := item.ID
			items = append(items, invoiceItemInsert{
				WorkOrderItemID: &workOrderItemID,
				LineNo:          item.LineNo,
				Description:     item.Description,
				Quantity:        1,
				SACCode:         &sacCode,
				UnitAmount:      item.Amount,
				TotalAmount:     item.Amount,
			})
		}
		return items
	}

	items := make([]invoiceItemInsert, 0, len(reqItems))
	for i, req := range reqItems {
		lineNo := req.LineNo
		if lineNo == 0 {
			lineNo = i + 1
		}
		quantity := 1.0
		if req.Quantity != nil {
			quantity = *req.Quantity
		}
		unitAmount := 0.0
		if req.UnitAmount != nil {
			unitAmount = *req.UnitAmount
		}
		taxAmount := 0.0
		if req.TaxAmount != nil {
			taxAmount = *req.TaxAmount
		}
		totalAmount := quantity * unitAmount
		if req.TotalAmount != nil {
			totalAmount = *req.TotalAmount
		}

		items = append(items, invoiceItemInsert{
			WorkOrderItemID: req.WorkOrderItemID,
			LineNo:          lineNo,
			Description:     req.Description,
			Quantity:        quantity,
			SACCode:         req.SACCode,
			UnitAmount:      unitAmount,
			TaxAmount:       taxAmount,
			TotalAmount:     totalAmount,
		})
	}
	return items
}

func insertInvoiceItems(ctx context.Context, tx pgx.Tx, invoiceID string, items []invoiceItemInsert) ([]models.WorkOrderInvoiceItem, error) {
	result := make([]models.WorkOrderInvoiceItem, 0, len(items))
	for _, req := range items {
		var item models.WorkOrderInvoiceItem
		err := tx.QueryRow(ctx,
			`INSERT INTO work_order_invoice_items
			 (invoice_id, work_order_item_id, line_no, description, quantity, sac_code, unit_amount, tax_amount, total_amount)
			 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
			 RETURNING id, invoice_id, work_order_item_id, line_no, description, quantity, sac_code,
			           unit_amount, tax_amount, total_amount, created_at, updated_at`,
			invoiceID, req.WorkOrderItemID, req.LineNo, req.Description, req.Quantity, req.SACCode, req.UnitAmount, req.TaxAmount, req.TotalAmount,
		).Scan(invoiceItemScanDest(&item)...)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, nil
}

func generateInvoiceNo(ctx context.Context, tx pgx.Tx, branchID, invoiceDate string) (string, error) {
	parsedDate, err := time.Parse("2006-01-02", invoiceDate)
	if err != nil {
		return "", err
	}

	fyStart := parsedDate.Year()
	if parsedDate.Month() < time.April {
		fyStart--
	}
	fyEnd := fyStart + 1
	suffix := fmt.Sprintf("%d-%02d", fyStart, fyEnd%100)
	pattern := fmt.Sprintf("^([0-9]+)/%s$", suffix)

	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtext($1))`, branchID+":invoice:"+suffix); err != nil {
		return "", err
	}

	var nextNumber int
	err = tx.QueryRow(ctx,
		`SELECT COALESCE(MAX(substring(invoice_no FROM $2)::INT), 0) + 1
		 FROM work_order_invoices
		 WHERE branch_id = $1
		   AND invoice_no ~ $2`,
		branchID, pattern,
	).Scan(&nextNumber)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%02d/%s", nextNumber, suffix), nil
}

func calculateInvoiceItemTotals(items []invoiceItemInsert) (grossAmount, taxAmount float64) {
	for _, item := range items {
		grossAmount += item.TotalAmount
		taxAmount += item.TaxAmount
	}
	return grossAmount, taxAmount
}

func invoiceScanDest(invoice *models.WorkOrderInvoice) []any {
	return []any{
		&invoice.ID, &invoice.WorkOrderID, &invoice.BranchID, &invoice.CreatedBy, &invoice.InvoiceNo, &invoice.InvoiceDate, &invoice.DueDate,
		&invoice.InvoiceTitle, &invoice.SupplyNote,
		&invoice.SellerName, &invoice.SellerAddress, &invoice.SellerEmail, &invoice.SellerPhone, &invoice.SellerGSTIN, &invoice.SellerLogoURL,
		&invoice.BillToName, &invoice.BillToAddress, &invoice.BillToEmail, &invoice.BillToPhone, &invoice.BillToWebsite,
		&invoice.Currency, &invoice.GrossAmount, &invoice.TaxAmount, &invoice.AdditionalAmount, &invoice.TotalAmount,
		&invoice.PaidAmount, &invoice.BalanceAmount, &invoice.Status, &invoice.PaymentStatus,
		&invoice.LUTOrderNumber, &invoice.ARNNumber, &invoice.Notes,
		&invoice.SignerName, &invoice.SignerTitle, &invoice.SignatureURL, &invoice.SealURL, &invoice.PDFURL,
		&invoice.IssuedAt, &invoice.CancelledAt, &invoice.CreatedAt, &invoice.UpdatedAt,
	}
}

func invoiceItemScanDest(item *models.WorkOrderInvoiceItem) []any {
	return []any{
		&item.ID, &item.InvoiceID, &item.WorkOrderItemID, &item.LineNo, &item.Description, &item.Quantity, &item.SACCode,
		&item.UnitAmount, &item.TaxAmount, &item.TotalAmount, &item.CreatedAt, &item.UpdatedAt,
	}
}

func paymentScanDest(payment *models.WorkOrderInvoicePayment) []any {
	return []any{
		&payment.ID, &payment.InvoiceID, &payment.RecordedBy, &payment.PaymentDate, &payment.Amount, &payment.Currency, &payment.Method, &payment.OtherMethod,
		&payment.ReferenceNo, &payment.PayerName, &payment.PayerAccountLast4, &payment.BankName, &payment.Status, &payment.Notes,
		&payment.VerifiedBy, &payment.VerifiedAt, &payment.CreatedAt, &payment.UpdatedAt,
	}
}

func statementScanDest(statement *models.WorkOrderInvoicePaymentStatement) []any {
	return []any{
		&statement.ID, &statement.PaymentID, &statement.UploadedBy, &statement.StatementURL, &statement.OriginalFilename,
		&statement.FileMIMEType, &statement.FileSizeBytes, &statement.Notes, &statement.CreatedAt,
	}
}

func stringValue(value *string, fallback string) string {
	if value == nil {
		return fallback
	}
	return *value
}

func stringPtrValue(value *string, fallback string) *string {
	if value != nil {
		return value
	}
	return &fallback
}
