package workorder

type UpsertAssetRequest struct {
	SignerName  string  `json:"signer_name" validate:"required,max=150"`
	SignerTitle *string `json:"signer_title" validate:"omitempty,max=100"`
}

type WorkOrderItemRequest struct {
	LineNo      int     `json:"line_no" validate:"omitempty,min=1"`
	Description string  `json:"description" validate:"required"`
	Amount      float64 `json:"amount" validate:"min=0"`
}

type CreateWorkOrderRequest struct {
	BranchID       string                 `json:"branch_id"`
	WorkOrderDate  string                 `json:"work_order_date" validate:"required"` // YYYY-MM-DD
	CompanyName    string                 `json:"company_name" validate:"omitempty,max=150"`
	CompanyAddress *string                `json:"company_address"`
	CompanyPhone   *string                `json:"company_phone" validate:"omitempty,max=50"`
	CompanyFax     *string                `json:"company_fax" validate:"omitempty,max=50"`
	CompanyEmail   *string                `json:"company_email" validate:"omitempty,email,max=150"`
	CompanyWebsite *string                `json:"company_website" validate:"omitempty,max=150"`
	CompanyLogoURL *string                `json:"company_logo_url"`
	BillToName     string                 `json:"bill_to_name" validate:"required,max=150"`
	BillToAddress  *string                `json:"bill_to_address"`
	BillToEmail    *string                `json:"bill_to_email" validate:"omitempty,email,max=150"`
	JobDetails     string                 `json:"job_details" validate:"required"`
	Currency       string                 `json:"currency" validate:"omitempty,len=3"`
	SubTotalAmount *float64               `json:"sub_total_amount" validate:"omitempty,min=0"`
	TotalAmount    *float64               `json:"total_amount" validate:"omitempty,min=0"`
	Status         string                 `json:"status" validate:"omitempty,oneof=draft issued cancelled"`
	Items          []WorkOrderItemRequest `json:"items" validate:"required,min=1,dive"`
}

type UpdateWorkOrderRequest struct {
	WorkOrderDate  *string                `json:"work_order_date"` // YYYY-MM-DD
	CompanyName    *string                `json:"company_name" validate:"omitempty,max=150"`
	CompanyAddress *string                `json:"company_address"`
	CompanyPhone   *string                `json:"company_phone" validate:"omitempty,max=50"`
	CompanyFax     *string                `json:"company_fax" validate:"omitempty,max=50"`
	CompanyEmail   *string                `json:"company_email" validate:"omitempty,email,max=150"`
	CompanyWebsite *string                `json:"company_website" validate:"omitempty,max=150"`
	CompanyLogoURL *string                `json:"company_logo_url"`
	BillToName     *string                `json:"bill_to_name" validate:"omitempty,max=150"`
	BillToAddress  *string                `json:"bill_to_address"`
	BillToEmail    *string                `json:"bill_to_email" validate:"omitempty,email,max=150"`
	JobDetails     *string                `json:"job_details"`
	Currency       *string                `json:"currency" validate:"omitempty,len=3"`
	SubTotalAmount *float64               `json:"sub_total_amount" validate:"omitempty,min=0"`
	TotalAmount    *float64               `json:"total_amount" validate:"omitempty,min=0"`
	Status         *string                `json:"status" validate:"omitempty,oneof=draft issued cancelled"`
	PDFURL         *string                `json:"pdf_url"`
	Items          []WorkOrderItemRequest `json:"items" validate:"omitempty,min=1,dive"`
}

type UpdateWorkOrderStatusRequest struct {
	Status string  `json:"status" validate:"required,oneof=draft issued cancelled"`
	PDFURL *string `json:"pdf_url"`
}
