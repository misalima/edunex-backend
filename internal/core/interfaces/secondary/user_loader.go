package secondary

import (
	"context"

	"github.com/google/uuid"
	"github.com/misalima/edunex-backend/internal/core/domain"
)

type UserLoader interface {
	InsertUser(ctx context.Context, user *domain.User) (uuid.UUID, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domain.User, error)
	UpdateUser(ctx context.Context, user *domain.User) error
	ListUsers(ctx context.Context) ([]*domain.User, error)
	UpdateRole(ctx context.Context, userID uuid.UUID, role string) error
	Exists(ctx context.Context, id uuid.UUID) (bool, error)
	UpsertUser(ctx context.Context, user *domain.User) error
}
