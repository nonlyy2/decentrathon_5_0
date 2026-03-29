package gemini

import (
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/assylkhan/invisionu-backend/internal/models"
)

type GeminiAnalysisResponse struct {
	ScoreLeadership          int      `json:"score_leadership"`
	ScoreMotivation          int      `json:"score_motivation"`
	ScoreGrowth              int      `json:"score_growth"`
	ScoreVision              int      `json:"score_vision"`
	ScoreCommunication       int      `json:"score_communication"`
	FinalScore               float64  `json:"final_score"`
	Category                 string   `json:"category"`
	AIGeneratedRisk          string   `json:"ai_generated_risk"`
	AIGeneratedScore         int      `json:"ai_generated_score"`
	IncompleteFlag           bool     `json:"incomplete_flag"`
	ExplanationLeadership    string   `json:"explanation_leadership"`
	ExplanationMotivation    string   `json:"explanation_motivation"`
	ExplanationGrowth        string   `json:"explanation_growth"`
	ExplanationVision        string   `json:"explanation_vision"`
	ExplanationCommunication string   `json:"explanation_communication"`
	Summary                  string   `json:"summary"`
	KeyStrengths             []string `json:"key_strengths"`
	RedFlags                 []string `json:"red_flags"`
}

func ParseBatchAnalysisResponse(jsonStr string, expected int) ([]*GeminiAnalysisResponse, error) {
	// Strip any markdown code fences the model might add
	jsonStr = strings.TrimSpace(jsonStr)
	if idx := strings.Index(jsonStr, "["); idx > 0 {
		jsonStr = jsonStr[idx:]
	}
	if idx := strings.LastIndex(jsonStr, "]"); idx >= 0 && idx < len(jsonStr)-1 {
		jsonStr = jsonStr[:idx+1]
	}

	var responses []*GeminiAnalysisResponse
	if err := json.Unmarshal([]byte(jsonStr), &responses); err != nil {
		return nil, fmt.Errorf("failed to parse batch response: %w", err)
	}
	if len(responses) != expected {
		return nil, fmt.Errorf("expected %d results but got %d", expected, len(responses))
	}

	for _, resp := range responses {
		resp.ScoreLeadership = clamp(resp.ScoreLeadership, 0, 100)
		resp.ScoreMotivation = clamp(resp.ScoreMotivation, 0, 100)
		resp.ScoreGrowth = clamp(resp.ScoreGrowth, 0, 100)
		resp.ScoreVision = clamp(resp.ScoreVision, 0, 100)
		resp.ScoreCommunication = clamp(resp.ScoreCommunication, 0, 100)
		resp.FinalScore = math.Round((float64(resp.ScoreLeadership)*0.25+
			float64(resp.ScoreMotivation)*0.25+
			float64(resp.ScoreGrowth)*0.20+
			float64(resp.ScoreVision)*0.15+
			float64(resp.ScoreCommunication)*0.15)*100) / 100
		resp.Category = scoreToCategory(resp.FinalScore)
		resp.AIGeneratedScore = clamp(resp.AIGeneratedScore, 0, 100)
		resp.AIGeneratedRisk = aiScoreToRisk(resp.AIGeneratedScore)
		if resp.KeyStrengths == nil {
			resp.KeyStrengths = []string{}
		}
		if resp.RedFlags == nil {
			resp.RedFlags = []string{}
		}
	}
	return responses, nil
}

func ParseAnalysisResponse(jsonStr string) (*GeminiAnalysisResponse, error) {
	var resp GeminiAnalysisResponse
	if err := json.Unmarshal([]byte(jsonStr), &resp); err != nil {
		return nil, fmt.Errorf("failed to parse Gemini response: %w", err)
	}

	// Clamp scores
	resp.ScoreLeadership = clamp(resp.ScoreLeadership, 0, 100)
	resp.ScoreMotivation = clamp(resp.ScoreMotivation, 0, 100)
	resp.ScoreGrowth = clamp(resp.ScoreGrowth, 0, 100)
	resp.ScoreVision = clamp(resp.ScoreVision, 0, 100)
	resp.ScoreCommunication = clamp(resp.ScoreCommunication, 0, 100)

	// Recalculate final score (don't trust LLM math)
	resp.FinalScore = math.Round((float64(resp.ScoreLeadership)*0.25+
		float64(resp.ScoreMotivation)*0.25+
		float64(resp.ScoreGrowth)*0.20+
		float64(resp.ScoreVision)*0.15+
		float64(resp.ScoreCommunication)*0.15)*100) / 100

	// Recalculate category
	resp.Category = scoreToCategory(resp.FinalScore)

	// Clamp and derive AI detection
	resp.AIGeneratedScore = clamp(resp.AIGeneratedScore, 0, 100)
	resp.AIGeneratedRisk = aiScoreToRisk(resp.AIGeneratedScore)

	// Ensure slices are not nil
	if resp.KeyStrengths == nil {
		resp.KeyStrengths = []string{}
	}
	if resp.RedFlags == nil {
		resp.RedFlags = []string{}
	}

	return &resp, nil
}

func ToAnalysis(resp *GeminiAnalysisResponse, candidateID int) *models.Analysis {
	return &models.Analysis{
		CandidateID:              candidateID,
		ScoreLeadership:          resp.ScoreLeadership,
		ScoreMotivation:          resp.ScoreMotivation,
		ScoreGrowth:              resp.ScoreGrowth,
		ScoreVision:              resp.ScoreVision,
		ScoreCommunication:       resp.ScoreCommunication,
		FinalScore:               resp.FinalScore,
		Category:                 resp.Category,
		AIGeneratedRisk:          resp.AIGeneratedRisk,
		AIGeneratedScore:         resp.AIGeneratedScore,
		IncompleteFlag:           resp.IncompleteFlag,
		ExplanationLeadership:    resp.ExplanationLeadership,
		ExplanationMotivation:    resp.ExplanationMotivation,
		ExplanationGrowth:        resp.ExplanationGrowth,
		ExplanationVision:        resp.ExplanationVision,
		ExplanationCommunication: resp.ExplanationCommunication,
		Summary:                  resp.Summary,
		KeyStrengths:             resp.KeyStrengths,
		RedFlags:                 resp.RedFlags,
		AnalyzedAt:               time.Now(),
		ModelUsed:                ModelName,
	}
}

func clamp(val, min, max int) int {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

func aiScoreToRisk(score int) string {
	switch {
	case score >= 66:
		return "high"
	case score >= 36:
		return "medium"
	default:
		return "low"
	}
}

func scoreToCategory(score float64) string {
	switch {
	case score >= 80:
		return "Strong Recommend"
	case score >= 65:
		return "Recommend"
	case score >= 50:
		return "Borderline"
	default:
		return "Not Recommended"
	}
}
