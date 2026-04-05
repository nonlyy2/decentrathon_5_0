package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type AnalysisHistoryEntry struct {
	ID                int       `json:"id"`
	CandidateID       int       `json:"candidate_id"`
	ScoreLeadership   int       `json:"score_leadership"`
	ScoreMotivation   int       `json:"score_motivation"`
	ScoreGrowth       int       `json:"score_growth"`
	ScoreVision       int       `json:"score_vision"`
	ScoreCommunication int      `json:"score_communication"`
	FinalScore        float64   `json:"final_score"`
	Category          string    `json:"category"`
	AIGeneratedRisk   string    `json:"ai_generated_risk"`
	AIGeneratedScore  int       `json:"ai_generated_score"`
	Summary           *string   `json:"summary"`
	KeyStrengths      []string  `json:"key_strengths"`
	RedFlags          []string  `json:"red_flags"`
	ModelUsed         *string   `json:"model_used"`
	AnalyzedAt        time.Time `json:"analyzed_at"`
	DurationMs        int       `json:"duration_ms"`
}

// GetAnalysisHistory возвращает историю анализов кандидата
func GetAnalysisHistory(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		candidateID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(400, gin.H{"error": "invalid candidate id"})
			return
		}

		rows, err := pool.Query(c.Request.Context(),
			`SELECT id, candidate_id, score_leadership, score_motivation, score_growth, score_vision,
				score_communication, final_score, category, COALESCE(ai_generated_risk, 'low'),
				COALESCE(ai_generated_score, 0), summary, key_strengths, red_flags,
				model_used, analyzed_at, COALESCE(duration_ms, 0)
			 FROM analysis_history
			 WHERE candidate_id = $1
			 ORDER BY analyzed_at DESC`, candidateID)
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to fetch analysis history"})
			return
		}
		defer rows.Close()

		entries := []AnalysisHistoryEntry{}
		for rows.Next() {
			var e AnalysisHistoryEntry
			if err := rows.Scan(&e.ID, &e.CandidateID, &e.ScoreLeadership, &e.ScoreMotivation,
				&e.ScoreGrowth, &e.ScoreVision, &e.ScoreCommunication, &e.FinalScore,
				&e.Category, &e.AIGeneratedRisk, &e.AIGeneratedScore,
				&e.Summary, &e.KeyStrengths, &e.RedFlags,
				&e.ModelUsed, &e.AnalyzedAt, &e.DurationMs); err != nil {
				continue
			}
			entries = append(entries, e)
		}

		c.JSON(http.StatusOK, gin.H{"history": entries})
	}
}
