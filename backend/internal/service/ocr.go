package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// OCRLineItem represents a single line item extracted from a document.
type OCRLineItem struct {
	Description string  `json:"description"`
	Quantity    float64 `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	Amount      float64 `json:"amount"`
}

// OCRResult is the normalized output from any OCR provider.
type OCRResult struct {
	DocumentType string                 `json:"document_type"`
	Vendor       string                 `json:"vendor"`
	TotalAmount  float64                `json:"total_amount"`
	TaxAmount    float64                `json:"tax_amount"`
	Date         string                 `json:"date"` // "2006-01-02" format
	Currency     string                 `json:"currency"`
	LineItems    []OCRLineItem          `json:"line_items"`
	RawData      map[string]interface{} `json:"raw_data"`
}

// OCRService is the interface all OCR providers must implement.
type OCRService interface {
	ExtractFromFile(ctx context.Context, filePath, docType string) (*OCRResult, error)
}

// FPTAIClient calls the FPT.AI document AI API.
type FPTAIClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// fptAIResponse is the raw response shape from FPT.AI invoice API.
type fptAIResponse struct {
	ErrorCode int              `json:"errorCode"`
	Data      []fptAIInvoice  `json:"data"`
}

type fptAIInvoice struct {
	Seller      string `json:"seller"`
	TotalAmount string `json:"total_amount"`
	Vat         string `json:"vat"`
	InvoiceDate string `json:"invoice_date"`
	InvoiceNo   string `json:"invoice_no"`
	Buyer       string `json:"buyer"`
}

func NewFPTAIClient(apiKey string) *FPTAIClient {
	return &FPTAIClient{
		apiKey:  apiKey,
		baseURL: "https://api.fpt.ai/vision/inovice/recognize",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// newFPTAIClientWithBaseURL creates a client with a custom base URL (for tests).
func newFPTAIClientWithBaseURL(apiKey, baseURL string, httpClient *http.Client) *FPTAIClient {
	return &FPTAIClient{
		apiKey:     apiKey,
		baseURL:    baseURL,
		httpClient: httpClient,
	}
}

// ExtractFromFile reads the file and calls FPT.AI API to extract invoice data.
func (c *FPTAIClient) ExtractFromFile(ctx context.Context, filePath, docType string) (*OCRResult, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	fw, err := mw.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return nil, fmt.Errorf("create form file: %w", err)
	}
	if _, err := io.Copy(fw, f); err != nil {
		return nil, fmt.Errorf("copy file: %w", err)
	}
	mw.Close()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL, &body)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", mw.FormDataContentType())
	req.Header.Set("api-key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fpt.ai request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fpt.ai returned status %d", resp.StatusCode)
	}

	var fptResp fptAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&fptResp); err != nil {
		return nil, fmt.Errorf("decode fpt.ai response: %w", err)
	}
	if fptResp.ErrorCode != 0 || len(fptResp.Data) == 0 {
		return nil, fmt.Errorf("fpt.ai error code %d or empty data", fptResp.ErrorCode)
	}

	inv := fptResp.Data[0]
	result := &OCRResult{
		DocumentType: docType,
		Vendor:       inv.Seller,
		TotalAmount:  parseAmount(inv.TotalAmount),
		TaxAmount:    parseAmount(inv.Vat),
		Date:         normDate(inv.InvoiceDate),
		Currency:     "VND",
		RawData:      map[string]interface{}{"fpt_response": fptResp},
	}
	return result, nil
}

// parseAmount converts a string like "1,000,000" or "1000000" to float64.
func parseAmount(s string) float64 {
	s = strings.ReplaceAll(s, ",", "")
	s = strings.ReplaceAll(s, ".", "")
	s = strings.TrimSpace(s)
	var v float64
	fmt.Sscanf(s, "%f", &v)
	return v
}

// normDate attempts to parse common VN date formats and return YYYY-MM-DD.
func normDate(s string) string {
	s = strings.TrimSpace(s)
	formats := []string{"02/01/2006", "2006-01-02", "02-01-2006", "01/02/2006"}
	for _, f := range formats {
		if t, err := time.Parse(f, s); err == nil {
			return t.Format("2006-01-02")
		}
	}
	return time.Now().Format("2006-01-02")
}
