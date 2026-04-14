package schedule

// DayInput represents one day's schedule within an office timing.
type DayInput struct {
	DayOfWeek    int     `json:"day_of_week" validate:"min=0,max=6"`   // 0=Sun ... 6=Sat
	IsWorkingDay bool    `json:"is_working_day"`
	StartTime    *string `json:"start_time"`    // "09:00:00" — required when is_working_day=true
	EndTime      *string `json:"end_time"`      // "18:00:00" — required when is_working_day=true
	BreakMinutes int     `json:"break_minutes"` // optional, default 0
}

type CreateOfficeTimingRequest struct {
	Name string     `json:"name" validate:"required,min=2,max=100"`
	Days []DayInput `json:"days" validate:"required,min=1,dive"`
}

type UpdateOfficeTimingRequest struct {
	Name string     `json:"name" validate:"required,min=2,max=100"`
	Days []DayInput `json:"days" validate:"required,min=1,dive"`
}
