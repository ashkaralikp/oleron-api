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

type CreateInvoiceItemRequest struct {
	LineNo          int      `json:"line_no" validate:"omitempty,min=1"`
	WorkOrderItemID *string  `json:"work_order_item_id"`
	Description     string   `json:"description" validate:"omitempty"`
	Quantity        *float64 `json:"quantity" validate:"omitempty,gt=0"`
	SACCode         *string  `json:"sac_code" validate:"omitempty,max=20"`
	UnitAmount      *float64 `json:"unit_amount" validate:"omitempty,min=0"`
	TaxAmount       *float64 `json:"tax_amount" validate:"omitempty,min=0"`
	TotalAmount     *float64 `json:"total_amount" validate:"omitempty,min=0"`
}

type CreateInvoiceRequest struct {
	InvoiceDate      string                     `json:"invoice_date" validate:"omitempty"` // YYYY-MM-DD
	DueDate          *string                    `json:"due_date" validate:"omitempty"`     // YYYY-MM-DD
	InvoiceTitle     *string                    `json:"invoice_title" validate:"omitempty,max=100"`
	SupplyNote       *string                    `json:"supply_note"`
	SellerName       *string                    `json:"seller_name" validate:"omitempty,max=150"`
	SellerAddress    *string                    `json:"seller_address"`
	SellerEmail      *string                    `json:"seller_email" validate:"omitempty,email,max=150"`
	SellerPhone      *string                    `json:"seller_phone" validate:"omitempty,max=50"`
	SellerGSTIN      *string                    `json:"seller_gstin" validate:"omitempty,max=30"`
	SellerLogoURL    *string                    `json:"seller_logo_url"`
	BillToName       *string                    `json:"bill_to_name" validate:"omitempty,max=150"`
	BillToAddress    *string                    `json:"bill_to_address"`
	BillToEmail      *string                    `json:"bill_to_email" validate:"omitempty,email,max=150"`
	BillToPhone      *string                    `json:"bill_to_phone" validate:"omitempty,max=50"`
	BillToWebsite    *string                    `json:"bill_to_website" validate:"omitempty,max=150"`
	Currency         *string                    `json:"currency" validate:"omitempty,len=3"`
	TaxAmount        *float64                   `json:"tax_amount" validate:"omitempty,min=0"`
	AdditionalAmount *float64                   `json:"additional_amount" validate:"omitempty,min=0"`
	LUTOrderNumber   *string                    `json:"lut_order_number" validate:"omitempty,max=100"`
	ARNNumber        *string                    `json:"arn_number" validate:"omitempty,max=100"`
	Notes            *string                    `json:"notes"`
	SignerName       *string                    `json:"signer_name" validate:"omitempty,max=150"`
	SignerTitle      *string                    `json:"signer_title" validate:"omitempty,max=100"`
	SignatureURL     *string                    `json:"signature_url"`
	SealURL          *string                    `json:"seal_url"`
	PDFURL           *string                    `json:"pdf_url"`
	Status           string                     `json:"status" validate:"omitempty,oneof=draft issued"`
	Items            []CreateInvoiceItemRequest `json:"items" validate:"omitempty,dive"`
}

type UpdateInvoiceStatusRequest struct {
	Status string  `json:"status" validate:"required,oneof=draft issued cancelled void"`
	PDFURL *string `json:"pdf_url"`
}

type CreateInvoicePaymentRequest struct {
	PaymentDate       string  `json:"payment_date" validate:"omitempty"` // YYYY-MM-DD
	Amount            float64 `json:"amount" validate:"required,gt=0"`
	Currency          string  `json:"currency" validate:"omitempty,len=3"`
	Method            string  `json:"method" validate:"required,oneof=bank_transfer cash cheque credit_card debit_card upi other"`
	OtherMethod       *string `json:"other_method" validate:"omitempty,max=100"`
	ReferenceNo       *string `json:"reference_no" validate:"omitempty,max=150"`
	PayerName         *string `json:"payer_name" validate:"omitempty,max=150"`
	PayerAccountLast4 *string `json:"payer_account_last4" validate:"omitempty,max=10"`
	BankName          *string `json:"bank_name" validate:"omitempty,max=150"`
	Status            string  `json:"status" validate:"omitempty,oneof=pending confirmed failed reversed rejected"`
	Notes             *string `json:"notes"`
}

type UpdateInvoicePaymentStatusRequest struct {
	Status string  `json:"status" validate:"required,oneof=pending confirmed failed reversed rejected"`
	Notes  *string `json:"notes"`
}
