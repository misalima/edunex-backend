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
	"gorm.io/gorm"
)

var _ secondary.LessonPlanLoader = (*LessonPlanRepository)(nil)

type lessonPlanModel struct {
	ID         uuid.UUID               `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	UserID     uuid.UUID               `gorm:"type:uuid;not null;index:idx_lesson_plans_user_id"`
	Title      string                  `gorm:"type:text;not null"`
	FilePath   string                  `gorm:"type:text;not null"`
	Teacher    *string                 `gorm:"type:text"`
	Discipline *string                 `gorm:"type:text"`
	GradeLevel *domain.GradeLevel      `gorm:"type:varchar(50)"`
	Status     domain.LessonPlanStatus `gorm:"type:lesson_plan_status;not null;default:'pending'"`
	CreatedAt  time.Time               `gorm:"column:created_at;autoCreateTime"`
}

func (lessonPlanModel) TableName() string {
	return "lesson_plans"
}

func (m *lessonPlanModel) toDomain() *domain.LessonPlan {
	if m == nil {
		return nil
	}
	return &domain.LessonPlan{
		ID:         m.ID,
		UserID:     m.UserID,
		Title:      m.Title,
		FilePath:   m.FilePath,
		GradeLevel: m.GradeLevel,
		Teacher:    m.Teacher,
		Discipline: m.Discipline,
		Status:     m.Status,
		CreatedAt:  m.CreatedAt,
	}
}

func fromDomainLessonPlan(lp *domain.LessonPlan) *lessonPlanModel {
	if lp == nil {
		return nil
	}
	m := &lessonPlanModel{
		UserID:     lp.UserID,
		Title:      lp.Title,
		FilePath:   lp.FilePath,
		GradeLevel: lp.GradeLevel,
		Teacher:    lp.Teacher,
		Discipline: lp.Discipline,
		Status:     lp.Status,
	}
	if lp.ID != uuid.Nil {
		m.ID = lp.ID
	}
	if !lp.CreatedAt.IsZero() {
		m.CreatedAt = lp.CreatedAt
	}
	return m
}

type LessonPlanRepository struct {
	db *gorm.DB
}

func NewLessonPlanRepository(db *gorm.DB) *LessonPlanRepository {
	return &LessonPlanRepository{db: db}
}

func (r *LessonPlanRepository) InsertLessonPlan(ctx context.Context, lp *domain.LessonPlan) (uuid.UUID, error) {
	m := fromDomainLessonPlan(lp)
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	logger.Log.Info("inserting lesson plan", zap.String("user_id", m.UserID.String()), zap.String("title", m.Title), zap.String("temp_id", m.ID.String()))
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		logger.Log.Error("failed to insert lesson plan", zap.Error(err), zap.String("user_id", m.UserID.String()), zap.String("title", m.Title))
		return uuid.Nil, domain_errors.WrapUnexpectedMsg(err, "failed to insert lesson plan")
	}
	logger.Log.Info("lesson plan inserted", zap.String("lesson_plan_id", m.ID.String()), zap.String("user_id", m.UserID.String()))
	return m.ID, nil
}

func (r *LessonPlanRepository) GetLessonPlanByID(ctx context.Context, id uuid.UUID) (*domain.LessonPlan, error) {
	logger.Log.Debug("fetching lesson plan by id", zap.String("lesson_plan_id", id.String()))
	m := &lessonPlanModel{}
	if err := r.db.WithContext(ctx).First(m, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Log.Info("lesson plan not found", zap.String("lesson_plan_id", id.String()))
			return nil, domain_errors.NewNotFoundMsg("lesson plan not found")
		}
		logger.Log.Error("failed to fetch lesson plan", zap.Error(err), zap.String("lesson_plan_id", id.String()))
		return nil, domain_errors.WrapUnexpectedMsg(err, "failed to fetch lesson plan")
	}
	logger.Log.Debug("lesson plan fetched", zap.String("lesson_plan_id", m.ID.String()), zap.String("user_id", m.UserID.String()))
	return m.toDomain(), nil
}

func (r *LessonPlanRepository) ListLessonPlans(ctx context.Context) ([]*domain.LessonPlan, error) {
	logger.Log.Debug("listing lesson plans")
	var models []lessonPlanModel
	if err := r.db.WithContext(ctx).Order("created_at desc").Find(&models).Error; err != nil {
		logger.Log.Error("failed to list lesson plans", zap.Error(err))
		return nil, domain_errors.WrapUnexpectedMsg(err, "failed to list lesson plans")
	}
	logger.Log.Info("lesson plans listed", zap.Int("count", len(models)))
	out := make([]*domain.LessonPlan, len(models))
	for i := range models {
		out[i] = models[i].toDomain()
	}
	return out, nil
}

func (r *LessonPlanRepository) UpdateLessonPlan(ctx context.Context, lp *domain.LessonPlan) error {
	if lp == nil || lp.ID == uuid.Nil {
		return domain_errors.NewBadRequestMsg("lesson plan id is required")
	}
	m := fromDomainLessonPlan(lp)
	logger.Log.Info("updating lesson plan", zap.String("lesson_plan_id", m.ID.String()))
	updates := map[string]interface{}{
		"updated_at": time.Now(),
	}
	if m.Title != "" {
		updates["title"] = m.Title
	}
	if m.FilePath != "" {
		updates["file_path"] = m.FilePath
	}
	if m.Status != "" {
		updates["status"] = m.Status
	}
	res := r.db.WithContext(ctx).Model(&lessonPlanModel{}).Where("id = ?", m.ID).Updates(updates)
	if res.Error != nil {
		logger.Log.Error("failed to update lesson plan", zap.Error(res.Error), zap.String("lesson_plan_id", m.ID.String()))
		return domain_errors.WrapUnexpectedMsg(res.Error, "failed to update lesson plan")
	}
	if res.RowsAffected == 0 {
		logger.Log.Info("no rows affected when updating lesson plan", zap.String("lesson_plan_id", m.ID.String()))
		return domain_errors.NewNotFoundMsg("lesson plan not found")
	}
	logger.Log.Info("lesson plan updated", zap.String("lesson_plan_id", m.ID.String()), zap.Int64("rows_affected", res.RowsAffected))
	return nil
}
