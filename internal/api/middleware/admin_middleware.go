package middleware

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/misalima/edunex-backend/internal/core/interfaces/primary"
)

func AdminOnly(svc primary.UserManager) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			userIDStr, ok := c.Get("user_id").(string)
			if !ok || userIDStr == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{"error": "user not authenticated"})
			}
			userID, err := uuid.Parse(userIDStr)
			if err != nil {
				return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid user id"})
			}

			ctx := c.Request().Context()
			user, err := svc.GetUserByID(ctx, userID)
			if err != nil {
				return c.JSON(http.StatusForbidden, map[string]string{"error": "failed to fetch user"})
			}

			if !user.IsAdmin() {
				return c.JSON(http.StatusForbidden, map[string]string{"error": "access denied"})
			}
			return next(c)
		}

	}
}
