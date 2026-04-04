package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// buildAnthropicResponse builds a fake Anthropic API response wrapping the given text.
func buildAnthropicResponse(text string) anthropicMessagesResponse {
	return anthropicMessagesResponse{
		Content: []anthropicContent{
			{Type: "text", Text: text},
		},
	}
}

func TestClaudeEnricher_EnrichInvoice_Success(t *testing.T) {
	enrichJSON := `{
		"vendor": "Công ty TNHH ABC",
		"total_amount": 1100000,
		"tax_amount": 100000,
		"date": "2026-03-28",
		"entries": [
			{"debit_account":"156","credit_account":"331","amount":1000000,"description":"Mua hàng hóa - Công ty TNHH ABC","confidence":0.95},
			{"debit_account":"133","credit_account":"331","amount":100000,"description":"Thuế GTGT đầu vào","confidence":0.95}
		]
	}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(buildAnthropicResponse(enrichJSON))
	}))
	defer srv.Close()

	enricher := newClaudeEnricherWithBaseURL("test-key", srv.URL, srv.Client())
	result, err := enricher.Enrich(context.Background(), "raw ocr text of invoice", "invoice")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Vendor != "Công ty TNHH ABC" {
		t.Errorf("expected vendor 'Công ty TNHH ABC', got %q", result.Vendor)
	}
	if result.TotalAmount != 1100000 {
		t.Errorf("expected total 1100000, got %f", result.TotalAmount)
	}
	if result.TaxAmount != 100000 {
		t.Errorf("expected tax 100000, got %f", result.TaxAmount)
	}
	if len(result.Entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(result.Entries))
	}
	if result.Entries[0].DebitAccount != "156" {
		t.Errorf("expected debit 156, got %s", result.Entries[0].DebitAccount)
	}
}

func TestClaudeEnricher_StripsMarkdownCodeBlock(t *testing.T) {
	// Claude sometimes wraps JSON in ```json ... ``` blocks
	enrichJSON := "```json\n{\"vendor\":\"ABC\",\"total_amount\":500000,\"tax_amount\":0,\"date\":\"2026-03-28\",\"entries\":[{\"debit_account\":\"642\",\"credit_account\":\"331\",\"amount\":500000,\"description\":\"Chi phí\",\"confidence\":0.8}]}\n```"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(buildAnthropicResponse(enrichJSON))
	}))
	defer srv.Close()

	enricher := newClaudeEnricherWithBaseURL("key", srv.URL, srv.Client())
	result, err := enricher.Enrich(context.Background(), "some text", "other")
	if err != nil {
		t.Fatalf("unexpected error stripping markdown: %v", err)
	}
	if result.Vendor != "ABC" {
		t.Errorf("expected vendor 'ABC', got %q", result.Vendor)
	}
}

func TestClaudeEnricher_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	enricher := newClaudeEnricherWithBaseURL("key", srv.URL, srv.Client())
	_, err := enricher.Enrich(context.Background(), "text", "invoice")
	if err == nil {
		t.Error("expected error for HTTP 500")
	}
}

func TestClaudeEnricher_MalformedJSON_FromClaude(t *testing.T) {
	// Claude returns garbage instead of JSON
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(buildAnthropicResponse("I cannot process this document."))
	}))
	defer srv.Close()

	enricher := newClaudeEnricherWithBaseURL("key", srv.URL, srv.Client())
	_, err := enricher.Enrich(context.Background(), "text", "invoice")
	if err == nil {
		t.Error("expected error when Claude returns non-JSON text")
	}
}

func TestClaudeEnricher_EmptyContent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(anthropicMessagesResponse{Content: []anthropicContent{}})
	}))
	defer srv.Close()

	enricher := newClaudeEnricherWithBaseURL("key", srv.URL, srv.Client())
	_, err := enricher.Enrich(context.Background(), "text", "invoice")
	if err == nil {
		t.Error("expected error for empty content array")
	}
}

func TestClaudeEnricher_ValidatesEntryAccounts(t *testing.T) {
	// Claude returns invalid account numbers — should error
	enrichJSON := `{
		"vendor":"X","total_amount":100,"tax_amount":0,"date":"2026-03-28",
		"entries":[{"debit_account":"INVALID","credit_account":"331","amount":100,"description":"test","confidence":0.5}]
	}`

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(buildAnthropicResponse(enrichJSON))
	}))
	defer srv.Close()

	enricher := newClaudeEnricherWithBaseURL("key", srv.URL, srv.Client())
	_, err := enricher.Enrich(context.Background(), "text", "invoice")
	if err == nil {
		t.Error("expected error for invalid VAS account number")
	}
}
