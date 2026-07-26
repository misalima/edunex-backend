package queue

import (
	"context"
	"errors"
	"io"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/misalima/edunex-backend/internal/core/domain"
	"github.com/misalima/edunex-backend/internal/core/interfaces/secondary"
	"github.com/misalima/edunex-backend/internal/infra/logger"
	"go.uber.org/zap"
)

func init() {
	logger.Log = zap.NewNop()
}

// ============================================================================
// MANUAL MOCK STRUCTS FOR PORTS AND INTERFACES
// ============================================================================

type mockAIProvider struct {
	analyzeFn func(ctx context.Context, text string) (*secondary.AnalysisResult, error)
}

func (m *mockAIProvider) Analyze(ctx context.Context, text string) (*secondary.AnalysisResult, error) {
	if m.analyzeFn != nil {
		return m.analyzeFn(ctx, text)
	}
	res := &secondary.AnalysisResult{}
	res.Metadata.Title = "Mock Lesson Plan"
	res.Metadata.Subject = "Mathematics"
	res.Metadata.GradeLevel = "High School"
	res.Analysis.AlignmentScore = 85
	res.Analysis.PedagogicalFeedback = "Good alignment."
	return res, nil
}

type mockDataExtractor struct {
	extractStorageFn func(ctx context.Context, objectPath string) (string, error)
}

func (m *mockDataExtractor) ExtractText(ctx context.Context, r io.Reader, contentType string) (*secondary.ExtractionResult, error) {
	return &secondary.ExtractionResult{Text: "Mock Extracted Text"}, nil
}

func (m *mockDataExtractor) ExtractFromStorage(ctx context.Context, objectPath string) (string, error) {
	if m.extractStorageFn != nil {
		return m.extractStorageFn(ctx, objectPath)
	}
	return "Mock Extracted Text from Storage", nil
}

type mockLessonPlanLoader struct {
	getByIDFn func(ctx context.Context, id uuid.UUID) (*domain.LessonPlan, error)
}

func (m *mockLessonPlanLoader) InsertLessonPlan(ctx context.Context, lp *domain.LessonPlan) (uuid.UUID, error) {
	return uuid.New(), nil
}

func (m *mockLessonPlanLoader) GetLessonPlanByID(ctx context.Context, id uuid.UUID) (*domain.LessonPlan, error) {
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id)
	}
	return &domain.LessonPlan{
		ID:       id,
		Title:    "Mock Lesson Plan",
		FilePath: "mock.pdf",
	}, nil
}

func (m *mockLessonPlanLoader) ListLessonPlans(ctx context.Context) ([]*domain.LessonPlan, error) {
	return nil, nil
}

func (m *mockLessonPlanLoader) UpdateLessonPlan(ctx context.Context, lp *domain.LessonPlan) error {
	return nil
}

func (m *mockLessonPlanLoader) DeleteLessonPlan(ctx context.Context, id uuid.UUID) error {
	return nil
}

type mockLessonPlanAnalysisLoader struct {
	saveFn func(ctx context.Context, analysis *domain.LessonPlanAnalysis) error
}

func (m *mockLessonPlanAnalysisLoader) InsertAnalysis(ctx context.Context, analysis *domain.LessonPlanAnalysis) error {
	if m.saveFn != nil {
		return m.saveFn(ctx, analysis)
	}
	return nil
}

func (m *mockLessonPlanAnalysisLoader) GetAnalysisByLessonPlanID(ctx context.Context, lessonPlanID uuid.UUID) (*domain.LessonPlanAnalysis, error) {
	return nil, nil
}

type mockAnalysisJobLoader struct {
	upsertFn      func(ctx context.Context, lessonPlanID uuid.UUID) (uuid.UUID, error)
	fetchFn       func(ctx context.Context) (*domain.AnalysisJob, error)
	fetchByIDFn   func(ctx context.Context, jobID uuid.UUID) (*domain.AnalysisJob, error)
	markDoneFn    func(ctx context.Context, jobID uuid.UUID) error
	markFailedFn  func(ctx context.Context, jobID uuid.UUID, errorMsg string, maxAttempts int) error
	getJobByLPFn  func(ctx context.Context, lessonPlanID uuid.UUID) (*domain.AnalysisJob, error)
	cleanupFn     func(ctx context.Context, staleThreshold time.Duration) error
	getStatsFn    func(ctx context.Context) (map[string]int64, error)
	saveAndDoneFn func(ctx context.Context, analysis *domain.LessonPlanAnalysis, jobID uuid.UUID) error
	getJobByIDFn  func(ctx context.Context, jobID uuid.UUID) (*domain.AnalysisJob, error)
}

func (m *mockAnalysisJobLoader) UpsertAnalysisJob(ctx context.Context, lessonPlanID uuid.UUID) (uuid.UUID, error) {
	if m.upsertFn != nil {
		return m.upsertFn(ctx, lessonPlanID)
	}
	return uuid.New(), nil
}

func (m *mockAnalysisJobLoader) FetchPendingJob(ctx context.Context) (*domain.AnalysisJob, error) {
	if m.fetchFn != nil {
		return m.fetchFn(ctx)
	}
	return nil, nil
}

func (m *mockAnalysisJobLoader) FetchPendingJobByID(ctx context.Context, jobID uuid.UUID) (*domain.AnalysisJob, error) {
	if m.fetchByIDFn != nil {
		return m.fetchByIDFn(ctx, jobID)
	}
	return nil, nil
}

func (m *mockAnalysisJobLoader) MarkJobDone(ctx context.Context, jobID uuid.UUID) error {
	if m.markDoneFn != nil {
		return m.markDoneFn(ctx, jobID)
	}
	return nil
}

func (m *mockAnalysisJobLoader) MarkJobFailed(ctx context.Context, jobID uuid.UUID, errorMsg string, maxAttempts int) error {
	if m.markFailedFn != nil {
		return m.markFailedFn(ctx, jobID, errorMsg, maxAttempts)
	}
	return nil
}

func (m *mockAnalysisJobLoader) GetJobByLessonPlanID(ctx context.Context, lessonPlanID uuid.UUID) (*domain.AnalysisJob, error) {
	if m.getJobByLPFn != nil {
		return m.getJobByLPFn(ctx, lessonPlanID)
	}
	return nil, nil
}

func (m *mockAnalysisJobLoader) CleanupStaleProcessingJobs(ctx context.Context, staleThreshold time.Duration) error {
	if m.cleanupFn != nil {
		return m.cleanupFn(ctx, staleThreshold)
	}
	return nil
}

func (m *mockAnalysisJobLoader) GetJobStatistics(ctx context.Context) (map[string]int64, error) {
	if m.getStatsFn != nil {
		return m.getStatsFn(ctx)
	}
	return map[string]int64{"pending": 0, "processing": 0, "done": 0, "failed": 0}, nil
}

func (m *mockAnalysisJobLoader) SaveAnalysisAndMarkDone(ctx context.Context, analysis *domain.LessonPlanAnalysis, jobID uuid.UUID) error {
	if m.saveAndDoneFn != nil {
		return m.saveAndDoneFn(ctx, analysis, jobID)
	}
	return nil
}

func (m *mockAnalysisJobLoader) GetJobByID(ctx context.Context, jobID uuid.UUID) (*domain.AnalysisJob, error) {
	if m.getJobByIDFn != nil {
		return m.getJobByIDFn(ctx, jobID)
	}
	return nil, nil
}

// ============================================================================
// ANALYSIS PIPELINE TEST CASES
// ============================================================================

func TestJobManager_SuccessfulPipeline(t *testing.T) {
	// Setup dependencies
	ai := &mockAIProvider{}
	ext := &mockDataExtractor{}
	lpLoader := &mockLessonPlanLoader{}
	analysisLoader := &mockLessonPlanAnalysisLoader{}
	jobLoader := &mockAnalysisJobLoader{}

	config := &JobManagerConfig{
		WorkerCount:    1,
		MaxAttempts:    3,
		PollInterval:   10 * time.Second,
		StaleThreshold: 30 * time.Minute,
	}

	jm := NewJobManager(nil, "", config, ai, ext, lpLoader, analysisLoader, jobLoader)

	jobID := uuid.New()
	lessonPlanID := uuid.New()

	job := &domain.AnalysisJob{
		ID:           jobID,
		LessonPlanID: lessonPlanID,
		Status:       "processing",
		Attempts:     0,
	}

	var saveCalled int32
	jobLoader.saveAndDoneFn = func(ctx context.Context, analysis *domain.LessonPlanAnalysis, jID uuid.UUID) error {
		atomic.AddInt32(&saveCalled, 1)
		if jID != jobID {
			t.Errorf("expected job ID %v, got %v", jobID, jID)
		}
		if analysis.LessonPlanID != lessonPlanID {
			t.Errorf("expected lesson plan ID %v, got %v", lessonPlanID, analysis.LessonPlanID)
		}
		if analysis.Title != "Mock Lesson Plan" {
			t.Errorf("expected title 'Mock Lesson Plan', got '%s'", analysis.Title)
		}
		if analysis.AlignmentScore != 85 {
			t.Errorf("expected alignment score 85, got %d", analysis.AlignmentScore)
		}
		return nil
	}

	// Trigger execution pipeline directly
	jm.processJobWithData(job, 1)

	if atomic.LoadInt32(&saveCalled) != 1 {
		t.Error("expected SaveAnalysisAndMarkDone to be called exactly once")
	}

	if jm.metrics.SuccessfulJobs.Load() != 1 {
		t.Errorf("expected successful jobs metric to be 1, got %d", jm.metrics.SuccessfulJobs.Load())
	}
	if jm.metrics.ProcessedJobs.Load() != 1 {
		t.Errorf("expected processed jobs metric to be 1, got %d", jm.metrics.ProcessedJobs.Load())
	}
}

func TestJobManager_PipelineExtractorFailure(t *testing.T) {
	// Setup dependencies
	ai := &mockAIProvider{}
	lpLoader := &mockLessonPlanLoader{}
	analysisLoader := &mockLessonPlanAnalysisLoader{}
	jobLoader := &mockAnalysisJobLoader{}

	// Force extractor to fail
	ext := &mockDataExtractor{
		extractStorageFn: func(ctx context.Context, objectPath string) (string, error) {
			return "", errors.New("supabase storage download failed")
		},
	}

	config := &JobManagerConfig{
		WorkerCount:    1,
		MaxAttempts:    3,
		PollInterval:   10 * time.Second,
		StaleThreshold: 30 * time.Minute,
	}

	jm := NewJobManager(nil, "", config, ai, ext, lpLoader, analysisLoader, jobLoader)

	jobID := uuid.New()
	lessonPlanID := uuid.New()

	job := &domain.AnalysisJob{
		ID:           jobID,
		LessonPlanID: lessonPlanID,
		Status:       "processing",
		Attempts:     0,
	}

	var failCalled int32
	jobLoader.markFailedFn = func(ctx context.Context, jID uuid.UUID, errMsg string, maxAtts int) error {
		atomic.AddInt32(&failCalled, 1)
		if jID != jobID {
			t.Errorf("expected job ID %v, got %v", jobID, jID)
		}
		if maxAtts != 3 {
			t.Errorf("expected max attempts 3, got %d", maxAtts)
		}
		return nil
	}

	// Trigger execution pipeline
	jm.processJobWithData(job, 1)

	if atomic.LoadInt32(&failCalled) != 1 {
		t.Error("expected MarkJobFailed to be called exactly once")
	}

	if jm.metrics.FailedJobs.Load() != 1 {
		t.Errorf("expected failed jobs metric to be 1, got %d", jm.metrics.FailedJobs.Load())
	}
	if jm.metrics.ProcessedJobs.Load() != 1 {
		t.Errorf("expected processed jobs metric to be 1, got %d", jm.metrics.ProcessedJobs.Load())
	}
}

func TestNormalizeConnString_AddsDisableSslMode(t *testing.T) {
	conn := "postgresql://postgres:secret@localhost:5432/edunex"

	got := normalizeConnString(conn)
	parsed, err := url.Parse(got)
	if err != nil {
		t.Fatalf("failed to parse normalized connection string: %v", err)
	}

	if parsed.Query().Get("sslmode") != "disable" {
		t.Fatalf("expected sslmode=disable, got %q", parsed.Query().Get("sslmode"))
	}
}

func TestNormalizeConnString_PreservesExistingSslMode(t *testing.T) {
	conn := "postgresql://postgres:secret@localhost:5432/edunex?sslmode=require"

	got := normalizeConnString(conn)
	parsed, err := url.Parse(got)
	if err != nil {
		t.Fatalf("failed to parse normalized connection string: %v", err)
	}

	if parsed.Query().Get("sslmode") != "require" {
		t.Fatalf("expected sslmode=require, got %q", parsed.Query().Get("sslmode"))
	}
}

func TestJobManager_PipelineAIFailure(t *testing.T) {
	// Setup dependencies
	ext := &mockDataExtractor{}
	lpLoader := &mockLessonPlanLoader{}
	analysisLoader := &mockLessonPlanAnalysisLoader{}
	jobLoader := &mockAnalysisJobLoader{}

	// Force AI to fail
	ai := &mockAIProvider{
		analyzeFn: func(ctx context.Context, text string) (*secondary.AnalysisResult, error) {
			return nil, errors.New("Groq API rate limit exceeded")
		},
	}

	config := &JobManagerConfig{
		WorkerCount:    1,
		MaxAttempts:    3,
		PollInterval:   10 * time.Second,
		StaleThreshold: 30 * time.Minute,
	}

	jm := NewJobManager(nil, "", config, ai, ext, lpLoader, analysisLoader, jobLoader)

	jobID := uuid.New()
	lessonPlanID := uuid.New()

	job := &domain.AnalysisJob{
		ID:           jobID,
		LessonPlanID: lessonPlanID,
		Status:       "processing",
		Attempts:     0,
	}

	var failCalled int32
	jobLoader.markFailedFn = func(ctx context.Context, jID uuid.UUID, errMsg string, maxAtts int) error {
		atomic.AddInt32(&failCalled, 1)
		if jID != jobID {
			t.Errorf("expected job ID %v, got %v", jobID, jID)
		}
		return nil
	}

	// Trigger execution pipeline
	jm.processJobWithData(job, 1)

	if atomic.LoadInt32(&failCalled) != 1 {
		t.Error("expected MarkJobFailed to be called exactly once")
	}

	if jm.metrics.FailedJobs.Load() != 1 {
		t.Errorf("expected failed jobs metric to be 1, got %d", jm.metrics.FailedJobs.Load())
	}
}

func TestJobManager_StartStopLifecycle(t *testing.T) {
	ai := &mockAIProvider{}
	ext := &mockDataExtractor{}
	lpLoader := &mockLessonPlanLoader{}
	analysisLoader := &mockLessonPlanAnalysisLoader{}
	jobLoader := &mockAnalysisJobLoader{}

	config := &JobManagerConfig{
		WorkerCount:    2,
		MaxAttempts:    3,
		PollInterval:   100 * time.Millisecond,
		StaleThreshold: 30 * time.Minute,
	}

	var cleanupCalled int32
	jobLoader.cleanupFn = func(ctx context.Context, staleThreshold time.Duration) error {
		atomic.StoreInt32(&cleanupCalled, 1)
		return nil
	}

	jm := NewJobManager(nil, "", config, ai, ext, lpLoader, analysisLoader, jobLoader)

	// Verify manager starts cleanly
	err := jm.Start()
	if err != nil {
		t.Fatalf("expected no error starting JobManager, got: %v", err)
	}

	if !jm.IsRunning() {
		t.Error("expected manager to be running after Start()")
	}

	// Let the ticker tick once or twice
	time.Sleep(150 * time.Millisecond)

	// Verify manager stops cleanly
	err = jm.Stop()
	if err != nil {
		t.Fatalf("expected no error stopping JobManager, got: %v", err)
	}

	if jm.IsRunning() {
		t.Error("expected manager not to be running after Stop()")
	}

	if atomic.LoadInt32(&cleanupCalled) != 1 {
		t.Error("expected CleanupStaleProcessingJobs to be called during startup")
	}
}
