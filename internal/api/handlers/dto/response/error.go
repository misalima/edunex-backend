package response

type ErrorResponse struct {
	Code    string `json:"code" example:"NOT_FOUND"`
	Message string `json:"message" example:"user not found"`
}
