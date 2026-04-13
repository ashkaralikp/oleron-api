package appointment

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

func (r *Repository) FindAll(ctx context.Context) ([]models.Appointment, error) {
	rows, err := r.db.Query(ctx,
		`SELECT id, branch_id, patient_id, doctor_id, date, start_time, end_time,
				status, notes, created_at, updated_at
		 FROM appointments ORDER BY date DESC, start_time ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var appointments []models.Appointment
	for rows.Next() {
		var a models.Appointment
		err := rows.Scan(
			&a.ID, &a.BranchID, &a.PatientID, &a.DoctorID,
			&a.Date, &a.StartTime, &a.EndTime, &a.Status,
			&a.Notes, &a.CreatedAt, &a.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		appointments = append(appointments, a)
	}
	return appointments, nil
}

func (r *Repository) FindByID(ctx context.Context, id string) (*models.Appointment, error) {
	var a models.Appointment
	err := r.db.QueryRow(ctx,
		`SELECT id, branch_id, patient_id, doctor_id, date, start_time, end_time,
				status, notes, created_at, updated_at
		 FROM appointments WHERE id = $1`, id,
	).Scan(
		&a.ID, &a.BranchID, &a.PatientID, &a.DoctorID,
		&a.Date, &a.StartTime, &a.EndTime, &a.Status,
		&a.Notes, &a.CreatedAt, &a.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *Repository) Create(ctx context.Context, a *models.Appointment) error {
	return r.db.QueryRow(ctx,
		`INSERT INTO appointments (branch_id, patient_id, doctor_id, date, start_time, end_time, status, notes)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING id, created_at, updated_at`,
		a.BranchID, a.PatientID, a.DoctorID, a.Date,
		a.StartTime, a.EndTime, a.Status, a.Notes,
	).Scan(&a.ID, &a.CreatedAt, &a.UpdatedAt)
}

func (r *Repository) Update(ctx context.Context, id string, a *models.Appointment) error {
	_, err := r.db.Exec(ctx,
		`UPDATE appointments
		 SET patient_id = $2, doctor_id = $3, date = $4, start_time = $5,
			 end_time = $6, status = $7, notes = $8, updated_at = NOW()
		 WHERE id = $1`,
		id, a.PatientID, a.DoctorID, a.Date,
		a.StartTime, a.EndTime, a.Status, a.Notes,
	)
	return err
}
