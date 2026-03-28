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
