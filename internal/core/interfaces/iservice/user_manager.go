package iservice

import (
	"context"

	"github.com/google/uuid"
	"github.com/misalima/edunex-backend/internal/core/domain"
)

type UserManager interface {
	CreateUser(ctx context.Context, user *domain.User) (*domain.User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	UpdateUser(ctx context.Context, user *domain.User) error
	ListUsers(ctx context.Context) ([]*domain.User, error)
}
