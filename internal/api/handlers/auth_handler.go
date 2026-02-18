package handlers

import (
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/misalima/edunex-backend/internal/api/handlers/dto/request"
	"github.com/misalima/edunex-backend/internal/api/handlers/dto/response"
	"github.com/misalima/edunex-backend/internal/core/interfaces/iservice"
)

type AuthHandler struct {
	authService iservice.AuthManager
	userService iservice.UserManager
}

func NewAuthHandler(authService iservice.AuthManager, userService iservice.UserManager) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		userService: userService,
	}
}

func (h *AuthHandler) Login(c echo.Context) error {
	var req request.LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request payload"})
	}

	loginData, err := h.authService.Login(c.Request().Context(), req.Email, req.Password)
	if err != nil {
		return handleDomainError(c, err)
	}

	cookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    loginData.RefreshToken,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode}

	c.SetCookie(cookie)

	userResponse := response.FromDomainUserToResponse(loginData.User)
	loginResponse := &response.LoginResponse{
		Token: loginData.AccessToken,
		User:  *userResponse,
	}

	return c.JSON(http.StatusOK, loginResponse)
}

func (h *AuthHandler) SignUp(c echo.Context) error {
	var req request.CreateUserRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request payload"})
	}

	if err := req.ValidateCreateUserRequest(); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	user := req.ToDomain()

	createdUser, err := h.userService.CreateUser(c.Request().Context(), user)
	if err != nil {
		return handleDomainError(c, err)
	}

	if createdUser == nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "could not create user"})
	}

	loginData, err := h.authService.Login(c.Request().Context(), createdUser.Email, req.Password)
	if err != nil {
		return handleDomainError(c, err)
	}

	cookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    loginData.RefreshToken,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // true em produção
		SameSite: http.SameSiteStrictMode,
	}
	c.SetCookie(cookie)

	userResponse := response.FromDomainUserToResponse(loginData.User)
	loginResponse := &response.LoginResponse{
		Token: loginData.AccessToken,
		User:  *userResponse,
	}

	return c.JSON(http.StatusCreated, loginResponse)
}

func (h *AuthHandler) Refresh(c echo.Context) error {
	cookie, err := c.Cookie("refresh_token")
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid session"})
	}

	newAccessToken, err := h.authService.RefreshToken(c.Request().Context(), cookie.Value)
	if err != nil {
		return handleDomainError(c, err)
	}

	resp := &response.RefreshTokenResponse{
		Token: newAccessToken,
	}

	return c.JSON(http.StatusOK, resp)
}

func (h *AuthHandler) Logout(c echo.Context) error {
	cookie, err := c.Cookie("refresh_token")
	if err == nil {
		_ = h.authService.Logout(c.Request().Context(), cookie.Value)
	}

	expiredCookie := &http.Cookie{
		Name:     "refresh_token",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(946684800, 0),
		HttpOnly: true,
		Secure:   false, // true em produção
		SameSite: http.SameSiteStrictMode,
	}
	c.SetCookie(expiredCookie)

	return c.NoContent(http.StatusNoContent)
}
