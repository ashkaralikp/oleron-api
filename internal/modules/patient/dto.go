package patient

// Request DTOs
type CreatePatientRequest struct {
	FirstName   string `json:"first_name" validate:"required"`
	LastName    string `json:"last_name" validate:"required"`
	Phone       string `json:"phone" validate:"required"`
	Email       string `json:"email"`
	DateOfBirth string `json:"date_of_birth"`
	Address     string `json:"address"`
	Gender      string `json:"gender"`
}

type UpdatePatientRequest struct {
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	DateOfBirth string `json:"date_of_birth"`
	Address     string `json:"address"`
	Gender      string `json:"gender"`
}

// Response DTOs
type PatientResponse struct {
	ID          string `json:"id"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Phone       string `json:"phone"`
	Email       string `json:"email"`
	DateOfBirth string `json:"date_of_birth"`
	Gender      string `json:"gender"`
	Address     string `json:"address"`
	CreatedAt   string `json:"created_at"`
}