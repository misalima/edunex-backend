package secondary

import (
	"context"

	"github.com/google/uuid"
	"github.com/misalima/edunex-backend/internal/core/domain"
)

type LessonPlanLoader interface {
	InsertLessonPlan(ctx context.Context, lp *domain.LessonPlan) (uuid.UUID, error)
	GetLessonPlanByID(ctx context.Context, id uuid.UUID) (*domain.LessonPlan, error)
	ListLessonPlans(ctx context.Context, userID uuid.UUID, params domain.PaginationParams) ([]*domain.LessonPlan, int64, error)
	UpdateLessonPlan(ctx context.Context, lp *domain.LessonPlan) error
	DeleteLessonPlan(ctx context.Context, id uuid.UUID) error
}
