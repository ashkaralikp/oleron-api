package models

import "time"

type BranchCalendar struct {
	ID        string    `json:"id"`
	BranchID  string    `json:"branch_id"`
	Date      time.Time `json:"date"`
	Type      string    `json:"type"`      // "holiday" or "working_day"
	Name      *string   `json:"name,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}
