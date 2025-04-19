package irepository

import (
	"context"
	"github.com/google/uuid"
	"github.com/misalima/edunex-backend/internal/core/domain"
)

type UserLoader interface {
	InsertUser(ctx context.Context, user *domain.User) (string, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	UpdateUser(ctx context.Context, user *domain.User) error
}
