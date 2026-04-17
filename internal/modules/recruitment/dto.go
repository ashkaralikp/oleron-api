package recruitment

// ─────────────────────────────────────────────
// VACANCY
// ─────────────────────────────────────────────

type CreateVacancyRequest struct {
	Title        string  `json:"title" validate:"required,max=150"`
	Department   *string `json:"department"`
	Description  *string `json:"description"`
	Requirements *string `json:"requirements"`
	Positions    int     `json:"positions"` // defaults to 1 if 0
	Deadline     *string `json:"deadline"`  // YYYY-MM-DD, optional
}

type UpdateVacancyRequest struct {
	Title        *string `json:"title" validate:"omitempty,max=150"`
	Department   *string `json:"department"`
	Description  *string `json:"description"`
	Requirements *string `json:"requirements"`
	Positions    *int    `json:"positions" validate:"omitempty,min=1"`
	Deadline     *string `json:"deadline"`
}

type UpdateVacancyStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=draft open closed cancelled"`
}

// ─────────────────────────────────────────────
// APPLICATION
// ─────────────────────────────────────────────

// ApplyRequest is the public-facing body used by candidates.
type ApplyRequest struct {
	FirstName   string  `json:"first_name" validate:"required,max=100"`
	LastName    string  `json:"last_name" validate:"required,max=100"`
	Email       string  `json:"email" validate:"required,email"`
	Phone       *string `json:"phone"`
	CVUrl       *string `json:"cv_url"`
	CoverLetter *string `json:"cover_letter"`
}

type UpdateApplicationStatusRequest struct {
	Status string  `json:"status" validate:"required,oneof=shortlisted rejected interview_scheduled hired withdrawn"`
	Notes  *string `json:"notes"`
}

// ─────────────────────────────────────────────
// INTERVIEW
// ─────────────────────────────────────────────

type CreateInterviewRequest struct {
	InterviewerID string  `json:"interviewer_id" validate:"required"`
	ScheduledAt   string  `json:"scheduled_at" validate:"required"` // RFC3339
	Type          string  `json:"type" validate:"required,oneof=phone video in_person"`
	Location      *string `json:"location"`
}

type UpdateInterviewRequest struct {
	ScheduledAt *string `json:"scheduled_at"`
	Type        *string `json:"type" validate:"omitempty,oneof=phone video in_person"`
	Location    *string `json:"location"`
	Outcome     *string `json:"outcome" validate:"omitempty,oneof=pending passed failed no_show"`
	Feedback    *string `json:"feedback"`
}

// ─────────────────────────────────────────────
// HIRE
// ─────────────────────────────────────────────

type HireRequest struct {
	EmployeeCode   string  `json:"employee_code" validate:"required"`
	HourlyRate     float64 `json:"hourly_rate" validate:"min=0"`
	Currency       string  `json:"currency"`        // defaults to USD
	JoiningDate    string  `json:"joining_date" validate:"required"` // YYYY-MM-DD
	Designation    *string `json:"designation"`
	EmploymentType string  `json:"employment_type"` // defaults to full_time
	TempPassword   string  `json:"temp_password" validate:"required,min=8"`
}

type HireResult struct {
	UserID     string `json:"user_id"`
	EmployeeID string `json:"employee_id"`
	Email      string `json:"email"`
	Message    string `json:"message"`
}
