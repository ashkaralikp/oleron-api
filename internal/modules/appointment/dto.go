package appointment

// Request DTOs
type CreateAppointmentRequest struct {
	PatientID string `json:"patient_id" validate:"required"`
	DoctorID  string `json:"doctor_id" validate:"required"`
	Date      string `json:"date" validate:"required"`
	StartTime string `json:"start_time" validate:"required"`
	EndTime   string `json:"end_time" validate:"required"`
	Notes     string `json:"notes"`
}

type UpdateAppointmentRequest struct {
	PatientID string `json:"patient_id"`
	DoctorID  string `json:"doctor_id"`
	Date      string `json:"date"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Status    string `json:"status"`
	Notes     string `json:"notes"`
}

// Response DTOs
type AppointmentResponse struct {
	ID        string `json:"id"`
	PatientID string `json:"patient_id"`
	DoctorID  string `json:"doctor_id"`
	Date      string `json:"date"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
	Status    string `json:"status"`
	Notes     string `json:"notes"`
	CreatedAt string `json:"created_at"`
}
