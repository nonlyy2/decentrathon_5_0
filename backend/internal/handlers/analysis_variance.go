package handlers

import (
	"net/http"

	"github.com/assylkhan/invisionu-backend/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CandidateScoreVariance struct {
	CandidateID   int     `json:"candidate_id"`
	FullName      string  `json:"full_name"`
	AnalysisCount int     `json:"analysis_count"`
	MeanScore     float64 `json:"mean_score"`
	StdDev        float64 `json:"std_dev"`
	MinScore      float64 `json:"min_score"`
	MaxScore      float64 `json:"max_score"`
	ScoreRange    float64 `json:"score_range"`
}

type VarianceSummary struct {
	Candidates       []CandidateScoreVariance `json:"candidates"`
	OverallMeanStdev float64                  `json:"overall_mean_stdev"`
	HighVarianceCount int                     `json:"high_variance_count"`
	TotalMultiAnalyzed int                    `json:"total_multi_analyzed"`
}

// GetAnalysisVariance returns stdev/variance data for candidates with multiple analyses.
// Auditor+ only.
func GetAnalysisVariance(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !middleware.HasLevel(c, "auditor") {
			c.JSON(403, gin.H{"error": "forbidden"})
			return
		}

		ctx := c.Request.Context()

		// For each candidate, combine current analysis + history to compute stats
		rows, err := pool.Query(ctx, `
			WITH all_scores AS (
				SELECT candidate_id, final_score FROM analyses
				UNION ALL
				SELECT candidate_id, final_score FROM analysis_history
			),
			candidate_stats AS (
				SELECT
					s.candidate_id,
					c.full_name,
					COUNT(*) as analysis_count,
					AVG(s.final_score) as mean_score,
					COALESCE(STDDEV_POP(s.final_score), 0) as std_dev,
					MIN(s.final_score) as min_score,
					MAX(s.final_score) as max_score
				FROM all_scores s
				JOIN candidates c ON c.id = s.candidate_id
				GROUP BY s.candidate_id, c.full_name
				HAVING COUNT(*) >= 2
			)
			SELECT candidate_id, full_name, analysis_count, mean_score, std_dev, min_score, max_score
			FROM candidate_stats
			ORDER BY std_dev DESC
		`)
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to query analysis variance"})
			return
		}
		defer rows.Close()

		var candidates []CandidateScoreVariance
		var totalStdev float64
		highVariance := 0

		for rows.Next() {
			var cv CandidateScoreVariance
			if err := rows.Scan(&cv.CandidateID, &cv.FullName, &cv.AnalysisCount,
				&cv.MeanScore, &cv.StdDev, &cv.MinScore, &cv.MaxScore); err != nil {
				continue
			}
			cv.ScoreRange = cv.MaxScore - cv.MinScore
			totalStdev += cv.StdDev
			if cv.StdDev > 5.0 {
				highVariance++
			}
			candidates = append(candidates, cv)
		}

		meanStdev := 0.0
		if len(candidates) > 0 {
			meanStdev = totalStdev / float64(len(candidates))
		}

		c.JSON(http.StatusOK, VarianceSummary{
			Candidates:         candidates,
			OverallMeanStdev:   meanStdev,
			HighVarianceCount:  highVariance,
			TotalMultiAnalyzed: len(candidates),
		})
	}
}
