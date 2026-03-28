package groq

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
		return nil, fmt.Errorf("groq API error: %w", err)
	}

	parsed, err := gemini.ParseAnalysisResponse(responseText)
	if err != nil {
		return nil, fmt.Errorf("response parsing error: %w", err)
	}

	analysis := gemini.ToAnalysis(parsed, candidate.ID)
	analysis.ModelUsed = "groq/" + c.ModelName()
	return analysis, nil
}
