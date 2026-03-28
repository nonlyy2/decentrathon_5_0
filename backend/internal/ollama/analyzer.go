package ollama

import (
	"context"
	"fmt"

	"github.com/assylkhan/invisionu-backend/internal/gemini"
	"github.com/assylkhan/invisionu-backend/internal/models"
)

func (c *Client) AnalyzeCandidate(ctx context.Context, candidate *models.Candidate) (*models.Analysis, error) {
	userMsg := gemini.BuildUserMessage(candidate)
	fullSystemPrompt := gemini.SystemPrompt + "\n\nExpected JSON response schema:\n" + gemini.ResponseSchema

	responseText, err := c.Generate(ctx, fullSystemPrompt, userMsg)
	if err != nil {
		return nil, fmt.Errorf("ollama API error: %w", err)
	}

	parsed, err := gemini.ParseAnalysisResponse(responseText)
	if err != nil {
		return nil, fmt.Errorf("response parsing error: %w", err)
	}

	analysis := gemini.ToAnalysis(parsed, candidate.ID)
	analysis.ModelUsed = "ollama/" + c.ModelName()
	return analysis, nil
}

func (c *Client) AnalyzeBatch(ctx context.Context, candidates []models.Candidate) ([]*models.Analysis, error) {
	userMsg := gemini.BuildBatchUserMessage(candidates)
	fullSystemPrompt := gemini.BatchSystemPrompt + "\n\nExpected JSON array schema:\n" + gemini.BatchResponseSchema
	responseText, err := c.Generate(ctx, fullSystemPrompt, userMsg)
	if err != nil {
		return nil, fmt.Errorf("ollama batch API error: %w", err)
	}
	parsed, err := gemini.ParseBatchAnalysisResponse(responseText, len(candidates))
	if err != nil {
		return nil, fmt.Errorf("batch response parsing error: %w", err)
	}
	analyses := make([]*models.Analysis, len(candidates))
	for i, r := range parsed {
		analyses[i] = gemini.ToAnalysis(r, candidates[i].ID)
		analyses[i].ModelUsed = "ollama/" + c.ModelName()
	}
	return analyses, nil
}
