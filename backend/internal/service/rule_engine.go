package service

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
	"time"

	"github.com/PhanAnh1001/ai-accounting/backend/internal/model"
)

// VASRuleEngine maps OCR results to Vietnamese Accounting Standard (VAS) entries.
type VASRuleEngine struct{}

func NewVASRuleEngine() *VASRuleEngine {
	return &VASRuleEngine{}
}

// Apply returns one or more double-entry accounting lines for the given document.
// All returned entries have Status=pending and a rule-based AIConfidence score.
func (e *VASRuleEngine) Apply(result *OCRResult, docType, docID, orgID string) []model.AccountingEntry {
	switch docType {
	case "invoice":
		return e.applyInvoice(result, docID, orgID)
	case "receipt":
		return e.applyReceipt(result, docID, orgID)
	case "bank_statement":
		return e.applyBankStatement(result, docID, orgID)
	default:
		return e.applyOther(result, docID, orgID)
	}
}

// applyInvoice generates entries for invoice documents.
// Default: purchase (156/331). Flip to sales if vendor contains sales keywords.
func (e *VASRuleEngine) applyInvoice(result *OCRResult, docID, orgID string) []model.AccountingEntry {
	if isSalesKeyword(result.Vendor) {
		return e.applySalesInvoice(result, docID, orgID)
	}
	return e.applyPurchaseInvoice(result, docID, orgID)
}

func (e *VASRuleEngine) applyPurchaseInvoice(result *OCRResult, docID, orgID string) []model.AccountingEntry {
	conf := 0.90
	goodsAmount := result.TotalAmount - result.TaxAmount
	if goodsAmount <= 0 {
		goodsAmount = result.TotalAmount
	}

	entries := []model.AccountingEntry{
		makeEntry(docID, orgID, result.Date,
			"Mua hàng hóa nhập kho - "+result.Vendor,
			"156", "331", goodsAmount, &conf),
	}

	// Separate VAT entry if tax amount is present
	if result.TaxAmount > 0 {
		vatConf := 0.92
		entries = append(entries, makeEntry(docID, orgID, result.Date,
			"Thuế GTGT đầu vào - "+result.Vendor,
			"133", "331", result.TaxAmount, &vatConf))
	}
	return entries
}

func (e *VASRuleEngine) applySalesInvoice(result *OCRResult, docID, orgID string) []model.AccountingEntry {
	conf := 0.85
	revenueAmount := result.TotalAmount - result.TaxAmount
	if revenueAmount <= 0 {
		revenueAmount = result.TotalAmount
	}

	entries := []model.AccountingEntry{
		makeEntry(docID, orgID, result.Date,
			"Doanh thu bán hàng - "+result.Vendor,
			"131", "511", revenueAmount, &conf),
	}

	if result.TaxAmount > 0 {
		vatConf := 0.88
		entries = append(entries, makeEntry(docID, orgID, result.Date,
			"Thuế GTGT đầu ra - "+result.Vendor,
			"131", "3331", result.TaxAmount, &vatConf))
	}
	return entries
}

// applyReceipt generates entries for receipt documents.
// Default: cash payment (331/111). "thu" keywords flip to cash receipt (111/131).
func (e *VASRuleEngine) applyReceipt(result *OCRResult, docID, orgID string) []model.AccountingEntry {
	conf := 0.87
	if isReceiptInKeyword(result.Vendor) {
		return []model.AccountingEntry{
			makeEntry(docID, orgID, result.Date,
				"Thu tiền từ khách hàng - "+result.Vendor,
				"111", "131", result.TotalAmount, &conf),
		}
	}
	return []model.AccountingEntry{
		makeEntry(docID, orgID, result.Date,
			"Chi tiền mặt thanh toán - "+result.Vendor,
			"331", "111", result.TotalAmount, &conf),
	}
}

// applyBankStatement generates entries for bank statement documents.
// "nhận"/"thu"/"từ khách" keywords → debit 112. Default → credit 112.
func (e *VASRuleEngine) applyBankStatement(result *OCRResult, docID, orgID string) []model.AccountingEntry {
	conf := 0.83
	if isBankReceiptKeyword(result.Vendor) {
		return []model.AccountingEntry{
			makeEntry(docID, orgID, result.Date,
				"Thu tiền qua ngân hàng - "+result.Vendor,
				"112", "131", result.TotalAmount, &conf),
		}
	}
	return []model.AccountingEntry{
		makeEntry(docID, orgID, result.Date,
			"Trả nợ qua ngân hàng - "+result.Vendor,
			"331", "112", result.TotalAmount, &conf),
	}
}

// applyOther handles documents of type "other" using keyword matching.
func (e *VASRuleEngine) applyOther(result *OCRResult, docID, orgID string) []model.AccountingEntry {
	conf := 0.75
	lower := strings.ToLower(result.Vendor)

	// Payroll / salary
	if containsAny(lower, "lương", "salary", "payroll", "bảng lương") {
		return []model.AccountingEntry{
			makeEntry(docID, orgID, result.Date,
				"Chi phí lương - "+result.Vendor,
				"622", "334", result.TotalAmount, &conf),
		}
	}

	// Utilities: electricity / water
	if containsAny(lower, "điện", "nước", "electricity", "water", "tiền điện", "tiền nước") {
		return []model.AccountingEntry{
			makeEntry(docID, orgID, result.Date,
				"Chi phí điện nước - "+result.Vendor,
				"6278", "331", result.TotalAmount, &conf),
		}
	}

	// Generic expense fallback
	lowConf := 0.60
	return []model.AccountingEntry{
		makeEntry(docID, orgID, result.Date,
			"Chi phí khác - "+result.Vendor,
			"642", "331", result.TotalAmount, &lowConf),
	}
}

// --- keyword helpers ---

func isSalesKeyword(vendor string) bool {
	lower := strings.ToLower(vendor)
	return containsAny(lower, "khách hàng", "bán hàng", "customer", "bán cho")
}

func isReceiptInKeyword(vendor string) bool {
	lower := strings.ToLower(vendor)
	return containsAny(lower, "thu tiền", "nhận tiền", "collect", "received from")
}

func isBankReceiptKeyword(vendor string) bool {
	lower := strings.ToLower(vendor)
	return containsAny(lower, "nhận tiền", "từ khách hàng", "received", "thu tiền")
}

func containsAny(s string, keywords ...string) bool {
	for _, kw := range keywords {
		if strings.Contains(s, kw) {
			return true
		}
	}
	return false
}

// --- entry builder ---

func makeEntry(docID, orgID, date, description, debit, credit string, amount float64, confidence *float64) model.AccountingEntry {
	entryDate := time.Now()
	if t, err := time.Parse("2006-01-02", date); err == nil {
		entryDate = t
	}
	now := time.Now()
	docIDPtr := &docID
	return model.AccountingEntry{
		ID:             genID(),
		OrganizationID: orgID,
		DocumentID:     docIDPtr,
		EntryDate:      entryDate,
		Description:    description,
		DebitAccount:   debit,
		CreditAccount:  credit,
		Amount:         amount,
		Status:         model.EntryPending,
		AIConfidence:   confidence,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

func genID() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
