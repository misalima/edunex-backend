package response

import (
	"time"

	"github.com/misalima/edunex-backend/internal/core/domain"
)

type LessonPlanResponse struct {
	ID          string    `json:"id"`
	UserID      string    `json:"user_id"`
	Title       string    `json:"title"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	DownloadURL string    `json:"download_url,omitempty"`
}

func FromDomainLessonPlanToResponse(lp *domain.LessonPlan) *LessonPlanResponse {
	if lp == nil {
		return nil
	}
	return &LessonPlanResponse{
		ID:        lp.ID.String(),
		UserID:    lp.UserID.String(),
		Title:     lp.Title,
		Status:    lp.Status,
		CreatedAt: lp.Created,
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
