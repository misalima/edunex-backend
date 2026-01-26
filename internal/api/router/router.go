package router

import (
	"github.com/labstack/echo/v4"
	"github.com/misalima/edunex-backend/internal/api/container"
	"github.com/misalima/edunex-backend/internal/api/middleware"
)

func RegisterRoutes(e *echo.Echo, c *container.Container) {
	// Rotas públicas
	e.GET("/api/health", c.GetHealthHandler().HealthHandler)
	e.POST("/api/sign-up", c.GetAuthHandler().SignUp)
	e.POST("/api/login", c.GetAuthHandler().Login)
	e.POST("/api/refresh", c.GetAuthHandler().Refresh)

	apiGroup := e.Group("/api")
	apiGroup.Use(middleware.AuthMiddleware(c.GetJWTService()))
	apiGroup.POST("/logout", c.GetAuthHandler().Logout)

	// Users
	userGroup := apiGroup.Group("/users")
	userGroup.POST("", c.GetUserHandler().CreateUser)
	userGroup.GET("", c.GetUserHandler().ListUsers)
	userGroup.GET("/:id", c.GetUserHandler().GetUserByID)
	userGroup.PUT("/:id", c.GetUserHandler().UpdateUser)

	// Admin
	adminGroup := apiGroup.Group("/admin")
	adminGroup.Use(middleware.AdminOnly)
	adminGroup.PATCH("/users/role", c.GetUserHandler().UpdateRole)

	// Lesson Plans
	lPGroup := apiGroup.Group("/lesson-plans")
	lPGroup.POST("", c.GetLessonPlanHandler().Create)
	lPGroup.GET("", c.GetLessonPlanHandler().List)
	lPGroup.GET("/:id", c.GetLessonPlanHandler().GetByID)
}
