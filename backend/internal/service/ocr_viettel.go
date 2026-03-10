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
	"time"
)

// ViettelAIClient calls the Viettel AI OCR API.
// Docs: https://viettelgroup.ai/document/ocr
// Auth: header "token: <api_key>"
// Free tier: contact viettelai@viettel.com.vn
type ViettelAIClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// viettelAIResponse is the response shape from Viettel AI OCR API.
type viettelAIResponse struct {
	Result string `json:"result"`
	Status int    `json:"status"`
	Error  string `json:"error,omitempty"`
}

// NewViettelAIClient creates a production client targeting viettelgroup.ai.
func NewViettelAIClient(apiKey string) *ViettelAIClient {
	return &ViettelAIClient{
		apiKey:  apiKey,
		baseURL: "https://viettelgroup.ai/cv/api/v1/ocr",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// newViettelAIClientWithBaseURL creates a client with a custom URL (for tests).
func newViettelAIClientWithBaseURL(apiKey, baseURL string, httpClient *http.Client) *ViettelAIClient {
	return &ViettelAIClient{
		apiKey:     apiKey,
		baseURL:    baseURL,
		httpClient: httpClient,
	}
}

// ExtractFromFile uploads the file to Viettel AI and parses the OCR result.
func (c *ViettelAIClient) ExtractFromFile(ctx context.Context, filePath, docType string) (*OCRResult, error) {
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
	req.Header.Set("token", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("viettel ai request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("viettel ai returned HTTP %d", resp.StatusCode)
	}

	var viettelResp viettelAIResponse
	if err := json.NewDecoder(resp.Body).Decode(&viettelResp); err != nil {
		return nil, fmt.Errorf("decode viettel ai response: %w", err)
	}

	if viettelResp.Status != 200 {
		return nil, fmt.Errorf("viettel ai error status %d: %s", viettelResp.Status, viettelResp.Error)
	}
	if viettelResp.Result == "" {
		return nil, fmt.Errorf("viettel ai returned empty result")
	}

	result := parseVNInvoiceText(viettelResp.Result, docType)
	result.RawData = map[string]interface{}{"full_text": viettelResp.Result}
	return result, nil
}
