package models

import "time"

type Appointment struct {
	ID          string    `json:"id"`
	BranchID    string    `json:"branch_id"`
	PatientID   string    `json:"patient_id"`
	DoctorID    string    `json:"doctor_id"`
	Date        string    `json:"date"`
	StartTime   string    `json:"start_time"`
	EndTime     string    `json:"end_time"`
	Status      string    `json:"status"` // scheduled, confirmed, completed, cancelled, no_show
	Notes       *string   `json:"notes,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
