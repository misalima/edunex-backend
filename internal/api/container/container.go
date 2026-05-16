package container

import (
	"fmt"
	"sync"

	"github.com/misalima/edunex-backend/cmd/app/config"
	"github.com/misalima/edunex-backend/internal/api/handlers"
	"github.com/misalima/edunex-backend/internal/core/interfaces/primary"
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

	userSvcOnce sync.Once
	userService *services.UserService

	userHdlOnce sync.Once
	userHandler *handlers.UserHandler

	healthOnce    sync.Once
	healthHandler *handlers.HealthHandler

	lessonPlanOnce    sync.Once
	lessonPlanHandler *handlers.LessonPlanHandler
}

func NewContainer(db *gorm.DB, cfg *config.Config) *Container {
	return &Container{db: db, cfg: cfg}
}

func (c *Container) GetJWTService() *security.JWTService {
	c.jwtOnce.Do(func() {
		expectedIssuer := fmt.Sprintf("%s/auth/v1", c.cfg.SupabaseURL)
		svc, err := security.NewJWTService(
			c.cfg.SupabaseJWTKX,
			c.cfg.SupabaseJWTKY,
			expectedIssuer,
			c.cfg.SupabaseURL,
			c.cfg.SupabaseAnonKey,
		)
		if err != nil {
			panic(fmt.Sprintf("failed to initialize JWT service: %v", err))
		}
		c.jwtService = svc
	})
	return c.jwtService
}

func (c *Container) GetJWTManager() security.JWTValidator {
	return c.GetJWTService()
}

func (c *Container) GetUserService() *services.UserService {
	c.userSvcOnce.Do(func() {
		userRepo := postgres.NewGormUserRepository(c.db)
		c.userService = services.NewUserService(userRepo)
	})
	return c.userService
}

func (c *Container) GetUserManager() primary.UserManager {
	return c.GetUserService()
}

func (c *Container) GetStorageClient() *supabase.Client {
	c.storageOnce.Do(func() {
		// Use cfg instead of os.Getenv
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
	c.userHdlOnce.Do(func() {
		svc := c.GetUserService()
		jwt := c.GetJWTManager()
		c.userHandler = handlers.NewUserHandler(svc, jwt)
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
