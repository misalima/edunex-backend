package domain

import (
	"time"

	"github.com/google/uuid"
)

type LessonPlanAnalysis struct {
	ID           uuid.UUID
	LessonPlanID uuid.UUID
	AnalysisText *string
	CreatedAt    time.Time
}
