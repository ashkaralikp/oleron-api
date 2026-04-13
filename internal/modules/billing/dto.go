package billing

// Request DTOs
type CreateBillingRequest struct {
	PatientID   string  `json:"patient_id" validate:"required"`
	TotalAmount float64 `json:"total_amount" validate:"required"`
	PaidAmount  float64 `json:"paid_amount"`
	Notes       string  `json:"notes"`
}

type UpdateBillingRequest struct {
	TotalAmount float64 `json:"total_amount"`
	PaidAmount  float64 `json:"paid_amount"`
	Status      string  `json:"status"`
	Notes       string  `json:"notes"`
}

// Response DTOs
type BillingResponse struct {
	ID          string  `json:"id"`
	PatientID   string  `json:"patient_id"`
	InvoiceNo   string  `json:"invoice_no"`
	TotalAmount float64 `json:"total_amount"`
	PaidAmount  float64 `json:"paid_amount"`
	Status      string  `json:"status"`
	Notes       string  `json:"notes"`
	CreatedBy   string  `json:"created_by"`
	CreatedAt   string  `json:"created_at"`
}
