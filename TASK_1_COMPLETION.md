# Task 1: Define the Analysis Domain Model — COMPLETED ✅

## Summary
Successfully established the core business entities for the AI-powered lesson plan analysis pipeline. All domain models are decoupled from infrastructure concerns and follow the architecture specification.

## Files Created/Modified

### New Files
1. **`internal/core/interfaces/secondary/analysis_repository.go`**
   - Defines two repository interfaces for persistence contracts
   - `AnalysisJobRepository`: CRUD operations for analysis jobs
   - `LessonPlanAnalysisRepository`: CRUD operations for analysis results
   - Enables loose coupling between core and infrastructure layers

2. **`internal/core/domain/ANALYSIS_DOMAIN_MODEL.md`**
   - Documentation of the domain model architecture
   - Explains entity relationships and design decisions

### Modified Files
1. **`internal/core/domain/lesson_plan_analysis.go`**
   - **Added:** `AnalysisMetadata` struct (title, subject, objectives, BNCC skills)
   - **Added:** `PedagogicalAnalysis` struct (feedback, alignment score, suggestions, missing elements)
   - **Added:** `AnalysisResult` struct (combines metadata + analysis per sprint contract)
   - **Enhanced:** `LessonPlanAnalysis` now includes both raw (`AnalysisText`) and structured (`StructuredData`) representations
   - All types include JSON tags for serialization

2. **`internal/core/domain/lesson_plan.go`**
   - **Removed:** `ProcessingStatus` field (eliminated ambiguity)
   - **Clarified:** `Status` field now explicitly represents pedagogical approval status only
   - **Added:** Inline comments explaining status semantics

---

## Domain Model Overview

### Three Main Entities

1. **LessonPlan** (`internal/core/domain/lesson_plan.go`)
   - Represents uploaded lesson plan document
   - `Status`: pedagogical approval status (pending/approved/needs_adjustment)
   - 1:1 relationship with AnalysisJob (processing) and LessonPlanAnalysis (result)

2. **AnalysisJob** (`internal/core/domain/analysis_job.go`)
   - Already existed; confirmed correct structure
   - Tracks asynchronous analysis task lifecycle
   - `Status`: processing status (pending/processing/done/failed)
   - Supports retry logic via `Attempts` and `ErrorMessage` fields

3. **LessonPlanAnalysis** (`internal/core/domain/lesson_plan_analysis.go`)
   - Stores final pedagogical analysis result
   - Includes AI-extracted metadata and pedagogical feedback
   - Structured data contract matches sprint LLM schema exactly
   - json tags for serialization with external systems

### Data Contract (LLM JSON Schema)
The domain model now supports the exact structure required:
```json
{
  "metadata": {
    "title": "string",
    "subject": "string",
    "grade_level": "string",
    "objectives": ["string"],
    "bncc_skills": ["string"]
  },
  "analysis": {
    "pedagogical_feedback": "string (markdown)",
    "alignment_score": "number (0-100)",
    "suggestions": ["string"],
    "missing_elements": ["string"]
  }
}
```

---

## Acceptance Criteria — ALL MET ✅

- ✅ **Domain entities exist in `internal/core/domain`**
  - `analysis_job.go` confirmed
  - `lesson_plan_analysis.go` enhanced with structured types
  - `lesson_plan.go` clarified

- ✅ **No infrastructure-specific types leak into core domain**
  - All types are pure Go: uuid, string, float64, []string, time.Time
  - No GORM, sql, or database-specific imports

- ✅ **Data contract for analysis is explicit and reusable**
  - `AnalysisResult`, `AnalysisMetadata`, `PedagogicalAnalysis` fully defined
  - JSON tags for serialization
  - Can be used in multiple contexts (API responses, storage, messaging)

- ✅ **Relationships between lesson plan, job, and result are clear**
  - Documentation in `ANALYSIS_DOMAIN_MODEL.md`
  - Inline comments in source code
  - Repository interfaces define the persistence contract

---

## Repository Interfaces

Located in `internal/core/interfaces/secondary/analysis_repository.go`:

### AnalysisJobRepository
```go
InsertJob(ctx context.Context, job *domain.AnalysisJob) (uuid.UUID, error)
GetJobByLessonPlanID(ctx context.Context, lessonPlanID uuid.UUID) (*domain.AnalysisJob, error)
FetchPendingJobs(ctx context.Context, limit int) ([]*domain.AnalysisJob, error)
UpdateJob(ctx context.Context, job *domain.AnalysisJob) error
```

### LessonPlanAnalysisRepository
```go
InsertAnalysis(ctx context.Context, a *domain.LessonPlanAnalysis) (uuid.UUID, error)
GetAnalysisByLessonPlanID(ctx context.Context, lessonPlanID uuid.UUID) (*domain.LessonPlanAnalysis, error)
```

---

## Build Status
✅ `go build ./cmd/app` — **PASSED**

No compilation errors or missing dependencies introduced.

---

## Ready for Next Task
Task 1 is complete and production-ready. The domain model provides a solid foundation for:
- Task 2: Repository interface implementation (Gorm models + converters)
- Task 3: Persistence layer (infra/postgres)
- Task 7: Analysis service orchestration
- Task 8: Async worker implementation

