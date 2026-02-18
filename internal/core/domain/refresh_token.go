package domain

import (
	"time"

	"github.com/google/uuid"
)

// RefreshToken represents a refresh token in the domain layer.
// Keep this struct free of persistence (GORM) tags to preserve hexagonal boundaries.
type RefreshToken struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	Token     string
	ExpiresAt time.Time
	Created   time.Time
}

// IsExpired returns true if the refresh token is already expired.
func (r *RefreshToken) IsExpired() bool {
	if r == nil {
		return true
	}
	return time.Now().After(r.ExpiresAt)
}
