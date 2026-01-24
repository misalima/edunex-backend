package router

import (
	"github.com/labstack/echo/v4"
	"github.com/misalima/edunex-backend/internal/api/container"
	"github.com/misalima/edunex-backend/internal/api/middleware"
)

func RegisterRoutes(e *echo.Echo, c *container.Container) {
	e.GET("/api/health", c.HealthHandler.HealthHandler)
	e.POST("/api/sign-up", c.AuthHandler.SignUp)
	e.POST("/api/login", c.AuthHandler.Login)
	e.POST("/api/refresh", c.AuthHandler.Refresh)

	apiGroup := e.Group("/api")
	apiGroup.Use(middleware.AuthMiddleware(c.JWTService))
	apiGroup.POST("/logout", c.AuthHandler.Logout)

	userGroup := apiGroup.Group("/users")
	userGroup.POST("", c.UserHandler.CreateUser)
	userGroup.GET("", c.UserHandler.ListUsers)
	userGroup.GET("/:id", c.UserHandler.GetUserByID)
	userGroup.PUT("/:id", c.UserHandler.UpdateUser)

	adminGroup := apiGroup.Group("/admin")
	adminGroup.Use(middleware.AdminOnly)
	adminGroup.PATCH("/users/role", c.UserHandler.UpdateRole)

}
