package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// buildVisionResponse creates a fake Google Vision API response with the given text.
func buildVisionResponse(fullText string) googleVisionResponse {
	return googleVisionResponse{
		Responses: []googleVisionAnnotateResponse{
			{
				FullTextAnnotation: &googleVisionFullText{
					Text: fullText,
				},
			},
		},
	}
}

func TestGoogleVisionClient_ExtractInvoice_Success(t *testing.T) {
	invoiceText := `CÔNG TY TNHH ABC
Địa chỉ: 123 Lê Lợi, TP.HCM
HÓA ĐƠN GTGT
Ngày: 28/03/2026
Số hóa đơn: 0001234
Người mua: Công ty XYZ
Mặt hàng: Phần mềm kế toán   10   909.091   9.090.910
Cộng tiền hàng: 9.090.910
Thuế GTGT 10%: 909.091
Tổng cộng tiền thanh toán: 10.000.001
`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify API key is in query params
		if r.URL.Query().Get("key") == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(buildVisionResponse(invoiceText))
	}))
	defer srv.Close()

	tmpFile, err := os.CreateTemp("", "invoice-*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString("fake-image-data")
	tmpFile.Close()

	client := newGoogleVisionClientWithBaseURL("test-api-key", srv.URL, srv.Client())
	result, err := client.ExtractFromFile(context.Background(), tmpFile.Name(), "invoice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.DocumentType != "invoice" {
		t.Errorf("expected DocumentType 'invoice', got %q", result.DocumentType)
	}
	if result.Vendor == "" {
		t.Error("expected non-empty vendor")
	}
	if result.TotalAmount <= 0 {
		t.Errorf("expected positive TotalAmount, got %f", result.TotalAmount)
	}
	if result.TaxAmount <= 0 {
		t.Errorf("expected positive TaxAmount, got %f", result.TaxAmount)
	}
	if result.Date == "" {
		t.Error("expected non-empty date")
	}
	if result.Currency != "VND" {
		t.Errorf("expected VND, got %s", result.Currency)
	}
}

func TestGoogleVisionClient_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	tmpFile, err := os.CreateTemp("", "*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	client := newGoogleVisionClientWithBaseURL("key", srv.URL, srv.Client())
	_, err = client.ExtractFromFile(context.Background(), tmpFile.Name(), "invoice")
	if err == nil {
		t.Error("expected error for HTTP 500 response")
	}
}

func TestGoogleVisionClient_MalformedJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{invalid`))
	}))
	defer srv.Close()

	tmpFile, err := os.CreateTemp("", "*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	client := newGoogleVisionClientWithBaseURL("key", srv.URL, srv.Client())
	_, err = client.ExtractFromFile(context.Background(), tmpFile.Name(), "invoice")
	if err == nil {
		t.Error("expected error for malformed JSON")
	}
}

func TestGoogleVisionClient_EmptyResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(googleVisionResponse{
			Responses: []googleVisionAnnotateResponse{},
		})
	}))
	defer srv.Close()

	tmpFile, err := os.CreateTemp("", "*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	client := newGoogleVisionClientWithBaseURL("key", srv.URL, srv.Client())
	_, err = client.ExtractFromFile(context.Background(), tmpFile.Name(), "invoice")
	if err == nil {
		t.Error("expected error for empty responses array")
	}
}

func TestGoogleVisionClient_ParseVNAmount(t *testing.T) {
	cases := []struct {
		input    string
		expected float64
	}{
		{"10.000.001", 10000001},
		{"1,100,000", 1100000},
		{"909.091", 909091},
		{"5000000", 5000000},
		{"", 0},
	}
	for _, c := range cases {
		got := parseVNAmount(c.input)
		if got != c.expected {
			t.Errorf("parseVNAmount(%q) = %f, want %f", c.input, got, c.expected)
		}
	}
}

func TestGoogleVisionClient_ExtractVendorFromText(t *testing.T) {
	text := "CÔNG TY TNHH ABC\nĐịa chỉ: 123 Lê Lợi\nHÓA ĐƠN"
	vendor := extractVendorFromText(text)
	if vendor == "" {
		t.Error("expected non-empty vendor from text")
	}
}
