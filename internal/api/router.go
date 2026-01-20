package api

import (
	"github.com/labstack/echo/v4"
	"github.com/misalima/edunex-backend/internal/api/container"
)

func RegisterRoutes(e *echo.Echo, c *container.Container) {
	e.GET("/health", c.HealthHandler.HealthHandler)

	userGroup := e.Group("/users")
	userGroup.POST("", c.UserHandler.CreateUser)
	userGroup.GET("", c.UserHandler.ListUsers)
}
