package container

import (
	"github.com/misalima/edunex-backend/internal/api/handlers"
	"github.com/misalima/edunex-backend/internal/core/services"
	"github.com/misalima/edunex-backend/internal/infra/postgres"
	"gorm.io/gorm"
)

type Container struct {
	UserHandler   *handlers.UserHandler
	HealthHandler *handlers.HealthHandler
}

func NewContainer(db *gorm.DB) *Container {
	userRepo := postgres.NewGormUserRepository(db)
	userSvc := services.NewUserService(userRepo)
	userHandler := handlers.NewUserHandler(userSvc)

	return &Container{
		UserHandler:   userHandler,
		HealthHandler: &handlers.HealthHandler{},
	}
}
