package handlers

import (
	"errors"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/misalima/edunex-backend/internal/api/handlers/dto/response"
	"github.com/misalima/edunex-backend/internal/core/domain_errors"
)

func getUserIDFromContext(c echo.Context) (uuid.UUID, error) {
	val := c.Get("user_id")
	if id, ok := val.(uuid.UUID); ok {
		return id, nil
	}
	if s, ok := val.(string); ok {
		return uuid.Parse(s)
	}
	return uuid.Nil, errors.New("user_id not found in context")
}

func getUserEmailFromContext(c echo.Context) (string, error) {
	val := c.Get("user_email")
	if email, ok := val.(string); ok {
		return email, nil
	}
	return "", errors.New("user_email not found in context")
}

func handleDomainError(c echo.Context, err error) error {
	status := domain_errors.MapErrorToStatus(err)
	code := domain_errors.CodeFromError(err)

	msg := err.Error()

	var de *domain_errors.DomainError
	if errors.As(err, &de) {
		msg = de.Message
	}

	return c.JSON(status, response.ErrorResponse{
		Code:    code,
		Message: msg,
	})
}
