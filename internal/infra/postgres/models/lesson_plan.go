package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/misalima/edunex-backend/internal/core/domain"
)

type LessonPlanModel struct {
	ID         uuid.UUID               `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID     uuid.UUID               `gorm:"type:uuid;not null;index:idx_lesson_plans_user_id"`
	Title      string                  `gorm:"type:text;not null"`
	FilePath   string                  `gorm:"type:text;not null"`
	RawContent *string                 `gorm:"type:text"`
	Teacher    *string                 `gorm:"type:text"`
	Discipline *string                 `gorm:"type:text"`
	GradeLevel *domain.GradeLevel      `gorm:"type:varchar(50)"`
	Status     domain.LessonPlanStatus `gorm:"type:lesson_plan_status;not null;default:'pending'"`
	CreatedAt  time.Time               `gorm:"column:created_at;autoCreateTime"`
}

func (LessonPlanModel) TableName() string {
	return "lesson_plans"
}

func (m *LessonPlanModel) ToDomain() *domain.LessonPlan {
	if m == nil {
		return nil
	}
	return &domain.LessonPlan{
		ID:         m.ID,
		UserID:     m.UserID,
		Title:      m.Title,
		FilePath:   m.FilePath,
		RawContent: m.RawContent,
		GradeLevel: m.GradeLevel,
		Teacher:    m.Teacher,
		Discipline: m.Discipline,
		Status:     m.Status,
		CreatedAt:  m.CreatedAt,
	}
}

func FromDomainLessonPlan(lp *domain.LessonPlan) *LessonPlanModel {
	if lp == nil {
		return nil
	}
	m := &LessonPlanModel{
		UserID:     lp.UserID,
		Title:      lp.Title,
		FilePath:   lp.FilePath,
		RawContent: lp.RawContent,
		GradeLevel: lp.GradeLevel,
		Teacher:    lp.Teacher,
		Discipline: lp.Discipline,
		Status:     lp.Status,
	}
	if lp.ID != uuid.Nil {
		m.ID = lp.ID
	}
	if !lp.CreatedAt.IsZero() {
		m.CreatedAt = lp.CreatedAt
	}
	return m
}
