package domain_errors

import "net/http"

var (
	ErrUserNotFound        = New("USER_NOT_FOUND", "user not found", http.StatusNotFound)
	ErrLessonPlanNotFound  = New("LESSON_PLAN_NOT_FOUND", "lesson plan not found", http.StatusNotFound)
	ErrInvalidCredentials  = New("INVALID_CREDENTIALS", "invalid credentials", http.StatusUnauthorized)
	ErrUnauthorized        = New("UNAUTHORIZED", "unauthorized access", http.StatusForbidden)
	ErrInvalidToken        = New("INVALID_TOKEN", "invalid token", http.StatusUnauthorized)
	ErrRefreshTokenExpired = New("REFRESH_TOKEN_EXPIRED", "refresh token expired", http.StatusUnauthorized)
	ErrBadRequest          = New("BAD_REQUEST", "invalid request", http.StatusBadRequest)
	ErrConflict            = New("CONFLICT", "data conflict", http.StatusConflict)
	ErrNotFound            = New("NOT_FOUND", "resource not found", http.StatusNotFound)
	ErrInternal            = New("INTERNAL_ERROR", "internal server error", http.StatusInternalServerError)
)
