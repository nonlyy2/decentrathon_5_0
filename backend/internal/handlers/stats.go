package handlers

import (
	"net/http"

	"github.com/assylkhan/invisionu-backend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func GetDashboardStats(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		stats := models.DashboardStats{
			CategoryCounts: make(map[string]int),
		}

		// Count by status
		rows, err := pool.Query(ctx, `SELECT status, COUNT(*) FROM candidates GROUP BY status`)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var status string
				var count int
				if rows.Scan(&status, &count) == nil {
					stats.TotalCandidates += count
					switch status {
					case "pending":
						stats.Pending = count
					case "analyzed":
						stats.Analyzed += count
					case "shortlisted":
						stats.Shortlisted = count
						stats.Analyzed += count
					case "rejected":
						stats.Rejected = count
						stats.Analyzed += count
					case "waitlisted":
						stats.Waitlisted = count
						stats.Analyzed += count
					}
				}
			}
		}

		// Average score
		pool.QueryRow(ctx, `SELECT COALESCE(AVG(final_score), 0) FROM analyses`).Scan(&stats.AvgScore)

		// Score distribution
		buckets := []struct {
			label string
			min   float64
			max   float64
		}{
			{"0-49", 0, 49.99},
			{"50-64", 50, 64.99},
			{"65-79", 65, 79.99},
			{"80-100", 80, 100},
		}
		for _, b := range buckets {
			var count int
			pool.QueryRow(ctx,
				`SELECT COUNT(*) FROM analyses WHERE final_score >= $1 AND final_score <= $2`,
				b.min, b.max).Scan(&count)
			stats.ScoreDistribution = append(stats.ScoreDistribution, models.ScoreBucket{
				Range: b.label,
				Count: count,
			})
		}

		// Category counts
		catRows, err := pool.Query(ctx, `SELECT category, COUNT(*) FROM analyses GROUP BY category`)
		if err == nil {
			defer catRows.Close()
			for catRows.Next() {
				var cat string
				var count int
				if catRows.Scan(&cat, &count) == nil {
					stats.CategoryCounts[cat] = count
				}
			}
		}

		c.JSON(http.StatusOK, stats)
	}
}
