package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/PhanAnh1001/ai-accounting/backend/internal/model"
)

func sampleAccountingEntry() *model.AccountingEntry {
	conf := 0.92
	docID := "doc-1"
	return &model.AccountingEntry{
		ID:             "entry-1",
		OrganizationID: "org-1",
		DocumentID:     &docID,
		EntryDate:      time.Date(2026, 3, 28, 0, 0, 0, 0, time.UTC),
		Description:    "Mua hàng hóa nhập kho",
		DebitAccount:   "156",
		CreditAccount:  "331",
		Amount:         11000000,
		Status:         model.EntryApproved,
		AIConfidence:   &conf,
		CreatedAt:      time.Now(),
	}
}

func TestMockMISAAdapter_SyncEntry_NoError(t *testing.T) {
	adapter := &MockMISAAdapter{}
	err := adapter.SyncEntry(context.Background(), sampleAccountingEntry())
	if err != nil {
		t.Errorf("expected no error from mock, got %v", err)
	}
}

func TestMISAClient_SyncEntry_CallsAPI(t *testing.T) {
	var received map[string]interface{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "wrong method", http.StatusMethodNotAllowed)
			return
		}
		if r.Header.Get("X-API-Key") == "" {
			http.Error(w, "missing api key", http.StatusUnauthorized)
			return
		}
		json.NewDecoder(r.Body).Decode(&received)
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "synced"})
	}))
	defer server.Close()

	client := NewMISAClient(server.URL, "test-api-key")
	entry := sampleAccountingEntry()
	err := client.SyncEntry(context.Background(), entry)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if received == nil {
		t.Fatal("expected request body to be sent")
	}
}

func TestMISAClient_SyncEntry_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "server error", http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewMISAClient(server.URL, "test-api-key")
	err := client.SyncEntry(context.Background(), sampleAccountingEntry())
	if err == nil {
		t.Error("expected error for server 500")
	}
}

func TestMISAClient_SyncEntry_Unreachable(t *testing.T) {
	// Use a non-existent server
	client := NewMISAClient("http://localhost:19999", "test-api-key")
	err := client.SyncEntry(context.Background(), sampleAccountingEntry())
	if err == nil {
		t.Error("expected error for unreachable server")
	}
}
