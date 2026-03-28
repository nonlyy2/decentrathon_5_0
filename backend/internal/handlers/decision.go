package handlers

import (
	"net/http"
	"strconv"

	"github.com/assylkhan/invisionu-backend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func MakeDecision(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		candidateID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(400, gin.H{"error": "invalid candidate id"})
			return
		}

		var req models.CreateDecisionRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			HandleValidationError(c, err)
			return
		}

		userID, _ := c.Get("user_id")

		// Verify candidate exists
		var exists bool
		pool.QueryRow(c.Request.Context(),
			`SELECT EXISTS(SELECT 1 FROM candidates WHERE id = $1)`, candidateID).Scan(&exists)
		if !exists {
			c.JSON(404, gin.H{"error": "candidate not found"})
			return
		}

		// Insert decision
		var decision models.Decision
		err = pool.QueryRow(c.Request.Context(),
			`INSERT INTO committee_decisions (candidate_id, decision, notes, decided_by)
			 VALUES ($1, $2, $3, $4)
			 RETURNING id, candidate_id, decision, notes, decided_by, decided_at`,
			candidateID, req.Decision, req.Notes, userID,
		).Scan(&decision.ID, &decision.CandidateID, &decision.Decision, &decision.Notes, &decision.DecidedBy, &decision.DecidedAt)

		if err != nil {
			c.JSON(500, gin.H{"error": "failed to save decision"})
			return
		}

		// Update candidate status
		statusMap := map[string]string{
			"shortlist": "shortlisted",
			"reject":    "rejected",
			"waitlist":  "waitlisted",
			"review":    "analyzed",
		}
		newStatus := statusMap[req.Decision]
		pool.Exec(c.Request.Context(),
			`UPDATE candidates SET status = $1 WHERE id = $2`, newStatus, candidateID)

		c.JSON(http.StatusCreated, decision)
	}
}

func GetDecisions(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		candidateID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(400, gin.H{"error": "invalid candidate id"})
			return
		}

		rows, err := pool.Query(c.Request.Context(),
			`SELECT id, candidate_id, decision, notes, decided_by, decided_at
			 FROM committee_decisions WHERE candidate_id = $1 ORDER BY decided_at DESC`, candidateID)
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to fetch decisions"})
			return
		}
		defer rows.Close()

		decisions := []models.Decision{}
		for rows.Next() {
			var d models.Decision
			if err := rows.Scan(&d.ID, &d.CandidateID, &d.Decision, &d.Notes, &d.DecidedBy, &d.DecidedAt); err == nil {
				decisions = append(decisions, d)
			}
		}

		c.JSON(http.StatusOK, decisions)
	}
}
