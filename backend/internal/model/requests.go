package model

// Document
type DocumentListResponse struct {
	Documents []*Document `json:"documents"`
	Total     int         `json:"total"`
}

// DayCount holds a count of documents uploaded on a single calendar day.
type DayCount struct {
	Date  string `json:"date"`  // YYYY-MM-DD
	Count int    `json:"count"`
}

// Accounting entries
type EntryListResponse struct {
	Entries []*AccountingEntry `json:"entries"`
	Total   int                `json:"total"`
}

type EntryActionResponse struct {
	Entry *AccountingEntry `json:"entry"`
}

// UpdateEntryFields holds the editable fields for PATCH /api/entries/{id}.
// Only pending entries may be updated.
type UpdateEntryFields struct {
	Description   string  `json:"description"`
	DebitAccount  string  `json:"debit_account"`
	CreditAccount string  `json:"credit_account"`
	Amount        float64 `json:"amount"`
}

// MISA Callback — structures matching MISA AMIS callback API spec.

// MISACallbackInput is the POST body MISA sends to our callback endpoint.
type MISACallbackInput struct {
	Success        bool   `json:"success"`
	AppID          string `json:"app_id"`
	ErrorCode      string `json:"error_code"`
	ErrorMessage   string `json:"error_message"`
	Signature      string `json:"signature"`
	DataType       int    `json:"data_type"`
	OrgCompanyCode string `json:"org_company_code"`
	Data           string `json:"data"` // JSON-encoded list of MISACallbackDetail
}

// MISACallbackOutput is the JSON response we return to MISA.
type MISACallbackOutput struct {
	Success      bool   `json:"Success"`
	ErrorCode    string `json:"ErrorCode,omitempty"`
	ErrorMessage string `json:"ErrorMessage"`
}

// MISACallbackDetail is one item inside the Data field of MISACallbackInput.
type MISACallbackDetail struct {
	OrgRefID             string  `json:"org_refid"`              // maps to our entry ID
	Success              bool    `json:"success"`
	ErrorCode            *string `json:"error_code"`
	ErrorMessage         string  `json:"error_message"`
	SessionID            *string `json:"session_id"`
	ErrorCallBackMessage *string `json:"error_call_back_message"`
	VoucherType          *int    `json:"voucher_type"`
}

// Auth
type RegisterRequest struct {
	Email       string `json:"email"`
	Password    string `json:"password"`
	FullName    string `json:"full_name"`
	CompanyName string `json:"company_name"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Token string `json:"token"`
	User  User   `json:"user"`
}
