package domain

import (
	"time"

	"github.com/google/uuid"
)

type LessonPlan struct {
	ID       uuid.UUID
	UserID   uuid.UUID
	Title    string
	FilePath string
	Status   string
	Created  time.Time
}
