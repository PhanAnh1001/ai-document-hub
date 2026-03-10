package service

import (
	"testing"
)

func TestVASRuleEngine_Apply_TableDriven(t *testing.T) {
	engine := NewVASRuleEngine()

	cases := []struct {
		name       string
		ocrResult  *OCRResult
		docType    string
		wantCount  int   // number of entries expected
		wantDebit  string // first entry debit
		wantCredit string // first entry credit
	}{
		{
			name: "invoice mua hàng hóa (no VAT)",
			ocrResult: &OCRResult{
				Vendor:      "Công ty ABC",
				TotalAmount: 1000000,
				TaxAmount:   0,
				Date:        "2026-03-28",
			},
			docType:    "invoice",
			wantCount:  1,
			wantDebit:  "156",
			wantCredit: "331",
		},
		{
			name: "invoice mua hàng hóa + VAT",
			ocrResult: &OCRResult{
				Vendor:      "Công ty ABC",
				TotalAmount: 1100000,
				TaxAmount:   100000,
				Date:        "2026-03-28",
			},
			docType:   "invoice",
			wantCount: 2, // 156/331 + 133/331
			wantDebit: "156",
			wantCredit: "331",
		},
		{
			name: "invoice bán hàng (sales keyword in vendor)",
			ocrResult: &OCRResult{
				Vendor:      "khách hàng Nguyễn Văn A",
				TotalAmount: 1100000,
				TaxAmount:   100000,
				Date:        "2026-03-28",
			},
			docType:   "invoice",
			wantCount: 2, // 131/511 + 131/3331
			wantDebit: "131",
			wantCredit: "511",
		},
		{
			name: "receipt chi tiền mặt",
			ocrResult: &OCRResult{
				Vendor:      "Cửa hàng XYZ",
				TotalAmount: 500000,
				TaxAmount:   0,
				Date:        "2026-03-28",
			},
			docType:    "receipt",
			wantCount:  1,
			wantDebit:  "331",
			wantCredit: "111",
		},
		{
			name: "receipt thu tiền (thu keyword)",
			ocrResult: &OCRResult{
				Vendor:      "thu tiền từ Công ty DEF",
				TotalAmount: 2000000,
				TaxAmount:   0,
				Date:        "2026-03-28",
			},
			docType:    "receipt",
			wantCount:  1,
			wantDebit:  "111",
			wantCredit: "131",
		},
		{
			name: "bank_statement chi",
			ocrResult: &OCRResult{
				Vendor:      "Ngân hàng thanh toán cho Công ty GHI",
				TotalAmount: 3000000,
				TaxAmount:   0,
				Date:        "2026-03-28",
			},
			docType:    "bank_statement",
			wantCount:  1,
			wantDebit:  "331",
			wantCredit: "112",
		},
		{
			name: "bank_statement thu",
			ocrResult: &OCRResult{
				Vendor:      "nhận tiền từ khách hàng",
				TotalAmount: 5000000,
				TaxAmount:   0,
				Date:        "2026-03-28",
			},
			docType:    "bank_statement",
			wantCount:  1,
			wantDebit:  "112",
			wantCredit: "131",
		},
		{
			name: "other lương nhân viên",
			ocrResult: &OCRResult{
				Vendor:      "bảng lương tháng 3",
				TotalAmount: 15000000,
				TaxAmount:   0,
				Date:        "2026-03-28",
			},
			docType:    "other",
			wantCount:  1,
			wantDebit:  "622",
			wantCredit: "334",
		},
		{
			name: "other điện nước",
			ocrResult: &OCRResult{
				Vendor:      "hóa đơn tiền điện tháng 3",
				TotalAmount: 500000,
				TaxAmount:   0,
				Date:        "2026-03-28",
			},
			docType:    "other",
			wantCount:  1,
			wantDebit:  "6278",
			wantCredit: "331",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			entries := engine.Apply(tc.ocrResult, tc.docType, "doc-1", "org-1")

			if len(entries) != tc.wantCount {
				t.Errorf("expected %d entries, got %d", tc.wantCount, len(entries))
				for i, e := range entries {
					t.Logf("  entry[%d]: %s/%s %.2f", i, e.DebitAccount, e.CreditAccount, e.Amount)
				}
				return
			}

			if entries[0].DebitAccount != tc.wantDebit {
				t.Errorf("expected debit %s, got %s", tc.wantDebit, entries[0].DebitAccount)
			}
			if entries[0].CreditAccount != tc.wantCredit {
				t.Errorf("expected credit %s, got %s", tc.wantCredit, entries[0].CreditAccount)
			}

			// All entries must have required fields
			for _, e := range entries {
				if e.ID == "" {
					t.Error("entry missing ID")
				}
				if e.OrganizationID != "org-1" {
					t.Errorf("expected org-1, got %s", e.OrganizationID)
				}
				if e.Amount <= 0 {
					t.Errorf("entry amount must be positive, got %f", e.Amount)
				}
				if e.Status != "pending" {
					t.Errorf("new entries must be pending, got %s", e.Status)
				}
			}
		})
	}
}

func TestVASRuleEngine_Apply_SetsConfidence(t *testing.T) {
	engine := NewVASRuleEngine()
	result := &OCRResult{
		Vendor: "Công ty ABC", TotalAmount: 1000000, Date: "2026-03-28",
	}
	entries := engine.Apply(result, "invoice", "doc-1", "org-1")
	for _, e := range entries {
		if e.AIConfidence == nil {
			t.Error("expected AI confidence to be set")
		}
		if *e.AIConfidence < 0 || *e.AIConfidence > 1 {
			t.Errorf("confidence out of range [0,1]: %f", *e.AIConfidence)
		}
	}
}

func TestVASRuleEngine_Apply_InvoiceWithVAT_SecondEntryIs133(t *testing.T) {
	engine := NewVASRuleEngine()
	result := &OCRResult{
		Vendor: "ABC", TotalAmount: 1100000, TaxAmount: 100000, Date: "2026-03-28",
	}
	entries := engine.Apply(result, "invoice", "doc-1", "org-1")
	if len(entries) < 2 {
		t.Fatalf("expected 2 entries for invoice with VAT, got %d", len(entries))
	}
	vatEntry := entries[1]
	if vatEntry.DebitAccount != "133" {
		t.Errorf("second entry should be 133 (VAT input), got %s", vatEntry.DebitAccount)
	}
	if vatEntry.Amount != 100000 {
		t.Errorf("VAT entry amount should be 100000, got %f", vatEntry.Amount)
	}
}
