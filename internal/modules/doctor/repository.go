package doctor

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

func (r *Repository) FindAll(ctx context.Context) ([]models.Doctor, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, branch_id, user_id, specialization, license_no, is_available, created_at, updated_at
		 FROM doctors ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var doctors []models.Doctor
	for rows.Next() {
		var d models.Doctor
		err := rows.Scan(
			&d.ID, &d.BranchID, &d.UserID, &d.Specialization,
			&d.LicenseNo, &d.IsAvailable, &d.CreatedAt, &d.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		doctors = append(doctors, d)
	}
	return doctors, nil
}

func (r *Repository) FindByID(ctx context.Context, id string) (*models.Doctor, error) {
	var d models.Doctor
	err := r.db.QueryRow(ctx,
		`SELECT id, branch_id, user_id, specialization, license_no, is_available, created_at, updated_at
		 FROM doctors WHERE id = $1`, id,
	).Scan(
		&d.ID, &d.BranchID, &d.UserID, &d.Specialization,
		&d.LicenseNo, &d.IsAvailable, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *Repository) Create(ctx context.Context, d *models.Doctor) error {
	return r.db.QueryRow(ctx,
		`INSERT INTO doctors (branch_id, user_id, specialization, license_no, is_available)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, created_at, updated_at`,
		d.BranchID, d.UserID, d.Specialization, d.LicenseNo, d.IsAvailable,
	).Scan(&d.ID, &d.CreatedAt, &d.UpdatedAt)
}

func (r *Repository) Update(ctx context.Context, id string, d *models.Doctor) error {
	_, err := r.db.Exec(ctx,
		`UPDATE doctors
		 SET specialization = $2, license_no = $3, is_available = $4, updated_at = NOW()
		 WHERE id = $1`,
		id, d.Specialization, d.LicenseNo, d.IsAvailable,
	)
	return err
}
