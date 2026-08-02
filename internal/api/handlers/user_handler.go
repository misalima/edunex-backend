package handlers

import (
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/misalima/edunex-backend/internal/api/handlers/dto/request"
	"github.com/misalima/edunex-backend/internal/api/handlers/dto/response"
	"github.com/misalima/edunex-backend/internal/core/domain"
	"github.com/misalima/edunex-backend/internal/core/interfaces/primary"
)

type UserHandler struct {
	svc primary.UserManager
}

func NewUserHandler(svc primary.UserManager) *UserHandler {
	return &UserHandler{svc: svc}
}

// ListUsers godoc
// @Summary List users
// @Description Returns the list of registered users.
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Success 200 {array} response.UserResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /users [get]
func (u *UserHandler) ListUsers(c echo.Context) error {
	users, err := u.svc.ListUsers(c.Request().Context())
	if err != nil {
		return handleDomainError(c, err)
	}

	resp := response.FromDomainListUserToResponse(users)
	return c.JSON(http.StatusOK, resp)
}

// GetUserByID godoc
// @Summary Get user by ID
// @Description Returns a user by its identifier.
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} response.UserResponse
// @Failure 400 {object} response.ErrorMessageResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /users/{id} [get]
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

// UpdateUser godoc
// @Summary Update user
// @Description Updates the name of a user.
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Param request body request.UpdateUserRequest true "User update payload"
// @Success 204 {string} string "No Content"
// @Failure 400 {object} response.ErrorMessageResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /users/{id} [put]
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

// UpdateRole godoc
// @Summary Update user role
// @Description Updates the role of a user. Intended for admin users.
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body request.UpdateRoleRequest true "Role update payload"
// @Success 200 {object} response.MessageResponse
// @Failure 400 {object} response.ErrorMessageResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 403 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /admin/users/role [patch]
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

// GetMe godoc
// @Summary Get current authenticated user
// @Description Returns the authenticated user's profile.
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.UserResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /me [get]
func (h *UserHandler) GetMe(c echo.Context) error {
	userIDStr, ok := c.Get("user_id").(string)
	if !ok || userIDStr == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "user not authenticated"})
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil || userID == uuid.Nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid user id"})
	}

	user, err := h.svc.GetUserByID(c.Request().Context(), userID)
	if err != nil {
		return handleDomainError(c, err)
	}
	if user == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "user not found"})
	}

	resp := response.FromDomainUserToResponse(user)
	return c.JSON(http.StatusOK, resp)
}

