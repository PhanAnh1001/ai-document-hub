package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/PhanAnh1001/ai-accounting/backend/internal/middleware"
	"github.com/PhanAnh1001/ai-accounting/backend/internal/model"
)

// SettingsOrgRepository is the interface SettingsHandler depends on.
type SettingsOrgRepository interface {
	GetByID(ctx context.Context, id string) (*model.Organization, error)
	Update(ctx context.Context, org *model.Organization) error
}

// SettingsHandler handles organization settings HTTP requests.
type SettingsHandler struct {
	repo SettingsOrgRepository
}

func NewSettingsHandler(repo SettingsOrgRepository) *SettingsHandler {
	return &SettingsHandler{repo: repo}
}

type settingsResponse struct {
	Organization *model.Organization `json:"organization"`
}

// Get handles GET /api/settings — returns current org settings.
func (h *SettingsHandler) Get(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	if orgID == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	org, err := h.repo.GetByID(r.Context(), orgID)
	if err != nil || org == nil {
		http.Error(w, `{"error":"organization not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settingsResponse{Organization: org})
}

type updateSettingsRequest struct {
	Name               string                   `json:"name"`
	TaxCode            *string                  `json:"tax_code"`
	AccountingStandard model.AccountingStandard  `json:"accounting_standard"`
	OCRProvider        *model.OCRProvider        `json:"ocr_provider,omitempty"`
	MISAAPIURL         *string                   `json:"misa_api_url"`
	MISAAPIKey         *string                   `json:"misa_api_key"`
}

// Update handles PUT /api/settings — updates org name, tax code, accounting standard.
func (h *SettingsHandler) Update(w http.ResponseWriter, r *http.Request) {
	orgID := middleware.GetOrgID(r.Context())
	if orgID == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req updateSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	org, err := h.repo.GetByID(r.Context(), orgID)
	if err != nil || org == nil {
		http.Error(w, `{"error":"organization not found"}`, http.StatusNotFound)
		return
	}

	// Validate accounting_standard if provided
	if req.AccountingStandard != "" {
		if req.AccountingStandard != model.TT133 && req.AccountingStandard != model.TT200 {
			http.Error(w, `{"error":"invalid accounting_standard, must be TT133 or TT200"}`, http.StatusBadRequest)
			return
		}
	}
	// Validate ocr_provider if provided
	if req.OCRProvider != nil {
		if *req.OCRProvider != model.OCRProviderFPT && *req.OCRProvider != model.OCRProviderMock {
			http.Error(w, `{"error":"invalid ocr_provider, must be fpt or mock"}`, http.StatusBadRequest)
			return
		}
	}

	if req.Name != "" {
		org.Name = req.Name
	}
	if req.TaxCode != nil {
		org.TaxCode = req.TaxCode
	}
	if req.AccountingStandard != "" {
		org.AccountingStandard = req.AccountingStandard
	}
	if req.OCRProvider != nil {
		org.OCRProvider = *req.OCRProvider
	}
	if req.MISAAPIURL != nil {
		org.MISAAPIURL = req.MISAAPIURL
	}
	if req.MISAAPIKey != nil {
		org.MISAAPIKey = req.MISAAPIKey
	}

	if err := h.repo.Update(r.Context(), org); err != nil {
		http.Error(w, `{"error":"failed to update settings"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(settingsResponse{Organization: org})
}
