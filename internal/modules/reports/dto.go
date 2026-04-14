package reports

// AttendanceFilter holds optional query params for filtering attendance records
type AttendanceFilter struct {
	BranchID  string // injected from JWT for manager/admin
	DateFrom  string // YYYY-MM-DD
	DateTo    string // YYYY-MM-DD
	UserID    string // filter by specific employee
	Status    string // present, absent, half_day, late, on_leave
}
