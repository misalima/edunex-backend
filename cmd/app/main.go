package main

import (
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/misalima/edunex-backend/internal/api"
	"github.com/misalima/edunex-backend/internal/infra/postgres"
	"log"
)

func main() {
	if err := godotenv.Load(".env"); err != nil {
		log.Println("Aviso: Arquivo .env não encontrado, usando variáveis de ambiente padrão")
	}

	postgres.Connect()
	defer postgres.Close()

	e := echo.New()

	api.RegisterRoutes(e)

	log.Println("Servidor iniciado na porta 8080")
	log.Fatal(e.Start(":8080"))

}
