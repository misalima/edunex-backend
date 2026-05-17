package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/misalima/edunex-backend/internal/core/domain"
	"github.com/misalima/edunex-backend/internal/core/domain_errors"
	"github.com/misalima/edunex-backend/internal/core/interfaces/secondary"
	"github.com/misalima/edunex-backend/internal/infra/logger"
	"go.uber.org/zap"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

var _ secondary.LessonPlanAnalysisLoader = (*LessonPlanAnalysisRepository)(nil)

type lessonPlanAnalysisModel struct {
	ID             uuid.UUID      `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	LessonPlanID   uuid.UUID      `gorm:"type:uuid;not null;uniqueIndex:idx_lpa_lesson_plan_id"`
	Title          string         `gorm:"type:text;not null"`
	Subject        string         `gorm:"type:text;not null"`
	GradeLevel     string         `gorm:"type:text;not null"`
	AlignmentScore int            `gorm:"type:integer;not null;check:alignment_score >= 0 AND alignment_score <= 100"`
	Feedback       string         `gorm:"type:text;not null"`
	Metadata       datatypes.JSON `gorm:"type:jsonb;not null"`
	Suggestions    datatypes.JSON `gorm:"type:jsonb;not null"`
	CreatedAt      time.Time      `gorm:"column:created_at;autoCreateTime"`
}

func (lessonPlanAnalysisModel) TableName() string {
	return "lesson_plan_analyses"
}

func (m *lessonPlanAnalysisModel) toDomain() *domain.LessonPlanAnalysis {
	if m == nil {
		return nil
	}
	return &domain.LessonPlanAnalysis{
		ID:             m.ID,
		LessonPlanID:   m.LessonPlanID,
		Title:          m.Title,
		Subject:        m.Subject,
		GradeLevel:     m.GradeLevel,
		AlignmentScore: m.AlignmentScore,
		Feedback:       m.Feedback,
		Metadata:       string(m.Metadata),
		Suggestions:    string(m.Suggestions),
		CreatedAt:      m.CreatedAt,
	}
}

func fromDomainAnalysis(lpa *domain.LessonPlanAnalysis) *lessonPlanAnalysisModel {
	if lpa == nil {
		return nil
	}
	return &lessonPlanAnalysisModel{
		ID:             lpa.ID,
		LessonPlanID:   lpa.LessonPlanID,
		Title:          lpa.Title,
		Subject:        lpa.Subject,
		GradeLevel:     lpa.GradeLevel,
		AlignmentScore: lpa.AlignmentScore,
		Feedback:       lpa.Feedback,
		Metadata:       datatypes.JSON(lpa.Metadata),
		Suggestions:    datatypes.JSON(lpa.Suggestions),
		CreatedAt:      lpa.CreatedAt,
	}
}

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

	m := fromDomainAnalysis(analysis)
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

	m := &lessonPlanAnalysisModel{}
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

	return m.toDomain(), nil
}
