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

// общий HTTP клиент с browser-заголовками
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

// извлекает 11-символьный ID из любого формата YouTube URL
func ExtractVideoID(rawURL string) (string, error) {
	matches := videoIDRegex.FindStringSubmatch(rawURL)
	if len(matches) < 2 {
		return "", fmt.Errorf("invalid YouTube URL")
	}
	return matches[1], nil
}

// true если видео доступно; 404 из oEmbed = невалидное; сетевые ошибки → assume valid
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

// часть ytInitialPlayerResponse captions JSON
type captionTrack struct {
	BaseURL      string `json:"baseUrl"`
	LanguageCode string `json:"languageCode"`
	Kind         string `json:"kind"` // "asr" = автогенерация
}

// получает текст субтитров: скрейпинг watch-страницы → captionTracks → timed-text XML
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

	// извлекаем captionTracks из ytInitialPlayerResponse
	tracks, err := extractCaptionTracks(page)
	if err != nil || len(tracks) == 0 {
		return "", fmt.Errorf("no captions found for video %s", videoID)
	}

	// приоритет: en ручные > en ASR > любые ручные > любые ASR
	best := chooseBestTrack(tracks)
	if best == nil {
		return "", fmt.Errorf("no suitable caption track found")
	}

	return fetchTimedText(best.BaseURL)
}

// извлекает caption tracks из ytInitialPlayerResponse
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

	// фоллбэк для нестандартных разметок страницы
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
	// порядок приоритета
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
	// любые ручные
	for i := range tracks {
		if tracks[i].Kind != "asr" {
			return &tracks[i]
		}
	}
	// любые
	if len(tracks) > 0 {
		return &tracks[0]
	}
	return nil
}

// XML от caption track URL
type timedTextXML struct {
	XMLName xml.Name        `xml:"transcript"`
	Texts   []timedTextNode `xml:"text"`
}

type timedTextNode struct {
	Value string `xml:",chardata"`
}

func fetchTimedText(baseURL string) (string, error) {
	// принудительно XML: YouTube часто возвращает fmt=json3/srv3
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

// проверяет URL и получает транскрипт; фоллбэк на аудио+STT если субтитры недоступны
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
		// фоллбэк 1: innertube API
		transcript, err = FetchTranscriptViaInnertube(videoID)
		if err != nil {
			// фоллбэк 2: yt-dlp субтитры
			transcript, err = FetchTranscriptViaYtDlpSubs(videoID)
			if err != nil {
				// фоллбэк 3: аудио STT
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
	}

	return transcript, true, nil
}

// использует innertube API — надёжнее скрейпинга при изменениях разметки
func FetchTranscriptViaInnertube(videoID string) (string, error) {
	payload := fmt.Sprintf(`{
		"context": {
			"client": {
				"clientName": "WEB",
				"clientVersion": "2.20240101.00.00",
				"hl": "en"
			}
		},
		"videoId": "%s"
	}`, videoID)

	req, err := http.NewRequest("POST", "https://www.youtube.com/youtubei/v1/player?prettyPrint=false", strings.NewReader(payload))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0 Safari/537.36")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("innertube request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
	if err != nil {
		return "", err
	}

	var result struct {
		Captions struct {
			PlayerCaptionsTracklistRenderer struct {
				CaptionTracks []captionTrack `json:"captionTracks"`
			} `json:"playerCaptionsTracklistRenderer"`
		} `json:"captions"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse innertube response: %w", err)
	}

	tracks := result.Captions.PlayerCaptionsTracklistRenderer.CaptionTracks
	if len(tracks) == 0 {
		return "", fmt.Errorf("no captions found via innertube for video %s", videoID)
	}

	best := chooseBestTrack(tracks)
	if best == nil {
		return "", fmt.Errorf("no suitable caption track found")
	}

	return fetchTimedText(best.BaseURL)
}

// скачивает субтитры через yt-dlp
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

	// сначала ручные субтитры, потом авто
	for _, args := range [][]string{
		{"--write-subs", "--sub-langs", "en,en-US,en-GB,ru,kk", "--skip-download", "--sub-format", "vtt/srt/best", "-o", outTemplate, videoURL},
		{"--write-auto-subs", "--sub-langs", "en,en-US,en-GB,ru,kk", "--skip-download", "--sub-format", "vtt/srt/best", "-o", outTemplate, videoURL},
	} {
		cmd := exec.Command("yt-dlp", args...)
		var stderr bytes.Buffer
		cmd.Stderr = &stderr
		_ = cmd.Run() // ошибку игнорируем — проверяем файлы

		// ищем файлы субтитров
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

// извлекает plain text из VTT/SRT
func parseSubtitleFile(content string) string {
	lines := strings.Split(content, "\n")
	var sb strings.Builder
	seen := make(map[string]bool)

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// пустые, временны́е метки, VTT-заголовки
		if line == "" || line == "WEBVTT" || strings.HasPrefix(line, "Kind:") || strings.HasPrefix(line, "Language:") {
			continue
		}
		// строки тайм-кодов
		if strings.Contains(line, "-->") {
			continue
		}
		// порядковые номера SRT
		if isNumericLine(line) {
			continue
		}
		// убираем VTT-теги: <c>, </c>, <00:00:01.000>
		cleaned := stripVTTTags(line)
		cleaned = strings.TrimSpace(cleaned)
		if cleaned == "" {
			continue
		}
		// дедупликация (авто-субтитры часто дублируют строки)
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

// скачивает аудио через yt-dlp и транскрибирует через Whisper/Alem STT
func FetchTranscriptViaAudio(videoID, apiKey, provider string) (string, error) {
	if _, err := exec.LookPath("yt-dlp"); err != nil {
		return "", fmt.Errorf("yt-dlp not installed — cannot extract audio")
	}

	tmpDir, err := os.MkdirTemp("", "ytaudio-*")
	if err != nil {
		return "", fmt.Errorf("failed to create temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	audioPath := filepath.Join(tmpDir, "audio.mp3")
	videoURL := "https://www.youtube.com/watch?v=" + videoID

	// только аудио, лимит 50 МБ
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

// отправляет аудио в Whisper (OpenAI) или Alem STT
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