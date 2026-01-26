package container

import (
	"os"
	"sync"

	"github.com/labstack/gommon/log"
	"github.com/misalima/edunex-backend/internal/api/handlers"
	"github.com/misalima/edunex-backend/internal/core/services"
	"github.com/misalima/edunex-backend/internal/infra/postgres"
	"github.com/misalima/edunex-backend/internal/infra/security"
	supabase "github.com/misalima/edunex-backend/internal/infra/storage"
	"gorm.io/gorm"
)

// Container centraliza a criação (lazy) e cache das dependências da aplicação.
// Implementado com sync.Once para inicialização única e thread-safe.
type Container struct {
	db *gorm.DB

	jwtOnce    sync.Once
	jwtService *security.JWTService

	storageOnce   sync.Once
	storageClient *supabase.Client

	userOnce    sync.Once
	userService *services.UserService
	userHandler *handlers.UserHandler

	authOnce    sync.Once
	authHandler *handlers.AuthHandler

	healthOnce    sync.Once
	healthHandler *handlers.HealthHandler

	lessonPlanOnce    sync.Once
	lessonPlanHandler *handlers.LessonPlanHandler
}

func NewContainer(db *gorm.DB) *Container {
	return &Container{db: db}
}

func (c *Container) GetJWTService() *security.JWTService {
	c.jwtOnce.Do(func() {
		secret := os.Getenv("JWT_SECRET")
		if secret == "" {
			log.Info("JWT secret not found, using default secret")
			secret = "default_secret_change_me"
		}
		c.jwtService = security.NewJWTService(secret)
	})
	return c.jwtService
}

func (c *Container) GetStorageClient() *supabase.Client {
	c.storageOnce.Do(func() {
		url := os.Getenv("SUPABASE_URL")
		key := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
		bucket := os.Getenv("SUPABASE_BUCKET")
		if url == "" || key == "" || bucket == "" {
			log.Fatal("supabase configuration is missing (SUPABASE_URL, SUPABASE_SERVICE_ROLE_KEY, SUPABASE_BUCKET)")
		}
		c.storageClient = supabase.NewClient(url, key, bucket)
	})
	return c.storageClient
}

func (c *Container) GetUserHandler() *handlers.UserHandler {
	c.userOnce.Do(func() {
		userRepo := postgres.NewGormUserRepository(c.db)
		c.userService = services.NewUserService(userRepo)
		c.userHandler = handlers.NewUserHandler(c.userService)
	})
	return c.userHandler
}

func (c *Container) GetAuthHandler() *handlers.AuthHandler {
	c.authOnce.Do(func() {
		authRepo := postgres.NewGormAuthRepository(c.db)
		userRepoForAuth := postgres.NewGormUserRepository(c.db)

		jwt := c.GetJWTService()

		authSvc := services.NewAuthService(authRepo, userRepoForAuth, jwt)

		c.GetUserHandler()

		c.authHandler = handlers.NewAuthHandler(authSvc, c.userService)
	})
	return c.authHandler
}

func (c *Container) GetHealthHandler() *handlers.HealthHandler {
	c.healthOnce.Do(func() {
		c.healthHandler = &handlers.HealthHandler{}
	})
	return c.healthHandler
}

func (c *Container) GetLessonPlanHandler() *handlers.LessonPlanHandler {
	c.lessonPlanOnce.Do(func() {
		lpRepo := postgres.NewLessonPlanRepository(c.db)
		storage := c.GetStorageClient()
		lpSvc := services.NewLessonPlanService(lpRepo, storage)
		c.lessonPlanHandler = handlers.NewLessonPlanHandler(lpSvc)
	})
	return c.lessonPlanHandler
}
