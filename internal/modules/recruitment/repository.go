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
// ASSIGNABLE USERS
// ─────────────────────────────────────────────

type AssignableUser struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Role      string `json:"role"`
}

// FindAssignableUsers returns all active manager and regional_manager users.
// Branch filtering is intentionally omitted — any authorised user (manager+)
// can assign a vacancy to any manager/regional_manager in the organisation.
func (r *Repository) FindAssignableUsers(ctx context.Context, branchIDs []string) ([]AssignableUser, error) {
	query := `SELECT id, first_name, last_name, role FROM users
	          WHERE role IN ('manager', 'regional_manager') AND status = 'active'
	          ORDER BY first_name, last_name`

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []AssignableUser
	for rows.Next() {
		var u AssignableUser
		if err := rows.Scan(&u.ID, &u.FirstName, &u.LastName, &u.Role); err != nil {
			return nil, err
		}
		result = append(result, u)
	}
	return result, nil
}

// ─────────────────────────────────────────────
// PUBLIC VACANCIES
// ─────────────────────────────────────────────

// FindPublicVacancies returns open vacancies that still have unfilled positions and
// have not passed their deadline. Internal fields (branch_id, created_by) are excluded.
func (r *Repository) FindPublicVacancies(ctx context.Context) ([]PublicVacancy, error) {
	rows, err := r.db.Query(ctx, `
		SELECT v.id, v.title, v.department, v.description, v.requirements,
		       v.positions, v.deadline, v.created_at, v.updated_at,
		       v.positions - COUNT(a.id) FILTER (WHERE a.status = 'hired') AS available_positions
		FROM vacancies v
		LEFT JOIN applications a ON a.vacancy_id = v.id
		WHERE v.status = 'open'
		  AND (v.deadline IS NULL OR v.deadline >= CURRENT_DATE)
		GROUP BY v.id
		HAVING (v.positions - COUNT(a.id) FILTER (WHERE a.status = 'hired')) > 0
		ORDER BY v.created_at DESC
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []PublicVacancy
	for rows.Next() {
		var v PublicVacancy
		if err := rows.Scan(
			&v.ID, &v.Title, &v.Department, &v.Description, &v.Requirements,
			&v.Positions, &v.Deadline, &v.CreatedAt, &v.UpdatedAt, &v.AvailablePositions,
		); err != nil {
			return nil, err
		}
		result = append(result, v)
	}
	return result, nil
}

// ─────────────────────────────────────────────
// VACANCIES
// ─────────────────────────────────────────────

func (r *Repository) FindAllVacancies(ctx context.Context, branchIDs []string, assignedToUserID string) ([]models.Vacancy, error) {
	query := `SELECT v.id, v.branch_id, v.created_by,
	                 COALESCE(ARRAY_AGG(va.user_id) FILTER (WHERE va.user_id IS NOT NULL), '{}') AS assigned_to,
	                 COALESCE(ARRAY_AGG(u.first_name || ' ' || u.last_name) FILTER (WHERE u.id IS NOT NULL), '{}') AS assigned_to_names,
	                 v.title, v.department,
	                 v.description, v.requirements, v.positions, v.status, v.deadline,
	                 v.created_at, v.updated_at,
	                 (SELECT COUNT(*) FROM applications a WHERE a.vacancy_id = v.id) AS application_count
	          FROM vacancies v
	          LEFT JOIN vacancy_assignees va ON va.vacancy_id = v.id
	          LEFT JOIN users u ON u.id = va.user_id`
	args := []any{}
	where := []string{}
	n := 1

	// Branch filter: users see vacancies in their branch scope
	hasBranchFilter := len(branchIDs) > 0
	if hasBranchFilter {
		where = append(where, fmt.Sprintf("v.branch_id::text = ANY($%d)", n))
		args = append(args, branchIDs)
		n++
	}
	// Assignment filter: users also see vacancies they are assigned to (even in other branches)
	if assignedToUserID != "" {
		assignClause := fmt.Sprintf(`EXISTS (SELECT 1 FROM vacancy_assignees va2 WHERE va2.vacancy_id = v.id AND va2.user_id::text = $%d)`, n)
		args = append(args, assignedToUserID)
		n++
		if hasBranchFilter {
			// Combine branch + assignment with OR so assigned vacancies are visible regardless of branch
			lastIdx := len(where) - 1
			existingBranchClause := where[lastIdx]
			where[lastIdx] = fmt.Sprintf("(%s OR %s)", existingBranchClause, assignClause)
		} else {
			where = append(where, assignClause)
		}
	}
	if len(where) > 0 {
		query += ` WHERE ` + strings.Join(where, " AND ")
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
			&v.ID, &v.BranchID, &v.CreatedBy,
			&v.AssignedTo, &v.AssignedToNames,
			&v.Title, &v.Department,
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
		`SELECT v.id, v.branch_id, v.created_by,
		        COALESCE(ARRAY_AGG(va.user_id) FILTER (WHERE va.user_id IS NOT NULL), '{}') AS assigned_to,
		        COALESCE(ARRAY_AGG(u.first_name || ' ' || u.last_name) FILTER (WHERE u.id IS NOT NULL), '{}') AS assigned_to_names,
		        v.title, v.department,
		        v.description, v.requirements, v.positions, v.status, v.deadline,
		        v.created_at, v.updated_at,
		        (SELECT COUNT(*) FROM applications a WHERE a.vacancy_id = v.id) AS application_count
		 FROM vacancies v
		 LEFT JOIN vacancy_assignees va ON va.vacancy_id = v.id
		 LEFT JOIN users u ON u.id = va.user_id
		 WHERE v.id = $1
		 GROUP BY v.id`, id,
	).Scan(
		&v.ID, &v.BranchID, &v.CreatedBy,
		&v.AssignedTo, &v.AssignedToNames,
		&v.Title, &v.Department,
		&v.Description, &v.Requirements, &v.Positions, &v.Status, &v.Deadline,
		&v.CreatedAt, &v.UpdatedAt, &v.ApplicationCount,
	)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *Repository) CreateVacancy(ctx context.Context, branchID, createdBy string, assignedTo []string, req CreateVacancyRequest) (*models.Vacancy, error) {
	positions := req.Positions
	if positions < 1 {
		positions = 1
	}

	var vacancyID string
	err := r.db.QueryRow(ctx,
		`INSERT INTO vacancies (branch_id, created_by, title, department, description, requirements, positions, deadline)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		 RETURNING id`,
		branchID, createdBy, req.Title, req.Department, req.Description, req.Requirements, positions, req.Deadline,
	).Scan(&vacancyID)
	if err != nil {
		return nil, err
	}

	// Insert assignees
	for _, uid := range assignedTo {
		if _, err := r.db.Exec(ctx,
			`INSERT INTO vacancy_assignees (vacancy_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
			vacancyID, uid,
		); err != nil {
			return nil, err
		}
	}

	// Fetch and return the full vacancy
	return r.FindVacancyByID(ctx, vacancyID)
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

	// Update scalar fields
	if len(fields) > 0 {
		fields = append(fields, "updated_at = NOW()")
		args = append(args, id)
		query := fmt.Sprintf(
			`UPDATE vacancies SET %s WHERE id = $%d`,
			strings.Join(fields, ", "), n,
		)
		if _, err := r.db.Exec(ctx, query, args...); err != nil {
			return nil, err
		}
	}

	// Update assignees (replace all)
	if req.AssignedTo != nil {
		if _, err := r.db.Exec(ctx, `DELETE FROM vacancy_assignees WHERE vacancy_id = $1`, id); err != nil {
			return nil, err
		}
		for _, uid := range req.AssignedTo {
			if _, err := r.db.Exec(ctx,
				`INSERT INTO vacancy_assignees (vacancy_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
				id, uid,
			); err != nil {
				return nil, err
			}
		}
	}

	return r.FindVacancyByID(ctx, id)
}

func (r *Repository) UpdateVacancyStatus(ctx context.Context, id, status string) (*models.Vacancy, error) {
	_, err := r.db.Exec(ctx,
		`UPDATE vacancies SET status = $2, updated_at = NOW() WHERE id = $1`,
		id, status,
	)
	if err != nil {
		return nil, err
	}
	return r.FindVacancyByID(ctx, id)
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

func (r *Repository) IsUserAssignedToVacancy(ctx context.Context, vacancyID, userID string) (bool, error) {
	var count int
	err := r.db.QueryRow(ctx,
		`SELECT COUNT(*) FROM vacancy_assignees WHERE vacancy_id = $1 AND user_id = $2`,
		vacancyID, userID,
	).Scan(&count)
	return count > 0, err
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

func (r *Repository) FindVacancyIDByApplicationID(ctx context.Context, applicationID string) (string, error) {
	var vacancyID string
	err := r.db.QueryRow(ctx,
		`SELECT vacancy_id FROM applications WHERE id = $1`, applicationID,
	).Scan(&vacancyID)
	return vacancyID, err
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

func (r *Repository) FindApplicationIDByInterviewID(ctx context.Context, interviewID string) (string, error) {
	var applicationID string
	err := r.db.QueryRow(ctx,
		`SELECT application_id FROM interviews WHERE id = $1`, interviewID,
	).Scan(&applicationID)
	return applicationID, err
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

	app.Interviews = []models.Interview{} // ensure empty array, never nil
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
	// Update application status to interview_scheduled and create the interview in a single round-trip
	var i models.Interview
	err := r.db.QueryRow(ctx,
		`WITH updated AS (
		    UPDATE applications SET status = 'interview_scheduled'::application_status, updated_at = NOW()
		    WHERE id = $1 AND status <> 'hired' AND status <> 'withdrawn' AND status <> 'rejected'
		)
		 INSERT INTO interviews (application_id, interviewer_id, scheduled_at, type, location)
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
