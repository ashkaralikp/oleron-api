package attendance

import "time"

// DayTiming holds the resolved schedule for an employee on a given day.
type DayTiming struct {
	IsWorkingDay bool
	StartTime    *string // "09:00:00"
	EndTime      *string // "18:00:00"
	BreakMinutes int
}

// PunchResult is returned after a punch event.
type PunchResult struct {
	Action      string     `json:"action"`       // "punch_in" or "punch_out"
	ID          string     `json:"id"`
	UserID      string     `json:"user_id"`
	WorkDate    time.Time  `json:"work_date"`
	PunchIn     *time.Time `json:"punch_in"`
	PunchOut    *time.Time `json:"punch_out"`
	WorkHours   *float64   `json:"work_hours"`
	Status      string     `json:"status"`
	Notes       *string    `json:"notes,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// TodayResult is returned by GET /attendance/today.
type TodayResult struct {
	PunchedIn  bool       `json:"punched_in"`
	PunchedOut bool       `json:"punched_out"`
	ID         *string    `json:"id,omitempty"`
	UserID     *string    `json:"user_id,omitempty"`
	WorkDate   *time.Time `json:"work_date,omitempty"`
	PunchIn    *time.Time `json:"punch_in,omitempty"`
	PunchOut   *time.Time `json:"punch_out,omitempty"`
	WorkHours  *float64   `json:"work_hours,omitempty"`
	Status     *string    `json:"status,omitempty"`
	Notes      *string    `json:"notes,omitempty"`
}
