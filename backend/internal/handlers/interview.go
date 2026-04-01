package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CreateTelegramInvite generates a deep-link for a candidate's Telegram interview.
// POST /candidates/:id/telegram-invite
func CreateTelegramInvite(pool *pgxpool.Pool, botUsername string) gin.HandlerFunc {
	return func(c *gin.Context) {
		candidateID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid candidate id"})
			return
		}

		// Check candidate exists and has Stage 1 score >= 65 (Recommended+)
		var status string
		var finalScore *float64
		var interviewStatus string
		err = pool.QueryRow(c.Request.Context(), `
			SELECT c.status, a.final_score, COALESCE(c.interview_status, 'not_invited')
			FROM candidates c
			LEFT JOIN analyses a ON a.candidate_id = c.id
			WHERE c.id = $1`, candidateID).Scan(&status, &finalScore, &interviewStatus)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "candidate not found"})
			return
		}

		override := c.Query("override") == "true"
		if !override && (finalScore == nil || *finalScore < 65) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "candidate must have Stage 1 score >= 65 (Recommended or above) to be invited"})
			return
		}

		if interviewStatus == "in_progress" || interviewStatus == "completed" {
			c.JSON(http.StatusConflict, gin.H{"error": "interview already " + interviewStatus})
			return
		}

		// Create or get existing invite
		var token string
		err = pool.QueryRow(c.Request.Context(), `
			INSERT INTO telegram_invites (candidate_id)
			VALUES ($1)
			ON CONFLICT (candidate_id) DO UPDATE SET
				status = 'pending',
				token = gen_random_uuid(),
				created_at = NOW(),
				expires_at = NOW() + INTERVAL '7 days',
				telegram_chat_id = NULL,
				linked_at = NULL
			RETURNING token`, candidateID).Scan(&token)
		if err != nil {
			log.Printf("Failed to create invite: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create invite"})
			return
		}

		// Update candidate interview status
		pool.Exec(c.Request.Context(), `
			UPDATE candidates SET interview_status = 'invited' WHERE id = $1`, candidateID)

		deepLink := ""
		if botUsername != "" {
			deepLink = "https://t.me/" + botUsername + "?start=" + token
		}

		c.JSON(http.StatusOK, gin.H{
			"token":     token,
			"deep_link": deepLink,
			"message":   "Invite created. Send the deep link to the candidate.",
		})
	}
}

// GetInterviewStatus returns the interview status and results for a candidate.
// GET /candidates/:id/interview
func GetInterviewStatus(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		candidateID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid candidate id"})
			return
		}

		// Get interview info
		var interview struct {
			ID             int     `json:"id"`
			Status         string  `json:"status"`
			Language       string  `json:"language"`
			QuestionsAsked int     `json:"questions_asked"`
			StartedAt      string  `json:"started_at"`
			CompletedAt    *string `json:"completed_at"`
			CurrentTopic   string  `json:"current_topic"`
		}
		err = pool.QueryRow(c.Request.Context(), `
			SELECT id, status, language, questions_asked,
				   started_at::text, completed_at::text, COALESCE(current_topic, '')
			FROM interviews WHERE candidate_id = $1`, candidateID).Scan(
			&interview.ID, &interview.Status, &interview.Language,
			&interview.QuestionsAsked, &interview.StartedAt, &interview.CompletedAt,
			&interview.CurrentTopic)
		if err != nil {
			// No interview yet — check for invite
			var inviteStatus *string
			var deepLink *string
			pool.QueryRow(c.Request.Context(), `
				SELECT status, token FROM telegram_invites WHERE candidate_id = $1`,
				candidateID).Scan(&inviteStatus, &deepLink)

			c.JSON(http.StatusOK, gin.H{
				"status":        "not_started",
				"invite_status": inviteStatus,
				"invite_token":  deepLink,
			})
			return
		}

		result := gin.H{
			"interview": interview,
		}

		// Get analysis if completed (or if interview finished but status not yet updated)
		{
			var analysis struct {
				ScoreLeadership        int      `json:"score_leadership"`
				ScoreGrit              int      `json:"score_grit"`
				ScoreAuthenticity      int      `json:"score_authenticity"`
				ScoreMotivation        int      `json:"score_motivation"`
				ScoreVision            int      `json:"score_vision"`
				FinalScore             float64  `json:"final_score"`
				Category               string   `json:"category"`
				ConsistencyScore       int      `json:"consistency_score"`
				StyleMatchScore        int      `json:"style_match_score"`
				SuspicionFlags         string   `json:"suspicion_flags_raw"`
				Summary                string   `json:"summary"`
				Strengths              []string `json:"strengths"`
				Concerns               []string `json:"concerns"`
				AnalyzedAt             string   `json:"analyzed_at"`
				ModelUsed              string   `json:"model_used"`
				ExplanationLeadership  string   `json:"explanation_leadership"`
				ExplanationGrit        string   `json:"explanation_grit"`
				ExplanationAuthenticity string  `json:"explanation_authenticity"`
				ExplanationMotivation  string   `json:"explanation_motivation"`
				ExplanationVision      string   `json:"explanation_vision"`
			}
			err = pool.QueryRow(c.Request.Context(), `
				SELECT score_leadership, score_grit, score_authenticity,
					   score_motivation, score_vision, final_score, category,
					   consistency_score, style_match_score,
					   COALESCE(suspicion_flags::text, '[]'),
					   COALESCE(summary, ''),
					   COALESCE(strengths, ARRAY[]::text[]),
					   COALESCE(concerns, ARRAY[]::text[]),
					   analyzed_at::text, COALESCE(model_used, ''),
					   COALESCE(explanation_leadership, ''),
					   COALESCE(explanation_grit, ''),
					   COALESCE(explanation_authenticity, ''),
					   COALESCE(explanation_motivation, ''),
					   COALESCE(explanation_vision, '')
				FROM interview_analyses WHERE candidate_id = $1`, candidateID).Scan(
				&analysis.ScoreLeadership, &analysis.ScoreGrit, &analysis.ScoreAuthenticity,
				&analysis.ScoreMotivation, &analysis.ScoreVision, &analysis.FinalScore,
				&analysis.Category, &analysis.ConsistencyScore, &analysis.StyleMatchScore,
				&analysis.SuspicionFlags, &analysis.Summary,
				&analysis.Strengths, &analysis.Concerns,
				&analysis.AnalyzedAt, &analysis.ModelUsed,
				&analysis.ExplanationLeadership, &analysis.ExplanationGrit,
				&analysis.ExplanationAuthenticity, &analysis.ExplanationMotivation,
				&analysis.ExplanationVision)
			if err == nil {
				// Parse suspicion flags from JSONB
				var flags []string
				json.Unmarshal([]byte(analysis.SuspicionFlags), &flags)

				result["analysis"] = gin.H{
					"score_leadership":        analysis.ScoreLeadership,
					"score_grit":              analysis.ScoreGrit,
					"score_authenticity":       analysis.ScoreAuthenticity,
					"score_motivation":        analysis.ScoreMotivation,
					"score_vision":            analysis.ScoreVision,
					"final_score":             analysis.FinalScore,
					"category":                analysis.Category,
					"consistency_score":       analysis.ConsistencyScore,
					"style_match_score":       analysis.StyleMatchScore,
					"suspicion_flags":         flags,
					"summary":                 analysis.Summary,
					"strengths":               analysis.Strengths,
					"concerns":                analysis.Concerns,
					"analyzed_at":             analysis.AnalyzedAt,
					"model_used":              analysis.ModelUsed,
					"explanation_leadership":   analysis.ExplanationLeadership,
					"explanation_grit":         analysis.ExplanationGrit,
					"explanation_authenticity": analysis.ExplanationAuthenticity,
					"explanation_motivation":   analysis.ExplanationMotivation,
					"explanation_vision":       analysis.ExplanationVision,
				}
			}

			// Get combined score
			var combinedScore *float64
			pool.QueryRow(c.Request.Context(), `
				SELECT combined_score FROM candidates WHERE id = $1`, candidateID).Scan(&combinedScore)
			result["combined_score"] = combinedScore
		}

		c.JSON(http.StatusOK, result)
	}
}

// ForceEvaluateInterview triggers evaluation for an interview that finished but wasn't evaluated.
// POST /candidates/:id/interview/evaluate
func ForceEvaluateInterview(pool *pgxpool.Pool, evaluateFn func(interviewID, candidateID int) error) gin.HandlerFunc {
	return func(c *gin.Context) {
		candidateID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid candidate id"})
			return
		}

		// Check interview exists
		var interviewID int
		var status string
		err = pool.QueryRow(c.Request.Context(),
			`SELECT id, status FROM interviews WHERE candidate_id = $1`, candidateID).Scan(&interviewID, &status)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "no interview found for this candidate"})
			return
		}

		// Check if already has analysis
		var hasAnalysis bool
		pool.QueryRow(c.Request.Context(),
			`SELECT EXISTS(SELECT 1 FROM interview_analyses WHERE candidate_id = $1)`, candidateID).Scan(&hasAnalysis)

		if hasAnalysis {
			c.JSON(http.StatusConflict, gin.H{"error": "interview already has analysis, use re-evaluate if needed"})
			return
		}

		// Launch evaluation async
		go func() {
			log.Printf("Force evaluation starting for interview %d (candidate %d)", interviewID, candidateID)
			if err := evaluateFn(interviewID, candidateID); err != nil {
				log.Printf("Force evaluation failed for interview %d: %v", interviewID, err)
			} else {
				log.Printf("Force evaluation completed for interview %d", interviewID)
			}
		}()

		c.JSON(http.StatusOK, gin.H{"message": "Evaluation started"})
	}
}

// ReEvaluateInterview re-runs evaluation for an interview (overwrites existing analysis).
// POST /candidates/:id/interview/re-evaluate
func ReEvaluateInterview(pool *pgxpool.Pool, evaluateFn func(interviewID, candidateID int) error) gin.HandlerFunc {
	return func(c *gin.Context) {
		candidateID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid candidate id"})
			return
		}

		var interviewID int
		err = pool.QueryRow(c.Request.Context(),
			`SELECT id FROM interviews WHERE candidate_id = $1`, candidateID).Scan(&interviewID)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "no interview found for this candidate"})
			return
		}

		go func() {
			log.Printf("Re-evaluation starting for interview %d (candidate %d)", interviewID, candidateID)
			if err := evaluateFn(interviewID, candidateID); err != nil {
				log.Printf("Re-evaluation failed for interview %d: %v", interviewID, err)
			} else {
				log.Printf("Re-evaluation completed for interview %d", interviewID)
			}
		}()

		c.JSON(http.StatusOK, gin.H{"message": "Re-evaluation started"})
	}
}

// GetInterviewTranscript returns the full message history for an interview.
// GET /candidates/:id/interview/messages
func GetInterviewTranscript(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		candidateID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid candidate id"})
			return
		}

		rows, err := pool.Query(c.Request.Context(), `
			SELECT im.role, im.content, im.message_type,
				   COALESCE(im.voice_duration_sec, 0),
				   COALESCE(im.response_time_sec, 0),
				   im.created_at::text
			FROM interview_messages im
			JOIN interviews i ON i.id = im.interview_id
			WHERE i.candidate_id = $1
			ORDER BY im.created_at ASC`, candidateID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch transcript"})
			return
		}
		defer rows.Close()

		type message struct {
			Role            string `json:"role"`
			Content         string `json:"content"`
			MessageType     string `json:"message_type"`
			VoiceDurationSec int   `json:"voice_duration_sec"`
			ResponseTimeSec int    `json:"response_time_sec"`
			CreatedAt       string `json:"created_at"`
		}

		var messages []message
		for rows.Next() {
			var m message
			if err := rows.Scan(&m.Role, &m.Content, &m.MessageType,
				&m.VoiceDurationSec, &m.ResponseTimeSec, &m.CreatedAt); err != nil {
				continue
			}
			messages = append(messages, m)
		}

		if messages == nil {
			messages = []message{}
		}

		c.JSON(http.StatusOK, messages)
	}
}
