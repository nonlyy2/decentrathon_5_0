package gemini

import (
	"context"
	"fmt"

	"github.com/assylkhan/invisionu-backend/internal/models"
)

func (c *Client) AnalyzeCandidate(ctx context.Context, candidate *models.Candidate) (*models.Analysis, error) {
	userMsg := BuildUserMessage(candidate)
	fullSystemPrompt := SystemPrompt + "\n\nExpected JSON response schema:\n" + ResponseSchema

	responseText, err := c.Generate(ctx, fullSystemPrompt, userMsg)
	if err != nil {
		return nil, fmt.Errorf("gemini API error: %w", err)
	}

	parsed, err := ParseAnalysisResponse(responseText)
	if err != nil {
		return nil, fmt.Errorf("response parsing error: %w", err)
	}

	return ToAnalysis(parsed, candidate.ID), nil
}

func (c *Client) AnalyzeBatch(ctx context.Context, candidates []models.Candidate) ([]*models.Analysis, error) {
	userMsg := BuildBatchUserMessage(candidates)
	fullSystemPrompt := BatchSystemPrompt + "\n\nExpected JSON array schema:\n" + BatchResponseSchema

	responseText, err := c.Generate(ctx, fullSystemPrompt, userMsg)
	if err != nil {
		return nil, fmt.Errorf("gemini batch API error: %w", err)
	}

	parsed, err := ParseBatchAnalysisResponse(responseText, len(candidates))
	if err != nil {
		return nil, fmt.Errorf("batch response parsing error: %w", err)
	}

	analyses := make([]*models.Analysis, len(candidates))
	for i, r := range parsed {
		analyses[i] = ToAnalysis(r, candidates[i].ID)
	}
	return analyses, nil
}
