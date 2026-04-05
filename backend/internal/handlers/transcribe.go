package handlers

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

	"github.com/gin-gonic/gin"
)

// TranscribeAudio accepts a browser-recorded audio file (webm/ogg/wav) and returns
// the transcribed text using the Alem Plus Speech-to-Text API.
func TranscribeAudio(alemAPIKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if alemAPIKey == "" {
			c.JSON(503, gin.H{"error": "speech-to-text service not configured"})
			return
		}

		file, _, err := c.Request.FormFile("audio")
		if err != nil {
			c.JSON(400, gin.H{"error": "audio file required"})
			return
		}
		defer file.Close()

		audioData, err := io.ReadAll(file)
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to read audio"})
			return
		}

		// Convert to 16kHz mono WAV — Alem STT requires WAV format
		wavData, err := convertAudioToWAV(audioData)
		if err != nil {
			c.JSON(500, gin.H{"error": "audio conversion failed: " + err.Error()})
			return
		}

		ctx, cancel := context.WithTimeout(c.Request.Context(), 60*time.Second)
		defer cancel()

		text, err := alemTranscribe(ctx, alemAPIKey, wavData)
		if err != nil {
			c.JSON(500, gin.H{"error": "transcription failed: " + err.Error()})
			return
		}

		c.JSON(200, gin.H{"text": text})
	}
}

// convertAudioToWAV converts any audio format to 16kHz mono WAV using ffmpeg.
// ffmpeg auto-detects the input format from the stream header.
func convertAudioToWAV(audioData []byte) ([]byte, error) {
	if _, err := exec.LookPath("ffmpeg"); err != nil {
		return nil, fmt.Errorf("ffmpeg is not installed")
	}
	cmd := exec.Command("ffmpeg", "-hide_banner", "-loglevel", "error",
		"-i", "pipe:0",
		"-ar", "16000", "-ac", "1", "-f", "wav",
		"pipe:1")
	cmd.Stdin = bytes.NewReader(audioData)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("ffmpeg error: %v — %s", err, stderr.String())
	}
	return stdout.Bytes(), nil
}

type alemTranscribeResponse struct {
	Text string `json:"text"`
}

// alemTranscribe sends WAV audio to the Alem Plus STT API and returns the text.
func alemTranscribe(ctx context.Context, apiKey string, wavData []byte) (string, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	part, err := writer.CreateFormFile("file", "voice.wav")
	if err != nil {
		return "", err
	}
	if _, err := part.Write(wavData); err != nil {
		return "", err
	}
	_ = writer.WriteField("model", "speech-to-text")
	writer.Close()

	req, err := http.NewRequestWithContext(ctx, "POST", "https://llm.alem.ai/v1/audio/transcriptions", &body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("alem API error (%d): %s", resp.StatusCode, string(respBody))
	}

	var result alemTranscribeResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", err
	}
	return result.Text, nil
}
