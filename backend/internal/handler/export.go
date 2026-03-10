package handler

import (
	"context"
	"encoding/csv"
	"fmt"
	"net/http"
	"time"

	"github.com/PhanAnh1001/ai-accounting/backend/internal/middleware"
	"github.com/PhanAnh1001/ai-accounting/backend/internal/model"
)

// ExportEntryRepository is the subset of EntryRepository needed for export.
type ExportEntryRepository interface {
	ListByOrg(ctx context.Context, orgID, status, q, dateFrom, dateTo, sortBy, sortDir string, limit, offset int) ([]*model.AccountingEntry, int, error)
}

// ExportHandler handles data export endpoints.
type ExportHandler struct {
	repo ExportEntryRepository
}

func NewExportHandler(repo ExportEntryRepository) *ExportHandler {
	return &ExportHandler{repo: repo}
}

// ExportCSV handles GET /api/entries/export?status=approved&from=2026-01-01&to=2026-03-31
// Returns all matching entries as a CSV file. Optionally filters by entry_date range.
func (h *ExportHandler) ExportCSV(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	if orgID == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	status := r.URL.Query().Get("status")

	// Parse optional date range filters (format: YYYY-MM-DD)
	var fromDate, toDate *time.Time
	if v := r.URL.Query().Get("from"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			fromDate = &t
		}
	}
	if v := r.URL.Query().Get("to"); v != "" {
		if t, err := time.Parse("2006-01-02", v); err == nil {
			// Include the entire "to" day
			end := t.Add(24*time.Hour - time.Second)
			toDate = &end
		}
	}

	// Fetch all entries matching current filters (large limit to get all for export)
	q := r.URL.Query().Get("q")
	entries, _, err := h.repo.ListByOrg(r.Context(), orgID, status, q, "", "", "", "", 10000, 0)
	if err != nil {
		http.Error(w, `{"error":"failed to fetch entries"}`, http.StatusInternalServerError)
		return
	}
	if entries == nil {
		entries = []*model.AccountingEntry{}
	}

	// Apply date range filter in-memory when specified
	if fromDate != nil || toDate != nil {
		filtered := entries[:0]
		for _, e := range entries {
			if fromDate != nil && e.EntryDate.Before(*fromDate) {
				continue
			}
			if toDate != nil && e.EntryDate.After(*toDate) {
				continue
			}
			filtered = append(filtered, e)
		}
		entries = filtered
	}

	today := time.Now().UTC().Format("2006-01-02")
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"entries-%s.csv\"", today))
	w.Header().Set("X-Export-Count", fmt.Sprintf("%d", len(entries)))
	w.Header().Set("Access-Control-Expose-Headers", "X-Export-Count")

	cw := csv.NewWriter(w)
	defer cw.Flush()

	// Write header
	cw.Write([]string{"id", "entry_date", "description", "debit_account", "credit_account", "amount", "status", "ai_confidence", "document_id"})

	for _, e := range entries {
		confidence := ""
		if e.AIConfidence != nil {
			confidence = fmt.Sprintf("%.2f", *e.AIConfidence)
		}
		docID := ""
		if e.DocumentID != nil {
			docID = *e.DocumentID
		}
		cw.Write([]string{
			e.ID,
			e.EntryDate.Format("2006-01-02"),
			e.Description,
			e.DebitAccount,
			e.CreditAccount,
			fmt.Sprintf("%.2f", e.Amount),
			string(e.Status),
			confidence,
			docID,
		})
	}
}
