package domain_errors

import (
	"errors"
	"fmt"
	"net/http"
)

type DomainError struct {
	Code       string
	Message    string
	HTTPStatus int
	Cause      error
}

func (e *DomainError) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s", e.Message, e.Cause.Error())
	}
	return e.Message
}

func (e *DomainError) Unwrap() error {
	return e.Cause
}

func (e *DomainError) Is(target error) bool {
	var t *DomainError
	if errors.As(target, &t) {
		return e.Code == t.Code
	}
	return false
}

func New(code, message string, httpStatus int) *DomainError {
	return &DomainError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
	}
}

func Wrap(code, message string, httpStatus int, cause error) *DomainError {
	return &DomainError{
		Code:       code,
		Message:    message,
		HTTPStatus: httpStatus,
		Cause:      cause,
	}
}

func CodeFromError(err error) string {
	var de *DomainError
	if errors.As(err, &de) {
		return de.Code
	}
	return "INTERNAL_ERROR"
}

func MapErrorToStatus(err error) int {
	var de *DomainError
	if errors.As(err, &de) && de.HTTPStatus != 0 {
		return de.HTTPStatus
	}

	if errors.Is(err, ErrInvalidCredentials) ||
		errors.Is(err, ErrInvalidToken) ||
		errors.Is(err, ErrRefreshTokenExpired) {
		return http.StatusUnauthorized
	}

	if errors.Is(err, ErrConflict) {
		return http.StatusConflict
	}

	if errors.Is(err, ErrBadRequest) {
		return http.StatusBadRequest
	}

	if errors.Is(err, ErrNotFound) || errors.Is(err, ErrUserNotFound) {
		return http.StatusNotFound
	}

	return http.StatusInternalServerError
}
