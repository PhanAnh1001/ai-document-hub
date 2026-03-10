package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestViettelAIClient_ExtractInvoice_Success(t *testing.T) {
	invoiceText := `CÔNG TY TNHH VIETTEL
Địa chỉ: 1 Giang Văn Minh, Hà Nội
HÓA ĐƠN GTGT
Ngày: 28/03/2026
Số hóa đơn: HĐ-0099
Người mua: Công ty ABC
Mặt hàng: Dịch vụ viễn thông   1   4.545.455   4.545.455
Cộng tiền hàng: 4.545.455
Thuế GTGT 10%: 454.545
Tổng cộng tiền thanh toán: 5.000.000
`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify token header is present
		if r.Header.Get("token") == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		// Verify multipart form with "file" field
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if r.MultipartForm.File["file"] == nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(viettelAIResponse{
			Result: invoiceText,
			Status: 200,
		})
	}))
	defer srv.Close()

	tmpFile, err := os.CreateTemp("", "invoice-*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.WriteString("fake-image-data")
	tmpFile.Close()

	client := newViettelAIClientWithBaseURL("test-token", srv.URL, srv.Client())
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
	if result.Currency != "VND" {
		t.Errorf("expected VND, got %s", result.Currency)
	}
}

func TestViettelAIClient_MissingToken_Returns401(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer srv.Close()

	tmpFile, err := os.CreateTemp("", "*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	client := newViettelAIClientWithBaseURL("", srv.URL, srv.Client())
	_, err = client.ExtractFromFile(context.Background(), tmpFile.Name(), "invoice")
	if err == nil {
		t.Error("expected error for 401 unauthorized")
	}
}

func TestViettelAIClient_HTTPError(t *testing.T) {
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

	client := newViettelAIClientWithBaseURL("token", srv.URL, srv.Client())
	_, err = client.ExtractFromFile(context.Background(), tmpFile.Name(), "invoice")
	if err == nil {
		t.Error("expected error for HTTP 500")
	}
}

func TestViettelAIClient_MalformedJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{bad json`))
	}))
	defer srv.Close()

	tmpFile, err := os.CreateTemp("", "*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	client := newViettelAIClientWithBaseURL("token", srv.URL, srv.Client())
	_, err = client.ExtractFromFile(context.Background(), tmpFile.Name(), "invoice")
	if err == nil {
		t.Error("expected error for malformed JSON")
	}
}

func TestViettelAIClient_EmptyResult(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(viettelAIResponse{Result: "", Status: 200})
	}))
	defer srv.Close()

	tmpFile, err := os.CreateTemp("", "*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	client := newViettelAIClientWithBaseURL("token", srv.URL, srv.Client())
	_, err = client.ExtractFromFile(context.Background(), tmpFile.Name(), "invoice")
	if err == nil {
		t.Error("expected error for empty result text")
	}
}

func TestViettelAIClient_ErrorStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(viettelAIResponse{
			Result: "",
			Status: 400,
			Error:  "invalid file format",
		})
	}))
	defer srv.Close()

	tmpFile, err := os.CreateTemp("", "*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	client := newViettelAIClientWithBaseURL("token", srv.URL, srv.Client())
	_, err = client.ExtractFromFile(context.Background(), tmpFile.Name(), "invoice")
	if err == nil {
		t.Error("expected error for non-200 status in response body")
	}
}
