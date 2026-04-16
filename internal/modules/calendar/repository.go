package calendar

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

// FindAll returns calendar entries filtered by branch and optional date range / type.
func (r *Repository) FindAll(ctx context.Context, f CalendarRangeFilter) ([]models.BranchCalendar, error) {
	query := `SELECT id, branch_id, date, type, name, created_at FROM branch_calendar WHERE 1=1`
	args := []any{}
	n := 1

	if f.BranchID != "" {
		query += fmt.Sprintf(` AND branch_id = $%d`, n)
		args = append(args, f.BranchID)
		n++
	}
	if f.From != "" {
		query += fmt.Sprintf(` AND date >= $%d`, n)
		args = append(args, f.From)
		n++
	}
	if f.To != "" {
		query += fmt.Sprintf(` AND date <= $%d`, n)
		args = append(args, f.To)
		n++
	}
	if f.Type != "" {
		query += fmt.Sprintf(` AND type = $%d`, n)
		args = append(args, f.Type)
		n++
	}
	query += ` ORDER BY date ASC`
	_ = n

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []models.BranchCalendar
	for rows.Next() {
		var e models.BranchCalendar
		if err := rows.Scan(&e.ID, &e.BranchID, &e.Date, &e.Type, &e.Name, &e.CreatedAt); err != nil {
			return nil, err
		}
		entries = append(entries, e)
	}
	return entries, nil
}

// FindByID returns a single calendar entry.
func (r *Repository) FindByID(ctx context.Context, id string) (*models.BranchCalendar, error) {
	var e models.BranchCalendar
	err := r.db.QueryRow(ctx,
		`SELECT id, branch_id, date, type, name, created_at
		 FROM branch_calendar WHERE id = $1`, id,
	).Scan(&e.ID, &e.BranchID, &e.Date, &e.Type, &e.Name, &e.CreatedAt)
	if err != nil {
		return nil, err
	}
	return &e, nil
}

// Create inserts a new calendar entry.
func (r *Repository) Create(ctx context.Context, branchID, date, entryType string, name *string) (*models.BranchCalendar, error) {
	var e models.BranchCalendar
	err := r.db.QueryRow(ctx,
		`INSERT INTO branch_calendar (branch_id, date, type, name)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, branch_id, date, type, name, created_at`,
		branchID, date, entryType, name,
	).Scan(&e.ID, &e.BranchID, &e.Date, &e.Type, &e.Name, &e.CreatedAt)
	return &e, err
}

// Update replaces the type and name of an existing entry.
func (r *Repository) Update(ctx context.Context, id, entryType string, name *string) (*models.BranchCalendar, error) {
	var e models.BranchCalendar
	err := r.db.QueryRow(ctx,
		`UPDATE branch_calendar SET type = $2, name = $3
		 WHERE id = $1
		 RETURNING id, branch_id, date, type, name, created_at`,
		id, entryType, name,
	).Scan(&e.ID, &e.BranchID, &e.Date, &e.Type, &e.Name, &e.CreatedAt)
	return &e, err
}

// Delete removes a calendar entry.
func (r *Repository) Delete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM branch_calendar WHERE id = $1`, id)
	return err
}
