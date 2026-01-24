package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func AdminOnly(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		userRole, ok := c.Get("user_role").(string)
		if !ok || userRole != "admin" {
			return c.JSON(http.StatusForbidden, map[string]string{"error": "acesso negado: apenas administradores"})
		}
		return next(c)
	}
}
