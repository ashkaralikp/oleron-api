package payroll

type GeneratePayrollRequest struct {
	PeriodFrom string  `json:"period_from" validate:"required"` // YYYY-MM-DD
	PeriodTo   string  `json:"period_to"   validate:"required"` // YYYY-MM-DD
	Currency   string  `json:"currency"    validate:"omitempty,len=3"`
	Notes      *string `json:"notes"`
}

type UpdateStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=approved paid"`
}
