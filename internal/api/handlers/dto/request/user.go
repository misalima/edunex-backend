package request

import (
	"github.com/google/uuid"
)

type UpdateUserRequest struct {
	Name string `json:"name"`
}

type UpdateRoleRequest struct {
	UserID uuid.UUID `json:"user_id" validate:"required"`
	Role   string    `json:"role" validate:"required"`
}
