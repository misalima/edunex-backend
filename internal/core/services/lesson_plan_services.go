package services

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/misalima/edunex-backend/internal/core/interfaces/primary"
	"go.uber.org/zap"

	"github.com/misalima/edunex-backend/internal/core/domain"
	"github.com/misalima/edunex-backend/internal/core/domain_errors"
	"github.com/misalima/edunex-backend/internal/core/interfaces/secondary"
	"github.com/misalima/edunex-backend/internal/infra/logger"
)

const DefaultLinkExpiration = 3600

type LessonPlanService struct {
	repo           secondary.LessonPlanLoader
	storage        secondary.StorageClient
	enqueuer       secondary.AnalysisJobEnqueuer
	jobLoader      secondary.AnalysisJobLoader
	analysisLoader secondary.LessonPlanAnalysisLoader
}

var _ primary.LessonPlanManager = (*LessonPlanService)(nil)

func NewLessonPlanService(
	repo secondary.LessonPlanLoader,
	storage secondary.StorageClient,
	enqueuer secondary.AnalysisJobEnqueuer,
	jobLoader secondary.AnalysisJobLoader,
	analysisLoader secondary.LessonPlanAnalysisLoader,
) *LessonPlanService {
	return &LessonPlanService{
		repo:           repo,
		storage:        storage,
		enqueuer:       enqueuer,
		jobLoader:      jobLoader,
		analysisLoader: analysisLoader,
	}
}

func (s *LessonPlanService) CreateLessonPlan(ctx context.Context, lp *domain.LessonPlan, reader io.Reader, filename, contentType string) (*domain.LessonPlan, error) {
	if lp == nil {
		return nil, domain_errors.NewBadRequestMsg("lesson plan is required")
	}
	if lp.UserID == uuid.Nil {
		return nil, domain_errors.NewBadRequestMsg("user id is required")
	}
	if strings.TrimSpace(lp.Title) == "" {
		return nil, domain_errors.NewBadRequestMsg("title is required")
	}

	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		ext = ""
	}

	objectPath := fmt.Sprintf("lesson_plans/%s/%s%s", lp.UserID.String(), uuid.New().String(), ext)

	logger.Log.Info("lesson plan upload start", zap.String("user_id", lp.UserID.String()), zap.String("object_path", objectPath), zap.String("title", lp.Title))

	uploadedURL, err := s.storage.Upload(ctx, objectPath, reader, contentType)
	if err != nil {
		logger.Log.Error("storage upload failed", zap.Error(err), zap.String("user_id", lp.UserID.String()), zap.String("object_path", objectPath))
		return nil, err
	}

	lp.FilePath = objectPath
	if lp.Status == "" {
		lp.Status = "pending"
	}

	id, err := s.repo.InsertLessonPlan(ctx, lp)
	if err != nil {
		logger.Log.Error("db insert failed, attempting cleanup", zap.Error(err), zap.String("object_path", objectPath), zap.String("user_id", lp.UserID.String()))
		if derr := s.storage.Delete(ctx, objectPath); derr != nil {
			logger.Log.Error("cleanup delete failed", zap.Error(derr), zap.String("object_path", objectPath))
		}
		return nil, err
	}

	createdLP, err := s.repo.GetLessonPlanByID(ctx, id)
	if err != nil {
		logger.Log.Error("failed to fetch created lesson plan", zap.Error(err), zap.String("lesson_plan_id", id.String()))
		return nil, err
	}

	if s.enqueuer != nil {
		enqueueCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		if jobID, err := s.enqueuer.Enqueue(enqueueCtx, createdLP.ID); err != nil {
			logger.Log.Error("failed to enqueue lesson plan analysis, performing cleanup",
				zap.Error(err),
				zap.String("lesson_plan_id", createdLP.ID.String()))
			if derr := s.repo.DeleteLessonPlan(ctx, createdLP.ID); derr != nil {
				logger.Log.Error("cleanup db delete failed", zap.Error(derr), zap.String("lesson_plan_id", createdLP.ID.String()))
			}
			if derr := s.storage.Delete(ctx, objectPath); derr != nil {
				logger.Log.Error("cleanup storage delete failed", zap.Error(derr), zap.String("object_path", objectPath))
			}
			return nil, domain_errors.WrapUnexpectedMsg(err, "failed to enqueue analysis job")
		} else {
			logger.Log.Info("lesson plan analysis enqueued",
				zap.String("lesson_plan_id", createdLP.ID.String()),
				zap.String("job_id", jobID.String()))
		}
	} else {
		logger.Log.Warn("lesson plan analysis enqueuer not configured",
			zap.String("lesson_plan_id", createdLP.ID.String()))
	}

	logger.Log.Info("lesson plan created", zap.String("lesson_plan_id", createdLP.ID.String()), zap.String("user_id", createdLP.UserID.String()), zap.String("access_url", uploadedURL))
	return createdLP, nil
}

func (s *LessonPlanService) GetLessonPlanWithSignedURL(ctx context.Context, id uuid.UUID) (*domain.LessonPlan, string, error) {
	lp, err := s.repo.GetLessonPlanByID(ctx, id)
	if err != nil {
		return nil, "", err
	}
	if lp == nil {
		return nil, "", domain_errors.NewNotFoundMsg("lesson plan not found")
	}

	signedURL, err := s.storage.SignURL(ctx, lp.FilePath, DefaultLinkExpiration)
	if err != nil {
		logger.Log.Error("failed to sign url", zap.Error(err), zap.String("lesson_plan_id", id.String()), zap.String("file_path", lp.FilePath))
		return lp, "", domain_errors.WrapUnexpectedMsg(err, "failed to create signed url")
	}

	return lp, signedURL, nil
}

func (s *LessonPlanService) ListLessonPlansWithSignedURLs(ctx context.Context, userID uuid.UUID, params domain.PaginationParams) ([]*domain.LessonPlan, map[string]string, int64, error) {
	lps, total, err := s.repo.ListLessonPlans(ctx, userID, params)
	if err != nil {
		return nil, nil, 0, err
	}

	urls := make(map[string]string, len(lps))
	for _, lp := range lps {
		if lp == nil {
			continue
		}
		signed, err := s.storage.SignURL(ctx, lp.FilePath, DefaultLinkExpiration)
		if err != nil {
			logger.Log.Error("failed to sign url for lesson plan", zap.Error(err), zap.String("lesson_plan_id", lp.ID.String()), zap.String("file_path", lp.FilePath))
			continue
		}
		urls[lp.ID.String()] = signed
	}

	return lps, urls, total, nil
}

func (s *LessonPlanService) DeleteLessonPlan(ctx context.Context, userID uuid.UUID, lessonPlanID uuid.UUID) error {
	if lessonPlanID == uuid.Nil {
		return domain_errors.NewBadRequestMsg("lesson plan id is required")
	}

	lp, err := s.repo.GetLessonPlanByID(ctx, lessonPlanID)
	if err != nil {
		return err
	}
	if lp == nil {
		return domain_errors.NewNotFoundMsg("lesson plan not found")
	}
	if lp.UserID != userID {
		return domain_errors.NewNotFoundMsg("lesson plan not found")
	}

	if err := s.storage.Delete(ctx, lp.FilePath); err != nil {
		logger.Log.Error("failed to delete file from storage, proceeding with db deletion",
			zap.Error(err),
			zap.String("lesson_plan_id", lessonPlanID.String()),
			zap.String("file_path", lp.FilePath),
		)
	}

	if err := s.repo.DeleteLessonPlan(ctx, lessonPlanID); err != nil {
		logger.Log.Error("failed to delete lesson plan from db",
			zap.Error(err),
			zap.String("lesson_plan_id", lessonPlanID.String()),
		)
		return err
	}

	logger.Log.Info("lesson plan deleted",
		zap.String("lesson_plan_id", lessonPlanID.String()),
		zap.String("user_id", userID.String()),
	)
	return nil
}

func (s *LessonPlanService) GetAnalysisStatus(ctx context.Context, lessonPlanID uuid.UUID) (*domain.LessonPlanAnalysisStatus, error) {
	if lessonPlanID == uuid.Nil {
		return nil, domain_errors.NewBadRequestMsg("lesson_plan_id is required")
	}

	lp, err := s.repo.GetLessonPlanByID(ctx, lessonPlanID)
	if err != nil {
		return nil, err
	}
	if lp == nil {
		return nil, domain_errors.NewNotFoundMsg("lesson plan not found")
	}

	job, err := s.jobLoader.GetJobByLessonPlanID(ctx, lessonPlanID)
	if err != nil {
		logger.Log.Error("failed to fetch job by lesson plan ID", zap.Error(err), zap.String("lesson_plan_id", lessonPlanID.String()))
		return nil, err
	}

	if job == nil {
		return nil, domain_errors.NewNotFoundMsg("analysis job not found")
	}

	resp := &domain.LessonPlanAnalysisStatus{
		Status:       job.Status,
		ErrorMessage: job.ErrorMessage,
	}

	if job.Status == "done" {
		analysis, err := s.analysisLoader.GetAnalysisByLessonPlanID(ctx, lessonPlanID)
		if err != nil {
			logger.Log.Error("failed to fetch final analysis result", zap.Error(err), zap.String("lesson_plan_id", lessonPlanID.String()))
			return nil, err
		}
		resp.Analysis = analysis
	}

	return resp, nil
}
