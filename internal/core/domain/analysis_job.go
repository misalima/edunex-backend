package domain

import (
	"time"

	"github.com/google/uuid"
)

type AnalysisJob struct {
	ID           uuid.UUID
	LessonPlanID uuid.UUID
	Status       ProcessingStatus
	Attempts     int
	ErrorMessage *string
	CreatedAt    time.Time
	StartedAt    *time.Time
	FinishedAt   *time.Time
}
