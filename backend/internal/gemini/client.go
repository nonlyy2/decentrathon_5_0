package gemini

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

const (
	BaseURL   = "https://generativelanguage.googleapis.com/v1beta/models"
	ModelName = "gemini-2.5-flash"
)

type Client struct {
	apiKey     string
	httpClient *http.Client
	model      string
}

type GeminiRequest struct {
	Contents          []Content        `json:"contents"`
	SystemInstruction *Content         `json:"systemInstruction,omitempty"`
	GenerationConfig  GenerationConfig `json:"generationConfig"`
}

type Content struct {
	Parts []Part `json:"parts"`
	Role  string `json:"role,omitempty"`
}

type Part struct {
	Text string `json:"text"`
}

type GenerationConfig struct {
	Temperature      float64 `json:"temperature"`
	TopP             float64 `json:"topP"`
	MaxOutputTokens  int     `json:"maxOutputTokens"`
	ResponseMimeType string  `json:"responseMimeType"`
}

type GeminiResponse struct {
	Candidates []struct {
		Content Content `json:"content"`
	} `json:"candidates"`
}

type GeminiError struct {
	StatusCode int
	Message    string
}

func (e *GeminiError) Error() string {
	return fmt.Sprintf("gemini API error (status %d): %s", e.StatusCode, e.Message)
}

func NewClient(apiKey string) *Client {
	return &Client{
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 120 * time.Second},
		model:      ModelName,
	}
}

func (c *Client) Generate(ctx context.Context, systemPrompt, userMessage string) (string, error) {
	return c.generate(ctx, systemPrompt, userMessage, 2048)
}

func (c *Client) GenerateLarge(ctx context.Context, systemPrompt, userMessage string) (string, error) {
	return c.generate(ctx, systemPrompt, userMessage, 24576)
}

// свободный текст (не JSON)
func (c *Client) GenerateText(ctx context.Context, systemPrompt, userMessage string) (string, error) {
	return c.generatePlain(ctx, systemPrompt, userMessage, 8192)
}

func (c *Client) generatePlain(ctx context.Context, systemPrompt, userMessage string, maxTokens int) (string, error) {
	url := fmt.Sprintf("%s/%s:generateContent?key=%s", BaseURL, c.model, c.apiKey)

	reqBody := GeminiRequest{
		SystemInstruction: &Content{Parts: []Part{{Text: systemPrompt}}},
		Contents:          []Content{{Role: "user", Parts: []Part{{Text: userMessage}}}},
		GenerationConfig: GenerationConfig{
			Temperature:      0.7,
			TopP:             0.95,
			MaxOutputTokens:  maxTokens,
			ResponseMimeType: "text/plain",
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return "", &GeminiError{StatusCode: resp.StatusCode, Message: string(respBody)}
	}

	var geminiResp GeminiResponse
	if err := json.Unmarshal(respBody, &geminiResp); err != nil {
		return "", fmt.Errorf("failed to parse response JSON: %w", err)
	}
	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("empty response from Gemini")
	}
	return geminiResp.Candidates[0].Content.Parts[0].Text, nil
}

func (c *Client) generate(ctx context.Context, systemPrompt, userMessage string, maxTokens int) (string, error) {
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(attempt*attempt) * 2 * time.Second
			log.Printf("Gemini attempt %d failed: %v, retrying in %v...", attempt, lastErr, backoff)
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return "", ctx.Err()
			}
		}

		result, err := c.doGenerate(ctx, systemPrompt, userMessage, maxTokens)
		if err == nil {
			return result, nil
		}
		lastErr = err

		if !isRetryable(err) {
			return "", err
		}
	}
	return "", fmt.Errorf("gemini failed after 3 attempts: %w", lastErr)
}

func (c *Client) doGenerate(ctx context.Context, systemPrompt, userMessage string, maxTokens int) (string, error) {
	url := fmt.Sprintf("%s/%s:generateContent?key=%s", BaseURL, c.model, c.apiKey)

	reqBody := GeminiRequest{
		SystemInstruction: &Content{Parts: []Part{{Text: systemPrompt}}},
		Contents:          []Content{{Role: "user", Parts: []Part{{Text: userMessage}}}},
		GenerationConfig: GenerationConfig{
			Temperature:      0.3,
			TopP:             0.95,
			MaxOutputTokens:  maxTokens,
			ResponseMimeType: "application/json",
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return "", &GeminiError{
			StatusCode: resp.StatusCode,
			Message:    string(respBody),
		}
	}

	var geminiResp GeminiResponse
	if err := json.Unmarshal(respBody, &geminiResp); err != nil {
		return "", fmt.Errorf("failed to parse response JSON: %w", err)
	}

	if len(geminiResp.Candidates) == 0 || len(geminiResp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("empty response from Gemini")
	}

	return geminiResp.Candidates[0].Content.Parts[0].Text, nil
}

func isRetryable(err error) bool {
	var ge *GeminiError
	if errors.As(err, &ge) {
		return ge.StatusCode == 429 || ge.StatusCode >= 500
	}
	return false
}
