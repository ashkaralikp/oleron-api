package timesheet

import "time"

// ── Estimate ─────────────────────────────────────────────────────

type EstimateRequest struct {
	Year          int     `json:"year" validate:"required,min=2000,max=2100"`
	Month         int     `json:"month" validate:"required,min=1,max=12"`
	SupportHours  float64 `json:"support_hours" validate:"min=0"`
	OvertimeHours float64 `json:"overtime_hours" validate:"min=0"`
}

type EstimateResponse struct {
	Year               int     `json:"year"`
	Month              int     `json:"month"`
	WholeMonthHours    float64 `json:"whole_month_hours"`
	SupportHours       float64 `json:"support_hours"`
	OvertimeHours      float64 `json:"overtime_hours"`
	FixedMonthlySalary float64 `json:"fixed_monthly_salary"`
	OTRate             float64 `json:"ot_rate"`
	HourlyRate         float64 `json:"hourly_rate"`
	Scenario           string  `json:"scenario"`
	EstimatedPay       float64 `json:"estimated_pay"`
	Currency           string  `json:"currency"`
}

// ── Submit (consultant) ───────────────────────────────────────────

type SubmitRequest struct {
	Year          int     `json:"year" validate:"required,min=2000,max=2100"`
	Month         int     `json:"month" validate:"required,min=1,max=12"`
	SupportHours  float64 `json:"support_hours" validate:"min=0"`
	OvertimeHours float64 `json:"overtime_hours" validate:"min=0"`
	Notes         string  `json:"notes"`
}

// ── Review (admin / manager) ──────────────────────────────────────

type ReviewRequest struct {
	Status     string `json:"status" validate:"required,oneof=approved rejected"`
	ReviewNote string `json:"review_note"`
}

// ── Shared response ───────────────────────────────────────────────

type TimesheetResponse struct {
	ID            string     `json:"id"`
	EmployeeID    string     `json:"employee_id"`
	EmployeeCode  string     `json:"employee_code"`
	FirstName     string     `json:"first_name"`
	LastName      string     `json:"last_name"`
	Year          int        `json:"year"`
	Month         int        `json:"month"`
	SupportHours  float64    `json:"support_hours"`
	OvertimeHours float64    `json:"overtime_hours"`
	Notes         *string    `json:"notes,omitempty"`
	Status        string     `json:"status"`
	ReviewerID    *string    `json:"reviewer_id,omitempty"`
	ReviewNote    *string    `json:"review_note,omitempty"`
	ReviewedAt    *time.Time `json:"reviewed_at,omitempty"`
	SubmittedAt   time.Time  `json:"submitted_at"`
}
