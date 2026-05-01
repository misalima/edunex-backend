package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/misalima/edunex-backend/internal/core/domain"
	"github.com/misalima/edunex-backend/internal/core/domain_errors"
	"github.com/misalima/edunex-backend/internal/core/interfaces/secondary"
	pgmodels "github.com/misalima/edunex-backend/internal/infra/postgres/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var _ secondary.AnalysisJobRepository = (*AnalysisRepository)(nil)
var _ secondary.LessonPlanAnalysisRepository = (*AnalysisRepository)(nil)

type AnalysisRepository struct {
	db *gorm.DB
}

func NewAnalysisRepository(db *gorm.DB) *AnalysisRepository {
	return &AnalysisRepository{db: db}
}

func (r *AnalysisRepository) CreateJob(ctx context.Context, job *domain.AnalysisJob) (*domain.AnalysisJob, error) {
	if job == nil {
		return nil, domain_errors.NewBadRequestMsg("analysis job is required")
	}
	if job.LessonPlanID == uuid.Nil {
		return nil, domain_errors.NewBadRequestMsg("lesson plan id is required")
	}

	m := pgmodels.FromDomainAnalysisJob(job)
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	if m.Status == "" {
		m.Status = domain.JobPending
	}

	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return nil, domain_errors.WrapUnexpectedMsg(err, "failed to create analysis job")
	}

	return m.ToDomain(), nil
}

func (r *AnalysisRepository) GetJobByLessonPlanID(ctx context.Context, lessonPlanID uuid.UUID) (*domain.AnalysisJob, error) {
	m := &pgmodels.AnalysisJobModel{}
	if err := r.db.WithContext(ctx).Where("lesson_plan_id = ?", lessonPlanID).First(m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain_errors.NewNotFoundMsg("analysis job not found")
		}
		return nil, domain_errors.WrapUnexpectedMsg(err, "failed to fetch analysis job")
	}

	return m.ToDomain(), nil
}

func (r *AnalysisRepository) ClaimPendingJobs(ctx context.Context, limit int) ([]*domain.AnalysisJob, error) {
	if limit <= 0 {
		limit = 10
	}

	var claimedJobs []*domain.AnalysisJob

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Step 1: Fetch pending jobs with row-level lock
		var models []pgmodels.AnalysisJobModel
		if err := tx.WithContext(ctx).
			Where("status = ?", domain.JobPending).
			Order("created_at asc").
			Limit(limit).
			Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Find(&models).Error; err != nil {
			return domain_errors.WrapUnexpectedMsg(err, "failed to fetch pending analysis jobs")
		}

		if len(models) == 0 {
			return nil
		}

		// Step 2: Claim jobs by updating status to processing
		res := tx.WithContext(ctx).
			Model(&pgmodels.AnalysisJobModel{}).
			Where("id IN ?", extractIDs(models)).
			Update("status", domain.JobProcessing)
		if res.Error != nil {
			return domain_errors.WrapUnexpectedMsg(res.Error, "failed to update analysis job status")
		}

		// Step 3: Convert claimed models to domain objects
		claimedJobs = make([]*domain.AnalysisJob, len(models))
		for i := range models {
			claimedJobs[i] = (&models[i]).ToDomain()
			claimedJobs[i].Status = domain.JobProcessing
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return claimedJobs, nil
}

func extractIDs(models []pgmodels.AnalysisJobModel) []uuid.UUID {
	ids := make([]uuid.UUID, len(models))
	for i := range models {
		ids[i] = models[i].ID
	}
	return ids
}

func (r *AnalysisRepository) MarkJobProcessing(ctx context.Context, jobID uuid.UUID, startedAt time.Time) error {
	res := r.db.WithContext(ctx).
		Model(&pgmodels.AnalysisJobModel{}).
		Where("id = ?", jobID).
		Updates(map[string]interface{}{
			"status":     domain.JobProcessing,
			"started_at": startedAt,
		})
	if res.Error != nil {
		return domain_errors.WrapUnexpectedMsg(res.Error, "failed to mark analysis job as processing")
	}
	if res.RowsAffected == 0 {
		return domain_errors.NewNotFoundMsg("analysis job not found")
	}
	return nil
}

func (r *AnalysisRepository) MarkJobCompleted(ctx context.Context, jobID uuid.UUID, result *domain.LessonPlanAnalysis, finishedAt time.Time) error {
	if result == nil {
		return domain_errors.NewBadRequestMsg("analysis result is required")
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if _, err := r.saveAnalysisResultTx(ctx, tx, result); err != nil {
			return err
		}

		res := tx.WithContext(ctx).
			Model(&pgmodels.AnalysisJobModel{}).
			Where("id = ?", jobID).
			Updates(map[string]interface{}{
				"status":        domain.JobDone,
				"finished_at":   finishedAt,
				"error_message": nil,
			})
		if res.Error != nil {
			return domain_errors.WrapUnexpectedMsg(res.Error, "failed to mark analysis job as completed")
		}
		if res.RowsAffected == 0 {
			return domain_errors.NewNotFoundMsg("analysis job not found")
		}

		return nil
	})
}

func (r *AnalysisRepository) MarkJobFailed(ctx context.Context, jobID uuid.UUID, reason string, finishedAt time.Time) error {
	res := r.db.WithContext(ctx).
		Model(&pgmodels.AnalysisJobModel{}).
		Where("id = ?", jobID).
		Updates(map[string]interface{}{
			"status":        domain.JobFailed,
			"finished_at":   finishedAt,
			"error_message": reason,
			"attempts":      gorm.Expr("attempts + 1"),
		})
	if res.Error != nil {
		return domain_errors.WrapUnexpectedMsg(res.Error, "failed to mark analysis job as failed")
	}
	if res.RowsAffected == 0 {
		return domain_errors.NewNotFoundMsg("analysis job not found")
	}
	return nil
}

func (r *AnalysisRepository) SaveAnalysisResult(ctx context.Context, a *domain.LessonPlanAnalysis) (*domain.LessonPlanAnalysis, error) {
	return r.saveAnalysisResultTx(ctx, r.db, a)
}

func (r *AnalysisRepository) saveAnalysisResultTx(ctx context.Context, tx *gorm.DB, a *domain.LessonPlanAnalysis) (*domain.LessonPlanAnalysis, error) {
	m, err := pgmodels.FromDomainLessonPlanAnalysis(a)
	if err != nil {
		return nil, err
	}
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}

	existing := &pgmodels.LessonPlanAnalysisModel{}
	query := tx.WithContext(ctx).Where("lesson_plan_id = ?", m.LessonPlanID).First(existing)
	if query.Error != nil {
		if !errors.Is(query.Error, gorm.ErrRecordNotFound) {
			return nil, domain_errors.WrapUnexpectedMsg(query.Error, "failed to fetch analysis result")
		}

		if err := tx.WithContext(ctx).Create(m).Error; err != nil {
			return nil, domain_errors.WrapUnexpectedMsg(err, "failed to save analysis result")
		}
		return m.ToDomain()
	}

	existing.AnalysisText = m.AnalysisText
	existing.StructuredData = m.StructuredData
	if err := tx.WithContext(ctx).Save(existing).Error; err != nil {
		return nil, domain_errors.WrapUnexpectedMsg(err, "failed to update analysis result")
	}

	return existing.ToDomain()
}

func (r *AnalysisRepository) GetAnalysisByLessonPlanID(ctx context.Context, lessonPlanID uuid.UUID) (*domain.LessonPlanAnalysis, error) {
	m := &pgmodels.LessonPlanAnalysisModel{}
	if err := r.db.WithContext(ctx).Where("lesson_plan_id = ?", lessonPlanID).First(m).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain_errors.NewNotFoundMsg("lesson plan analysis not found")
		}
		return nil, domain_errors.WrapUnexpectedMsg(err, "failed to fetch lesson plan analysis")
	}

	return m.ToDomain()
}
