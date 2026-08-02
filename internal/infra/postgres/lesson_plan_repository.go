package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/misalima/edunex-backend/internal/core/domain"
	"github.com/misalima/edunex-backend/internal/core/domain_errors"
	"github.com/misalima/edunex-backend/internal/core/interfaces/secondary"
	"github.com/misalima/edunex-backend/internal/infra/logger"
	"github.com/misalima/edunex-backend/internal/infra/postgres/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

var _ secondary.LessonPlanLoader = (*LessonPlanRepository)(nil)

type LessonPlanRepository struct {
	db *gorm.DB
}

func NewLessonPlanRepository(db *gorm.DB) *LessonPlanRepository {
	return &LessonPlanRepository{db: db}
}

func (r *LessonPlanRepository) InsertLessonPlan(ctx context.Context, lp *domain.LessonPlan) (uuid.UUID, error) {
	m := models.FromDomainLessonPlan(lp)
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	logger.Log.Info("inserting lesson plan", zap.String("user_id", m.UserID.String()), zap.String("title", m.Title), zap.String("temp_id", m.ID.String()))
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" && pgErr.ConstraintName == "lesson_plans_user_id_fkey" {
			logger.Log.Info("user not found when inserting lesson plan", zap.String("user_id", m.UserID.String()), zap.String("title", m.Title))
			return uuid.Nil, domain_errors.Wrap(
				domain_errors.ErrUserNotFound.Code,
				domain_errors.ErrUserNotFound.Message,
				domain_errors.ErrUserNotFound.HTTPStatus,
				err,
			)
		}
		logger.Log.Error("failed to insert lesson plan", zap.Error(err), zap.String("user_id", m.UserID.String()), zap.String("title", m.Title))
		return uuid.Nil, domain_errors.WrapUnexpectedMsg(err, "failed to insert lesson plan")
	}
	logger.Log.Info("lesson plan inserted", zap.String("lesson_plan_id", m.ID.String()), zap.String("user_id", m.UserID.String()))
	return m.ID, nil
}

func (r *LessonPlanRepository) GetLessonPlanByID(ctx context.Context, id uuid.UUID) (*domain.LessonPlan, error) {
	logger.Log.Debug("fetching lesson plan by id", zap.String("lesson_plan_id", id.String()))
	m := &models.LessonPlanModel{}
	if err := r.db.WithContext(ctx).First(m, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.Log.Info("lesson plan not found", zap.String("lesson_plan_id", id.String()))
			return nil, domain_errors.NewNotFoundMsg("lesson plan not found")
		}
		logger.Log.Error("failed to fetch lesson plan", zap.Error(err), zap.String("lesson_plan_id", id.String()))
		return nil, domain_errors.WrapUnexpectedMsg(err, "failed to fetch lesson plan")
	}
	logger.Log.Debug("lesson plan fetched", zap.String("lesson_plan_id", m.ID.String()), zap.String("user_id", m.UserID.String()))
	return m.ToDomain(), nil
}

func (r *LessonPlanRepository) ListLessonPlans(ctx context.Context, userID uuid.UUID, params domain.PaginationParams) ([]*domain.LessonPlan, int64, error) {
	logger.Log.Debug("listing lesson plans",
		zap.String("user_id", userID.String()),
		zap.Int("limit", params.Limit),
		zap.Int("offset", params.Offset),
	)

	var total int64
	if err := r.db.WithContext(ctx).
		Model(&models.LessonPlanModel{}).
		Where("user_id = ?", userID).
		Count(&total).Error; err != nil {
		logger.Log.Error("failed to count lesson plans", zap.Error(err), zap.String("user_id", userID.String()))
		return nil, 0, domain_errors.WrapUnexpectedMsg(err, "failed to count lesson plans")
	}

	var dbModels []models.LessonPlanModel
	query := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("created_at desc")

	if params.Limit > 0 {
		query = query.Limit(params.Limit)
	}
	if params.Offset > 0 {
		query = query.Offset(params.Offset)
	}

	if err := query.Find(&dbModels).Error; err != nil {
		logger.Log.Error("failed to list lesson plans", zap.Error(err), zap.String("user_id", userID.String()))
		return nil, 0, domain_errors.WrapUnexpectedMsg(err, "failed to list lesson plans")
	}

	logger.Log.Info("lesson plans listed",
		zap.String("user_id", userID.String()),
		zap.Int("count", len(dbModels)),
		zap.Int64("total", total),
	)

	out := make([]*domain.LessonPlan, len(dbModels))
	for i := range dbModels {
		out[i] = dbModels[i].ToDomain()
	}
	return out, total, nil
}

func (r *LessonPlanRepository) UpdateLessonPlan(ctx context.Context, lp *domain.LessonPlan) error {
	if lp == nil || lp.ID == uuid.Nil {
		return domain_errors.NewBadRequestMsg("lesson plan id is required")
	}
	m := models.FromDomainLessonPlan(lp)
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
	res := r.db.WithContext(ctx).Model(&models.LessonPlanModel{}).Where("id = ?", m.ID).Updates(updates)
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

func (r *LessonPlanRepository) DeleteLessonPlan(ctx context.Context, id uuid.UUID) error {
	logger.Log.Info("deleting lesson plan", zap.String("lesson_plan_id", id.String()))
	res := r.db.WithContext(ctx).Where("id = ?", id).Delete(&models.LessonPlanModel{})
	if res.Error != nil {
		logger.Log.Error("failed to delete lesson plan", zap.Error(res.Error), zap.String("lesson_plan_id", id.String()))
		return domain_errors.WrapUnexpectedMsg(res.Error, "failed to delete lesson plan")
	}
	return nil
}
