package models

import "time"

type RegionalManagerWorkOrderAsset struct {
	ID                string    `json:"id"`
	RegionalManagerID string    `json:"regional_manager_id"`
	SignerName        string    `json:"signer_name"`
	SignerTitle       *string   `json:"signer_title,omitempty"`
	SignatureURL      string    `json:"signature_url"`
	SealURL           string    `json:"seal_url"`
	UploadedBy        *string   `json:"uploaded_by,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

type WorkOrder struct {
	ID             string          `json:"id"`
	BranchID       string          `json:"branch_id"`
	CreatedBy      string          `json:"created_by"`
	ManagerAssetID *string         `json:"manager_asset_id,omitempty"`
	WorkOrderNo    string          `json:"work_order_no"`
	WorkOrderDate  time.Time       `json:"work_order_date"`
	CompanyName    string          `json:"company_name"`
	CompanyAddress *string         `json:"company_address,omitempty"`
	CompanyPhone   *string         `json:"company_phone,omitempty"`
	CompanyFax     *string         `json:"company_fax,omitempty"`
	CompanyEmail   *string         `json:"company_email,omitempty"`
	CompanyWebsite *string         `json:"company_website,omitempty"`
	CompanyLogoURL *string         `json:"company_logo_url,omitempty"`
	BillToName     string          `json:"bill_to_name"`
	BillToAddress  *string         `json:"bill_to_address,omitempty"`
	BillToEmail    *string         `json:"bill_to_email,omitempty"`
	JobDetails     string          `json:"job_details"`
	SignerName     *string         `json:"signer_name,omitempty"`
	SignerTitle    *string         `json:"signer_title,omitempty"`
	SignatureURL   *string         `json:"signature_url,omitempty"`
	SealURL        *string         `json:"seal_url,omitempty"`
	Currency       string          `json:"currency"`
	SubTotalAmount float64         `json:"sub_total_amount"`
	TotalAmount    float64         `json:"total_amount"`
	Status         string          `json:"status"`
	PDFURL         *string         `json:"pdf_url,omitempty"`
	IssuedAt       *time.Time      `json:"issued_at,omitempty"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
	Items          []WorkOrderItem `json:"items,omitempty"`
}

type WorkOrderItem struct {
	ID          string    `json:"id"`
	WorkOrderID string    `json:"work_order_id"`
	LineNo      int       `json:"line_no"`
	Description string    `json:"description"`
	Amount      float64   `json:"amount"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type WorkOrderInvoice struct {
	ID               string                 `json:"id"`
	WorkOrderID      string                 `json:"work_order_id"`
	BranchID         string                 `json:"branch_id"`
	CreatedBy        string                 `json:"created_by"`
	InvoiceNo        string                 `json:"invoice_no"`
	InvoiceDate      time.Time              `json:"invoice_date"`
	DueDate          *time.Time             `json:"due_date,omitempty"`
	InvoiceTitle     string                 `json:"invoice_title"`
	SupplyNote       *string                `json:"supply_note,omitempty"`
	SellerName       string                 `json:"seller_name"`
	SellerAddress    *string                `json:"seller_address,omitempty"`
	SellerEmail      *string                `json:"seller_email,omitempty"`
	SellerPhone      *string                `json:"seller_phone,omitempty"`
	SellerGSTIN      *string                `json:"seller_gstin,omitempty"`
	SellerLogoURL    *string                `json:"seller_logo_url,omitempty"`
	BillToName       string                 `json:"bill_to_name"`
	BillToAddress    *string                `json:"bill_to_address,omitempty"`
	BillToEmail      *string                `json:"bill_to_email,omitempty"`
	BillToPhone      *string                `json:"bill_to_phone,omitempty"`
	BillToWebsite    *string                `json:"bill_to_website,omitempty"`
	Currency         string                 `json:"currency"`
	GrossAmount      float64                `json:"gross_amount"`
	TaxAmount        float64                `json:"tax_amount"`
	AdditionalAmount float64                `json:"additional_amount"`
	TotalAmount      float64                `json:"total_amount"`
	PaidAmount       float64                `json:"paid_amount"`
	BalanceAmount    float64                `json:"balance_amount"`
	Status           string                 `json:"status"`
	PaymentStatus    string                 `json:"payment_status"`
	LUTOrderNumber   *string                `json:"lut_order_number,omitempty"`
	ARNNumber        *string                `json:"arn_number,omitempty"`
	Notes            *string                `json:"notes,omitempty"`
	SignerName       *string                `json:"signer_name,omitempty"`
	SignerTitle      *string                `json:"signer_title,omitempty"`
	SignatureURL     *string                `json:"signature_url,omitempty"`
	SealURL          *string                `json:"seal_url,omitempty"`
	PDFURL           *string                `json:"pdf_url,omitempty"`
	IssuedAt         *time.Time             `json:"issued_at,omitempty"`
	CancelledAt      *time.Time             `json:"cancelled_at,omitempty"`
	CreatedAt        time.Time              `json:"created_at"`
	UpdatedAt        time.Time              `json:"updated_at"`
	Items            []WorkOrderInvoiceItem `json:"items,omitempty"`
}

type WorkOrderInvoiceItem struct {
	ID              string    `json:"id"`
	InvoiceID       string    `json:"invoice_id"`
	WorkOrderItemID *string   `json:"work_order_item_id,omitempty"`
	LineNo          int       `json:"line_no"`
	Description     string    `json:"description"`
	Quantity        float64   `json:"quantity"`
	SACCode         *string   `json:"sac_code,omitempty"`
	UnitAmount      float64   `json:"unit_amount"`
	TaxAmount       float64   `json:"tax_amount"`
	TotalAmount     float64   `json:"total_amount"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type WorkOrderInvoicePayment struct {
	ID                string                             `json:"id"`
	InvoiceID         string                             `json:"invoice_id"`
	RecordedBy        string                             `json:"recorded_by"`
	PaymentDate       time.Time                          `json:"payment_date"`
	Amount            float64                            `json:"amount"`
	Currency          string                             `json:"currency"`
	Method            string                             `json:"method"`
	OtherMethod       *string                            `json:"other_method,omitempty"`
	ReferenceNo       *string                            `json:"reference_no,omitempty"`
	PayerName         *string                            `json:"payer_name,omitempty"`
	PayerAccountLast4 *string                            `json:"payer_account_last4,omitempty"`
	BankName          *string                            `json:"bank_name,omitempty"`
	Status            string                             `json:"status"`
	Notes             *string                            `json:"notes,omitempty"`
	VerifiedBy        *string                            `json:"verified_by,omitempty"`
	VerifiedAt        *time.Time                         `json:"verified_at,omitempty"`
	CreatedAt         time.Time                          `json:"created_at"`
	UpdatedAt         time.Time                          `json:"updated_at"`
	Statements        []WorkOrderInvoicePaymentStatement `json:"statements,omitempty"`
}

type WorkOrderInvoicePaymentStatement struct {
	ID               string    `json:"id"`
	PaymentID        string    `json:"payment_id"`
	UploadedBy       string    `json:"uploaded_by"`
	StatementURL     string    `json:"statement_url"`
	OriginalFilename *string   `json:"original_filename,omitempty"`
	FileMIMEType     *string   `json:"file_mime_type,omitempty"`
	FileSizeBytes    *int64    `json:"file_size_bytes,omitempty"`
	Notes            *string   `json:"notes,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
}
