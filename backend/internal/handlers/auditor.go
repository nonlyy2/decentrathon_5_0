package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ManagerStats — данные о производительности менеджера.
type ManagerStats struct {
	UserID          int     `json:"user_id"`
	Email           string  `json:"email"`
	FullName        *string `json:"full_name"`
	Role            string  `json:"role"`
	TotalDecisions  int     `json:"total_decisions"`
	Upvotes         int     `json:"upvotes"`
	Downvotes       int     `json:"downvotes"`
	Shortlists      int     `json:"shortlists"`
	Rejects         int     `json:"rejects"`
	Waitlists       int     `json:"waitlists"`
	Reviews         int     `json:"reviews"`
	SuccessfulCases int     `json:"successful_cases"` // решения, где кандидат попал в шортлист
	EfficiencyScore float64 `json:"efficiency_score"` // успешных / всего * 100
	LastActiveAt    *string `json:"last_active_at"`
}

// GetManagerPerformance — аналитика активности менеджеров (auditor+).
func GetManagerPerformance(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()

			rows, err := pool.Query(ctx, `
			SELECT
				u.id,
				u.email,
				u.full_name,
				u.role,
				COUNT(cd.id) AS total_decisions,
				COUNT(CASE WHEN cd.decision = 'upvote' THEN 1 END) AS upvotes,
				COUNT(CASE WHEN cd.decision = 'downvote' THEN 1 END) AS downvotes,
				COUNT(CASE WHEN cd.decision = 'shortlist' THEN 1 END) AS shortlists,
				COUNT(CASE WHEN cd.decision = 'reject' THEN 1 END) AS rejects,
				COUNT(CASE WHEN cd.decision = 'waitlist' THEN 1 END) AS waitlists,
				COUNT(CASE WHEN cd.decision = 'review' THEN 1 END) AS reviews,
				COUNT(CASE WHEN c.status = 'shortlisted' THEN 1 END) AS successful_cases,
				MAX(cd.decided_at::text) AS last_active_at
			FROM users u
			LEFT JOIN committee_decisions cd ON cd.decided_by = u.id
			LEFT JOIN candidates c ON c.id = cd.candidate_id
			WHERE u.role IN ('manager', 'committee', 'tech-admin', 'superadmin', 'admin')
			GROUP BY u.id, u.email, u.full_name, u.role
			ORDER BY total_decisions DESC
		`)
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to fetch manager stats"})
			return
		}
		defer rows.Close()

		var stats []ManagerStats
		for rows.Next() {
			var s ManagerStats
			if err := rows.Scan(
				&s.UserID, &s.Email, &s.FullName, &s.Role,
				&s.TotalDecisions, &s.Upvotes, &s.Downvotes,
				&s.Shortlists, &s.Rejects, &s.Waitlists, &s.Reviews,
				&s.SuccessfulCases, &s.LastActiveAt,
			); err == nil {
				if s.TotalDecisions > 0 {
					s.EfficiencyScore = float64(s.SuccessfulCases) / float64(s.TotalDecisions) * 100
				}
				stats = append(stats, s)
			}
		}

		var totalCandidates, shortlisted, rejected, pending int
		pool.QueryRow(ctx, `SELECT COUNT(*) FROM candidates`).Scan(&totalCandidates)
		pool.QueryRow(ctx, `SELECT COUNT(*) FROM candidates WHERE status='shortlisted'`).Scan(&shortlisted)
		pool.QueryRow(ctx, `SELECT COUNT(*) FROM candidates WHERE status='rejected'`).Scan(&rejected)
		pool.QueryRow(ctx, `SELECT COUNT(*) FROM candidates WHERE status='pending'`).Scan(&pending)

		// Тренд решений за 30 дней
		trendRows, err := pool.Query(ctx, `
			SELECT DATE(decided_at) as day, COUNT(*) as count
			FROM committee_decisions
			WHERE decided_at >= NOW() - INTERVAL '30 days'
			GROUP BY DATE(decided_at)
			ORDER BY day ASC
		`)
		type DayCount struct {
			Day   string `json:"day"`
			Count int    `json:"count"`
		}
		var trend []DayCount
		if err == nil {
			defer trendRows.Close()
			for trendRows.Next() {
				var dc DayCount
				if trendRows.Scan(&dc.Day, &dc.Count) == nil {
					trend = append(trend, dc)
				}
			}
		}

		c.JSON(http.StatusOK, gin.H{
			"managers": stats,
			"summary": gin.H{
				"total_candidates": totalCandidates,
				"shortlisted":      shortlisted,
				"rejected":         rejected,
				"pending":          pending,
			},
			"decision_trend": trend,
		})
	}
}
