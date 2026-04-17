package models

import "time"

type PayrollRun struct {
	ID          string         `json:"id"`
	BranchID    string         `json:"branch_id"`
	PeriodFrom  time.Time      `json:"period_from"`
	PeriodTo    time.Time      `json:"period_to"`
	GeneratedBy string         `json:"generated_by"`
	Status      string         `json:"status"`   // draft, approved, paid
	TotalAmount float64        `json:"total_amount"`
	Currency    string         `json:"currency"`
	Notes       *string        `json:"notes,omitempty"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	Items       []PayrollItem  `json:"items,omitempty"`
}

type PayrollItem struct {
	ID           string    `json:"id"`
	PayrollRunID string    `json:"payroll_run_id"`
	EmployeeID   string    `json:"employee_id"`
	UserID       string    `json:"user_id"`
	// Joined from users + employees
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	Email        string    `json:"email"`
	EmployeeCode string    `json:"employee_code"`
	// Attendance summary
	WorkingDays  int       `json:"working_days"`  // expected working days in period
	PresentDays  int       `json:"present_days"`
	AbsentDays   int       `json:"absent_days"`
	LeaveDays    int       `json:"leave_days"`
	TotalHours   float64   `json:"total_hours"`
	// Pay
	HourlyRate   float64   `json:"hourly_rate"`
	Currency     string    `json:"currency"`
	GrossPay     float64   `json:"gross_pay"`
	Deductions   float64   `json:"deductions"`
	NetPay       float64   `json:"net_pay"`
	CreatedAt    time.Time `json:"created_at"`
}
