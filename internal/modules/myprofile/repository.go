package myprofile

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

func (r *Repository) FindMyProfileByID(ctx context.Context, id string) (*models.User, error) {
	var u models.User
	err := r.db.QueryRow(ctx,
		`SELECT id, branch_id, first_name, last_name, email, phone,
		        password_hash, role, status, avatar_url, last_login_at,
		        created_at, updated_at
		 FROM users WHERE id = $1`, id,
	).Scan(
		&u.ID, &u.BranchID, &u.FirstName, &u.LastName,
		&u.Email, &u.Phone, &u.PasswordHash, &u.Role,
		&u.Status, &u.AvatarURL, &u.LastLoginAt,
		&u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &u, nil
}

func (r *Repository) UpdateMyProfile(ctx context.Context, id string, u *models.User) error {
	_, err := r.db.Exec(ctx,
		`UPDATE users
		 SET first_name = $2, last_name = $3, email = $4, phone = $5, avatar_url = $6
		 WHERE id = $1`,
		id, u.FirstName, u.LastName, u.Email, u.Phone, u.AvatarURL,
	)
	return err
}

func (r *Repository) UpdateMyPassword(ctx context.Context, id string, passwordHash string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE users
		 SET password_hash = $2
		 WHERE id = $1`,
		id, passwordHash,
	)
	return err
}

func (r *Repository) FindMenusByRole(ctx context.Context, role string) ([]models.Menu, error) {
	rows, err := r.db.Query(ctx,
		`SELECT DISTINCT m.id, m.parent_id, m.label, m.path, m.resource,
		        m.sort_order, m.is_active, m.created_at, m.updated_at,
		        COALESCE(rp.can_view, TRUE) AS can_view,
		        COALESCE(rp.can_create, FALSE) AS can_create,
		        COALESCE(rp.can_edit, FALSE) AS can_edit,
		        COALESCE(rp.can_delete, FALSE) AS can_delete
		 FROM menus m
		 LEFT JOIN role_permissions rp ON m.resource = rp.resource AND rp.role = $1
		 WHERE m.is_active = TRUE
		   AND (
		       m.resource IS NULL
		       OR rp.can_view = TRUE
		   )
		 ORDER BY m.sort_order ASC, m.created_at ASC`, role,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var menus []models.Menu
	for rows.Next() {
		var m models.Menu
		var perms models.MenuPermissions
		err := rows.Scan(
			&m.ID, &m.ParentID, &m.Label, &m.Path,
			&m.Resource, &m.SortOrder, &m.IsActive, &m.CreatedAt, &m.UpdatedAt,
			&perms.CanView, &perms.CanCreate, &perms.CanEdit, &perms.CanDelete,
		)
		if err != nil {
			return nil, err
		}
		m.Permissions = &perms
		menus = append(menus, m)
	}
	return menus, nil
}
