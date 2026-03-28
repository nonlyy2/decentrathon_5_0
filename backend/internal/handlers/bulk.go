package handlers

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type BulkDecisionRequest struct {
	CandidateIDs []int  `json:"candidate_ids" binding:"required,min=1"`
	Decision     string `json:"decision" binding:"required,oneof=shortlist reject waitlist review"`
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
