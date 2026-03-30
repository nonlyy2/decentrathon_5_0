package handlers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/assylkhan/invisionu-backend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

func CreateCandidate(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.CreateCandidateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			HandleValidationError(c, err)
			return
		}

		var candidate models.Candidate
		err := pool.QueryRow(c.Request.Context(),
			`INSERT INTO candidates (full_name, email, phone, telegram, age, city, school, graduation_year, achievements, extracurriculars, essay, motivation_statement, disability)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
			 RETURNING id, full_name, email, phone, telegram, age, city, school, graduation_year, achievements, extracurriculars, essay, motivation_statement, disability, created_at, status`,
			req.FullName, req.Email, req.Phone, req.Telegram, req.Age, req.City, req.School, req.GraduationYear,
			req.Achievements, req.Extracurriculars, req.Essay, req.MotivationStatement, req.Disability,
		).Scan(&candidate.ID, &candidate.FullName, &candidate.Email, &candidate.Phone, &candidate.Telegram, &candidate.Age, &candidate.City,
			&candidate.School, &candidate.GraduationYear, &candidate.Achievements, &candidate.Extracurriculars,
			&candidate.Essay, &candidate.MotivationStatement, &candidate.Disability, &candidate.CreatedAt, &candidate.Status)

		if err != nil {
			if isDuplicateKey(err) {
				c.JSON(409, gin.H{"error": "email already registered"})
				return
			}
			c.JSON(500, gin.H{"error": "failed to create candidate"})
			return
		}

		c.JSON(201, candidate)
	}
}

func SubmitApplication(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.CreateCandidateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			HandleValidationError(c, err)
			return
		}

		if len(req.Essay) < 50 {
			c.JSON(400, gin.H{"error": "essay must be at least 50 characters"})
			return
		}

		var id int
		err := pool.QueryRow(c.Request.Context(),
			`INSERT INTO candidates (full_name, email, phone, telegram, age, city, school, graduation_year, achievements, extracurriculars, essay, motivation_statement, disability, status)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,'pending')
			 RETURNING id`,
			req.FullName, req.Email, req.Phone, req.Telegram, req.Age, req.City, req.School, req.GraduationYear,
			req.Achievements, req.Extracurriculars, req.Essay, req.MotivationStatement, req.Disability,
		).Scan(&id)

		if err != nil {
			if isDuplicateKey(err) {
				c.JSON(409, gin.H{"error": "email already registered"})
				return
			}
			c.JSON(500, gin.H{"error": "failed to submit application"})
			return
		}

		c.JSON(201, gin.H{"message": "Application submitted successfully", "id": id})
	}
}

func ListCandidates(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		status := c.Query("status")
		search := c.Query("search")
		sortBy := c.DefaultQuery("sort_by", "created_at")
		order := c.DefaultQuery("order", "desc")
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

		if limit > 100 {
			limit = 100
		}

		// Whitelist sort columns
		allowedSorts := map[string]string{
			"created_at":  "c.created_at",
			"full_name":   "c.full_name",
			"final_score": "a.final_score",
			"analyzed_at": "a.analyzed_at",
		}
		sortCol, ok := allowedSorts[sortBy]
		if !ok {
			sortCol = "c.created_at"
		}
		if order != "asc" {
			order = "desc"
		}

		// Build query
		where := "WHERE 1=1"
		args := []interface{}{}
		argIdx := 1

		if status != "" && status != "all" {
			where += fmt.Sprintf(" AND c.status = $%d", argIdx)
			args = append(args, status)
			argIdx++
		}
		if search != "" {
			where += fmt.Sprintf(" AND (c.full_name ILIKE $%d OR c.email ILIKE $%d)", argIdx, argIdx)
			args = append(args, "%"+search+"%")
			argIdx++
		}

		// Count
		var total int
		countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM candidates c LEFT JOIN analyses a ON c.id = a.candidate_id %s`, where)
		pool.QueryRow(c.Request.Context(), countQuery, args...).Scan(&total)

		// Fetch
		query := fmt.Sprintf(
			`SELECT c.id, c.full_name, c.email, c.city, c.school, c.status, c.created_at, a.final_score, a.category, a.analyzed_at, a.model_used
			 FROM candidates c
			 LEFT JOIN analyses a ON c.id = a.candidate_id
			 %s ORDER BY %s %s NULLS LAST LIMIT $%d OFFSET $%d`,
			where, sortCol, order, argIdx, argIdx+1,
		)
		args = append(args, limit, offset)

		rows, err := pool.Query(c.Request.Context(), query, args...)
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to fetch candidates"})
			return
		}
		defer rows.Close()

		candidates := []models.CandidateListItem{}
		for rows.Next() {
			var item models.CandidateListItem
			if err := rows.Scan(&item.ID, &item.FullName, &item.Email, &item.City, &item.School,
				&item.Status, &item.CreatedAt, &item.FinalScore, &item.Category, &item.AnalyzedAt, &item.ModelUsed); err != nil {
				continue
			}
			candidates = append(candidates, item)
		}

		c.JSON(http.StatusOK, gin.H{
			"candidates": candidates,
			"total":      total,
			"limit":      limit,
			"offset":     offset,
		})
	}
}

func GetCandidate(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(400, gin.H{"error": "invalid candidate id"})
			return
		}

		// Get candidate
		var candidate models.Candidate
		err = pool.QueryRow(c.Request.Context(),
			`SELECT id, full_name, email, phone, telegram, age, city, school, graduation_year, achievements, extracurriculars, essay, motivation_statement, disability, created_at, status
			 FROM candidates WHERE id = $1`, id,
		).Scan(&candidate.ID, &candidate.FullName, &candidate.Email, &candidate.Phone, &candidate.Telegram, &candidate.Age, &candidate.City,
			&candidate.School, &candidate.GraduationYear, &candidate.Achievements, &candidate.Extracurriculars,
			&candidate.Essay, &candidate.MotivationStatement, &candidate.Disability, &candidate.CreatedAt, &candidate.Status)

		if err != nil {
			c.JSON(404, gin.H{"error": "candidate not found"})
			return
		}

		detail := models.CandidateDetail{
			Candidate: candidate,
			Decisions: []models.Decision{},
		}

		// Get analysis (optional)
		var analysis models.Analysis
		err = pool.QueryRow(c.Request.Context(),
			`SELECT id, candidate_id, score_leadership, score_motivation, score_growth, score_vision, score_communication,
			 final_score, category, ai_generated_risk, COALESCE(ai_generated_score, 0), incomplete_flag,
			 explanation_leadership, explanation_motivation, explanation_growth, explanation_vision, explanation_communication,
			 summary, key_strengths, red_flags, analyzed_at, model_used
			 FROM analyses WHERE candidate_id = $1`, id,
		).Scan(&analysis.ID, &analysis.CandidateID, &analysis.ScoreLeadership, &analysis.ScoreMotivation,
			&analysis.ScoreGrowth, &analysis.ScoreVision, &analysis.ScoreCommunication,
			&analysis.FinalScore, &analysis.Category, &analysis.AIGeneratedRisk, &analysis.AIGeneratedScore, &analysis.IncompleteFlag,
			&analysis.ExplanationLeadership, &analysis.ExplanationMotivation, &analysis.ExplanationGrowth,
			&analysis.ExplanationVision, &analysis.ExplanationCommunication,
			&analysis.Summary, &analysis.KeyStrengths, &analysis.RedFlags, &analysis.AnalyzedAt, &analysis.ModelUsed)

		if err == nil {
			detail.Analysis = &analysis
		}

		// Get decisions with user email
		rows, err := pool.Query(c.Request.Context(),
			`SELECT cd.id, cd.candidate_id, cd.decision, cd.notes, cd.decided_by, u.email, cd.decided_at
			 FROM committee_decisions cd
			 LEFT JOIN users u ON u.id = cd.decided_by
			 WHERE cd.candidate_id = $1 ORDER BY cd.decided_at DESC`, id)
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var d models.Decision
				if err := rows.Scan(&d.ID, &d.CandidateID, &d.Decision, &d.Notes, &d.DecidedBy, &d.DecidedByEmail, &d.DecidedAt); err == nil {
					detail.Decisions = append(detail.Decisions, d)
				}
			}
		}

		c.JSON(http.StatusOK, detail)
	}
}

func UpdateCandidateStatus(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(400, gin.H{"error": "invalid candidate id"})
			return
		}

		var req struct {
			Status string `json:"status" binding:"required,oneof=pending analyzed shortlisted rejected waitlisted"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			HandleValidationError(c, err)
			return
		}

		tag, err := pool.Exec(c.Request.Context(),
			`UPDATE candidates SET status = $1 WHERE id = $2`, req.Status, id)
		if err != nil || tag.RowsAffected() == 0 {
			c.JSON(404, gin.H{"error": "candidate not found"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "status updated", "status": req.Status})
	}
}
