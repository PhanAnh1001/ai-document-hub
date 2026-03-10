package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"regexp"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/PhanAnh1001/ai-accounting/backend/internal/middleware"
	"github.com/PhanAnh1001/ai-accounting/backend/internal/model"
	"github.com/PhanAnh1001/ai-accounting/backend/internal/service"
)

// EntryRepository is the interface the EntryHandler depends on.
type EntryRepository interface {
	ListByOrg(ctx context.Context, orgID, status, q, dateFrom, dateTo, sortBy, sortDir string, limit, offset int) ([]*model.AccountingEntry, int, error)
	UpdateStatus(ctx context.Context, id, orgID string, status model.EntryStatus) error
	UpdateStatusWithReason(ctx context.Context, id, orgID string, status model.EntryStatus, reason string) error
	UpdateEntry(ctx context.Context, id, orgID string, fields model.UpdateEntryFields) error
	BulkUpdateStatus(ctx context.Context, ids []string, orgID string, status model.EntryStatus) (int, error)
	GetByID(ctx context.Context, id, orgID string) (*model.AccountingEntry, error)
	CountByStatus(ctx context.Context, orgID string) (map[string]int, error)
	SumApprovedAmount(ctx context.Context, orgID string) (float64, error)
	CountByDay(ctx context.Context, orgID string, days int) ([]model.DayCount, error)
}

// EntryHandler handles accounting entry HTTP requests.
type EntryHandler struct {
	repo        EntryRepository
	misaAdapter service.MISAAdapter
}

func NewEntryHandler(repo EntryRepository, misa ...service.MISAAdapter) *EntryHandler {
	h := &EntryHandler{repo: repo}
	if len(misa) > 0 {
		h.misaAdapter = misa[0]
	}
	return h
}

// Stats handles GET /api/entries/stats — returns entry counts by status.
func (h *EntryHandler) Stats(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	if orgID == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	counts, err := h.repo.CountByStatus(r.Context(), orgID)
	if err != nil {
		http.Error(w, `{"error":"failed to get stats"}`, http.StatusInternalServerError)
		return
	}

	approvedAmount, err := h.repo.SumApprovedAmount(r.Context(), orgID)
	if err != nil {
		http.Error(w, `{"error":"failed to get stats"}`, http.StatusInternalServerError)
		return
	}

	trend, err := h.repo.CountByDay(r.Context(), orgID, 7)
	if err != nil {
		http.Error(w, `{"error":"failed to get stats"}`, http.StatusInternalServerError)
		return
	}
	if trend == nil {
		trend = []model.DayCount{}
	}

	total := 0
	for _, v := range counts {
		total += v
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"total":                 total,
		"by_status":             counts,
		"total_approved_amount": approvedAmount,
		"daily_trend":           trend,
	})
}

// List handles GET /api/entries?status=pending&limit=20&offset=0
func (h *EntryHandler) List(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	if orgID == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	status := r.URL.Query().Get("status")
	q := r.URL.Query().Get("q")
	dateFrom := r.URL.Query().Get("date_from")
	dateTo := r.URL.Query().Get("date_to")
	sortBy := r.URL.Query().Get("sort_by")
	sortDir := r.URL.Query().Get("sort_dir")
	limit := queryInt(r, "limit", 20)
	offset := queryInt(r, "offset", 0)

	entries, total, err := h.repo.ListByOrg(r.Context(), orgID, status, q, dateFrom, dateTo, sortBy, sortDir, limit, offset)
	if err != nil {
		http.Error(w, `{"error":"failed to list entries"}`, http.StatusInternalServerError)
		return
	}

	if entries == nil {
		entries = []*model.AccountingEntry{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(model.EntryListResponse{
		Entries: entries,
		Total:   total,
	})
}

// Update handles PATCH /api/entries/{id} — edits fields of a pending entry.
func (h *EntryHandler) Update(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	if orgID == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, `{"error":"missing entry id"}`, http.StatusBadRequest)
		return
	}

	existing, err := h.repo.GetByID(r.Context(), id, orgID)
	if err != nil || existing == nil {
		http.Error(w, `{"error":"entry not found"}`, http.StatusNotFound)
		return
	}

	if existing.Status != model.EntryPending {
		http.Error(w, `{"error":"only pending entries can be edited"}`, http.StatusBadRequest)
		return
	}

	var fields model.UpdateEntryFields
	if err := json.NewDecoder(r.Body).Decode(&fields); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}
	if fields.Description == "" || fields.DebitAccount == "" || fields.CreditAccount == "" || fields.Amount <= 0 {
		http.Error(w, `{"error":"description, debit_account, credit_account and amount are required"}`, http.StatusBadRequest)
		return
	}
	if !vasAccountRe.MatchString(fields.DebitAccount) || !vasAccountRe.MatchString(fields.CreditAccount) {
		http.Error(w, `{"error":"debit_account and credit_account must be 3-4 digit VAS account numbers"}`, http.StatusBadRequest)
		return
	}

	if err := h.repo.UpdateEntry(r.Context(), id, orgID, fields); err != nil {
		http.Error(w, `{"error":"entry not found"}`, http.StatusNotFound)
		return
	}

	entry, err := h.repo.GetByID(r.Context(), id, orgID)
	if err != nil || entry == nil {
		http.Error(w, `{"error":"entry not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(model.EntryActionResponse{Entry: entry})
}

// BulkApprove handles POST /api/entries/bulk-approve — approves multiple pending entries at once.
func (h *EntryHandler) BulkApprove(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	if orgID == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req struct {
		IDs []string `json:"ids"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || len(req.IDs) == 0 {
		http.Error(w, `{"error":"ids must be a non-empty array"}`, http.StatusBadRequest)
		return
	}

	updated, err := h.repo.BulkUpdateStatus(r.Context(), req.IDs, orgID, model.EntryApproved)
	if err != nil {
		http.Error(w, `{"error":"failed to bulk approve"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{"updated": updated})
}

// Get handles GET /api/entries/{id}
func (h *EntryHandler) Get(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	if orgID == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, `{"error":"missing entry id"}`, http.StatusBadRequest)
		return
	}

	entry, err := h.repo.GetByID(r.Context(), id, orgID)
	if err != nil || entry == nil {
		http.Error(w, `{"error":"entry not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(model.EntryActionResponse{Entry: entry})
}

// Approve handles POST /api/entries/{id}/approve
func (h *EntryHandler) Approve(w http.ResponseWriter, r *http.Request) {
	h.updateStatus(w, r, model.EntryApproved)
}

// Reject handles POST /api/entries/{id}/reject
// Accepts optional JSON body: {"reason": "..."}
func (h *EntryHandler) Reject(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	if orgID == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, `{"error":"missing entry id"}`, http.StatusBadRequest)
		return
	}

	// Parse optional reason from body
	var body struct {
		Reason string `json:"reason"`
	}
	// Ignore decode errors — reason is optional
	_ = json.NewDecoder(r.Body).Decode(&body)

	var err error
	if body.Reason != "" {
		err = h.repo.UpdateStatusWithReason(r.Context(), id, orgID, model.EntryRejected, body.Reason)
	} else {
		err = h.repo.UpdateStatus(r.Context(), id, orgID, model.EntryRejected)
	}
	if err != nil {
		http.Error(w, `{"error":"entry not found"}`, http.StatusNotFound)
		return
	}

	entry, err := h.repo.GetByID(r.Context(), id, orgID)
	if err != nil || entry == nil {
		http.Error(w, `{"error":"entry not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(model.EntryActionResponse{Entry: entry})
}

func (h *EntryHandler) updateStatus(w http.ResponseWriter, r *http.Request, status model.EntryStatus) {
	orgID := middleware.GetOrgID(r.Context())
	if orgID == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, `{"error":"missing entry id"}`, http.StatusBadRequest)
		return
	}

	if err := h.repo.UpdateStatus(r.Context(), id, orgID, status); err != nil {
		http.Error(w, `{"error":"entry not found"}`, http.StatusNotFound)
		return
	}

	entry, err := h.repo.GetByID(r.Context(), id, orgID)
	if err != nil || entry == nil {
		http.Error(w, `{"error":"entry not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(model.EntryActionResponse{Entry: entry})
}

// Sync handles POST /api/entries/{id}/sync — sends approved entry to MISA AMIS.
func (h *EntryHandler) Sync(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	if orgID == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, `{"error":"missing entry id"}`, http.StatusBadRequest)
		return
	}

	entry, err := h.repo.GetByID(r.Context(), id, orgID)
	if err != nil || entry == nil {
		http.Error(w, `{"error":"entry not found"}`, http.StatusNotFound)
		return
	}

	if entry.Status != model.EntryApproved {
		http.Error(w, `{"error":"only approved entries can be synced"}`, http.StatusBadRequest)
		return
	}

	if h.misaAdapter != nil {
		if err := h.misaAdapter.SyncEntry(r.Context(), entry); err != nil {
			http.Error(w, `{"error":"failed to sync to MISA"}`, http.StatusBadGateway)
			return
		}
	}

	if err := h.repo.UpdateStatus(r.Context(), id, orgID, model.EntrySynced); err != nil {
		http.Error(w, `{"error":"failed to update entry status"}`, http.StatusInternalServerError)
		return
	}

	entry.Status = model.EntrySynced
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(model.EntryActionResponse{Entry: entry})
}

// vasAccountRe matches 3-4 digit Vietnamese accounting standard (VAS) account numbers.
var vasAccountRe = regexp.MustCompile(`^\d{3,4}$`)

const maxListLimit = 100

func queryInt(r *http.Request, key string, defaultVal int) int {
	v := r.URL.Query().Get(key)
	if v == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(v)
	if err != nil || n < 0 {
		return defaultVal
	}
	if key == "limit" && n > maxListLimit {
		return maxListLimit
	}
	return n
}
