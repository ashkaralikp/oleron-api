package models

import "time"

type Patient struct {
	ID          string    `json:"id"`
	BranchID    string    `json:"branch_id"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	Phone       string    `json:"phone"`
	Email       *string   `json:"email,omitempty"`
	DateOfBirth *string   `json:"date_of_birth,omitempty"`
	Gender      *string   `json:"gender,omitempty"`
	Address     *string   `json:"address,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
