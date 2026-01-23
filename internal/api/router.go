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
	userGroup.GET("/:id", c.UserHandler.GetUserByID)
	userGroup.PUT("/:id", c.UserHandler.UpdateUser)

	authGroup := e.Group("/auth")
	authGroup.POST("/sign-up", c.AuthHandler.SignUp)
	authGroup.POST("/login", c.AuthHandler.Login)
	authGroup.POST("/refresh", c.AuthHandler.Refresh)
	authGroup.POST("/logout", c.AuthHandler.Logout)
}
