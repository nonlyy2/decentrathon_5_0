package ollama

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
)

const (
	DefaultURL   = "http://localhost:11434"
	DefaultModel = "mistral:7b"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
	model      string
}

type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Format   string    `json:"format"`
	Stream   bool      `json:"stream"`
	Options  Options   `json:"options"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type Options struct {
	Temperature float64 `json:"temperature"`
	TopP        float64 `json:"top_p"`
	NumPredict  int     `json:"num_predict"`
}

type ChatResponse struct {
	Message Message `json:"message"`
}

func NewClient(baseURL, model string) *Client {
	if baseURL == "" {
		baseURL = DefaultURL
	}
	if model == "" {
		model = DefaultModel
	}
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 300 * time.Second},
		model:      model,
	}
}

func (c *Client) ModelName() string {
	return c.model
}

// генерирует свободный текст (не JSON)
func (c *Client) GenerateText(ctx context.Context, systemPrompt, userMessage string) (string, error) {
	return c.generatePlain(ctx, systemPrompt, userMessage)
}

func (c *Client) generatePlain(ctx context.Context, systemPrompt, userMessage string) (string, error) {
	url := fmt.Sprintf("%s/api/chat", c.baseURL)

	reqBody := ChatRequest{
		Model: c.model,
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userMessage},
		},
		Format: "", // свободный текст
		Stream: false,
		Options: Options{
			Temperature: 0.7,
			TopP:        0.95,
			NumPredict:  4096,
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("ollama error (status %d): %s", resp.StatusCode, string(body))
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}
	return chatResp.Message.Content, nil
}

func (c *Client) Generate(ctx context.Context, systemPrompt, userMessage string) (string, error) {
	var lastErr error
	for attempt := 0; attempt < 2; attempt++ {
		if attempt > 0 {
			backoff := 3 * time.Second
			log.Printf("Ollama attempt %d failed: %v, retrying in %v...", attempt, lastErr, backoff)
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return "", ctx.Err()
			}
		}

		result, err := c.doGenerate(ctx, systemPrompt, userMessage)
		if err == nil {
			return result, nil
		}
		lastErr = err

		// не ретраим нефатальные ошибки (модель не найдена, bad request)
		if isNonRetryable(lastErr) {
			return "", lastErr
		}
	}
	return "", fmt.Errorf("ollama failed after 2 attempts: %w", lastErr)
}

// нефатальные ошибки — не ретраим
func isNonRetryable(err error) bool {
	msg := err.Error()
	for _, s := range []string{"status 404", "status 400", "not found", "marshal"} {
		if strings.Contains(msg, s) {
			return true
		}
	}
	return false
}

func (c *Client) doGenerate(ctx context.Context, systemPrompt, userMessage string) (string, error) {
	url := fmt.Sprintf("%s/api/chat", c.baseURL)

	reqBody := ChatRequest{
		Model: c.model,
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userMessage},
		},
		Format: "json",
		Stream: false,
		Options: Options{
			Temperature: 0.1,
			TopP:        0.95,
			NumPredict:  4096,
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
		return "", fmt.Errorf("ollama HTTP request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("ollama API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(respBody, &chatResp); err != nil {
		return "", fmt.Errorf("failed to parse ollama response: %w", err)
	}

	if chatResp.Message.Content == "" {
		return "", fmt.Errorf("empty response from ollama")
	}

	return chatResp.Message.Content, nil
}
