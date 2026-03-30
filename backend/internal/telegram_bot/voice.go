package telegram_bot

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os/exec"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// AlemSTTClient handles audio transcription via Alem Plus speech-to-text API.
type AlemSTTClient struct {
	apiKey     string
	httpClient *http.Client
}

// NewAlemSTTClient creates an Alem Plus speech-to-text transcription client.
func NewAlemSTTClient(apiKey string) *AlemSTTClient {
	return &AlemSTTClient{
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 120 * time.Second},
	}
}

type alemSTTResponse struct {
	Text string `json:"text"`
}

// convertOGGToWAV converts OGG/Opus audio (from Telegram voice messages) to WAV using ffmpeg.
// Returns the original data unchanged if ffmpeg is not installed.
func convertOGGToWAV(oggData []byte) ([]byte, string, error) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		// ffmpeg not available — return original OGG data
		return oggData, "voice.ogg", nil
	}
	cmd := exec.Command("ffmpeg", "-hide_banner", "-loglevel", "error",
		"-i", "pipe:0",
		"-ar", "16000", "-ac", "1", "-f", "wav",
		"pipe:1")
	cmd.Stdin = bytes.NewReader(oggData)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, "", fmt.Errorf("ffmpeg: %v — %s", err, stderr.String())
	}
	return stdout.Bytes(), "voice.wav", nil
}

// Transcribe sends audio data to Alem Plus speech-to-text API and returns the transcribed text.
func (w *AlemSTTClient) Transcribe(ctx context.Context, audioData []byte, language string) (string, error) {
	// Alem STT does not yet support OGG — convert to WAV first
	audioData, filename, err := convertOGGToWAV(audioData)
	if err != nil {
		return "", fmt.Errorf("audio conversion: %w", err)
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return "", fmt.Errorf("create form file: %w", err)
	}
	if _, err := part.Write(audioData); err != nil {
		return "", fmt.Errorf("write audio data: %w", err)
	}

	_ = writer.WriteField("model", "speech-to-text")
	if language != "" {
		langMap := map[string]string{"en": "en", "ru": "ru", "kz": "kk"}
		if code, ok := langMap[language]; ok {
			_ = writer.WriteField("language", code)
		}
	}
	writer.Close()

	req, err := http.NewRequestWithContext(ctx, "POST", "https://llm.alem.ai/v1/audio/transcriptions", &body)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+w.apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := w.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("alem stt request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("alem stt API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	var result alemSTTResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("parse alem stt response: %w", err)
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
