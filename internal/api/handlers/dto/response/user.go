package response

import "github.com/misalima/edunex-backend/internal/core/domain"

type UserResponse struct {
	ID        string `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name      string `json:"name" example:"Maria Silva"`
	Email     string `json:"email" example:"maria@edunex.example"`
	Role      string `json:"role" example:"teacher"`
	CreatedAt string `json:"created_at" example:"2026-05-16T15:30:00Z"`
	UpdatedAt string `json:"updated_at" example:"2026-05-16T15:30:00Z"`
}

func FromDomainUserToResponse(user *domain.User) *UserResponse {
	return &UserResponse{
		ID:        user.ID.String(),
		Name:      user.Name,
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.Created.String(),
		UpdatedAt: user.Updated.String(),
	}
}

func FromDomainListUserToResponse(users []*domain.User) []*UserResponse {
	response := make([]*UserResponse, len(users))
	for i, user := range users {
		response[i] = FromDomainUserToResponse(user)
	}
	return response
}
