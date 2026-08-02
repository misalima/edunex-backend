package secondary

import (
	"context"

	"github.com/google/uuid"
	"github.com/misalima/edunex-backend/internal/core/domain"
)

type LessonPlanAnalysisLoader interface {
	InsertAnalysis(ctx context.Context, analysis *domain.LessonPlanAnalysis) error
	GetAnalysisByLessonPlanID(ctx context.Context, lessonPlanID uuid.UUID) (*domain.LessonPlanAnalysis, error)
}
