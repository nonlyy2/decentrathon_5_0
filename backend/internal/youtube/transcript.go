package youtube

import (
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

// ExtractVideoID extracts the 11-char video ID from any YouTube URL format.
func ExtractVideoID(rawURL string) (string, error) {
	matches := videoIDRegex.FindStringSubmatch(rawURL)
	if len(matches) < 2 {
		return "", fmt.Errorf("invalid YouTube URL")
	}
	return matches[1], nil
}

// CheckAccessibility returns true if the video is publicly viewable.
// Uses YouTube's oEmbed endpoint (no API key required).
// Only returns false when YouTube explicitly responds with 404 (video not found/private).
// Network errors, timeouts, or rate-limit responses are treated as "assume valid"
// to avoid false negatives from server-side IP blocks.
func CheckAccessibility(videoID string) bool {
	client := &http.Client{Timeout: 10 * time.Second}
	videoURL := url.QueryEscape("https://www.youtube.com/watch?v=" + videoID)
	oembedURL := "https://www.youtube.com/oembed?url=" + videoURL + "&format=json"

	req, err := http.NewRequest("GET", oembedURL, nil)
	if err != nil {
		return true // assume valid on request build failure
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (compatible; bot)")

	resp, err := client.Do(req)
	if err != nil {
		// Network error / timeout — cannot verify, assume valid
		return true
	}
	defer resp.Body.Close()

	// 404 = video not found, deleted, or private
	// Any other code (200, 401, 403, 429…) — treat as valid/unknown
	return resp.StatusCode != 404
}

// timedTextResponse maps the XML returned by the YouTube timedtext API.
type timedTextResponse struct {
	XMLName xml.Name        `xml:"transcript"`
	Texts   []timedTextNode `xml:"text"`
}

type timedTextNode struct {
	Value string `xml:",chardata"`
}

// FetchTranscript tries to retrieve a plain-text transcript for the video.
// It cycles through common language codes and returns the first non-empty result.
func FetchTranscript(videoID string) (string, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	langs := []string{"en", "en-US", "en-GB", "ru", "kk"}

	for _, lang := range langs {
		u := fmt.Sprintf("https://www.youtube.com/api/timedtext?lang=%s&v=%s", lang, videoID)
		resp, err := client.Get(u)
		if err != nil {
			continue
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil || len(body) == 0 {
			continue
		}

		var tt timedTextResponse
		if err := xml.Unmarshal(body, &tt); err != nil || len(tt.Texts) == 0 {
			continue
		}

		var sb strings.Builder
		for _, node := range tt.Texts {
			text := strings.TrimSpace(html.UnescapeString(node.Value))
			if text != "" {
				sb.WriteString(text)
				sb.WriteString(" ")
			}
		}
		result := strings.TrimSpace(sb.String())
		if result != "" {
			return result, nil
		}
	}

	return "", fmt.Errorf("no transcript available for video %s", videoID)
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

	transcript, _ = FetchTranscript(videoID)
	return transcript, true, nil
}
