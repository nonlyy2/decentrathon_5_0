package ollama

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

// OllamaAnalysisResponse — JSON от локальной модели
type OllamaAnalysisResponse struct {
	ScoreLeadership          int      `json:"score_leadership"`
	ScoreMotivation          int      `json:"score_motivation"`
	ScoreGrowth              int      `json:"score_growth"`
	ScoreVision              int      `json:"score_vision"`
	ScoreCommunication       int      `json:"score_communication"`
	AIGeneratedScore         int      `json:"ai_generated_score"`
	Summary                  string   `json:"summary"`
	ExplanationLeadership    string   `json:"explanation_leadership"`
	ExplanationMotivation    string   `json:"explanation_motivation"`
	ExplanationGrowth        string   `json:"explanation_growth"`
	ExplanationVision        string   `json:"explanation_vision"`
	ExplanationCommunication string   `json:"explanation_communication"`
	KeyStrengths             []string `json:"key_strengths"`
	RedFlags                 []string `json:"red_flags"`
	RecommendedMajor         string   `json:"recommended_major"`
	MajorReasonNote          string   `json:"major_reason_note"`
}

var jsonCodeFenceRe = regexp.MustCompile("(?s)```(?:json)?\\s*(\\{.*?\\})\\s*```")

// extractJSON — первый JSON-объект из ответа (снимает markdown-обёртку)
func extractJSON(raw string) (string, error) {
	raw = strings.TrimSpace(raw)

	if m := jsonCodeFenceRe.FindStringSubmatch(raw); len(m) == 2 {
		return m[1], nil
	}

	start := strings.Index(raw, "{")
	end := strings.LastIndex(raw, "}")
	if start >= 0 && end > start {
		return raw[start : end+1], nil
	}

	return "", fmt.Errorf("no JSON object found in response")
}

// ParseOllamaResponse — парсит и валидирует ответ модели
func ParseOllamaResponse(raw string) (*OllamaAnalysisResponse, error) {
	jsonStr, err := extractJSON(raw)
	if err != nil {
		return nil, fmt.Errorf("JSON extraction failed: %w (raw: %.200s)", err, raw)
	}

	var resp OllamaAnalysisResponse
	if err := json.Unmarshal([]byte(jsonStr), &resp); err != nil {
		return nil, fmt.Errorf("JSON parse failed: %w (extracted: %.300s)", err, jsonStr)
	}

	resp.ScoreLeadership = clampInt(resp.ScoreLeadership)
	resp.ScoreMotivation = clampInt(resp.ScoreMotivation)
	resp.ScoreGrowth = clampInt(resp.ScoreGrowth)
	resp.ScoreVision = clampInt(resp.ScoreVision)
	resp.ScoreCommunication = clampInt(resp.ScoreCommunication)
	resp.AIGeneratedScore = clampInt(resp.AIGeneratedScore)

	// Все нули — модель не обработала промпт
	total := resp.ScoreLeadership + resp.ScoreMotivation + resp.ScoreGrowth +
		resp.ScoreVision + resp.ScoreCommunication
	if total == 0 {
		return nil, fmt.Errorf("model returned all-zero scores — likely failed to process the prompt (check model size and context)")
	}

	if resp.KeyStrengths == nil {
		resp.KeyStrengths = []string{}
	}
	if resp.RedFlags == nil {
		resp.RedFlags = []string{}
	}

	return &resp, nil
}

func clampInt(v int) int {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return v
}
