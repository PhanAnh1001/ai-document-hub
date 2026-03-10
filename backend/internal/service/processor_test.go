package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/PhanAnh1001/ai-accounting/backend/internal/model"
)

// stubAIEnricher implements AIEnricher for tests.
type stubAIEnricher struct {
	result *ClaudeEnrichResult
	err    error
}

func (s *stubAIEnricher) Enrich(_ context.Context, _, _ string) (*ClaudeEnrichResult, error) {
	return s.result, s.err
}

// mockDocumentUpdater implements DocumentUpdater for tests.
type mockDocumentUpdater struct {
	updateStatusFn  func(ctx context.Context, id string, status model.DocumentStatus) error
	updateOCRDataFn func(ctx context.Context, id string, ocrData []byte, status model.DocumentStatus) error
}

func (m *mockDocumentUpdater) UpdateStatus(ctx context.Context, id string, status model.DocumentStatus) error {
	if m.updateStatusFn != nil {
		return m.updateStatusFn(ctx, id, status)
	}
	return nil
}

func (m *mockDocumentUpdater) UpdateOCRData(ctx context.Context, id string, ocrData []byte, status model.DocumentStatus) error {
	if m.updateOCRDataFn != nil {
		return m.updateOCRDataFn(ctx, id, ocrData, status)
	}
	return nil
}

// mockEntryCreator implements EntryCreator for tests.
type mockEntryCreator struct {
	createBatchFn func(ctx context.Context, entries []*model.AccountingEntry) error
	calledWith    []*model.AccountingEntry
}

func (m *mockEntryCreator) CreateBatch(ctx context.Context, entries []*model.AccountingEntry) error {
	m.calledWith = entries
	if m.createBatchFn != nil {
		return m.createBatchFn(ctx, entries)
	}
	return nil
}

func sampleDocument() *model.Document {
	return &model.Document{
		ID:             "doc-1",
		OrganizationID: "org-1",
		FileURL:        "/tmp/test-invoice.pdf",
		DocumentType:   model.DocInvoice,
		Status:         model.StatusUploaded,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}
}

func TestProcessorService_HappyPath(t *testing.T) {
	var statusUpdates []model.DocumentStatus
	docUpdater := &mockDocumentUpdater{
		updateStatusFn: func(_ context.Context, _ string, status model.DocumentStatus) error {
			statusUpdates = append(statusUpdates, status)
			return nil
		},
	}
	entryCreator := &mockEntryCreator{}
	ocrSvc := &MockOCRService{Result: DefaultMockOCRResult()}
	ruleEngine := NewVASRuleEngine()

	processor := NewProcessorService(ocrSvc, ruleEngine, docUpdater, entryCreator)
	err := processor.Process(context.Background(), sampleDocument())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should have updated status: processing → booked
	if len(statusUpdates) < 1 {
		t.Error("expected at least 1 status update")
	}
	lastStatus := statusUpdates[len(statusUpdates)-1]
	if lastStatus != model.StatusBooked {
		t.Errorf("expected final status 'booked', got %s", lastStatus)
	}

	// Should have created entries
	if len(entryCreator.calledWith) == 0 {
		t.Error("expected CreateBatch to be called with entries")
	}
}

func TestProcessorService_OCRError_SetsErrorStatus(t *testing.T) {
	var statusUpdates []model.DocumentStatus
	docUpdater := &mockDocumentUpdater{
		updateStatusFn: func(_ context.Context, _ string, status model.DocumentStatus) error {
			statusUpdates = append(statusUpdates, status)
			return nil
		},
	}
	entryCreator := &mockEntryCreator{}
	ocrSvc := &MockOCRService{Err: errors.New("OCR service unavailable")}
	ruleEngine := NewVASRuleEngine()

	processor := NewProcessorService(ocrSvc, ruleEngine, docUpdater, entryCreator)
	err := processor.Process(context.Background(), sampleDocument())

	// Processor should NOT return the error to caller (errors are absorbed)
	_ = err

	// Status should end as "error"
	found := false
	for _, s := range statusUpdates {
		if s == model.StatusError {
			found = true
		}
	}
	if !found {
		t.Errorf("expected status 'error' in updates, got %v", statusUpdates)
	}

	// No entries should be created
	if len(entryCreator.calledWith) > 0 {
		t.Error("expected no entries created on OCR error")
	}
}

func TestProcessorService_WithClaude_UsesClaudeEntries(t *testing.T) {
	entryCreator := &mockEntryCreator{}
	ocrSvc := &MockOCRService{Result: DefaultMockOCRResult()}
	docUpdater := &mockDocumentUpdater{}

	enrichResult := &ClaudeEnrichResult{
		Vendor: "Công ty Test", TotalAmount: 1100000, TaxAmount: 100000, Date: "2026-03-28",
		Entries: []ClaudeEntryLine{
			{DebitAccount: "156", CreditAccount: "331", Amount: 1000000, Description: "Mua hàng", Confidence: 0.95},
			{DebitAccount: "133", CreditAccount: "331", Amount: 100000, Description: "Thuế GTGT", Confidence: 0.95},
		},
	}

	processor := NewProcessorService(ocrSvc, NewVASRuleEngine(), docUpdater, entryCreator)
	processor.WithClaudeEnricher(&stubAIEnricher{result: enrichResult})
	processor.Process(context.Background(), sampleDocument())

	if len(entryCreator.calledWith) != 2 {
		t.Errorf("expected 2 entries from Claude, got %d", len(entryCreator.calledWith))
	}
	if entryCreator.calledWith[0].DebitAccount != "156" {
		t.Errorf("expected debit 156, got %s", entryCreator.calledWith[0].DebitAccount)
	}
}

func TestProcessorService_ClaudeFails_FallsBackToVASRules(t *testing.T) {
	entryCreator := &mockEntryCreator{}
	ocrSvc := &MockOCRService{Result: DefaultMockOCRResult()}
	docUpdater := &mockDocumentUpdater{}

	processor := NewProcessorService(ocrSvc, NewVASRuleEngine(), docUpdater, entryCreator)
	processor.WithClaudeEnricher(&stubAIEnricher{err: errors.New("claude unavailable")})
	processor.Process(context.Background(), sampleDocument())

	if len(entryCreator.calledWith) == 0 {
		t.Error("expected entries from VAS rule fallback when Claude fails")
	}
}

func TestProcessorService_CreatesEntriesWithOrgID(t *testing.T) {
	entryCreator := &mockEntryCreator{}
	ocrSvc := &MockOCRService{Result: DefaultMockOCRResult()}
	ruleEngine := NewVASRuleEngine()
	docUpdater := &mockDocumentUpdater{}

	processor := NewProcessorService(ocrSvc, ruleEngine, docUpdater, entryCreator)

	doc := sampleDocument()
	doc.OrganizationID = "org-999"
	processor.Process(context.Background(), doc)

	for _, e := range entryCreator.calledWith {
		if e.OrganizationID != "org-999" {
			t.Errorf("expected org-999, got %s", e.OrganizationID)
		}
	}
}
