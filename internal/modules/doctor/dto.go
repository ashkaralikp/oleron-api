package doctor

// Request DTOs
type CreateDoctorRequest struct {
	UserID         string `json:"user_id" validate:"required"`
	Specialization string `json:"specialization" validate:"required"`
	LicenseNo      string `json:"license_no" validate:"required"`
}

type UpdateDoctorRequest struct {
	Specialization string `json:"specialization"`
	LicenseNo      string `json:"license_no"`
	IsAvailable    *bool  `json:"is_available"`
}

// Response DTOs
type DoctorResponse struct {
	ID             string `json:"id"`
	UserID         string `json:"user_id"`
	Specialization string `json:"specialization"`
	LicenseNo      string `json:"license_no"`
	IsAvailable    bool   `json:"is_available"`
	CreatedAt      string `json:"created_at"`
}
