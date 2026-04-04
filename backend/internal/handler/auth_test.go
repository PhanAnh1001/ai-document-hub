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

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"

	"github.com/PhanAnh1001/ai-accounting/backend/internal/middleware"
	"github.com/PhanAnh1001/ai-accounting/backend/internal/model"
)

// --- mock repositories ---

type mockAuthUserRepo struct {
	getByEmailFn      func(ctx context.Context, email string) (*model.User, error)
	createFn          func(ctx context.Context, user *model.User) error
	getByIDFn         func(ctx context.Context, id string) (*model.User, error)
	updatePasswordFn  func(ctx context.Context, id, hash string) error
}

func (m *mockAuthUserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	if m.getByEmailFn != nil {
		return m.getByEmailFn(ctx, email)
	}
	return nil, pgx.ErrNoRows
}

func (m *mockAuthUserRepo) Create(ctx context.Context, user *model.User) error {
	if m.createFn != nil {
		return m.createFn(ctx, user)
	}
	return nil
}

func (m *mockAuthUserRepo) GetByID(ctx context.Context, id string) (*model.User, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return nil, fmt.Errorf("not found")
}

func (m *mockAuthUserRepo) UpdatePassword(ctx context.Context, id, hash string) error {
	if m.updatePasswordFn != nil {
		return m.updatePasswordFn(ctx, id, hash)
	}
	return nil
}

type mockAuthOrgRepo struct {
	createFn func(ctx context.Context, org *model.Organization) error
}

func (m *mockAuthOrgRepo) Create(ctx context.Context, org *model.Organization) error {
	if m.createFn != nil {
		return m.createFn(ctx, org)
	}
	return nil
}

func sampleHashedUser(email string) *model.User {
	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)
	name := "Test User"
	return &model.User{
		ID:             "user-1",
		OrganizationID: "org-1",
		Email:          email,
		PasswordHash:   string(hash),
		FullName:       &name,
		Role:           model.RoleOwner,
		CreatedAt:      time.Now(),
	}
}

func TestAuthHandler_Register_Success(t *testing.T) {
	userRepo := &mockAuthUserRepo{
		// email doesn't exist yet
		getByEmailFn: func(_ context.Context, _ string) (*model.User, error) {
			return nil, pgx.ErrNoRows
		},
	}
	orgRepo := &mockAuthOrgRepo{}
	h := NewAuthHandler(userRepo, orgRepo, "test-secret")

	body := `{"email":"test@example.com","password":"password123","company_name":"Công ty Test","full_name":"Nguyen Van A"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.Register(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["token"] == nil || resp["token"] == "" {
		t.Error("expected non-empty token in response")
	}
}

func TestAuthHandler_Register_MissingFields_Returns400(t *testing.T) {
	h := NewAuthHandler(&mockAuthUserRepo{}, &mockAuthOrgRepo{}, "secret")

	body := `{"email":"test@example.com","password":"pass123"}` // missing company_name
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.Register(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestAuthHandler_Register_DuplicateEmail_Returns409(t *testing.T) {
	userRepo := &mockAuthUserRepo{
		getByEmailFn: func(_ context.Context, _ string) (*model.User, error) {
			return sampleHashedUser("test@example.com"), nil // already exists
		},
	}
	h := NewAuthHandler(userRepo, &mockAuthOrgRepo{}, "secret")

	body := `{"email":"test@example.com","password":"pass123","company_name":"Co"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.Register(rr, req)

	if rr.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d", rr.Code)
	}
}

func TestAuthHandler_Login_Success(t *testing.T) {
	userRepo := &mockAuthUserRepo{
		getByEmailFn: func(_ context.Context, _ string) (*model.User, error) {
			return sampleHashedUser("test@example.com"), nil
		},
	}
	h := NewAuthHandler(userRepo, &mockAuthOrgRepo{}, "test-secret")

	body := `{"email":"test@example.com","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.Login(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["token"] == nil || resp["token"] == "" {
		t.Error("expected non-empty token")
	}
}

func TestAuthHandler_Login_WrongPassword_Returns401(t *testing.T) {
	userRepo := &mockAuthUserRepo{
		getByEmailFn: func(_ context.Context, _ string) (*model.User, error) {
			return sampleHashedUser("test@example.com"), nil
		},
	}
	h := NewAuthHandler(userRepo, &mockAuthOrgRepo{}, "secret")

	body := `{"email":"test@example.com","password":"wrongpassword"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.Login(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestAuthHandler_Login_UnknownEmail_Returns401(t *testing.T) {
	userRepo := &mockAuthUserRepo{
		getByEmailFn: func(_ context.Context, _ string) (*model.User, error) {
			return nil, fmt.Errorf("get user by email: %w", pgx.ErrNoRows)
		},
	}
	h := NewAuthHandler(userRepo, &mockAuthOrgRepo{}, "secret")

	body := `{"email":"nobody@example.com","password":"pass"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/login", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()

	h.Login(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func buildAuthUserContext(r *http.Request, userID, orgID string) *http.Request {
	ctx := context.WithValue(r.Context(), middleware.UserIDKey, userID)
	ctx = context.WithValue(ctx, middleware.OrgIDKey, orgID)
	return r.WithContext(ctx)
}

func TestAuthHandler_ChangePassword_Success(t *testing.T) {
	oldHash, _ := bcrypt.GenerateFromPassword([]byte("OldPass1!"), bcrypt.MinCost)
	user := &model.User{
		ID:           "user-1",
		PasswordHash: string(oldHash),
	}

	var capturedHash string
	repo := &mockAuthUserRepo{
		getByIDFn: func(_ context.Context, id string) (*model.User, error) {
			return user, nil
		},
		updatePasswordFn: func(_ context.Context, id, hash string) error {
			capturedHash = hash
			return nil
		},
	}
	h := NewAuthHandler(repo, &mockAuthOrgRepo{}, "secret")

	body := `{"current_password":"OldPass1!","new_password":"NewPass1!"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/change-password", strings.NewReader(body))
	req = buildAuthUserContext(req, "user-1", "org-1")
	rr := httptest.NewRecorder()

	h.ChangePassword(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rr.Code, rr.Body.String())
	}
	if capturedHash == "" {
		t.Error("expected UpdatePassword to be called")
	}
	// Verify new hash works
	if err := bcrypt.CompareHashAndPassword([]byte(capturedHash), []byte("NewPass1!")); err != nil {
		t.Error("new password hash doesn't match expected password")
	}
}

func TestAuthHandler_ChangePassword_WrongCurrentPassword(t *testing.T) {
	oldHash, _ := bcrypt.GenerateFromPassword([]byte("OldPass1!"), bcrypt.MinCost)
	user := &model.User{ID: "user-1", PasswordHash: string(oldHash)}

	repo := &mockAuthUserRepo{
		getByIDFn: func(_ context.Context, _ string) (*model.User, error) { return user, nil },
	}
	h := NewAuthHandler(repo, &mockAuthOrgRepo{}, "secret")

	body := `{"current_password":"WrongPass!","new_password":"NewPass1!"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/change-password", strings.NewReader(body))
	req = buildAuthUserContext(req, "user-1", "org-1")
	rr := httptest.NewRecorder()

	h.ChangePassword(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestAuthHandler_ChangePassword_Unauthorized(t *testing.T) {
	h := NewAuthHandler(&mockAuthUserRepo{}, &mockAuthOrgRepo{}, "secret")

	req := httptest.NewRequest(http.MethodPost, "/api/auth/change-password",
		strings.NewReader(`{"current_password":"a","new_password":"b"}`))
	rr := httptest.NewRecorder()

	h.ChangePassword(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rr.Code)
	}
}

func TestAuthHandler_ChangePassword_ShortNewPassword(t *testing.T) {
	oldHash, _ := bcrypt.GenerateFromPassword([]byte("OldPass1!"), bcrypt.MinCost)
	user := &model.User{ID: "user-1", PasswordHash: string(oldHash)}

	repo := &mockAuthUserRepo{
		getByIDFn: func(_ context.Context, _ string) (*model.User, error) { return user, nil },
	}
	h := NewAuthHandler(repo, &mockAuthOrgRepo{}, "secret")

	body := `{"current_password":"OldPass1!","new_password":"short"}`
	req := httptest.NewRequest(http.MethodPost, "/api/auth/change-password", strings.NewReader(body))
	req = buildAuthUserContext(req, "user-1", "org-1")
	rr := httptest.NewRecorder()

	h.ChangePassword(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}
