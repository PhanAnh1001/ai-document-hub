package handler

import (
	"context"
	"encoding/csv"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/PhanAnh1001/ai-accounting/backend/internal/model"
)

func sampleExportEntry(id string, status model.EntryStatus) *model.AccountingEntry {
	conf := 0.90
	docID := "doc-1"
	return &model.AccountingEntry{
		ID:             id,
		OrganizationID: "org-1",
		DocumentID:     &docID,
		EntryDate:      time.Date(2026, 3, 28, 0, 0, 0, 0, time.UTC),
		Description:    "Mua hàng hóa",
		DebitAccount:   "156",
		CreditAccount:  "331",
		Amount:         1500000,
		Status:         status,
		AIConfidence:   &conf,
		CreatedAt:      time.Now(),
	}
}

func TestExportHandler_ExportCSV_ReturnsCSV(t *testing.T) {
	entries := []*model.AccountingEntry{
		sampleExportEntry("e-1", model.EntryApproved),
		sampleExportEntry("e-2", model.EntryApproved),
	}
	repo := &mockEntryRepository{
		listByOrgFn: func(_ context.Context, orgID, status, _, _, _, _, _ string, limit, offset int) ([]*model.AccountingEntry, int, error) {
			return entries, len(entries), nil
		},
	}
	h := NewExportHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/entries/export?status=approved", nil)
	req = buildEntryAuthContext(req, "org-1")
	rr := httptest.NewRecorder()

	h.ExportCSV(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	ct := rr.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/csv") {
		t.Errorf("expected Content-Type text/csv, got %s", ct)
	}
	cd := rr.Header().Get("Content-Disposition")
	if !strings.Contains(cd, "attachment") {
		t.Errorf("expected Content-Disposition attachment, got %s", cd)
	}

	r := csv.NewReader(rr.Body)
	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("parse CSV: %v", err)
	}
	// Header row + 2 data rows
	if len(records) != 3 {
		t.Errorf("expected 3 rows (header+2), got %d", len(records))
	}
	// Verify header fields
	header := records[0]
	if header[0] != "id" {
		t.Errorf("expected header[0]='id', got %q", header[0])
	}
	// Verify first data row
	if records[1][0] != "e-1" {
		t.Errorf("expected first row id='e-1', got %q", records[1][0])
	}
}

func TestExportHandler_ExportCSV_Unauthorized(t *testing.T) {
	h := NewExportHandler(&mockEntryRepository{})

	req := httptest.NewRequest(http.MethodGet, "/api/entries/export", nil)
	rr := httptest.NewRecorder()

	h.ExportCSV(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestExportHandler_ExportCSV_FiltersByDateRange(t *testing.T) {
	entries := []*model.AccountingEntry{
		{ID: "e-in", EntryDate: time.Date(2026, 3, 15, 0, 0, 0, 0, time.UTC), Status: model.EntryApproved,
			DebitAccount: "156", CreditAccount: "331", Description: "In range"},
		{ID: "e-out", EntryDate: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC), Status: model.EntryApproved,
			DebitAccount: "156", CreditAccount: "331", Description: "Out of range"},
	}
	repo := &mockEntryRepository{
		listByOrgFn: func(_ context.Context, _, _, _, _, _, _, _ string, _, _ int) ([]*model.AccountingEntry, int, error) {
			return entries, len(entries), nil
		},
	}
	h := NewExportHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/entries/export?from=2026-03-01&to=2026-03-31", nil)
	req = buildEntryAuthContext(req, "org-1")
	rr := httptest.NewRecorder()

	h.ExportCSV(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	r := csv.NewReader(rr.Body)
	records, _ := r.ReadAll()
	// header + 1 matching entry
	if len(records) != 2 {
		t.Errorf("expected 2 rows (header + 1 entry in range), got %d", len(records))
	}
	if len(records) > 1 && !strings.Contains(records[1][0], "e-in") {
		t.Errorf("expected e-in in first data row, got %s", records[1][0])
	}
}

func TestExportHandler_ExportCSV_ReturnsExportCountHeader(t *testing.T) {
	entries := []*model.AccountingEntry{
		sampleExportEntry("e-1", model.EntryApproved),
		sampleExportEntry("e-2", model.EntryApproved),
		sampleExportEntry("e-3", model.EntryApproved),
	}
	repo := &mockEntryRepository{
		listByOrgFn: func(_ context.Context, _, _, _, _, _, _, _ string, _, _ int) ([]*model.AccountingEntry, int, error) {
			return entries, len(entries), nil
		},
	}
	h := NewExportHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/entries/export?status=approved", nil)
	req = buildEntryAuthContext(req, "org-1")
	rr := httptest.NewRecorder()

	h.ExportCSV(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	count := rr.Header().Get("X-Export-Count")
	if count != "3" {
		t.Errorf("expected X-Export-Count=3, got %q", count)
	}
}

func TestExportHandler_ExportCSV_EmptyReturnsHeaderOnly(t *testing.T) {
	h := NewExportHandler(&mockEntryRepository{})

	req := httptest.NewRequest(http.MethodGet, "/api/entries/export", nil)
	req = buildEntryAuthContext(req, "org-1")
	rr := httptest.NewRecorder()

	h.ExportCSV(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	r := csv.NewReader(rr.Body)
	records, err := r.ReadAll()
	if err != nil {
		t.Fatalf("parse CSV: %v", err)
	}
	// Only header row
	if len(records) != 1 {
		t.Errorf("expected 1 row (header only), got %d", len(records))
	}
}

func TestExportHandler_ExportCSV_PassesSearchQueryToRepo(t *testing.T) {
	var capturedQuery string
	repo := &mockEntryRepository{
		listByOrgFn: func(_ context.Context, orgID, status, q, _, _, _, _ string, limit, offset int) ([]*model.AccountingEntry, int, error) {
			capturedQuery = q
			return []*model.AccountingEntry{}, 0, nil
		},
	}
	h := NewExportHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/entries/export?q=mua+h%C3%A0ng", nil)
	req = buildEntryAuthContext(req, "org-1")
	rr := httptest.NewRecorder()

	h.ExportCSV(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if capturedQuery != "mua hàng" {
		t.Errorf("expected q='mua hàng', got %q", capturedQuery)
	}
}
