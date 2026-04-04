package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/PhanAnh1001/ai-accounting/backend/internal/model"
)

// MISAAdapter is the interface for syncing approved accounting entries to MISA AMIS.
type MISAAdapter interface {
	SyncEntry(ctx context.Context, entry *model.AccountingEntry) error
}

// MockMISAAdapter is a no-op adapter for use in tests and CI environments.
type MockMISAAdapter struct{}

func (m *MockMISAAdapter) SyncEntry(_ context.Context, _ *model.AccountingEntry) error {
	return nil
}

// MISAClient calls the MISA AMIS API to sync accounting entries.
type MISAClient struct {
	apiURL     string
	apiKey     string
	httpClient *http.Client
}

func NewMISAClient(apiURL, apiKey string) *MISAClient {
	return &MISAClient{
		apiURL: apiURL,
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

type misaSyncRequest struct {
	EntryID       string  `json:"entry_id"`
	EntryDate     string  `json:"entry_date"`
	Description   string  `json:"description"`
	DebitAccount  string  `json:"debit_account"`
	CreditAccount string  `json:"credit_account"`
	Amount        float64 `json:"amount"`
}

// SyncEntry sends an approved accounting entry to MISA AMIS API.
func (c *MISAClient) SyncEntry(ctx context.Context, entry *model.AccountingEntry) error {
	payload := misaSyncRequest{
		EntryID:       entry.ID,
		EntryDate:     entry.EntryDate.Format("2006-01-02"),
		Description:   entry.Description,
		DebitAccount:  entry.DebitAccount,
		CreditAccount: entry.CreditAccount,
		Amount:        entry.Amount,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal misa request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.apiURL+"/api/accounting/entries", bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("create misa request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("misa api request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("misa api returned status %d", resp.StatusCode)
	}
	return nil
}
