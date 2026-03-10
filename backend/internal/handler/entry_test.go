package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/PhanAnh1001/ai-accounting/backend/internal/middleware"
	"github.com/PhanAnh1001/ai-accounting/backend/internal/model"
)

// mockEntryRepository implements EntryRepository for testing.
type mockEntryRepository struct {
	listByOrgFn               func(ctx context.Context, orgID, status, q, dateFrom, dateTo, sortBy, sortDir string, limit, offset int) ([]*model.AccountingEntry, int, error)
	updateStatusFn            func(ctx context.Context, id, orgID string, status model.EntryStatus) error
	updateStatusWithReasonFn  func(ctx context.Context, id, orgID string, status model.EntryStatus, reason string) error
	updateEntryFn             func(ctx context.Context, id, orgID string, fields model.UpdateEntryFields) error
	bulkUpdateStatusFn        func(ctx context.Context, ids []string, orgID string, status model.EntryStatus) (int, error)
	getByIDFn                 func(ctx context.Context, id, orgID string) (*model.AccountingEntry, error)
	countByStatusFn           func(ctx context.Context, orgID string) (map[string]int, error)
	sumApprovedAmountFn       func(ctx context.Context, orgID string) (float64, error)
	countByDayFn              func(ctx context.Context, orgID string, days int) ([]model.DayCount, error)
}

func (m *mockEntryRepository) ListByOrg(ctx context.Context, orgID, status, q, dateFrom, dateTo, sortBy, sortDir string, limit, offset int) ([]*model.AccountingEntry, int, error) {
	if m.listByOrgFn != nil {
		return m.listByOrgFn(ctx, orgID, status, q, dateFrom, dateTo, sortBy, sortDir, limit, offset)
	}
	return []*model.AccountingEntry{}, 0, nil
}

func (m *mockEntryRepository) UpdateEntry(ctx context.Context, id, orgID string, fields model.UpdateEntryFields) error {
	if m.updateEntryFn != nil {
		return m.updateEntryFn(ctx, id, orgID, fields)
	}
	return nil
}

func (m *mockEntryRepository) UpdateStatus(ctx context.Context, id, orgID string, status model.EntryStatus) error {
	if m.updateStatusFn != nil {
		return m.updateStatusFn(ctx, id, orgID, status)
	}
	return nil
}

func (m *mockEntryRepository) UpdateStatusWithReason(ctx context.Context, id, orgID string, status model.EntryStatus, reason string) error {
	if m.updateStatusWithReasonFn != nil {
		return m.updateStatusWithReasonFn(ctx, id, orgID, status, reason)
	}
	return nil
}

func (m *mockEntryRepository) BulkUpdateStatus(ctx context.Context, ids []string, orgID string, status model.EntryStatus) (int, error) {
	if m.bulkUpdateStatusFn != nil {
		return m.bulkUpdateStatusFn(ctx, ids, orgID, status)
	}
	return len(ids), nil
}

func (m *mockEntryRepository) GetByID(ctx context.Context, id, orgID string) (*model.AccountingEntry, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id, orgID)
	}
	return nil, nil
}

func (m *mockEntryRepository) CountByStatus(ctx context.Context, orgID string) (map[string]int, error) {
	if m.countByStatusFn != nil {
		return m.countByStatusFn(ctx, orgID)
	}
	return map[string]int{}, nil
}

func (m *mockEntryRepository) SumApprovedAmount(ctx context.Context, orgID string) (float64, error) {
	if m.sumApprovedAmountFn != nil {
		return m.sumApprovedAmountFn(ctx, orgID)
	}
	return 0, nil
}

func (m *mockEntryRepository) CountByDay(ctx context.Context, orgID string, days int) ([]model.DayCount, error) {
	if m.countByDayFn != nil {
		return m.countByDayFn(ctx, orgID, days)
	}
	return []model.DayCount{}, nil
}

// buildEntryAuthContext injects orgID into request context.
func buildEntryAuthContext(r *http.Request, orgID string) *http.Request {
	ctx := context.WithValue(r.Context(), middleware.OrgIDKey, orgID)
	return r.WithContext(ctx)
}

func sampleEntry(id string, status model.EntryStatus) *model.AccountingEntry {
	conf := 0.92
	docID := "doc-1"
	return &model.AccountingEntry{
		ID:             id,
		OrganizationID: "org-1",
		DocumentID:     &docID,
		EntryDate:      time.Now(),
		Description:    "Mua hàng hóa nhập kho",
		DebitAccount:   "156",
		CreditAccount:  "331",
		Amount:         1000000,
		Status:         status,
		AIConfidence:   &conf,
		CreatedAt:      time.Now(),
	}
}

func TestEntryHandler_List_ReturnsPendingEntries(t *testing.T) {
	entries := []*model.AccountingEntry{
		sampleEntry("entry-1", model.EntryPending),
		sampleEntry("entry-2", model.EntryPending),
	}
	repo := &mockEntryRepository{
		listByOrgFn: func(_ context.Context, orgID, status, _, _, _, _, _ string, _, _ int) ([]*model.AccountingEntry, int, error) {
			return entries, len(entries), nil
		},
	}
	h := NewEntryHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/entries?status=pending", nil)
	req = buildEntryAuthContext(req, "org-1")
	rr := httptest.NewRecorder()

	h.List(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp model.EntryListResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(resp.Entries))
	}
	if resp.Total != 2 {
		t.Errorf("expected total 2, got %d", resp.Total)
	}
}

func TestEntryHandler_List_EmptyReturnsSliceNotNull(t *testing.T) {
	h := NewEntryHandler(&mockEntryRepository{})

	req := httptest.NewRequest(http.MethodGet, "/api/entries", nil)
	req = buildEntryAuthContext(req, "org-1")
	rr := httptest.NewRecorder()

	h.List(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var resp model.EntryListResponse
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp.Entries == nil {
		t.Error("entries should not be null")
	}
}

func TestEntryHandler_List_Unauthorized(t *testing.T) {
	h := NewEntryHandler(&mockEntryRepository{})

	req := httptest.NewRequest(http.MethodGet, "/api/entries", nil)
	// No auth context
	rr := httptest.NewRecorder()

	h.List(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestEntryHandler_Approve_UpdatesStatusToApproved(t *testing.T) {
	entry := sampleEntry("entry-1", model.EntryPending)
	var calledWith model.EntryStatus
	repo := &mockEntryRepository{
		updateStatusFn: func(_ context.Context, id, orgID string, status model.EntryStatus) error {
			calledWith = status
			return nil
		},
		getByIDFn: func(_ context.Context, id, orgID string) (*model.AccountingEntry, error) {
			entry.Status = calledWith
			return entry, nil
		},
	}
	h := NewEntryHandler(repo)

	req := httptest.NewRequest(http.MethodPost, "/api/entries/entry-1/approve", nil)
	req = buildEntryAuthContext(req, "org-1")

	// Inject chi URL params
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "entry-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	h.Approve(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	if calledWith != model.EntryApproved {
		t.Errorf("expected UpdateStatus called with approved, got %s", calledWith)
	}
	var resp model.EntryActionResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Entry == nil {
		t.Error("expected entry in response")
	}
}

func TestEntryHandler_Reject_UpdatesStatusToRejected(t *testing.T) {
	entry := sampleEntry("entry-1", model.EntryPending)
	var calledWith model.EntryStatus
	repo := &mockEntryRepository{
		updateStatusFn: func(_ context.Context, id, orgID string, status model.EntryStatus) error {
			calledWith = status
			return nil
		},
		getByIDFn: func(_ context.Context, id, orgID string) (*model.AccountingEntry, error) {
			entry.Status = calledWith
			return entry, nil
		},
	}
	h := NewEntryHandler(repo)

	req := httptest.NewRequest(http.MethodPost, "/api/entries/entry-1/reject", nil)
	req = buildEntryAuthContext(req, "org-1")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "entry-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	h.Reject(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	if calledWith != model.EntryRejected {
		t.Errorf("expected UpdateStatus called with rejected, got %s", calledWith)
	}
}

func TestEntryHandler_Approve_NotFound(t *testing.T) {
	repo := &mockEntryRepository{
		updateStatusFn: func(_ context.Context, _, _ string, _ model.EntryStatus) error {
			return fmt.Errorf("not found")
		},
	}
	h := NewEntryHandler(repo)

	req := httptest.NewRequest(http.MethodPost, "/api/entries/bad-id/approve", nil)
	req = buildEntryAuthContext(req, "org-1")

	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "bad-id")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rr := httptest.NewRecorder()
	h.Approve(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestEntryHandler_Approve_Unauthorized(t *testing.T) {
	h := NewEntryHandler(&mockEntryRepository{})

	req := httptest.NewRequest(http.MethodPost, "/api/entries/entry-1/approve", nil)
	// No auth context
	rr := httptest.NewRecorder()
	h.Approve(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestEntryHandler_Sync_UpdatesStatusToSynced(t *testing.T) {
	entry := sampleEntry("entry-1", model.EntryApproved)
	var calledWith model.EntryStatus
	repo := &mockEntryRepository{
		getByIDFn: func(_ context.Context, id, orgID string) (*model.AccountingEntry, error) {
			return entry, nil
		},
		updateStatusFn: func(_ context.Context, id, orgID string, status model.EntryStatus) error {
			calledWith = status
			return nil
		},
	}
	h := NewEntryHandler(repo)

	req := httptest.NewRequest(http.MethodPost, "/api/entries/entry-1/sync", nil)
	req = buildEntryAuthContext(req, "org-1")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "entry-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rr := httptest.NewRecorder()

	h.Sync(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	if calledWith != model.EntrySynced {
		t.Errorf("expected UpdateStatus called with synced, got %s", calledWith)
	}
}

func TestEntryHandler_Sync_RejectsNonApprovedEntry(t *testing.T) {
	entry := sampleEntry("entry-1", model.EntryPending) // not approved
	repo := &mockEntryRepository{
		getByIDFn: func(_ context.Context, _, _ string) (*model.AccountingEntry, error) {
			return entry, nil
		},
	}
	h := NewEntryHandler(repo)

	req := httptest.NewRequest(http.MethodPost, "/api/entries/entry-1/sync", nil)
	req = buildEntryAuthContext(req, "org-1")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "entry-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rr := httptest.NewRecorder()

	h.Sync(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for non-approved entry, got %d", rr.Code)
	}
}

func TestEntryHandler_Sync_Unauthorized(t *testing.T) {
	h := NewEntryHandler(&mockEntryRepository{})

	req := httptest.NewRequest(http.MethodPost, "/api/entries/entry-1/sync", nil)
	rr := httptest.NewRecorder()
	h.Sync(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestEntryHandler_Get_ReturnsEntry(t *testing.T) {
	entry := sampleEntry("entry-1", model.EntryPending)
	repo := &mockEntryRepository{
		getByIDFn: func(_ context.Context, id, orgID string) (*model.AccountingEntry, error) {
			if id == "entry-1" && orgID == "org-1" {
				return entry, nil
			}
			return nil, fmt.Errorf("not found")
		},
	}
	h := NewEntryHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/entries/entry-1", nil)
	req = buildEntryAuthContext(req, "org-1")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "entry-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rr := httptest.NewRecorder()

	h.Get(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp model.EntryActionResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Entry == nil {
		t.Fatal("expected entry in response")
	}
	if resp.Entry.ID != "entry-1" {
		t.Errorf("expected id 'entry-1', got %s", resp.Entry.ID)
	}
}

func TestEntryHandler_Get_NotFound(t *testing.T) {
	repo := &mockEntryRepository{
		getByIDFn: func(_ context.Context, _, _ string) (*model.AccountingEntry, error) {
			return nil, fmt.Errorf("not found")
		},
	}
	h := NewEntryHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/entries/bad-id", nil)
	req = buildEntryAuthContext(req, "org-1")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "bad-id")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rr := httptest.NewRecorder()

	h.Get(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestEntryHandler_List_CapsLimitAt100(t *testing.T) {
	var capturedLimit int
	repo := &mockEntryRepository{
		listByOrgFn: func(_ context.Context, _, _, _, _, _, _, _ string, limit, _ int) ([]*model.AccountingEntry, int, error) {
			capturedLimit = limit
			return []*model.AccountingEntry{}, 0, nil
		},
	}
	h := NewEntryHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/entries?limit=999", nil)
	req = buildEntryAuthContext(req, "org-1")
	rr := httptest.NewRecorder()

	h.List(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if capturedLimit > 100 {
		t.Errorf("expected limit capped at 100, got %d", capturedLimit)
	}
}

func TestEntryHandler_Get_Unauthorized(t *testing.T) {
	h := NewEntryHandler(&mockEntryRepository{})

	req := httptest.NewRequest(http.MethodGet, "/api/entries/entry-1", nil)
	rr := httptest.NewRecorder()
	h.Get(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestEntryHandler_List_PassesSearchQuery(t *testing.T) {
	var capturedQuery string
	repo := &mockEntryRepository{
		listByOrgFn: func(_ context.Context, orgID, status, q, _, _, _, _ string, limit, offset int) ([]*model.AccountingEntry, int, error) {
			capturedQuery = q
			return []*model.AccountingEntry{}, 0, nil
		},
	}
	h := NewEntryHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/entries?q=mua+h%C3%A0ng", nil)
	req = buildEntryAuthContext(req, "org-1")
	rr := httptest.NewRecorder()
	h.List(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if capturedQuery != "mua hàng" {
		t.Errorf("expected query 'mua hàng', got %q", capturedQuery)
	}
}

func TestEntryHandler_List_PassesDateRange(t *testing.T) {
	var capturedDateFrom, capturedDateTo string
	repo := &mockEntryRepository{
		listByOrgFn: func(_ context.Context, orgID, status, q, dateFrom, dateTo, _, _ string, limit, offset int) ([]*model.AccountingEntry, int, error) {
			capturedDateFrom = dateFrom
			capturedDateTo = dateTo
			return []*model.AccountingEntry{}, 0, nil
		},
	}
	h := NewEntryHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/entries?date_from=2026-03-01&date_to=2026-03-31", nil)
	req = buildEntryAuthContext(req, "org-1")
	rr := httptest.NewRecorder()
	h.List(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if capturedDateFrom != "2026-03-01" {
		t.Errorf("expected dateFrom '2026-03-01', got %q", capturedDateFrom)
	}
	if capturedDateTo != "2026-03-31" {
		t.Errorf("expected dateTo '2026-03-31', got %q", capturedDateTo)
	}
}

func TestEntryHandler_Stats_ReturnsCountsByStatus(t *testing.T) {
	repo := &mockEntryRepository{
		countByStatusFn: func(_ context.Context, orgID string) (map[string]int, error) {
			return map[string]int{
				"pending":  5,
				"approved": 3,
				"rejected": 1,
				"synced":   2,
				"draft":    0,
			}, nil
		},
	}
	h := NewEntryHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/entries/stats", nil)
	req = buildEntryAuthContext(req, "org-1")
	rr := httptest.NewRecorder()

	h.Stats(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	byStatus, ok := resp["by_status"].(map[string]interface{})
	if !ok {
		t.Fatal("expected by_status field")
	}
	if byStatus["pending"] != float64(5) {
		t.Errorf("expected pending=5, got %v", byStatus["pending"])
	}
	if total, ok := resp["total"].(float64); !ok || total != 11 {
		t.Errorf("expected total=11, got %v", resp["total"])
	}
}

func TestEntryHandler_Stats_IncludesTotalApprovedAmount(t *testing.T) {
	repo := &mockEntryRepository{
		countByStatusFn: func(_ context.Context, _ string) (map[string]int, error) {
			return map[string]int{"approved": 2}, nil
		},
		sumApprovedAmountFn: func(_ context.Context, orgID string) (float64, error) {
			return 5500000.0, nil
		},
	}
	h := NewEntryHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/entries/stats", nil)
	req = buildEntryAuthContext(req, "org-1")
	rr := httptest.NewRecorder()

	h.Stats(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	amt, ok := resp["total_approved_amount"].(float64)
	if !ok || amt != 5500000.0 {
		t.Errorf("expected total_approved_amount=5500000, got %v", resp["total_approved_amount"])
	}
}

func TestEntryHandler_Stats_Unauthorized(t *testing.T) {
	h := NewEntryHandler(&mockEntryRepository{})

	req := httptest.NewRequest(http.MethodGet, "/api/entries/stats", nil)
	rr := httptest.NewRecorder()

	h.Stats(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestEntryHandler_Update_Success(t *testing.T) {
	updated := sampleEntry("entry-1", model.EntryPending)
	updated.Description = "Mua NVL nhập kho"
	updated.DebitAccount = "152"
	updated.CreditAccount = "331"
	updated.Amount = 2000000

	repo := &mockEntryRepository{
		getByIDFn: func(_ context.Context, id, orgID string) (*model.AccountingEntry, error) {
			if id == "entry-1" {
				return updated, nil
			}
			return nil, fmt.Errorf("not found")
		},
		updateEntryFn: func(_ context.Context, id, orgID string, fields model.UpdateEntryFields) error {
			if id != "entry-1" {
				return fmt.Errorf("not found")
			}
			return nil
		},
	}
	h := NewEntryHandler(repo)

	body := `{"description":"Mua NVL nhập kho","debit_account":"152","credit_account":"331","amount":2000000}`
	req := httptest.NewRequest(http.MethodPatch, "/api/entries/entry-1", strings.NewReader(body))
	req = buildEntryAuthContext(req, "org-1")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "entry-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rr := httptest.NewRecorder()

	h.Update(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp model.EntryActionResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Entry.Description != "Mua NVL nhập kho" {
		t.Errorf("expected updated description, got %q", resp.Entry.Description)
	}
}

func TestEntryHandler_Update_OnlyPendingAllowed(t *testing.T) {
	repo := &mockEntryRepository{
		getByIDFn: func(_ context.Context, id, _ string) (*model.AccountingEntry, error) {
			return sampleEntry(id, model.EntryApproved), nil
		},
	}
	h := NewEntryHandler(repo)

	body := `{"description":"new","debit_account":"152","credit_account":"331","amount":1000}`
	req := httptest.NewRequest(http.MethodPatch, "/api/entries/entry-1", strings.NewReader(body))
	req = buildEntryAuthContext(req, "org-1")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "entry-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rr := httptest.NewRecorder()

	h.Update(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for non-pending entry, got %d", rr.Code)
	}
}

func TestEntryHandler_Update_Unauthorized(t *testing.T) {
	h := NewEntryHandler(&mockEntryRepository{})

	req := httptest.NewRequest(http.MethodPatch, "/api/entries/entry-1", strings.NewReader("{}"))
	rr := httptest.NewRecorder()

	h.Update(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestEntryHandler_Update_EntryNotFound(t *testing.T) {
	repo := &mockEntryRepository{
		getByIDFn: func(_ context.Context, _, _ string) (*model.AccountingEntry, error) {
			return nil, fmt.Errorf("not found")
		},
	}
	h := NewEntryHandler(repo)

	body := `{"description":"x","debit_account":"111","credit_account":"112","amount":100}`
	req := httptest.NewRequest(http.MethodPatch, "/api/entries/bad-id", strings.NewReader(body))
	req = buildEntryAuthContext(req, "org-1")
	rctx2 := chi.NewRouteContext()
	rctx2.URLParams.Add("id", "bad-id")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx2))
	rr := httptest.NewRecorder()

	h.Update(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestEntryHandler_BulkApprove_Success(t *testing.T) {
	var capturedIDs []string
	repo := &mockEntryRepository{
		bulkUpdateStatusFn: func(_ context.Context, ids []string, _ string, status model.EntryStatus) (int, error) {
			capturedIDs = ids
			if status != model.EntryApproved {
				t.Errorf("expected approved status, got %s", status)
			}
			return len(ids), nil
		},
	}
	h := NewEntryHandler(repo)

	body := `{"ids":["entry-1","entry-2","entry-3"]}`
	req := httptest.NewRequest(http.MethodPost, "/api/entries/bulk-approve", strings.NewReader(body))
	req = buildEntryAuthContext(req, "org-1")
	rr := httptest.NewRecorder()

	h.BulkApprove(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	if len(capturedIDs) != 3 {
		t.Errorf("expected 3 ids, got %d", len(capturedIDs))
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["updated"] != float64(3) {
		t.Errorf("expected updated=3, got %v", resp["updated"])
	}
}

func TestEntryHandler_BulkApprove_EmptyIDs(t *testing.T) {
	h := NewEntryHandler(&mockEntryRepository{})

	body := `{"ids":[]}`
	req := httptest.NewRequest(http.MethodPost, "/api/entries/bulk-approve", strings.NewReader(body))
	req = buildEntryAuthContext(req, "org-1")
	rr := httptest.NewRecorder()

	h.BulkApprove(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for empty ids, got %d", rr.Code)
	}
}

func TestEntryHandler_BulkApprove_Unauthorized(t *testing.T) {
	h := NewEntryHandler(&mockEntryRepository{})

	req := httptest.NewRequest(http.MethodPost, "/api/entries/bulk-approve", strings.NewReader(`{"ids":["x"]}`))
	rr := httptest.NewRecorder()

	h.BulkApprove(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestEntryHandler_Update_RejectsInvalidVASAccount(t *testing.T) {
	pendingEntry := &model.AccountingEntry{
		ID:     "entry-1",
		Status: model.EntryPending,
	}
	repo := &mockEntryRepository{
		getByIDFn: func(_ context.Context, id, _ string) (*model.AccountingEntry, error) {
			return pendingEntry, nil
		},
	}
	h := NewEntryHandler(repo)

	tests := []struct {
		name          string
		debitAccount  string
		creditAccount string
	}{
		{"non-numeric debit", "abc", "331"},
		{"non-numeric credit", "156", "xyz"},
		{"too short debit", "15", "331"},
		{"too long debit", "15678", "331"},
		{"too short credit", "156", "33"},
		{"too long credit", "156", "33112"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			body := fmt.Sprintf(
				`{"description":"test","debit_account":%q,"credit_account":%q,"amount":1000}`,
				tc.debitAccount, tc.creditAccount,
			)
			req := httptest.NewRequest(http.MethodPatch, "/api/entries/entry-1", strings.NewReader(body))
			req = buildEntryAuthContext(req, "org-1")
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("id", "entry-1")
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
			rr := httptest.NewRecorder()

			h.Update(rr, req)

			if rr.Code != http.StatusBadRequest {
				t.Errorf("[%s] expected 400, got %d: %s", tc.name, rr.Code, rr.Body.String())
			}
		})
	}
}

func TestEntryHandler_Update_AcceptsValidVASAccount(t *testing.T) {
	pendingEntry := &model.AccountingEntry{
		ID:     "entry-1",
		Status: model.EntryPending,
	}
	updated := &model.AccountingEntry{
		ID:            "entry-1",
		Status:        model.EntryPending,
		Description:   "valid",
		DebitAccount:  "156",
		CreditAccount: "3311",
		Amount:        5000,
	}
	repo := &mockEntryRepository{
		getByIDFn: func(_ context.Context, id, _ string) (*model.AccountingEntry, error) {
			if id == "entry-1" {
				return updated, nil
			}
			return pendingEntry, nil
		},
	}
	h := NewEntryHandler(repo)

	body := `{"description":"valid","debit_account":"156","credit_account":"3311","amount":5000}`
	req := httptest.NewRequest(http.MethodPatch, "/api/entries/entry-1", strings.NewReader(body))
	req = buildEntryAuthContext(req, "org-1")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "entry-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rr := httptest.NewRecorder()

	h.Update(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestEntryHandler_Reject_WithReason_CallsUpdateStatusWithReason(t *testing.T) {
	entry := sampleEntry("entry-1", model.EntryPending)
	var capturedReason string
	repo := &mockEntryRepository{
		updateStatusWithReasonFn: func(_ context.Context, id, orgID string, status model.EntryStatus, reason string) error {
			capturedReason = reason
			entry.Status = model.EntryRejected
			r := reason
			entry.RejectReason = &r
			return nil
		},
		getByIDFn: func(_ context.Context, id, orgID string) (*model.AccountingEntry, error) {
			return entry, nil
		},
	}
	h := NewEntryHandler(repo)

	body := `{"reason":"Sai tài khoản kế toán"}`
	req := httptest.NewRequest(http.MethodPost, "/api/entries/entry-1/reject", strings.NewReader(body))
	req = buildEntryAuthContext(req, "org-1")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "entry-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rr := httptest.NewRecorder()

	h.Reject(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	if capturedReason != "Sai tài khoản kế toán" {
		t.Errorf("expected reason 'Sai tài khoản kế toán', got %q", capturedReason)
	}

	var resp model.EntryActionResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.Entry.RejectReason == nil || *resp.Entry.RejectReason != "Sai tài khoản kế toán" {
		t.Errorf("expected reject_reason in response entry")
	}
}

func TestEntryHandler_Reject_NoReason_CallsUpdateStatus(t *testing.T) {
	entry := sampleEntry("entry-1", model.EntryPending)
	var updateStatusCalled bool
	var updateStatusWithReasonCalled bool
	repo := &mockEntryRepository{
		updateStatusFn: func(_ context.Context, _, _ string, status model.EntryStatus) error {
			updateStatusCalled = true
			entry.Status = status
			return nil
		},
		updateStatusWithReasonFn: func(_ context.Context, _, _ string, _ model.EntryStatus, _ string) error {
			updateStatusWithReasonCalled = true
			return nil
		},
		getByIDFn: func(_ context.Context, id, orgID string) (*model.AccountingEntry, error) {
			return entry, nil
		},
	}
	h := NewEntryHandler(repo)

	req := httptest.NewRequest(http.MethodPost, "/api/entries/entry-1/reject", nil)
	req = buildEntryAuthContext(req, "org-1")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "entry-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rr := httptest.NewRecorder()

	h.Reject(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	if !updateStatusCalled {
		t.Error("expected UpdateStatus to be called when no reason provided")
	}
	if updateStatusWithReasonCalled {
		t.Error("expected UpdateStatusWithReason NOT to be called when no reason provided")
	}
}

func TestEntryHandler_List_PassesSortParams(t *testing.T) {
	var capturedSortBy, capturedSortDir string
	repo := &mockEntryRepository{
		listByOrgFn: func(_ context.Context, orgID, status, q, dateFrom, dateTo, sortBy, sortDir string, limit, offset int) ([]*model.AccountingEntry, int, error) {
			capturedSortBy = sortBy
			capturedSortDir = sortDir
			return []*model.AccountingEntry{}, 0, nil
		},
	}
	h := NewEntryHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/entries?sort_by=entry_date&sort_dir=asc", nil)
	req = buildEntryAuthContext(req, "org-1")
	rr := httptest.NewRecorder()
	h.List(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if capturedSortBy != "entry_date" {
		t.Errorf("expected sortBy 'entry_date', got %q", capturedSortBy)
	}
	if capturedSortDir != "asc" {
		t.Errorf("expected sortDir 'asc', got %q", capturedSortDir)
	}
}

func TestEntryHandler_Stats_IncludesDailyTrend(t *testing.T) {
	// Stats handler should return a daily_trend field with DayCount data
	repo := &mockEntryRepository{
		countByStatusFn: func(_ context.Context, _ string) (map[string]int, error) {
			return map[string]int{"pending": 2}, nil
		},
		sumApprovedAmountFn: func(_ context.Context, _ string) (float64, error) {
			return 0, nil
		},
		countByDayFn: func(_ context.Context, orgID string, days int) ([]model.DayCount, error) {
			return []model.DayCount{
				{Date: "2026-03-28", Count: 5},
				{Date: "2026-03-29", Count: 3},
			}, nil
		},
	}
	h := NewEntryHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/entries/stats", nil)
	req = buildEntryAuthContext(req, "org-1")
	rr := httptest.NewRecorder()

	h.Stats(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	trend, ok := resp["daily_trend"]
	if !ok {
		t.Fatal("expected daily_trend field in stats response")
	}
	trendArr, ok := trend.([]interface{})
	if !ok || len(trendArr) != 2 {
		t.Errorf("expected daily_trend to be array of 2, got %v", trend)
	}
}
