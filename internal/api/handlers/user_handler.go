package handlers

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/misalima/edunex-backend/internal/api/handlers/dto/request"
	"github.com/misalima/edunex-backend/internal/api/handlers/dto/response"
	"github.com/misalima/edunex-backend/internal/api/util"
	"github.com/misalima/edunex-backend/internal/core/domain"
	"github.com/misalima/edunex-backend/internal/core/interfaces/iservice"
	"github.com/misalima/edunex-backend/internal/infra/security"
)

type UserHandler struct {
	svc    iservice.UserManager
	jwtSvc security.JWTManager
}

func NewUserHandler(svc iservice.UserManager, jwtSvc security.JWTManager) *UserHandler {
	return &UserHandler{svc: svc, jwtSvc: jwtSvc}
}

func (u *UserHandler) ListUsers(c echo.Context) error {
	users, err := u.svc.ListUsers(c.Request().Context())
	if err != nil {
		return handleDomainError(c, err)
	}

	resp := response.FromDomainListUserToResponse(users)
	return c.JSON(http.StatusOK, resp)
}

func (u *UserHandler) GetUserByID(c echo.Context) error {

	idParam := strings.TrimSpace(c.Param("id"))
	id, err := uuid.Parse(idParam)
	if err != nil {
		log.Errorf("Error parsing user id %s: %s", id, err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid user id"})
	}

	user, err := u.svc.GetUserByID(c.Request().Context(), id)
	if err != nil {
		return handleDomainError(c, err)
	}
	if user == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "user not found"})
	}

	resp := response.FromDomainUserToResponse(user)
	return c.JSON(http.StatusOK, resp)
}

func (u *UserHandler) UpdateUser(c echo.Context) error {
	idParam := c.Param("id")
	id, err := uuid.Parse(idParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid user id"})
	}

	var req request.UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		log.Error(err)
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request payload"})
	}

	user := domain.User{
		ID:   id,
		Name: req.Name,
	}

	err = u.svc.UpdateUser(c.Request().Context(), &user)
	if err != nil {
		return handleDomainError(c, err)
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *UserHandler) UpdateRole(c echo.Context) error {
	var req request.UpdateRoleRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request payload"})
	}

	if req.UserID == uuid.Nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid user id"})
	}

	err := h.svc.UpdateUserRole(c.Request().Context(), req.UserID, req.Role)
	if err != nil {
		return handleDomainError(c, err)
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "user role updated successfully"})
}

func (h *UserHandler) GetMe(c echo.Context) error {
	tokenStr := util.ExtractBearerToken(c)

	claims, err := h.jwtSvc.ValidateTokenViaAPI(tokenStr)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}
	userEmail := claims.Email
	if userID == uuid.Nil || userEmail == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
	}

	user, err := h.svc.GetOrCreateUserFromSupabaseID(c.Request().Context(), userID, userEmail)
	if err != nil {
		return handleDomainError(c, err)
	}

	resp := response.FromDomainUserToResponse(user)
	return c.JSON(http.StatusOK, resp)
}
