# Analysis Pipeline: Future Robustness Improvements

This document outlines recommended enhancements to the analysis job pipeline for production readiness. These improvements are deferred from MVP but should be implemented as the system scales to support multiple concurrent workers and higher job volumes.

---

## 1. Max Retry Attempts with Permanent Failure Handling

**Current State:**  
- `MarkJobFailed` increments `attempts` counter but has no upper limit
- Failed jobs can theoretically be retried indefinitely

**Improvement:**  
- Set a configurable max retry threshold (recommended: 3 attempts)
- After exceeding max attempts, transition job to `ABANDONED` or `PERMANENTLY_FAILED` state
- Log and alert on jobs that exceed retry limits

**Implementation Details:**
- Add column `max_attempts` to `analysis_jobs` table (default: 3)
- Modify `MarkJobFailed` to check `attempts >= max_attempts` before allowing retry
- Create a UI/admin endpoint to inspect abandoned jobs for manual investigation
- Add metrics to track jobs by final status (done, failed, abandoned)

**Benefits:**
- Prevents infinite retry loops
- Enables manual intervention for problematic jobs
- Clear visibility into system health

---

## 2. Dead Letter Queue / Inspection Interface for Failed Jobs

**Current State:**  
- Failed jobs are marked with error messages but no easy way to inspect or manually retry

**Improvement:**  
- Create a separate view/query for failed/abandoned jobs with filtering and sorting
- Implement admin endpoint `GET /admin/analysis-jobs/failed` with pagination
- Allow selective manual retry: `POST /admin/analysis-jobs/{id}/retry`
- Track retry history (who retried, when, result)

**Implementation Details:**
- Add `last_retry_by` and `last_retry_at` fields to `analysis_jobs` table
- Create a `JobFailureAnalysis` domain type capturing context (error, lesson plan details, etc.)
- Implement filtering by:
  - Status (failed, abandoned)
  - Date range
  - Lesson plan ID
  - Error type / substring
- Expose summary statistics: total failed, by error type, success rate after manual retry

**Benefits:**
- Operational visibility and debugging
- Manual recovery path for transient failures
- Audit trail for regulatory compliance

---

## 3. Transaction Timeout & Lock Release on Worker Stall

**Current State:**  
- If a worker crashes after claiming a job, the row remains locked in `processing` state
- Other workers cannot reclaim it without manual intervention

**Improvement:**  
- Add `claimed_at` timestamp when transitioning to `processing`
- Implement a lock reaper process or periodic cleanup job
- Unclaim jobs that have been `processing` for > threshold (e.g., 30 minutes without heartbeat)
- Transition stalled jobs back to `pending` with error note

**Implementation Details:**
- Add `claimed_at` TIMESTAMP column to `analysis_jobs`
- Create a cleanup job that runs every 5 minutes:
  ```sql
  UPDATE analysis_jobs SET status = 'pending', attempts = attempts + 1, 
    error_message = 'Reclaimed: previous worker timeout'
  WHERE status = 'processing' AND claimed_at < now() - INTERVAL '30 minutes'
  ```
- Add configurable `ANALYSIS_JOB_PROCESSING_TIMEOUT_MINUTES` environment variable
- Log all reclaimed jobs for debugging

**Benefits:**
- Self-healing system tolerant of worker crashes
- No manual intervention required
- Bounded resource usage

---

## 4. Graceful Shutdown and Worker Lifecycle Management

**Current State:**  
- Worker runs indefinitely; no clean shutdown mechanism
- Jobs in progress may be lost on process termination

**Improvement:**  
- Implement graceful shutdown using Go context cancellation
- Upon shutdown signal (SIGTERM/SIGINT):
  1. Stop accepting new jobs
  2. Wait for in-progress jobs to complete (with timeout)
  3. Return claimed jobs to `pending` if timeout exceeded
  4. Log shutdown event with job status summary
- Implement heartbeat mechanism to signal worker liveness

**Implementation Details:**
- Add `worker_id` to `analysis_jobs` table (generated per worker instance)
- Create `WorkerRegistry` to track active workers (in-memory or Redis)
- Use Go's `signal.Notify()` to catch SIGTERM
- Implement graceful drain pattern:
  ```go
  select {
    case <-ctx.Done():
      // Shutdown signal received
      returnClaimedJobsToPending()
      close(jobChannel)
      return
    case job := <-jobChannel:
      processJob(job)
  }
  ```
- Add `--shutdown-timeout` flag (default: 60s)

**Benefits:**
- Zero job loss on restart
- Clean resource cleanup
- Production-grade operational practices

---

## 5. Observability: Logging, Metrics, and Tracing

**Current State:**  
- Basic error logging but no performance metrics or structured traces
- Difficult to diagnose bottlenecks or correlate failures

**Improvement:**  
- Add structured logging with correlation IDs (trace IDs) for end-to-end tracking
- Export Prometheus metrics for job pipeline health
- Implement distributed tracing (optional: OpenTelemetry)
- Create dashboards for job throughput, latency, and error rates

**Implementation Details:**

#### Metrics (Prometheus)
- `analysis_jobs_total{status, worker_id}` — Counter of jobs by final status
- `analysis_jobs_duration_seconds{status}` — Histogram of processing duration
- `analysis_jobs_current{status}` — Gauge of jobs in each state
- `analysis_jobs_attempts_total` — Counter of retry attempts
- `analysis_job_claiming_duration_seconds` — Histogram of `ClaimPendingJobs` latency

#### Logging
- Log at key lifecycle points:
  - Job claimed (with timestamp, worker_id)
  - Processing started/completed
  - Retry decisions
  - Failures with context (error type, attempt count)
- Use structured logging (JSON) with fields:
  - `trace_id` — Correlate across services
  - `job_id`, `lesson_plan_id`, `user_id`
  - `duration_ms`, `attempt`, `status`

#### Tracing (Optional)
- Use OpenTelemetry to trace:
  - `ClaimPendingJobs` duration
  - File extraction time
  - LLM API call latency
  - Database write latency
- Export to Jaeger or Datadog for visualization

#### Dashboards
- Job throughput (jobs/min by status)
- P50/P95/P99 latencies
- Error rate by error type
- Worker utilization (jobs claimed per worker)
- System health (pending queue depth, abandoned jobs count)

**Benefits:**
- Production-grade observability
- Data-driven optimization decisions
- Faster incident response and root cause analysis

---

## Implementation Priority

### Phase 1 (Critical for multi-worker)
1. **Max Retry Attempts** — Prevents runaway retries
2. **Lock Release on Timeout** — Handles worker crashes gracefully
3. **Graceful Shutdown** — Ensures clean deployments

### Phase 2 (Recommended for production)
4. **Dead Letter Queue Interface** — Operational visibility
5. **Observability** — Monitoring and debugging

---

## Related Documents

- [Analysis Domain Model](../architecture/ANALYSIS_DOMAIN_MODEL.md) — Domain entities for jobs and analysis
- [Architecture.md](../architecture/ARCHITECTURE.md) — ADR-06 on job claiming with SELECT FOR UPDATE SKIP LOCKED

---

## Notes for Implementation Team

- Keep retry logic simple; resist over-engineering fault tolerance before observability exists
- Use feature flags for gradual rollout of improvements
- Add integration tests for failure scenarios (simulate worker crash, network timeout, etc.)
- Monitor improvements through Prometheus metrics before declaring success

