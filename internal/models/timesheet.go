package models

import "time"

type ConsultantTimesheet struct {
	ID            string     `json:"id"`
	EmployeeID    string     `json:"employee_id"`
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
	UpdatedAt     time.Time  `json:"updated_at"`
}
