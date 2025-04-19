package domain

import (
	"github.com/google/uuid"
	"time"
)

type User struct {
	ID       uuid.UUID
	Name     string
	Email    string
	Password string
	Created  time.Time
	Updated  time.Time
}
