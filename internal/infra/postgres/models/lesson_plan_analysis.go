package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/misalima/edunex-backend/internal/core/domain"
	"gorm.io/datatypes"
)

type LessonPlanAnalysisModel struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	LessonPlanID   uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:idx_lpa_lesson_plan_id"`
	Title          string         `gorm:"type:text;not null"`
	Subject        string         `gorm:"type:text;not null"`
	GradeLevel     string         `gorm:"type:text;not null"`
	AlignmentScore int            `gorm:"type:integer;not null;check:alignment_score >= 0 AND alignment_score <= 100"`
	Feedback       string         `gorm:"type:text;not null"`
	Metadata       datatypes.JSON `gorm:"type:jsonb;not null"`
	Suggestions    datatypes.JSON `gorm:"type:jsonb;not null"`
	CreatedAt      time.Time      `gorm:"column:created_at;autoCreateTime"`
}

func (LessonPlanAnalysisModel) TableName() string {
	return "lesson_plan_analyses"
}

func (m *LessonPlanAnalysisModel) ToDomain() *domain.LessonPlanAnalysis {
	if m == nil {
		return nil
	}
	return &domain.LessonPlanAnalysis{
		ID:             m.ID,
		LessonPlanID:   m.LessonPlanID,
		Title:          m.Title,
		Subject:        m.Subject,
		GradeLevel:     m.GradeLevel,
		AlignmentScore: m.AlignmentScore,
		Feedback:       m.Feedback,
		Metadata:       string(m.Metadata),
		Suggestions:    string(m.Suggestions),
		CreatedAt:      m.CreatedAt,
	}
}

func FromDomainAnalysis(lpa *domain.LessonPlanAnalysis) *LessonPlanAnalysisModel {
	if lpa == nil {
		return nil
	}
	return &LessonPlanAnalysisModel{
		ID:             lpa.ID,
		LessonPlanID:   lpa.LessonPlanID,
		Title:          lpa.Title,
		Subject:        lpa.Subject,
		GradeLevel:     lpa.GradeLevel,
		AlignmentScore: lpa.AlignmentScore,
		Feedback:       lpa.Feedback,
		Metadata:       datatypes.JSON(lpa.Metadata),
		Suggestions:    datatypes.JSON(lpa.Suggestions),
		CreatedAt:      lpa.CreatedAt,
	}
}
