package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

type DailyEntry struct {
	Day     int     `json:"day"`
	Support float64 `json:"support"`
	OT      float64 `json:"ot"`
}

type DailyEntries []DailyEntry

// Scan implements sql.Scanner interface for reading from DB
func (d *DailyEntries) Scan(value interface{}) error {
	if value == nil {
		*d = nil
		return nil
	}
	bytes := value.([]byte)
	return json.Unmarshal(bytes, &d)
}

// Value implements driver.Valuer interface for writing to DB
func (d DailyEntries) Value() (driver.Value, error) {
	if len(d) == 0 {
		return nil, nil
	}
	return json.Marshal(d)
}

type ConsultantTimesheet struct {
	ID            string       `json:"id"`
	EmployeeID    string       `json:"employee_id"`
	Year          int          `json:"year"`
	Month         int          `json:"month"`
	SupportHours  float64      `json:"support_hours"`
	OvertimeHours float64      `json:"overtime_hours"`
	Notes         *string      `json:"notes,omitempty"`
	Status        string       `json:"status"`
	ReviewerID    *string      `json:"reviewer_id,omitempty"`
	ReviewNote    *string      `json:"review_note,omitempty"`
	ReviewedAt    *time.Time   `json:"reviewed_at,omitempty"`
	SubmittedAt   time.Time    `json:"submitted_at"`
	UpdatedAt     time.Time    `json:"updated_at"`
	Details       DailyEntries `json:"details,omitempty"`
}
