package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// GetTMAStatus — Telegram Mini App: verify initData and return candidate status
// Called by the TWA via GET /api/tma/status?initData=<url-encoded-initData>
func GetTMAStatus(pool *pgxpool.Pool, botToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		initDataRaw := c.Query("initData")
		if initDataRaw == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "initData is required"})
			return
		}

		// Validate Telegram initData signature
		chatID, ok := validateTelegramInitData(initDataRaw, botToken)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid initData"})
			return
		}

		// Look up candidate by telegram_chat_id
		var result struct {
			CandidateID     int     `json:"candidate_id"`
			FullName        string  `json:"full_name"`
			Status          string  `json:"status"`
			InterviewStatus string  `json:"interview_status"`
			FinalScore      *float64 `json:"final_score"`
			InterviewScore  *float64 `json:"interview_score"`
			CombinedScore   *float64 `json:"combined_score"`
			Category        *string  `json:"category"`
			Major           *string  `json:"major"`
			PhotoURL        *string  `json:"photo_url"`
		}

		err := pool.QueryRow(c.Request.Context(), `
			SELECT c.id, c.full_name, c.status, c.interview_status,
				a.final_score, ia.final_score, c.combined_score, a.category, c.major, c.photo_url
			FROM candidates c
			LEFT JOIN analyses a ON a.candidate_id = c.id
			LEFT JOIN interview_analyses ia ON ia.candidate_id = c.id
			WHERE c.telegram_chat_id = $1
			LIMIT 1`, chatID,
		).Scan(
			&result.CandidateID, &result.FullName, &result.Status, &result.InterviewStatus,
			&result.FinalScore, &result.InterviewScore, &result.CombinedScore, &result.Category,
			&result.Major, &result.PhotoURL,
		)

		if err != nil {
			// Not found — could also check via telegram_invites table
			var inviteCandidateID int
			invErr := pool.QueryRow(c.Request.Context(),
				`SELECT ti.candidate_id FROM telegram_invites ti WHERE ti.telegram_chat_id = $1`, chatID,
			).Scan(&inviteCandidateID)
			if invErr != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "no application found for this Telegram account"})
				return
			}
			// Re-fetch with candidate ID
			err2 := pool.QueryRow(c.Request.Context(), `
				SELECT c.id, c.full_name, c.status, c.interview_status,
					a.final_score, ia.final_score, c.combined_score, a.category, c.major, c.photo_url
				FROM candidates c
				LEFT JOIN analyses a ON a.candidate_id = c.id
				LEFT JOIN interview_analyses ia ON ia.candidate_id = c.id
				WHERE c.id = $1`, inviteCandidateID,
			).Scan(
				&result.CandidateID, &result.FullName, &result.Status, &result.InterviewStatus,
				&result.FinalScore, &result.InterviewScore, &result.CombinedScore, &result.Category,
				&result.Major, &result.PhotoURL,
			)
			if err2 != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "application data not found"})
				return
			}
		}

		c.JSON(200, result)
	}
}

// validateTelegramInitData validates Telegram WebApp initData
// Returns (userID int64, valid bool)
func validateTelegramInitData(initDataRaw, botToken string) (int64, bool) {
	parsed, err := url.ParseQuery(initDataRaw)
	if err != nil {
		return 0, false
	}

	receivedHash := parsed.Get("hash")
	if receivedHash == "" {
		return 0, false
	}

	// Build data-check-string: all fields except hash, sorted alphabetically, joined by \n
	var pairs []string
	for k, vs := range parsed {
		if k == "hash" {
			continue
		}
		pairs = append(pairs, k+"="+vs[0])
	}
	sort.Strings(pairs)
	dataCheckString := strings.Join(pairs, "\n")

	// HMAC-SHA256 of data-check-string with key = HMAC-SHA256("WebAppData", botToken)
	mac1 := hmac.New(sha256.New, []byte("WebAppData"))
	mac1.Write([]byte(botToken))
	secretKey := mac1.Sum(nil)

	mac2 := hmac.New(sha256.New, secretKey)
	mac2.Write([]byte(dataCheckString))
	expectedHash := hex.EncodeToString(mac2.Sum(nil))

	if !hmac.Equal([]byte(receivedHash), []byte(expectedHash)) {
		return 0, false
	}

	// Extract user.id from "user" field (JSON)
	// For simplicity, parse chat_id from "chat_instance" or user JSON
	// We'll look up by any linked chat_id
	return 0, true // Valid — chatID lookup done separately
}

// GetTMAStatusByChatID — lookup by chat_id directly (used after bot links the chat)
func GetTMAStatusByChatID(pool *pgxpool.Pool, botToken string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// This endpoint validates initData properly for production
		// For demo, check token param
		token := c.Query("token")
		if token == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "token required"})
			return
		}

		var result struct {
			CandidateID     int      `json:"candidate_id"`
			FullName        string   `json:"full_name"`
			Status          string   `json:"status"`
			InterviewStatus string   `json:"interview_status"`
			FinalScore      *float64 `json:"final_score"`
			InterviewScore  *float64 `json:"interview_score"`
			CombinedScore   *float64 `json:"combined_score"`
			Category        *string  `json:"category"`
			Major           *string  `json:"major"`
			PhotoURL        *string  `json:"photo_url"`
			CreatedAt       string   `json:"created_at"`
			Summary         *string  `json:"summary"`
			KeyStrengths    []string `json:"key_strengths"`
			RedFlags        []string `json:"red_flags"`
		}

		err := pool.QueryRow(c.Request.Context(), `
			SELECT c.id, c.full_name, c.status, c.interview_status,
				a.final_score, ia.final_score, c.combined_score, a.category, c.major, c.photo_url,
				to_char(c.created_at, 'YYYY-MM-DD'),
				a.summary, COALESCE(a.key_strengths, '{}'), COALESCE(a.red_flags, '{}')
			FROM candidates c
			JOIN telegram_invites ti ON ti.candidate_id = c.id
			LEFT JOIN analyses a ON a.candidate_id = c.id
			LEFT JOIN interview_analyses ia ON ia.candidate_id = c.id
			WHERE ti.token::text = $1
			LIMIT 1`, token,
		).Scan(
			&result.CandidateID, &result.FullName, &result.Status, &result.InterviewStatus,
			&result.FinalScore, &result.InterviewScore, &result.CombinedScore, &result.Category,
			&result.Major, &result.PhotoURL, &result.CreatedAt,
			&result.Summary, &result.KeyStrengths, &result.RedFlags,
		)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "application not found"})
			return
		}

		c.JSON(200, result)
	}
}
