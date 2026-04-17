package payroll

import (
	"context"
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

// EmployeePayData holds raw data needed to compute one employee's pay.
type EmployeePayData struct {
	EmployeeID   string
	UserID       string
	FirstName    string
	LastName     string
	Email        string
	EmployeeCode string
	HourlyRate   float64
	Currency     string
	// Attendance summary for the period
	PresentDays int
	AbsentDays  int
	LeaveDays   int
	TotalHours  float64
}

// FetchEmployeePayData returns pay data for all employees in a branch for the given period.
func (r *Repository) FetchEmployeePayData(ctx context.Context, branchID, from, to string) ([]EmployeePayData, error) {
	rows, err := r.db.Query(ctx,
		`SELECT
		    e.id, e.user_id, u.first_name, u.last_name, u.email, e.employee_code,
		    COALESCE(e.hourly_rate, 0), COALESCE(e.currency, 'USD'),
		    COUNT(a.id) FILTER (WHERE a.status IN ('present','late_in','early_out','late_in_early_out')) AS present_days,
		    COUNT(a.id) FILTER (WHERE a.status = 'absent')   AS absent_days,
		    COUNT(a.id) FILTER (WHERE a.status = 'on_leave') AS leave_days,
		    COALESCE(SUM(a.work_hours) FILTER (WHERE a.work_hours IS NOT NULL), 0) AS total_hours
		 FROM employees e
		 JOIN users u ON u.id = e.user_id
		 LEFT JOIN attendance a
		    ON a.user_id = e.user_id
		    AND a.work_date BETWEEN $2 AND $3
		 WHERE e.branch_id = $1
		   AND u.status = 'active'
		 GROUP BY e.id, e.user_id, u.first_name, u.last_name, u.email, e.employee_code, e.hourly_rate, e.currency`,
		branchID, from, to,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []EmployeePayData
	for rows.Next() {
		var d EmployeePayData
		if err := rows.Scan(
			&d.EmployeeID, &d.UserID, &d.FirstName, &d.LastName, &d.Email, &d.EmployeeCode,
			&d.HourlyRate, &d.Currency,
			&d.PresentDays, &d.AbsentDays, &d.LeaveDays, &d.TotalHours,
		); err != nil {
			return nil, err
		}
		result = append(result, d)
	}
	return result, nil
}

// CountWorkingDays counts expected working days in the period for a branch,
// based on office_timing_days minus branch_calendar holidays.
func (r *Repository) CountWorkingDays(ctx context.Context, branchID, from, to string) (int, error) {
	var count int
	err := r.db.QueryRow(ctx,
		`WITH date_series AS (
		    SELECT generate_series($2::date, $3::date, '1 day'::interval)::date AS d
		),
		timing_days AS (
		    SELECT otd.day_of_week
		    FROM branches b
		    JOIN office_timings ot ON ot.id = b.office_timing_id
		    JOIN office_timing_days otd ON otd.office_timing_id = ot.id AND otd.is_working_day = TRUE
		    WHERE b.id = $1
		),
		holidays AS (
		    SELECT date FROM branch_calendar
		    WHERE branch_id = $1 AND type = 'holiday'
		      AND date BETWEEN $2::date AND $3::date
		),
		extra_working AS (
		    SELECT date FROM branch_calendar
		    WHERE branch_id = $1 AND type = 'working_day'
		      AND date BETWEEN $2::date AND $3::date
		)
		SELECT COUNT(*) FROM (
		    -- Regular working days from schedule, minus holidays
		    SELECT d FROM date_series
		    WHERE EXTRACT(DOW FROM d)::int IN (SELECT day_of_week FROM timing_days)
		      AND d NOT IN (SELECT date FROM holidays)
		    UNION
		    -- Extra working days from calendar overrides
		    SELECT date FROM extra_working
		) counted`,
		branchID, from, to,
	).Scan(&count)
	return count, err
}

// CreatePayrollRun inserts the run and all items atomically.
func (r *Repository) CreatePayrollRun(
	ctx context.Context,
	branchID, generatedBy, from, to, currency string,
	notes *string,
	total float64,
	items []models.PayrollItem,
) (*models.PayrollRun, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var run models.PayrollRun
	err = tx.QueryRow(ctx,
		`INSERT INTO payroll_runs (branch_id, period_from, period_to, generated_by, currency, total_amount, notes)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, branch_id, period_from, period_to, generated_by, status, total_amount, currency, notes, created_at, updated_at`,
		branchID, from, to, generatedBy, currency, total, notes,
	).Scan(
		&run.ID, &run.BranchID, &run.PeriodFrom, &run.PeriodTo, &run.GeneratedBy,
		&run.Status, &run.TotalAmount, &run.Currency, &run.Notes, &run.CreatedAt, &run.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	for i := range items {
		items[i].PayrollRunID = run.ID
		_, err = tx.Exec(ctx,
			`INSERT INTO payroll_items
			 (payroll_run_id, employee_id, user_id, working_days, present_days, absent_days, leave_days,
			  total_hours, hourly_rate, currency, gross_pay, deductions, net_pay)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
			run.ID, items[i].EmployeeID, items[i].UserID,
			items[i].WorkingDays, items[i].PresentDays, items[i].AbsentDays, items[i].LeaveDays,
			items[i].TotalHours, items[i].HourlyRate, items[i].Currency,
			items[i].GrossPay, items[i].Deductions, items[i].NetPay,
		)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	run.Items = items
	return &run, nil
}

// FindAllRuns returns payroll runs for a branch (or all branches for super_admin).
func (r *Repository) FindAllRuns(ctx context.Context, branchID string) ([]models.PayrollRun, error) {
	query := `SELECT id, branch_id, period_from, period_to, generated_by, status, total_amount, currency, notes, created_at, updated_at
	          FROM payroll_runs`
	args := []any{}
	if branchID != "" {
		query += ` WHERE branch_id = $1`
		args = append(args, branchID)
	}
	query += ` ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var runs []models.PayrollRun
	for rows.Next() {
		var run models.PayrollRun
		if err := rows.Scan(
			&run.ID, &run.BranchID, &run.PeriodFrom, &run.PeriodTo, &run.GeneratedBy,
			&run.Status, &run.TotalAmount, &run.Currency, &run.Notes, &run.CreatedAt, &run.UpdatedAt,
		); err != nil {
			return nil, err
		}
		runs = append(runs, run)
	}
	return runs, nil
}

// FindRunByID returns a payroll run with all its items (joined with user/employee data).
func (r *Repository) FindRunByID(ctx context.Context, id string) (*models.PayrollRun, error) {
	var run models.PayrollRun
	err := r.db.QueryRow(ctx,
		`SELECT id, branch_id, period_from, period_to, generated_by, status, total_amount, currency, notes, created_at, updated_at
		 FROM payroll_runs WHERE id = $1`, id,
	).Scan(
		&run.ID, &run.BranchID, &run.PeriodFrom, &run.PeriodTo, &run.GeneratedBy,
		&run.Status, &run.TotalAmount, &run.Currency, &run.Notes, &run.CreatedAt, &run.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	rows, err := r.db.Query(ctx,
		`SELECT pi.id, pi.payroll_run_id, pi.employee_id, pi.user_id,
		        u.first_name, u.last_name, u.email, e.employee_code,
		        pi.working_days, pi.present_days, pi.absent_days, pi.leave_days,
		        pi.total_hours, pi.hourly_rate, pi.currency,
		        pi.gross_pay, pi.deductions, pi.net_pay, pi.created_at
		 FROM payroll_items pi
		 JOIN users u ON u.id = pi.user_id
		 JOIN employees e ON e.id = pi.employee_id
		 WHERE pi.payroll_run_id = $1
		 ORDER BY u.first_name, u.last_name`, id,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item models.PayrollItem
		if err := rows.Scan(
			&item.ID, &item.PayrollRunID, &item.EmployeeID, &item.UserID,
			&item.FirstName, &item.LastName, &item.Email, &item.EmployeeCode,
			&item.WorkingDays, &item.PresentDays, &item.AbsentDays, &item.LeaveDays,
			&item.TotalHours, &item.HourlyRate, &item.Currency,
			&item.GrossPay, &item.Deductions, &item.NetPay, &item.CreatedAt,
		); err != nil {
			return nil, err
		}
		run.Items = append(run.Items, item)
	}
	return &run, nil
}

// UpdateStatus updates the status of a payroll run (approved / paid).
func (r *Repository) UpdateStatus(ctx context.Context, id, status string) (*models.PayrollRun, error) {
	var run models.PayrollRun
	err := r.db.QueryRow(ctx,
		`UPDATE payroll_runs SET status = $2, updated_at = NOW()
		 WHERE id = $1
		 RETURNING id, branch_id, period_from, period_to, generated_by, status, total_amount, currency, notes, created_at, updated_at`,
		id, status,
	).Scan(
		&run.ID, &run.BranchID, &run.PeriodFrom, &run.PeriodTo, &run.GeneratedBy,
		&run.Status, &run.TotalAmount, &run.Currency, &run.Notes, &run.CreatedAt, &run.UpdatedAt,
	)
	return &run, err
}

// DeleteRun deletes a draft payroll run (items cascade).
func (r *Repository) DeleteRun(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM payroll_runs WHERE id = $1 AND status = 'draft'`, id)
	return err
}

// scanRun is a helper used internally.
var _ = pgx.ErrNoRows // ensure pgx imported
var _ = time.Now      // ensure time imported
