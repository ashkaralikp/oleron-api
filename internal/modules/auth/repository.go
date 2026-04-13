package auth

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

func (r *Repository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	var user models.User
	err := r.db.QueryRow(ctx,
		`SELECT id, branch_id, first_name, last_name, email, phone,
				password_hash, role, status, avatar_url, last_login_at,
				created_at, updated_at
		 FROM users WHERE email = $1`, email,
	).Scan(
		&user.ID, &user.BranchID, &user.FirstName, &user.LastName,
		&user.Email, &user.Phone, &user.PasswordHash, &user.Role,
		&user.Status, &user.AvatarURL, &user.LastLoginAt,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) FindByID(ctx context.Context, id string) (*models.User, error) {
	var user models.User
	err := r.db.QueryRow(ctx,
		`SELECT id, branch_id, first_name, last_name, email, phone,
				password_hash, role, status, avatar_url, last_login_at,
				created_at, updated_at
		 FROM users WHERE id = $1`, id,
	).Scan(
		&user.ID, &user.BranchID, &user.FirstName, &user.LastName,
		&user.Email, &user.Phone, &user.PasswordHash, &user.Role,
		&user.Status, &user.AvatarURL, &user.LastLoginAt,
		&user.CreatedAt, &user.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *Repository) UpdateLastLogin(ctx context.Context, userID string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE users SET last_login_at = NOW() WHERE id = $1`, userID,
	)
	return err
}
