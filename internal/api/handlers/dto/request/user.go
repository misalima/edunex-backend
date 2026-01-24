package request

import (
	"errors"

	"github.com/misalima/edunex-backend/internal/core/domain"
)

type CreateUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UpdateUserRequest struct {
	Name string `json:"name"`
}

func (req *CreateUserRequest) ToDomain() *domain.User {
	return &domain.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	}
}

func (req *CreateUserRequest) ValidateCreateUserRequest() error {
	if req.Name == "" {
		return errors.New("name is required")
	}
	if req.Email == "" {
		return errors.New("email is required")
	}
	if req.Password == "" {
		return errors.New("password is required")
	}
	return nil
}

type UpdateRoleRequest struct {
	UserID string `json:"user_id" validate:"required"`
	Role   string `json:"role" validate:"required"`
}
