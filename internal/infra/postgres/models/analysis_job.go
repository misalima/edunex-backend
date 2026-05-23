package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/misalima/edunex-backend/internal/core/domain"
)

type AnalysisJobModel struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey"`
	LessonPlanID uuid.UUID `gorm:"type:uuid;uniqueIndex"`
	Status       string    `gorm:"index"`
	Attempts     int
	ErrorMessage string
	CreatedAt    time.Time
	StartedAt    *time.Time
	FinishedAt   *time.Time
}

// TableName specifies the table name for GORM
func (AnalysisJobModel) TableName() string {
	return "analysis_jobs"
}

func (m *AnalysisJobModel) ToDomain() *domain.AnalysisJob {
	if m == nil {
		return nil
	}
	return &domain.AnalysisJob{
		ID:           m.ID,
		LessonPlanID: m.LessonPlanID,
		Status:       m.Status,
		Attempts:     m.Attempts,
		ErrorMessage: m.ErrorMessage,
		CreatedAt:    m.CreatedAt,
		StartedAt:    m.StartedAt,
		FinishedAt:   m.FinishedAt,
	}
}
