package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/PhanAnh1001/ai-accounting/backend/internal/model"
)

// DocumentUpdater is the subset of DocumentRepository used by the processor.
type DocumentUpdater interface {
	UpdateStatus(ctx context.Context, id string, status model.DocumentStatus) error
	UpdateOCRData(ctx context.Context, id string, ocrData []byte, status model.DocumentStatus) error
}

// EntryCreator is the subset of EntryRepository used by the processor.
type EntryCreator interface {
	CreateBatch(ctx context.Context, entries []*model.AccountingEntry) error
}

// AIEnricher is the interface for Claude-based OCR enrichment.
// Separated from the concrete type to allow mocking in tests.
type AIEnricher interface {
	Enrich(ctx context.Context, ocrText, docType string) (*ClaudeEnrichResult, error)
}

// ProcessorService orchestrates the document processing pipeline:
// document → OCR → Claude enricher (optional) → rule engine fallback → accounting entries
type ProcessorService struct {
	ocr       OCRService
	rules     *VASRuleEngine
	claude    AIEnricher // nil = disabled, falls back to VAS rules
	docRepo   DocumentUpdater
	entryRepo EntryCreator
}

func NewProcessorService(
	ocr OCRService,
	rules *VASRuleEngine,
	docRepo DocumentUpdater,
	entryRepo EntryCreator,
) *ProcessorService {
	return &ProcessorService{
		ocr:       ocr,
		rules:     rules,
		docRepo:   docRepo,
		entryRepo: entryRepo,
	}
}

// WithClaudeEnricher attaches an AI enricher to the processor.
// When set, it is called after OCR to improve extraction and suggest accounts.
// If it fails, the processor falls back to the VAS rule engine automatically.
func (p *ProcessorService) WithClaudeEnricher(e AIEnricher) *ProcessorService {
	p.claude = e
	return p
}

// Process runs the full pipeline for a document.
// On OCR error, the document status is set to "error" and the error is logged (not returned).
func (p *ProcessorService) Process(ctx context.Context, doc *model.Document) error {
	// Step 1: mark processing
	if err := p.docRepo.UpdateStatus(ctx, doc.ID, model.StatusProcessing); err != nil {
		log.Printf("processor: update status processing doc=%s: %v", doc.ID, err)
		return nil
	}

	// Step 2: call OCR or XML parser based on file extension
	extractor := p.ocr
	if strings.HasSuffix(strings.ToLower(doc.FileURL), ".xml") {
		extractor = NewXMLInvoiceParser()
	}
	ocrResult, err := extractor.ExtractFromFile(ctx, doc.FileURL, string(doc.DocumentType))
	if err != nil {
		log.Printf("processor: OCR failed doc=%s: %v", doc.ID, err)
		_ = p.docRepo.UpdateStatus(ctx, doc.ID, model.StatusError)
		return nil
	}

	// Step 3: persist OCR data
	ocrJSON, _ := json.Marshal(ocrResult)
	if err := p.docRepo.UpdateOCRData(ctx, doc.ID, ocrJSON, model.StatusExtracted); err != nil {
		log.Printf("processor: save OCR data doc=%s: %v", doc.ID, err)
		_ = p.docRepo.UpdateStatus(ctx, doc.ID, model.StatusError)
		return nil
	}

	// Step 4: generate accounting entries via Claude (preferred) or VAS rule engine (fallback)
	entries := p.generateEntries(ctx, ocrResult, doc)

	// Step 5: persist entries
	if err := p.entryRepo.CreateBatch(ctx, entries); err != nil {
		log.Printf("processor: create entries doc=%s: %v", doc.ID, err)
		_ = p.docRepo.UpdateStatus(ctx, doc.ID, model.StatusError)
		return nil
	}

	// Step 6: mark document as booked
	if err := p.docRepo.UpdateStatus(ctx, doc.ID, model.StatusBooked); err != nil {
		log.Printf("processor: update status booked doc=%s: %v", doc.ID, err)
	}
	return nil
}

// generateEntries tries Claude first (if configured), falls back to VAS rule engine.
func (p *ProcessorService) generateEntries(ctx context.Context, ocrResult *OCRResult, doc *model.Document) []*model.AccountingEntry {
	if p.claude != nil {
		rawText := extractRawText(ocrResult)
		enriched, err := p.claude.Enrich(ctx, rawText, string(doc.DocumentType))
		if err != nil {
			log.Printf("processor: claude enricher failed doc=%s, falling back to VAS rules: %v", doc.ID, err)
		} else {
			return claudeResultToEntries(enriched, doc.ID, doc.OrganizationID)
		}
	}

	// Fallback: keyword-based VAS rule engine
	rawEntries := p.rules.Apply(ocrResult, string(doc.DocumentType), doc.ID, doc.OrganizationID)
	entries := make([]*model.AccountingEntry, len(rawEntries))
	for i := range rawEntries {
		entries[i] = &rawEntries[i]
	}
	return entries
}

// extractRawText gets the best available text from the OCR result.
func extractRawText(r *OCRResult) string {
	if t, ok := r.RawData["full_text"].(string); ok && t != "" {
		return t
	}
	// Reconstruct a minimal text summary for Claude from structured fields
	return strings.Join([]string{
		"Vendor: " + r.Vendor,
		"Date: " + r.Date,
		"Total: " + fmt.Sprintf("%.0f", r.TotalAmount),
		"Tax: " + fmt.Sprintf("%.0f", r.TaxAmount),
	}, "\n")
}

// claudeResultToEntries converts Claude's suggestion into AccountingEntry records.
func claudeResultToEntries(r *ClaudeEnrichResult, docID, orgID string) []*model.AccountingEntry {
	entries := make([]*model.AccountingEntry, 0, len(r.Entries))
	for _, line := range r.Entries {
		conf := line.Confidence
		e := makeEntry(docID, orgID, r.Date, line.Description, line.DebitAccount, line.CreditAccount, line.Amount, &conf)
		entries = append(entries, &e)
	}
	return entries
}
