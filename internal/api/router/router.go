package router

import (
	"github.com/labstack/echo/v4"
	"github.com/misalima/edunex-backend/internal/api/container"
	"github.com/misalima/edunex-backend/internal/api/middleware"
	echoSwagger "github.com/swaggo/echo-swagger" // echo-swagger middleware (alias)
)

func RegisterRoutes(e *echo.Echo, c *container.Container) {
	// Rota do Swagger UI
	e.GET("/swagger/*", echoSwagger.WrapHandler)

	// Rotas públicas
	e.GET("/health", c.GetHealthHandler().HealthHandler)

	// WebSocket: ticket is a short-lived single-use token (never put the JWT in the URL).
	e.GET("/api/v1/ws/lesson-plans", c.GetWsHandler().ConnectWs)

	v1 := e.Group("/api/v1")
	v1.Use(middleware.AuthMiddleware(c.GetAuthenticator()))
	v1.GET("/me", c.GetUserHandler().GetMe)
	v1.POST("/ws/ticket", c.GetWsHandler().IssueTicket)

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
	lPGroup.DELETE("/:id", c.GetLessonPlanHandler().Delete)
	lPGroup.POST("/:lesson_plan_id/analyze", c.GetAnalysisJobHandler().Analyze)
	lPGroup.GET("/:id/analysis", c.GetLessonPlanHandler().GetAnalysis)

	// Analysis Jobs
	analysisGroup := v1.Group("/analysis-jobs")
	analysisGroup.GET("/:job_id", c.GetAnalysisJobHandler().GetJobStatus)
	analysisGroup.GET("/metrics", c.GetAnalysisJobHandler().GetMetrics)
}
