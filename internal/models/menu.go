package models

import "time"

type MenuPermissions struct {
	CanView   bool `json:"can_view"`
	CanCreate bool `json:"can_create"`
	CanEdit   bool `json:"can_edit"`
	CanDelete bool `json:"can_delete"`
}

type Menu struct {
	ID          string           `json:"id"`
	ParentID    *string          `json:"parent_id"`
	Label       string           `json:"label"`
	Path        *string          `json:"path,omitempty"`
	Resource    *string          `json:"resource,omitempty"`
	SortOrder   int              `json:"sort_order"`
	IsActive    bool             `json:"is_active"`
	Permissions *MenuPermissions `json:"permissions,omitempty"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
	Children    []Menu           `json:"children,omitempty"`
}
