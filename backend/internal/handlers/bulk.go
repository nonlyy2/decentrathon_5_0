package handlers

import (
	"fmt"
	"net/http"

	"github.com/assylkhan/invisionu-backend/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BulkDecisionRequest struct {
	CandidateIDs []int  `json:"candidate_ids" binding:"required,min=1"`
	Decision     string `json:"decision" binding:"required,oneof=shortlist reject waitlist review pending upvote downvote"`
	Notes        string `json:"notes"`
}

// decisionsWithRecord — типы решений, записываемых в committee_decisions.
// "pending" — только сброс статуса, запись не создаётся.
var decisionsWithRecord = map[string]bool{
	"shortlist": true,
	"reject":    true,
	"waitlist":  true,
	"review":    true,
	"upvote":    true,
	"downvote":  true,
}

func BulkDecision(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if middleware.IsRole(c, "auditor") {
			c.JSON(403, gin.H{"error": "auditors cannot make bulk decisions"})
			return
		}
		var req BulkDecisionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			HandleValidationError(c, err)
			return
		}

		userID, _ := c.Get("user_id")
		ctx := c.Request.Context()

		statusMap := map[string]string{
			"shortlist": "shortlisted",
			"reject":    "rejected",
			"waitlist":  "waitlisted",
			"review":    "analyzed",
			"pending":   "pending",
			"upvote":    "",
			"downvote":  "",
		}
		newStatus := statusMap[req.Decision]

		success := 0
		for _, id := range req.CandidateIDs {
			// Запись в committee_decisions только для нужных типов решений.
			if decisionsWithRecord[req.Decision] {
				_, err := pool.Exec(ctx,
					`INSERT INTO committee_decisions (candidate_id, decision, notes, decided_by)
					 VALUES ($1, $2, $3, $4)
					 ON CONFLICT (candidate_id, decided_by) DO UPDATE SET decision=$2, notes=$3, decided_at=NOW()`,
					id, req.Decision, req.Notes, userID)
				if err != nil {
					continue
				}
			}

			// upvote/downvote не меняют статус кандидата.
			if newStatus == "" {
				success++
				continue
			}

			_, err := pool.Exec(ctx, `UPDATE candidates SET status = $1 WHERE id = $2`, newStatus, id)
			if err != nil {
				continue
			}
			success++
		}

		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("%d candidates updated to %s", success, req.Decision),
			"updated": success,
		})
	}
}

// AutoAcceptRequest — тело запроса для авто-шортлиста топ-N кандидатов.
type AutoAcceptRequest struct {
	Count    int      `json:"count" binding:"required,min=1"`
	Major    string   `json:"major"`
	MinScore *float64 `json:"min_score"`
	MaxScore *float64 `json:"max_score"`
	MinAge   *int     `json:"min_age"`
	MaxAge   *int     `json:"max_age"`
}

// AutoAcceptTopN — шортлистит топ-N проанализированных кандидатов с учётом фильтров.
func AutoAcceptTopN(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if middleware.IsRole(c, "auditor") {
			c.JSON(403, gin.H{"error": "auditors cannot auto-accept candidates"})
			return
		}
		var req AutoAcceptRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			HandleValidationError(c, err)
			return
		}

		userID, _ := c.Get("user_id")
		ctx := c.Request.Context()

		where := "WHERE c.status = 'analyzed'"
		args := []interface{}{}
		argIdx := 1

		if req.Major != "" {
			where += fmt.Sprintf(" AND c.major = $%d", argIdx)
			args = append(args, req.Major)
			argIdx++
		}
		if req.MinScore != nil {
			where += fmt.Sprintf(" AND a.final_score >= $%d", argIdx)
			args = append(args, *req.MinScore)
			argIdx++
		}
		if req.MaxScore != nil {
			where += fmt.Sprintf(" AND a.final_score <= $%d", argIdx)
			args = append(args, *req.MaxScore)
			argIdx++
		}
		if req.MinAge != nil {
			where += fmt.Sprintf(" AND c.age >= $%d", argIdx)
			args = append(args, *req.MinAge)
			argIdx++
		}
		if req.MaxAge != nil {
			where += fmt.Sprintf(" AND c.age <= $%d", argIdx)
			args = append(args, *req.MaxAge)
			argIdx++
		}

		args = append(args, req.Count)
		query := fmt.Sprintf(`
			SELECT c.id FROM candidates c
			JOIN analyses a ON a.candidate_id = c.id
			%s
			ORDER BY a.final_score DESC
			LIMIT $%d`, where, argIdx)

		rows, err := pool.Query(ctx, query, args...)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch candidates"})
			return
		}
		defer rows.Close()

		var ids []int
		for rows.Next() {
			var id int
			if err := rows.Scan(&id); err == nil {
				ids = append(ids, id)
			}
		}

		if len(ids) == 0 {
			c.JSON(http.StatusOK, gin.H{"message": "No analyzed candidates to accept", "accepted": 0})
			return
		}

		accepted := 0
		for _, id := range ids {
			pool.Exec(ctx,
				`INSERT INTO committee_decisions (candidate_id, decision, notes, decided_by) VALUES ($1, 'shortlist', 'Auto-accepted (top N)', $2) ON CONFLICT (candidate_id, decided_by) DO UPDATE SET decision='shortlist', decided_at=NOW()`,
				id, userID)
			_, err := pool.Exec(ctx, `UPDATE candidates SET status = 'shortlisted' WHERE id = $1`, id)
			if err == nil {
				accepted++
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"message":  fmt.Sprintf("Auto-accepted top %d candidates", accepted),
			"accepted": accepted,
		})
	}
}
