package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
	"github.com/misalima/edunex-backend/internal/api/handlers/dto/request"
	"github.com/misalima/edunex-backend/internal/api/handlers/dto/response"
	"github.com/misalima/edunex-backend/internal/core/domain"
	"github.com/misalima/edunex-backend/internal/core/interfaces/iservice"
)

type UserHandler struct {
	svc iservice.UserManager
}

func NewUserHandler(svc iservice.UserManager) *UserHandler {
	return &UserHandler{svc: svc}
}

func (u *UserHandler) CreateUser(c echo.Context) error {
	var req request.CreateUserRequest

	if err := c.Bind(&req); err != nil {
		log.Error(err)
		return c.JSON(http.StatusBadRequest, err)
	}

	if err := req.ValidateCreateUserRequest(); err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	user := domain.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	}

	createdUser, err := u.svc.CreateUser(c.Request().Context(), &user)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	userResponse := response.FromDomainUserToResponse(createdUser)

	return c.JSON(http.StatusCreated, userResponse)
}

func (u *UserHandler) ListUsers(c echo.Context) error {
	users, err := u.svc.ListUsers(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, err)
	}

	resp := response.FromDomainListUserToResponse(users)
	return c.JSON(http.StatusOK, resp)
}
