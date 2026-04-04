package model

import "time"

type AccountingStandard string

const (
	TT133 AccountingStandard = "TT133"
	TT200 AccountingStandard = "TT200"
)

type OCRProvider string

const (
	OCRProviderFPT  OCRProvider = "fpt"
	OCRProviderMock OCRProvider = "mock"
)

type UserRole string

const (
	RoleOwner  UserRole = "owner"
	RoleAdmin  UserRole = "admin"
	RoleMember UserRole = "member"
)

type DocumentType string

const (
	DocInvoice       DocumentType = "invoice"
	DocReceipt       DocumentType = "receipt"
	DocBankStatement DocumentType = "bank_statement"
	DocOther         DocumentType = "other"
)

type DocumentStatus string

const (
	StatusUploaded   DocumentStatus = "uploaded"
	StatusProcessing DocumentStatus = "processing"
	StatusExtracted  DocumentStatus = "extracted"
	StatusBooked     DocumentStatus = "booked"
	StatusError      DocumentStatus = "error"
)

type EntryStatus string

const (
	EntryDraft     EntryStatus = "draft"
	EntryPending   EntryStatus = "pending"
	EntryApproved  EntryStatus = "approved"
	EntryRejected  EntryStatus = "rejected"
	EntrySynced    EntryStatus = "synced"
	EntryErrorSync EntryStatus = "error_sync" // MISA callback returned failure for this entry
)

type Organization struct {
	ID                 string             `json:"id" db:"id"`
	Name               string             `json:"name" db:"name"`
	TaxCode            *string            `json:"tax_code,omitempty" db:"tax_code"`
	AccountingStandard AccountingStandard `json:"accounting_standard" db:"accounting_standard"`
	OCRProvider        OCRProvider        `json:"ocr_provider" db:"ocr_provider"`
	MISAAPIURL         *string            `json:"misa_api_url,omitempty" db:"misa_api_url"`
	MISAAPIKey         *string            `json:"misa_api_key,omitempty" db:"misa_api_key"`
	CreatedAt          time.Time          `json:"created_at" db:"created_at"`
	UpdatedAt          time.Time          `json:"updated_at" db:"updated_at"`
}

type User struct {
	ID             string    `json:"id" db:"id"`
	OrganizationID string   `json:"organization_id" db:"organization_id"`
	Email          string    `json:"email" db:"email"`
	PasswordHash   string    `json:"-" db:"password_hash"`
	FullName       *string   `json:"full_name,omitempty" db:"full_name"`
	Role           UserRole  `json:"role" db:"role"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
}

type Document struct {
	ID             string         `json:"id" db:"id"`
	OrganizationID string         `json:"organization_id" db:"organization_id"`
	FileURL        string         `json:"file_url" db:"file_url"`
	FileName       string         `json:"file_name" db:"file_name"`
	DocumentType   DocumentType   `json:"document_type" db:"document_type"`
	Status         DocumentStatus `json:"status" db:"status"`
	OCRData        []byte         `json:"ocr_data,omitempty" db:"ocr_data"`
	CreatedBy      string         `json:"created_by" db:"created_by"`
	CreatedAt      time.Time      `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at" db:"updated_at"`
}

type AccountingEntry struct {
	ID             string      `json:"id" db:"id"`
	OrganizationID string      `json:"organization_id" db:"organization_id"`
	DocumentID     *string     `json:"document_id,omitempty" db:"document_id"`
	DocumentName   *string     `json:"document_name,omitempty" db:"document_name"`
	EntryDate      time.Time   `json:"entry_date" db:"entry_date"`
	Description    string      `json:"description" db:"description"`
	DebitAccount   string      `json:"debit_account" db:"debit_account"`
	CreditAccount  string      `json:"credit_account" db:"credit_account"`
	Amount         float64     `json:"amount" db:"amount"`
	Status         EntryStatus `json:"status" db:"status"`
	RejectReason   *string     `json:"reject_reason,omitempty" db:"reject_reason"`
	AIConfidence   *float64    `json:"ai_confidence,omitempty" db:"ai_confidence"`
	CreatedAt      time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at" db:"updated_at"`
}
