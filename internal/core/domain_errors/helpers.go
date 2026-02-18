package domain_errors

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
)

func defaultCodeFromStatus(status int) string {
	switch status {
	case http.StatusBadRequest:
		return "BAD_REQUEST"
	case http.StatusUnauthorized:
		return "UNAUTHORIZED"
	case http.StatusForbidden:
		return "FORBIDDEN"
	case http.StatusNotFound:
		return "NOT_FOUND"
	case http.StatusConflict:
		return "CONFLICT"
	case http.StatusUnprocessableEntity:
		return "UNPROCESSABLE_ENTITY"
	case http.StatusInternalServerError:
		return "INTERNAL_ERROR"
	default:
		text := http.StatusText(status)
		if text == "" {
			return "ERROR"
		}
		return toUpperSnake(text)
	}
}

func toUpperSnake(s string) string {
	s = strings.TrimSpace(s)
	return strings.ToUpper(strings.ReplaceAll(s, " ", "_"))
}

func NewWithStatus(message string, httpStatus int) *DomainError {
	code := defaultCodeFromStatus(httpStatus)
	return New(code, message, httpStatus)
}

func NewUnexpectedMsg(message string) *DomainError {
	return NewWithStatus(message, http.StatusInternalServerError)
}

func NewfWithStatus(httpStatus int, format string, args ...interface{}) *DomainError {
	return NewWithStatus(fmt.Sprintf(format, args...), httpStatus)
}

func NewUnexpectedfMsg(format string, args ...interface{}) *DomainError {
	return NewUnexpectedMsg(fmt.Sprintf(format, args...))
}

func WrapWithStatus(cause error, message string, httpStatus int) *DomainError {
	code := defaultCodeFromStatus(httpStatus)
	return Wrap(code, message, httpStatus, cause)
}

func WrapUnexpectedMsg(cause error, message string) *DomainError {
	return WrapWithStatus(cause, message, http.StatusInternalServerError)
}

func WrapfWithStatus(cause error, httpStatus int, format string, args ...interface{}) *DomainError {
	return WrapWithStatus(cause, fmt.Sprintf(format, args...), httpStatus)
}

func NewBadRequestMsg(message string) *DomainError {
	return NewWithStatus(message, http.StatusBadRequest)
}

func NewUnauthorizedMsg(message string) *DomainError {
	return NewWithStatus(message, http.StatusUnauthorized)
}

func NewForbiddenMsg(message string) *DomainError {
	return NewWithStatus(message, http.StatusForbidden)
}

func NewNotFoundMsg(message string) *DomainError {
	return NewWithStatus(message, http.StatusNotFound)
}

func NewConflictMsg(message string) *DomainError {
	return NewWithStatus(message, http.StatusConflict)
}

// AsDomain tenta extrair *DomainError de um error (nil se não for possível)
func AsDomain(err error) *DomainError {
	var de *DomainError
	if errors.As(err, &de) {
		return de
	}
	return nil
}
