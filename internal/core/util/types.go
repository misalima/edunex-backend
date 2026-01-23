package util

import "github.com/misalima/edunex-backend/internal/core/domain"

type LoginResponse struct {
	AccessToken  string
	RefreshToken string
	User         *domain.User
}
