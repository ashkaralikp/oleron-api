package contact

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
