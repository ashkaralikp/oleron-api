package calendar

type CreateCalendarEntryRequest struct {
	Date string  `json:"date" validate:"required"` // YYYY-MM-DD
	Type string  `json:"type" validate:"required,oneof=holiday working_day"`
	Name *string `json:"name"` // e.g. "Christmas Day", "Makeup Saturday"
}

type UpdateCalendarEntryRequest struct {
	Type string  `json:"type" validate:"required,oneof=holiday working_day"`
	Name *string `json:"name"`
}

// CalendarRangeFilter is used for GET /branch-calendar?from=&to=
type CalendarRangeFilter struct {
	BranchID string
	From     string // YYYY-MM-DD
	To       string // YYYY-MM-DD
	Type     string // optional: "holiday" or "working_day"
}
