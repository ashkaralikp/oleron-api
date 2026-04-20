package timesheet

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

type EmployeeConfig struct {
	EmployeeID         string
	UserID             string
	BranchID           string
	FixedMonthlySalary float64
	OTRate             float64
	Currency           string
}

// FetchConfigByUserID resolves the employee config from a user_id (for employee role).
func (r *Repository) FetchConfigByUserID(ctx context.Context, userID string) (*EmployeeConfig, error) {
	var c EmployeeConfig
	err := r.db.QueryRow(ctx,
		`SELECT id, user_id, branch_id,
		        COALESCE(fixed_monthly_salary, 0),
		        COALESCE(ot_rate, 0),
		        currency
		 FROM employees WHERE user_id = $1`, userID,
	).Scan(&c.EmployeeID, &c.UserID, &c.BranchID, &c.FixedMonthlySalary, &c.OTRate, &c.Currency)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

// FetchConfigByEmployeeID resolves the employee config from an explicit employee_id.
func (r *Repository) FetchConfigByEmployeeID(ctx context.Context, employeeID string) (*EmployeeConfig, error) {
	var c EmployeeConfig
	err := r.db.QueryRow(ctx,
		`SELECT id, user_id, branch_id,
		        COALESCE(fixed_monthly_salary, 0),
		        COALESCE(ot_rate, 0),
		        currency
		 FROM employees WHERE id = $1`, employeeID,
	).Scan(&c.EmployeeID, &c.UserID, &c.BranchID, &c.FixedMonthlySalary, &c.OTRate, &c.Currency)
	if err != nil {
		return nil, err
	}
	return &c, nil
}
