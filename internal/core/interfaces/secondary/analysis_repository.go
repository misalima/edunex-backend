package secondary

import (
	"context"

	"github.com/google/uuid"
	"github.com/misalima/edunex-backend/internal/core/domain"
)

type AnalysisJobRepository interface {
	InsertJob(ctx context.Context, job *domain.AnalysisJob) (uuid.UUID, error)
	GetJobByLessonPlanID(ctx context.Context, lessonPlanID uuid.UUID) (*domain.AnalysisJob, error)
	FetchPendingJobs(ctx context.Context, limit int) ([]*domain.AnalysisJob, error)
	UpdateJob(ctx context.Context, job *domain.AnalysisJob) error
}

type LessonPlanAnalysisRepository interface {
	InsertAnalysis(ctx context.Context, a *domain.LessonPlanAnalysis) (uuid.UUID, error)
	GetAnalysisByLessonPlanID(ctx context.Context, lessonPlanID uuid.UUID) (*domain.LessonPlanAnalysis, error)
}
