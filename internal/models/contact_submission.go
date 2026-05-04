package models

import "time"

type ContactSubmission struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Company   *string   `json:"company,omitempty"`
	Email     string    `json:"email"`
	Phone     *string   `json:"phone,omitempty"`
	Category  *string   `json:"category,omitempty"`
	Message   string    `json:"message"`
	Status    string    `json:"status"`
	IPAddress *string   `json:"ip_address,omitempty"`
	UserAgent *string   `json:"user_agent,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
