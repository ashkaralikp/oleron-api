package workorder

import (
	"context"
	"errors"
	"fmt"
	"time"

	"rmp-api/internal/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) UpsertAsset(ctx context.Context, regionalManagerID, signerName string, signerTitle *string, signatureURL, sealURL string, uploadedBy *string) (*models.RegionalManagerWorkOrderAsset, error) {
	var asset models.RegionalManagerWorkOrderAsset
	err := r.db.QueryRow(ctx,
		`INSERT INTO regional_manager_work_order_assets
		 (regional_manager_id, signer_name, signer_title, signature_url, seal_url, uploaded_by)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (regional_manager_id) DO UPDATE SET
		   signer_name = EXCLUDED.signer_name,
		   signer_title = EXCLUDED.signer_title,
		   signature_url = EXCLUDED.signature_url,
		   seal_url = EXCLUDED.seal_url,
		   uploaded_by = EXCLUDED.uploaded_by,
		   updated_at = NOW()
		 RETURNING id, regional_manager_id, signer_name, signer_title, signature_url, seal_url, uploaded_by, created_at, updated_at`,
		regionalManagerID, signerName, signerTitle, signatureURL, sealURL, uploadedBy,
	).Scan(
		&asset.ID, &asset.RegionalManagerID, &asset.SignerName, &asset.SignerTitle,
		&asset.SignatureURL, &asset.SealURL, &asset.UploadedBy, &asset.CreatedAt, &asset.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &asset, nil
}

func (r *Repository) FindAssetByManagerID(ctx context.Context, regionalManagerID string) (*models.RegionalManagerWorkOrderAsset, error) {
	var asset models.RegionalManagerWorkOrderAsset
	err := r.db.QueryRow(ctx,
		`SELECT id, regional_manager_id, signer_name, signer_title, signature_url, seal_url, uploaded_by, created_at, updated_at
		 FROM regional_manager_work_order_assets
		 WHERE regional_manager_id = $1`, regionalManagerID,
	).Scan(
		&asset.ID, &asset.RegionalManagerID, &asset.SignerName, &asset.SignerTitle,
		&asset.SignatureURL, &asset.SealURL, &asset.UploadedBy, &asset.CreatedAt, &asset.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &asset, nil
}

func (r *Repository) FindAssetByID(ctx context.Context, id string) (*models.RegionalManagerWorkOrderAsset, error) {
	var asset models.RegionalManagerWorkOrderAsset
	err := r.db.QueryRow(ctx,
		`SELECT id, regional_manager_id, signer_name, signer_title, signature_url, seal_url, uploaded_by, created_at, updated_at
		 FROM regional_manager_work_order_assets
		 WHERE id = $1`, id,
	).Scan(
		&asset.ID, &asset.RegionalManagerID, &asset.SignerName, &asset.SignerTitle,
		&asset.SignatureURL, &asset.SealURL, &asset.UploadedBy, &asset.CreatedAt, &asset.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &asset, nil
}

func (r *Repository) CreateWorkOrder(ctx context.Context, branchID, createdBy string, asset *models.RegionalManagerWorkOrderAsset, req CreateWorkOrderRequest) (*models.WorkOrder, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	total := calculateTotal(req.Items)
	subTotal := total
	if req.SubTotalAmount != nil {
		subTotal = *req.SubTotalAmount
	}
	if req.TotalAmount != nil {
		total = *req.TotalAmount
	}

	companyName := req.CompanyName
	if companyName == "" {
		companyName = "Oleron.Inc"
	}
	currency := req.Currency
	if currency == "" {
		currency = "USD"
	}
	status := req.Status
	if status == "" {
		status = "draft"
	}
	workOrderNo, err := generateWorkOrderNo(ctx, tx, branchID, req.WorkOrderDate)
	if err != nil {
		return nil, err
	}

	var managerAssetID *string
	var signerName, signerTitle, signatureURL, sealURL *string
	if asset != nil {
		managerAssetID = &asset.ID
		signerName = &asset.SignerName
		signerTitle = asset.SignerTitle
		signatureURL = &asset.SignatureURL
		sealURL = &asset.SealURL
	}

	var wo models.WorkOrder
	err = tx.QueryRow(ctx,
		`INSERT INTO work_orders
		 (branch_id, created_by, manager_asset_id, work_order_no, work_order_date,
			  company_name, company_address, company_phone, company_fax, company_email, company_website, company_logo_url,
			  bill_to_name, bill_to_address, bill_to_email, job_details,
			  signer_name, signer_title, signature_url, seal_url, currency, sub_total_amount, total_amount, status, issued_at)
			 VALUES
			 ($1, $2, $3, $4, $5,
				  $6, $7, $8, $9, $10, $11, $12,
				  $13, $14, $15, $16,
				  $17, $18, $19, $20, $21, $22, $23, $24::work_order_status,
				  CASE WHEN $24::work_order_status = 'issued' THEN NOW() ELSE NULL END)
			 RETURNING id, branch_id, created_by, manager_asset_id, work_order_no, work_order_date,
			           company_name, company_address, company_phone, company_fax, company_email, company_website, company_logo_url,
			           bill_to_name, bill_to_address, bill_to_email, job_details,
			           signer_name, signer_title, signature_url, seal_url, currency, sub_total_amount, total_amount,
			           status, pdf_url, issued_at, created_at, updated_at`,
		branchID, createdBy, managerAssetID, workOrderNo, req.WorkOrderDate,
		companyName, req.CompanyAddress, req.CompanyPhone, req.CompanyFax, req.CompanyEmail, req.CompanyWebsite, req.CompanyLogoURL,
		req.BillToName, req.BillToAddress, req.BillToEmail, req.JobDetails,
		signerName, signerTitle, signatureURL, sealURL, currency, subTotal, total, status,
	).Scan(workOrderScanDest(&wo)...)
	if err != nil {
		return nil, err
	}

	items, err := insertItems(ctx, tx, wo.ID, req.Items)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	wo.Items = items
	return &wo, nil
}

func (r *Repository) FindAllWorkOrders(ctx context.Context, branchIDs []string) ([]models.WorkOrder, error) {
	query := `SELECT id, branch_id, created_by, manager_asset_id, work_order_no, work_order_date,
		                 company_name, company_address, company_phone, company_fax, company_email, company_website, company_logo_url,
		                 bill_to_name, bill_to_address, bill_to_email, job_details,
		                 signer_name, signer_title, signature_url, seal_url, currency, sub_total_amount, total_amount,
	                 status, pdf_url, issued_at, created_at, updated_at
	          FROM work_orders`
	args := []any{}
	if len(branchIDs) > 0 {
		query += ` WHERE branch_id::text = ANY($1)`
		args = append(args, branchIDs)
	}
	query += ` ORDER BY work_order_date DESC, created_at DESC`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.WorkOrder
	for rows.Next() {
		var wo models.WorkOrder
		if err := rows.Scan(workOrderScanDest(&wo)...); err != nil {
			return nil, err
		}
		result = append(result, wo)
	}
	return result, rows.Err()
}

func (r *Repository) FindWorkOrderByID(ctx context.Context, id string) (*models.WorkOrder, error) {
	var wo models.WorkOrder
	err := r.db.QueryRow(ctx,
		`SELECT id, branch_id, created_by, manager_asset_id, work_order_no, work_order_date,
			        company_name, company_address, company_phone, company_fax, company_email, company_website, company_logo_url,
			        bill_to_name, bill_to_address, bill_to_email, job_details,
			        signer_name, signer_title, signature_url, seal_url, currency, sub_total_amount, total_amount,
		        status, pdf_url, issued_at, created_at, updated_at
		 FROM work_orders
		 WHERE id = $1`, id,
	).Scan(workOrderScanDest(&wo)...)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	items, err := r.FindItemsByWorkOrderID(ctx, id)
	if err != nil {
		return nil, err
	}
	wo.Items = items
	return &wo, nil
}

func (r *Repository) FindItemsByWorkOrderID(ctx context.Context, workOrderID string) ([]models.WorkOrderItem, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, work_order_id, line_no, description, amount, created_at, updated_at
		 FROM work_order_items
		 WHERE work_order_id = $1
		 ORDER BY line_no ASC`, workOrderID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []models.WorkOrderItem
	for rows.Next() {
		var item models.WorkOrderItem
		if err := rows.Scan(&item.ID, &item.WorkOrderID, &item.LineNo, &item.Description, &item.Amount, &item.CreatedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (r *Repository) UpdateWorkOrder(ctx context.Context, id string, req UpdateWorkOrderRequest) (*models.WorkOrder, error) {
	current, err := r.FindWorkOrderByID(ctx, id)
	if err != nil || current == nil {
		return current, err
	}

	patch := updatePatchFromCurrent(current, req)

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var wo models.WorkOrder
	err = tx.QueryRow(ctx,
		`UPDATE work_orders SET
		   work_order_date = $2,
		   company_name = $3,
		   company_address = $4,
		   company_phone = $5,
		   company_fax = $6,
		   company_email = $7,
		   company_website = $8,
		   company_logo_url = $9,
			   bill_to_name = $10,
			   bill_to_address = $11,
			   bill_to_email = $12,
			   job_details = $13,
			   currency = $14,
			   sub_total_amount = $15,
			   total_amount = $16,
				   status = $17::work_order_status,
				   pdf_url = $18,
				   issued_at = CASE
				     WHEN $17::work_order_status = 'issued' AND issued_at IS NULL THEN NOW()
				     WHEN $17::work_order_status <> 'issued' THEN NULL
				     ELSE issued_at
				   END,
		   updated_at = NOW()
		 WHERE id = $1
			 RETURNING id, branch_id, created_by, manager_asset_id, work_order_no, work_order_date,
			           company_name, company_address, company_phone, company_fax, company_email, company_website, company_logo_url,
			           bill_to_name, bill_to_address, bill_to_email, job_details,
			           signer_name, signer_title, signature_url, seal_url, currency, sub_total_amount, total_amount,
		           status, pdf_url, issued_at, created_at, updated_at`,
		id,
		patch.WorkOrderDate,
		patch.CompanyName,
		patch.CompanyAddress,
		patch.CompanyPhone,
		patch.CompanyFax,
		patch.CompanyEmail,
		patch.CompanyWebsite,
		patch.CompanyLogoURL,
		patch.BillToName,
		patch.BillToAddress,
		patch.BillToEmail,
		patch.JobDetails,
		patch.Currency,
		patch.SubTotalAmount,
		patch.TotalAmount,
		patch.Status,
		patch.PDFURL,
	).Scan(workOrderScanDest(&wo)...)
	if err != nil {
		return nil, err
	}

	if req.Items != nil {
		if _, err := tx.Exec(ctx, `DELETE FROM work_order_items WHERE work_order_id = $1`, id); err != nil {
			return nil, err
		}
		wo.Items, err = insertItems(ctx, tx, id, req.Items)
		if err != nil {
			return nil, err
		}
	} else {
		wo.Items, err = r.FindItemsByWorkOrderID(ctx, id)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &wo, nil
}

func (r *Repository) UpdateStatus(ctx context.Context, id, status string, pdfURL *string) (*models.WorkOrder, error) {
	var wo models.WorkOrder
	err := r.db.QueryRow(ctx,
		`UPDATE work_orders SET
			   status = $2::work_order_status,
			   pdf_url = COALESCE($3, pdf_url),
			   issued_at = CASE
			     WHEN $2::work_order_status = 'issued' AND issued_at IS NULL THEN NOW()
			     WHEN $2::work_order_status <> 'issued' THEN NULL
			     ELSE issued_at
			   END,
		   updated_at = NOW()
		 WHERE id = $1
			 RETURNING id, branch_id, created_by, manager_asset_id, work_order_no, work_order_date,
			           company_name, company_address, company_phone, company_fax, company_email, company_website, company_logo_url,
			           bill_to_name, bill_to_address, bill_to_email, job_details,
			           signer_name, signer_title, signature_url, seal_url, currency, sub_total_amount, total_amount,
		           status, pdf_url, issued_at, created_at, updated_at`,
		id, status, pdfURL,
	).Scan(workOrderScanDest(&wo)...)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	items, err := r.FindItemsByWorkOrderID(ctx, id)
	if err != nil {
		return nil, err
	}
	wo.Items = items
	return &wo, nil
}

func (r *Repository) DeleteWorkOrder(ctx context.Context, id string) (bool, error) {
	tag, err := r.db.Exec(ctx, `DELETE FROM work_orders WHERE id = $1 AND status = 'draft'`, id)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

func generateWorkOrderNo(ctx context.Context, tx pgx.Tx, branchID, workOrderDate string) (string, error) {
	parsedDate, err := time.Parse("2006-01-02", workOrderDate)
	if err != nil {
		return "", err
	}

	year := parsedDate.Year()
	prefix := fmt.Sprintf("%d-WRO", year)
	pattern := fmt.Sprintf("^%s([0-9]+)$", prefix)

	if _, err := tx.Exec(ctx, `SELECT pg_advisory_xact_lock(hashtext($1))`, branchID+":"+prefix); err != nil {
		return "", err
	}

	var nextNumber int
	err = tx.QueryRow(ctx,
		`SELECT COALESCE(MAX(substring(work_order_no FROM $2)::INT), 0) + 1
		 FROM work_orders
		 WHERE branch_id = $1
		   AND work_order_no ~ $2`,
		branchID,
		pattern,
	).Scan(&nextNumber)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s%03d", prefix, nextNumber), nil
}

func workOrderScanDest(wo *models.WorkOrder) []any {
	return []any{
		&wo.ID, &wo.BranchID, &wo.CreatedBy, &wo.ManagerAssetID, &wo.WorkOrderNo, &wo.WorkOrderDate,
		&wo.CompanyName, &wo.CompanyAddress, &wo.CompanyPhone, &wo.CompanyFax, &wo.CompanyEmail, &wo.CompanyWebsite, &wo.CompanyLogoURL,
		&wo.BillToName, &wo.BillToAddress, &wo.BillToEmail, &wo.JobDetails,
		&wo.SignerName, &wo.SignerTitle, &wo.SignatureURL, &wo.SealURL, &wo.Currency, &wo.SubTotalAmount, &wo.TotalAmount,
		&wo.Status, &wo.PDFURL, &wo.IssuedAt, &wo.CreatedAt, &wo.UpdatedAt,
	}
}

func insertItems(ctx context.Context, tx pgx.Tx, workOrderID string, items []WorkOrderItemRequest) ([]models.WorkOrderItem, error) {
	result := make([]models.WorkOrderItem, 0, len(items))
	for i, req := range items {
		lineNo := req.LineNo
		if lineNo == 0 {
			lineNo = i + 1
		}

		var item models.WorkOrderItem
		err := tx.QueryRow(ctx,
			`INSERT INTO work_order_items (work_order_id, line_no, description, amount)
			 VALUES ($1, $2, $3, $4)
			 RETURNING id, work_order_id, line_no, description, amount, created_at, updated_at`,
			workOrderID, lineNo, req.Description, req.Amount,
		).Scan(&item.ID, &item.WorkOrderID, &item.LineNo, &item.Description, &item.Amount, &item.CreatedAt, &item.UpdatedAt)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	return result, nil
}

func calculateTotal(items []WorkOrderItemRequest) float64 {
	var total float64
	for _, item := range items {
		total += item.Amount
	}
	return total
}

type workOrderPatch struct {
	WorkOrderDate  time.Time
	CompanyName    string
	CompanyAddress *string
	CompanyPhone   *string
	CompanyFax     *string
	CompanyEmail   *string
	CompanyWebsite *string
	CompanyLogoURL *string
	BillToName     string
	BillToAddress  *string
	BillToEmail    *string
	JobDetails     string
	Currency       string
	SubTotalAmount float64
	TotalAmount    float64
	Status         string
	PDFURL         *string
}

func updatePatchFromCurrent(current *models.WorkOrder, req UpdateWorkOrderRequest) workOrderPatch {
	patch := workOrderPatch{
		WorkOrderDate:  current.WorkOrderDate,
		CompanyName:    current.CompanyName,
		CompanyAddress: current.CompanyAddress,
		CompanyPhone:   current.CompanyPhone,
		CompanyFax:     current.CompanyFax,
		CompanyEmail:   current.CompanyEmail,
		CompanyWebsite: current.CompanyWebsite,
		CompanyLogoURL: current.CompanyLogoURL,
		BillToName:     current.BillToName,
		BillToAddress:  current.BillToAddress,
		BillToEmail:    current.BillToEmail,
		JobDetails:     current.JobDetails,
		Currency:       current.Currency,
		SubTotalAmount: current.SubTotalAmount,
		TotalAmount:    current.TotalAmount,
		Status:         current.Status,
		PDFURL:         current.PDFURL,
	}

	if req.WorkOrderDate != nil {
		if parsed, err := time.Parse("2006-01-02", *req.WorkOrderDate); err == nil {
			patch.WorkOrderDate = parsed
		}
	}
	if req.CompanyName != nil {
		patch.CompanyName = *req.CompanyName
	}
	if req.CompanyAddress != nil {
		patch.CompanyAddress = req.CompanyAddress
	}
	if req.CompanyPhone != nil {
		patch.CompanyPhone = req.CompanyPhone
	}
	if req.CompanyFax != nil {
		patch.CompanyFax = req.CompanyFax
	}
	if req.CompanyEmail != nil {
		patch.CompanyEmail = req.CompanyEmail
	}
	if req.CompanyWebsite != nil {
		patch.CompanyWebsite = req.CompanyWebsite
	}
	if req.CompanyLogoURL != nil {
		patch.CompanyLogoURL = req.CompanyLogoURL
	}
	if req.BillToName != nil {
		patch.BillToName = *req.BillToName
	}
	if req.BillToAddress != nil {
		patch.BillToAddress = req.BillToAddress
	}
	if req.BillToEmail != nil {
		patch.BillToEmail = req.BillToEmail
	}
	if req.JobDetails != nil {
		patch.JobDetails = *req.JobDetails
	}
	if req.Currency != nil {
		patch.Currency = *req.Currency
	}
	if req.SubTotalAmount != nil {
		patch.SubTotalAmount = *req.SubTotalAmount
	}
	if req.TotalAmount != nil {
		patch.TotalAmount = *req.TotalAmount
	}
	if req.Status != nil {
		patch.Status = *req.Status
	}
	if req.PDFURL != nil {
		patch.PDFURL = req.PDFURL
	}
	if req.Items != nil {
		total := calculateTotal(req.Items)
		if req.SubTotalAmount == nil {
			patch.SubTotalAmount = total
		}
		if req.TotalAmount == nil {
			patch.TotalAmount = total
		}
	}

	return patch
}
