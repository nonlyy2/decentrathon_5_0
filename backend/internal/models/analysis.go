package models

import "time"

type Analysis struct {
	ID                       int       `json:"id"`
	CandidateID              int       `json:"candidate_id"`
	ScoreLeadership          int       `json:"score_leadership"`
	ScoreMotivation          int       `json:"score_motivation"`
	ScoreGrowth              int       `json:"score_growth"`
	ScoreVision              int       `json:"score_vision"`
	ScoreCommunication       int       `json:"score_communication"`
	FinalScore               float64   `json:"final_score"`
	Category                 string    `json:"category"`
	AIGeneratedRisk          string    `json:"ai_generated_risk"`
	AIGeneratedScore         int       `json:"ai_generated_score"`
	IncompleteFlag           bool      `json:"incomplete_flag"`
	ExplanationLeadership    string    `json:"explanation_leadership"`
	ExplanationMotivation    string    `json:"explanation_motivation"`
	ExplanationGrowth        string    `json:"explanation_growth"`
	ExplanationVision        string    `json:"explanation_vision"`
	ExplanationCommunication string    `json:"explanation_communication"`
	Summary                  string    `json:"summary"`
	KeyStrengths             []string  `json:"key_strengths"`
	RedFlags                 []string  `json:"red_flags"`
	AnalyzedAt               time.Time `json:"analyzed_at"`
	ModelUsed                string    `json:"model_used"`
	DurationMs               int       `json:"duration_ms"`
	RecommendedMajor         *string   `json:"recommended_major"`
	MajorReasonNote          *string   `json:"major_reason_note"`
}

func (a *Analysis) IsHighRisk() bool {
	return a.AIGeneratedRisk == "high"
}
