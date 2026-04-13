package models

import "time"

type Doctor struct {
	ID             string    `json:"id"`
	BranchID       string    `json:"branch_id"`
	UserID         string    `json:"user_id"`
	Specialization string    `json:"specialization"`
	LicenseNo      string    `json:"license_no"`
	IsAvailable    bool      `json:"is_available"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}
