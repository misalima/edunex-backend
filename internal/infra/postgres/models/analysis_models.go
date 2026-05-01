package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/misalima/edunex-backend/internal/core/domain"
	"github.com/misalima/edunex-backend/internal/core/domain_errors"
)

type AnalysisJobModel struct {
	ID           uuid.UUID               `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	LessonPlanID uuid.UUID               `gorm:"type:uuid;not null;unique"`
	Status       domain.ProcessingStatus `gorm:"type:job_status_enum;not null;default:'pending'"`
	Attempts     int                     `gorm:"type:int;not null;default:0"`
	ErrorMessage *string                 `gorm:"type:text"`
	CreatedAt    time.Time               `gorm:"column:created_at;autoCreateTime"`
	StartedAt    *time.Time              `gorm:"column:started_at"`
	FinishedAt   *time.Time              `gorm:"column:finished_at"`
}

func (AnalysisJobModel) TableName() string {
	return "analysis_jobs"
}

func (m *AnalysisJobModel) ToDomain() *domain.AnalysisJob {
	if m == nil {
		return nil
	}

	return &domain.AnalysisJob{
		ID:           m.ID,
		LessonPlanID: m.LessonPlanID,
		Status:       m.Status,
		Attempts:     m.Attempts,
		ErrorMessage: m.ErrorMessage,
		CreatedAt:    m.CreatedAt,
		StartedAt:    m.StartedAt,
		FinishedAt:   m.FinishedAt,
	}
}

func FromDomainAnalysisJob(job *domain.AnalysisJob) *AnalysisJobModel {
	if job == nil {
		return nil
	}

	m := &AnalysisJobModel{
		LessonPlanID: job.LessonPlanID,
		Status:       job.Status,
		Attempts:     job.Attempts,
		ErrorMessage: job.ErrorMessage,
		StartedAt:    job.StartedAt,
		FinishedAt:   job.FinishedAt,
	}

	if job.ID != uuid.Nil {
		m.ID = job.ID
	}
	if !job.CreatedAt.IsZero() {
		m.CreatedAt = job.CreatedAt
	}

	return m
}

type analysisMetadataPayload struct {
	Title      string   `json:"title"`
	Subject    string   `json:"subject"`
	GradeLevel string   `json:"grade_level"`
	Objectives []string `json:"objectives"`
	BnccSkills []string `json:"bncc_skills"`
}

type pedagogicalAnalysisPayload struct {
	PedagogicalFeedback string   `json:"pedagogical_feedback"`
	AlignmentScore      float64  `json:"alignment_score"`
	Suggestions         []string `json:"suggestions"`
	MissingElements     []string `json:"missing_elements"`
}

type analysisResultPayload struct {
	Metadata analysisMetadataPayload    `json:"metadata"`
	Analysis pedagogicalAnalysisPayload `json:"analysis"`
}

type LessonPlanAnalysisModel struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	LessonPlanID   uuid.UUID `gorm:"type:uuid;not null;unique"`
	AnalysisText   string    `gorm:"type:text;not null"`
	StructuredData []byte    `gorm:"type:jsonb;not null"`
	CreatedAt      time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (LessonPlanAnalysisModel) TableName() string {
	return "lesson_plan_analyses"
}

func toPayload(result domain.AnalysisResult) analysisResultPayload {
	return analysisResultPayload{
		Metadata: analysisMetadataPayload{
			Title:      result.Metadata.Title,
			Subject:    result.Metadata.Subject,
			GradeLevel: result.Metadata.GradeLevel,
			Objectives: result.Metadata.Objectives,
			BnccSkills: result.Metadata.BnccSkills,
		},
		Analysis: pedagogicalAnalysisPayload{
			PedagogicalFeedback: result.Analysis.PedagogicalFeedback,
			AlignmentScore:      result.Analysis.AlignmentScore,
			Suggestions:         result.Analysis.Suggestions,
			MissingElements:     result.Analysis.MissingElements,
		},
	}
}

func fromPayload(payload analysisResultPayload) domain.AnalysisResult {
	return domain.AnalysisResult{
		Metadata: domain.AnalysisMetadata{
			Title:      payload.Metadata.Title,
			Subject:    payload.Metadata.Subject,
			GradeLevel: payload.Metadata.GradeLevel,
			Objectives: payload.Metadata.Objectives,
			BnccSkills: payload.Metadata.BnccSkills,
		},
		Analysis: domain.PedagogicalAnalysis{
			PedagogicalFeedback: payload.Analysis.PedagogicalFeedback,
			AlignmentScore:      payload.Analysis.AlignmentScore,
			Suggestions:         payload.Analysis.Suggestions,
			MissingElements:     payload.Analysis.MissingElements,
		},
	}
}

func (m *LessonPlanAnalysisModel) ToDomain() (*domain.LessonPlanAnalysis, error) {
	if m == nil {
		return nil, nil
	}

	var payload analysisResultPayload
	if err := json.Unmarshal(m.StructuredData, &payload); err != nil {
		return nil, domain_errors.WrapUnexpectedMsg(err, "failed to parse analysis structured_data")
	}

	return &domain.LessonPlanAnalysis{
		ID:             m.ID,
		LessonPlanID:   m.LessonPlanID,
		AnalysisText:   m.AnalysisText,
		StructuredData: fromPayload(payload),
		CreatedAt:      m.CreatedAt,
	}, nil
}

func FromDomainLessonPlanAnalysis(a *domain.LessonPlanAnalysis) (*LessonPlanAnalysisModel, error) {
	if a == nil {
		return nil, domain_errors.NewBadRequestMsg("analysis result is required")
	}
	if a.LessonPlanID == uuid.Nil {
		return nil, domain_errors.NewBadRequestMsg("lesson plan id is required")
	}
	if a.AnalysisText == "" {
		return nil, domain_errors.NewBadRequestMsg("analysis text is required")
	}

	encoded, err := json.Marshal(toPayload(a.StructuredData))
	if err != nil {
		return nil, domain_errors.WrapUnexpectedMsg(err, "failed to encode analysis structured_data")
	}

	m := &LessonPlanAnalysisModel{
		LessonPlanID:   a.LessonPlanID,
		AnalysisText:   a.AnalysisText,
		StructuredData: encoded,
	}

	if a.ID != uuid.Nil {
		m.ID = a.ID
	}
	if !a.CreatedAt.IsZero() {
		m.CreatedAt = a.CreatedAt
	}

	return m, nil
}
