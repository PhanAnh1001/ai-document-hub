package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"
	"time"
)

const claudeModel = "claude-haiku-4-5-20251001"

const claudeSystemPrompt = `You are an expert Vietnamese accountant specializing in VAS (Vietnamese Accounting Standards).
Given OCR text from a financial document, extract key fields and suggest the correct double-entry accounting lines.

Respond ONLY with valid JSON in this exact format (no markdown, no explanation):
{
  "vendor": "company or person name",
  "total_amount": 1000000,
  "tax_amount": 0,
  "date": "YYYY-MM-DD",
  "entries": [
    {
      "debit_account": "156",
      "credit_account": "331",
      "amount": 1000000,
      "description": "brief Vietnamese description",
      "confidence": 0.95
    }
  ]
}

VAS account rules:
- Purchase invoice (mua hàng): debit 156, credit 331 (goods); debit 133, credit 331 (VAT if present)
- Sales invoice (bán hàng): debit 131, credit 511 (revenue); debit 131, credit 3331 (VAT if present)
- Cash payment receipt (phiếu chi): debit 331, credit 111
- Cash receipt (phiếu thu): debit 111, credit 131
- Bank debit (ủy nhiệm chi): debit 331, credit 112
- Bank credit (báo có): debit 112, credit 131
- Salary (lương): debit 622 or 641 or 642, credit 334
- Utilities electricity/water (điện/nước): debit 6278, credit 331
- Other expense: debit 642, credit 331

Account numbers must be 3-4 digits only. date must be YYYY-MM-DD.`

// ClaudeEnrichResult is the structured output from Claude after analyzing OCR text.
type ClaudeEnrichResult struct {
	Vendor      string             `json:"vendor"`
	TotalAmount float64            `json:"total_amount"`
	TaxAmount   float64            `json:"tax_amount"`
	Date        string             `json:"date"`
	Entries     []ClaudeEntryLine  `json:"entries"`
}

// ClaudeEntryLine is a single suggested accounting entry line from Claude.
type ClaudeEntryLine struct {
	DebitAccount  string  `json:"debit_account"`
	CreditAccount string  `json:"credit_account"`
	Amount        float64 `json:"amount"`
	Description   string  `json:"description"`
	Confidence    float64 `json:"confidence"`
}

// --- Anthropic API request/response types ---

type anthropicMessagesRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	System    string             `json:"system"`
	Messages  []anthropicMessage `json:"messages"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicMessagesResponse struct {
	Content []anthropicContent `json:"content"`
	Error   *anthropicError    `json:"error,omitempty"`
}

type anthropicContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type anthropicError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

// ClaudeEnricher calls the Anthropic API to enrich OCR output with structured
// data extraction and VAS account suggestions.
type ClaudeEnricher struct {
	apiKey     string
	baseURL    string
	httpClient *http.Client
}

// NewClaudeEnricher creates a production enricher using the Anthropic API.
func NewClaudeEnricher(apiKey string) *ClaudeEnricher {
	return &ClaudeEnricher{
		apiKey:  apiKey,
		baseURL: "https://api.anthropic.com/v1/messages",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// newClaudeEnricherWithBaseURL creates an enricher with a custom URL (for tests).
func newClaudeEnricherWithBaseURL(apiKey, baseURL string, httpClient *http.Client) *ClaudeEnricher {
	return &ClaudeEnricher{
		apiKey:     apiKey,
		baseURL:    baseURL,
		httpClient: httpClient,
	}
}

// Enrich sends OCR text to Claude and returns structured accounting data.
func (e *ClaudeEnricher) Enrich(ctx context.Context, ocrText, docType string) (*ClaudeEnrichResult, error) {
	userMsg := fmt.Sprintf("Document type: %s\n\nOCR text:\n%s", docType, ocrText)

	reqBody := anthropicMessagesRequest{
		Model:     claudeModel,
		MaxTokens: 1024,
		System:    claudeSystemPrompt,
		Messages: []anthropicMessage{
			{Role: "user", Content: userMsg},
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal claude request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.baseURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("create claude request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", e.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("claude api request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("claude api returned HTTP %d", resp.StatusCode)
	}

	var apiResp anthropicMessagesResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, fmt.Errorf("decode claude response: %w", err)
	}
	if apiResp.Error != nil {
		return nil, fmt.Errorf("claude api error: %s", apiResp.Error.Message)
	}
	if len(apiResp.Content) == 0 {
		return nil, fmt.Errorf("claude returned empty content")
	}

	rawText := apiResp.Content[0].Text
	rawText = stripMarkdownCodeBlock(rawText)

	var result ClaudeEnrichResult
	if err := json.Unmarshal([]byte(rawText), &result); err != nil {
		return nil, fmt.Errorf("parse claude JSON output: %w", err)
	}

	if err := validateEnrichResult(&result); err != nil {
		return nil, fmt.Errorf("invalid claude result: %w", err)
	}

	return &result, nil
}

// stripMarkdownCodeBlock removes ```json ... ``` or ``` ... ``` wrappers that
// Claude occasionally adds around JSON output.
func stripMarkdownCodeBlock(s string) string {
	s = strings.TrimSpace(s)
	re := regexp.MustCompile("(?s)^```(?:json)?\\s*\\n?(.*?)\\n?```$")
	if m := re.FindStringSubmatch(s); len(m) > 1 {
		return strings.TrimSpace(m[1])
	}
	return s
}

// validateEnrichResult checks that Claude returned sensible VAS account numbers.
var vasAccountRe = regexp.MustCompile(`^\d{3,4}$`)

func validateEnrichResult(r *ClaudeEnrichResult) error {
	if len(r.Entries) == 0 {
		return fmt.Errorf("no entries suggested")
	}
	for i, e := range r.Entries {
		if !vasAccountRe.MatchString(e.DebitAccount) {
			return fmt.Errorf("entry %d: invalid debit account %q (must be 3-4 digits)", i, e.DebitAccount)
		}
		if !vasAccountRe.MatchString(e.CreditAccount) {
			return fmt.Errorf("entry %d: invalid credit account %q (must be 3-4 digits)", i, e.CreditAccount)
		}
	}
	return nil
}
