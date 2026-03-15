package domain

type LessonPlanStatus string

const (
	LessonPlanPending         LessonPlanStatus = "pending"
	LessonPlanApproved        LessonPlanStatus = "approved"
	LessonPlanNeedsAdjustment LessonPlanStatus = "needs_adjustment"
)

func (s LessonPlanStatus) IsValid() bool {
	return s == LessonPlanPending || s == LessonPlanApproved || s == LessonPlanNeedsAdjustment
}
