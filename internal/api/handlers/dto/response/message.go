package response

type MessageResponse struct {
	Message string `json:"message" example:"user role updated successfully"`
}

type ErrorMessageResponse struct {
	Error string `json:"error" example:"invalid request payload"`
}
