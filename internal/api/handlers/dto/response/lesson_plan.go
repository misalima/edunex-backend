package response

import (
	"time"

	"github.com/misalima/edunex-backend/internal/core/domain"
)

type LessonPlanResponse struct {
	ID          string    `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	UserID      string    `json:"user_id" example:"11111111-1111-1111-1111-111111111111"`
	Title       string    `json:"title" example:"Plano de aula de Matemática"`
	Teacher     string    `json:"teacher,omitempty" example:"Ana Paula"`
	GradeLevel  string    `json:"grade_level,omitempty" example:"1ª SÉRIE"`
	Discipline  string    `json:"discipline,omitempty" example:"Matemática"`
	Status      string    `json:"status" example:"pending"`
	CreatedAt   time.Time `json:"created_at" example:"2026-05-16T15:30:00Z"`
	DownloadURL string    `json:"download_url,omitempty" example:"https://storage.example.com/signed-url"`
}

func FromDomainLessonPlanToResponse(lp *domain.LessonPlan) *LessonPlanResponse {
	if lp == nil {
		return nil
	}
	return &LessonPlanResponse{
		ID:         lp.ID.String(),
		UserID:     lp.UserID.String(),
		Title:      lp.Title,
		Teacher:    derefString(lp.Teacher),
		GradeLevel: derefGradeLevel(lp.GradeLevel),
		Discipline: derefString(lp.Discipline),
		Status:     string(lp.Status),
		CreatedAt:  lp.CreatedAt,
	}
}

func FromDomainLessonPlanWithURL(lp *domain.LessonPlan, signedURL string) *LessonPlanResponse {
	dto := FromDomainLessonPlanToResponse(lp)
	if dto == nil {
		return nil
	}
	dto.DownloadURL = signedURL
	return dto
}

func FromDomainLessonPlanList(lps []*domain.LessonPlan) []*LessonPlanResponse {
	if lps == nil {
		return nil
	}
	out := make([]*LessonPlanResponse, 0, len(lps))
	for _, lp := range lps {
		if lp == nil {
			continue
		}
		out = append(out, FromDomainLessonPlanToResponse(lp))
	}
	return out
}

func FromDomainLessonPlanListWithURLs(lps []*domain.LessonPlan, urls map[string]string) []*LessonPlanResponse {
	if lps == nil {
		return nil
	}
	out := make([]*LessonPlanResponse, 0, len(lps))
	for _, lp := range lps {
		if lp == nil {
			continue
		}
		dto := FromDomainLessonPlanToResponse(lp)
		if dto == nil {
			continue
		}
		if urls != nil {
			if u, ok := urls[lp.ID.String()]; ok {
				dto.DownloadURL = u
			}
		}
		out = append(out, dto)
	}
	return out
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func derefGradeLevel(gl *domain.GradeLevel) string {
	if gl == nil {
		return ""
	}
	return string(*gl)
}
