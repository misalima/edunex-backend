package iservice

import (
	"context"

	"github.com/misalima/edunex-backend/internal/core/util"
)

type AuthManager interface {
	Login(ctx context.Context, email, password string) (*util.LoginResponse, error)
	RefreshToken(ctx context.Context, refreshToken string) (string, error)
	Logout(ctx context.Context, refreshToken string) error
}
