package models

import "time"

type OfficeTiming struct {
	ID        string           `json:"id"`
	BranchID  string           `json:"branch_id"`
	Name      string           `json:"name"`
	IsActive  bool             `json:"is_active"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
	Days      []OfficeTimingDay `json:"days,omitempty"`
}

type OfficeTimingDay struct {
	ID             string  `json:"id"`
	OfficeTimingID string  `json:"office_timing_id"`
	DayOfWeek      int     `json:"day_of_week"`     // 0=Sun, 1=Mon ... 6=Sat
	IsWorkingDay   bool    `json:"is_working_day"`
	StartTime      *string `json:"start_time"`      // "09:00:00"
	EndTime        *string `json:"end_time"`        // "18:00:00"
	BreakMinutes   int     `json:"break_minutes"`
}
