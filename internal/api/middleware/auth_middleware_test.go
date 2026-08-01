package middleware_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/misalima/edunex-backend/internal/api/middleware"
	"github.com/misalima/edunex-backend/internal/core/util"
)

type mockAuthenticator struct {
	authenticateFn func(ctx context.Context, token string) (*util.TokenClaims, error)
}

func (m *mockAuthenticator) Authenticate(ctx context.Context, token string) (*util.TokenClaims, error) {
	if m.authenticateFn != nil {
		return m.authenticateFn(ctx, token)
	}
	return nil, errors.New("not implemented")
}

func TestAuthMiddleware_MissingToken(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	authSvc := &mockAuthenticator{}
	mw := middleware.AuthMiddleware(authSvc)

	handler := mw(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
	if rec.Body.String() != "{\"error\":\"token not provided\"}\n" {
		t.Errorf("unexpected body: %s", rec.Body.String())
	}
}

func TestAuthMiddleware_InvalidTokenFormat(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	req.Header.Set("Authorization", "InvalidFormat123")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	authSvc := &mockAuthenticator{}
	mw := middleware.AuthMiddleware(authSvc)

	handler := mw(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
	if rec.Body.String() != "{\"error\":\"invalid token format\"}\n" {
		t.Errorf("unexpected body: %s", rec.Body.String())
	}
}

func TestAuthMiddleware_AuthFailure(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-jwt")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	authSvc := &mockAuthenticator{
		authenticateFn: func(ctx context.Context, token string) (*util.TokenClaims, error) {
			return nil, errors.New("signature expired")
		},
	}
	mw := middleware.AuthMiddleware(authSvc)

	handler := mw(func(c echo.Context) error {
		return c.NoContent(http.StatusOK)
	})

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
	if rec.Body.String() != "{\"error\":\"signature expired\"}\n" {
		t.Errorf("unexpected body: %s", rec.Body.String())
	}
}

func TestAuthMiddleware_Success(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/test", nil)
	req.Header.Set("Authorization", "Bearer valid-jwt-token")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	var receivedUserID, receivedEmail string

	authSvc := &mockAuthenticator{
		authenticateFn: func(ctx context.Context, token string) (*util.TokenClaims, error) {
			if token != "valid-jwt-token" {
				t.Errorf("expected token 'valid-jwt-token', got '%s'", token)
			}
			return &util.TokenClaims{
				UserID: "550e8400-e29b-41d4-a716-446655440000",
				Email:  "prof@edunex.com",
			}, nil
		},
	}
	mw := middleware.AuthMiddleware(authSvc)

	handler := mw(func(c echo.Context) error {
		receivedUserID = c.Get("user_id").(string)
		receivedEmail = c.Get("user_email").(string)
		return c.NoContent(http.StatusOK)
	})

	err := handler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if receivedUserID != "550e8400-e29b-41d4-a716-446655440000" {
		t.Errorf("expected user_id '550e8400-e29b-41d4-a716-446655440000', got '%s'", receivedUserID)
	}
	if receivedEmail != "prof@edunex.com" {
		t.Errorf("expected user_email 'prof@edunex.com', got '%s'", receivedEmail)
	}
}
