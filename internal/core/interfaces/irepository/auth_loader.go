package irepository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/misalima/edunex-backend/internal/core/domain"
)

type AuthLoader interface {
	CreateRefreshToken(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) (*domain.RefreshToken, error)
	FindRefreshTokenByToken(ctx context.Context, token string) (*domain.RefreshToken, error)
	DeleteRefreshTokenByToken(ctx context.Context, token string) error
}
