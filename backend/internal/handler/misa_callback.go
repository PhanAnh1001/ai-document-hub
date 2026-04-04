package handler

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"

	"github.com/PhanAnh1001/ai-accounting/backend/internal/model"
)

// CallbackDataType mirrors the MISA spec enum values.
const (
	callbackSaveVoucher   = 1
	callbackDeleteVoucher = 2
)

// CallbackEntryRepository is the minimal interface the callback handler needs.
type CallbackEntryRepository interface {
	UpdateStatusByID(ctx context.Context, id string, status model.EntryStatus) error
}

// MISACallbackHandler handles the POST /api/misa/callback endpoint.
// MISA AMIS calls this endpoint asynchronously after processing save/delete voucher requests.
type MISACallbackHandler struct {
	repo     CallbackEntryRepository
	appID    string // MISA-provided app_id used as HMAC key for signature validation
	syncMode bool   // when true, processCallback runs synchronously (for testing)
}

// NewMISACallbackHandler creates a new handler.
// When appID is empty, signature validation is skipped (useful for local dev/testing).
func NewMISACallbackHandler(repo CallbackEntryRepository, appID string, opts ...MISACallbackHandlerOption) *MISACallbackHandler {
	h := &MISACallbackHandler{repo: repo, appID: appID}
	for _, o := range opts {
		o(h)
	}
	return h
}

// MISACallbackHandlerOption configures a MISACallbackHandler.
type MISACallbackHandlerOption func(*MISACallbackHandler)

// WithSyncProcessing makes the handler process callbacks synchronously.
// Only use this in tests — production code should use async processing.
func WithSyncProcessing() MISACallbackHandlerOption {
	return func(h *MISACallbackHandler) { h.syncMode = true }
}

// HandleCallback receives the MISA AMIS callback POST request.
// It always returns HTTP 200 with a JSON body — even on errors — as required by the MISA spec.
func (h *MISACallbackHandler) HandleCallback(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var input model.MISACallbackInput
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		writeCallbackError(w, "InvalidParam", "invalid request body")
		return
	}

	// Validate HMAC-SHA256 signature when appID is configured.
	// Signature = SHA256HMAC(data_string, app_id)
	if h.appID != "" {
		expected := generateHMACSHA256(input.Data, h.appID)
		if !hmac.Equal([]byte(expected), []byte(input.Signature)) {
			writeCallbackError(w, "InvalidParam", "Signature invalid")
			return
		}
	}

	// Process asynchronously so MISA gets a fast response — matching the pattern in MISA demo code.
	// syncMode is only set in unit tests to avoid goroutine race conditions.
	if h.syncMode {
		h.processCallback(input)
	} else {
		go h.processCallback(input)
	}

	json.NewEncoder(w).Encode(model.MISACallbackOutput{Success: true})
}

// processCallback handles the actual state transitions after the signature check.
func (h *MISACallbackHandler) processCallback(input model.MISACallbackInput) {
	switch input.DataType {
	case callbackSaveVoucher, callbackDeleteVoucher:
		var items []model.MISACallbackDetail
		if err := json.Unmarshal([]byte(input.Data), &items); err != nil {
			log.Printf("misa callback: failed to unmarshal data (type=%d): %v", input.DataType, err)
			return
		}
		for _, item := range items {
			h.applyCallbackItem(input.DataType, item)
		}
	default:
		// Ignore other data types (UpdateDocumentRef=4, UpdateTaxInfoASP=5 are MISA-internal)
		log.Printf("misa callback: ignored data_type=%d", input.DataType)
	}
}

// applyCallbackItem updates the entry status based on a single callback detail item.
func (h *MISACallbackHandler) applyCallbackItem(dataType int, item model.MISACallbackDetail) {
	if item.OrgRefID == "" {
		log.Printf("misa callback: skipping item with empty org_refid")
		return
	}

	ctx := context.Background()
	var newStatus model.EntryStatus

	switch dataType {
	case callbackSaveVoucher:
		if item.Success {
			newStatus = model.EntrySynced
		} else {
			newStatus = model.EntryErrorSync
			log.Printf("misa callback: save voucher failed for entry %s — code=%v msg=%s",
				item.OrgRefID, item.ErrorCode, item.ErrorMessage)
		}
	case callbackDeleteVoucher:
		if item.Success {
			// Voucher deleted in MISA → revert to approved (not yet synced)
			newStatus = model.EntryApproved
		} else {
			newStatus = model.EntryErrorSync
			log.Printf("misa callback: delete voucher failed for entry %s — code=%v msg=%s",
				item.OrgRefID, item.ErrorCode, item.ErrorMessage)
		}
	}

	if err := h.repo.UpdateStatusByID(ctx, item.OrgRefID, newStatus); err != nil {
		log.Printf("misa callback: failed to update entry %s to %s: %v", item.OrgRefID, newStatus, err)
	}
}

// generateHMACSHA256 computes HMAC-SHA256(input, key) and returns the lowercase hex string.
// This matches the GeneratorSHA256HMAC function in the MISA demo code.
func generateHMACSHA256(input, key string) string {
	if input == "" {
		input = ""
	}
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(input))
	return hex.EncodeToString(mac.Sum(nil))
}

// writeCallbackError writes a failure JSON response (still HTTP 200 per MISA spec).
func writeCallbackError(w http.ResponseWriter, code, message string) {
	json.NewEncoder(w).Encode(model.MISACallbackOutput{
		Success:      false,
		ErrorCode:    code,
		ErrorMessage: message,
	})
}
