# Domain Model Architecture: Lesson Plan Analysis Pipeline

## Overview
The lesson plan analysis pipeline consists of three main domain entities that work together to manage the asynchronous pedagogical analysis workflow.

## Domain Entities

### 1. LessonPlan
**Located:** `internal/core/domain/lesson_plan.go`

Represents a lesson plan document uploaded by a user.

**Key Fields:**
- `ID`, `UserID`, `Title`, `FilePath`: Core metadata
- `Status`: `LessonPlanStatus` enum (`pending`, `approved`, `needs_adjustment`)
  - This represents the **pedagogical approval status**, not the analysis processing status
  - Managed separately by domain/business logic

**Relationship:** One lesson plan can have **one analysis job** and **one analysis result**.

---

### 2. AnalysisJob
**Located:** `internal/core/domain/analysis_job.go`

Represents an asynchronous job in the analysis pipeline queue.

**Key Fields:**
- `ID`, `LessonPlanID`: Job identity and reference to the lesson plan
- `Status`: `ProcessingStatus` enum (`pending`, `processing`, `done`, `failed`)
  - Tracks where the job is in the processing pipeline
  - Incremented and checked by the background worker
- `Attempts`: Number of processing attempts (for retry logic)
- `ErrorMessage`: Captured error details if the job fails
- `CreatedAt`, `StartedAt`, `FinishedAt`: Lifecycle timestamps

**Lifecycle:**
```
pending → processing → done ✓
       ↓
    failed (retry possible)
```

**Relationship:** One analysis job is **uniquely** tied to one lesson plan (1:1).

---

### 3. LessonPlanAnalysis
**Located:** `internal/core/domain/lesson_plan_analysis.go`

Represents the final, persisted analysis result for a lesson plan.

**Key Types:**

#### AnalysisMetadata
Extracted metadata from the lesson plan:
- `Title`, `Subject`, `GradeLevel`: Document attributes
- `Objectives`: Learning objectives extracted by the LLM
- `BnccSkills`: BNCC competencies identified in the content

#### PedagogicalAnalysis
AI-generated pedagogical feedback:
- `PedagogicalFeedback`: Markdown-formatted suggestions and commentary
- `AlignmentScore`: 0-100 score of BNCC alignment
- `Suggestions`: Actionable improvement recommendations
- `MissingElements`: Pedagogical components that are missing

#### AnalysisResult
The complete, structured LLM response:
```json
{
  "metadata": { ... AnalysisMetadata ... },
  "analysis": { ... PedagogicalAnalysis ... }
}
```

#### LessonPlanAnalysis
The persisted entity:
- `ID`, `LessonPlanID`: Analysis identity and reference
- `AnalysisText`: Raw JSON string returned by the LLM (audit trail)
- `StructuredData`: Parsed `AnalysisResult` object (for easy access)
- `CreatedAt`: When the analysis was completed

**Relationship:** One analysis result is **uniquely** tied to one lesson plan (1:1).

---

## Entity Relationships

```
┌─────────────────┐
│  LessonPlan     │
├─────────────────┤
│ ID (UUID)       │
│ Title           │
│ Status (pedagogical approval)
│ FilePath        │
└─────────────────┘
        │
        │ 1:1
        ├─────────────────────────────────────────┐
        │                                         │
        ▼                                         ▼
┌─────────────────┐                    ┌──────────────────────┐
│ AnalysisJob     │                    │ LessonPlanAnalysis   │
├─────────────────┤                    ├──────────────────────┤
│ ID (UUID)       │                    │ ID (UUID)            │
│ LessonPlanID    │                    │ LessonPlanID         │
│ Status (job)    │                    │ AnalysisText (raw)   │
│ Attempts        │                    │ StructuredData       │
│ ErrorMessage    │                    │ CreatedAt            │
│ Timestamps      │                    └──────────────────────┘
└─────────────────┘
```

---

## Key Design Decisions

1. **Separation of Concerns:**
   - `LessonPlan.Status`: Pedagogical approval status (human decision)
   - `AnalysisJob.Status`: Analysis processing status (system lifecycle)
   - `LessonPlanAnalysis`: Immutable result storage

2. **Domain Independence (Pure Core):**
   - All entities are POJO-like types with **no infrastructure dependencies**
   - No `json` tags (serialization is a concern of API/infra layers)
   - No GORM tags, SQL concerns, or external library coupling
   - Database mapping and serialization happen at the `internal/infra` and `internal/api` layers via converters/DTOs

3. **Structured Data Contract:**
   - `AnalysisResult` follows the LLM schema defined in the sprint specification
   - Both raw (`AnalysisText`) and parsed (`StructuredData`) versions are stored for audit and reliability
   - Serialization to JSON happens only in:
     - DTOs for API responses (`internal/api/handlers/dto/response`)
     - Database converters when storing JSON in Postgres (`internal/infra/postgres`)

4. **Error Tracking:**
   - `AnalysisJob.ErrorMessage` captures failure context
   - `Attempts` enables retry logic without circular dependencies

---

## Serialization Strategy

**In the Core (Domain):** NO `json` tags or any serialization concerns
```go
// ✅ CORRECT - Pure domain, no external technology
type AnalysisMetadata struct {
    Title      string
    Subject    string
    GradeLevel string
}
```

**In the API Layer (DTOs):** Add `json` tags for REST responses
```go
// ✅ CORRECT - DTO for API, can have json tags
type AnalysisMetadataResponse struct {
    Title      string   `json:"title"`
    Subject    string   `json:"subject"`
    GradeLevel string   `json:"grade_level"`
}
```

**In the Infra Layer (Gorm Models):** Add `json` tags if storing structured data as JSON in the database
```go
// ✅ CORRECT - DB model, can have db and json tags
type LessonPlanAnalysisModel struct {
    ID             uuid.UUID `gorm:"primaryKey"`
    StructuredData *string   `gorm:"type:jsonb" json:"structured_data"`
}
```

This keeps the core clean from technology-specific concerns while allowing flexible serialization at the boundaries.

---

**Located:** `internal/core/interfaces/secondary/analysis_repository.go`

Two interfaces define the persistence contracts:

### AnalysisJobRepository
- `InsertJob(ctx, job) → UUID`
- `GetJobByLessonPlanID(ctx, lessonPlanID) → *AnalysisJob`
- `FetchPendingJobs(ctx, limit) → []*AnalysisJob` (used by worker)
- `UpdateJob(ctx, job) → error`

### LessonPlanAnalysisRepository
- `InsertAnalysis(ctx, analysis) → UUID`
- `GetAnalysisByLessonPlanID(ctx, lessonPlanID) → *LessonPlanAnalysis`

---

## Next Steps

1. **Task 3 (Persistence):** Implement GORM models in `internal/infra/postgres` with converters.
2. **Task 7 (Service):** Create `internal/core/services/analysis_service.go` using these interfaces and the extractor/AI client.
3. **Task 8 (Worker):** Implement background worker consuming `FetchPendingJobs()` and calling the analysis service.



