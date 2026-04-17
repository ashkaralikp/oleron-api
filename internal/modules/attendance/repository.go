package attendance

import (
	"context"
	"time"

	"rmp-api/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

// FindTodayRecord returns the attendance record for the given user for today, or nil if none exists.
func (r *Repository) FindTodayRecord(ctx context.Context, userID string) (*models.Attendance, error) {
	var a models.Attendance
	err := r.db.QueryRow(ctx,
		`SELECT id, user_id, work_date, punch_in, punch_out, work_hours, status, notes, created_at, updated_at
		 FROM attendance
		 WHERE user_id = $1 AND work_date = CURRENT_DATE`,
		userID,
	).Scan(
		&a.ID, &a.UserID, &a.WorkDate,
		&a.PunchIn, &a.PunchOut, &a.WorkHours,
		&a.Status, &a.Notes, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

// CreatePunchIn creates a new attendance record with punch_in time and initial status.
func (r *Repository) CreatePunchIn(ctx context.Context, userID string, punchIn time.Time, status string) (*models.Attendance, error) {
	var a models.Attendance
	err := r.db.QueryRow(ctx,
		`INSERT INTO attendance (user_id, work_date, punch_in, status)
		 VALUES ($1, CURRENT_DATE, $2, $3)
		 RETURNING id, user_id, work_date, punch_in, punch_out, work_hours, status, notes, created_at, updated_at`,
		userID, punchIn, status,
	).Scan(
		&a.ID, &a.UserID, &a.WorkDate,
		&a.PunchIn, &a.PunchOut, &a.WorkHours,
		&a.Status, &a.Notes, &a.CreatedAt, &a.UpdatedAt,
	)
	return &a, err
}

// UpdatePunchOut sets punch_out, work_hours, and final status on an existing record.
func (r *Repository) UpdatePunchOut(ctx context.Context, id string, punchOut time.Time, workHours float64, status string) (*models.Attendance, error) {
	var a models.Attendance
	err := r.db.QueryRow(ctx,
		`UPDATE attendance
		 SET punch_out = $2, work_hours = $3, status = $4, updated_at = NOW()
		 WHERE id = $1
		 RETURNING id, user_id, work_date, punch_in, punch_out, work_hours, status, notes, created_at, updated_at`,
		id, punchOut, workHours, status,
	).Scan(
		&a.ID, &a.UserID, &a.WorkDate,
		&a.PunchIn, &a.PunchOut, &a.WorkHours,
		&a.Status, &a.Notes, &a.CreatedAt, &a.UpdatedAt,
	)
	return &a, err
}

// FindDayTiming resolves the office timing for a user on a given day of the week.
// Checks employee.office_timing_id first, falls back to branch.office_timing_id.
func (r *Repository) FindDayTiming(ctx context.Context, userID string, dayOfWeek int) (*DayTiming, error) {
	var dt DayTiming
	err := r.db.QueryRow(ctx,
		`SELECT otd.is_working_day, otd.start_time, otd.end_time, otd.break_minutes
		 FROM employees e
		 JOIN users u ON u.id = e.user_id
		 JOIN office_timings ot
		   ON ot.id = COALESCE(
		       e.office_timing_id,
		       (SELECT office_timing_id FROM branches WHERE id = e.branch_id)
		   )
		 JOIN office_timing_days otd
		   ON otd.office_timing_id = ot.id AND otd.day_of_week = $2
		 WHERE e.user_id = $1`,
		userID, dayOfWeek,
	).Scan(&dt.IsWorkingDay, &dt.StartTime, &dt.EndTime, &dt.BreakMinutes)
	if err != nil {
		return nil, err
	}
	return &dt, nil
}

// IsHoliday checks if today is marked as a holiday in branch_calendar for the user's branch.
func (r *Repository) IsHoliday(ctx context.Context, userID string) (bool, string, error) {
	var calType, name string
	err := r.db.QueryRow(ctx,
		`SELECT type, COALESCE(name, '')
		 FROM branch_calendar
		 WHERE branch_id = (SELECT branch_id FROM users WHERE id = $1)
		   AND date = CURRENT_DATE`,
		userID,
	).Scan(&calType, &name)
	if err != nil {
		// no row = not a holiday
		return false, "", nil
	}
	return calType == "holiday", name, nil
}
