package services

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/misalima/edunex-backend/internal/core/domain_errors"
	"github.com/misalima/edunex-backend/internal/core/interfaces/irepository"
	"github.com/misalima/edunex-backend/internal/core/interfaces/iservice"
	"github.com/misalima/edunex-backend/internal/core/util"
	"github.com/misalima/edunex-backend/internal/infra/logger"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	authRepo irepository.AuthLoader
	userRepo irepository.UserLoader
	jwtSvc   iservice.JWTManager
}

func NewAuthService(authRepo irepository.AuthLoader, userRepo irepository.UserLoader, jwtService iservice.JWTManager) *AuthService {
	return &AuthService{
		authRepo: authRepo,
		userRepo: userRepo,
		jwtSvc:   jwtService,
	}
}

func (s *AuthService) Login(ctx context.Context, email, password string) (*util.LoginResponse, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, email)
	if err != nil {
		logger.Log.Error("login failed", zap.Error(err))
		return nil, err
	}
	if user == nil {
		logger.Log.Error("login failed: user not found",
			zap.String("email", email),
			zap.Error(err))
		return nil, domain_errors.ErrUserNotFound
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		logger.Log.Error("login failed: invalid credentials", zap.Error(err))
		return nil, domain_errors.ErrInvalidCredentials
	}

	accessToken, err := s.jwtSvc.GenerateToken(user.ID.String(), user.Role)
	if err != nil {
		logger.Log.Error("login failed: error generating access token", zap.Error(err))
		return nil, domain_errors.WrapUnexpectedMsg(err, "error generating access token")
	}

	refreshToken := uuid.New().String()
	expiresAt := time.Now().Add(time.Hour * 7 * 24)

	refToken, err := s.authRepo.CreateRefreshToken(ctx, user.ID, refreshToken, expiresAt)
	if err != nil {
		logger.Log.Error("login failed: error creating refresh token", zap.Error(err))
		return nil, domain_errors.WrapUnexpectedMsg(err, "error creating refresh token")
	}

	user.Password = ""

	return &util.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refToken.Token,
		User:         user,
	}, nil
}

func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (string, error) {
	tokenData, err := s.authRepo.FindRefreshTokenByToken(ctx, refreshToken)
	if err != nil {
		return "", domain_errors.NewUnauthorizedMsg("invalid session")
	}

	if tokenData == nil || time.Now().After(tokenData.ExpiresAt) {
		_ = s.authRepo.DeleteRefreshTokenByToken(ctx, refreshToken)
		return "", domain_errors.NewUnauthorizedMsg("invalid session")
	}

	user, err := s.userRepo.GetUserByID(ctx, tokenData.UserID)
	if err != nil || user == nil {
		return "", domain_errors.ErrUserNotFound
	}

	newAccessToken, err := s.jwtSvc.GenerateToken(tokenData.UserID.String(), user.Role)
	if err != nil {
		return "", domain_errors.WrapUnexpectedMsg(err, "error generating new access token")
	}

	return newAccessToken, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	if err := s.authRepo.DeleteRefreshTokenByToken(ctx, refreshToken); err != nil {
		return domain_errors.WrapUnexpectedMsg(err, "error deleting refresh token")
	}
	return nil
}
