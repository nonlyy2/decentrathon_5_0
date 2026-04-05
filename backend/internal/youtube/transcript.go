package youtube

import (
	"bytes"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var videoIDRegex = regexp.MustCompile(`(?:youtube\.com/watch\?(?:.*&)?v=|youtu\.be/|youtube\.com/embed/)([a-zA-Z0-9_-]{11})`)

// httpClient shared, with browser-like headers
var httpClient = &http.Client{Timeout: 15 * time.Second}

func browserGet(rawURL string) (*http.Response, error) {
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0 Safari/537.36")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")
	return httpClient.Do(req)
}

// ExtractVideoID extracts the 11-char video ID from any YouTube URL format.
func ExtractVideoID(rawURL string) (string, error) {
	matches := videoIDRegex.FindStringSubmatch(rawURL)
	if len(matches) < 2 {
		return "", fmt.Errorf("invalid YouTube URL")
	}
	return matches[1], nil
}

// CheckAccessibility returns true if the video is publicly viewable.
// Only HTTP 404 from oEmbed counts as "definitely invalid".
// Network/timeout/other errors → assume valid to avoid false negatives.
func CheckAccessibility(videoID string) bool {
	videoURL := url.QueryEscape("https://www.youtube.com/watch?v=" + videoID)
	oembedURL := "https://www.youtube.com/oembed?url=" + videoURL + "&format=json"

	resp, err := browserGet(oembedURL)
	if err != nil {
		return true // network error → can't verify, assume valid
	}
	defer resp.Body.Close()
	return resp.StatusCode != 404
}

// captionTrack is part of the ytInitialPlayerResponse captions JSON.
type captionTrack struct {
	BaseURL      string `json:"baseUrl"`
	LanguageCode string `json:"languageCode"`
	Kind         string `json:"kind"` // "asr" = auto-generated
}

// FetchTranscript fetches plain-text transcript by:
//  1. Scraping the video watch page for ytInitialPlayerResponse
//  2. Extracting captionTracks from it
//  3. Fetching and parsing the timed-text XML from the best available track
func FetchTranscript(videoID string) (string, error) {
	watchURL := "https://www.youtube.com/watch?v=" + videoID
	resp, err := browserGet(watchURL)
	if err != nil {
		return "", fmt.Errorf("failed to fetch watch page: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 4*1024*1024)) // 4 MB cap
	if err != nil {
		return "", fmt.Errorf("failed to read watch page: %w", err)
	}
	page := string(body)

	// Extract captionTracks JSON array from ytInitialPlayerResponse
	tracks, err := extractCaptionTracks(page)
	if err != nil || len(tracks) == 0 {
		return "", fmt.Errorf("no captions found for video %s", videoID)
	}

	// Prefer: en manual > en ASR > any manual > any ASR
	best := chooseBestTrack(tracks)
	if best == nil {
		return "", fmt.Errorf("no suitable caption track found")
	}

	return fetchTimedText(best.BaseURL)
}

// extractCaptionTracks pulls caption tracks from ytInitialPlayerResponse.
var (
	captionTracksRe = regexp.MustCompile(`(?s)"captionTracks"\s*:\s*(\[.*?\])`)
	initRespMarkers = []string{
		"ytInitialPlayerResponse = ",
		"var ytInitialPlayerResponse = ",
		"window[\"ytInitialPlayerResponse\"] = ",
	}
)

func extractCaptionTracks(page string) ([]captionTrack, error) {
	if initialJSON := extractInitialPlayerResponseJSON(page); initialJSON != "" {
		var parsed struct {
			Captions struct {
				PlayerCaptionsTracklistRenderer struct {
					CaptionTracks []captionTrack `json:"captionTracks"`
				} `json:"playerCaptionsTracklistRenderer"`
			} `json:"captions"`
		}
		if err := json.Unmarshal([]byte(initialJSON), &parsed); err == nil && len(parsed.Captions.PlayerCaptionsTracklistRenderer.CaptionTracks) > 0 {
			return parsed.Captions.PlayerCaptionsTracklistRenderer.CaptionTracks, nil
		}
	}

	// Fallback for alternative page layouts.
	m := captionTracksRe.FindStringSubmatch(page)
	if len(m) < 2 {
		return nil, fmt.Errorf("captionTracks not found in page")
	}
	raw := strings.ReplaceAll(m[1], `\u0026`, "&")

	var tracks []captionTrack
	if err := json.Unmarshal([]byte(raw), &tracks); err != nil {
		return nil, fmt.Errorf("failed to parse captionTracks: %w", err)
	}
	return tracks, nil
}

func extractInitialPlayerResponseJSON(page string) string {
	for _, marker := range initRespMarkers {
		idx := strings.Index(page, marker)
		if idx == -1 {
			continue
		}
		start := strings.Index(page[idx+len(marker):], "{")
		if start == -1 {
			continue
		}
		start += idx + len(marker)
		return extractBalancedJSONObject(page[start:])
	}
	return ""
}

func extractBalancedJSONObject(s string) string {
	if s == "" || s[0] != '{' {
		return ""
	}
	depth := 0
	inString := false
	escaped := false
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if inString {
			if escaped {
				escaped = false
				continue
			}
			if ch == '\\' {
				escaped = true
				continue
			}
			if ch == '"' {
				inString = false
			}
			continue
		}
		if ch == '"' {
			inString = true
			continue
		}
		if ch == '{' {
			depth++
		} else if ch == '}' {
			depth--
			if depth == 0 {
				return s[:i+1]
			}
		}
	}
	return ""
}

func chooseBestTrack(tracks []captionTrack) *captionTrack {
	// Priority order
	for _, lang := range []string{"en", "en-US", "en-GB"} {
		for i := range tracks {
			if tracks[i].LanguageCode == lang && tracks[i].Kind != "asr" {
				return &tracks[i]
			}
		}
	}
	for _, lang := range []string{"en", "en-US", "en-GB"} {
		for i := range tracks {
			if tracks[i].LanguageCode == lang {
				return &tracks[i]
			}
		}
	}
	// Any manual track
	for i := range tracks {
		if tracks[i].Kind != "asr" {
			return &tracks[i]
		}
	}
	// Any track
	if len(tracks) > 0 {
		return &tracks[0]
	}
	return nil
}

// timedTextXML maps the XML returned by the caption track URL.
type timedTextXML struct {
	XMLName xml.Name        `xml:"transcript"`
	Texts   []timedTextNode `xml:"text"`
}

type timedTextNode struct {
	Value string `xml:",chardata"`
}

func fetchTimedText(baseURL string) (string, error) {
	// Always force XML format. YouTube often returns base URLs with fmt=json3/srv3.
	parsed, err := url.Parse(baseURL)
	if err != nil {
		return "", fmt.Errorf("invalid timed text URL: %w", err)
	}
	q := parsed.Query()
	q.Set("fmt", "xml")
	parsed.RawQuery = q.Encode()
	u := parsed.String()

	resp, err := browserGet(u)
	if err != nil {
		return "", fmt.Errorf("failed to fetch timed text: %w", err)
	}
	defer resp.Body.Close()

	xmlBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read timed text: %w", err)
	}

	var tt timedTextXML
	if err := xml.Unmarshal(xmlBytes, &tt); err != nil || len(tt.Texts) == 0 {
		return "", fmt.Errorf("failed to parse timed text XML: %w", err)
	}

	var sb strings.Builder
	for _, node := range tt.Texts {
		text := strings.TrimSpace(html.UnescapeString(node.Value))
		if text != "" {
			sb.WriteString(text)
			sb.WriteString(" ")
		}
	}
	return strings.TrimSpace(sb.String()), nil
}

// ValidateAndFetch validates the YouTube URL and fetches its transcript.
// If no captions are found, falls back to audio download + STT if credentials are provided.
// Returns: transcript (may be empty string), isValid (video accessible), error.
func ValidateAndFetch(rawURL string, sttAPIKey, sttProvider string) (transcript string, isValid bool, err error) {
	videoID, err := ExtractVideoID(rawURL)
	if err != nil {
		return "", false, err
	}

	isValid = CheckAccessibility(videoID)
	if !isValid {
		return "", false, fmt.Errorf("video is private, deleted, or not accessible")
	}

	transcript, err = FetchTranscript(videoID)
	if err != nil {
		// Fallback 1: try yt-dlp subtitle extraction (more reliable than page scraping)
		transcript, err = FetchTranscriptViaYtDlpSubs(videoID)
		if err != nil {
			// Fallback 2: audio-based transcription if STT is configured
			if sttAPIKey != "" {
				transcript, err = FetchTranscriptViaAudio(videoID, sttAPIKey, sttProvider)
				if err != nil {
					return "", true, fmt.Errorf("captions unavailable, audio transcription failed: %w", err)
				}
				return transcript, true, nil
			}
			return "", true, fmt.Errorf("no captions available (configure WHISPER_API_KEY to enable audio transcription)")
		}
	}

	return transcript, true, nil
}

// FetchTranscriptViaYtDlpSubs uses yt-dlp to download subtitles (more reliable than page scraping).
func FetchTranscriptViaYtDlpSubs(videoID string) (string, error) {
	if _, err := exec.LookPath("yt-dlp"); err != nil {
		return "", fmt.Errorf("yt-dlp not installed")
	}

	tmpDir, err := os.MkdirTemp("", "ytsubs-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	videoURL := "https://www.youtube.com/watch?v=" + videoID
	outTemplate := filepath.Join(tmpDir, "subs")

	// Try manual subs first, then auto-generated
	for _, args := range [][]string{
		{"--write-subs", "--sub-langs", "en,en-US,en-GB,ru,kk", "--skip-download", "--sub-format", "vtt/srt/best", "-o", outTemplate, videoURL},
		{"--write-auto-subs", "--sub-langs", "en,en-US,en-GB,ru,kk", "--skip-download", "--sub-format", "vtt/srt/best", "-o", outTemplate, videoURL},
	} {
		cmd := exec.Command("yt-dlp", args...)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		_ = cmd.Run() // ignore error, check for files

		// Look for any subtitle file
		matches, _ := filepath.Glob(filepath.Join(tmpDir, "subs.*"))
		for _, m := range matches {
			ext := strings.ToLower(filepath.Ext(m))
			if ext == ".vtt" || ext == ".srt" {
				data, err := os.ReadFile(m)
				if err != nil {
					continue
				}
				text := parseSubtitleFile(string(data))
				if len(text) > 20 {
					return text, nil
				}
			}
		}
	}

	return "", fmt.Errorf("yt-dlp found no subtitles for video %s", videoID)
}

// parseSubtitleFile extracts plain text from VTT/SRT subtitle content.
func parseSubtitleFile(content string) string {
	lines := strings.Split(content, "\n")
	var sb strings.Builder
	seen := make(map[string]bool)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip empty lines, timestamps, sequence numbers, and VTT headers
		if line == "" || line == "WEBVTT" || strings.HasPrefix(line, "Kind:") || strings.HasPrefix(line, "Language:") {
			continue
		}
		// Skip timestamp lines (00:00:01.000 --> 00:00:04.000)
		if strings.Contains(line, "-->") {
			continue
		}
		// Skip pure numeric lines (SRT sequence numbers)
		if isNumericLine(line) {
			continue
		}
		// Strip VTT tags like <c>, </c>, <00:00:01.000>
		cleaned := stripVTTTags(line)
		cleaned = strings.TrimSpace(cleaned)
		if cleaned == "" {
			continue
		}
		// Deduplicate (auto-subs often repeat lines)
		if !seen[cleaned] {
			seen[cleaned] = true
			sb.WriteString(cleaned)
			sb.WriteString(" ")
		}
	}
	return strings.TrimSpace(sb.String())
}

func isNumericLine(s string) bool {
	for _, c := range s {
		if c < '0' || c > '9' {
			return false
		}
	}
	return len(s) > 0
}

var vttTagRe = regexp.MustCompile(`<[^>]+>`)

func stripVTTTags(s string) string {
	return vttTagRe.ReplaceAllString(s, "")
}

// FetchTranscriptViaAudio downloads the video audio using yt-dlp and transcribes it via Whisper/Alem STT.
func FetchTranscriptViaAudio(videoID, apiKey, provider string) (string, error) {
	if _, err := exec.LookPath("yt-dlp"); err != nil {
		return "", fmt.Errorf("yt-dlp not installed — cannot extract audio")
	}

	// Create temp dir
	tmpDir, err := os.MkdirTemp("", "ytaudio-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	audioPath := filepath.Join(tmpDir, "audio.mp3")
	videoURL := "https://www.youtube.com/watch?v=" + videoID

	// Download audio only (max 50 MB, bail fast on long videos)
	cmd := exec.Command("yt-dlp",
		"--extract-audio",
		"--audio-format", "mp3",
		"--audio-quality", "5",
		"--max-filesize", "50m",
		"--no-playlist",
		"-o", audioPath,
		videoURL,
	)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("yt-dlp failed: %v — %s", err, stderr.String())
	}

	audioData, err := os.ReadFile(audioPath)
	if err != nil {
		return "", fmt.Errorf("failed to read downloaded audio: %w", err)
	}

	return transcribeAudio(audioData, "audio.mp3", apiKey, provider)
}

// transcribeAudio sends audio bytes to Whisper (OpenAI) or Alem STT.
func transcribeAudio(audioData []byte, filename, apiKey, provider string) (string, error) {
	var apiURL string
	switch provider {
	case "alem":
		apiURL = "https://llm.alem.ai/v1/audio/transcriptions"
	default: // openai / whisper
		apiURL = "https://api.openai.com/v1/audio/transcriptions"
	}

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	part, err := w.CreateFormFile("file", filename)
	if err != nil {
		return "", err
	}
	if _, err := part.Write(audioData); err != nil {
		return "", err
	}
	_ = w.WriteField("model", "whisper-1")
	w.Close()

	req, err := http.NewRequest("POST", apiURL, &body)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", w.FormDataContentType())

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("STT request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("STT API error %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("parse STT response: %w", err)
	}
	return strings.TrimSpace(result.Text), nil
}