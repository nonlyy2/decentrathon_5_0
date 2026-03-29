package ollama

import (
	"context"
	"fmt"
	"math"

	"github.com/assylkhan/invisionu-backend/internal/gemini"
	"github.com/assylkhan/invisionu-backend/internal/models"
)

// AnalyzeCandidate uses the dedicated Ollama prompt (much shorter than the Gemini one)
// and a robust parser that handles common local-model quirks.
func (c *Client) AnalyzeCandidate(ctx context.Context, candidate *models.Candidate) (*models.Analysis, error) {
	userMsg := BuildPrompt(candidate)

	responseText, err := c.Generate(ctx, SystemPrompt, userMsg)
	if err != nil {
		return nil, fmt.Errorf("ollama error: %w", err)
	}

	parsed, err := ParseOllamaResponse(responseText)
	if err != nil {
		return nil, fmt.Errorf("ollama response parse failed: %w", err)
	}

	return toAnalysis(parsed, candidate.ID, "ollama/"+c.ModelName()), nil
}

// AnalyzeBatch processes each candidate individually for Ollama.
// Local models can't handle batch prompts reliably — too much context.
func (c *Client) AnalyzeBatch(ctx context.Context, candidates []models.Candidate) ([]*models.Analysis, error) {
	analyses := make([]*models.Analysis, len(candidates))
	for i := range candidates {
		select {
		case <-ctx.Done():
			return analyses, ctx.Err()
		default:
		}
		a, err := c.AnalyzeCandidate(ctx, &candidates[i])
		if err != nil {
			return nil, fmt.Errorf("candidate %d: %w", candidates[i].ID, err)
		}
		analyses[i] = a
	}
	return analyses, nil
}

func toAnalysis(r *OllamaAnalysisResponse, candidateID int, modelUsed string) *models.Analysis {
	finalScore := math.Round((float64(r.ScoreLeadership)*0.25+
		float64(r.ScoreMotivation)*0.25+
		float64(r.ScoreGrowth)*0.20+
		float64(r.ScoreVision)*0.15+
		float64(r.ScoreCommunication)*0.15)*100) / 100

	return &models.Analysis{
		CandidateID:              candidateID,
		ScoreLeadership:          r.ScoreLeadership,
		ScoreMotivation:          r.ScoreMotivation,
		ScoreGrowth:              r.ScoreGrowth,
		ScoreVision:              r.ScoreVision,
		ScoreCommunication:       r.ScoreCommunication,
		FinalScore:               finalScore,
		Category:                 gemini.ScoreToCategory(finalScore),
		AIGeneratedRisk:          gemini.AIScoreToRisk(r.AIGeneratedScore),
		AIGeneratedScore:         r.AIGeneratedScore,
		IncompleteFlag:           false,
		ExplanationLeadership:    r.ExplanationLeadership,
		ExplanationMotivation:    r.ExplanationMotivation,
		ExplanationGrowth:        r.ExplanationGrowth,
		ExplanationVision:        r.ExplanationVision,
		ExplanationCommunication: r.ExplanationCommunication,
		Summary:                  r.Summary,
		KeyStrengths:             r.KeyStrengths,
		RedFlags:                 r.RedFlags,
		ModelUsed:                modelUsed,
	}
}
