package handlers

import (
	"net/http"
	"strconv"

	"github.com/assylkhan/invisionu-backend/internal/middleware"
	"github.com/assylkhan/invisionu-backend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

const reviewThreshold = 4

// upvoteThreshold: net score (upvotes - downvotes) needed for auto-shortlist
const upvoteThreshold = 3

func MakeDecision(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Auditors cannot make decisions — they are read-only observers
		if middleware.IsRole(c, "auditor") {
			c.JSON(403, gin.H{"error": "auditors cannot make decisions on candidates"})
			return
		}

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

		// Check if this user already voted on this candidate
		var existingVote int
		err = pool.QueryRow(c.Request.Context(),
			`SELECT id FROM committee_decisions WHERE candidate_id = $1 AND decided_by = $2`,
			candidateID, userID).Scan(&existingVote)
		if err == nil {
			// User already voted — update their vote
			_, err = pool.Exec(c.Request.Context(),
				`UPDATE committee_decisions SET decision = $1, notes = $2, decided_at = NOW()
				 WHERE candidate_id = $3 AND decided_by = $4`,
				req.Decision, req.Notes, candidateID, userID)
			if err != nil {
				c.JSON(500, gin.H{"error": "failed to update decision"})
				return
			}
		} else {
			// Insert new decision
			_, err = pool.Exec(c.Request.Context(),
				`INSERT INTO committee_decisions (candidate_id, decision, notes, decided_by)
				 VALUES ($1, $2, $3, $4)`,
				candidateID, req.Decision, req.Notes, userID)
			if err != nil {
				c.JSON(500, gin.H{"error": "failed to save decision"})
				return
			}
		}

		// Count total unique reviewers and tally votes
		type voteCount struct {
			Decision string
			Count    int
		}
		rows, err := pool.Query(c.Request.Context(),
			`SELECT decision, COUNT(*) FROM committee_decisions WHERE candidate_id = $1 GROUP BY decision`,
			candidateID)
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to tally votes"})
			return
		}
		defer rows.Close()

		var totalVotes int
		var votes []voteCount
		for rows.Next() {
			var v voteCount
			if rows.Scan(&v.Decision, &v.Count) == nil {
				votes = append(votes, v)
				totalVotes += v.Count
			}
		}

		// Determine if consensus reached
		statusMap := map[string]string{
			"shortlist": "shortlisted",
			"reject":    "rejected",
			"waitlist":  "waitlisted",
			"review":    "analyzed",
		}

		consensusReached := false
		var winningDecision string

		// Check for upvote/downvote net score
		var upvotes, downvotes int
		hasVoteSystem := false
		for _, v := range votes {
			if v.Decision == "upvote" {
				upvotes += v.Count
				hasVoteSystem = true
			} else if v.Decision == "downvote" {
				downvotes += v.Count
				hasVoteSystem = true
			}
		}

		if hasVoteSystem {
			netScore := upvotes - downvotes
			if netScore >= upvoteThreshold {
				winningDecision = "upvote"
				consensusReached = true
				pool.Exec(c.Request.Context(),
					`UPDATE candidates SET status = 'shortlisted' WHERE id = $1`, candidateID)
			} else if netScore <= -upvoteThreshold {
				winningDecision = "downvote"
				consensusReached = true
				pool.Exec(c.Request.Context(),
					`UPDATE candidates SET status = 'rejected' WHERE id = $1`, candidateID)
			}
		} else if totalVotes >= reviewThreshold {
			// Legacy: find majority decision among shortlist/reject/waitlist/review
			maxCount := 0
			for _, v := range votes {
				if v.Count > maxCount {
					maxCount = v.Count
					winningDecision = v.Decision
				}
			}
			consensusReached = true
			newStatus := statusMap[winningDecision]
			pool.Exec(c.Request.Context(),
				`UPDATE candidates SET status = $1 WHERE id = $2`, newStatus, candidateID)
		}

		// Get updated decision list for response
		decRows, _ := pool.Query(c.Request.Context(),
			`SELECT cd.id, cd.candidate_id, cd.decision, cd.notes, cd.decided_by, u.email, cd.decided_at
			 FROM committee_decisions cd
			 LEFT JOIN users u ON u.id = cd.decided_by
			 WHERE cd.candidate_id = $1 ORDER BY cd.decided_at DESC`, candidateID)
		var decisions []models.Decision
		if decRows != nil {
			defer decRows.Close()
			for decRows.Next() {
				var d models.Decision
				if decRows.Scan(&d.ID, &d.CandidateID, &d.Decision, &d.Notes, &d.DecidedBy, &d.DecidedByEmail, &d.DecidedAt) == nil {
					decisions = append(decisions, d)
				}
			}
		}

		c.JSON(http.StatusCreated, gin.H{
			"decisions":         decisions,
			"total_reviews":     totalVotes,
			"required_reviews":  reviewThreshold,
			"consensus_reached": consensusReached,
			"winning_decision":  winningDecision,
			"upvotes":           upvotes,
			"downvotes":         downvotes,
			"net_score":         upvotes - downvotes,
			"upvote_threshold":  upvoteThreshold,
		})
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
			`SELECT cd.id, cd.candidate_id, cd.decision, cd.notes, cd.decided_by, u.email, cd.decided_at
			 FROM committee_decisions cd
			 LEFT JOIN users u ON u.id = cd.decided_by
			 WHERE cd.candidate_id = $1 ORDER BY cd.decided_at DESC`, candidateID)
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to fetch decisions"})
			return
		}
		defer rows.Close()

		decisions := []models.Decision{}
		for rows.Next() {
			var d models.Decision
			if err := rows.Scan(&d.ID, &d.CandidateID, &d.Decision, &d.Notes, &d.DecidedBy, &d.DecidedByEmail, &d.DecidedAt); err == nil {
				decisions = append(decisions, d)
			}
		}

		// Count votes summary
		var totalVotes int
		type voteCount struct {
			Decision string `json:"decision"`
			Count    int    `json:"count"`
		}
		var voteSummary []voteCount
		voteRows, err := pool.Query(c.Request.Context(),
			`SELECT decision, COUNT(*) FROM committee_decisions WHERE candidate_id = $1 GROUP BY decision`, candidateID)
		if err == nil {
			defer voteRows.Close()
			for voteRows.Next() {
				var v voteCount
				if voteRows.Scan(&v.Decision, &v.Count) == nil {
					voteSummary = append(voteSummary, v)
					totalVotes += v.Count
				}
			}
		}

		var upvotes, downvotes int
		for _, v := range voteSummary {
			if v.Decision == "upvote" {
				upvotes = v.Count
			} else if v.Decision == "downvote" {
				downvotes = v.Count
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"decisions":        decisions,
			"vote_summary":     voteSummary,
			"total_reviews":    totalVotes,
			"required_reviews": reviewThreshold,
			"upvotes":          upvotes,
			"downvotes":        downvotes,
			"net_score":        upvotes - downvotes,
			"upvote_threshold": upvoteThreshold,
		})
	}
}
