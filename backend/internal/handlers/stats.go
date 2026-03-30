package handlers

import (
	"fmt"
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

		// Score statistics for distribution graphs
		var mean, median float64
		pool.QueryRow(ctx, `SELECT COALESCE(AVG(final_score), 0) FROM analyses`).Scan(&mean)
		pool.QueryRow(ctx, `SELECT COALESCE(PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY final_score), 0) FROM analyses`).Scan(&median)
		stats.ScoreMean = mean
		stats.ScoreMedian = median

		// Per-dimension means
		var meanL, meanM, meanG, meanV, meanC float64
		pool.QueryRow(ctx, `SELECT COALESCE(AVG(score_leadership),0), COALESCE(AVG(score_motivation),0),
			COALESCE(AVG(score_growth),0), COALESCE(AVG(score_vision),0), COALESCE(AVG(score_communication),0) FROM analyses`).Scan(
			&meanL, &meanM, &meanG, &meanV, &meanC)
		stats.DimensionMeans = map[string]float64{
			"leadership":    meanL,
			"motivation":    meanM,
			"growth":        meanG,
			"vision":        meanV,
			"communication": meanC,
		}

		// Per-dimension score distributions (buckets of 10)
		dimensions := []string{"score_leadership", "score_motivation", "score_growth", "score_vision", "score_communication"}
		dimDistributions := map[string][]models.ScoreBucket{}
		for _, dim := range dimensions {
			var buckets []models.ScoreBucket
			for lo := 0; lo < 100; lo += 10 {
				hi := lo + 9
				if lo == 90 {
					hi = 100
				}
				var cnt int
				pool.QueryRow(ctx, fmt.Sprintf(`SELECT COUNT(*) FROM analyses WHERE %s >= $1 AND %s <= $2`, dim, dim), lo, hi).Scan(&cnt)
				buckets = append(buckets, models.ScoreBucket{
					Range: fmt.Sprintf("%d-%d", lo, hi),
					Count: cnt,
				})
			}
			// strip "score_" prefix
			key := dim[6:]
			dimDistributions[key] = buckets
		}
		stats.DimensionDistributions = dimDistributions

		c.JSON(http.StatusOK, stats)
	}
}
