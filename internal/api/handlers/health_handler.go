package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type HealthHandler struct{}

// HealthResponse represents the response returned by the health endpoint.
type HealthResponse struct {
	Status  string `json:"status" example:"ok"`
	Message string `json:"message" example:"API is running correctly!"`
}

// HealthHandler is a simple handler used to verify the API status.
// @Summary Health check
// @Description Returns the API status.
// @Tags Health
// @Produce json
// @Success 200 {object} HealthResponse
// @Router /health [get]
func (h *HealthHandler) HealthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, HealthResponse{
		Status:  "ok",
		Message: "API is running correctly!",
	})
}
