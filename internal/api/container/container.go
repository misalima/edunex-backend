package container

import (
	"os"

	"github.com/labstack/gommon/log"
	"github.com/misalima/edunex-backend/internal/api/handlers"
	"github.com/misalima/edunex-backend/internal/core/services"
	"github.com/misalima/edunex-backend/internal/infra/postgres"
	"github.com/misalima/edunex-backend/internal/infra/security"
	"gorm.io/gorm"
)

type Container struct {
	UserHandler   *handlers.UserHandler
	HealthHandler *handlers.HealthHandler
	AuthHandler   *handlers.AuthHandler
}

func NewContainer(db *gorm.DB) *Container {
	userRepo := postgres.NewGormUserRepository(db)
	userSvc := services.NewUserService(userRepo)
	userHandler := handlers.NewUserHandler(userSvc)

	authRepo := postgres.NewGormAuthRepository(db)
	jwtSvc := security.NewJWTService(getJWTSecret())
	authSvc := services.NewAuthService(authRepo, userRepo, jwtSvc)
	authHandler := handlers.NewAuthHandler(authSvc, userSvc)

	return &Container{
		UserHandler:   userHandler,
		HealthHandler: &handlers.HealthHandler{},
		AuthHandler:   authHandler,
	}
}

func getJWTSecret() string {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Info("JWT secret not found, using default secret")
		secret = "secret"
	}
	return secret
}
