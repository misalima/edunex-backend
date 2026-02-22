package container

import (
	"fmt"
	"sync"

	"github.com/misalima/edunex-backend/cmd/app/config"
	"github.com/misalima/edunex-backend/internal/api/handlers"
	"github.com/misalima/edunex-backend/internal/core/interfaces/iservice"
	"github.com/misalima/edunex-backend/internal/core/services"
	"github.com/misalima/edunex-backend/internal/infra/postgres"
	"github.com/misalima/edunex-backend/internal/infra/security"
	supabase "github.com/misalima/edunex-backend/internal/infra/storage"
	"gorm.io/gorm"
)

// Container centraliza a criação (lazy) e cache das dependências da aplicação.
type Container struct {
	db  *gorm.DB
	cfg *config.Config

	jwtOnce    sync.Once
	jwtService *security.JWTService

	storageOnce   sync.Once
	storageClient *supabase.Client

	userOnce    sync.Once
	userService *services.UserService
	userHandler *handlers.UserHandler

	healthOnce    sync.Once
	healthHandler *handlers.HealthHandler

	lessonPlanOnce    sync.Once
	lessonPlanHandler *handlers.LessonPlanHandler
}

func NewContainer(db *gorm.DB, cfg *config.Config) *Container {
	return &Container{db: db, cfg: cfg}
}

func (c *Container) GetJWTService() iservice.JWTManager {
	c.jwtOnce.Do(func() {
		expectedIssuer := fmt.Sprintf("%s/auth/v1", c.cfg.SupabaseURL)
		c.jwtService = security.NewJWTService(
			c.cfg.SupabaseJWTSecret,
			expectedIssuer,
			c.cfg.SupabaseURL,
			c.cfg.SupabaseAnonKey,
		)
	})
	return c.jwtService
}

func (c *Container) GetStorageClient() *supabase.Client {
	c.storageOnce.Do(func() {
		// Usa o cfg em vez de os.Getenv
		if c.cfg.SupabaseURL == "" || c.cfg.SupabaseServiceKey == "" || c.cfg.SupabaseBucket == "" {
			panic("supabase configuration is missing (SUPABASE_URL, SUPABASE_SERVICE_ROLE_KEY, SUPABASE_BUCKET)")
		}
		c.storageClient = supabase.NewClient(
			c.cfg.SupabaseURL,
			c.cfg.SupabaseServiceKey,
			c.cfg.SupabaseBucket,
		)
	})
	return c.storageClient
}

func (c *Container) GetUserHandler() *handlers.UserHandler {
	c.userOnce.Do(func() {
		userRepo := postgres.NewGormUserRepository(c.db)
		c.userService = services.NewUserService(userRepo)

		c.userHandler = handlers.NewUserHandler(c.userService, c.GetJWTService())
	})
	return c.userHandler
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
