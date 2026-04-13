package models

import "time"

type Billing struct {
	ID          string    `json:"id"`
	BranchID    string    `json:"branch_id"`
	PatientID   string    `json:"patient_id"`
	InvoiceNo   string    `json:"invoice_no"`
	TotalAmount float64   `json:"total_amount"`
	PaidAmount  float64   `json:"paid_amount"`
	Status      string    `json:"status"` // pending, paid, partial, cancelled
	Notes       *string   `json:"notes,omitempty"`
	CreatedBy   string    `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
