package models

import "time"

type Branch struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	Address   *string   `json:"address,omitempty"`
	Phone     *string   `json:"phone,omitempty"`
	Email     *string   `json:"email,omitempty"`
	LogoURL   *string   `json:"logo_url,omitempty"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
