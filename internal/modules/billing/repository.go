package billing

import (
	"context"

	"rmp-api/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindAll(ctx context.Context) ([]models.Billing, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, branch_id, patient_id, invoice_no, total_amount, paid_amount,
				status, notes, created_by, created_at, updated_at
		 FROM billing ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var billings []models.Billing
	for rows.Next() {
		var b models.Billing
		err := rows.Scan(
			&b.ID, &b.BranchID, &b.PatientID, &b.InvoiceNo,
			&b.TotalAmount, &b.PaidAmount, &b.Status, &b.Notes,
			&b.CreatedBy, &b.CreatedAt, &b.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		billings = append(billings, b)
	}
	return billings, nil
}

func (r *Repository) FindByID(ctx context.Context, id string) (*models.Billing, error) {
	var b models.Billing
	err := r.db.QueryRow(ctx,
		`SELECT id, branch_id, patient_id, invoice_no, total_amount, paid_amount,
				status, notes, created_by, created_at, updated_at
		 FROM billing WHERE id = $1`, id,
	).Scan(
		&b.ID, &b.BranchID, &b.PatientID, &b.InvoiceNo,
		&b.TotalAmount, &b.PaidAmount, &b.Status, &b.Notes,
		&b.CreatedBy, &b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *Repository) Create(ctx context.Context, b *models.Billing) error {
	return r.db.QueryRow(ctx,
		`INSERT INTO billing (branch_id, patient_id, invoice_no, total_amount, paid_amount, status, notes, created_by)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING id, created_at, updated_at`,
		b.BranchID, b.PatientID, b.InvoiceNo, b.TotalAmount,
		b.PaidAmount, b.Status, b.Notes, b.CreatedBy,
	).Scan(&b.ID, &b.CreatedAt, &b.UpdatedAt)
}

func (r *Repository) Update(ctx context.Context, id string, b *models.Billing) error {
	_, err := r.db.Exec(ctx,
		`UPDATE billing
		 SET total_amount = $2, paid_amount = $3, status = $4, notes = $5, updated_at = NOW()
		 WHERE id = $1`,
		id, b.TotalAmount, b.PaidAmount, b.Status, b.Notes,
	)
	return err
}
