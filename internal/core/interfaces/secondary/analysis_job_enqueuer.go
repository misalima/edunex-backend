package secondary

import (
	"context"

	"github.com/google/uuid"
)

// AnalysisJobEnqueuer represents a port capable of enqueueing lesson plan analysis jobs.
type AnalysisJobEnqueuer interface {
	Enqueue(ctx context.Context, lessonPlanID uuid.UUID) (uuid.UUID, error)
}
