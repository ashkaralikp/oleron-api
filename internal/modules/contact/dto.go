package contact

type CreateSubmissionRequest struct {
	Name     string  `json:"name" validate:"required,max=150"`
	Company  *string `json:"company" validate:"omitempty,max=150"`
	Email    string  `json:"email" validate:"required,email,max=255"`
	Phone    *string `json:"phone" validate:"omitempty,max=50"`
	Category *string `json:"category" validate:"omitempty,max=100"`
	Message  string  `json:"message" validate:"required,max=5000"`
}
