package services_test

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/misalima/edunex-backend/internal/core/domain"
	"github.com/misalima/edunex-backend/internal/core/services"
	"github.com/misalima/edunex-backend/internal/core/util"
)

type mockJWTValidator struct {
	validateTokenFn func(token string) (*util.TokenClaims, error)
}

func (m *mockJWTValidator) ValidateToken(token string) (*util.TokenClaims, error) {
	if m.validateTokenFn != nil {
		return m.validateTokenFn(token)
	}
	return nil, errors.New("not implemented")
}

func (m *mockJWTValidator) ValidateTokenViaAPI(token string) (*util.TokenClaims, error) {
	return nil, errors.New("not implemented")
}

type mockUserLoader struct {
	existsCalls    atomic.Int64
	upsertCalls    atomic.Int64
	existsFn       func(ctx context.Context, id uuid.UUID) (bool, error)
	upsertFn       func(ctx context.Context, user *domain.User) error
	getUserByIDFn  func(ctx context.Context, id uuid.UUID) (*domain.User, error)
	getUserByEmail func(ctx context.Context, email string) (*domain.User, error)
	insertUserFn   func(ctx context.Context, user *domain.User) (uuid.UUID, error)
	updateUserFn   func(ctx context.Context, user *domain.User) error
	listUsersFn    func(ctx context.Context) ([]*domain.User, error)
	updateRoleFn   func(ctx context.Context, userID uuid.UUID, role string) error
}

func (m *mockUserLoader) Exists(ctx context.Context, id uuid.UUID) (bool, error) {
	m.existsCalls.Add(1)
	if m.existsFn != nil {
		return m.existsFn(ctx, id)
	}
	return false, nil
}

func (m *mockUserLoader) UpsertUser(ctx context.Context, user *domain.User) error {
	m.upsertCalls.Add(1)
	if m.upsertFn != nil {
		return m.upsertFn(ctx, user)
	}
	return nil
}

func (m *mockUserLoader) InsertUser(ctx context.Context, user *domain.User) (uuid.UUID, error) {
	if m.insertUserFn != nil {
		return m.insertUserFn(ctx, user)
	}
	return uuid.New(), nil
}

func (m *mockUserLoader) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	if m.getUserByIDFn != nil {
		return m.getUserByIDFn(ctx, id)
	}
	return nil, nil
}

func (m *mockUserLoader) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	if m.getUserByEmail != nil {
		return m.getUserByEmail(ctx, email)
	}
	return nil, nil
}

func (m *mockUserLoader) UpdateUser(ctx context.Context, user *domain.User) error {
	if m.updateUserFn != nil {
		return m.updateUserFn(ctx, user)
	}
	return nil
}

func (m *mockUserLoader) ListUsers(ctx context.Context) ([]*domain.User, error) {
	if m.listUsersFn != nil {
		return m.listUsersFn(ctx)
	}
	return nil, nil
}

func (m *mockUserLoader) UpdateRole(ctx context.Context, userID uuid.UUID, role string) error {
	if m.updateRoleFn != nil {
		return m.updateRoleFn(ctx, userID, role)
	}
	return nil
}

func TestAuthService_Authenticate_UserExistsInDB(t *testing.T) {
	userID := uuid.New()
	email := "teacher@edunex.com"

	mockJWT := &mockJWTValidator{
		validateTokenFn: func(token string) (*util.TokenClaims, error) {
			return &util.TokenClaims{
				UserID: userID.String(),
				Email:  email,
			}, nil
		},
	}

	mockRepo := &mockUserLoader{
		existsFn: func(ctx context.Context, id uuid.UUID) (bool, error) {
			if id != userID {
				t.Errorf("expected userID %s, got %s", userID, id)
			}
			return true, nil
		},
	}

	svc := services.NewAuthService(mockJWT, mockRepo)

	claims, err := svc.Authenticate(context.Background(), "valid-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if claims.UserID != userID.String() {
		t.Errorf("expected UserID %s, got %s", userID.String(), claims.UserID)
	}
	if claims.Email != email {
		t.Errorf("expected Email %s, got %s", email, claims.Email)
	}
	if calls := mockRepo.existsCalls.Load(); calls != 1 {
		t.Errorf("expected 1 Exists call, got %d", calls)
	}
	if calls := mockRepo.upsertCalls.Load(); calls != 0 {
		t.Errorf("expected 0 Upsert calls, got %d", calls)
	}
}

func TestAuthService_Authenticate_LazySyncsNewUser(t *testing.T) {
	userID := uuid.New()
	email := "new.teacher@edunex.com"

	mockJWT := &mockJWTValidator{
		validateTokenFn: func(token string) (*util.TokenClaims, error) {
			return &util.TokenClaims{
				UserID: userID.String(),
				Email:  email,
			}, nil
		},
	}

	var upsertedUser *domain.User
	mockRepo := &mockUserLoader{
		existsFn: func(ctx context.Context, id uuid.UUID) (bool, error) {
			return false, nil // user does not exist
		},
		upsertFn: func(ctx context.Context, user *domain.User) error {
			upsertedUser = user
			return nil
		},
	}

	svc := services.NewAuthService(mockJWT, mockRepo)

	claims, err := svc.Authenticate(context.Background(), "valid-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if claims.UserID != userID.String() {
		t.Errorf("expected UserID %s, got %s", userID.String(), claims.UserID)
	}
	if claims.Email != email {
		t.Errorf("expected Email %s, got %s", email, claims.Email)
	}
	if calls := mockRepo.existsCalls.Load(); calls != 1 {
		t.Errorf("expected 1 Exists call, got %d", calls)
	}
	if calls := mockRepo.upsertCalls.Load(); calls != 1 {
		t.Errorf("expected 1 Upsert call, got %d", calls)
	}
	if upsertedUser == nil {
		t.Fatal("expected upserted user to not be nil")
	}
	if upsertedUser.ID != userID {
		t.Errorf("expected ID %s, got %s", userID, upsertedUser.ID)
	}
	if upsertedUser.Email != email {
		t.Errorf("expected Email %s, got %s", email, upsertedUser.Email)
	}
	if upsertedUser.Name != "new.teacher" {
		t.Errorf("expected Name 'new.teacher', got '%s'", upsertedUser.Name)
	}
	if upsertedUser.Role != "coordinator" {
		t.Errorf("expected Role 'coordinator', got '%s'", upsertedUser.Role)
	}
}

func TestAuthService_Authenticate_CacheHit(t *testing.T) {
	userID := uuid.New()
	email := "cached@edunex.com"

	mockJWT := &mockJWTValidator{
		validateTokenFn: func(token string) (*util.TokenClaims, error) {
			return &util.TokenClaims{
				UserID: userID.String(),
				Email:  email,
			}, nil
		},
	}

	mockRepo := &mockUserLoader{
		existsFn: func(ctx context.Context, id uuid.UUID) (bool, error) {
			return true, nil
		},
	}

	svc := services.NewAuthServiceWithTTL(mockJWT, mockRepo, 1*time.Minute)

	// First call - cache miss, hits DB
	claims1, err := svc.Authenticate(context.Background(), "token")
	if err != nil {
		t.Fatalf("first authenticate call failed: %v", err)
	}
	if claims1.UserID != userID.String() {
		t.Errorf("expected UserID %s, got %s", userID.String(), claims1.UserID)
	}
	if calls := mockRepo.existsCalls.Load(); calls != 1 {
		t.Errorf("expected 1 Exists call after first invoke, got %d", calls)
	}

	// Second call - cache hit, does NOT hit DB
	claims2, err := svc.Authenticate(context.Background(), "token")
	if err != nil {
		t.Fatalf("second authenticate call failed: %v", err)
	}
	if claims2.UserID != userID.String() {
		t.Errorf("expected UserID %s, got %s", userID.String(), claims2.UserID)
	}
	if calls := mockRepo.existsCalls.Load(); calls != 1 {
		t.Errorf("expected still 1 Exists call on cache hit, got %d", calls)
	}
}

func TestAuthService_Authenticate_InvalidToken(t *testing.T) {
	mockJWT := &mockJWTValidator{
		validateTokenFn: func(token string) (*util.TokenClaims, error) {
			return nil, errors.New("signature is invalid")
		},
	}

	mockRepo := &mockUserLoader{}
	svc := services.NewAuthService(mockJWT, mockRepo)

	claims, err := svc.Authenticate(context.Background(), "bad-token")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if claims != nil {
		t.Errorf("expected nil claims, got %v", claims)
	}
	if !strings.Contains(err.Error(), "invalid token") {
		t.Errorf("expected error to contain 'invalid token', got: %v", err)
	}
	if calls := mockRepo.existsCalls.Load(); calls != 0 {
		t.Errorf("expected 0 Exists calls on invalid token, got %d", calls)
	}
}

func TestAuthService_Authenticate_InvalidUserID(t *testing.T) {
	mockJWT := &mockJWTValidator{
		validateTokenFn: func(token string) (*util.TokenClaims, error) {
			return &util.TokenClaims{
				UserID: "not-a-valid-uuid",
				Email:  "test@edunex.com",
			}, nil
		},
	}

	mockRepo := &mockUserLoader{}
	svc := services.NewAuthService(mockJWT, mockRepo)

	claims, err := svc.Authenticate(context.Background(), "token")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if claims != nil {
		t.Errorf("expected nil claims, got %v", claims)
	}
	if !strings.Contains(err.Error(), "invalid user id") {
		t.Errorf("expected error to contain 'invalid user id', got: %v", err)
	}
	if calls := mockRepo.existsCalls.Load(); calls != 0 {
		t.Errorf("expected 0 Exists calls on invalid user id, got %d", calls)
	}
}

func TestAuthService_Authenticate_DBError(t *testing.T) {
	userID := uuid.New()
	mockJWT := &mockJWTValidator{
		validateTokenFn: func(token string) (*util.TokenClaims, error) {
			return &util.TokenClaims{
				UserID: userID.String(),
				Email:  "test@edunex.com",
			}, nil
		},
	}

	mockRepo := &mockUserLoader{
		existsFn: func(ctx context.Context, id uuid.UUID) (bool, error) {
			return false, errors.New("db connection timeout")
		},
	}

	svc := services.NewAuthService(mockJWT, mockRepo)

	claims, err := svc.Authenticate(context.Background(), "token")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if claims != nil {
		t.Errorf("expected nil claims, got %v", claims)
	}
	if !strings.Contains(err.Error(), "failed to check user existence") {
		t.Errorf("expected error to contain 'failed to check user existence', got: %v", err)
	}
}

func TestAuthService_Authenticate_UpsertError(t *testing.T) {
	userID := uuid.New()
	mockJWT := &mockJWTValidator{
		validateTokenFn: func(token string) (*util.TokenClaims, error) {
			return &util.TokenClaims{
				UserID: userID.String(),
				Email:  "test@edunex.com",
			}, nil
		},
	}

	mockRepo := &mockUserLoader{
		existsFn: func(ctx context.Context, id uuid.UUID) (bool, error) {
			return false, nil
		},
		upsertFn: func(ctx context.Context, user *domain.User) error {
			return errors.New("unique constraint violation")
		},
	}

	svc := services.NewAuthService(mockJWT, mockRepo)

	claims, err := svc.Authenticate(context.Background(), "token")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if claims != nil {
		t.Errorf("expected nil claims, got %v", claims)
	}
	if !strings.Contains(err.Error(), "failed to sync user") {
		t.Errorf("expected error to contain 'failed to sync user', got: %v", err)
	}
}
