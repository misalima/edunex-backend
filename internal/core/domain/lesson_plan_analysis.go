package domain

import (
	"time"

	"github.com/google/uuid"
)

// LessonPlanAnalysis represents the structured AI-generated analysis of a lesson plan.
// Fields extracted by the AI enable searching and filtering based on curriculum alignment,
// subject area, and grade level, independently of what the coordinator provided during upload.
type LessonPlanAnalysis struct {
	ID             uuid.UUID
	LessonPlanID   uuid.UUID
	Title          string
	Subject        string
	GradeLevel     string
	AlignmentScore int
	Feedback       string
	Metadata       string // JSON: objectives, bncc_skills (searchable)
	Suggestions    string // JSON: array of actionable suggestions
	CreatedAt      time.Time
}

// LessonPlanAnalysisStatus represents the pedagogical status and eventual result of a lesson plan analysis.
type LessonPlanAnalysisStatus struct {
	Status       string
	ErrorMessage string
	Analysis     *LessonPlanAnalysis
}
