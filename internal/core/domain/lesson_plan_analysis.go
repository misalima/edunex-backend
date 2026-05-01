package domain

import (
	"time"

	"github.com/google/uuid"
)

// AnalysisMetadata maps the "metadata" block returned by the LLM.
// It contains structured metadata about the lesson plan extracted and validated by the AI.
type AnalysisMetadata struct {
	Title      string
	Subject    string
	GradeLevel string
	Objectives []string
	BnccSkills []string
}

// PedagogicalAnalysis maps the "analysis" block returned by the LLM.
// It contains the AI's pedagogical feedback and suggestions aligned with BNCC standards.
type PedagogicalAnalysis struct {
	PedagogicalFeedback string  // markdown formatted
	AlignmentScore      float64 // 0-100
	Suggestions         []string
	MissingElements     []string
}

// AnalysisResult is the parsed, structured result returned by the LLM.
// It follows the data contract defined in the sprint specification.
type AnalysisResult struct {
	Metadata AnalysisMetadata
	Analysis PedagogicalAnalysis
}

// LessonPlanAnalysis represents the complete analysis result persisted for a lesson plan.
// It includes both the raw LLM response (AnalysisText) and the parsed structured data (StructuredData).
type LessonPlanAnalysis struct {
	ID             uuid.UUID
	LessonPlanID   uuid.UUID
	AnalysisText   string         // raw JSON string returned by the LLM (for audit/debug)
	StructuredData AnalysisResult // parsed structured representation
	CreatedAt      time.Time
}
