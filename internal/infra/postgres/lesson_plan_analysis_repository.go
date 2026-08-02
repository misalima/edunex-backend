package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/misalima/edunex-backend/internal/core/domain"
	"github.com/misalima/edunex-backend/internal/core/domain_errors"
	"github.com/misalima/edunex-backend/internal/core/interfaces/secondary"
	"github.com/misalima/edunex-backend/internal/infra/logger"
	"github.com/misalima/edunex-backend/internal/infra/postgres/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var _ secondary.LessonPlanAnalysisLoader = (*LessonPlanAnalysisRepository)(nil)

type LessonPlanAnalysisRepository struct {
	db *gorm.DB
}

func NewLessonPlanAnalysisRepository(db *gorm.DB) *LessonPlanAnalysisRepository {
	return &LessonPlanAnalysisRepository{db: db}
}

func (r *LessonPlanAnalysisRepository) InsertAnalysis(ctx context.Context, analysis *domain.LessonPlanAnalysis) error {
	if analysis == nil {
		return domain_errors.NewBadRequestMsg("analysis is required")
	}
	if analysis.LessonPlanID == uuid.Nil {
		return domain_errors.NewBadRequestMsg("lesson_plan_id is required")
	}

	m := models.FromDomainAnalysis(analysis)
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}

	logger.Log.Info("inserting lesson plan analysis",
		zap.String("analysis_id", m.ID.String()),
		zap.String("lesson_plan_id", m.LessonPlanID.String()),
		zap.String("title", m.Title),
		zap.Int("alignment_score", m.AlignmentScore))

	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		logger.Log.Error("failed to insert lesson plan analysis",
			zap.Error(err),
			zap.String("lesson_plan_id", m.LessonPlanID.String()))
		return domain_errors.WrapUnexpectedMsg(err, "failed to insert lesson plan analysis")
	}

	logger.Log.Info("lesson plan analysis inserted",
		zap.String("analysis_id", m.ID.String()),
		zap.String("lesson_plan_id", m.LessonPlanID.String()))

	return nil
}

func (r *LessonPlanAnalysisRepository) GetAnalysisByLessonPlanID(ctx context.Context, lessonPlanID uuid.UUID) (*domain.LessonPlanAnalysis, error) {
	if lessonPlanID == uuid.Nil {
		return nil, domain_errors.NewBadRequestMsg("lesson_plan_id is required")
	}

	logger.Log.Debug("fetching lesson plan analysis", zap.String("lesson_plan_id", lessonPlanID.String()))

	m := &models.LessonPlanAnalysisModel{}
	if err := r.db.WithContext(ctx).First(m, "lesson_plan_id = ?", lessonPlanID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Log.Debug("lesson plan analysis not found", zap.String("lesson_plan_id", lessonPlanID.String()))
			return nil, domain_errors.NewNotFoundMsg("lesson plan analysis not found")
		}
		logger.Log.Error("failed to fetch lesson plan analysis",
			zap.Error(err),
			zap.String("lesson_plan_id", lessonPlanID.String()))
		return nil, domain_errors.WrapUnexpectedMsg(err, "failed to fetch lesson plan analysis")
	}

	logger.Log.Debug("lesson plan analysis fetched",
		zap.String("analysis_id", m.ID.String()),
		zap.String("lesson_plan_id", m.LessonPlanID.String()))

	return m.ToDomain(), nil
}

