package groq

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"golang.org/x/time/rate"
)

const (
	BaseURL      = "https://api.groq.com/openai/v1/chat/completions"
	DefaultModel = "llama-3.3-70b-versatile"
)

type Client struct {
	apiKey     string
	httpClient *http.Client
	model      string
	limiter    *rate.Limiter
}

type chatRequest struct {
	Model       string    `json:"model"`
	Messages    []message `json:"messages"`
	Temperature float64   `json:"temperature"`
	TopP        float64   `json:"top_p"`
	MaxTokens   int       `json:"max_tokens"`
	ResponseFormat *responseFormat `json:"response_format,omitempty"`
}

type message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type responseFormat struct {
	Type string `json:"type"`
}

type chatResponse struct {
	Choices []struct {
		Message message `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
	} `json:"error,omitempty"`
}

func NewClient(apiKey, model string) *Client {
	if model == "" {
		model = DefaultModel
	}
	return &Client{
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 60 * time.Second},
		model:      model,
		limiter:    rate.NewLimiter(rate.Every(10*time.Second), 1), // ~3 req/min to stay under 6000 TPM
	}
}

func (c *Client) ModelName() string {
	return c.model
}

func (c *Client) Generate(ctx context.Context, systemPrompt, userMessage string) (string, error) {
	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		if attempt > 0 {
			backoff := 20 * time.Second
			log.Printf("Groq attempt %d failed: %v, retrying in %v...", attempt, lastErr, backoff)
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return "", ctx.Err()
			}
		}

		if err := c.limiter.Wait(ctx); err != nil {
			return "", err
		}

		result, err := c.doGenerate(ctx, systemPrompt, userMessage)
		if err == nil {
			return result, nil
		}
		lastErr = err

		if isNonRetryable(lastErr) {
			return "", lastErr
		}
	}
	return "", fmt.Errorf("groq failed after 2 attempts: %w", lastErr)
}

func (c *Client) doGenerate(ctx context.Context, systemPrompt, userMessage string) (string, error) {
	reqBody := chatRequest{
		Model: c.model,
		Messages: []message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userMessage},
		},
		Temperature: 0.3,
		TopP:        0.95,
		MaxTokens:   2048,
		ResponseFormat: &responseFormat{Type: "json_object"},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", BaseURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("groq HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("groq API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var chatResp chatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return "", fmt.Errorf("failed to parse groq response: %w", err)
	}

	if chatResp.Error != nil {
		return "", fmt.Errorf("groq error: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 || chatResp.Choices[0].Message.Content == "" {
		return "", fmt.Errorf("empty response from groq")
	}

	return chatResp.Choices[0].Message.Content, nil
}

func isNonRetryable(err error) bool {
	msg := err.Error()
	for _, s := range []string{"status 401", "status 400", "status 403", "status 404", "invalid_api_key"} {
		if strings.Contains(msg, s) {
			return true
		}
	}
	return false
}
