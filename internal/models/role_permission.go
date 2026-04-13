package models

import "time"

type RolePermission struct {
	ID        string    `json:"id"`
	Role      string    `json:"role"`
	Resource  string    `json:"resource"`
	CanView   bool      `json:"can_view"`
	CanCreate bool      `json:"can_create"`
	CanEdit   bool      `json:"can_edit"`
	CanDelete bool      `json:"can_delete"`
	CreatedAt time.Time `json:"created_at"`
}
