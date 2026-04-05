package handlers

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/assylkhan/invisionu-backend/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// assignTaskRoundRobin — наименее загруженный менеджер, вызов через go
func assignTaskRoundRobin(pool *pgxpool.Pool, candidateID int) {
	ctx := context.Background()

	var managerID int
	err := pool.QueryRow(ctx,
		`SELECT u.id
		 FROM users u
		 WHERE u.role IN ('manager', 'admin', 'superadmin')
		 ORDER BY (
			SELECT COUNT(*) FROM review_tasks rt
			WHERE rt.assigned_to = u.id AND rt.status != 'completed'
		 ) ASC, u.id ASC
		 LIMIT 1`,
	).Scan(&managerID)
	if err != nil {
		log.Printf("assignTaskRoundRobin: no managers found for candidate %d: %v", candidateID, err)
		return
	}

	_, err = pool.Exec(ctx,
		`INSERT INTO review_tasks (candidate_id, assigned_to, status)
		 VALUES ($1, $2, 'pending')
		 ON CONFLICT (candidate_id) DO NOTHING`,
		candidateID, managerID)
	if err != nil {
		log.Printf("assignTaskRoundRobin: failed to insert task for candidate %d: %v", candidateID, err)
	}
}

type ReviewTask struct {
	ID           int        `json:"id"`
	CandidateID  int        `json:"candidate_id"`
	CandidateName string    `json:"candidate_name"`
	CandidateEmail string   `json:"candidate_email"`
	AssignedTo   *int       `json:"assigned_to"`
	AssigneeName *string    `json:"assignee_name"`
	AssigneeEmail *string   `json:"assignee_email"`
	Status       string     `json:"status"`
	ReviewComplexity *float64 `json:"review_complexity"`
	CreatedAt    time.Time  `json:"created_at"`
	CompletedAt  *time.Time `json:"completed_at"`
}

// GetTasks — список задач на проверку
func GetTasks(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		filter := c.DefaultQuery("filter", "all") // all/mine/unassigned
		userID := middleware.GetUserID(c)

		query := `SELECT rt.id, rt.candidate_id, c.full_name, c.email, rt.assigned_to,
					u.full_name, u.email, rt.status, c.review_complexity, rt.created_at, rt.completed_at
				  FROM review_tasks rt
				  JOIN candidates c ON c.id = rt.candidate_id
				  LEFT JOIN users u ON u.id = rt.assigned_to`

		var args []interface{}
		switch filter {
		case "mine":
			query += " WHERE rt.assigned_to = $1"
			args = append(args, userID)
		case "unassigned":
			query += " WHERE rt.assigned_to IS NULL"
		}
		query += " ORDER BY rt.created_at DESC"

		rows, err := pool.Query(c.Request.Context(), query, args...)
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to fetch tasks"})
			return
		}
		defer rows.Close()

		tasks := []ReviewTask{}
		for rows.Next() {
			var t ReviewTask
			if err := rows.Scan(&t.ID, &t.CandidateID, &t.CandidateName, &t.CandidateEmail,
				&t.AssignedTo, &t.AssigneeName, &t.AssigneeEmail, &t.Status,
				&t.ReviewComplexity, &t.CreatedAt, &t.CompletedAt); err != nil {
				continue
			}
			tasks = append(tasks, t)
		}

		c.JSON(http.StatusOK, gin.H{"tasks": tasks})
	}
}

// AssignNewTasks — round-robin назначение новых кандидатов
func AssignNewTasks(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

		managerRows, err := pool.Query(ctx,
			`SELECT u.id, u.full_name,
				COALESCE((SELECT COUNT(*) FROM review_tasks rt WHERE rt.assigned_to = u.id AND rt.status != 'completed'), 0) AS active_tasks
			 FROM users u
			 WHERE u.role IN ('manager', 'admin', 'superadmin')
			 ORDER BY active_tasks ASC, u.id ASC`)
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to fetch managers"})
			return
		}
		defer managerRows.Close()

		type manager struct {
			ID   int
			Name string
		}
		var managers []manager
		for managerRows.Next() {
			var m manager
			var active int
			if err := managerRows.Scan(&m.ID, &m.Name, &active); err == nil {
				managers = append(managers, m)
			}
		}

		if len(managers) == 0 {
			c.JSON(400, gin.H{"error": "no managers available"})
			return
		}

		// Кандидаты без задач
		unassignedRows, err := pool.Query(ctx,
			`SELECT c.id FROM candidates c
			 WHERE NOT EXISTS (SELECT 1 FROM review_tasks rt WHERE rt.candidate_id = c.id)
			 ORDER BY c.review_complexity ASC NULLS LAST, c.created_at ASC`)
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to fetch unassigned candidates"})
			return
		}
		defer unassignedRows.Close()

		var candidateIDs []int
		for unassignedRows.Next() {
			var id int
			if err := unassignedRows.Scan(&id); err == nil {
				candidateIDs = append(candidateIDs, id)
			}
		}

		if len(candidateIDs) == 0 {
			c.JSON(http.StatusOK, gin.H{"message": "no new candidates to assign", "assigned": 0})
			return
		}

		assigned := 0
		for i, candID := range candidateIDs {
			mgr := managers[i%len(managers)]
			_, err := pool.Exec(ctx,
				`INSERT INTO review_tasks (candidate_id, assigned_to, status) VALUES ($1, $2, 'pending')
				 ON CONFLICT (candidate_id) DO NOTHING`,
				candID, mgr.ID)
			if err == nil {
				assigned++
			}
		}

		c.JSON(http.StatusOK, gin.H{"message": "tasks assigned", "assigned": assigned})
	}
}

// UpdateTaskStatus — смена статуса задачи
func UpdateTaskStatus(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		taskID, err := strconv.Atoi(c.Param("taskId"))
		if err != nil {
			c.JSON(400, gin.H{"error": "invalid task id"})
			return
		}

		var req struct {
			Status string `json:"status" binding:"required"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": "status required"})
			return
		}

		if req.Status != "pending" && req.Status != "in_progress" && req.Status != "completed" {
			c.JSON(400, gin.H{"error": "invalid status"})
			return
		}

		var completedAt interface{}
		if req.Status == "completed" {
			now := time.Now()
			completedAt = &now
		}

		_, err = pool.Exec(c.Request.Context(),
			`UPDATE review_tasks SET status = $1, completed_at = $2 WHERE id = $3`,
			req.Status, completedAt, taskID)
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to update task"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "task updated"})
	}
}
