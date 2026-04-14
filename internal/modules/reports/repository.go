package reports

import (
	"context"
	"fmt"

	"rmp-api/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	db *pgxpool.Pool
}

func NewRepository(db *pgxpool.Pool) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindAttendance(ctx context.Context, f AttendanceFilter) ([]models.Attendance, error) {
	query := `
		SELECT a.id, a.user_id, a.work_date, a.punch_in, a.punch_out,
		       a.work_hours, a.status, a.notes, a.created_at, a.updated_at,
		       u.first_name, u.last_name, u.email, u.branch_id,
		       e.employee_code
		FROM attendance a
		JOIN users u ON u.id = a.user_id
		LEFT JOIN employees e ON e.user_id = a.user_id
		WHERE 1=1`

	args := []any{}
	i := 1

	if f.BranchID != "" {
		query += fmt.Sprintf(" AND u.branch_id = $%d", i)
		args = append(args, f.BranchID)
		i++
	}
	if f.DateFrom != "" {
		query += fmt.Sprintf(" AND a.work_date >= $%d", i)
		args = append(args, f.DateFrom)
		i++
	}
	if f.DateTo != "" {
		query += fmt.Sprintf(" AND a.work_date <= $%d", i)
		args = append(args, f.DateTo)
		i++
	}
	if f.UserID != "" {
		query += fmt.Sprintf(" AND a.user_id = $%d", i)
		args = append(args, f.UserID)
		i++
	}
	if f.Status != "" {
		query += fmt.Sprintf(" AND a.status = $%d", i)
		args = append(args, f.Status)
		i++
	}

	query += " ORDER BY a.work_date DESC, u.first_name ASC"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var records []models.Attendance
	for rows.Next() {
		var a models.Attendance
		err := rows.Scan(
			&a.ID, &a.UserID, &a.WorkDate, &a.PunchIn, &a.PunchOut,
			&a.WorkHours, &a.Status, &a.Notes, &a.CreatedAt, &a.UpdatedAt,
			&a.FirstName, &a.LastName, &a.Email, &a.BranchID,
			&a.EmployeeCode,
		)
		if err != nil {
			return nil, err
		}
		records = append(records, a)
	}
	return records, nil
}
