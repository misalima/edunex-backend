package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/misalima/edunex-backend/internal/core/domain"
	"gorm.io/gorm"
)

// RefreshTokenModel maps to refresh_tokens table
type RefreshTokenModel struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index:idx_refresh_tokens_user_id"`
	Token     string    `gorm:"type:text;uniqueIndex;not null"`
	ExpiresAt time.Time `gorm:"column:expires_at;not null"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (RefreshTokenModel) TableName() string {
	return "refresh_tokens"
}

// toDomain converts GORM model to domain.RefreshToken
func (m *RefreshTokenModel) toDomain() *domain.RefreshToken {
	if m == nil {
		return nil
	}
	return &domain.RefreshToken{
		ID:        m.ID,
		UserID:    m.UserID,
		Token:     m.Token,
		ExpiresAt: m.ExpiresAt,
		Created:   m.CreatedAt,
	}
}

// GormAuthRepository is a GORM implementation for refresh token operations
type GormAuthRepository struct {
	db *gorm.DB
}

// NewGormAuthRepository creates a new auth repository
func NewGormAuthRepository(db *gorm.DB) *GormAuthRepository {
	return &GormAuthRepository{db: db}
}

// CreateRefreshToken inserts a new refresh token and returns the domain model
func (r *GormAuthRepository) CreateRefreshToken(ctx context.Context, userID uuid.UUID, token string, expiresAt time.Time) (*domain.RefreshToken, error) {
	m := &RefreshTokenModel{
		UserID:    userID,
		Token:     token,
		ExpiresAt: expiresAt,
	}
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return nil, err
	}
	return m.toDomain(), nil
}

// FindRefreshTokenByToken looks up a refresh token by its token value and returns the domain model. Returns (nil, nil) if not found.
func (r *GormAuthRepository) FindRefreshTokenByToken(ctx context.Context, token string) (*domain.RefreshToken, error) {
	m := &RefreshTokenModel{}
	if err := r.db.WithContext(ctx).First(m, "token = ?", token).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return m.toDomain(), nil
}

// DeleteRefreshTokenByToken deletes a refresh token by its token value
func (r *GormAuthRepository) DeleteRefreshTokenByToken(ctx context.Context, token string) error {
	if err := r.db.WithContext(ctx).Where("token = ?", token).Delete(&RefreshTokenModel{}).Error; err != nil {
		return err
	}
	return nil
}
