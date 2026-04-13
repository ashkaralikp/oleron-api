package report

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetPatientCount(ctx context.Context, startDate, endDate string) (int, error) {
	var count int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM patients WHERE created_at BETWEEN $1 AND $2`,
		startDate, endDate,
	).Scan(&count)
	return count, err
}

func (r *Repository) GetBillingTotal(ctx context.Context, startDate, endDate string) (float64, error) {
	var total float64
	err := r.db.QueryRow(ctx,
		`SELECT COALESCE(SUM(total_amount), 0) FROM billing WHERE created_at BETWEEN $1 AND $2`,
		startDate, endDate,
	).Scan(&total)
	return total, err
}

func (r *Repository) GetAppointmentCount(ctx context.Context, startDate, endDate string) (int, error) {
	var count int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM appointments WHERE created_at BETWEEN $1 AND $2`,
		startDate, endDate,
	).Scan(&count)
	return count, err
}
