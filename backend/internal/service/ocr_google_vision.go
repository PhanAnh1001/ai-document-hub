package service

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

// GoogleVisionClient calls the Google Cloud Vision API for document text detection.
// Free tier: 1,000 units/month (document text detection counts as 1 unit per page).
type GoogleVisionClient struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// --- Google Vision API request/response types ---

type googleVisionRequest struct {
	Requests []googleVisionAnnotateRequest `json:"requests"`
}

type googleVisionAnnotateRequest struct {
	Image    googleVisionImage    `json:"image"`
	Features []googleVisionFeature `json:"features"`
}

type googleVisionImage struct {
	Content string `json:"content"` // base64-encoded file bytes
}

type googleVisionFeature struct {
	Type       string `json:"type"`
	MaxResults int    `json:"maxResults"`
}

type googleVisionResponse struct {
	Responses []googleVisionAnnotateResponse `json:"responses"`
}

type googleVisionAnnotateResponse struct {
	FullTextAnnotation *googleVisionFullText `json:"fullTextAnnotation"`
	Error              *googleVisionError    `json:"error"`
}

type googleVisionFullText struct {
	Text string `json:"text"`
}

type googleVisionError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// NewGoogleVisionClient creates a client using the given API key.
func NewGoogleVisionClient(apiKey string) *GoogleVisionClient {
	return &GoogleVisionClient{
		apiKey:  apiKey,
		baseURL: "https://vision.googleapis.com/v1/images:annotate",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// newGoogleVisionClientWithBaseURL creates a client with a custom URL (for tests).
func newGoogleVisionClientWithBaseURL(apiKey, baseURL string, httpClient *http.Client) *GoogleVisionClient {
	return &GoogleVisionClient{
		apiKey:     apiKey,
		baseURL:    baseURL,
		httpClient: httpClient,
	}
}

// ExtractFromFile reads the file, sends to Google Vision API, and parses the result.
func (c *GoogleVisionClient) ExtractFromFile(ctx context.Context, filePath, docType string) (*OCRResult, error) {
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	encoded := base64.StdEncoding.EncodeToString(fileBytes)

	reqBody := googleVisionRequest{
		Requests: []googleVisionAnnotateRequest{
			{
				Image: googleVisionImage{Content: encoded},
				Features: []googleVisionFeature{
					{Type: "DOCUMENT_TEXT_DETECTION", MaxResults: 1},
				},
			},
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	apiURL := fmt.Sprintf("%s?key=%s", c.baseURL, c.apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, apiURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("google vision request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google vision returned status %d", resp.StatusCode)
	}

	var visionResp googleVisionResponse
	if err := json.NewDecoder(resp.Body).Decode(&visionResp); err != nil {
		return nil, fmt.Errorf("decode google vision response: %w", err)
	}

	if len(visionResp.Responses) == 0 {
		return nil, fmt.Errorf("google vision returned empty responses")
	}

	annotate := visionResp.Responses[0]
	if annotate.Error != nil {
		return nil, fmt.Errorf("google vision error %d: %s", annotate.Error.Code, annotate.Error.Message)
	}
	if annotate.FullTextAnnotation == nil {
		return nil, fmt.Errorf("google vision returned no text annotation")
	}

	text := annotate.FullTextAnnotation.Text
	result := parseVNInvoiceText(text, docType)
	result.RawData = map[string]interface{}{"full_text": text}
	return result, nil
}

// parseVNInvoiceText parses raw OCR text from a Vietnamese invoice into structured OCRResult.
func parseVNInvoiceText(text, docType string) *OCRResult {
	return &OCRResult{
		DocumentType: docType,
		Vendor:       extractVendorFromText(text),
		TotalAmount:  extractTotalFromText(text),
		TaxAmount:    extractTaxFromText(text),
		Date:         extractDateFromText(text),
		Currency:     "VND",
	}
}

// extractVendorFromText returns the first non-empty line that looks like a company name.
func extractVendorFromText(text string) string {
	lines := strings.Split(strings.TrimSpace(text), "\n")

	// Check for explicit "Người bán:" / "Đơn vị bán:" labels
	labelRe := regexp.MustCompile(`(?i)(người bán|đơn vị bán hàng|công ty bán|seller)[:\s]+(.+)`)
	for _, line := range lines {
		if m := labelRe.FindStringSubmatch(line); len(m) > 2 {
			if v := strings.TrimSpace(m[2]); v != "" {
				return v
			}
		}
	}

	// Fall back: first line containing "công ty", "doanh nghiệp", "cty", "tnhh", "cp", "ltd"
	companyRe := regexp.MustCompile(`(?i)(công ty|doanh nghiệp|cty|tnhh|trách nhiệm|cổ phần|ltd|co\.|inc\.)`)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if companyRe.MatchString(line) {
			return line
		}
	}

	// Last resort: return the first non-empty line
	for _, line := range lines {
		if line = strings.TrimSpace(line); line != "" {
			return line
		}
	}
	return ""
}

// extractTotalFromText finds the total payment amount in Vietnamese invoice text.
func extractTotalFromText(text string) float64 {
	// Match Vietnamese total labels, largest number in that line wins
	patterns := []string{
		`(?i)tổng\s*cộng[^:\n]*[:\s]+([\d.,]+)`,
		`(?i)tổng\s*tiền[^:\n]*[:\s]+([\d.,]+)`,
		`(?i)thành\s*tiền[^:\n]*[:\s]+([\d.,]+)`,
		`(?i)total[^:\n]*[:\s]+([\d.,]+)`,
	}
	for _, p := range patterns {
		re := regexp.MustCompile(p)
		if m := re.FindStringSubmatch(text); len(m) > 1 {
			if v := parseVNAmount(m[1]); v > 0 {
				return v
			}
		}
	}
	return 0
}

// extractTaxFromText finds VAT / tax amount in the text.
func extractTaxFromText(text string) float64 {
	patterns := []string{
		`(?i)thuế\s*gtgt[^:\n]*[:\s]+([\d.,]+)`,
		`(?i)vat[^:\n]*[:\s]+([\d.,]+)`,
		`(?i)thuế[^:\n]*[:\s]+([\d.,]+)`,
	}
	for _, p := range patterns {
		re := regexp.MustCompile(p)
		if m := re.FindStringSubmatch(text); len(m) > 1 {
			if v := parseVNAmount(m[1]); v > 0 {
				return v
			}
		}
	}
	return 0
}

// extractDateFromText finds the invoice date in common Vietnamese formats.
func extractDateFromText(text string) string {
	// Match "Ngày DD/MM/YYYY" or "Ngày: DD/MM/YYYY" or standalone date patterns
	patterns := []string{
		`(?i)ngày[:\s]*(\d{1,2}[/-]\d{1,2}[/-]\d{4})`,
		`(?i)date[:\s]*(\d{1,2}[/-]\d{1,2}[/-]\d{4})`,
		`\b(\d{1,2}/\d{1,2}/\d{4})\b`,
		`\b(\d{4}-\d{2}-\d{2})\b`,
	}
	for _, p := range patterns {
		re := regexp.MustCompile(p)
		if m := re.FindStringSubmatch(text); len(m) > 1 {
			return normDate(m[1])
		}
	}
	return time.Now().Format("2006-01-02")
}

// parseVNAmount parses Vietnamese number formats like "10.000.001" or "1,100,000".
// VN uses "." as thousand separator and "," as decimal separator (opposite of US).
// We detect format by checking if the last separator segment length == 3 (thousands).
func parseVNAmount(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}

	// Count dots and commas to determine format
	dots := strings.Count(s, ".")
	commas := strings.Count(s, ",")

	var clean string
	switch {
	case dots > 1 && commas == 0:
		// VN format: "10.000.001" → remove dots
		clean = strings.ReplaceAll(s, ".", "")
	case commas > 1 && dots == 0:
		// US/EU format: "1,100,000" → remove commas
		clean = strings.ReplaceAll(s, ",", "")
	case dots == 1 && commas == 1:
		// Mixed "1.000,50" (VN decimal) or "1,000.50" (US decimal)
		if strings.LastIndex(s, ".") < strings.LastIndex(s, ",") {
			// "1.000,50" → VN decimal
			clean = strings.ReplaceAll(s, ".", "")
			clean = strings.ReplaceAll(clean, ",", ".")
		} else {
			// "1,000.50" → US decimal
			clean = strings.ReplaceAll(s, ",", "")
		}
	default:
		clean = strings.ReplaceAll(strings.ReplaceAll(s, ".", ""), ",", "")
	}

	var v float64
	fmt.Sscanf(clean, "%f", &v)
	return v
}
