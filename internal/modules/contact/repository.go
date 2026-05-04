package contact

import (
	"context"
	"errors"

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

func (r *Repository) CreateSubmission(ctx context.Context, req CreateSubmissionRequest, ipAddress, userAgent *string) (*models.ContactSubmission, error) {
	var submission models.ContactSubmission

	err := r.db.QueryRow(ctx,
		`INSERT INTO contact_submissions (name, company, email, phone, category, message, ip_address, user_agent)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING id, name, company, email, phone, category, message, status,
		           host(ip_address), user_agent, created_at, updated_at`,
		req.Name, req.Company, req.Email, req.Phone, req.Category, req.Message, ipAddress, userAgent,
	).Scan(
		&submission.ID,
		&submission.Name,
		&submission.Company,
		&submission.Email,
		&submission.Phone,
		&submission.Category,
		&submission.Message,
		&submission.Status,
		&submission.IPAddress,
		&submission.UserAgent,
		&submission.CreatedAt,
		&submission.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &submission, nil
}

func (r *Repository) GetAll(ctx context.Context, statusFilter string) ([]*models.ContactSubmission, error) {
	query := `SELECT id, name, company, email, phone, category, message, status,
	                 host(ip_address), user_agent, created_at, updated_at
	          FROM contact_submissions`
	args := []any{}

	if statusFilter != "" {
		query += ` WHERE status = $1`
		args = append(args, statusFilter)
	}

	query += ` ORDER BY created_at DESC`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var submissions []*models.ContactSubmission
	for rows.Next() {
		var s models.ContactSubmission
		if err := rows.Scan(
			&s.ID, &s.Name, &s.Company, &s.Email, &s.Phone, &s.Category,
			&s.Message, &s.Status, &s.IPAddress, &s.UserAgent, &s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, err
		}
		submissions = append(submissions, &s)
	}

	return submissions, rows.Err()
}

func (r *Repository) GetByID(ctx context.Context, id string) (*models.ContactSubmission, error) {
	var s models.ContactSubmission

	err := r.db.QueryRow(ctx,
		`SELECT id, name, company, email, phone, category, message, status,
		        host(ip_address), user_agent, created_at, updated_at
		 FROM contact_submissions WHERE id = $1`, id,
	).Scan(
		&s.ID, &s.Name, &s.Company, &s.Email, &s.Phone, &s.Category,
		&s.Message, &s.Status, &s.IPAddress, &s.UserAgent, &s.CreatedAt, &s.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func (r *Repository) UpdateStatus(ctx context.Context, id, status string) (*models.ContactSubmission, error) {
	var s models.ContactSubmission

	err := r.db.QueryRow(ctx,
		`UPDATE contact_submissions SET status = $1
		 WHERE id = $2
		 RETURNING id, name, company, email, phone, category, message, status,
		           host(ip_address), user_agent, created_at, updated_at`,
		status, id,
	).Scan(
		&s.ID, &s.Name, &s.Company, &s.Email, &s.Phone, &s.Category,
		&s.Message, &s.Status, &s.IPAddress, &s.UserAgent, &s.CreatedAt, &s.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &s, nil
}

func (r *Repository) Delete(ctx context.Context, id string) (bool, error) {
	tag, err := r.db.Exec(ctx, `DELETE FROM contact_submissions WHERE id = $1`, id)
	if err != nil {
		return false, err
	}

	return tag.RowsAffected() > 0, nil
}
