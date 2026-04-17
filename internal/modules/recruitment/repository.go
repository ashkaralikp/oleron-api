package recruitment

import (
	"context"
	"errors"
	"fmt"
	"strings"
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

// ─────────────────────────────────────────────
// VACANCIES
// ─────────────────────────────────────────────

func (r *Repository) FindAllVacancies(ctx context.Context, branchID string) ([]models.Vacancy, error) {
	query := `SELECT v.id, v.branch_id, v.created_by, v.title, v.department,
	                 v.description, v.requirements, v.positions, v.status, v.deadline,
	                 v.created_at, v.updated_at,
	                 COUNT(a.id) AS application_count
	          FROM vacancies v
	          LEFT JOIN applications a ON a.vacancy_id = v.id`
	args := []any{}
	if branchID != "" {
		query += ` WHERE v.branch_id = $1`
		args = append(args, branchID)
	}
	query += ` GROUP BY v.id ORDER BY v.created_at DESC`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.Vacancy
	for rows.Next() {
		var v models.Vacancy
		if err := rows.Scan(
			&v.ID, &v.BranchID, &v.CreatedBy, &v.Title, &v.Department,
			&v.Description, &v.Requirements, &v.Positions, &v.Status, &v.Deadline,
			&v.CreatedAt, &v.UpdatedAt, &v.ApplicationCount,
		); err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	return result, nil
}

func (r *Repository) FindVacancyByID(ctx context.Context, id string) (*models.Vacancy, error) {
	var v models.Vacancy
	err := r.db.QueryRow(ctx,
		`SELECT v.id, v.branch_id, v.created_by, v.title, v.department,
		        v.description, v.requirements, v.positions, v.status, v.deadline,
		        v.created_at, v.updated_at,
		        COUNT(a.id) AS application_count
		 FROM vacancies v
		 LEFT JOIN applications a ON a.vacancy_id = v.id
		 WHERE v.id = $1
		 GROUP BY v.id`, id,
	).Scan(
		&v.ID, &v.BranchID, &v.CreatedBy, &v.Title, &v.Department,
		&v.Description, &v.Requirements, &v.Positions, &v.Status, &v.Deadline,
		&v.CreatedAt, &v.UpdatedAt, &v.ApplicationCount,
	)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *Repository) CreateVacancy(ctx context.Context, branchID, createdBy string, req CreateVacancyRequest) (*models.Vacancy, error) {
	positions := req.Positions
	if positions < 1 {
		positions = 1
	}
	var v models.Vacancy
	err := r.db.QueryRow(ctx,
		`INSERT INTO vacancies (branch_id, created_by, title, department, description, requirements, positions, deadline)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING id, branch_id, created_by, title, department, description, requirements, positions, status, deadline, created_at, updated_at`,
		branchID, createdBy, req.Title, req.Department, req.Description, req.Requirements, positions, req.Deadline,
	).Scan(
		&v.ID, &v.BranchID, &v.CreatedBy, &v.Title, &v.Department,
		&v.Description, &v.Requirements, &v.Positions, &v.Status, &v.Deadline,
		&v.CreatedAt, &v.UpdatedAt,
	)
	return &v, err
}

func (r *Repository) UpdateVacancy(ctx context.Context, id string, req UpdateVacancyRequest) (*models.Vacancy, error) {
	fields := []string{}
	args := []any{}
	n := 1

	if req.Title != nil {
		fields = append(fields, fmt.Sprintf("title = $%d", n))
		args = append(args, *req.Title)
		n++
	}
	if req.Department != nil {
		fields = append(fields, fmt.Sprintf("department = $%d", n))
		args = append(args, *req.Department)
		n++
	}
	if req.Description != nil {
		fields = append(fields, fmt.Sprintf("description = $%d", n))
		args = append(args, *req.Description)
		n++
	}
	if req.Requirements != nil {
		fields = append(fields, fmt.Sprintf("requirements = $%d", n))
		args = append(args, *req.Requirements)
		n++
	}
	if req.Positions != nil {
		fields = append(fields, fmt.Sprintf("positions = $%d", n))
		args = append(args, *req.Positions)
		n++
	}
	if req.Deadline != nil {
		fields = append(fields, fmt.Sprintf("deadline = $%d", n))
		args = append(args, *req.Deadline)
		n++
	}
	if len(fields) == 0 {
		return r.FindVacancyByID(ctx, id)
	}

	fields = append(fields, "updated_at = NOW()")
	args = append(args, id)
	query := fmt.Sprintf(
		`UPDATE vacancies SET %s WHERE id = $%d
		 RETURNING id, branch_id, created_by, title, department, description, requirements, positions, status, deadline, created_at, updated_at`,
		strings.Join(fields, ", "), n,
	)
	var v models.Vacancy
	err := r.db.QueryRow(ctx, query, args...).Scan(
		&v.ID, &v.BranchID, &v.CreatedBy, &v.Title, &v.Department,
		&v.Description, &v.Requirements, &v.Positions, &v.Status, &v.Deadline,
		&v.CreatedAt, &v.UpdatedAt,
	)
	return &v, err
}

func (r *Repository) UpdateVacancyStatus(ctx context.Context, id, status string) (*models.Vacancy, error) {
	var v models.Vacancy
	err := r.db.QueryRow(ctx,
		`UPDATE vacancies SET status = $2, updated_at = NOW()
		 WHERE id = $1
		 RETURNING id, branch_id, created_by, title, department, description, requirements, positions, status, deadline, created_at, updated_at`,
		id, status,
	).Scan(
		&v.ID, &v.BranchID, &v.CreatedBy, &v.Title, &v.Department,
		&v.Description, &v.Requirements, &v.Positions, &v.Status, &v.Deadline,
		&v.CreatedAt, &v.UpdatedAt,
	)
	return &v, err
}

func (r *Repository) DeleteVacancy(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM vacancies WHERE id = $1`, id)
	return err
}

// ─────────────────────────────────────────────
// OWNERSHIP HELPERS
// ─────────────────────────────────────────────

func (r *Repository) FindVacancyBranchID(ctx context.Context, vacancyID string) (string, error) {
	var branchID string
	err := r.db.QueryRow(ctx, `SELECT branch_id FROM vacancies WHERE id = $1`, vacancyID).Scan(&branchID)
	return branchID, err
}

func (r *Repository) FindVacancyBranchIDByApplicationID(ctx context.Context, applicationID string) (string, error) {
	var branchID string
	err := r.db.QueryRow(ctx,
		`SELECT v.branch_id FROM applications a
		 JOIN vacancies v ON v.id = a.vacancy_id
		 WHERE a.id = $1`, applicationID,
	).Scan(&branchID)
	return branchID, err
}

func (r *Repository) FindVacancyBranchIDByInterviewID(ctx context.Context, interviewID string) (string, error) {
	var branchID string
	err := r.db.QueryRow(ctx,
		`SELECT v.branch_id FROM interviews i
		 JOIN applications a ON a.id = i.application_id
		 JOIN vacancies v ON v.id = a.vacancy_id
		 WHERE i.id = $1`, interviewID,
	).Scan(&branchID)
	return branchID, err
}

// ─────────────────────────────────────────────
// APPLICATIONS
// ─────────────────────────────────────────────

func (r *Repository) CreateApplication(ctx context.Context, vacancyID string, req ApplyRequest) (*models.Application, error) {
	// Verify vacancy exists and is open
	var status string
	err := r.db.QueryRow(ctx, `SELECT status FROM vacancies WHERE id = $1`, vacancyID).Scan(&status)
	if err != nil {
		return nil, errors.New("vacancy not found")
	}
	if status != "open" {
		return nil, errors.New("vacancy is not open for applications")
	}

	var app models.Application
	err = r.db.QueryRow(ctx,
		`INSERT INTO applications (vacancy_id, first_name, last_name, email, phone, cv_url, cover_letter)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)
		 RETURNING id, vacancy_id, first_name, last_name, email, phone, cv_url, cover_letter, status, notes, applied_at, updated_at`,
		vacancyID, req.FirstName, req.LastName, req.Email, req.Phone, req.CVUrl, req.CoverLetter,
	).Scan(
		&app.ID, &app.VacancyID, &app.FirstName, &app.LastName, &app.Email,
		&app.Phone, &app.CVUrl, &app.CoverLetter, &app.Status, &app.Notes,
		&app.AppliedAt, &app.UpdatedAt,
	)
	return &app, err
}

func (r *Repository) FindApplicationsByVacancy(ctx context.Context, vacancyID, statusFilter string) ([]models.Application, error) {
	query := `SELECT id, vacancy_id, first_name, last_name, email, phone, cv_url, cover_letter,
	                 status, notes, applied_at, updated_at
	          FROM applications WHERE vacancy_id = $1`
	args := []any{vacancyID}
	if statusFilter != "" {
		query += ` AND status = $2`
		args = append(args, statusFilter)
	}
	query += ` ORDER BY applied_at DESC`

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.Application
	for rows.Next() {
		var a models.Application
		if err := rows.Scan(
			&a.ID, &a.VacancyID, &a.FirstName, &a.LastName, &a.Email,
			&a.Phone, &a.CVUrl, &a.CoverLetter, &a.Status, &a.Notes,
			&a.AppliedAt, &a.UpdatedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, a)
	}
	return result, nil
}

func (r *Repository) FindApplicationByID(ctx context.Context, id string) (*models.Application, error) {
	var app models.Application
	err := r.db.QueryRow(ctx,
		`SELECT id, vacancy_id, first_name, last_name, email, phone, cv_url, cover_letter,
		        status, notes, applied_at, updated_at
		 FROM applications WHERE id = $1`, id,
	).Scan(
		&app.ID, &app.VacancyID, &app.FirstName, &app.LastName, &app.Email,
		&app.Phone, &app.CVUrl, &app.CoverLetter, &app.Status, &app.Notes,
		&app.AppliedAt, &app.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Fetch interviews
	rows, err := r.db.Query(ctx,
		`SELECT id, application_id, interviewer_id, scheduled_at, type, location,
		        outcome, feedback, created_at, updated_at
		 FROM interviews WHERE application_id = $1 ORDER BY scheduled_at`, id,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var i models.Interview
		if err := rows.Scan(
			&i.ID, &i.ApplicationID, &i.InterviewerID, &i.ScheduledAt,
			&i.Type, &i.Location, &i.Outcome, &i.Feedback,
			&i.CreatedAt, &i.UpdatedAt,
		); err != nil {
			return nil, err
		}
		app.Interviews = append(app.Interviews, i)
	}
	return &app, nil
}

func (r *Repository) UpdateApplicationStatus(ctx context.Context, id, status string, notes *string) (*models.Application, error) {
	var app models.Application
	err := r.db.QueryRow(ctx,
		`UPDATE applications
		 SET status = $2, notes = COALESCE($3, notes), updated_at = NOW()
		 WHERE id = $1
		 RETURNING id, vacancy_id, first_name, last_name, email, phone, cv_url, cover_letter,
		           status, notes, applied_at, updated_at`,
		id, status, notes,
	).Scan(
		&app.ID, &app.VacancyID, &app.FirstName, &app.LastName, &app.Email,
		&app.Phone, &app.CVUrl, &app.CoverLetter, &app.Status, &app.Notes,
		&app.AppliedAt, &app.UpdatedAt,
	)
	return &app, err
}

func (r *Repository) DeleteApplication(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM applications WHERE id = $1`, id)
	return err
}

// ─────────────────────────────────────────────
// INTERVIEWS
// ─────────────────────────────────────────────

func (r *Repository) CreateInterview(ctx context.Context, applicationID, interviewerID string, scheduledAt time.Time, iType string, location *string) (*models.Interview, error) {
	var i models.Interview
	err := r.db.QueryRow(ctx,
		`INSERT INTO interviews (application_id, interviewer_id, scheduled_at, type, location)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, application_id, interviewer_id, scheduled_at, type, location,
		           outcome, feedback, created_at, updated_at`,
		applicationID, interviewerID, scheduledAt, iType, location,
	).Scan(
		&i.ID, &i.ApplicationID, &i.InterviewerID, &i.ScheduledAt,
		&i.Type, &i.Location, &i.Outcome, &i.Feedback,
		&i.CreatedAt, &i.UpdatedAt,
	)
	return &i, err
}

func (r *Repository) UpdateInterview(ctx context.Context, id string, req UpdateInterviewRequest) (*models.Interview, error) {
	fields := []string{}
	args := []any{}
	n := 1

	if req.ScheduledAt != nil {
		fields = append(fields, fmt.Sprintf("scheduled_at = $%d", n))
		args = append(args, *req.ScheduledAt)
		n++
	}
	if req.Type != nil {
		fields = append(fields, fmt.Sprintf("type = $%d", n))
		args = append(args, *req.Type)
		n++
	}
	if req.Location != nil {
		fields = append(fields, fmt.Sprintf("location = $%d", n))
		args = append(args, *req.Location)
		n++
	}
	if req.Outcome != nil {
		fields = append(fields, fmt.Sprintf("outcome = $%d", n))
		args = append(args, *req.Outcome)
		n++
	}
	if req.Feedback != nil {
		fields = append(fields, fmt.Sprintf("feedback = $%d", n))
		args = append(args, *req.Feedback)
		n++
	}

	if len(fields) == 0 {
		var i models.Interview
		err := r.db.QueryRow(ctx,
			`SELECT id, application_id, interviewer_id, scheduled_at, type, location,
			        outcome, feedback, created_at, updated_at
			 FROM interviews WHERE id = $1`, id,
		).Scan(
			&i.ID, &i.ApplicationID, &i.InterviewerID, &i.ScheduledAt,
			&i.Type, &i.Location, &i.Outcome, &i.Feedback,
			&i.CreatedAt, &i.UpdatedAt,
		)
		return &i, err
	}

	fields = append(fields, "updated_at = NOW()")
	args = append(args, id)
	query := fmt.Sprintf(
		`UPDATE interviews SET %s WHERE id = $%d
		 RETURNING id, application_id, interviewer_id, scheduled_at, type, location,
		           outcome, feedback, created_at, updated_at`,
		strings.Join(fields, ", "), n,
	)
	var i models.Interview
	err := r.db.QueryRow(ctx, query, args...).Scan(
		&i.ID, &i.ApplicationID, &i.InterviewerID, &i.ScheduledAt,
		&i.Type, &i.Location, &i.Outcome, &i.Feedback,
		&i.CreatedAt, &i.UpdatedAt,
	)
	return &i, err
}

func (r *Repository) DeleteInterview(ctx context.Context, id string) error {
	_, err := r.db.Exec(ctx, `DELETE FROM interviews WHERE id = $1`, id)
	return err
}

// ─────────────────────────────────────────────
// HIRE
// ─────────────────────────────────────────────

func (r *Repository) HireApplicant(ctx context.Context, applicationID, branchID, passwordHash string, req HireRequest) (*HireResult, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	// Fetch applicant data
	var firstName, lastName, email string
	var phone *string
	err = tx.QueryRow(ctx,
		`SELECT first_name, last_name, email, phone FROM applications WHERE id = $1`, applicationID,
	).Scan(&firstName, &lastName, &email, &phone)
	if err != nil {
		return nil, errors.New("application not found")
	}

	currency := req.Currency
	if currency == "" {
		currency = "USD"
	}
	employmentType := req.EmploymentType
	if employmentType == "" {
		employmentType = "full_time"
	}

	// Create user account
	var userID string
	err = tx.QueryRow(ctx,
		`INSERT INTO users (branch_id, first_name, last_name, email, phone, password_hash, role, status)
		 VALUES ($1, $2, $3, $4, $5, $6, 'employee', 'active')
		 RETURNING id`,
		branchID, firstName, lastName, email, phone, passwordHash,
	).Scan(&userID)
	if err != nil {
		return nil, errors.New("failed to create user account: email may already be registered")
	}

	// Create employee record
	var employeeID string
	err = tx.QueryRow(ctx,
		`INSERT INTO employees (user_id, branch_id, employee_code, hourly_rate, currency, joining_date, designation, employment_type)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING id`,
		userID, branchID, req.EmployeeCode, req.HourlyRate, currency,
		req.JoiningDate, req.Designation, employmentType,
	).Scan(&employeeID)
	if err != nil {
		return nil, errors.New("failed to create employee record: employee code may already exist")
	}

	// Mark application as hired
	_, err = tx.Exec(ctx,
		`UPDATE applications SET status = 'hired', updated_at = NOW() WHERE id = $1`, applicationID,
	)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &HireResult{
		UserID:     userID,
		EmployeeID: employeeID,
		Email:      email,
		Message:    "candidate successfully hired and employee account created",
	}, nil
}
