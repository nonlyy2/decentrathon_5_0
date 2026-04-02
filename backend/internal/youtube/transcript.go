package youtube

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/url"
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
	u := baseURL
	
	// Ищем любой существующий параметр fmt=... и заменяем его на fmt=xml
	re := regexp.MustCompile(`([?&])fmt=[^&]+`)
	if re.MatchString(u) {
		u = re.ReplaceAllString(u, "${1}fmt=xml")
	} else {
		// Если параметра нет вообще, добавляем его
		if strings.Contains(u, "?") {
			u += "&fmt=xml"
		} else {
			u += "?fmt=xml"
		}
	}

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
// Returns: transcript (may be empty string), isValid (video accessible), error.
func ValidateAndFetch(rawURL string) (transcript string, isValid bool, err error) {
	videoID, err := ExtractVideoID(rawURL)
	if err != nil {
		return "", false, err
	}

	isValid = CheckAccessibility(videoID)
	if !isValid {
		return "", false, fmt.Errorf("video is private, deleted, or not accessible")
	}

	// Перестаем игнорировать ошибку!
	transcript, err = FetchTranscript(videoID)
	if err != nil {
		// Возвращаем isValid = true (видео доступно), но с ошибкой парсинга
		return "", true, fmt.Errorf("failed to extract transcript: %w", err) 
	}
	
	return transcript, true, nil
}