package api

import (
	"github.com/labstack/echo/v4"
	"github.com/misalima/edunex-backend/internal/api/handlers"
)

func RegisterRoutes(e *echo.Echo) {
	e.GET("/health", handlers.HealthHandler)
}
