package response

import "github.com/misalima/edunex-backend/internal/core/domain"

type UserResponse struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
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
