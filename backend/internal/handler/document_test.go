package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/PhanAnh1001/ai-accounting/backend/internal/middleware"
	"github.com/PhanAnh1001/ai-accounting/backend/internal/model"
)

// mockDocumentRepository implements DocumentRepository interface for testing.
type mockDocumentRepository struct {
	createFn        func(ctx context.Context, doc *model.Document) error
	listByOrgFn     func(ctx context.Context, orgID, status, q string, limit, offset int) ([]*model.Document, int, error)
	getByIDFn       func(ctx context.Context, id, orgID string) (*model.Document, error)
	countByDayFn    func(ctx context.Context, orgID string, days int, docType string) ([]model.DayCount, error)
	updateStatusFn  func(ctx context.Context, id string, status model.DocumentStatus) error
}

func (m *mockDocumentRepository) UpdateStatus(ctx context.Context, id string, status model.DocumentStatus) error {
	if m.updateStatusFn != nil {
		return m.updateStatusFn(ctx, id, status)
	}
	return nil
}

// mockDocumentProcessor implements DocumentProcessor for testing.
type mockDocumentProcessor struct {
	processFn func(ctx context.Context, doc *model.Document) error
}

func (m *mockDocumentProcessor) Process(ctx context.Context, doc *model.Document) error {
	if m.processFn != nil {
		return m.processFn(ctx, doc)
	}
	return nil
}

func (m *mockDocumentRepository) Create(ctx context.Context, doc *model.Document) error {
	if m.createFn != nil {
		return m.createFn(ctx, doc)
	}
	return nil
}

func (m *mockDocumentRepository) ListByOrg(ctx context.Context, orgID, status, q string, limit, offset int) ([]*model.Document, int, error) {
	if m.listByOrgFn != nil {
		return m.listByOrgFn(ctx, orgID, status, q, limit, offset)
	}
	return []*model.Document{}, 0, nil
}

func (m *mockDocumentRepository) GetByID(ctx context.Context, id, orgID string) (*model.Document, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id, orgID)
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockDocumentRepository) CountByDay(ctx context.Context, orgID string, days int, docType string) ([]model.DayCount, error) {
	if m.countByDayFn != nil {
		return m.countByDayFn(ctx, orgID, days, docType)
	}
	return []model.DayCount{}, nil
}

// buildAuthContext injects orgID and userID into request context.
func buildAuthContext(r *http.Request, orgID, userID string) *http.Request {
	ctx := context.WithValue(r.Context(), middleware.OrgIDKey, orgID)
	ctx = context.WithValue(ctx, middleware.UserIDKey, userID)
	return r.WithContext(ctx)
}

// buildUploadRequest creates a multipart form request with a file attachment.
func buildUploadRequest(t *testing.T, filename, content, docType string) *http.Request {
	t.Helper()
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)

	fw, err := w.CreateFormFile("file", filename)
	if err != nil {
		t.Fatal(err)
	}
	io.WriteString(fw, content)

	if docType != "" {
		w.WriteField("document_type", docType)
	}
	w.Close()

	req := httptest.NewRequest(http.MethodPost, "/api/documents", &buf)
	req.Header.Set("Content-Type", w.FormDataContentType())
	return req
}

func TestDocumentHandler_Upload_Success(t *testing.T) {
	repo := &mockDocumentRepository{}
	h := NewDocumentHandler(repo, t.TempDir())

	req := buildUploadRequest(t, "invoice.pdf", "%PDF-test", "invoice")
	req = buildAuthContext(req, "org-1", "user-1")
	rr := httptest.NewRecorder()

	h.Upload(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}
	var doc model.Document
	if err := json.NewDecoder(rr.Body).Decode(&doc); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if doc.ID == "" {
		t.Error("expected non-empty document ID")
	}
	if doc.OrganizationID != "org-1" {
		t.Errorf("expected org-1, got %s", doc.OrganizationID)
	}
	if doc.DocumentType != model.DocInvoice {
		t.Errorf("expected invoice, got %s", doc.DocumentType)
	}
	if doc.Status != model.StatusUploaded {
		t.Errorf("expected uploaded, got %s", doc.Status)
	}
}

func TestDocumentHandler_Upload_MissingFile(t *testing.T) {
	h := NewDocumentHandler(&mockDocumentRepository{}, t.TempDir())

	req := httptest.NewRequest(http.MethodPost, "/api/documents", strings.NewReader(""))
	req.Header.Set("Content-Type", "multipart/form-data; boundary=xxx")
	req = buildAuthContext(req, "org-1", "user-1")
	rr := httptest.NewRecorder()

	h.Upload(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestDocumentHandler_Upload_InvalidDocType(t *testing.T) {
	h := NewDocumentHandler(&mockDocumentRepository{}, t.TempDir())

	req := buildUploadRequest(t, "file.pdf", "content", "INVALID_TYPE")
	req = buildAuthContext(req, "org-1", "user-1")
	rr := httptest.NewRecorder()

	h.Upload(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestDocumentHandler_Upload_UnauthenticatedContext(t *testing.T) {
	h := NewDocumentHandler(&mockDocumentRepository{}, t.TempDir())

	req := buildUploadRequest(t, "file.pdf", "content", "invoice")
	// No auth context injected
	rr := httptest.NewRecorder()

	h.Upload(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestDocumentHandler_List_ReturnsOrgDocuments(t *testing.T) {
	now := time.Now()
	docs := []*model.Document{
		{ID: "doc-1", OrganizationID: "org-1", FileName: "a.pdf", DocumentType: model.DocInvoice, Status: model.StatusUploaded, CreatedAt: now},
		{ID: "doc-2", OrganizationID: "org-1", FileName: "b.pdf", DocumentType: model.DocReceipt, Status: model.StatusUploaded, CreatedAt: now},
	}
	repo := &mockDocumentRepository{
		listByOrgFn: func(_ context.Context, orgID, _, _ string, _, _ int) ([]*model.Document, int, error) {
			return docs, len(docs), nil
		},
	}
	h := NewDocumentHandler(repo, t.TempDir())

	req := httptest.NewRequest(http.MethodGet, "/api/documents", nil)
	req = buildAuthContext(req, "org-1", "user-1")
	rr := httptest.NewRecorder()

	h.List(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp model.DocumentListResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(resp.Documents) != 2 {
		t.Errorf("expected 2 documents, got %d", len(resp.Documents))
	}
	if resp.Total != 2 {
		t.Errorf("expected total 2, got %d", resp.Total)
	}
}

func TestDocumentHandler_List_EmptyList(t *testing.T) {
	h := NewDocumentHandler(&mockDocumentRepository{}, t.TempDir())

	req := httptest.NewRequest(http.MethodGet, "/api/documents", nil)
	req = buildAuthContext(req, "org-1", "user-1")
	rr := httptest.NewRecorder()

	h.List(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var resp model.DocumentListResponse
	json.NewDecoder(rr.Body).Decode(&resp)
	if resp.Documents == nil {
		t.Error("documents field should not be nil")
	}
	if resp.Total != 0 {
		t.Errorf("expected 0, got %d", resp.Total)
	}
}

func TestDocumentHandler_List_UnauthenticatedContext(t *testing.T) {
	h := NewDocumentHandler(&mockDocumentRepository{}, t.TempDir())

	req := httptest.NewRequest(http.MethodGet, "/api/documents", nil)
	rr := httptest.NewRecorder()

	h.List(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestDocumentHandler_List_PassesStatusFilter(t *testing.T) {
	var capturedStatus string
	repo := &mockDocumentRepository{
		listByOrgFn: func(_ context.Context, _, status, _ string, _, _ int) ([]*model.Document, int, error) {
			capturedStatus = status
			return []*model.Document{}, 0, nil
		},
	}
	h := NewDocumentHandler(repo, t.TempDir())

	req := httptest.NewRequest(http.MethodGet, "/api/documents?status=error", nil)
	req = buildAuthContext(req, "org-1", "user-1")
	rr := httptest.NewRecorder()

	h.List(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if capturedStatus != "error" {
		t.Errorf("expected status 'error' passed to repo, got %q", capturedStatus)
	}
}

func TestDocumentHandler_List_PassesPagination(t *testing.T) {
	var capturedLimit, capturedOffset int
	repo := &mockDocumentRepository{
		listByOrgFn: func(_ context.Context, _, _, _ string, limit, offset int) ([]*model.Document, int, error) {
			capturedLimit = limit
			capturedOffset = offset
			return []*model.Document{}, 0, nil
		},
	}
	h := NewDocumentHandler(repo, t.TempDir())

	req := httptest.NewRequest(http.MethodGet, "/api/documents?limit=10&offset=20", nil)
	req = buildAuthContext(req, "org-1", "user-1")
	rr := httptest.NewRecorder()

	h.List(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if capturedLimit != 10 {
		t.Errorf("expected limit 10, got %d", capturedLimit)
	}
	if capturedOffset != 20 {
		t.Errorf("expected offset 20, got %d", capturedOffset)
	}
}

func TestDocumentHandler_Get_ReturnsDocument(t *testing.T) {
	now := time.Now()
	doc := &model.Document{
		ID:             "doc-1",
		OrganizationID: "org-1",
		FileName:       "invoice.pdf",
		DocumentType:   model.DocInvoice,
		Status:         model.StatusBooked,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
	repo := &mockDocumentRepository{
		getByIDFn: func(_ context.Context, id, orgID string) (*model.Document, error) {
			if id == "doc-1" && orgID == "org-1" {
				return doc, nil
			}
			return nil, fmt.Errorf("not found")
		},
	}
	h := NewDocumentHandler(repo, t.TempDir())

	req := httptest.NewRequest(http.MethodGet, "/api/documents/doc-1", nil)
	req = buildAuthContext(req, "org-1", "user-1")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "doc-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rr := httptest.NewRecorder()

	h.Get(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp model.Document
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp.ID != "doc-1" {
		t.Errorf("expected id 'doc-1', got %s", resp.ID)
	}
}

func TestDocumentHandler_Get_NotFound(t *testing.T) {
	repo := &mockDocumentRepository{
		getByIDFn: func(_ context.Context, _, _ string) (*model.Document, error) {
			return nil, fmt.Errorf("not found")
		},
	}
	h := NewDocumentHandler(repo, t.TempDir())

	req := httptest.NewRequest(http.MethodGet, "/api/documents/bad-id", nil)
	req = buildAuthContext(req, "org-1", "user-1")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "bad-id")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rr := httptest.NewRecorder()

	h.Get(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestDocumentHandler_Get_Unauthorized(t *testing.T) {
	h := NewDocumentHandler(&mockDocumentRepository{}, t.TempDir())

	req := httptest.NewRequest(http.MethodGet, "/api/documents/doc-1", nil)
	// No auth context
	rr := httptest.NewRecorder()
	h.Get(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestDocumentHandler_List_CapsLimitAt100(t *testing.T) {
	var capturedLimit int
	repo := &mockDocumentRepository{
		listByOrgFn: func(_ context.Context, _, _, _ string, limit, _ int) ([]*model.Document, int, error) {
			capturedLimit = limit
			return []*model.Document{}, 0, nil
		},
	}
	h := NewDocumentHandler(repo, t.TempDir())

	req := httptest.NewRequest(http.MethodGet, "/api/documents?limit=500", nil)
	req = buildAuthContext(req, "org-1", "user-1")
	rr := httptest.NewRecorder()

	h.List(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if capturedLimit > 100 {
		t.Errorf("expected limit capped at 100, got %d", capturedLimit)
	}
}

func TestDocumentHandler_List_DefaultPagination(t *testing.T) {
	var capturedLimit, capturedOffset int
	repo := &mockDocumentRepository{
		listByOrgFn: func(_ context.Context, _, _, _ string, limit, offset int) ([]*model.Document, int, error) {
			capturedLimit = limit
			capturedOffset = offset
			return []*model.Document{}, 0, nil
		},
	}
	h := NewDocumentHandler(repo, t.TempDir())

	req := httptest.NewRequest(http.MethodGet, "/api/documents", nil)
	req = buildAuthContext(req, "org-1", "user-1")
	rr := httptest.NewRecorder()

	h.List(rr, req)

	if capturedLimit != 20 {
		t.Errorf("expected default limit 20, got %d", capturedLimit)
	}
	if capturedOffset != 0 {
		t.Errorf("expected default offset 0, got %d", capturedOffset)
	}
}

func TestDocumentHandler_List_PassesSearchQuery(t *testing.T) {
	var capturedQ string
	repo := &mockDocumentRepository{
		listByOrgFn: func(_ context.Context, _, _, q string, _, _ int) ([]*model.Document, int, error) {
			capturedQ = q
			return []*model.Document{}, 0, nil
		},
	}
	h := NewDocumentHandler(repo, t.TempDir())

	req := httptest.NewRequest(http.MethodGet, "/api/documents?q=invoice", nil)
	req = buildAuthContext(req, "org-1", "user-1")
	rr := httptest.NewRecorder()

	h.List(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if capturedQ != "invoice" {
		t.Errorf("expected q='invoice' passed to repo, got %q", capturedQ)
	}
}

func TestDocumentHandler_Stats_ReturnsDaily(t *testing.T) {
	repo := &mockDocumentRepository{
		countByDayFn: func(ctx context.Context, orgID string, days int, docType string) ([]model.DayCount, error) {
			if orgID != "org-1" {
				t.Errorf("unexpected orgID: %s", orgID)
			}
			return []model.DayCount{
				{Date: "2026-03-28", Count: 3},
				{Date: "2026-03-27", Count: 1},
			}, nil
		},
	}
	h := NewDocumentHandler(repo, t.TempDir())

	req := httptest.NewRequest(http.MethodGet, "/api/documents/stats?days=7", nil)
	req = buildAuthContext(req, "org-1", "user-1")
	rr := httptest.NewRecorder()

	h.Stats(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}

	var resp struct {
		Days []model.DayCount `json:"days"`
	}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.Days) != 2 {
		t.Errorf("expected 2 days, got %d", len(resp.Days))
	}
	if resp.Days[0].Count != 3 {
		t.Errorf("expected count 3, got %d", resp.Days[0].Count)
	}
}

func TestDocumentHandler_Stats_PassesDocTypeFilter(t *testing.T) {
	var capturedType string
	repo := &mockDocumentRepository{
		countByDayFn: func(ctx context.Context, orgID string, days int, docType string) ([]model.DayCount, error) {
			capturedType = docType
			return []model.DayCount{}, nil
		},
	}
	h := NewDocumentHandler(repo, t.TempDir())

	req := httptest.NewRequest(http.MethodGet, "/api/documents/stats?days=7&type=invoice", nil)
	req = buildAuthContext(req, "org-1", "user-1")
	rr := httptest.NewRecorder()

	h.Stats(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if capturedType != "invoice" {
		t.Errorf("expected docType 'invoice', got %q", capturedType)
	}
}

func TestDocumentHandler_Stats_Unauthorized(t *testing.T) {
	h := NewDocumentHandler(&mockDocumentRepository{}, t.TempDir())

	req := httptest.NewRequest(http.MethodGet, "/api/documents/stats", nil)
	// No auth context
	rr := httptest.NewRecorder()
	h.Stats(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestDocumentHandler_Retry_ResetsStatusAndReprocesses(t *testing.T) {
	doc := &model.Document{
		ID:             "doc-1",
		OrganizationID: "org-1",
		Status:         model.StatusError,
		FileName:       "invoice.pdf",
		DocumentType:   model.DocInvoice,
	}

	updateStatusCalled := false
	processCalled := false

	repo := &mockDocumentRepository{
		getByIDFn: func(_ context.Context, id, orgID string) (*model.Document, error) {
			if id == "doc-1" && orgID == "org-1" {
				return doc, nil
			}
			return nil, fmt.Errorf("not found")
		},
		updateStatusFn: func(_ context.Context, id string, status model.DocumentStatus) error {
			updateStatusCalled = true
			if status != model.StatusProcessing {
				t.Errorf("expected StatusProcessing, got %s", status)
			}
			return nil
		},
	}

	processor := &mockDocumentProcessor{
		processFn: func(_ context.Context, d *model.Document) error {
			processCalled = true
			return nil
		},
	}

	h := NewDocumentHandler(repo, t.TempDir(), processor)

	req := httptest.NewRequest(http.MethodPost, "/api/documents/doc-1/retry", nil)
	req = buildAuthContext(req, "org-1", "user-1")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "doc-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rr := httptest.NewRecorder()

	h.Retry(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	if !updateStatusCalled {
		t.Error("expected UpdateStatus to be called")
	}
	// processor.Process is called in a goroutine — wait briefly
	time.Sleep(20 * time.Millisecond)
	if !processCalled {
		t.Error("expected processor.Process to be called")
	}
}

func TestDocumentHandler_Retry_OnlyAllowsErrorStatus(t *testing.T) {
	doc := &model.Document{
		ID:             "doc-1",
		OrganizationID: "org-1",
		Status:         model.StatusBooked, // not error
	}
	repo := &mockDocumentRepository{
		getByIDFn: func(_ context.Context, id, _ string) (*model.Document, error) {
			return doc, nil
		},
	}
	h := NewDocumentHandler(repo, t.TempDir())

	req := httptest.NewRequest(http.MethodPost, "/api/documents/doc-1/retry", nil)
	req = buildAuthContext(req, "org-1", "user-1")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "doc-1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rr := httptest.NewRecorder()

	h.Retry(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for non-error document, got %d", rr.Code)
	}
}

func TestDocumentHandler_Retry_NotFound(t *testing.T) {
	repo := &mockDocumentRepository{
		getByIDFn: func(_ context.Context, _, _ string) (*model.Document, error) {
			return nil, fmt.Errorf("not found")
		},
	}
	h := NewDocumentHandler(repo, t.TempDir())

	req := httptest.NewRequest(http.MethodPost, "/api/documents/bad-id/retry", nil)
	req = buildAuthContext(req, "org-1", "user-1")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "bad-id")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rr := httptest.NewRecorder()

	h.Retry(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}
