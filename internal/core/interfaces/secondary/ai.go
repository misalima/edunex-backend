package secondary

import "context"

// AnalysisResult is the structured AI output used by the analysis pipeline.
type AnalysisResult struct {
	Metadata struct {
		Title      string   `json:"title"`
		Subject    string   `json:"subject"`
		GradeLevel string   `json:"grade_level"`
		Objectives []string `json:"objectives"`
		BNCCSkills []string `json:"bncc_skills"`
	} `json:"metadata"`
	Analysis struct {
		PedagogicalFeedback string   `json:"pedagogical_feedback"`
		AlignmentScore      int      `json:"alignment_score"`
		Suggestions         []string `json:"suggestions"`
		MissingElements     []string `json:"missing_elements"`
	} `json:"analysis"`
}

// AIProvider abstracts the AI model integration used to analyze lesson plans.
// Analyze returns a structured analysis payload ready for persistence.
type AIProvider interface {
	Analyze(ctx context.Context, text string) (*AnalysisResult, error)
}
