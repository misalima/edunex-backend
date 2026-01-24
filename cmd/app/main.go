package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/misalima/edunex-backend/cmd/app/config"
	"github.com/misalima/edunex-backend/internal/api/container"
	"github.com/misalima/edunex-backend/internal/api/router"
	"github.com/misalima/edunex-backend/internal/infra/postgres"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("Warning: .env file not found. Using environment variables.")
	}

	cfg := config.Load()

	db, err := postgres.InitDB(cfg.DBURL)
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatalf("failed to get sql.DB from gorm.DB: %v", err)
	}

	e := echo.New()
	setupMiddleware(e)

	ctn := container.NewContainer(db)
	router.RegisterRoutes(e, ctn)

	log.Printf("Server starting at port %s", cfg.Port)

	go func() {
		if err := e.Start(":" + cfg.Port); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("error starting server:", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Print("Starting graceful shutdown...")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatalf("error shutting down the server: %v", err)
	}

	if err := sqlDB.Close(); err != nil {
		e.Logger.Errorf("error closing database connections: %v", err)
	}

	log.Print("Server shut down successfully")
}

func setupMiddleware(e *echo.Echo) {
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
}
