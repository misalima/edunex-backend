package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/misalima/edunex-backend/internal/core/domain"
	"github.com/misalima/edunex-backend/internal/infra/postgres/models"
	"gorm.io/gorm"
)

// AnalysisJobRepository handles database operations for analysis jobs
type AnalysisJobRepository struct {
	db *gorm.DB
}

// NewAnalysisJobRepository creates a new analysis job repository
func NewAnalysisJobRepository(db *gorm.DB) *AnalysisJobRepository {
	return &AnalysisJobRepository{db: db}
}

// UpsertAnalysisJob inserts or updates an analysis job using ON CONFLICT
// Returns the job ID and error
func (r *AnalysisJobRepository) UpsertAnalysisJob(ctx context.Context, lessonPlanID uuid.UUID) (uuid.UUID, error) {
	jobID := uuid.New()

	result := r.db.WithContext(ctx).Exec(`
		INSERT INTO analysis_jobs (id, lesson_plan_id, status, attempts, created_at)
		VALUES (?, ?, ?, 0, now())
		ON CONFLICT (lesson_plan_id) DO UPDATE SET
			status = 'pending',
			attempts = 0,
			error_message = NULL,
			created_at = now(),
			started_at = NULL,
			finished_at = NULL
		WHERE analysis_jobs.status IN ('done', 'failed')
	`, jobID, lessonPlanID, "pending")

	if result.Error != nil {
		return uuid.Nil, fmt.Errorf("failed to upsert analysis job: %w", result.Error)
	}

	// If no rows were affected, fetch the existing job ID
	if result.RowsAffected == 0 {
		var existingJob models.AnalysisJobModel
		if err := r.db.WithContext(ctx).
			Where("lesson_plan_id = ?", lessonPlanID).
			First(&existingJob).Error; err != nil {
			return uuid.Nil, fmt.Errorf("failed to fetch existing job: %w", err)
		}
		return existingJob.ID, nil
	}

	return jobID, nil
}

// FetchPendingJob fetches a single pending job and atomically marks it as processing
// Uses SELECT FOR UPDATE SKIP LOCKED to avoid race conditions
func (r *AnalysisJobRepository) FetchPendingJob(ctx context.Context) (*domain.AnalysisJob, error) {
	var job models.AnalysisJobModel

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Raw query to use SELECT FOR UPDATE SKIP LOCKED
		if err := tx.Raw(`
			SELECT * FROM analysis_jobs
			WHERE status = 'pending'
			ORDER BY created_at ASC
			LIMIT 1
			FOR UPDATE SKIP LOCKED
		`).Scan(&job).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil // No pending jobs
			}
			return fmt.Errorf("failed to fetch pending job: %w", err)
		}

		// Mark as processing atomically within the same transaction
		if job.ID != uuid.Nil {
			now := time.Now()
			if err := tx.Model(&job).
				Updates(map[string]interface{}{
					"status":     domain.JobProcessing,
					"started_at": now,
				}).Error; err != nil {
				return fmt.Errorf("failed to mark job as processing: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	if job.ID == uuid.Nil {
		return nil, nil // No pending jobs found
	}

	return job.ToDomain(), nil
}

// FetchPendingJobByID fetches a specific pending job by ID and atomically marks it as processing.
// Returns nil if the job is not in pending status (already claimed by another worker).
func (r *AnalysisJobRepository) FetchPendingJobByID(ctx context.Context, jobID uuid.UUID) (*domain.AnalysisJob, error) {
	var job models.AnalysisJobModel

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Fetch specific job with lock if pending
		if err := tx.Raw(`
			SELECT * FROM analysis_jobs
			WHERE id = ? AND status = 'pending'
			FOR UPDATE SKIP LOCKED
		`, jobID).Scan(&job).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				return nil // Job not found or not pending
			}
			return fmt.Errorf("failed to fetch job: %w", err)
		}

		// Mark as processing atomically within the same transaction
		if job.ID != uuid.Nil {
			now := time.Now()
			if err := tx.Model(&job).
				Updates(map[string]interface{}{
					"status":     domain.JobProcessing,
					"started_at": now,
				}).Error; err != nil {
				return fmt.Errorf("failed to mark job as processing: %w", err)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	if job.ID == uuid.Nil {
		return nil, nil // Job not found or not pending
	}

	return job.ToDomain(), nil
}

// MarkJobDone marks a job as completed
func (r *AnalysisJobRepository) MarkJobDone(ctx context.Context, jobID uuid.UUID) error {
	now := time.Now()
	result := r.db.WithContext(ctx).Model(&models.AnalysisJobModel{}).
		Where("id = ?", jobID).
		Updates(map[string]interface{}{
			"status":        domain.JobDone,
			"finished_at":   now,
			"error_message": nil,
		})

	if result.Error != nil {
		return fmt.Errorf("failed to mark job as done: %w", result.Error)
	}

	return nil
}

// MarkJobFailed marks a job as failed and increments attempt count
func (r *AnalysisJobRepository) MarkJobFailed(ctx context.Context, jobID uuid.UUID, errorMsg string, maxAttempts int) error {
	var job models.AnalysisJobModel

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Fetch current job
		if err := tx.WithContext(ctx).Where("id = ?", jobID).First(&job).Error; err != nil {
			return fmt.Errorf("failed to fetch job: %w", err)
		}

		now := time.Now()
		updates := map[string]interface{}{
			"finished_at":   now,
			"error_message": errorMsg,
			"attempts":      job.Attempts + 1,
		}

		// Decide retry or final failure
		if job.Attempts+1 >= maxAttempts {
			updates["status"] = domain.JobFailed
		} else {
			updates["status"] = domain.JobPending
			updates["started_at"] = nil
		}

		return tx.Model(&job).Updates(updates).Error
	})

	if err != nil {
		return fmt.Errorf("failed to mark job as failed: %w", err)
	}

	return nil
}

// GetJobByLessonPlanID fetches a job by lesson plan ID
func (r *AnalysisJobRepository) GetJobByLessonPlanID(ctx context.Context, lessonPlanID uuid.UUID) (*domain.AnalysisJob, error) {
	var job models.AnalysisJobModel

	if err := r.db.WithContext(ctx).
		Where("lesson_plan_id = ?", lessonPlanID).
		First(&job).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to fetch job: %w", err)
	}

	return job.ToDomain(), nil
}

// CleanupStaleProcessingJobs marks processing jobs that haven't been updated as pending (recovery from crashes)
func (r *AnalysisJobRepository) CleanupStaleProcessingJobs(ctx context.Context, staleThreshold time.Duration) error {
	staleBefore := time.Now().Add(-staleThreshold)

	result := r.db.WithContext(ctx).
		Model(&models.AnalysisJobModel{}).
		Where("status = ? AND started_at < ?", domain.JobProcessing, staleBefore).
		Updates(map[string]interface{}{
			"status":     domain.JobPending,
			"started_at": nil,
		})

	if result.Error != nil {
		return fmt.Errorf("failed to cleanup stale jobs: %w", result.Error)
	}

	return nil
}

// GetJobStatistics returns job statistics
func (r *AnalysisJobRepository) GetJobStatistics(ctx context.Context) (map[string]int64, error) {
	type StatusCount struct {
		Status string
		Count  int64
	}

	var counts []StatusCount
	if err := r.db.WithContext(ctx).
		Model(&models.AnalysisJobModel{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Scan(&counts).Error; err != nil {
		return nil, fmt.Errorf("failed to get job statistics: %w", err)
	}

	stats := map[string]int64{
		"pending":    0,
		"processing": 0,
		"done":       0,
		"failed":     0,
	}

	for _, sc := range counts {
		stats[sc.Status] = sc.Count
	}

	return stats, nil
}

// SaveAnalysisAndMarkDone saves the analysis result and atomically marks the job as done.
// This ensures consistency: either both operations happen or neither does.
func (r *AnalysisJobRepository) SaveAnalysisAndMarkDone(ctx context.Context, analysis *domain.LessonPlanAnalysis, jobID uuid.UUID) error {
	// Insert analysis model (which implements InsertAnalysis) and mark job done in one transaction
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Insert the analysis
		model := models.FromDomainAnalysis(analysis)
		if model.ID == uuid.Nil {
			model.ID = uuid.New()
		}

		// Delete any existing analysis for this lesson plan to prevent unique constraint violation on re-analysis
		if err := tx.Where("lesson_plan_id = ?", analysis.LessonPlanID).Delete(&models.LessonPlanAnalysisModel{}).Error; err != nil {
			return fmt.Errorf("failed to delete existing analysis: %w", err)
		}

		if err := tx.Create(model).Error; err != nil {
			return fmt.Errorf("failed to insert analysis: %w", err)
		}

		// Mark job as done
		now := time.Now()
		if err := tx.Model(&models.AnalysisJobModel{}).
			Where("id = ?", jobID).
			Updates(map[string]interface{}{
				"status":        domain.JobDone,
				"finished_at":   now,
				"error_message": nil,
			}).Error; err != nil {
			return fmt.Errorf("failed to mark job as done: %w", err)
		}

		return nil
	})

	return err
}

// GetJobByID fetches an analysis job by ID
func (r *AnalysisJobRepository) GetJobByID(ctx context.Context, jobID uuid.UUID) (*domain.AnalysisJob, error) {
	var job models.AnalysisJobModel

	if err := r.db.WithContext(ctx).
		Where("id = ?", jobID).
		First(&job).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to fetch job by ID: %w", err)
	}

	return job.ToDomain(), nil
}
