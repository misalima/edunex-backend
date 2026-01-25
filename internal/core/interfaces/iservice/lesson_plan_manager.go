package iservice

import (
	"context"
	"io"

	"github.com/google/uuid"
	"github.com/misalima/edunex-backend/internal/core/domain"
)

type LessonPlanManager interface {
	CreateLessonPlan(ctx context.Context, lp *domain.LessonPlan, fileReader io.Reader, filename, contentType string) (*domain.LessonPlan, error)
	GetLessonPlanWithSignedURL(ctx context.Context, id uuid.UUID) (*domain.LessonPlan, string, error)
	ListLessonPlansWithSignedURLs(ctx context.Context) ([]*domain.LessonPlan, map[string]string, error)
}
