package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"errors"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/PhanAnh1001/ai-accounting/backend/internal/middleware"
	"github.com/PhanAnh1001/ai-accounting/backend/internal/model"
)

// AuthUserRepository is the subset of UserRepository needed by AuthHandler.
type AuthUserRepository interface {
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	Create(ctx context.Context, user *model.User) error
	GetByID(ctx context.Context, id string) (*model.User, error)
	UpdatePassword(ctx context.Context, id, hash string) error
}

// AuthOrgRepository is the subset of OrganizationRepository needed by AuthHandler.
type AuthOrgRepository interface {
	Create(ctx context.Context, org *model.Organization) error
}

type AuthHandler struct {
	userRepo AuthUserRepository
	orgRepo  AuthOrgRepository
	secret   string
}

func NewAuthHandler(userRepo AuthUserRepository, orgRepo AuthOrgRepository, secret string) *AuthHandler {
	return &AuthHandler{userRepo: userRepo, orgRepo: orgRepo, secret: secret}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req model.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if req.Email == "" || req.Password == "" || req.CompanyName == "" {
		http.Error(w, `{"error":"email, password, and company_name are required"}`, http.StatusBadRequest)
		return
	}

	// Check if email already exists
	if _, err := h.userRepo.GetByEmail(r.Context(), req.Email); err == nil {
		http.Error(w, `{"error":"email already registered"}`, http.StatusConflict)
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	orgID := generateID()
	org := &model.Organization{
		ID:                 orgID,
		Name:               req.CompanyName,
		AccountingStandard: model.TT133,
	}
	if err := h.orgRepo.Create(r.Context(), org); err != nil {
		http.Error(w, `{"error":"failed to create organization"}`, http.StatusInternalServerError)
		return
	}

	userID := generateID()
	fullName := req.FullName
	user := &model.User{
		ID:             userID,
		OrganizationID: orgID,
		Email:          req.Email,
		PasswordHash:   string(hash),
		FullName:       &fullName,
		Role:           model.RoleOwner,
	}
	if err := h.userRepo.Create(r.Context(), user); err != nil {
		http.Error(w, `{"error":"failed to create user"}`, http.StatusInternalServerError)
		return
	}

	token, err := h.generateToken(userID, orgID)
	if err != nil {
		http.Error(w, `{"error":"failed to generate token"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(model.AuthResponse{Token: token, User: *user})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req model.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	user, err := h.userRepo.GetByEmail(r.Context(), req.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, `{"error":"invalid credentials"}`, http.StatusUnauthorized)
			return
		}
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		http.Error(w, `{"error":"invalid credentials"}`, http.StatusUnauthorized)
		return
	}

	token, err := h.generateToken(user.ID, user.OrganizationID)
	if err != nil {
		http.Error(w, `{"error":"failed to generate token"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(model.AuthResponse{Token: token, User: *user})
}

// ChangePassword handles POST /api/auth/change-password.
func (h *AuthHandler) ChangePassword(w http.ResponseWriter, r *http.Request) {
	userID := middleware.GetUserID(r.Context())
	if userID == "" {
		http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
		return
	}

	var req struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"invalid request body"}`, http.StatusBadRequest)
		return
	}

	if len(req.NewPassword) < 8 {
		http.Error(w, `{"error":"new_password must be at least 8 characters"}`, http.StatusBadRequest)
		return
	}

	user, err := h.userRepo.GetByID(r.Context(), userID)
	if err != nil {
		http.Error(w, `{"error":"user not found"}`, http.StatusNotFound)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.CurrentPassword)); err != nil {
		http.Error(w, `{"error":"current password is incorrect"}`, http.StatusUnauthorized)
		return
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, `{"error":"internal server error"}`, http.StatusInternalServerError)
		return
	}

	if err := h.userRepo.UpdatePassword(r.Context(), userID, string(newHash)); err != nil {
		http.Error(w, `{"error":"failed to update password"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "password updated successfully"})
}

func (h *AuthHandler) generateToken(userID, orgID string) (string, error) {
	claims := middleware.Claims{
		UserID: userID,
		OrgID:  orgID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(h.secret))
}
