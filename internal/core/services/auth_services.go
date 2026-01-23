package services

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/misalima/edunex-backend/internal/core/interfaces/irepository"
	"github.com/misalima/edunex-backend/internal/core/interfaces/iservice"
	"github.com/misalima/edunex-backend/internal/core/util"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("credenciais inválidas")
	ErrUserNotFound       = errors.New("usuário não encontrado")
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
		return nil, ErrUserNotFound
	}
	if user == nil {
		return nil, ErrUserNotFound
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	accessToken, err := s.jwtSvc.GenerateToken(user.ID.String())
	if err != nil {
		return nil, err
	}

	refreshToken := uuid.New().String()
	expiresAt := time.Now().Add(time.Hour * 7 * 24)

	refToken, err := s.authRepo.CreateRefreshToken(ctx, user.ID, refreshToken, expiresAt)
	if err != nil {
		return nil, errors.New("error creating refresh token")
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
		return "", errors.New("invalid session")
	}

	if time.Now().After(tokenData.ExpiresAt) {
		_ = s.authRepo.DeleteRefreshTokenByToken(ctx, refreshToken)
		return "", errors.New("invalid session")
	}

	newAccessToken, err := s.jwtSvc.GenerateToken(tokenData.UserID.String())
	if err != nil {
		return "", errors.New("error generating new access token")
	}

	return newAccessToken, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	return s.authRepo.DeleteRefreshTokenByToken(ctx, refreshToken)
}
