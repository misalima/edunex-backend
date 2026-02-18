package domain

import (
	"time"

	"github.com/google/uuid"
)

type LessonPlanAnalysis struct {
	ID           uuid.UUID
	LessonPlanID uuid.UUID
	AnalysisText *string
	Status       string
	CreatedAt    time.Time
}
