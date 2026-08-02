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
	"golang.org/x/sync/singleflight"
)

const (
	defaultCacheTTL        = 5 * time.Minute
	defaultCleanupInterval = 10 * time.Minute
	maxCacheEntries        = 10000
)

var _ primary.Authenticator = (*AuthService)(nil)

type AuthService struct {
	jwtValidator security.JWTValidator
	userRepo     secondary.UserLoader
	userCache    map[uuid.UUID]time.Time
	cacheTTL     time.Duration
	cacheMu      sync.RWMutex
	sfGroup      singleflight.Group
	stopCleaner  chan struct{}
	cleanerOnce  sync.Once
}

// NewAuthService creates a new AuthService with the default cache TTL (5 minutes) and automatic periodic cleanup.
func NewAuthService(
	jwtValidator security.JWTValidator,
	userRepo secondary.UserLoader,
) *AuthService {
	return NewAuthServiceWithTTL(jwtValidator, userRepo, defaultCacheTTL, defaultCleanupInterval)
}

// NewAuthServiceWithTTL creates a new AuthService with a custom cache TTL and cleanup interval.
func NewAuthServiceWithTTL(
	jwtValidator security.JWTValidator,
	userRepo secondary.UserLoader,
	cacheTTL time.Duration,
	cleanupInterval time.Duration,
) *AuthService {
	s := &AuthService{
		jwtValidator: jwtValidator,
		userRepo:     userRepo,
		userCache:    make(map[uuid.UUID]time.Time),
		cacheTTL:     cacheTTL,
		stopCleaner:  make(chan struct{}),
	}

	if cleanupInterval > 0 {
		go s.startCleaner(cleanupInterval)
	}

	return s
}

// startCleaner periodically purges expired user entries from the in-memory cache.
func (s *AuthService) startCleaner(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.CleanupExpired()
		case <-s.stopCleaner:
			return
		}
	}
}

// CleanupExpired purges all entries older than cacheTTL from memory.
func (s *AuthService) CleanupExpired() int {
	now := time.Now()
	s.cacheMu.Lock()
	defer s.cacheMu.Unlock()

	deletedCount := 0
	for id, cachedAt := range s.userCache {
		if now.Sub(cachedAt) > s.cacheTTL {
			delete(s.userCache, id)
			deletedCount++
		}
	}
	return deletedCount
}

// CacheSize returns the current number of cached users.
func (s *AuthService) CacheSize() int {
	s.cacheMu.RLock()
	defer s.cacheMu.RUnlock()
	return len(s.userCache)
}

// Close stops the background cache cleaner goroutine.
func (s *AuthService) Close() {
	s.cleanerOnce.Do(func() {
		close(s.stopCleaner)
	})
}

// Authenticate validates the JWT signature and transparently ensures the user exists locally in Postgres.
// It uses an in-memory cache and singleflight deduplication to avoid cache stampedes and redundant DB lookups.
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

	if claims.Email == "" {
		return nil, fmt.Errorf("email claim is required in token")
	}

	// 2. Check thread-safe in-memory cache first (fast path)
	s.cacheMu.RLock()
	cachedAt, found := s.userCache[userID]
	s.cacheMu.RUnlock()

	if found && time.Since(cachedAt) < s.cacheTTL {
		return claims, nil
	}

	// 3. Prevent cache stampede using singleflight for concurrent requests of the same user
	_, err, _ = s.sfGroup.Do(userID.String(), func() (interface{}, error) {
		// Double-check cache inside singleflight lock
		s.cacheMu.RLock()
		cachedAt, found := s.userCache[userID]
		s.cacheMu.RUnlock()

		if found && time.Since(cachedAt) < s.cacheTTL {
			return nil, nil
		}

		// Check if user exists in local database
		exists, err := s.userRepo.Exists(ctx, userID)
		if err != nil {
			logger.Log.Error("failed to verify user existence in db",
				zap.Error(err),
				zap.String("user_id", userID.String()))
			return nil, fmt.Errorf("failed to check user existence: %w", err)
		}

		if !exists {
			// Lazy sync: create user in local Postgres
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

		// Update in-memory cache with overflow protection
		s.cacheMu.Lock()
		if len(s.userCache) >= maxCacheEntries {
			now := time.Now()
			for id, ts := range s.userCache {
				if now.Sub(ts) > s.cacheTTL {
					delete(s.userCache, id)
				}
			}
		}
		s.userCache[userID] = time.Now()
		s.cacheMu.Unlock()

		return nil, nil
	})

	if err != nil {
		return nil, err
	}

	return claims, nil
}
