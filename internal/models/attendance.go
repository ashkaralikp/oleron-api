package models

import "time"

type Attendance struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	WorkDate  time.Time  `json:"work_date"`
	PunchIn   *time.Time `json:"punch_in,omitempty"`
	PunchOut  *time.Time `json:"punch_out,omitempty"`
	WorkHours *float64   `json:"work_hours,omitempty"`
	Status    string     `json:"status"`
	Notes     *string    `json:"notes,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`

	// Joined from users + employees
	FirstName    string  `json:"first_name"`
	LastName     string  `json:"last_name"`
	Email        string  `json:"email"`
	EmployeeCode *string `json:"employee_code,omitempty"`
	BranchID     string  `json:"branch_id"`
}
