package domain

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID       uuid.UUID
	Name     string
	Email    string
	Password string
	Role     string
	Created  time.Time
	Updated  time.Time
}

func (u *User) IsAdmin() bool {
	return u.Role == "admin"
}

func (u *User) IsPrincipal() bool {
	return u.Role == "principal"
}
