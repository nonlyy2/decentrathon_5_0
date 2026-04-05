package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/assylkhan/invisionu-backend/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PrivateNote struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	CandidateID int       `json:"candidate_id"`
	Content     string    `json:"content"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func GetPrivateNote(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		candidateID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(400, gin.H{"error": "invalid candidate id"})
			return
		}

		var note PrivateNote
		err = pool.QueryRow(c.Request.Context(),
			`SELECT id, user_id, candidate_id, content, COALESCE(created_at, updated_at), updated_at FROM private_notes WHERE user_id = $1 AND candidate_id = $2`,
			userID, candidateID,
		).Scan(&note.ID, &note.UserID, &note.CandidateID, &note.Content, &note.CreatedAt, &note.UpdatedAt)

		if err != nil {
			c.JSON(http.StatusOK, PrivateNote{UserID: userID, CandidateID: candidateID, Content: ""})
			return
		}
		c.JSON(http.StatusOK, note)
	}
}

func SavePrivateNote(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		candidateID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(400, gin.H{"error": "invalid candidate id"})
			return
		}

		var req struct {
			Content string `json:"content"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "content required"})
			return
		}

		_, err = pool.Exec(c.Request.Context(),
			`INSERT INTO private_notes (user_id, candidate_id, content, created_at, updated_at)
			 VALUES ($1, $2, $3, NOW(), NOW())
			 ON CONFLICT (user_id, candidate_id) DO UPDATE SET content = $3, updated_at = NOW()`,
			userID, candidateID, req.Content,
		)
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to save note"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "note saved"})
	}
}

func DeletePrivateNote(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := middleware.GetUserID(c)
		candidateID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(400, gin.H{"error": "invalid candidate id"})
			return
		}

		tag, err := pool.Exec(c.Request.Context(),
			`DELETE FROM private_notes WHERE user_id = $1 AND candidate_id = $2`,
			userID, candidateID,
		)
		if err != nil || tag.RowsAffected() == 0 {
			c.JSON(404, gin.H{"error": "note not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "note deleted"})
	}
}
