package report

// Request DTOs
type ReportRequest struct {
	Type      string `json:"type" validate:"required"` // daily, weekly, monthly
	StartDate string `json:"start_date" validate:"required"`
	EndDate   string `json:"end_date" validate:"required"`
}

// Response DTOs
type ReportResponse struct {
	Type           string      `json:"type"`
	StartDate      string      `json:"start_date"`
	EndDate        string      `json:"end_date"`
	TotalPatients  int         `json:"total_patients"`
	TotalBilling   float64     `json:"total_billing"`
	TotalAppoints  int         `json:"total_appointments"`
	Data           interface{} `json:"data,omitempty"`
}
