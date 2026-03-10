package handler_test

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/PhanAnh1001/ai-accounting/backend/internal/handler"
	"github.com/PhanAnh1001/ai-accounting/backend/internal/model"
)

// mockCallbackEntryRepo implements handler.CallbackEntryRepository for tests.
type mockCallbackEntryRepo struct {
	updated map[string]model.EntryStatus
	err     error
}

func (m *mockCallbackEntryRepo) UpdateStatusByID(_ context.Context, id string, status model.EntryStatus) error {
	if m.err != nil {
		return m.err
	}
	if m.updated == nil {
		m.updated = map[string]model.EntryStatus{}
	}
	m.updated[id] = status
	return nil
}

// generateSignature creates the HMAC-SHA256 signature MISA uses to sign callback data.
func generateSignature(data, appID string) string {
	mac := hmac.New(sha256.New, []byte(appID))
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil))
}

func buildCallbackBody(t *testing.T, appID string, dataType int, dataItems interface{}) []byte {
	t.Helper()
	dataJSON, err := json.Marshal(dataItems)
	if err != nil {
		t.Fatalf("marshal data items: %v", err)
	}
	dataStr := string(dataJSON)
	sig := generateSignature(dataStr, appID)

	body := map[string]interface{}{
		"success":          true,
		"app_id":           appID,
		"error_code":       "",
		"error_message":    "",
		"signature":        sig,
		"data_type":        dataType,
		"org_company_code": "testcompany",
		"data":             dataStr,
	}
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal body: %v", err)
	}
	return b
}

const testAppID = "0e0a14cf-9e4b-4af9-875b-c490f34a581b"

func TestMISACallback_SaveVoucher_Success(t *testing.T) {
	repo := &mockCallbackEntryRepo{}
	h := handler.NewMISACallbackHandler(repo, testAppID, handler.WithSyncProcessing())

	dataItems := []map[string]interface{}{
		{
			"org_refid":     "entry-uuid-001",
			"success":       true,
			"error_code":    nil,
			"error_message": "",
			"session_id":    "sess-001",
			"voucher_type":  1,
		},
	}
	body := buildCallbackBody(t, testAppID, 1 /* SaveVoucher */, dataItems)

	req := httptest.NewRequest(http.MethodPost, "/api/misa/callback", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.HandleCallback(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var out model.MISACallbackOutput
	if err := json.NewDecoder(w.Body).Decode(&out); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if !out.Success {
		t.Errorf("expected Success=true, got false: %s", out.ErrorMessage)
	}
	if got := repo.updated["entry-uuid-001"]; got != model.EntrySynced {
		t.Errorf("expected entry status synced, got %q", got)
	}
}

func TestMISACallback_SaveVoucher_ItemFailed(t *testing.T) {
	repo := &mockCallbackEntryRepo{}
	h := handler.NewMISACallbackHandler(repo, testAppID, handler.WithSyncProcessing())

	errCode := "IsCreatedVoucher"
	dataItems := []map[string]interface{}{
		{
			"org_refid":     "entry-uuid-002",
			"success":       false,
			"error_code":    errCode,
			"error_message": "Đã sinh chứng từ kế toán, không thể cập nhật được.",
			"session_id":    "sess-002",
			"voucher_type":  1,
		},
	}
	body := buildCallbackBody(t, testAppID, 1, dataItems)

	req := httptest.NewRequest(http.MethodPost, "/api/misa/callback", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.HandleCallback(w, req)

	// Callback itself succeeds (we accepted the call)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var out model.MISACallbackOutput
	json.NewDecoder(w.Body).Decode(&out)
	if !out.Success {
		t.Errorf("expected Success=true even when item failed")
	}
	// Entry status should be set to error for failed items
	if got := repo.updated["entry-uuid-002"]; got != model.EntryStatus("error_sync") {
		t.Errorf("expected entry status error_sync, got %q", got)
	}
}

func TestMISACallback_InvalidSignature(t *testing.T) {
	repo := &mockCallbackEntryRepo{}
	h := handler.NewMISACallbackHandler(repo, testAppID, handler.WithSyncProcessing())

	body := map[string]interface{}{
		"success":          true,
		"app_id":           testAppID,
		"signature":        "invalidsignature",
		"data_type":        1,
		"org_company_code": "testcompany",
		"data":             `[{"org_refid":"abc","success":true}]`,
	}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/misa/callback", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.HandleCallback(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 even for invalid sig, got %d", w.Code)
	}
	var out model.MISACallbackOutput
	json.NewDecoder(w.Body).Decode(&out)
	if out.Success {
		t.Errorf("expected Success=false for invalid signature")
	}
	if out.ErrorCode != "InvalidParam" {
		t.Errorf("expected ErrorCode=InvalidParam, got %q", out.ErrorCode)
	}
}

func TestMISACallback_InvalidJSON(t *testing.T) {
	repo := &mockCallbackEntryRepo{}
	h := handler.NewMISACallbackHandler(repo, testAppID, handler.WithSyncProcessing())

	req := httptest.NewRequest(http.MethodPost, "/api/misa/callback", bytes.NewReader([]byte("not json")))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.HandleCallback(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var out model.MISACallbackOutput
	json.NewDecoder(w.Body).Decode(&out)
	if out.Success {
		t.Errorf("expected Success=false for invalid JSON")
	}
}

func TestMISACallback_DeleteVoucher(t *testing.T) {
	repo := &mockCallbackEntryRepo{}
	h := handler.NewMISACallbackHandler(repo, testAppID, handler.WithSyncProcessing())

	dataItems := []map[string]interface{}{
		{
			"org_refid":     "entry-uuid-003",
			"success":       true,
			"error_code":    nil,
			"error_message": "",
			"voucher_type":  1,
		},
	}
	body := buildCallbackBody(t, testAppID, 2 /* DeleteVoucher */, dataItems)

	req := httptest.NewRequest(http.MethodPost, "/api/misa/callback", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.HandleCallback(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var out model.MISACallbackOutput
	json.NewDecoder(w.Body).Decode(&out)
	if !out.Success {
		t.Errorf("expected Success=true")
	}
	// DeleteVoucher success → status should revert to approved (un-synced)
	if got := repo.updated["entry-uuid-003"]; got != model.EntryApproved {
		t.Errorf("expected entry status approved after delete, got %q", got)
	}
}

func TestMISACallback_EmptyAppID_SkipsSignatureCheck(t *testing.T) {
	// When MISA_APP_ID is not configured, skip signature validation (dev mode)
	repo := &mockCallbackEntryRepo{}
	h := handler.NewMISACallbackHandler(repo, "" /* no app_id */, handler.WithSyncProcessing())

	dataItems := []map[string]interface{}{
		{"org_refid": "entry-uuid-004", "success": true, "error_message": ""},
	}
	dataJSON, _ := json.Marshal(dataItems)
	body := map[string]interface{}{
		"success":   true,
		"app_id":    "some-id",
		"signature": "doesnt-matter",
		"data_type": 1,
		"data":      string(dataJSON),
	}
	b, _ := json.Marshal(body)

	req := httptest.NewRequest(http.MethodPost, "/api/misa/callback", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.HandleCallback(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var out model.MISACallbackOutput
	json.NewDecoder(w.Body).Decode(&out)
	if !out.Success {
		t.Errorf("expected Success=true when app_id not configured: %s", out.ErrorMessage)
	}
}
