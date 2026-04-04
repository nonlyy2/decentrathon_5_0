package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/assylkhan/invisionu-backend/internal/middleware"
	"github.com/assylkhan/invisionu-backend/internal/models"
	"github.com/assylkhan/invisionu-backend/internal/youtube"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

var phoneRegex = regexp.MustCompile(`^\+?[0-9\s\-()]+$`)
var telegramRegex = regexp.MustCompile(`^@?[a-zA-Z0-9_]+$`)

func nilIntIfZero(i int) *int {
	if i == 0 {
		return nil
	}
	return &i
}

// countTotalChars returns the total character count of the 4 essay fields
func countTotalChars(achievements, extracurriculars, essay, motivation string) int {
	return len(achievements) + len(extracurriculars) + len(essay) + len(motivation)
}

// recalcAllComplexity recalculates review_complexity for ALL candidates as a percentage
// of the current global maximum total_chars. The candidate with the most text = 100%.
func recalcAllComplexity(pool *pgxpool.Pool) {
	ctx := context.Background()

	// Find current global max total_chars
	var maxChars int
	err := pool.QueryRow(ctx,
		`SELECT COALESCE(MAX(
			COALESCE(LENGTH(achievements),0) + COALESCE(LENGTH(extracurriculars),0) +
			COALESCE(LENGTH(essay),0) + COALESCE(LENGTH(motivation_statement),0)
		), 1) FROM candidates`,
	).Scan(&maxChars)
	if err != nil || maxChars == 0 {
		maxChars = 1
	}

	// Update all candidates' review_complexity as percentage of max
	_, err = pool.Exec(ctx,
		`UPDATE candidates SET review_complexity = ROUND(
			(COALESCE(LENGTH(achievements),0) + COALESCE(LENGTH(extracurriculars),0) +
			 COALESCE(LENGTH(essay),0) + COALESCE(LENGTH(motivation_statement),0))::numeric
			/ $1 * 100, 2
		)`, maxChars)
	if err != nil {
		log.Printf("recalcAllComplexity: %v", err)
	}
}

// RecalcComplexity is an HTTP handler to manually trigger recalculation
func RecalcComplexity(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		recalcAllComplexity(pool)
		c.JSON(200, gin.H{"message": "review complexity recalculated for all candidates"})
	}
}

func validateCandidateFields(req *models.CreateCandidateRequest) map[string]string {
	errs := map[string]string{}

	// Phone: only digits, spaces, +, -, ()
	if req.Phone != "" && !phoneRegex.MatchString(req.Phone) {
		errs["Phone"] = "phone must contain only digits and + - ( ) characters"
	}

	// Telegram: only latin letters, digits, underscore (no cyrillic)
	if req.Telegram != "" {
		tg := strings.TrimPrefix(req.Telegram, "@")
		if !telegramRegex.MatchString(tg) {
			errs["Telegram"] = "telegram username must contain only Latin letters, digits, and underscores"
		}
	}

	// City: no digits
	if req.City != "" {
		for _, ch := range req.City {
			if unicode.IsDigit(ch) {
				errs["City"] = "city name must not contain digits"
				break
			}
		}
	}

	return errs
}

// fetchAndStoreTranscript runs asynchronously after a candidate is saved.
// It validates the YouTube URL and stores the transcript (if any).
func fetchAndStoreTranscript(pool *pgxpool.Pool, candidateID int, youtubeURL, sttAPIKey, sttProvider string) {
	transcript, isValid, err := youtube.ValidateAndFetch(youtubeURL, sttAPIKey, sttProvider)
	if err != nil {
		log.Printf("youtube[%d]: %v", candidateID, err)
	}
	var transcriptVal *string
	if transcript != "" {
		transcriptVal = &transcript
	}
	_, dbErr := pool.Exec(context.Background(),
		`UPDATE candidates SET youtube_url_valid=$1, youtube_transcript=$2 WHERE id=$3`,
		isValid, transcriptVal, candidateID)
	if dbErr != nil {
		log.Printf("youtube[%d]: failed to save transcript: %v", candidateID, dbErr)
	}
}

func CreateCandidate(pool *pgxpool.Pool, sttAPIKey, sttProvider string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.CreateCandidateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			HandleValidationError(c, err)
			return
		}

		if errs := validateCandidateFields(&req); len(errs) > 0 {
			c.JSON(400, ErrorResponse{Error: "validation failed", Details: errs})
			return
		}

		var candidate models.Candidate
		err := pool.QueryRow(c.Request.Context(),
			`INSERT INTO candidates (full_name, email, phone, telegram, age, city, school, graduation_year, achievements, extracurriculars, essay, motivation_statement, disability, major, youtube_url)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15)
			 RETURNING id, full_name, email, phone, telegram, age, city, school, graduation_year, achievements, extracurriculars, essay, motivation_statement, disability, major, photo_url, photo_ai_flag, photo_ai_note, keywords, created_at, status, youtube_url, youtube_transcript, youtube_url_valid`,
			req.FullName, req.Email, req.Phone, req.Telegram, req.Age, req.City, req.School, req.GraduationYear,
			req.Achievements, req.Extracurriculars, req.Essay, req.MotivationStatement, req.Disability, req.Major, req.YouTubeURL,
		).Scan(&candidate.ID, &candidate.FullName, &candidate.Email, &candidate.Phone, &candidate.Telegram, &candidate.Age, &candidate.City,
			&candidate.School, &candidate.GraduationYear, &candidate.Achievements, &candidate.Extracurriculars,
			&candidate.Essay, &candidate.MotivationStatement, &candidate.Disability, &candidate.Major,
			&candidate.PhotoURL, &candidate.PhotoAIFlag, &candidate.PhotoAINote, &candidate.Keywords,
			&candidate.CreatedAt, &candidate.Status, &candidate.YouTubeURL, &candidate.YouTubeTranscript, &candidate.YouTubeURLValid)

		if err != nil {
			if isDuplicateKey(err) {
				c.JSON(409, gin.H{"error": "email already registered"})
				return
			}
			c.JSON(500, gin.H{"error": "failed to create candidate"})
			return
		}

		go fetchAndStoreTranscript(pool, candidate.ID, req.YouTubeURL, sttAPIKey, sttProvider)
		go recalcAllComplexity(pool)

		c.JSON(201, candidate)
	}
}

func SubmitApplication(pool *pgxpool.Pool, sttAPIKey, sttProvider string, emailSvc ...*EmailService) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req models.CreateCandidateRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			HandleValidationError(c, err)
			return
		}

		if errs := validateCandidateFields(&req); len(errs) > 0 {
			c.JSON(400, ErrorResponse{Error: "validation failed", Details: errs})
			return
		}

		if len(req.Essay) < 50 {
			c.JSON(400, gin.H{"error": "essay must be at least 50 characters"})
			return
		}

		// Initial review_complexity will be recalculated after insert
		totalChars := countTotalChars(req.Achievements, req.Extracurriculars, req.Essay, req.MotivationStatement)
		reviewComplexity := float64(totalChars) // temporary, will be recalculated

		// Build full_name from first_name + last_name if provided
		fullNameVal := req.FullName
		if fullNameVal == "" && (req.FirstName != "" || req.LastName != "") {
			fullNameVal = strings.TrimSpace(req.LastName + " " + req.FirstName)
		}

		var id int
		var fullName, email string
		err := pool.QueryRow(c.Request.Context(),
			`INSERT INTO candidates (full_name, first_name, last_name, patronymic, email, phone, telegram, age, date_of_birth, gender, city, home_country, school, graduation_year, nationality, iin, identity_doc_type, instagram, whatsapp, achievements, extracurriculars, essay, motivation_statement, disability, major, youtube_url, exam_type, ielts_score, toefl_score, certificate_type, personality_answers, review_complexity, status)
			 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21,$22,$23,$24,$25,$26,$27,$28,$29,$30,$31,$32,'pending')
			 RETURNING id, full_name, email`,
			fullNameVal, nilIfEmpty(req.FirstName), nilIfEmpty(req.LastName), nilIfEmpty(req.Patronymic),
			req.Email, nilIfEmpty(req.Phone), nilIfEmpty(req.Telegram),
			nilIntIfZero(req.Age), nilIfEmpty(req.DateOfBirth), nilIfEmpty(req.Gender),
			nilIfEmpty(req.City), nilIfEmpty(req.HomeCountry), nilIfEmpty(req.School), nilIntIfZero(req.GraduationYear),
			nilIfEmpty(req.Nationality), nilIfEmpty(req.IIN), nilIfEmpty(req.IdentityDocType),
			nilIfEmpty(req.Instagram), nilIfEmpty(req.WhatsApp),
			nilIfEmpty(req.Achievements), nilIfEmpty(req.Extracurriculars), req.Essay,
			nilIfEmpty(req.MotivationStatement), req.Disability, req.Major, nilIfEmpty(req.YouTubeURL),
			nilIfEmpty(req.ExamType), req.IELTSScore, req.TOEFLScore,
			nilIfEmpty(req.CertificateType), nilIfEmpty(req.PersonalityAnswers),
			reviewComplexity,
		).Scan(&id, &fullName, &email)

		if err != nil {
			if isDuplicateKey(err) {
				c.JSON(409, gin.H{"error": "email already registered"})
				return
			}
			c.JSON(500, gin.H{"error": "failed to submit application"})
			return
		}

		// Async: validate YouTube URL and fetch transcript
		go fetchAndStoreTranscript(pool, id, req.YouTubeURL, sttAPIKey, sttProvider)

		// Recalculate complexity for ALL candidates (new max may have changed)
		go recalcAllComplexity(pool)

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
			"net_score":   "COALESCE((SELECT COUNT(*) FILTER (WHERE cd2.decision='upvote') - COUNT(*) FILTER (WHERE cd2.decision='downvote') FROM committee_decisions cd2 WHERE cd2.candidate_id=c.id), 0)",
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
				a.final_score, a.category, a.analyzed_at, a.model_used, c.photo_url, c.photo_ai_flag, c.age,
				COALESCE((SELECT COUNT(*) FILTER (WHERE cd.decision='upvote') - COUNT(*) FILTER (WHERE cd.decision='downvote') FROM committee_decisions cd WHERE cd.candidate_id=c.id), 0) AS net_score
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
				&item.ModelUsed, &item.PhotoURL, &item.PhotoAIFlag, &item.Age, &item.NetScore); err != nil {
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
				photo_url, photo_ai_flag, photo_ai_note, COALESCE(keywords, '{}'), created_at, status,
				youtube_url, youtube_transcript, youtube_url_valid,
				first_name, last_name, patronymic, date_of_birth, gender, nationality, iin,
				identity_doc_type, instagram, whatsapp, home_country,
				exam_type, ielts_score, toefl_score, english_cert_url,
				certificate_type, certificate_url, additional_docs_url,
				personality_answers, review_complexity
			 FROM candidates WHERE id = $1`, id,
		).Scan(&candidate.ID, &candidate.FullName, &candidate.Email, &candidate.Phone, &candidate.Telegram,
			&candidate.Age, &candidate.City, &candidate.School, &candidate.GraduationYear,
			&candidate.Achievements, &candidate.Extracurriculars, &candidate.Essay,
			&candidate.MotivationStatement, &candidate.Disability, &candidate.Major,
			&candidate.PhotoURL, &candidate.PhotoAIFlag, &candidate.PhotoAINote,
			&candidate.Keywords, &candidate.CreatedAt, &candidate.Status,
			&candidate.YouTubeURL, &candidate.YouTubeTranscript, &candidate.YouTubeURLValid,
			&candidate.FirstName, &candidate.LastName, &candidate.Patronymic,
			&candidate.DateOfBirth, &candidate.Gender, &candidate.Nationality, &candidate.IIN,
			&candidate.IdentityDocType, &candidate.Instagram, &candidate.WhatsApp, &candidate.HomeCountry,
			&candidate.ExamType, &candidate.IELTSScore, &candidate.TOEFLScore, &candidate.EnglishCertURL,
			&candidate.CertificateType, &candidate.CertificateURL, &candidate.AdditionalDocsURL,
			&candidate.PersonalityAnswers, &candidate.ReviewComplexity)

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
			 summary, key_strengths, red_flags, analyzed_at, model_used, recommended_major, major_reason_note
			 FROM analyses WHERE candidate_id = $1`, id,
		).Scan(&analysis.ID, &analysis.CandidateID, &analysis.ScoreLeadership, &analysis.ScoreMotivation,
			&analysis.ScoreGrowth, &analysis.ScoreVision, &analysis.ScoreCommunication,
			&analysis.FinalScore, &analysis.Category, &analysis.AIGeneratedRisk, &analysis.AIGeneratedScore, &analysis.IncompleteFlag,
			&analysis.ExplanationLeadership, &analysis.ExplanationMotivation, &analysis.ExplanationGrowth,
			&analysis.ExplanationVision, &analysis.ExplanationCommunication,
			&analysis.Summary, &analysis.KeyStrengths, &analysis.RedFlags, &analysis.AnalyzedAt, &analysis.ModelUsed,
			&analysis.RecommendedMajor, &analysis.MajorReasonNote)

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
		if middleware.IsRole(c, "auditor") {
			c.JSON(403, gin.H{"error": "auditors cannot delete candidates"})
			return
		}
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

func UpdateCandidate(pool *pgxpool.Pool, sttAPIKey, sttProvider string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if middleware.IsRole(c, "auditor") {
			c.JSON(403, gin.H{"error": "auditors cannot modify candidates"})
			return
		}
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
			YouTubeURL          *string `json:"youtube_url"`
		}
		if err := c.ShouldBindJSON(&req); err != nil {
			HandleValidationError(c, err)
			return
		}

		tag, err := pool.Exec(c.Request.Context(),
			`UPDATE candidates SET full_name=$1, email=$2, phone=$3, telegram=$4, age=$5, city=$6, school=$7,
			 graduation_year=$8, achievements=$9, extracurriculars=$10, essay=$11, motivation_statement=$12, major=$13, youtube_url=$14
			 WHERE id=$15`,
			req.FullName, req.Email, req.Phone, req.Telegram, req.Age, req.City, req.School,
			req.GraduationYear, req.Achievements, req.Extracurriculars, req.Essay, req.MotivationStatement, req.Major, req.YouTubeURL, id)
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
		// Re-fetch transcript if YouTube URL was provided
		if req.YouTubeURL != nil && *req.YouTubeURL != "" {
			go fetchAndStoreTranscript(pool, id, *req.YouTubeURL, sttAPIKey, sttProvider)
		}
		c.JSON(200, gin.H{"message": "candidate updated"})
	}
}

func UpdateCandidateStatus(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		if middleware.IsRole(c, "auditor") {
			c.JSON(403, gin.H{"error": "auditors cannot change candidate status"})
			return
		}
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

// FetchTranscriptManually re-validates the YouTube URL and fetches/refreshes the transcript.
// POST /candidates/:id/fetch-transcript
func FetchTranscriptManually(pool *pgxpool.Pool, sttAPIKey, sttProvider string) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(400, gin.H{"error": "invalid candidate id"})
			return
		}

		var youtubeURL *string
		err = pool.QueryRow(c.Request.Context(),
			`SELECT youtube_url FROM candidates WHERE id = $1`, id).Scan(&youtubeURL)
		if err != nil {
			c.JSON(404, gin.H{"error": "candidate not found"})
			return
		}
		if youtubeURL == nil || *youtubeURL == "" {
			c.JSON(400, gin.H{"error": "no YouTube URL on this candidate"})
			return
		}

		transcript, isValid, fetchErr := youtube.ValidateAndFetch(*youtubeURL, sttAPIKey, sttProvider)
		var transcriptVal *string
		if transcript != "" {
			transcriptVal = &transcript
		}
		pool.Exec(c.Request.Context(),
			`UPDATE candidates SET youtube_url_valid=$1, youtube_transcript=$2 WHERE id=$3`,
			isValid, transcriptVal, id)

		if !isValid {
			c.JSON(422, gin.H{"error": fetchErr.Error(), "youtube_url_valid": false})
			return
		}
		if transcript == "" {
			c.JSON(200, gin.H{"message": "Video is accessible but no transcript/captions are available", "youtube_url_valid": true, "youtube_transcript": nil})
			return
		}
		c.JSON(200, gin.H{"message": "Transcript fetched successfully", "youtube_url_valid": true, "youtube_transcript": transcript})
	}
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
