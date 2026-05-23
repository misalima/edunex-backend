# Job Queue Processing System - Implementation Guide

## Overview

This document describes the background job processing system for lesson plan analysis. The system uses:
- **Worker Pool**: N concurrent workers processing analysis jobs
- **Postgres NOTIFY/LISTEN**: Instant job notifications with polling fallback
- **Atomic Transactions**: Race condition-free job state transitions
- **Graceful Shutdown**: Proper context cancellation and resource cleanup
- **Observability**: Metrics tracking and structured logging

## Architecture

### Components

#### 1. **AnalysisJobRepository** (`internal/infra/postgres/analysis_job_repository.go`)
Manages database operations for `analysis_jobs` table:
- `UpsertAnalysisJob()` - Insert or reactivate a job (ON CONFLICT handling)
- `FetchPendingJob()` - Atomically fetch and mark job as processing (SELECT FOR UPDATE SKIP LOCKED)
- `MarkJobDone()` - Transition to done state
- `MarkJobFailed()` - Handle failures with retry tracking and logic
- `CleanupStaleProcessingJobs()` - Recovery from crashed workers
- `GetJobStatistics()` - Metrics retrieval

#### 2. **NotificationListener** (`internal/infra/queue/notification_listener.go`)
Listens to Postgres NOTIFY events:
- Uses `pq.Listener` to receive real-time notifications
- Pushes job IDs to internal Go channel
- Automatic reconnection with backoff
- Graceful shutdown support

#### 3. **JobManager** (`internal/infra/queue/job_manager.go`)
Orchestrates the entire job processing system:
- Initializes worker pool (default: 3 workers)
- Starts notification listener (with polling fallback)
- Manages job lifecycle: enqueue → processing → done/failed
- Implements retry logic with immediate retries and tracking
- Tracks metrics (processed jobs, success rate, etc.)
- Graceful startup/shutdown

**Configuration** (DefaultJobManagerConfig):
- `WorkerCount`: 3 (number of concurrent workers)
- `MaxAttempts`: 3 (retry limit)
- `PollInterval`: 10 seconds (fallback polling for missed notifications)
- `StaleThreshold`: 30 minutes (mark stale processing jobs as pending)

#### 4. **AnalysisJobHandler** (`internal/api/handlers/analysis_job_handler.go`)
HTTP endpoints for job management:
- `POST /api/v1/lesson-plans/{lesson_plan_id}/analyze` - Enqueue analysis
- `GET /api/v1/analysis-jobs/{job_id}` - Get job status
- `GET /api/v1/analysis-jobs/metrics` - Job statistics

#### 5. **Extractor** (`internal/infra/extractor/extractor.go`)
Extractor with optional storage capability:
- Downloads files from Supabase storage using pre-configured `StorageClient`
- Extracts text content (PDF, DOCX, plain text) using unified parsing limits
- Implements `DataExtractor` interface

## Database Schema

```sql
-- Existing table (see config/postgres/init.sql)
CREATE TABLE analysis_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    lesson_plan_id UUID NOT NULL UNIQUE,
    status job_status_enum NOT NULL DEFAULT 'pending',
    attempts INT NOT NULL DEFAULT 0,
    error_message TEXT,
    created_at TIMESTAMP DEFAULT now(),
    started_at TIMESTAMP,
    finished_at TIMESTAMP,
    FOREIGN KEY (lesson_plan_id) REFERENCES lesson_plans(id) ON DELETE CASCADE
);

-- Indexes
CREATE INDEX idx_analysis_jobs_status ON analysis_jobs(status);
CREATE INDEX idx_analysis_jobs_created_at ON analysis_jobs(created_at);
```

## Job Lifecycle

```
┌─────────────┐
│   HTTP      │ POST /lesson-plans/{id}/analyze
│  Request    │
└──────┬──────┘
       │
       v
┌─────────────────────┐
│  Enqueue Operation  │
├─────────────────────┤
│ 1. INSERT/UPDATE on │
│    CONFLICT         │
│ 2. NOTIFY postgres  │
│ 3. Push to channel  │
└──────┬──────────────┘
       │
       ├─────────────────────────────────────┐
       │                                     │
       v (NOTIFY)                     v (Polling)
   ┌───────────┐                ┌──────────────┐
   │ Listener  │                │ Poll Timer   │
   │ (Real-    │                │ (Every 10s)  │
   │  time)    │                └──────┬───────┘
   └─────┬─────┘                       │
       │                               │
       └───────────┬───────────────────┘
                   │
                   v
         ┌──────────────────────┐
         │  Job Channel         │
         │  (Internal Queue)    │
         └──────────┬───────────┘
                    │
        ┌───────────┼───────────┬─────────────┐
        │           │           │             │
        v           v           v             v
    ┌──────┐   ┌──────┐   ┌──────┐    ┌──────────┐
    │Worker│   │Worker│   │Worker│    │Fallback  │
    │  #1  │   │  #2  │   │  #3  │    │Polling   │
    └──┬───┘   └──┬───┘   └──┬───┘    └──────────┘
       │          │          │
       └──────────┼──────────┘
                  │
                  v
      ┌─────────────────────────────┐
      │ FetchPendingJob (atomic)    │
      │ SELECT ... FOR UPDATE       │
      │ SKIP LOCKED                 │
      │ + UPDATE status=processing  │
      └──────────┬──────────────────┘
                 │
                 v
      ┌──────────────────────────┐
      │ Extract from Storage     │
      │ (Download file)          │
      │ + Extract text           │
      └──────────┬───────────────┘
                 │
                 v
      ┌──────────────────────────┐
      │ AI Analysis (Groq API)   │
      └──────────┬───────────────┘
                 │
                 v
      ┌──────────────────────────┐
      │ Save Analysis Result     │
      │ (lesson_plan_analyses)   │
      └──────────┬───────────────┘
                 │
       ┌─────────┴────────────────┐
       │ Success?                 │
       │                          │
    YES│                       NO │ ERROR
       │                          │
       v                          v
   ┌────────┐           ┌──────────────────┐
   │ DONE   │           │ MarkJobFailed    │
   │ State  │           │ ├─ attempts++    │
   │        │           │ ├─ save error    │
   └────────┘           │ └─ retry logic   │
                        │    (if < max     │
                        │     attempts)    │
                        └──────────────────┘
                               │
                    ┌──────────┴─────────┐
                    │                    │
                 RETRY             FAILED
                    │                    │
                    v                    v
                ┌────────┐          ┌────────┐
                │PENDING │          │ FAILED │
                │ State  │          │ State  │
                └────────┘          └────────┘
```

## Job Status Transitions

| Current | Action                  | Next      | Notes |
|---------|-------------------------|-----------|-------|
| PENDING | Worker fetches job      | PROCESSING| SELECT FOR UPDATE |
| PROCESSING | Analysis succeeds    | DONE      | Analysis saved |
| PROCESSING | Analysis fails       | PENDING*  | attempts++ if < max_attempts |
| PROCESSING | Analysis fails       | FAILED    | attempts >= max_attempts |
| DONE    | (terminal)              | -         | Job complete |
| FAILED  | (terminal)              | -         | All retries exhausted |

*Job transitions back to PENDING for retry

## Atomic Operations

### 1. FetchPendingJob (SELECT FOR UPDATE SKIP LOCKED)
```go
BEGIN;
  SELECT * FROM analysis_jobs
  WHERE status = 'pending'
  ORDER BY created_at ASC
  LIMIT 1
  FOR UPDATE SKIP LOCKED;
  
  UPDATE analysis_jobs
  SET status = 'processing', started_at = now()
  WHERE id = ?;
COMMIT;
```
**Benefits:**
- Prevents race conditions between workers
- `SKIP LOCKED` ensures only one worker gets a job
- Atomic read + update in same transaction

### 2. MarkJobFailed (Retry Logic)
```go
BEGIN;
  SELECT attempts FROM analysis_jobs WHERE id = ?;
  
  UPDATE analysis_jobs
  SET 
    attempts = attempts + 1,
    status = IF(attempts + 1 >= max_attempts, 'failed', 'pending'),
    finished_at = now(),
    error_message = ?,
    started_at = NULL (if retrying)
  WHERE id = ?;
COMMIT;
```

## Notification System

### Real-time: Postgres NOTIFY/LISTEN
1. Enqueue endpoint calls `NOTIFY analysis_jobs_channel, '{job_id}'`
2. NotificationListener constantly receives notifications
3. Pushes job IDs to internal Go channel
4. Workers consume from channel

### Fallback: Periodic Polling
- If notification listener fails: periodic polling (every 10 seconds)
- Polls `SELECT id FROM analysis_jobs WHERE status = 'pending'`
- Ensures jobs are never lost

### Error Handling on Startup
- CleanupStaleProcessingJobs checks for jobs stuck in processing state
- If `started_at` older than 30 minutes → mark as pending or failed
- Allows recovery from crashed worker instances

## Configuration & Integration

### 1. Environment Variables
```bash
# Required for AI analysis
GROQ_API_KEY=xxxxx

# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=xxxxx
DB_NAME=edunex

# Storage
SUPABASE_URL=https://xxxxx.supabase.co
SUPABASE_SERVICE_ROLE_KEY=xxxxx
SUPABASE_BUCKET=lesson-plans

# Existing
SUPABASE_ANON_KEY=xxxxx
SUPABASE_JWT_K_X=xxxxx
SUPABASE_JWT_K_Y=xxxxx
```

### 2. Initialization (main.go)
```go
// Create container first
ctn := container.NewContainer(db, cfg)

// Start job manager BEFORE HTTP server
jobManager := ctn.GetJobManager()
if err := jobManager.Start(); err != nil {
    log.Fatalf("failed to start job manager: %v", err)
}

// Start HTTP server...

// On shutdown: stop job manager FIRST
if err := jobManager.Stop(); err != nil {
    log.Printf("error stopping job manager: %v", err)
}
```

### 3. Routes Registered
```bash
POST   /api/v1/lesson-plans/{lesson_plan_id}/analyze  # Enqueue
GET    /api/v1/analysis-jobs/{job_id}                  # Status
GET    /api/v1/analysis-jobs/metrics                   # Metrics
```

## Examples

### 1. Enqueue Analysis
```bash
curl -X POST \
  http://localhost:8080/api/v1/lesson-plans/550e8400-12ab-34cd-e8fg-761234567890/analyze \
  -H "Authorization: Bearer {jwt_token}"

# Response (202 Accepted)
{
  "job_id": "660f9500-23bc-45de-f9gh-872345678901",
  "lesson_plan_id": "550e8400-12ab-34cd-e8fg-761234567890",
  "status": "pending"
}
```

### 2. Get Job Status
```bash
curl -X GET \
  http://localhost:8080/api/v1/analysis-jobs/660f9500-23bc-45de-f9gh-872345678901 \
  -H "Authorization: Bearer {jwt_token}"

# Response
{
  "job_id": "660f9500-23bc-45de-f9gh-872345678901",
  "status": "processing"
}
```

### 3. Get Metrics
```bash
curl -X GET \
  http://localhost:8080/api/v1/analysis-jobs/metrics \
  -H "Authorization: Bearer {jwt_token}"

# Response
{
  "processed_jobs": 42,
  "successful_jobs": 40,
  "failed_jobs": 2,
  "pending_jobs": 3,
  "processing_jobs": 1,
  "done_jobs": 40,
  "failed_jobs_db": 2
}
```

## Observability

### Logging
- Structured logging with Zap
- Log levels: Debug, Info, Warn, Error
- Includes: job_id, lesson_plan_id, worker_id, error details, timing

### Metrics Tracked
- `ProcessedJobs`: Total jobs processed
- `SuccessfulJobs`: Completed successfully
- `FailedJobs`: Failed (all retries exhausted)
- `PendingJobs`: Waiting to be processed
- `ProcessingJobs`: Currently being processed
- `DoneJobs`: Completed and saved
- `FailedJobsDB`: Failed in database

### Key Log Messages
```
level=info msg="Job manager started" worker_count=3 max_attempts=3
level=info msg="Notification listener started"
level=info msg="Worker started" worker_id=0
level=debug msg="job enqueued" job_id=xxx lesson_plan_id=yyy
level=info msg="Processing job" job_id=xxx worker_id=0
level=info msg="Job completed successfully" job_id=xxx alignment_score=85
level=error msg="failed to extract content" job_id=xxx error=xxx
level=info msg="Stopping job manager..."
level=info msg="All workers stopped gracefully"
```

## Graceful Shutdown

On termination signal (SIGTERM/SIGINT):
1. Stop accepting new HTTP requests
2. Stop job manager (first priority)
   - Cancel worker goroutines
   - Wait for running jobs to finish (max 10s timeout)
   - Stop notification listener
   - Stop polling timer
3. Close database connections
4. Server shutdown complete

## Failure Scenarios & Recovery

### Scenario 1: Notification Listener Fails
- Notification listener reports error
- Falls back to periodic polling
- No jobs lost; polling catches them

### Scenario 2: Worker Crashes While Processing
- Job stuck in `processing` state
- On next startup: `CleanupStaleProcessingJobs` marks as pending
- Job gets reprocessed

### Scenario 3: Storage Download Fails
- Job marked as failed with error message
- If `attempts < max_attempts`: marked pending for retry
- If `attempts >= max_attempts`: marked failed permanently

### Scenario 4: AI Analysis Fails (rate limit/API down)
- Error message saved in `error_message` field
- Immediate retry with tracking
- Max 3 attempts by default
- After max attempts: marked as failed

### Scenario 5: Database Connection Lost
- Notification listener fails gracefully
- Falls back to polling
- Workers retry database operations with context timeout
- On reconnection: normal operation resumes

## Performance Considerations

### Throughput
- Default: 3 workers × 10 seconds = ~6.4 jobs/minute with ideal network/AI
- Bottleneck: AI API response time (~10-30 seconds)
- Can increase `WorkerCount` for higher throughput

### Resource Usage
- Memory: ~1-2 MB per worker goroutine
- CPU: Idle waiting for I/O (network, database, AI API)
- Database connections: 1 per worker (already in pool)

### Backpressure
- Job channel buffered with `WorkerCount × 2`
- Enqueue endpoint non-blocking (notifications fallback to polling)
- HTTP requests return immediately (202 Accepted)

## Future Improvements

1. **Job Priorities**: Support high/medium/low priority jobs with separate queues
2. **Scheduled Jobs**: Support scheduling analysis for future dates
3. **Batch Operations**: Analyze multiple lesson plans in one request
4. **Webhooks**: Notify external systems when jobs complete
5. **Job History**: Archiving successful jobs after 30 days
6. **Dead Letter Queues**: Route permanently failed jobs for manual review
7. **Distributed Architecture**: Multiple instances with shared job queue (via Postgres)
8. **Metrics Export**: Prometheus integration for Grafana dashboards

