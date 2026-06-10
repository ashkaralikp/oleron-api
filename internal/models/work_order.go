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
