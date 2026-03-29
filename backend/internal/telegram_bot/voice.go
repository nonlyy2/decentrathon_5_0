package telegram_bot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// WhisperClient handles audio transcription via OpenAI Whisper API.
type WhisperClient struct {
	apiKey     string
	httpClient *http.Client
}

// NewWhisperClient creates a Whisper transcription client.
func NewWhisperClient(apiKey string) *WhisperClient {
	return &WhisperClient{
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 120 * time.Second},
	}
}

type whisperResponse struct {
	Text string `json:"text"`
}

// Transcribe sends audio data to OpenAI Whisper API and returns the transcribed text.
func (w *WhisperClient) Transcribe(ctx context.Context, audioData []byte, language string) (string, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	part, err := writer.CreateFormFile("file", "voice.ogg")
	if err != nil {
		return "", fmt.Errorf("create form file: %w", err)
	}
	if _, err := part.Write(audioData); err != nil {
		return "", fmt.Errorf("write audio data: %w", err)
	}

	_ = writer.WriteField("model", "whisper-1")
	if language != "" {
		// Map our language codes to Whisper language codes
		langMap := map[string]string{"en": "en", "ru": "ru", "kz": "kk"}
		if code, ok := langMap[language]; ok {
			_ = writer.WriteField("language", code)
		}
	}
	writer.Close()

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/audio/transcriptions", &body)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+w.apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("whisper request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("whisper API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result whisperResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("parse whisper response: %w", err)
	}

	return result.Text, nil
}

// downloadVoiceFile downloads a voice message from Telegram servers.
func downloadVoiceFile(bot *tgbotapi.BotAPI, fileID string) ([]byte, error) {
	file, err := bot.GetFile(tgbotapi.FileConfig{FileID: fileID})
	if err != nil {
		return nil, fmt.Errorf("get file info: %w", err)
	}

	fileURL := file.Link(bot.Token)
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(fileURL)
	if err != nil {
		return nil, fmt.Errorf("download file: %w", err)
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	return data, nil
}
