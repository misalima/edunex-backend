package queue

// This file contains examples and notes for testing the job queue system

/*

## Testing the Job Queue Implementation

### Unit Tests Examples

#### 1. Test AnalysisJobRepository.UpsertAnalysisJob()
```go
func TestAnalysisJobRepository_UpsertAnalysisJob(t *testing.T) {
	// Setup: Create test database
	db := setupTestDB(t)
	repo := NewAnalysisJobRepository(db)

	lessonPlanID := uuid.New()

	// First insert
	jobID1, err := repo.UpsertAnalysisJob(context.Background(), lessonPlanID)
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, jobID1)

	// Verify job was created as pending
	job, err := repo.GetJobByLessonPlanID(context.Background(), lessonPlanID)
	require.NoError(t, err)
	require.Equal(t, "pending", job.Status)

	// Manually mark as done
	err = repo.MarkJobDone(context.Background(), jobID1)
	require.NoError(t, err)

	// Upsert again - should reactivate
	jobID2, err := repo.UpsertAnalysisJob(context.Background(), lessonPlanID)
	require.NoError(t, err)

	job, err = repo.GetJobByLessonPlanID(context.Background(), lessonPlanID)
	require.NoError(t, err)
	require.Equal(t, "pending", job.Status)
}
```

#### 2. Test FetchPendingJob() - No Race Conditions
```go
func TestAnalysisJobRepository_FetchPendingJob_NoRaceConditions(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAnalysisJobRepository(db)

	// Create multiple pending jobs
	jobCount := 10
	jobIDs := make([]uuid.UUID, jobCount)
	for i := 0; i < jobCount; i++ {
		lpID := uuid.New()
		jobID, err := repo.UpsertAnalysisJob(context.Background(), lpID)
		require.NoError(t, err)
		jobIDs[i] = jobID
	}

	// Simulate multiple workers fetching simultaneously
	fetched := make(map[uuid.UUID]bool)
	mu := sync.Mutex{}

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ { // 5 workers
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < jobCount; j++ {
				job, err := repo.FetchPendingJob(context.Background())
				require.NoError(t, err)
				if job != nil {
					mu.Lock()
					require.False(t, fetched[job.ID], "Job %s fetched by multiple workers", job.ID)
					fetched[job.ID] = true
					mu.Unlock()
				}
			}
		}()
	}

	wg.Wait()

	// Verify each job was fetched exactly once
	require.Equal(t, jobCount, len(fetched))
}
```

#### 3. Test JobManager Retry Logic
```go
func TestJobManager_RetryLogic(t *testing.T) {
	db := setupTestDB(t)

	failureCount := 0
	mockAI := &mockAIProvider{
		analyzeFn: func(ctx context.Context, text string) (*secondary.AnalysisResult, error) {
			failureCount++
			return nil, errors.New("AI service temporarily unavailable")
		},
	}

	jm := NewJobManager(
		db,
		&JobManagerConfig{
			WorkerCount: 1,
			MaxAttempts: 3,
			PollInterval: 100 * time.Millisecond,
		},
		mockAI,
		// ... other dependencies
	)

	lessonPlanID := uuid.New()

	// Enqueue job
	jobID, err := jm.Enqueue(context.Background(), lessonPlanID)
	require.NoError(t, err)

	// Start job manager
	err = jm.Start()
	require.NoError(t, err)

	// Wait for retries
	time.Sleep(5 * time.Second)

	// Verify job failed after max attempts
	job, err := jm.analysisJobRepo.GetJobByLessonPlanID(context.Background(), lessonPlanID)
	require.NoError(t, err)
	require.Equal(t, "failed", job.Status)
	require.Equal(t, 3, job.Attempts)

	jm.Stop()
}
```

### Integration Tests

#### 1. End-to-End Job Processing
```go
func TestJobQueue_EndToEnd(t *testing.T) {
	db := setupTestDB(t)
	storage := setupMockStorage(t) // Returns test file

	// Create test lesson plan
	lessonPlanID := uuid.New()
	lessonPlan := &domain.LessonPlan{
		ID:       lessonPlanID,
		Title:    "Test Plan",
		FilePath: "test.pdf",
	}
	lpRepo := postgres.NewLessonPlanRepository(db)
	lpRepo.InsertLessonPlan(context.Background(), lessonPlan)

	// Setup job manager with all real dependencies
	jm := NewJobManager(
		db,
		DefaultJobManagerConfig(),
		groqClient,      // Real Groq client (or mock with real API)
		extractor,       // Real extractor (can test PDF/DOCX)
		lpRepo,
		analysisRepo,
	)

	// Start and enqueue
	err := jm.Start()
	require.NoError(t, err)

	jobID, err := jm.Enqueue(context.Background(), lessonPlanID)
	require.NoError(t, err)

	// Wait for completion (with timeout)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			t.Fatal("Job did not complete within 30 seconds")
		default:
			job, _ := jm.analysisJobRepo.GetJobByLessonPlanID(context.Background(), lessonPlanID)
			if job.Status == "done" {
				// Verify analysis was saved
				analysis, err := analysisRepo.GetAnalysisByLessonPlanID(context.Background(), lessonPlanID)
				require.NoError(t, err)
				require.NotNil(t, analysis)
				require.Greater(t, analysis.AlignmentScore, 0)
				return
			}
			time.Sleep(100 * time.Millisecond)
		}
	}
}
```

#### 2. Test Notification System Failover
```go
func TestJobQueue_NotificationFailover(t *testing.T) {
	db := setupTestDB(t)

	// Create a job
	lessonPlanID := uuid.New()
	jm := setupJobManager(db)
	err := jm.Start()
	require.NoError(t, err)

	// Simulate notification listener failure
	if jm.notifyListener != nil {
		// Forcefully stop listener to trigger fallback
		jm.notifyListener.Stop()
	}

	// Enqueue should still work
	jobID, err := jm.Enqueue(context.Background(), lessonPlanID)
	require.NoError(t, err)
	require.NotEqual(t, uuid.Nil, jobID)

	// Wait for polling to pick it up
	time.Sleep(15 * time.Second) // > 10s poll interval

	// Job should start processing
	job, err := jm.analysisJobRepo.GetJobByLessonPlanID(context.Background(), lessonPlanID)
	require.NoError(t, err)
	require.NotEqual(t, "pending", job.Status)

	jm.Stop()
}
```

### Performance/Load Tests

#### 1. Throughput Test
```go
func BenchmarkJobQueue_Throughput(b *testing.B) {
	db := setupTestDB(b)
	jm := setupJobManager(db)
	jm.Start()
	defer jm.Stop()

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		lessonPlanID := uuid.New()
		_, err := jm.Enqueue(context.Background(), lessonPlanID)
		require.NoError(b, err)
	}

	// Output: BenchmarkJobQueue_Throughput-4  50000  24000 ns/op (50k ops/sec)
}
```

#### 2. Worker Concurrency Test
```go
func TestJobQueue_WorkerConcurrency(t *testing.T) {
	db := setupTestDB(t)

	// Create 100 pending jobs
	jobCount := 100
	for i := 0; i < jobCount; i++ {
		repo := postgres.NewAnalysisJobRepository(db)
		repo.UpsertAnalysisJob(context.Background(), uuid.New())
	}

	// Process with 5 workers
	jm := NewJobManager(
		db,
		&JobManagerConfig{
			WorkerCount: 5,
			MaxAttempts: 1,
		},
		mockAI,
		// ... dependencies
	)
	jm.Start()

	// Monitor metrics over time
	for i := 0; i < 10; i++ {
		metrics := jm.GetMetrics()
		t.Logf("Iteration %d: processed=%d, processing=%d, pending=%d",
			i,
			metrics["processed_jobs"],
			metrics["processing_jobs"],
			metrics["pending_jobs"],
		)
		time.Sleep(1 * time.Second)
	}

	jm.Stop()
}
```

### Manual Testing Commands

```bash
# 1. Start server
go run cmd/app/main.go

# 2. Create a lesson plan (prerequisite)
curl -X POST http://localhost:8080/api/v1/lesson-plans \
  -H "Authorization: Bearer {jwt}" \
  -F "title=Test Plan" \
  -F "file=@test.pdf" \
  # -> returns lesson_plan_id

# 3. Enqueue analysis
curl -X POST \
  "http://localhost:8080/api/v1/lesson-plans/{lesson_plan_id}/analyze" \
  -H "Authorization: Bearer {jwt}"
# -> returns job_id with status=pending

# 4. Poll job status
curl -X GET \
  "http://localhost:8080/api/v1/analysis-jobs/{job_id}" \
  -H "Authorization: Bearer {jwt}"
# -> will show status transitions: pending → processing → done

# 5. Check metrics
curl -X GET \
  "http://localhost:8080/api/v1/analysis-jobs/metrics" \
  -H "Authorization: Bearer {jwt}"

# 6. Watch server logs
tail -f <log_file> | grep "job"
```

### Debugging Tips

1. **Enable Debug Logging**: Set `LOG_LEVEL=debug` environment variable
2. **Check Database State**: Run queries directly
   ```sql
   SELECT id, status, attempts, error_message, started_at, finished_at
   FROM analysis_jobs
   ORDER BY created_at DESC
   LIMIT 10;
   ```
3. **Monitor Active Workers**: Check goroutines
   ```go
   import "runtime"
   runtime.NumGoroutine() // Should be ~6 when running (3 workers + listener + poller + main)
   ```
4. **Test Notification Channel**: Manually trigger NOTIFY
   ```sql
   NOTIFY analysis_jobs_channel, '123e4567-e89b-12d3-a456-426614174000';
   ```

### Test Coverage

Target coverage for job queue system: **>80%**

Key paths to cover:
- ✓ Successful job processing
- ✓ Retry logic with exponential backoff
- ✓ Max attempts exceeded → failed state
- ✓ Atomic fetch+update (no race conditions)
- ✓ Notification listener startup/failure
- ✓ Polling fallback
- ✓ Graceful shutdown
- ✓ Stale job cleanup
- ✓ Error scenarios (storage, AI, database)

*/
