package secondary

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/misalima/edunex-backend/internal/core/domain"
)

type AnalysisJobRepository interface {
	CreateJob(ctx context.Context, job *domain.AnalysisJob) (*domain.AnalysisJob, error)
	GetJobByLessonPlanID(ctx context.Context, lessonPlanID uuid.UUID) (*domain.AnalysisJob, error)
	ClaimPendingJobs(ctx context.Context, limit int) ([]*domain.AnalysisJob, error)
	MarkJobProcessing(ctx context.Context, jobID uuid.UUID, startedAt time.Time) error
	MarkJobCompleted(ctx context.Context, jobID uuid.UUID, result *domain.LessonPlanAnalysis, finishedAt time.Time) error
	MarkJobFailed(ctx context.Context, jobID uuid.UUID, reason string, finishedAt time.Time) error
}

type LessonPlanAnalysisRepository interface {
	SaveAnalysisResult(ctx context.Context, a *domain.LessonPlanAnalysis) (*domain.LessonPlanAnalysis, error)
	GetAnalysisByLessonPlanID(ctx context.Context, lessonPlanID uuid.UUID) (*domain.LessonPlanAnalysis, error)
}
