package service

import "context"

// MockOCRService returns a fixed OCRResult for testing and CI environments.
type MockOCRService struct {
	Result *OCRResult
	Err    error
}

func (m *MockOCRService) ExtractFromFile(_ context.Context, _, _ string) (*OCRResult, error) {
	return m.Result, m.Err
}

// DefaultMockOCRResult returns a realistic invoice OCR result for testing.
func DefaultMockOCRResult() *OCRResult {
	return &OCRResult{
		DocumentType: "invoice",
		Vendor:       "Công ty TNHH ABC",
		TotalAmount:  909090.91,
		TaxAmount:    90909.09,
		Date:         "2026-03-28",
		Currency:     "VND",
		LineItems: []OCRLineItem{
			{Description: "Hàng hóa A", Quantity: 10, UnitPrice: 90909.09, Amount: 909090.91},
		},
		RawData: map[string]interface{}{"source": "mock"},
	}
}
