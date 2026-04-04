package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/PhanAnh1001/ai-accounting/backend/internal/middleware"
	"github.com/PhanAnh1001/ai-accounting/backend/internal/model"
)

// DocumentRepository is the interface the handler depends on.
// The concrete repository.DocumentRepository satisfies this interface.
type DocumentRepository interface {
	Create(ctx context.Context, doc *model.Document) error
	ListByOrg(ctx context.Context, orgID, status, q string, limit, offset int) ([]*model.Document, int, error)
	GetByID(ctx context.Context, id, orgID string) (*model.Document, error)
	CountByDay(ctx context.Context, orgID string, days int, docType string) ([]model.DayCount, error)
	UpdateStatus(ctx context.Context, id string, status model.DocumentStatus) error
}

// DocumentProcessor is the interface for triggering async document processing.
type DocumentProcessor interface {
	Process(ctx context.Context, doc *model.Document) error
}

// DocumentHandler handles document-related HTTP requests.
type DocumentHandler struct {
	repo      DocumentRepository
	uploadDir string
	processor DocumentProcessor // optional: nil disables async processing
}

func NewDocumentHandler(repo DocumentRepository, uploadDir string, processor ...DocumentProcessor) *DocumentHandler {
	h := &DocumentHandler{repo: repo, uploadDir: uploadDir}
	if len(processor) > 0 {
		h.processor = processor[0]
	}
	return h
}

var validDocTypes = map[string]model.DocumentType{
	"invoice":       model.DocInvoice,
	"receipt":       model.DocReceipt,
	"bank_statement": model.DocBankStatement,
	"other":         model.DocOther,
}

const maxUploadSize = 10 << 20 // 10 MB

// Upload handles POST /api/documents (multipart/form-data).
func (h *DocumentHandler) Upload(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	userID := middleware.GetUserID(r.Context())
	if orgID == "" || userID == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		http.Error(w, `{"error":"file too large or invalid form"}`, http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, `{"error":"file is required"}`, http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Validate document_type
	rawType := strings.ToLower(strings.TrimSpace(r.FormValue("document_type")))
	if rawType == "" {
		rawType = "other"
	}
	docType, ok := validDocTypes[rawType]
	if !ok {
		http.Error(w, `{"error":"invalid document_type"}`, http.StatusBadRequest)
		return
	}

	// Save file to upload directory
	docID := generateID()
	safeName := sanitizeFilename(header.Filename)
	orgDir := filepath.Join(h.uploadDir, orgID)
	if err := os.MkdirAll(orgDir, 0755); err != nil {
		http.Error(w, `{"error":"failed to create storage directory"}`, http.StatusInternalServerError)
		return
	}

	destPath := filepath.Join(orgDir, docID+"-"+safeName)
	dest, err := os.Create(destPath)
	if err != nil {
		http.Error(w, `{"error":"failed to save file"}`, http.StatusInternalServerError)
		return
	}
	defer dest.Close()

	if _, err := io.Copy(dest, file); err != nil {
		http.Error(w, `{"error":"failed to write file"}`, http.StatusInternalServerError)
		return
	}

	now := time.Now()
	doc := &model.Document{
		ID:             docID,
		OrganizationID: orgID,
		FileName:       safeName,
		FileURL:        destPath,
		DocumentType:   docType,
		Status:         model.StatusUploaded,
		CreatedBy:      userID,
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := h.repo.Create(r.Context(), doc); err != nil {
		http.Error(w, `{"error":"failed to save document record"}`, http.StatusInternalServerError)
		return
	}

	// Trigger async processing (OCR → rule engine → entries).
	// Wrapped in recover to prevent goroutine panics from crashing the server.
	if h.processor != nil {
		go func() {
			defer func() {
				if r := recover(); r != nil {
					// Document stays in "uploaded" status; can be retried manually.
				}
			}()
			h.processor.Process(context.Background(), doc)
		}()
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(doc)
}

// List handles GET /api/documents?status=&limit=20&offset=0
func (h *DocumentHandler) List(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	if orgID == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	status := r.URL.Query().Get("status")
	q := r.URL.Query().Get("q")
	limit := queryInt(r, "limit", 20)
	offset := queryInt(r, "offset", 0)

	docs, total, err := h.repo.ListByOrg(r.Context(), orgID, status, q, limit, offset)
	if err != nil {
		http.Error(w, `{"error":"failed to list documents"}`, http.StatusInternalServerError)
		return
	}

	// Return empty slice, never null
	if docs == nil {
		docs = []*model.Document{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(model.DocumentListResponse{
		Documents: docs,
		Total:     total,
	})
}

// Get handles GET /api/documents/{id} — returns a single document by ID.
func (h *DocumentHandler) Get(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	if orgID == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, `{"error":"missing document id"}`, http.StatusBadRequest)
		return
	}

	doc, err := h.repo.GetByID(r.Context(), id, orgID)
	if err != nil || doc == nil {
		http.Error(w, `{"error":"document not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(doc)
}

// Stats handles GET /api/documents/stats — returns daily upload counts for the last N days.
func (h *DocumentHandler) Stats(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	if orgID == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	days := queryInt(r, "days", 7)
	if days < 1 || days > 90 {
		days = 7
	}

	docType := r.URL.Query().Get("type")
	counts, err := h.repo.CountByDay(r.Context(), orgID, days, docType)
	if err != nil {
		http.Error(w, `{"error":"failed to get stats"}`, http.StatusInternalServerError)
		return
	}

	if counts == nil {
		counts = []model.DayCount{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"days": counts,
	})
}

// Retry handles POST /api/documents/{id}/retry — re-triggers processing for a failed document.
func (h *DocumentHandler) Retry(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	if orgID == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, `{"error":"missing document id"}`, http.StatusBadRequest)
		return
	}

	doc, err := h.repo.GetByID(r.Context(), id, orgID)
	if err != nil || doc == nil {
		http.Error(w, `{"error":"document not found"}`, http.StatusNotFound)
		return
	}

	if doc.Status != model.StatusError {
		http.Error(w, `{"error":"only error documents can be retried"}`, http.StatusBadRequest)
		return
	}

	if err := h.repo.UpdateStatus(r.Context(), id, model.StatusProcessing); err != nil {
		http.Error(w, `{"error":"failed to update document status"}`, http.StatusInternalServerError)
		return
	}
	doc.Status = model.StatusProcessing

	if h.processor != nil {
		go func() {
			defer func() { recover() }()
			h.processor.Process(context.Background(), doc)
		}()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(doc)
}

// sanitizeFilename removes path separators and dangerous characters.
func sanitizeFilename(name string) string {
	name = filepath.Base(name)
	replacer := strings.NewReplacer(
		"..", "",
		"/", "_",
		"\\", "_",
		" ", "_",
	)
	result := replacer.Replace(name)
	if result == "" || result == "." {
		return fmt.Sprintf("file_%d", time.Now().UnixNano())
	}
	return result
}
