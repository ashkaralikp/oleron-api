package models

import "time"

type Employee struct {
	ID             string     `json:"id"`
	UserID         string     `json:"user_id"`
	BranchID        string     `json:"branch_id"`
	OfficeTimingID  *string    `json:"office_timing_id,omitempty"`
	ManagerID       *string    `json:"manager_id,omitempty"`
	EmployeeCode   string     `json:"employee_code"`
	Designation    *string    `json:"designation,omitempty"`
	EmploymentType string     `json:"employment_type"`
	HourlyRate     *float64   `json:"hourly_rate,omitempty"`
	Currency       string     `json:"currency"`
	JoiningDate    time.Time  `json:"joining_date"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`

	// Joined from users table
	FirstName string  `json:"first_name"`
	LastName  string  `json:"last_name"`
	Email     string  `json:"email"`
	Phone     *string `json:"phone,omitempty"`
	Status    string  `json:"status"`
	AvatarURL *string `json:"avatar_url,omitempty"`
}
