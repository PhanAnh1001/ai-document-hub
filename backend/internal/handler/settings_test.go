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

	"github.com/PhanAnh1001/ai-accounting/backend/internal/middleware"
	"github.com/PhanAnh1001/ai-accounting/backend/internal/model"
)

type mockSettingsOrgRepository struct {
	getByIDFn func(ctx context.Context, id string) (*model.Organization, error)
	updateFn  func(ctx context.Context, org *model.Organization) error
}

func (m *mockSettingsOrgRepository) GetByID(ctx context.Context, id string) (*model.Organization, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockSettingsOrgRepository) Update(ctx context.Context, org *model.Organization) error {
	if m.updateFn != nil {
		return m.updateFn(ctx, org)
	}
	return nil
}

func sampleOrg() *model.Organization {
	taxCode := "0123456789"
	return &model.Organization{
		ID:                 "org-1",
		Name:               "Công ty ABC",
		TaxCode:            &taxCode,
		AccountingStandard: model.TT200,
		CreatedAt:          time.Now(),
		UpdatedAt:          time.Now(),
	}
}

func buildSettingsAuthContext(r *http.Request, orgID string) *http.Request {
	ctx := context.WithValue(r.Context(), middleware.OrgIDKey, orgID)
	return r.WithContext(ctx)
}

func TestSettingsHandler_Get_ReturnsOrganization(t *testing.T) {
	org := sampleOrg()
	repo := &mockSettingsOrgRepository{
		getByIDFn: func(_ context.Context, id string) (*model.Organization, error) {
			return org, nil
		},
	}
	h := NewSettingsHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/settings", nil)
	req = buildSettingsAuthContext(req, "org-1")
	rr := httptest.NewRecorder()

	h.Get(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	orgResp, ok := resp["organization"].(map[string]interface{})
	if !ok {
		t.Fatal("expected organization field in response")
	}
	if orgResp["name"] != "Công ty ABC" {
		t.Errorf("expected name 'Công ty ABC', got %v", orgResp["name"])
	}
}

func TestSettingsHandler_Get_Unauthorized(t *testing.T) {
	h := NewSettingsHandler(&mockSettingsOrgRepository{})

	req := httptest.NewRequest(http.MethodGet, "/api/settings", nil)
	rr := httptest.NewRecorder()

	h.Get(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestSettingsHandler_Get_NotFound(t *testing.T) {
	repo := &mockSettingsOrgRepository{
		getByIDFn: func(_ context.Context, _ string) (*model.Organization, error) {
			return nil, fmt.Errorf("not found")
		},
	}
	h := NewSettingsHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/settings", nil)
	req = buildSettingsAuthContext(req, "org-1")
	rr := httptest.NewRecorder()

	h.Get(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestSettingsHandler_Update_SavesFields(t *testing.T) {
	var saved *model.Organization
	org := sampleOrg()
	repo := &mockSettingsOrgRepository{
		getByIDFn: func(_ context.Context, _ string) (*model.Organization, error) {
			return org, nil
		},
		updateFn: func(_ context.Context, o *model.Organization) error {
			saved = o
			return nil
		},
	}
	h := NewSettingsHandler(repo)

	body := `{"name":"Công ty XYZ","tax_code":"9876543210","accounting_standard":"TT133"}`
	req := httptest.NewRequest(http.MethodPut, "/api/settings", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = buildSettingsAuthContext(req, "org-1")
	rr := httptest.NewRecorder()

	h.Update(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	if saved == nil {
		t.Fatal("expected Update to be called")
	}
	if saved.Name != "Công ty XYZ" {
		t.Errorf("expected name 'Công ty XYZ', got %s", saved.Name)
	}
	if saved.TaxCode == nil || *saved.TaxCode != "9876543210" {
		t.Errorf("expected tax_code '9876543210', got %v", saved.TaxCode)
	}
	if saved.AccountingStandard != model.TT133 {
		t.Errorf("expected TT133, got %s", saved.AccountingStandard)
	}
}

func TestSettingsHandler_Update_Unauthorized(t *testing.T) {
	h := NewSettingsHandler(&mockSettingsOrgRepository{})

	req := httptest.NewRequest(http.MethodPut, "/api/settings", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.Update(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestSettingsHandler_Update_OCRProvider(t *testing.T) {
	var saved *model.Organization
	org := sampleOrg()
	repo := &mockSettingsOrgRepository{
		getByIDFn: func(_ context.Context, _ string) (*model.Organization, error) {
			return org, nil
		},
		updateFn: func(_ context.Context, o *model.Organization) error {
			saved = o
			return nil
		},
	}
	h := NewSettingsHandler(repo)

	body := `{"name":"Công ty ABC","ocr_provider":"fpt"}`
	req := httptest.NewRequest(http.MethodPut, "/api/settings", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = buildSettingsAuthContext(req, "org-1")
	rr := httptest.NewRecorder()

	h.Update(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	if saved == nil {
		t.Fatal("expected Update to be called")
	}
	if saved.OCRProvider != model.OCRProviderFPT {
		t.Errorf("expected ocr_provider 'fpt', got %s", saved.OCRProvider)
	}
}

func TestSettingsHandler_Update_MISAFields(t *testing.T) {
	var saved *model.Organization
	org := sampleOrg()
	repo := &mockSettingsOrgRepository{
		getByIDFn: func(_ context.Context, _ string) (*model.Organization, error) {
			return org, nil
		},
		updateFn: func(_ context.Context, o *model.Organization) error {
			saved = o
			return nil
		},
	}
	h := NewSettingsHandler(repo)

	body := `{"name":"Công ty ABC","misa_api_url":"https://misa.example.com","misa_api_key":"secret-key-123"}`
	req := httptest.NewRequest(http.MethodPut, "/api/settings", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = buildSettingsAuthContext(req, "org-1")
	rr := httptest.NewRecorder()

	h.Update(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	if saved == nil {
		t.Fatal("expected Update to be called")
	}
	if saved.MISAAPIURL == nil || *saved.MISAAPIURL != "https://misa.example.com" {
		t.Errorf("expected misa_api_url 'https://misa.example.com', got %v", saved.MISAAPIURL)
	}
	if saved.MISAAPIKey == nil || *saved.MISAAPIKey != "secret-key-123" {
		t.Errorf("expected misa_api_key 'secret-key-123', got %v", saved.MISAAPIKey)
	}
}

func TestSettingsHandler_Update_InvalidOCRProvider_Returns400(t *testing.T) {
	org := sampleOrg()
	repo := &mockSettingsOrgRepository{
		getByIDFn: func(_ context.Context, _ string) (*model.Organization, error) {
			return org, nil
		},
	}
	h := NewSettingsHandler(repo)

	body := `{"ocr_provider":"unknown_provider"}`
	req := httptest.NewRequest(http.MethodPut, "/api/settings", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = buildSettingsAuthContext(req, "org-1")
	rr := httptest.NewRecorder()

	h.Update(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid ocr_provider, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestSettingsHandler_Update_InvalidAccountingStandard_Returns400(t *testing.T) {
	org := sampleOrg()
	repo := &mockSettingsOrgRepository{
		getByIDFn: func(_ context.Context, _ string) (*model.Organization, error) {
			return org, nil
		},
	}
	h := NewSettingsHandler(repo)

	body := `{"name":"Công ty ABC","accounting_standard":"INVALID"}`
	req := httptest.NewRequest(http.MethodPut, "/api/settings", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req = buildSettingsAuthContext(req, "org-1")
	rr := httptest.NewRecorder()

	h.Update(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for invalid accounting_standard, got %d: %s", rr.Code, rr.Body.String())
	}
}

func TestSettingsHandler_Get_ReturnsOCRProvider(t *testing.T) {
	ocrProvider := model.OCRProviderFPT
	org := sampleOrg()
	org.OCRProvider = ocrProvider
	repo := &mockSettingsOrgRepository{
		getByIDFn: func(_ context.Context, _ string) (*model.Organization, error) {
			return org, nil
		},
	}
	h := NewSettingsHandler(repo)

	req := httptest.NewRequest(http.MethodGet, "/api/settings", nil)
	req = buildSettingsAuthContext(req, "org-1")
	rr := httptest.NewRecorder()

	h.Get(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	orgResp := resp["organization"].(map[string]interface{})
	if orgResp["ocr_provider"] != "fpt" {
		t.Errorf("expected ocr_provider 'fpt', got %v", orgResp["ocr_provider"])
	}
}
