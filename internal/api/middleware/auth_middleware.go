package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/misalima/edunex-backend/internal/core/interfaces/primary"
)

// AuthMiddleware validates bearer tokens and ensures the user exists locally in Postgres.
func AuthMiddleware(authService primary.Authenticator) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "token not provided"})
			}

			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == "" || tokenString == authHeader {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token format"})
			}

			claims, err := authService.Authenticate(c.Request().Context(), tokenString)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": err.Error()})
			}

			c.Set("user_id", claims.UserID)
			c.Set("user_email", claims.Email)

			return next(c)
		}
	}
}
