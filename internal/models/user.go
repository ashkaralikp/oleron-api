package models

import "time"

type User struct {
	ID           string     `json:"id"`
	BranchID     string     `json:"branch_id"`
	FirstName    string     `json:"first_name"`
	LastName     string     `json:"last_name"`
	Email        string     `json:"email"`
	Phone        *string    `json:"phone,omitempty"`
	PasswordHash string     `json:"-"`
	Role         string     `json:"role"`
	Status       string     `json:"status"`
	AvatarURL    *string    `json:"avatar_url,omitempty"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}
