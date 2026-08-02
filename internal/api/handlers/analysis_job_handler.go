package handlers

import (
	"errors"
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/misalima/edunex-backend/internal/api/handlers/dto/response"
	"github.com/misalima/edunex-backend/internal/core/domain_errors"
	"github.com/misalima/edunex-backend/internal/infra/logger"
	"github.com/misalima/edunex-backend/internal/infra/queue"
	"go.uber.org/zap"
)

type AnalysisJobHandler struct {
	jobManager *queue.JobManager
}

func NewAnalysisJobHandler(jobManager *queue.JobManager) *AnalysisJobHandler {
	return &AnalysisJobHandler{jobManager: jobManager}
}

// Analyze godoc
// @Summary Analyze lesson plan
// @Description Enqueues a lesson plan for AI analysis processing in background
// @Tags Analysis
// @Produce json
// @Security BearerAuth
// @Param lesson_plan_id path string true "Lesson plan ID"
// @Success 202 {object} response.EnqueueAnalysisResponse
// @Failure 400 {object} response.ErrorMessageResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /lesson-plans/{lesson_plan_id}/analyze [post]
func (h *AnalysisJobHandler) Analyze(c echo.Context) error {
	ctx := c.Request().Context()

	lessonPlanIDStr := c.Param("lesson_plan_id")
	lessonPlanID, err := uuid.Parse(lessonPlanIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid lesson_plan_id format"})
	}

	jobID, err := h.jobManager.Enqueue(ctx, lessonPlanID)
	if err != nil {
		if errors.Is(err, domain_errors.ErrLessonPlanNotFound) || errors.Is(err, domain_errors.ErrNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "lesson plan not found"})
		}

		logger.Log.Error("failed to enqueue analysis job",
			zap.Error(err),
			zap.String("lesson_plan_id", lessonPlanID.String()))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to enqueue analysis"})
	}

	logger.Log.Info("analysis job enqueued",
		zap.String("job_id", jobID.String()),
		zap.String("lesson_plan_id", lessonPlanID.String()))

	return c.JSON(http.StatusAccepted, &response.EnqueueAnalysisResponse{
		JobID:        jobID.String(),
		LessonPlanID: lessonPlanID.String(),
		Status:       "pending",
	})
}

// GetJobStatus godoc
// @Summary Get analysis job status
// @Description Returns the status of an analysis job
// @Tags Analysis
// @Produce json
// @Security BearerAuth
// @Param job_id path string true "Job ID"
// @Success 200 {object} response.AnalysisJobResponse
// @Failure 400 {object} response.ErrorMessageResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 404 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /analysis-jobs/{job_id} [get]
func (h *AnalysisJobHandler) GetJobStatus(c echo.Context) error {
	ctx := c.Request().Context()

	jobIDStr := c.Param("job_id")
	jobID, err := uuid.Parse(jobIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid job_id format"})
	}

	logger.Log.Debug("fetching job status", zap.String("job_id", jobID.String()))

	job, err := h.jobManager.GetJobByID(ctx, jobID)
	if err != nil {
		logger.Log.Error("failed to fetch job status", zap.Error(err), zap.String("job_id", jobID.String()))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to fetch job status"})
	}

	if job == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "job not found"})
	}

	return c.JSON(http.StatusOK, response.FromDomainAnalysisJob(job))
}

// GetMetrics godoc
// @Summary Get job metrics
// @Description Returns job processing metrics and statistics
// @Tags Analysis
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.JobMetricsResponse
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /analysis-jobs/metrics [get]
func (h *AnalysisJobHandler) GetMetrics(c echo.Context) error {
	metrics := h.jobManager.GetMetrics()

	logger.Log.Debug("job metrics retrieved",
		zap.Any("metrics", metrics))

	dbStats, _ := metrics["db_stats"].(map[string]int64)

	resp := &response.JobMetricsResponse{
		ProcessedJobs:  getInt64Metric(metrics, "processed_jobs"),
		SuccessfulJobs: getInt64Metric(metrics, "successful_jobs"),
		FailedJobs:     getInt64Metric(metrics, "failed_jobs"),
		RetriedJobs:    getInt64Metric(metrics, "retried_jobs"),
		ActiveWorkers:  getIntMetric(metrics, "active_workers"),
		DbStats:        dbStats,
	}

	return c.JSON(http.StatusOK, resp)
}

func getInt64Metric(m map[string]interface{}, key string) int64 {
	if val, ok := m[key].(int64); ok {
		return val
	}
	return 0
}

func getIntMetric(m map[string]interface{}, key string) int {
	if val, ok := m[key].(int); ok {
		return val
	}
	return 0
}
