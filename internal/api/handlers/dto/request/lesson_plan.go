package request

import (
	"time"

	"github.com/google/uuid"
	"github.com/misalima/edunex-backend/internal/core/domain"
)

type CreateLessonPlanRequest struct {
	Title string `json:"title" form:"title"`
}

func (r *CreateLessonPlanRequest) ToDomain(userID uuid.UUID) *domain.LessonPlan {
	return &domain.LessonPlan{
		ID:       uuid.Nil,
		UserID:   userID,
		Title:    r.Title,
		FilePath: "",
		Status:   "pending",
		Created:  time.Time{},
	}
}
