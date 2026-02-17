package services_test

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/misalima/edunex-backend/internal/core/domain"
	"github.com/stretchr/testify/mock"
)

type MockAuthRepo struct {
	mock.Mock
}

func (m *MockAuthRepo) DeleteRefreshTokenByToken(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockAuthRepo) FindRefreshTokenByToken(ctx context.Context, token string) (*domain.RefreshToken, error) {
	args := m.Called(ctx, token)
	return args.Get(0).(*domain.RefreshToken), args.Error(1)
}

func (m *MockAuthRepo) Login(ctx context.Context, username, password string) (bool, error) {
	args := m.Called(ctx, username, password)
	return args.Bool(0), args.Error(1)
}

func (m *MockAuthRepo) CreateRefreshToken(ctx context.Context, userID uuid.UUID, token string, expiry time.Time) (*domain.RefreshToken, error) {
	args := m.Called(ctx, userID, token, expiry)
	return args.Get(0).(*domain.RefreshToken), args.Error(1)
}

type MockUserRepo struct {
	mock.Mock
}

func (m *MockUserRepo) UpdateUser(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepo) UpdateRole(ctx context.Context, userID uuid.UUID, role string) error {
	args := m.Called(ctx, userID, role)
	return args.Error(0)
}

func (m *MockUserRepo) ListUsers(ctx context.Context) ([]*domain.User, error) {
	args := m.Called(ctx)
	return args.Get(0).([]*domain.User), args.Error(1)
}

func (m *MockUserRepo) InsertUser(ctx context.Context, user *domain.User) (uuid.UUID, error) {
	args := m.Called(ctx, user)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockUserRepo) GetUserByID(ctx context.Context, userID uuid.UUID) (*domain.User, error) {
	args := m.Called(ctx, userID)
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepo) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	args := m.Called(ctx, email)
	return args.Get(0).(*domain.User), args.Error(1)
}

type MockJWTManager struct {
	mock.Mock
}

func (m *MockJWTManager) GenerateToken(userID, role string) (string, error) {
	args := m.Called(userID, role)
	return args.String(0), args.Error(1)
}