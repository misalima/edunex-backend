package domain

import (
	"time"

	"github.com/google/uuid"
)

type LessonPlan struct {
	ID         uuid.UUID
	UserID     uuid.UUID
	Title      string
	FilePath   string
	RawContent *string
	Teacher    *string
	Discipline *string
	GradeLevel *GradeLevel
	Status     LessonPlanStatus // pending, approved, or needs_adjustment (pedagogical approval status)
	CreatedAt  time.Time
}
