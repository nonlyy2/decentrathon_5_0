package gemini

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync/atomic"

	"github.com/assylkhan/invisionu-backend/internal/models"
)

// PoolClient rotates over multiple API keys. On 429 it skips to the next key.
type PoolClient struct {
	clients []*Client
	idx     atomic.Uint64
}

func NewPool(keys []string) *PoolClient {
	clients := make([]*Client, len(keys))
	for i, k := range keys {
		clients[i] = NewClient(k)
	}
	return &PoolClient{clients: clients}
}

func (p *PoolClient) Generate(ctx context.Context, systemPrompt, userMessage string) (string, error) {
	n := len(p.clients)
	start := int(p.idx.Add(1)-1) % n

	for i := 0; i < n; i++ {
		idx := (start + i) % n
		c := p.clients[idx]

		// Respect per-key rate limiter
		if err := c.limiter.Wait(ctx); err != nil {
			return "", err
		}

		// Call doGenerate directly — skips per-key retry loop so we rotate immediately on 429
		result, err := c.doGenerate(ctx, systemPrompt, userMessage)
		if err == nil {
			return result, nil
		}

		var ge *GeminiError
		if errors.As(err, &ge) && ge.StatusCode == 429 {
			log.Printf("Gemini key #%d rate-limited (429), rotating to key #%d...", idx, (idx+1)%n)
			continue
		}

		return "", err // non-rate-limit error — don't rotate
	}

	return "", fmt.Errorf("all %d Gemini API keys are rate-limited (429)", n)
}

func (p *PoolClient) AnalyzeBatch(ctx context.Context, candidates []models.Candidate) ([]*models.Analysis, error) {
	userMsg := BuildBatchUserMessage(candidates)
	fullSystemPrompt := BatchSystemPrompt + "\n\nExpected JSON array schema:\n" + BatchResponseSchema

	responseText, err := p.Generate(ctx, fullSystemPrompt, userMsg)
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

func (p *PoolClient) AnalyzeCandidate(ctx context.Context, candidate *models.Candidate) (*models.Analysis, error) {
	userMsg := BuildUserMessage(candidate)
	fullSystemPrompt := SystemPrompt + "\n\nExpected JSON response schema:\n" + ResponseSchema

	responseText, err := p.Generate(ctx, fullSystemPrompt, userMsg)
	if err != nil {
		return nil, fmt.Errorf("gemini API error: %w", err)
	}

	parsed, err := ParseAnalysisResponse(responseText)
	if err != nil {
		return nil, fmt.Errorf("response parsing error: %w", err)
	}

	return ToAnalysis(parsed, candidate.ID), nil
}
