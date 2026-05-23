package handlers

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
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
// @Success 202 {object} map[string]string
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
		logger.Log.Error("failed to enqueue analysis job",
			zap.Error(err),
			zap.String("lesson_plan_id", lessonPlanID.String()))
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to enqueue analysis"})
	}

	logger.Log.Info("analysis job enqueued",
		zap.String("job_id", jobID.String()),
		zap.String("lesson_plan_id", lessonPlanID.String()))

	return c.JSON(http.StatusAccepted, map[string]string{
		"job_id":         jobID.String(),
		"lesson_plan_id": lessonPlanID.String(),
		"status":         "pending",
	})
}

// GetJobStatus godoc
// @Summary Get analysis job status
// @Description Returns the status of an analysis job
// @Tags Analysis
// @Produce json
// @Security BearerAuth
// @Param job_id path string true "Job ID"
// @Success 200 {object} map[string]interface{}
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

	resp := map[string]interface{}{
		"job_id":         job.ID.String(),
		"lesson_plan_id": job.LessonPlanID.String(),
		"status":         job.Status,
		"attempts":      job.Attempts,
		"error_message":  job.ErrorMessage,
		"created_at":     job.CreatedAt,
		"started_at":     job.StartedAt,
		"finished_at":    job.FinishedAt,
	}

	return c.JSON(http.StatusOK, resp)
}

// GetMetrics godoc
// @Summary Get job metrics
// @Description Returns job processing metrics and statistics
// @Tags Analysis
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} response.ErrorResponse
// @Failure 500 {object} response.ErrorResponse
// @Router /analysis-jobs/metrics [get]
func (h *AnalysisJobHandler) GetMetrics(c echo.Context) error {
	metrics := h.jobManager.GetMetrics()

	logger.Log.Debug("job metrics retrieved",
		zap.Any("metrics", metrics))

	return c.JSON(http.StatusOK, metrics)
}
