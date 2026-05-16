package request

import (
	"github.com/google/uuid"
	"github.com/misalima/edunex-backend/internal/core/domain"
)

type CreateLessonPlanRequest struct {
	Title      string `json:"title" form:"title" example:"Plano de aula de Matemática"`
	Teacher    string `json:"teacher" form:"teacher" example:"Ana Paula"`
	Discipline string `json:"discipline" form:"discipline" example:"Matemática"`
	GradeLevel string `json:"grade_level" form:"grade_level" example:"1ª SÉRIE"`
}

func (r *CreateLessonPlanRequest) ToDomain(userID uuid.UUID) *domain.LessonPlan {
	return &domain.LessonPlan{
		UserID:     userID,
		Title:      r.Title,
		Teacher:    nullableString(r.Teacher),
		GradeLevel: nullableGradeLevel(r.GradeLevel),
		Discipline: nullableString(r.Discipline),
		Status:     domain.LessonPlanPending,
	}
}

func nullableString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func nullableGradeLevel(s string) *domain.GradeLevel {
	if s == "" {
		return nil
	}
	gl := domain.GradeLevel(s)
	return &gl
}
