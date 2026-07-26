package response

import (
	"time"

	"github.com/misalima/edunex-backend/internal/core/domain"
)

type EnqueueAnalysisResponse struct {
	JobID        string `json:"job_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	LessonPlanID string `json:"lesson_plan_id" example:"11111111-1111-1111-1111-111111111111"`
	Status       string `json:"status" example:"pending"`
}

type AnalysisJobResponse struct {
	JobID        string     `json:"job_id" example:"550e8400-e29b-41d4-a716-446655440000"`
	LessonPlanID string     `json:"lesson_plan_id" example:"11111111-1111-1111-1111-111111111111"`
	Status       string     `json:"status" example:"processing"`
	Attempts     int        `json:"attempts" example:"1"`
	ErrorMessage string     `json:"error_message,omitempty" example:"API timeout"`
	CreatedAt    time.Time  `json:"created_at" example:"2026-05-16T15:30:00Z"`
	StartedAt    *time.Time `json:"started_at,omitempty" example:"2026-05-16T15:30:01Z"`
	FinishedAt   *time.Time `json:"finished_at,omitempty" example:"2026-05-16T15:30:05Z"`
}

func FromDomainAnalysisJob(job *domain.AnalysisJob) *AnalysisJobResponse {
	if job == nil {
		return nil
	}
	return &AnalysisJobResponse{
		JobID:        job.ID.String(),
		LessonPlanID: job.LessonPlanID.String(),
		Status:       job.Status,
		Attempts:     job.Attempts,
		ErrorMessage: job.ErrorMessage,
		CreatedAt:    job.CreatedAt,
		StartedAt:    job.StartedAt,
		FinishedAt:   job.FinishedAt,
	}
}

type JobMetricsResponse struct {
	ProcessedJobs  int64            `json:"processed_jobs" example:"42"`
	SuccessfulJobs int64            `json:"successful_jobs" example:"40"`
	FailedJobs     int64            `json:"failed_jobs" example:"2"`
	RetriedJobs    int64            `json:"retried_jobs" example:"3"`
	ActiveWorkers  int              `json:"active_workers" example:"4"`
	DbStats        map[string]int64 `json:"db_stats"`
}

type LessonPlanAnalysisResponse struct {
	ID             string    `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	LessonPlanID   string    `json:"lesson_plan_id" example:"11111111-1111-1111-1111-111111111111"`
	Title          string    `json:"title" example:"Plano de aula de Matemática"`
	Subject        string    `json:"subject" example:"Matemática"`
	GradeLevel     string    `json:"grade_level" example:"1ª SÉRIE"`
	AlignmentScore int       `json:"alignment_score" example:"85"`
	Feedback       string    `json:"feedback" example:"O plano está bem estruturado..."`
	Metadata       string    `json:"metadata" example:"{\"objectives\": [\"Utilizar números inteiros\"]}"`
	Suggestions    string    `json:"suggestions" example:"[\"Adicionar atividades práticas\"]"`
	CreatedAt      time.Time `json:"created_at" example:"2026-05-16T15:30:00Z"`
}

func FromDomainAnalysis(a *domain.LessonPlanAnalysis) *LessonPlanAnalysisResponse {
	if a == nil {
		return nil
	}
	return &LessonPlanAnalysisResponse{
		ID:             a.ID.String(),
		LessonPlanID:   a.LessonPlanID.String(),
		Title:          a.Title,
		Subject:        a.Subject,
		GradeLevel:     a.GradeLevel,
		AlignmentScore: a.AlignmentScore,
		Feedback:       a.Feedback,
		Metadata:       a.Metadata,
		Suggestions:    a.Suggestions,
		CreatedAt:      a.CreatedAt,
	}
}

type LessonPlanAnalysisStatusResponse struct {
	Status       string                      `json:"status" example:"done"`
	ErrorMessage string                      `json:"error_message,omitempty" example:""`
	Analysis     *LessonPlanAnalysisResponse `json:"analysis,omitempty"`
}

func FromDomainAnalysisStatus(status *domain.LessonPlanAnalysisStatus) *LessonPlanAnalysisStatusResponse {
	if status == nil {
		return nil
	}
	resp := &LessonPlanAnalysisStatusResponse{
		Status:       status.Status,
		ErrorMessage: status.ErrorMessage,
	}
	if status.Analysis != nil {
		resp.Analysis = FromDomainAnalysis(status.Analysis)
	}
	return resp
}
