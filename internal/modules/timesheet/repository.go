package timesheet

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// ── Estimate config ───────────────────────────────────────────────

type EmployeeConfig struct {
	EmployeeID         string
	UserID             string
	BranchID           string
	FixedMonthlySalary float64
	OTRate             float64
	Currency           string
}

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

// ── Timesheet CRUD ────────────────────────────────────────────────

const tsCols = `
	ct.id, ct.employee_id, e.employee_code, u.first_name, u.last_name,
	ct.year, ct.month, ct.support_hours, ct.overtime_hours, ct.notes,
	ct.status, ct.reviewer_id, ct.review_note, ct.reviewed_at, ct.submitted_at`

const tsJoin = `
	FROM consultant_timesheets ct
	JOIN employees e ON e.id = ct.employee_id
	JOIN users     u ON u.id = e.user_id`

type scannable interface {
	Scan(...any) error
}

func scanTS(row scannable) (*TimesheetResponse, error) {
	var t TimesheetResponse
	err := row.Scan(
		&t.ID, &t.EmployeeID, &t.EmployeeCode, &t.FirstName, &t.LastName,
		&t.Year, &t.Month, &t.SupportHours, &t.OvertimeHours, &t.Notes,
		&t.Status, &t.ReviewerID, &t.ReviewNote, &t.ReviewedAt, &t.SubmittedAt,
	)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *Repository) Submit(ctx context.Context, employeeID string, req SubmitRequest) (*TimesheetResponse, error) {
	var notes *string
	if req.Notes != "" {
		notes = &req.Notes
	}

	var id string
	var submittedAt time.Time
	err := r.db.QueryRow(ctx,
		`INSERT INTO consultant_timesheets
		   (employee_id, year, month, support_hours, overtime_hours, notes)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 ON CONFLICT (employee_id, year, month) DO UPDATE
		   SET support_hours  = EXCLUDED.support_hours,
		       overtime_hours = EXCLUDED.overtime_hours,
		       notes          = EXCLUDED.notes,
		       status         = 'pending',
		       reviewer_id    = NULL,
		       review_note    = NULL,
		       reviewed_at    = NULL,
		       updated_at     = NOW()
		 RETURNING id, submitted_at`,
		employeeID, req.Year, req.Month, req.SupportHours, req.OvertimeHours, notes,
	).Scan(&id, &submittedAt)
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id)
}

func (r *Repository) GetMine(ctx context.Context, employeeID string, year, month int) (*TimesheetResponse, error) {
	return scanTS(r.db.QueryRow(ctx,
		`SELECT`+tsCols+tsJoin+`
		 WHERE ct.employee_id = $1 AND ct.year = $2 AND ct.month = $3`,
		employeeID, year, month,
	))
}

func (r *Repository) GetAll(ctx context.Context, role, branchID string) ([]TimesheetResponse, error) {
	var query string
	var args []any

	if role == "super_admin" {
		query = `SELECT` + tsCols + tsJoin + `
		         ORDER BY ct.year DESC, ct.month DESC, ct.submitted_at DESC`
	} else {
		query = `SELECT` + tsCols + tsJoin + `
		         WHERE e.branch_id = $1
		         ORDER BY ct.year DESC, ct.month DESC, ct.submitted_at DESC`
		args = append(args, branchID)
	}

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []TimesheetResponse
	for rows.Next() {
		t, err := scanTS(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *t)
	}
	return out, nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (*TimesheetResponse, error) {
	return scanTS(r.db.QueryRow(ctx,
		`SELECT`+tsCols+tsJoin+` WHERE ct.id = $1`, id,
	))
}

func (r *Repository) Review(ctx context.Context, id, status, reviewNote, reviewerUserID string) (*TimesheetResponse, error) {
	var note *string
	if reviewNote != "" {
		note = &reviewNote
	}
	_, err := r.db.Exec(ctx,
		`UPDATE consultant_timesheets
		 SET status      = $2,
		     review_note = $3,
		     reviewer_id = $4,
		     reviewed_at = NOW(),
		     updated_at  = NOW()
		 WHERE id = $1`,
		id, status, note, reviewerUserID,
	)
	if err != nil {
		return nil, err
	}
	return r.GetByID(ctx, id)
}

func (r *Repository) IsTimesheetInBranch(ctx context.Context, timesheetID, branchID string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx,
		`SELECT EXISTS(
		   SELECT 1
		   FROM consultant_timesheets ct
		   JOIN employees e ON e.id = ct.employee_id
		   WHERE ct.id = $1 AND e.branch_id = $2
		 )`, timesheetID, branchID,
	).Scan(&exists)
	return exists, err
}
