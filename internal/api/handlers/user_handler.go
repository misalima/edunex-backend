package handlers

import "github.com/misalima/edunex-backend/internal/core/interfaces/iservice"

type UserHandler struct {
	svc iservice.UserManager
}

func NewUserHandler(svc iservice.UserManager) *UserHandler {
	return &UserHandler{svc: svc}
}
