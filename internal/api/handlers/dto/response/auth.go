package response

type LoginResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

type RefreshTokenResponse struct {
	Token string `json:"token"`
}
