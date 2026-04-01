package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

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
			`INSERT INTO candidates (full_name, email, phone, telegram, age, city, school, graduation_year, achievements, extracurriculars, essay, motivation_statement, disability, major)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
			 RETURNING id, full_name, email, phone, telegram, age, city, school, graduation_year, achievements, extracurriculars, essay, motivation_statement, disability, major, photo_url, photo_ai_flag, photo_ai_note, keywords, created_at, status`,
			req.FullName, req.Email, req.Phone, req.Telegram, req.Age, req.City, req.School, req.GraduationYear,
			req.Achievements, req.Extracurriculars, req.Essay, req.MotivationStatement, req.Disability, req.Major,
		).Scan(&candidate.ID, &candidate.FullName, &candidate.Email, &candidate.Phone, &candidate.Telegram, &candidate.Age, &candidate.City,
			&candidate.School, &candidate.GraduationYear, &candidate.Achievements, &candidate.Extracurriculars,
			&candidate.Essay, &candidate.MotivationStatement, &candidate.Disability, &candidate.Major,
			&candidate.PhotoURL, &candidate.PhotoAIFlag, &candidate.PhotoAINote, &candidate.Keywords,
			&candidate.CreatedAt, &candidate.Status)

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

func SubmitApplication(pool *pgxpool.Pool, emailSvc ...*EmailService) gin.HandlerFunc {
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
		var fullName, email string
		err := pool.QueryRow(c.Request.Context(),
			`INSERT INTO candidates (full_name, email, phone, telegram, age, city, school, graduation_year, achievements, extracurriculars, essay, motivation_statement, disability, major, status)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,'pending')
			 RETURNING id, full_name, email`,
			req.FullName, req.Email, req.Phone, req.Telegram, req.Age, req.City, req.School, req.GraduationYear,
			req.Achievements, req.Extracurriculars, req.Essay, req.MotivationStatement, req.Disability, req.Major,
		).Scan(&id, &fullName, &email)

		if err != nil {
			if isDuplicateKey(err) {
				c.JSON(409, gin.H{"error": "email already registered"})
				return
			}
			c.JSON(500, gin.H{"error": "failed to submit application"})
			return
		}

		// Send confirmation email (fire-and-forget)
		if len(emailSvc) > 0 && emailSvc[0] != nil && emailSvc[0].Enabled() {
			go emailSvc[0].SendApplicationReceived(email, fullName, id)
		}

		c.JSON(201, gin.H{"message": "Application submitted successfully", "id": id})
	}
}

func ListCandidates(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		status := c.Query("status")
		search := c.Query("search")
		major := c.Query("major")
		sortBy := c.DefaultQuery("sort_by", "created_at")
		order := c.DefaultQuery("order", "desc")
		limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
		offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

		// Numeric filters
		minScore, _ := strconv.ParseFloat(c.Query("min_score"), 64)
		maxScore, _ := strconv.ParseFloat(c.Query("max_score"), 64)
		minAge, _ := strconv.Atoi(c.Query("min_age"))
		maxAge, _ := strconv.Atoi(c.Query("max_age"))

		if limit > 100 {
			limit = 100
		}

		allowedSorts := map[string]string{
			"created_at":  "c.created_at",
			"full_name":   "c.full_name",
			"final_score": "a.final_score",
			"analyzed_at": "a.analyzed_at",
			"age":         "c.age",
		}
		sortCol, ok := allowedSorts[sortBy]
		if !ok {
			sortCol = "c.created_at"
		}
		if order != "asc" {
			order = "desc"
		}

		where := "WHERE 1=1"
		args := []interface{}{}
		argIdx := 1

		if status != "" && status != "all" {
			where += fmt.Sprintf(" AND c.status = $%d", argIdx)
			args = append(args, status)
			argIdx++
		}
		if search != "" {
			// Search by name, email, OR keywords
			where += fmt.Sprintf(
				` AND (c.full_name ILIKE $%d OR c.email ILIKE $%d OR $%d ILIKE ANY(c.keywords) OR c.achievements ILIKE $%d OR c.essay ILIKE $%d)`,
				argIdx, argIdx, argIdx, argIdx, argIdx)
			args = append(args, "%"+search+"%")
			argIdx++
		}
		if major != "" {
			where += fmt.Sprintf(" AND c.major = $%d", argIdx)
			args = append(args, major)
			argIdx++
		}
		if minScore > 0 {
			where += fmt.Sprintf(" AND a.final_score >= $%d", argIdx)
			args = append(args, minScore)
			argIdx++
		}
		if maxScore > 0 {
			where += fmt.Sprintf(" AND a.final_score <= $%d", argIdx)
			args = append(args, maxScore)
			argIdx++
		}
		if minAge > 0 {
			where += fmt.Sprintf(" AND c.age >= $%d", argIdx)
			args = append(args, minAge)
			argIdx++
		}
		if maxAge > 0 {
			where += fmt.Sprintf(" AND c.age <= $%d", argIdx)
			args = append(args, maxAge)
			argIdx++
		}

		var total int
		countQuery := fmt.Sprintf(`SELECT COUNT(*) FROM candidates c LEFT JOIN analyses a ON c.id = a.candidate_id %s`, where)
		pool.QueryRow(c.Request.Context(), countQuery, args...).Scan(&total)

		query := fmt.Sprintf(
			`SELECT c.id, c.full_name, c.email, c.city, c.school, c.major, c.status, c.created_at,
				a.final_score, a.category, a.analyzed_at, a.model_used, c.photo_url, c.photo_ai_flag, c.age
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
				&item.Major, &item.Status, &item.CreatedAt, &item.FinalScore, &item.Category, &item.AnalyzedAt,
				&item.ModelUsed, &item.PhotoURL, &item.PhotoAIFlag, &item.Age); err != nil {
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

		var candidate models.Candidate
		err = pool.QueryRow(c.Request.Context(),
			`SELECT id, full_name, email, phone, telegram, age, city, school, graduation_year,
				achievements, extracurriculars, essay, motivation_statement, disability, major,
				photo_url, photo_ai_flag, photo_ai_note, COALESCE(keywords, '{}'), created_at, status
			 FROM candidates WHERE id = $1`, id,
		).Scan(&candidate.ID, &candidate.FullName, &candidate.Email, &candidate.Phone, &candidate.Telegram,
			&candidate.Age, &candidate.City, &candidate.School, &candidate.GraduationYear,
			&candidate.Achievements, &candidate.Extracurriculars, &candidate.Essay,
			&candidate.MotivationStatement, &candidate.Disability, &candidate.Major,
			&candidate.PhotoURL, &candidate.PhotoAIFlag, &candidate.PhotoAINote,
			&candidate.Keywords, &candidate.CreatedAt, &candidate.Status)

		if err != nil {
			c.JSON(404, gin.H{"error": "candidate not found"})
			return
		}

		detail := models.CandidateDetail{
			Candidate: candidate,
			Decisions: []models.Decision{},
		}

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

func DeleteCandidate(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(400, gin.H{"error": "invalid candidate id"})
			return
		}
		tag, err := pool.Exec(c.Request.Context(), `DELETE FROM candidates WHERE id = $1`, id)
		if err != nil || tag.RowsAffected() == 0 {
			c.JSON(404, gin.H{"error": "candidate not found"})
			return
		}
		c.JSON(200, gin.H{"message": "candidate deleted"})
	}
}

func UpdateCandidate(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(400, gin.H{"error": "invalid candidate id"})
			return
		}

		var req struct {
			FullName            string  `json:"full_name" binding:"required"`
			Email               string  `json:"email" binding:"required,email"`
			Phone               *string `json:"phone"`
			Telegram            *string `json:"telegram"`
			Age                 *int    `json:"age"`
			City                *string `json:"city"`
			School              *string `json:"school"`
			GraduationYear      *int    `json:"graduation_year"`
			Achievements        *string `json:"achievements"`
			Extracurriculars    *string `json:"extracurriculars"`
			Essay               string  `json:"essay" binding:"required"`
			MotivationStatement *string `json:"motivation_statement"`
			Major               *string `json:"major"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			HandleValidationError(c, err)
			return
		}

		tag, err := pool.Exec(c.Request.Context(),
			`UPDATE candidates SET full_name=$1, email=$2, phone=$3, telegram=$4, age=$5, city=$6, school=$7,
			 graduation_year=$8, achievements=$9, extracurriculars=$10, essay=$11, motivation_statement=$12, major=$13
			 WHERE id=$14`,
			req.FullName, req.Email, req.Phone, req.Telegram, req.Age, req.City, req.School,
			req.GraduationYear, req.Achievements, req.Extracurriculars, req.Essay, req.MotivationStatement, req.Major, id)
		if err != nil {
			if isDuplicateKey(err) {
				c.JSON(409, gin.H{"error": "email already registered"})
				return
			}
			c.JSON(500, gin.H{"error": "failed to update candidate"})
			return
		}
		if tag.RowsAffected() == 0 {
			c.JSON(404, gin.H{"error": "candidate not found"})
			return
		}
		c.JSON(200, gin.H{"message": "candidate updated"})
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

// UpdateCandidateKeywords — save extracted keywords from AI analysis
func UpdateCandidateKeywords(pool *pgxpool.Pool, candidateID int, keywords []string) error {
	seen := map[string]bool{}
	unique := []string{}
	for _, kw := range keywords {
		kw = strings.ToLower(strings.TrimSpace(kw))
		if kw != "" && !seen[kw] {
			seen[kw] = true
			unique = append(unique, kw)
		}
	}
	ctx := context.Background()
	_, err := pool.Exec(ctx,
		`UPDATE candidates SET keywords = $1 WHERE id = $2`, unique, candidateID)
	return err
}

// GetSimilarCandidates returns candidates whose score is within ±3% of the given candidate.
func GetSimilarCandidates(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(400, gin.H{"error": "invalid candidate id"})
			return
		}

		// Get this candidate's score
		var score *float64
		pool.QueryRow(c.Request.Context(),
			`SELECT a.final_score FROM analyses a WHERE a.candidate_id = $1`, id).Scan(&score)
		if score == nil {
			c.JSON(200, []interface{}{})
			return
		}

		margin := 3.0
		lo := *score - margin
		hi := *score + margin

		rows, err := pool.Query(c.Request.Context(), `
			SELECT c.id, c.full_name, c.major, a.final_score, a.category, c.status
			FROM candidates c
			JOIN analyses a ON a.candidate_id = c.id
			WHERE c.id != $1 AND a.final_score BETWEEN $2 AND $3
			ORDER BY ABS(a.final_score - $4) ASC
			LIMIT 10`, id, lo, hi, *score)
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to query similar candidates"})
			return
		}
		defer rows.Close()

		type similarItem struct {
			ID         int      `json:"id"`
			FullName   string   `json:"full_name"`
			Major      *string  `json:"major"`
			FinalScore float64  `json:"final_score"`
			Category   string   `json:"category"`
			Status     string   `json:"status"`
		}
		var results []similarItem
		for rows.Next() {
			var s similarItem
			if err := rows.Scan(&s.ID, &s.FullName, &s.Major, &s.FinalScore, &s.Category, &s.Status); err == nil {
				results = append(results, s)
			}
		}
		if results == nil {
			results = []similarItem{}
		}
		c.JSON(200, results)
	}
}
