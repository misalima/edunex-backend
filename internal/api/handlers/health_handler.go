package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// HealthHandler é um handler simples para verificar o status da API
func HealthHandler(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{
		"status":  "ok",
		"message": "API está funcionando corretamente!",
	})
}
