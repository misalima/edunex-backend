package domain_errors

import "net/http"

var (
	ErrUserNotFound        = New("USER_NOT_FOUND", "usuário não encontrado", http.StatusNotFound)
	ErrInvalidCredentials  = New("INVALID_CREDENTIALS", "credenciais inválidas", http.StatusUnauthorized)
	ErrUnauthorized        = New("UNAUTHORIZED", "acesso não autorizado", http.StatusForbidden)
	ErrInvalidToken        = New("INVALID_TOKEN", "token inválido", http.StatusUnauthorized)
	ErrRefreshTokenExpired = New("REFRESH_TOKEN_EXPIRED", "refresh token expirado", http.StatusUnauthorized)
	ErrBadRequest          = New("BAD_REQUEST", "requisição inválida", http.StatusBadRequest)
	ErrConflict            = New("CONFLICT", "conflito de dados", http.StatusConflict)
	ErrNotFound            = New("NOT_FOUND", "recurso não encontrado", http.StatusNotFound)
	ErrInternal            = New("INTERNAL_ERROR", "erro interno do servidor", http.StatusInternalServerError)
)
