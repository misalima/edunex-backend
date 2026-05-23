package secondary

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/misalima/edunex-backend/internal/core/domain"
)

// AnalysisJobLoader defines the ports for retrieving, locking, and updating background analysis jobs.
type AnalysisJobLoader interface {
	UpsertAnalysisJob(ctx context.Context, lessonPlanID uuid.UUID) (uuid.UUID, error)
	FetchPendingJob(ctx context.Context) (*domain.AnalysisJob, error)
	FetchPendingJobByID(ctx context.Context, jobID uuid.UUID) (*domain.AnalysisJob, error)
	MarkJobDone(ctx context.Context, jobID uuid.UUID) error
	MarkJobFailed(ctx context.Context, jobID uuid.UUID, errorMsg string, maxAttempts int) error
	GetJobByLessonPlanID(ctx context.Context, lessonPlanID uuid.UUID) (*domain.AnalysisJob, error)
	CleanupStaleProcessingJobs(ctx context.Context, staleThreshold time.Duration) error
	GetJobStatistics(ctx context.Context) (map[string]int64, error)
	SaveAnalysisAndMarkDone(ctx context.Context, analysis *domain.LessonPlanAnalysis, jobID uuid.UUID) error
	GetJobByID(ctx context.Context, jobID uuid.UUID) (*domain.AnalysisJob, error)
}
