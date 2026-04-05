package handlers

import (
	"fmt"
	"net/http"
	"strconv"

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

		rows, err := pool.Query(ctx, `SELECT status, COUNT(*) FROM candidates GROUP BY status`)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var status string
				var count int
				if rows.Scan(&status, &count) == nil {
					stats.TotalCandidates += count
					switch status {
					case "pending", "in_progress":
						stats.Pending += count
					case "analyzed", "initial_screening", "application_review":
						stats.Analyzed += count
					case "interview_stage", "committee_review":
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

		pool.QueryRow(ctx, `SELECT COALESCE(AVG(final_score), 0) FROM analyses`).Scan(&stats.AvgScore)

		// Распределение по баллам
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

		var mean, median float64
		pool.QueryRow(ctx, `SELECT COALESCE(AVG(final_score), 0) FROM analyses`).Scan(&mean)
		pool.QueryRow(ctx, `SELECT COALESCE(PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY final_score), 0) FROM analyses`).Scan(&median)
		stats.ScoreMean = mean
		stats.ScoreMedian = median

		// Средние по измерениям (5 критериев)
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

		// Бакеты по 10; top_n — фильтр топ-N
		topNParam := c.Query("top_n")
		topN := 0
		if topNParam != "" {
			if n, err := strconv.Atoi(topNParam); err == nil && n > 0 {
				topN = n
			}
		}

		dimTable := "analyses"
		if topN > 0 {
			dimTable = fmt.Sprintf("(SELECT * FROM analyses ORDER BY final_score DESC LIMIT %d)", topN)
		}

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
				pool.QueryRow(ctx, fmt.Sprintf(`SELECT COUNT(*) FROM %s AS a WHERE a.%s >= $1 AND a.%s <= $2`, dimTable, dim, dim), lo, hi).Scan(&cnt)
				buckets = append(buckets, models.ScoreBucket{
					Range: fmt.Sprintf("%d-%d", lo, hi),
					Count: cnt,
				})
			}
				key := dim[6:] // убираем "score_"
			dimDistributions[key] = buckets
		}
		stats.DimensionDistributions = dimDistributions

		// Пересчёт средних для top-N
		if topN > 0 {
			pool.QueryRow(ctx, fmt.Sprintf(`SELECT COALESCE(AVG(score_leadership),0), COALESCE(AVG(score_motivation),0),
				COALESCE(AVG(score_growth),0), COALESCE(AVG(score_vision),0), COALESCE(AVG(score_communication),0) FROM %s AS a`, dimTable)).Scan(
				&meanL, &meanM, &meanG, &meanV, &meanC)
			stats.DimensionMeans = map[string]float64{
				"leadership":    meanL,
				"motivation":    meanM,
				"growth":        meanG,
				"vision":        meanV,
				"communication": meanC,
			}
		}

		// IELTS
		ieltsRanges := []struct{ label string; lo, hi float64 }{
			{"5.0-5.5", 4.5, 5.99}, {"6.0-6.5", 6.0, 6.99}, {"7.0-7.5", 7.0, 7.99}, {"8.0-9.0", 8.0, 9.0},
		}
		for _, r := range ieltsRanges {
			var cnt int
			pool.QueryRow(ctx, `SELECT COUNT(*) FROM candidates WHERE exam_type='IELTS' AND ielts_score >= $1 AND ielts_score <= $2`, r.lo, r.hi).Scan(&cnt)
			stats.IELTSDistribution = append(stats.IELTSDistribution, models.ScoreBucket{Range: r.label, Count: cnt})
		}
		pool.QueryRow(ctx, `SELECT COUNT(*) FROM candidates WHERE exam_type='IELTS' AND ielts_score IS NOT NULL`).Scan(&stats.IELTSCount)

		// TOEFL
		toeflRanges := []struct{ label string; lo, hi int }{
			{"60-70", 60, 70}, {"71-80", 71, 80}, {"81-90", 81, 90}, {"91-100", 91, 100}, {"101-120", 101, 120},
		}
		for _, r := range toeflRanges {
			var cnt int
			pool.QueryRow(ctx, `SELECT COUNT(*) FROM candidates WHERE exam_type='TOEFL' AND toefl_score >= $1 AND toefl_score <= $2`, r.lo, r.hi).Scan(&cnt)
			stats.TOEFLDistribution = append(stats.TOEFLDistribution, models.ScoreBucket{Range: r.label, Count: cnt})
		}
		pool.QueryRow(ctx, `SELECT COUNT(*) FROM candidates WHERE exam_type='TOEFL' AND toefl_score IS NOT NULL`).Scan(&stats.TOEFLCount)

		// UNT
		untRanges := []struct{ label string; lo, hi int }{
			{"0-50", 0, 50}, {"51-80", 51, 80}, {"81-100", 81, 100}, {"101-120", 101, 120}, {"121-140", 121, 140},
		}
		for _, r := range untRanges {
			var cnt int
			pool.QueryRow(ctx, `SELECT COUNT(*) FROM candidates WHERE certificate_type='UNT' AND unt_score >= $1 AND unt_score <= $2`, r.lo, r.hi).Scan(&cnt)
			stats.UNTDistribution = append(stats.UNTDistribution, models.ScoreBucket{Range: r.label, Count: cnt})
		}
		pool.QueryRow(ctx, `SELECT COUNT(*) FROM candidates WHERE certificate_type='UNT' AND unt_score IS NOT NULL`).Scan(&stats.UNTCount)

		// NIS Grade
		nisGrades := []string{"A", "B", "C", "D"}
		for _, g := range nisGrades {
			var cnt int
			pool.QueryRow(ctx, `SELECT COUNT(*) FROM candidates WHERE certificate_type='NIS 12 Grade Certificate' AND nis_grade = $1`, g).Scan(&cnt)
			stats.NISDistribution = append(stats.NISDistribution, models.ScoreBucket{Range: g, Count: cnt})
		}
		pool.QueryRow(ctx, `SELECT COUNT(*) FROM candidates WHERE certificate_type='NIS 12 Grade Certificate' AND nis_grade IS NOT NULL`).Scan(&stats.NISCount)

		c.JSON(http.StatusOK, stats)
	}
}

func GetCityDistribution(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := pool.Query(c.Request.Context(),
			`SELECT city, COUNT(*) as count FROM candidates WHERE city IS NOT NULL AND city != '' GROUP BY city ORDER BY count DESC`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		type item struct {
			City  string `json:"city"`
			Count int    `json:"count"`
		}
		var result []item
		for rows.Next() {
			var r item
			if err := rows.Scan(&r.City, &r.Count); err == nil {
				result = append(result, r)
			}
		}
		c.JSON(http.StatusOK, result)
	}
}

func GetMajorDistribution(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := pool.Query(c.Request.Context(),
			`SELECT major, COUNT(*) as count FROM candidates WHERE major IS NOT NULL AND major != '' GROUP BY major ORDER BY count DESC`)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		type item struct {
			Major string `json:"major"`
			Count int    `json:"count"`
		}
		var result []item
		for rows.Next() {
			var r item
			if err := rows.Scan(&r.Major, &r.Count); err == nil {
				result = append(result, r)
			}
		}
		c.JSON(http.StatusOK, result)
	}
}
