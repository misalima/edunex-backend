package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/misalima/edunex-backend/internal/api"
	"github.com/misalima/edunex-backend/internal/api/container"
	"github.com/misalima/edunex-backend/internal/infra/postgres"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("Aviso: Arquivo .env não encontrado, usando variáveis de ambiente padrão")
	}

	// Build DSN from environment if not provided
	dsn := os.Getenv("DATABASE_DSN")
	if dsn == "" {
		dsn = fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASSWORD"),
			os.Getenv("DB_HOST"),
			os.Getenv("DB_PORT"),
			os.Getenv("DB_NAME"),
		)
	}

	// Initialize GORM DB using the postgres package (GORM initializer)
	db, err := postgres.InitDB(dsn)
	if err != nil {
		log.Fatalf("failed to initialize database: %v", err)
	}

	// Initialize the API container with dependencies
	ctn := container.NewContainer(db)
	e := echo.New()
	api.RegisterRoutes(e, ctn)

	e.Use(middleware.Logger())

	log.Println("Servidor iniciado na porta 8080")
	log.Fatal(e.Start(":8080"))
}
