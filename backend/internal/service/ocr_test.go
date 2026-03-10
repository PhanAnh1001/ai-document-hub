package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestFPTAIClient_ExtractInvoice_Success(t *testing.T) {
	// Setup mock FPT.AI server
	fptResp := fptAIResponse{
		ErrorCode: 0,
		Data: []fptAIInvoice{
			{
				Seller:      "Công ty ABC",
				TotalAmount: "1,100,000",
				Vat:         "100,000",
				InvoiceDate: "28/03/2026",
				InvoiceNo:   "0001234",
			},
		},
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("api-key") == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(fptResp)
	}))
	defer srv.Close()

	// Create temp file to upload
	tmpFile, err := os.CreateTemp("", "invoice-*.pdf")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString("%PDF-test-content")
	tmpFile.Close()

	client := newFPTAIClientWithBaseURL("test-api-key", srv.URL, srv.Client())
	result, err := client.ExtractFromFile(context.Background(), tmpFile.Name(), "invoice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Vendor != "Công ty ABC" {
		t.Errorf("expected vendor 'Công ty ABC', got %q", result.Vendor)
	}
	if result.TotalAmount != 1100000 {
		t.Errorf("expected total 1100000, got %f", result.TotalAmount)
	}
	if result.TaxAmount != 100000 {
		t.Errorf("expected tax 100000, got %f", result.TaxAmount)
	}
	if result.Date != "2026-03-28" {
		t.Errorf("expected date '2026-03-28', got %q", result.Date)
	}
	if result.Currency != "VND" {
		t.Errorf("expected VND, got %s", result.Currency)
	}
}

func TestFPTAIClient_APIError_Returns500(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	tmpFile, err := os.CreateTemp("", "*.pdf")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	client := newFPTAIClientWithBaseURL("key", srv.URL, srv.Client())
	_, err = client.ExtractFromFile(context.Background(), tmpFile.Name(), "invoice")
	if err == nil {
		t.Error("expected error for 500 response")
	}
}

func TestFPTAIClient_MalformedJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{invalid json`))
	}))
	defer srv.Close()

	tmpFile, err := os.CreateTemp("", "*.pdf")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	client := newFPTAIClientWithBaseURL("key", srv.URL, srv.Client())
	_, err = client.ExtractFromFile(context.Background(), tmpFile.Name(), "invoice")
	if err == nil {
		t.Error("expected error for malformed JSON")
	}
}

func TestFPTAIClient_EmptyData(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(fptAIResponse{ErrorCode: 0, Data: []fptAIInvoice{}})
	}))
	defer srv.Close()

	tmpFile, err := os.CreateTemp("", "*.pdf")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	client := newFPTAIClientWithBaseURL("key", srv.URL, srv.Client())
	_, err = client.ExtractFromFile(context.Background(), tmpFile.Name(), "invoice")
	if err == nil {
		t.Error("expected error for empty data")
	}
}

func TestMockOCRService_ReturnsDefaultResult(t *testing.T) {
	mock := &MockOCRService{Result: DefaultMockOCRResult()}
	result, err := mock.ExtractFromFile(context.Background(), "any.pdf", "invoice")
	if err != nil {
		t.Fatal(err)
	}
	if result.Vendor == "" {
		t.Error("expected non-empty vendor")
	}
	if result.TotalAmount <= 0 {
		t.Error("expected positive total amount")
	}
}
