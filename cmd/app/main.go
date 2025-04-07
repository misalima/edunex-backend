package main

import (
	"github.com/misalima/edunex-backend/internal/api"
	"log"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/misalima/edunex-backend/internal/infra/postgres"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("Aviso: Arquivo .env não encontrado, usando variáveis de ambiente padrão")
	}

	db.Connect()
	defer db.Close()

	e := echo.New()

	api.RegisterRoutes(e)

	log.Println("Servidor iniciado na porta 8080")
	log.Fatal(e.Start(":8080"))
}
