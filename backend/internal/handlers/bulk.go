package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BulkDecisionRequest struct {
	CandidateIDs []int  `json:"candidate_ids" binding:"required,min=1"`
	Decision     string `json:"decision" binding:"required,oneof=shortlist reject waitlist review pending"`
	Notes        string `json:"notes"`
}

func BulkDecision(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
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
		}
		newStatus := statusMap[req.Decision]

		success := 0
		for _, id := range req.CandidateIDs {
			// Insert decision
			_, err := pool.Exec(ctx,
				`INSERT INTO committee_decisions (candidate_id, decision, notes, decided_by) VALUES ($1, $2, $3, $4)`,
				id, req.Decision, req.Notes, userID)
			if err != nil {
				continue
			}
			// Update status
			_, err = pool.Exec(ctx, `UPDATE candidates SET status = $1 WHERE id = $2`, newStatus, id)
			if err != nil {
				continue
			}
			success++
		}

		c.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("%d candidates updated to %s", success, newStatus),
			"updated": success,
		})
	}
}

// AutoAcceptRequest defines the request body for auto-accepting top N candidates.
type AutoAcceptRequest struct {
	Count int `json:"count" binding:"required,min=1"`
}

// AutoAcceptTopN automatically shortlists the top N analyzed candidates by score.
func AutoAcceptTopN(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req AutoAcceptRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			HandleValidationError(c, err)
			return
		}

		userID, _ := c.Get("user_id")
		ctx := c.Request.Context()

		// Get top N analyzed candidates ordered by final_score desc
		rows, err := pool.Query(ctx, `
			SELECT c.id FROM candidates c
			JOIN analyses a ON a.candidate_id = c.id
			WHERE c.status = 'analyzed'
			ORDER BY a.final_score DESC
			LIMIT $1`, req.Count)
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
				`INSERT INTO committee_decisions (candidate_id, decision, notes, decided_by) VALUES ($1, 'shortlist', 'Auto-accepted (top N)', $2)`,
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
