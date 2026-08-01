package services

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/misalima/edunex-backend/internal/core/domain"
	"github.com/misalima/edunex-backend/internal/core/interfaces/primary"
	"github.com/misalima/edunex-backend/internal/core/interfaces/secondary"
	"github.com/misalima/edunex-backend/internal/core/util"
	"github.com/misalima/edunex-backend/internal/infra/logger"
	"github.com/misalima/edunex-backend/internal/infra/security"
	"go.uber.org/zap"
)

const defaultCacheTTL = 5 * time.Minute

var _ primary.Authenticator = (*AuthService)(nil)

type AuthService struct {
	jwtValidator security.JWTValidator
	userRepo     secondary.UserLoader
	userCache    map[uuid.UUID]time.Time
	cacheTTL     time.Duration
	cacheMu      sync.RWMutex
}

// NewAuthService creates a new AuthService with the default cache TTL (5 minutes).
func NewAuthService(
	jwtValidator security.JWTValidator,
	userRepo secondary.UserLoader,
) *AuthService {
	return NewAuthServiceWithTTL(jwtValidator, userRepo, defaultCacheTTL)
}

// NewAuthServiceWithTTL creates a new AuthService with a custom cache TTL.
func NewAuthServiceWithTTL(
	jwtValidator security.JWTValidator,
	userRepo secondary.UserLoader,
	cacheTTL time.Duration,
) *AuthService {
	return &AuthService{
		jwtValidator: jwtValidator,
		userRepo:     userRepo,
		userCache:    make(map[uuid.UUID]time.Time),
		cacheTTL:     cacheTTL,
	}
}

// Authenticate validates the JWT signature and transparently ensures the user exists locally in Postgres.
func (s *AuthService) Authenticate(ctx context.Context, tokenString string) (*util.TokenClaims, error) {
	// 1. Validate JWT signature locally
	claims, err := s.jwtValidator.ValidateToken(tokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil || userID == uuid.Nil {
		return nil, fmt.Errorf("invalid user id in token claims: %w", err)
	}

	// 2. Check thread-safe in-memory cache first to avoid unnecessary database queries
	s.cacheMu.RLock()
	cachedAt, found := s.userCache[userID]
	s.cacheMu.RUnlock()

	if found && time.Since(cachedAt) < s.cacheTTL {
		return claims, nil
	}

	// 3. Check if user exists in local database
	exists, err := s.userRepo.Exists(ctx, userID)
	if err != nil {
		logger.Log.Error("failed to verify user existence in db",
			zap.Error(err),
			zap.String("user_id", userID.String()))
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}

	if !exists {
		// 4. Lazy sync: create user in local Postgres
		userName := extractNameFromEmail(claims.Email)

		user := domain.User{
			ID:    userID,
			Email: claims.Email,
			Name:  userName,
			Role:  "coordinator",
		}

		if err := s.userRepo.UpsertUser(ctx, &user); err != nil {
			logger.Log.Error("failed to lazy-sync user to postgres",
				zap.Error(err),
				zap.String("user_id", userID.String()))
			return nil, fmt.Errorf("failed to sync user: %w", err)
		}

		logger.Log.Info("user lazily synchronized to postgres",
			zap.String("user_id", userID.String()),
			zap.String("email", claims.Email))
	}

	// 5. Update in-memory cache
	s.cacheMu.Lock()
	s.userCache[userID] = time.Now()
	s.cacheMu.Unlock()

	return claims, nil
}
