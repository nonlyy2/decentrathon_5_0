package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/assylkhan/invisionu-backend/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Comment struct {
	ID          int       `json:"id"`
	CandidateID int       `json:"candidate_id"`
	UserID      int       `json:"user_id"`
	UserEmail   string    `json:"user_email"`
	Content     string    `json:"content"`
	CreatedAt   time.Time `json:"created_at"`
}

func GetComments(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		candidateID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(400, gin.H{"error": "invalid candidate id"})
			return
		}

		rows, err := pool.Query(c.Request.Context(),
			`SELECT cm.id, cm.candidate_id, cm.user_id, COALESCE(u.email, 'unknown'), cm.content, cm.created_at
			 FROM comments cm LEFT JOIN users u ON cm.user_id = u.id
			 WHERE cm.candidate_id = $1 ORDER BY cm.created_at DESC`, candidateID)
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to fetch comments"})
			return
		}
		defer rows.Close()

		comments := []Comment{}
		for rows.Next() {
			var cm Comment
			if err := rows.Scan(&cm.ID, &cm.CandidateID, &cm.UserID, &cm.UserEmail, &cm.Content, &cm.CreatedAt); err == nil {
				comments = append(comments, cm)
			}
		}

		c.JSON(http.StatusOK, comments)
	}
}

func AddComment(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if middleware.IsRole(c, "auditor") {
			c.JSON(403, gin.H{"error": "auditors cannot add comments on candidates"})
			return
		}

		candidateID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(400, gin.H{"error": "invalid candidate id"})
			return
		}

		var req struct {
			Content string `json:"content" binding:"required,min=1"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			HandleValidationError(c, err)
			return
		}

		userID, _ := c.Get("user_id")

		var cm Comment
		err = pool.QueryRow(c.Request.Context(),
			`INSERT INTO comments (candidate_id, user_id, content) VALUES ($1, $2, $3)
			 RETURNING id, candidate_id, user_id, content, created_at`,
			candidateID, userID, req.Content,
		).Scan(&cm.ID, &cm.CandidateID, &cm.UserID, &cm.Content, &cm.CreatedAt)

		if err != nil {
			c.JSON(500, gin.H{"error": "failed to add comment"})
			return
		}

		pool.QueryRow(c.Request.Context(), `SELECT email FROM users WHERE id = $1`, userID).Scan(&cm.UserEmail)

		c.JSON(http.StatusCreated, cm)
	}
}

func DeleteComment(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		commentID, err := strconv.Atoi(c.Param("commentId"))
		if err != nil {
			c.JSON(400, gin.H{"error": "invalid comment id"})
			return
		}

		userID, _ := c.Get("user_id")
		userRole, _ := c.Get("user_role")

		// Удаление: свои или admin — любые
		query := `DELETE FROM comments WHERE id = $1 AND user_id = $2`
		args := []interface{}{commentID, userID}
		if userRole == "admin" {
			query = `DELETE FROM comments WHERE id = $1`
			args = []interface{}{commentID}
		}

		tag, err := pool.Exec(c.Request.Context(), query, args...)
		if err != nil || tag.RowsAffected() == 0 {
			c.JSON(404, gin.H{"error": "comment not found"})
			return
		}

		c.JSON(200, gin.H{"message": "comment deleted"})
	}
}
