package schedule

import (
	"context"
	"fmt"

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

// FindAll returns all office timings (filtered by branch if branchID is set).
func (r *Repository) FindAll(ctx context.Context, branchID string) ([]models.OfficeTiming, error) {
	query := `SELECT id, branch_id, name, is_active, created_at, updated_at FROM office_timings`
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

	var timings []models.OfficeTiming
	for rows.Next() {
		var t models.OfficeTiming
		if err := rows.Scan(&t.ID, &t.BranchID, &t.Name, &t.IsActive, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		timings = append(timings, t)
	}
	return timings, nil
}

// FindByID returns a single office timing with its days.
func (r *Repository) FindByID(ctx context.Context, id string) (*models.OfficeTiming, error) {
	var t models.OfficeTiming
	err := r.db.QueryRow(ctx,
		`SELECT id, branch_id, name, is_active, created_at, updated_at
		 FROM office_timings WHERE id = $1`, id,
	).Scan(&t.ID, &t.BranchID, &t.Name, &t.IsActive, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}

	days, err := r.findDays(ctx, id)
	if err != nil {
		return nil, err
	}
	t.Days = days
	return &t, nil
}

func (r *Repository) findDays(ctx context.Context, officeTimingID string) ([]models.OfficeTimingDay, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, office_timing_id, day_of_week, is_working_day, start_time, end_time, break_minutes
		 FROM office_timing_days WHERE office_timing_id = $1 ORDER BY day_of_week`, officeTimingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var days []models.OfficeTimingDay
	for rows.Next() {
		var d models.OfficeTimingDay
		if err := rows.Scan(
			&d.ID, &d.OfficeTimingID, &d.DayOfWeek,
			&d.IsWorkingDay, &d.StartTime, &d.EndTime, &d.BreakMinutes,
		); err != nil {
			return nil, err
		}
		days = append(days, d)
	}
	return days, nil
}

// Create inserts a new office timing and its days atomically.
func (r *Repository) Create(ctx context.Context, branchID, name string, days []DayInput) (*models.OfficeTiming, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var t models.OfficeTiming
	err = tx.QueryRow(ctx,
		`INSERT INTO office_timings (branch_id, name)
		 VALUES ($1, $2)
		 RETURNING id, branch_id, name, is_active, created_at, updated_at`,
		branchID, name,
	).Scan(&t.ID, &t.BranchID, &t.Name, &t.IsActive, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}

	if err := insertDaysInTx(ctx, tx, t.ID, days); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	t.Days, _ = r.findDays(ctx, t.ID)
	return &t, nil
}

// Update replaces the timing name and all its days atomically.
func (r *Repository) Update(ctx context.Context, id, name string, days []DayInput) (*models.OfficeTiming, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var t models.OfficeTiming
	err = tx.QueryRow(ctx,
		`UPDATE office_timings SET name = $1, updated_at = NOW()
		 WHERE id = $2
		 RETURNING id, branch_id, name, is_active, created_at, updated_at`,
		name, id,
	).Scan(&t.ID, &t.BranchID, &t.Name, &t.IsActive, &t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		return nil, err
	}

	if _, err := tx.Exec(ctx, `DELETE FROM office_timing_days WHERE office_timing_id = $1`, id); err != nil {
		return nil, err
	}
	if err := insertDaysInTx(ctx, tx, id, days); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	t.Days, _ = r.findDays(ctx, t.ID)
	return &t, nil
}

// Delete removes an office timing (days cascade via FK).
func (r *Repository) Delete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM office_timings WHERE id = $1`, id)
	return err
}

// Activate sets branches.office_timing_id to the given timing.
// Verifies the timing belongs to the branch (non-super_admin callers pass their branchID).
func (r *Repository) Activate(ctx context.Context, timingID, branchID string) error {
	var ownerBranchID string
	err := r.db.QueryRow(ctx,
		`SELECT branch_id FROM office_timings WHERE id = $1`, timingID,
	).Scan(&ownerBranchID)
	if err != nil {
		return fmt.Errorf("timing not found")
	}
	if branchID != "" && ownerBranchID != branchID {
		return fmt.Errorf("forbidden")
	}

	_, err = r.db.Exec(ctx,
		`UPDATE branches SET office_timing_id = $1, updated_at = NOW() WHERE id = $2`,
		timingID, ownerBranchID,
	)
	return err
}

// insertDaysInTx inserts office_timing_days rows within an open transaction.
func insertDaysInTx(ctx context.Context, tx pgx.Tx, timingID string, days []DayInput) error {
	for _, d := range days {
		_, err := tx.Exec(ctx,
			`INSERT INTO office_timing_days
			 (office_timing_id, day_of_week, is_working_day, start_time, end_time, break_minutes)
			 VALUES ($1, $2, $3, $4, $5, $6)`,
			timingID, d.DayOfWeek, d.IsWorkingDay, d.StartTime, d.EndTime, d.BreakMinutes,
		)
		if err != nil {
			return err
		}
	}
	return nil
}
