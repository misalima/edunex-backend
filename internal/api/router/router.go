package router

import (
	"github.com/labstack/echo/v4"
	"github.com/misalima/edunex-backend/internal/api/container"
	"github.com/misalima/edunex-backend/internal/api/middleware"
)

func RegisterRoutes(e *echo.Echo, c *container.Container) {
	// Rotas públicas
	e.GET("/health", c.GetHealthHandler().HealthHandler)

	v1 := e.Group("/api/v1")
	v1.GET("/me", c.GetUserHandler().GetMe)
	v1.Use(middleware.AuthMiddleware(c.GetJWTManager()))

	// Users
	userGroup := v1.Group("/users")
	userGroup.GET("", c.GetUserHandler().ListUsers)
	userGroup.GET("/:id", c.GetUserHandler().GetUserByID)
	userGroup.PUT("/:id", c.GetUserHandler().UpdateUser)

	// Admin
	adminGroup := v1.Group("/admin")
	adminGroup.Use(middleware.AdminOnly(c.GetUserManager()))
	adminGroup.PATCH("/users/role", c.GetUserHandler().UpdateRole)

	// Lesson Plans
	lPGroup := v1.Group("/lesson-plans")
	lPGroup.POST("", c.GetLessonPlanHandler().Create)
	lPGroup.GET("", c.GetLessonPlanHandler().List)
	lPGroup.GET("/:id", c.GetLessonPlanHandler().GetByID)
}
