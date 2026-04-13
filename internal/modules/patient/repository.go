package patient

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

func (r *Repository) FindAll(ctx context.Context) ([]models.Patient, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, branch_id, first_name, last_name, phone, email,
				date_of_birth, gender, address, created_at, updated_at
		 FROM patients ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var patients []models.Patient
	for rows.Next() {
		var p models.Patient
		err := rows.Scan(
			&p.ID, &p.BranchID, &p.FirstName, &p.LastName,
			&p.Phone, &p.Email, &p.DateOfBirth, &p.Gender,
			&p.Address, &p.CreatedAt, &p.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		patients = append(patients, p)
	}
	return patients, nil
}

func (r *Repository) FindByID(ctx context.Context, id string) (*models.Patient, error) {
	var p models.Patient
	err := r.db.QueryRow(ctx,
		`SELECT id, branch_id, first_name, last_name, phone, email,
				date_of_birth, gender, address, created_at, updated_at
		 FROM patients WHERE id = $1`, id,
	).Scan(
		&p.ID, &p.BranchID, &p.FirstName, &p.LastName,
		&p.Phone, &p.Email, &p.DateOfBirth, &p.Gender,
		&p.Address, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repository) Create(ctx context.Context, p *models.Patient) error {
	return r.db.QueryRow(ctx,
		`INSERT INTO patients (branch_id, first_name, last_name, phone, email, date_of_birth, gender, address)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING id, created_at, updated_at`,
		p.BranchID, p.FirstName, p.LastName, p.Phone,
		p.Email, p.DateOfBirth, p.Gender, p.Address,
	).Scan(&p.ID, &p.CreatedAt, &p.UpdatedAt)
}

func (r *Repository) Update(ctx context.Context, id string, p *models.Patient) error {
	_, err := r.db.Exec(ctx,
		`UPDATE patients
		 SET first_name = $2, last_name = $3, phone = $4, email = $5,
			 date_of_birth = $6, gender = $7, address = $8, updated_at = NOW()
		 WHERE id = $1`,
		id, p.FirstName, p.LastName, p.Phone, p.Email,
		p.DateOfBirth, p.Gender, p.Address,
	)
	return err
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM patients WHERE id = $1`, id)
	return err
}
