package test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/misalima/edunex-backend/internal/core/domain"
	"github.com/misalima/edunex-backend/internal/core/domain_errors"
	"github.com/misalima/edunex-backend/internal/core/services"
	"github.com/misalima/edunex-backend/internal/infra/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mocks
type mockAuthRepo struct {
	mock.Mock
}

func (m *mockAuthRepo) CreateRefreshToken(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) (*domain.RefreshToken, error) {
	args := m.Called(ctx, userID, token, expiresAt)
	return args.Get(0).(*domain.RefreshToken), args.Error(1)
}

func (m *mockAuthRepo) FindRefreshTokenByToken(ctx context.Context, token string) (*domain.RefreshToken, error) {
	args := m.Called(ctx, token)
	return args.Get(0).(*domain.RefreshToken), args.Error(1)
}

func (m *mockAuthRepo) DeleteRefreshTokenByToken(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

type mockUserRepo struct {
	mock.Mock
}

func (m *mockUserRepo) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserRepo) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *mockUserRepo) InsertUser(ctx context.Context, user *domain.User) (uuid.UUID, error) {
	args := m.Called(ctx, user)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *mockUserRepo) UpdateUser(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *mockUserRepo) ListUsers(ctx context.Context) ([]*domain.User, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*domain.User), args.Error(1)
}

func (m *mockUserRepo) UpdateRole(ctx context.Context, userID uuid.UUID, role string) error {
	args := m.Called(ctx, userID, role)
	return args.Error(0)
}

type mockJWTService struct {
	mock.Mock
}

func init() {
	// Initialize logger for tests
	logger.InitLogger()
}

func (m *mockJWTService) GenerateToken(userID string, role string) (string, error) {
	args := m.Called(userID, role)
	return args.String(0), args.Error(1)
}

func (m *mockJWTService) ValidateToken(token string) (string, string, error) {
	args := m.Called(token)
	return args.String(0), args.String(1), args.Error(2)
}

func TestAuthService_Login_Success(t *testing.T) {
	// Setup
	ctx := context.Background()
	userID := uuid.New()
	email := "test@example.com"
	password := "password123"
	hashedPassword := "$2a$10$FW69DWtPJ.5vRnq1EeUMUuPVf1FI17XhXGOW94qlNJ1P.gK3bdB3." // bcrypt hash válido

	user := &domain.User{
		ID:       userID,
		Email:    email,
		Password: hashedPassword,
		Role:     "admin",
	}

	authRepo := new(mockAuthRepo)
	userRepo := new(mockUserRepo)
	jwtSvc := new(mockJWTService)

	authService := services.NewAuthService(authRepo, userRepo, jwtSvc)

	// Expectations
	userRepo.On("GetUserByEmail", ctx, email).Return(user, nil)
	jwtSvc.On("GenerateToken", userID.String(), "admin").Return("access-token", nil)
	authRepo.On("CreateRefreshToken", ctx, userID, mock.Anything, mock.Anything).Return(&domain.RefreshToken{
		Token: "refresh-token",
	}, nil)

	// Execute
	result, err := authService.Login(ctx, email, password)

	// Verify
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "access-token", result.AccessToken)
	assert.Equal(t, "refresh-token", result.RefreshToken)
	assert.Equal(t, userID, result.User.ID)
	assert.Equal(t, "", result.User.Password) // Password should be cleared

	authRepo.AssertExpectations(t)
	userRepo.AssertExpectations(t)
	jwtSvc.AssertExpectations(t)
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	// Setup
	ctx := context.Background()
	email := "nonexistent@example.com"
	password := "password123"

	authRepo := new(mockAuthRepo)
	userRepo := new(mockUserRepo)
	jwtSvc := new(mockJWTService)

	authService := services.NewAuthService(authRepo, userRepo, jwtSvc)

	// Expectations
	userRepo.On("GetUserByEmail", ctx, email).Return((*domain.User)(nil), domain_errors.ErrUserNotFound)

	// Execute
	result, err := authService.Login(ctx, email, password)

	// Verify
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Equal(t, domain_errors.ErrUserNotFound, err)

	userRepo.AssertExpectations(t)
}

// Adicionar mais testes conforme necessário
