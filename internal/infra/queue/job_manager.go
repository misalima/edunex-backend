package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"github.com/misalima/edunex-backend/internal/core/domain"
	"github.com/misalima/edunex-backend/internal/core/interfaces/secondary"
	"github.com/misalima/edunex-backend/internal/infra/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// JobManagerConfig holds configuration for the job manager
type JobManagerConfig struct {
	WorkerCount    int
	MaxAttempts    int
	PollInterval   time.Duration
	StaleThreshold time.Duration
}

// DefaultJobManagerConfig returns a sensible default configuration
func DefaultJobManagerConfig() *JobManagerConfig {
	return &JobManagerConfig{
		WorkerCount:    3,
		MaxAttempts:    3,
		PollInterval:   10 * time.Second,
		StaleThreshold: 30 * time.Minute,
	}
}

// JobManager orchestrates job processing with worker pool and notification system
type JobManager struct {
	db              *gorm.DB
	dbURL           string // PostgreSQL connection string for listener
	config          *JobManagerConfig
	aiProvider      secondary.AIProvider
	extractors      secondary.DataExtractor
	lessonPlanRepo  secondary.LessonPlanLoader
	analysisLoader  secondary.LessonPlanAnalysisLoader
	analysisJobRepo secondary.AnalysisJobLoader
	jobChan         chan uuid.UUID
	workerWg        sync.WaitGroup
	notifyListener  *NotificationListener
	ctx             context.Context
	cancel          context.CancelFunc
	isRunning       bool
	metrics         *JobMetrics
	pollTicker      *time.Ticker
	mu              sync.RWMutex
}

var _ secondary.AnalysisJobEnqueuer = (*JobManager)(nil)

// JobMetrics tracks statistics about job processing
type JobMetrics struct {
	ProcessedJobs  atomic.Int64
	FailedJobs     atomic.Int64
	SuccessfulJobs atomic.Int64
	LastUpdated    time.Time
	mu             sync.RWMutex
}

// NewJobManager creates a new job manager
// dbURL should be the PostgreSQL connection string (used for NOTIFY/LISTEN listener)
func NewJobManager(
	db *gorm.DB,
	dbURL string,
	config *JobManagerConfig,
	aiProvider secondary.AIProvider,
	extractors secondary.DataExtractor,
	lessonPlanRepo secondary.LessonPlanLoader,
	analysisLoader secondary.LessonPlanAnalysisLoader,
	analysisJobRepo secondary.AnalysisJobLoader,
) *JobManager {
	ctx, cancel := context.WithCancel(context.Background())

	return &JobManager{
		db:              db,
		dbURL:           dbURL,
		config:          config,
		aiProvider:      aiProvider,
		extractors:      extractors,
		lessonPlanRepo:  lessonPlanRepo,
		analysisLoader:  analysisLoader,
		analysisJobRepo: analysisJobRepo,
		jobChan:         make(chan uuid.UUID, config.WorkerCount*2),
		ctx:             ctx,
		cancel:          cancel,
		metrics:         &JobMetrics{},
	}
}

// Start initializes the worker pool and notification listener
func (jm *JobManager) Start() error {
	jm.mu.Lock()
	if jm.isRunning {
		jm.mu.Unlock()
		return fmt.Errorf("job manager is already running")
	}
	jm.mu.Unlock()

	// Initialize notification listener with explicit connection string
	if jm.dbURL == "" {
		logger.Log.Warn("database URL not provided, notification listener will be unavailable")
	} else {
		listener, err := NewNotificationListener(jm.dbURL, jm.jobChan)
		if err != nil {
			logger.Log.Warn("failed to create notification listener, will use polling fallback", zap.Error(err))
		} else {
			if err := listener.Start(); err != nil {
				logger.Log.Warn("failed to start notification listener, will use polling fallback", zap.Error(err))
			} else {
				jm.notifyListener = listener
			}
		}
	}

	// Cleanup stale jobs left from previous instance
	if err := jm.analysisJobRepo.CleanupStaleProcessingJobs(jm.ctx, jm.config.StaleThreshold); err != nil {
		logger.Log.Error("failed to cleanup stale jobs", zap.Error(err))
		// Don't fail startup for this
	}

	// Start worker pool
	for i := 0; i < jm.config.WorkerCount; i++ {
		jm.workerWg.Add(1)
		go jm.worker(i)
	}

	// Start polling fallback
	jm.pollTicker = time.NewTicker(jm.config.PollInterval)
	jm.workerWg.Add(1)
	go jm.pollFallback()

	jm.mu.Lock()
	jm.isRunning = true
	jm.mu.Unlock()

	logger.Log.Info("Job manager started",
		zap.Int("worker_count", jm.config.WorkerCount),
		zap.Int("max_attempts", jm.config.MaxAttempts),
	)

	return nil
}

// Stop gracefully shuts down the job manager
func (jm *JobManager) Stop() error {
	jm.mu.Lock()
	if !jm.isRunning {
		jm.mu.Unlock()
		return fmt.Errorf("job manager is not running")
	}
	jm.mu.Unlock()

	logger.Log.Info("Stopping job manager...")

	// Stop polling
	if jm.pollTicker != nil {
		jm.pollTicker.Stop()
	}

	// Stop notification listener
	if jm.notifyListener != nil && jm.notifyListener.IsActive() {
		if err := jm.notifyListener.Stop(); err != nil {
			logger.Log.Error("failed to stop notification listener", zap.Error(err))
		}
	}

	// Signal workers to stop
	jm.cancel()

	// Wait for workers with timeout
	done := make(chan struct{})
	go func() {
		jm.workerWg.Wait()
		close(done)
	}()

	timeout := time.After(10 * time.Second)
	select {
	case <-done:
		logger.Log.Info("All workers stopped gracefully")
	case <-timeout:
		logger.Log.Warn("Job manager shutdown timeout")
	}

	// Close job channel
	close(jm.jobChan)

	jm.mu.Lock()
	jm.isRunning = false
	jm.mu.Unlock()

	return nil
}

// Enqueue enqueues a new job for a lesson plan
func (jm *JobManager) Enqueue(ctx context.Context, lessonPlanID uuid.UUID) (uuid.UUID, error) {
	// Insert/update job in database
	jobID, err := jm.analysisJobRepo.UpsertAnalysisJob(ctx, lessonPlanID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to upsert job: %w", err)
	}

	// Send NOTIFY to database
	err = jm.db.WithContext(ctx).
		Exec("NOTIFY analysis_jobs_channel, ?", jobID.String()).Error
	if err != nil {
		logger.Log.Warn("failed to send notification", zap.Error(err), zap.String("job_id", jobID.String()))
		// Don't fail enqueue - polling will pick it up
	}

	// Try to push to channel (non-blocking, fallback to polling)
	select {
	case jm.jobChan <- jobID:
		logger.Log.Debug("job enqueued", zap.String("job_id", jobID.String()), zap.String("lesson_plan_id", lessonPlanID.String()))
	default:
		logger.Log.Debug("job channel full, relying on polling", zap.String("job_id", jobID.String()))
	}

	return jobID, nil
}

// worker processes jobs from the queue.
// It can receive either:
// - A specific job ID to process
// - uuid.Nil as a "wake-up" signal to check for any pending jobs
func (jm *JobManager) worker(id int) {
	defer jm.workerWg.Done()

	logger.Log.Info("Worker started", zap.Int("worker_id", id))

	for {
		select {
		case <-jm.ctx.Done():
			logger.Log.Info("Worker stopping", zap.Int("worker_id", id))
			return

		case jobID := <-jm.jobChan:
			if jobID == uuid.Nil {
				// Wake-up signal from polling fallback: try to fetch any pending job
				jm.fetchAndProcessNextJob(id)
			} else {
				// Specific job ID from notification
				jm.processJob(jobID, id)
			}
		}
	}
}

// fetchAndProcessNextJob tries to claim and process the next available pending job.
// Used when receiving wake-up signal from polling fallback.
func (jm *JobManager) fetchAndProcessNextJob(workerID int) {
	job, err := jm.analysisJobRepo.FetchPendingJob(jm.ctx)
	if err != nil {
		logger.Log.Error("failed to fetch next job", zap.Error(err), zap.Int("worker_id", workerID))
		return
	}

	if job == nil {
		logger.Log.Debug("no pending jobs available", zap.Int("worker_id", workerID))
		return
	}

	// Process the claimed job
	jm.processJobWithData(job, workerID)
}

// processJob processes a specific job by ID.
// Fetches the job with atomic lock-and-claim (SELECT FOR UPDATE) to prevent race conditions.
func (jm *JobManager) processJob(jobID uuid.UUID, workerID int) {
	logger.Log.Info("Processing job", zap.String("job_id", jobID.String()), zap.Int("worker_id", workerID))

	// Fetch and claim the specific job atomically.
	// FetchPendingJob uses SELECT FOR UPDATE SKIP LOCKED to ensure only one worker gets it.
	// If this specific job is not pending, we'll skip it (another worker may have claimed it first).
	job, err := jm.analysisJobRepo.FetchPendingJobByID(jm.ctx, jobID)
	if err != nil {
		logger.Log.Error("failed to fetch job", zap.Error(err), zap.String("job_id", jobID.String()))
		jm.handleJobFailure(jobID, err.Error())
		return
	}

	if job == nil {
		logger.Log.Debug("job not found or already claimed", zap.String("job_id", jobID.String()), zap.Int("worker_id", workerID))
		return
	}

	jm.processJobWithData(job, workerID)
}

// processJobWithData does the actual work of processing a job that's already been fetched and claimed.
func (jm *JobManager) processJobWithData(job *domain.AnalysisJob, workerID int) {
	// Get lesson plan
	lessonPlan, err := jm.lessonPlanRepo.GetLessonPlanByID(jm.ctx, job.LessonPlanID)
	if err != nil {
		errorMsg := fmt.Sprintf("failed to fetch lesson plan: %v", err)
		logger.Log.Error(errorMsg, zap.Error(err), zap.String("job_id", job.ID.String()))
		jm.handleJobFailure(job.ID, errorMsg)
		return
	}

	if lessonPlan == nil {
		errorMsg := "lesson plan not found"
		logger.Log.Error(errorMsg, zap.String("job_id", job.ID.String()))
		jm.handleJobFailure(job.ID, errorMsg)
		return
	}

	// Extract content from file
	content, err := jm.extractors.ExtractFromStorage(jm.ctx, lessonPlan.FilePath)
	if err != nil {
		errorMsg := fmt.Sprintf("failed to extract content: %v", err)
		logger.Log.Error(errorMsg, zap.Error(err), zap.String("job_id", job.ID.String()), zap.String("file_path", lessonPlan.FilePath))
		jm.handleJobFailure(job.ID, errorMsg)
		return
	}

	// Analyze with AI
	analysisResult, err := jm.aiProvider.Analyze(jm.ctx, content)
	if err != nil {
		errorMsg := fmt.Sprintf("AI analysis failed: %v", err)
		logger.Log.Error(errorMsg, zap.Error(err), zap.String("job_id", job.ID.String()))
		jm.handleJobFailure(job.ID, errorMsg)
		return
	}

	// Persist analysis result and mark job as done in a single transaction
	analysis, err := analysisResultToDomain(job.LessonPlanID, analysisResult)
	if err != nil {
		errorMsg := fmt.Sprintf("failed to build analysis domain model: %v", err)
		logger.Log.Error(errorMsg, zap.Error(err), zap.String("job_id", job.ID.String()))
		jm.handleJobFailure(job.ID, errorMsg)
		return
	}

	// Save analysis and mark job done atomically
	if err := jm.analysisJobRepo.SaveAnalysisAndMarkDone(jm.ctx, analysis, job.ID); err != nil {
		errorMsg := fmt.Sprintf("failed to save analysis and mark done: %v", err)
		logger.Log.Error(errorMsg, zap.Error(err), zap.String("job_id", job.ID.String()))
		jm.handleJobFailure(job.ID, errorMsg)
		return
	}

	jm.metrics.SuccessfulJobs.Add(1)
	jm.metrics.ProcessedJobs.Add(1)

	logger.Log.Info("Job completed successfully",
		zap.String("job_id", job.ID.String()),
		zap.String("lesson_plan_id", job.LessonPlanID.String()),
		zap.Int("worker_id", workerID),
		zap.Int("alignment_score", analysisResult.Analysis.AlignmentScore),
	)
}

// handleJobFailure handles job failure with retry logic
func (jm *JobManager) handleJobFailure(jobID uuid.UUID, errorMsg string) {
	if err := jm.analysisJobRepo.MarkJobFailed(jm.ctx, jobID, errorMsg, jm.config.MaxAttempts); err != nil {
		logger.Log.Error("failed to mark job as failed", zap.Error(err), zap.String("job_id", jobID.String()))
	}

	jm.metrics.FailedJobs.Add(1)
	jm.metrics.ProcessedJobs.Add(1)
}

// pollFallback periodically polls for pending jobs as a fallback when notifications fail.
// To avoid re-enqueueing the same job multiple times, this only pushes a "wake-up" signal
// to trigger worker pool to check for pending jobs, rather than pushing individual job IDs.
func (jm *JobManager) pollFallback() {
	defer jm.workerWg.Done()

	logger.Log.Info("Polling fallback started", zap.Duration("interval", jm.config.PollInterval))

	// Sentinel value to signal workers to check for pending jobs
	sentinelJobID := uuid.Nil

	for {
		select {
		case <-jm.ctx.Done():
			logger.Log.Info("Polling fallback stopped")
			return

		case <-jm.pollTicker.C:
			// Instead of fetching and enqueueing individual jobs (which can cause duplicates),
			// we send a "wake-up" signal to trigger worker pool to check for any pending jobs.
			// This way, FetchPendingJob with SELECT FOR UPDATE handles the actual claiming.
			select {
			case jm.jobChan <- sentinelJobID:
				logger.Log.Debug("polling signal sent to trigger workers")
			case <-jm.ctx.Done():
				return
			default:
				// Channel full, will retry on next poll
				logger.Log.Debug("polling signal dropped (channel full)")
			}
		}
	}
}

// GetMetrics returns current job metrics
func (jm *JobManager) GetMetrics() map[string]interface{} {
	jm.metrics.mu.RLock()
	defer jm.metrics.mu.RUnlock()

	stats, err := jm.analysisJobRepo.GetJobStatistics(jm.ctx)
	if err != nil {
		logger.Log.Error("failed to get job statistics", zap.Error(err))
		stats = map[string]int64{}
	}

	return map[string]interface{}{
		"processed_jobs":  jm.metrics.ProcessedJobs.Load(),
		"successful_jobs": jm.metrics.SuccessfulJobs.Load(),
		"failed_jobs":     jm.metrics.FailedJobs.Load(),
		"pending_jobs":    stats["pending"],
		"processing_jobs": stats["processing"],
		"done_jobs":       stats["done"],
		"failed_jobs_db":  stats["failed"],
	}
}

// IsRunning returns whether the job manager is currently running
func (jm *JobManager) IsRunning() bool {
	jm.mu.RLock()
	defer jm.mu.RUnlock()
	return jm.isRunning
}

func analysisResultToDomain(lessonPlanID uuid.UUID, result *secondary.AnalysisResult) (*domain.LessonPlanAnalysis, error) {
	if result == nil {
		return nil, fmt.Errorf("analysis result is required")
	}
	if lessonPlanID == uuid.Nil {
		return nil, fmt.Errorf("lesson_plan_id is required")
	}

	metadataJSON, err := json.Marshal(map[string]any{
		"title":       result.Metadata.Title,
		"subject":     result.Metadata.Subject,
		"grade_level": result.Metadata.GradeLevel,
		"objectives":  result.Metadata.Objectives,
		"bncc_skills": result.Metadata.BNCCSkills,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	suggestionsJSON, err := json.Marshal(map[string]any{
		"suggestions":      result.Analysis.Suggestions,
		"missing_elements": result.Analysis.MissingElements,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal suggestions: %w", err)
	}

	return &domain.LessonPlanAnalysis{
		ID:             uuid.New(),
		LessonPlanID:   lessonPlanID,
		Title:          result.Metadata.Title,
		Subject:        result.Metadata.Subject,
		GradeLevel:     result.Metadata.GradeLevel,
		AlignmentScore: result.Analysis.AlignmentScore,
		Feedback:       result.Analysis.PedagogicalFeedback,
		Metadata:       string(metadataJSON),
		Suggestions:    string(suggestionsJSON),
		CreatedAt:      time.Now(),
	}, nil
}

// GetJobByID fetches an analysis job by ID
func (jm *JobManager) GetJobByID(ctx context.Context, jobID uuid.UUID) (*domain.AnalysisJob, error) {
	return jm.analysisJobRepo.GetJobByID(ctx, jobID)
}
