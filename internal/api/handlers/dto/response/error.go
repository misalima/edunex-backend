package response

import (
	"errors"

	"github.com/labstack/echo/v4"
	"github.com/misalima/edunex-backend/internal/core/domain_errors"
)

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func JSONError(c echo.Context, err error) error {
	status := domain_errors.MapErrorToStatus(err)
	code := domain_errors.CodeFromError(err)

	msg := err.Error()

	var de *domain_errors.DomainError
	if errors.As(err, &de) {
		msg = de.Message
	}

	return c.JSON(status, ErrorResponse{
		Code:    code,
		Message: msg,
	})
}
