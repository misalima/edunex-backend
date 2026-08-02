package request

import (
	"github.com/google/uuid"
)

type UpdateUserRequest struct {
	Name string `json:"name" example:"Maria Silva"`
}

type UpdateRoleRequest struct {
	UserID uuid.UUID `json:"user_id" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	Role   string    `json:"role" validate:"required" example:"teacher"`
}
