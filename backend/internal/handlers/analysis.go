package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/assylkhan/invisionu-backend/internal/models"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
)

// AIProviders maps provider names to their single-candidate analyze functions
type AIProviders map[string]func(ctx context.Context, candidate *models.Candidate) (*models.Analysis, error)

// AIBatchProviders maps provider names to their batch analyze functions
type AIBatchProviders map[string]func(ctx context.Context, candidates []models.Candidate) ([]*models.Analysis, error)

// GetAIProviders returns the list of available AI providers
func GetAIProviders(providers AIProviders, defaultProvider string) gin.HandlerFunc {
	return func(c *gin.Context) {
		available := make([]string, 0, len(providers))
		for k := range providers {
			available = append(available, k)
		}
		sort.Strings(available)
		c.JSON(200, gin.H{
			"providers":        available,
			"default_provider": defaultProvider,
		})
	}
}

func GetAnalysis(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		candidateID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(400, gin.H{"error": "invalid candidate id"})
			return
		}

		var analysis models.Analysis
		err = pool.QueryRow(c.Request.Context(),
			`SELECT id, candidate_id, score_leadership, score_motivation, score_growth, score_vision, score_communication,
			 final_score, category, ai_generated_risk, COALESCE(ai_generated_score, 0), incomplete_flag,
			 explanation_leadership, explanation_motivation, explanation_growth, explanation_vision, explanation_communication,
			 summary, key_strengths, red_flags, analyzed_at, model_used
			 FROM analyses WHERE candidate_id = $1`, candidateID,
		).Scan(&analysis.ID, &analysis.CandidateID, &analysis.ScoreLeadership, &analysis.ScoreMotivation,
			&analysis.ScoreGrowth, &analysis.ScoreVision, &analysis.ScoreCommunication,
			&analysis.FinalScore, &analysis.Category, &analysis.AIGeneratedRisk, &analysis.AIGeneratedScore, &analysis.IncompleteFlag,
			&analysis.ExplanationLeadership, &analysis.ExplanationMotivation, &analysis.ExplanationGrowth,
			&analysis.ExplanationVision, &analysis.ExplanationCommunication,
			&analysis.Summary, &analysis.KeyStrengths, &analysis.RedFlags, &analysis.AnalyzedAt, &analysis.ModelUsed)

		if err != nil {
			c.JSON(404, gin.H{"error": "analysis not found"})
			return
		}

		c.JSON(http.StatusOK, analysis)
	}
}

func DeleteAnalysis(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		candidateID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(400, gin.H{"error": "invalid candidate id"})
			return
		}
		ctx := c.Request.Context()
		tag, err := pool.Exec(ctx, `DELETE FROM analyses WHERE candidate_id = $1`, candidateID)
		if err != nil || tag.RowsAffected() == 0 {
			c.JSON(404, gin.H{"error": "analysis not found"})
			return
		}
		pool.Exec(ctx, `UPDATE candidates SET status = 'pending' WHERE id = $1`, candidateID)
		candidateAnalyses.Delete(candidateID)
		c.JSON(200, gin.H{"message": "analysis deleted"})
	}
}

func DeleteAllAnalyses(pool *pgxpool.Pool) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := c.Request.Context()
		// Delete analyses only for candidates without a committee decision (not shortlisted/rejected/waitlisted)
		tag, err := pool.Exec(ctx,
			`DELETE FROM analyses WHERE candidate_id IN (
				SELECT id FROM candidates WHERE status NOT IN ('shortlisted', 'rejected', 'waitlisted')
			)`)
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to delete analyses"})
			return
		}
		// Reset those candidates back to pending
		pool.Exec(ctx, `UPDATE candidates SET status = 'pending' WHERE status = 'analyzed'`)
		// Clear in-memory status cache
		candidateAnalyses.Range(func(key, _ any) bool {
			candidateAnalyses.Delete(key)
			return true
		})
		c.JSON(200, gin.H{"deleted": tag.RowsAffected()})
	}
}

func SaveAnalysis(ctx context.Context, pool *pgxpool.Pool, analysis *models.Analysis) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx,
		`INSERT INTO analyses (candidate_id, score_leadership, score_motivation, score_growth, score_vision, score_communication,
		 final_score, category, ai_generated_risk, ai_generated_score, incomplete_flag,
		 explanation_leadership, explanation_motivation, explanation_growth, explanation_vision, explanation_communication,
		 summary, key_strengths, red_flags, model_used)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20)`,
		analysis.CandidateID, analysis.ScoreLeadership, analysis.ScoreMotivation, analysis.ScoreGrowth,
		analysis.ScoreVision, analysis.ScoreCommunication, analysis.FinalScore, analysis.Category,
		analysis.AIGeneratedRisk, analysis.AIGeneratedScore, analysis.IncompleteFlag,
		analysis.ExplanationLeadership, analysis.ExplanationMotivation, analysis.ExplanationGrowth,
		analysis.ExplanationVision, analysis.ExplanationCommunication,
		analysis.Summary, analysis.KeyStrengths, analysis.RedFlags, analysis.ModelUsed)

	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `UPDATE candidates SET status = 'analyzed' WHERE id = $1`, analysis.CandidateID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// resolveProvider picks the analyze function based on ?provider= query param or default
func resolveProvider(c *gin.Context, providers AIProviders, defaultProvider string) (func(ctx context.Context, candidate *models.Candidate) (*models.Analysis, error), bool) {
	providerName := c.Query("provider")
	if providerName == "" {
		providerName = defaultProvider
	}
	fn, ok := providers[providerName]
	if !ok {
		c.JSON(400, gin.H{"error": fmt.Sprintf("unknown AI provider: %s", providerName)})
		return nil, false
	}
	return fn, true
}

// AnalyzeSingleCandidate creates a handler — providers map is injected from main
func AnalyzeSingleCandidate(pool *pgxpool.Pool, providers AIProviders, defaultProvider string) gin.HandlerFunc {
	return func(c *gin.Context) {
		candidateID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(400, gin.H{"error": "invalid candidate id"})
			return
		}

		analyzeFunc, ok := resolveProvider(c, providers, defaultProvider)
		if !ok {
			return
		}

		// Get candidate
		var candidate models.Candidate
		err = pool.QueryRow(c.Request.Context(),
			`SELECT id, full_name, email, phone, telegram, age, city, school, graduation_year, achievements, extracurriculars, essay, motivation_statement, created_at, status
			 FROM candidates WHERE id = $1`, candidateID,
		).Scan(&candidate.ID, &candidate.FullName, &candidate.Email, &candidate.Phone, &candidate.Telegram, &candidate.Age, &candidate.City,
			&candidate.School, &candidate.GraduationYear, &candidate.Achievements, &candidate.Extracurriculars,
			&candidate.Essay, &candidate.MotivationStatement, &candidate.CreatedAt, &candidate.Status)
		if err != nil {
			c.JSON(404, gin.H{"error": "candidate not found"})
			return
		}

		// Check existing analysis
		force := c.Query("force") == "true"
		var existingID int
		err = pool.QueryRow(c.Request.Context(),
			`SELECT id FROM analyses WHERE candidate_id = $1`, candidateID).Scan(&existingID)
		if err == nil && !force {
			c.JSON(409, gin.H{"error": "candidate already analyzed, use ?force=true to re-analyze"})
			return
		}
		if err == nil && force {
			pool.Exec(c.Request.Context(), `DELETE FROM analyses WHERE candidate_id = $1`, candidateID)
		}

		if analyzeFunc == nil {
			c.JSON(503, gin.H{"error": "AI provider not configured"})
			return
		}

		// Check if already running
		if raw, ok := candidateAnalyses.Load(candidateID); ok {
			if raw.(candidateAnalysisState).Running {
				c.JSON(409, gin.H{"error": "analysis already running"})
				return
			}
		}

		// Mark as running and return immediately — analysis runs in background
		candidateAnalyses.Store(candidateID, candidateAnalysisState{Running: true})
		c.JSON(202, gin.H{"message": "analysis started"})

		go func(cand models.Candidate) {
			ctx := context.Background()
			analysis, err := analyzeFunc(ctx, &cand)
			if err != nil {
				log.Printf("Analysis failed for candidate %d: %v", cand.ID, err)
				candidateAnalyses.Store(cand.ID, candidateAnalysisState{Running: false, Failed: true, ErrMsg: err.Error()})
				return
			}
			if err := SaveAnalysis(ctx, pool, analysis); err != nil {
				log.Printf("Save failed for candidate %d: %v", cand.ID, err)
				candidateAnalyses.Store(cand.ID, candidateAnalysisState{Running: false, Failed: true, ErrMsg: "failed to save analysis"})
				return
			}
			candidateAnalyses.Store(cand.ID, candidateAnalysisState{Running: false, Failed: false})
			log.Printf("Analysis completed for candidate %d", cand.ID)
		}(candidate)
	}
}

// Per-candidate async analysis tracking
type candidateAnalysisState struct {
	Running bool
	Failed  bool
	ErrMsg  string
}

var candidateAnalyses sync.Map // map[int]candidateAnalysisState

func GetCandidateAnalysisStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		candidateID, err := strconv.Atoi(c.Param("id"))
		if err != nil {
			c.JSON(400, gin.H{"error": "invalid candidate id"})
			return
		}
		if raw, ok := candidateAnalyses.Load(candidateID); ok {
			state := raw.(candidateAnalysisState)
			c.JSON(200, gin.H{
				"running": state.Running,
				"failed":  state.Failed,
				"error":   state.ErrMsg,
			})
			return
		}
		c.JSON(200, gin.H{"running": false, "failed": false, "error": ""})
	}
}

// Bulk analysis
var batchCancel context.CancelFunc

var batchStatus struct {
	sync.Mutex
	Running   bool     `json:"running"`
	Processed int      `json:"processed"`
	Total     int      `json:"total"`
	Errors    []string `json:"errors"`
}

func AnalyzeAllPending(pool *pgxpool.Pool, providers AIProviders, batchProviders AIBatchProviders, defaultProvider string) gin.HandlerFunc {
	return func(c *gin.Context) {
		analyzeFunc, ok := resolveProvider(c, providers, defaultProvider)
		if !ok {
			return
		}
		if analyzeFunc == nil {
			c.JSON(503, gin.H{"error": "AI provider not configured"})
			return
		}
		providerName := c.Query("provider")
		if providerName == "" {
			providerName = defaultProvider
		}
		batchFunc := batchProviders[providerName] // may be nil

		batchStatus.Lock()
		if batchStatus.Running {
			batchStatus.Unlock()
			c.JSON(409, gin.H{"error": "batch analysis already running"})
			return
		}
		batchStatus.Unlock()

		rows, err := pool.Query(c.Request.Context(),
			`SELECT id, full_name, email, phone, telegram, age, city, school, graduation_year, achievements, extracurriculars, essay, motivation_statement, created_at, status
			 FROM candidates WHERE status = 'pending'`)
		if err != nil {
			c.JSON(500, gin.H{"error": "failed to query candidates"})
			return
		}

		var candidates []models.Candidate
		for rows.Next() {
			var cand models.Candidate
			if err := rows.Scan(&cand.ID, &cand.FullName, &cand.Email, &cand.Phone, &cand.Telegram, &cand.Age, &cand.City,
				&cand.School, &cand.GraduationYear, &cand.Achievements, &cand.Extracurriculars,
				&cand.Essay, &cand.MotivationStatement, &cand.CreatedAt, &cand.Status); err == nil {
				candidates = append(candidates, cand)
			}
		}
		rows.Close()

		if len(candidates) == 0 {
			c.JSON(200, gin.H{"message": "no pending candidates"})
			return
		}

		ctx, cancel := context.WithCancel(context.Background())
		batchCancel = cancel

		batchStatus.Lock()
		batchStatus.Running = true
		batchStatus.Processed = 0
		batchStatus.Total = len(candidates)
		batchStatus.Errors = []string{}
		batchStatus.Unlock()

		const batchSize = 10
		const maxConcurrent = 5

		go func() {
			defer func() {
				cancel()
				batchStatus.Lock()
				batchStatus.Running = false
				batchStatus.Unlock()
			}()

			isRateLimited := func(err error) bool {
				msg := err.Error()
				return strings.Contains(msg, "429") || strings.Contains(msg, "RESOURCE_EXHAUSTED") ||
					strings.Contains(msg, "rate-limited") || strings.Contains(msg, "rate limit")
			}

			// Split into batches
			var batches [][]models.Candidate
			for i := 0; i < len(candidates); i += batchSize {
				end := i + batchSize
				if end > len(candidates) {
					end = len(candidates)
				}
				batches = append(batches, candidates[i:end])
			}

			type batchResult struct {
				analyses []*models.Analysis
				batch    []models.Candidate
				err      error
				batchIdx int
			}

			sem := make(chan struct{}, maxConcurrent)
			results := make(chan batchResult, len(batches))
			rateLimitHit := false

			var wg sync.WaitGroup
			for idx, b := range batches {
				if rateLimitHit {
					break
				}
				// Check for cancellation
				select {
				case <-ctx.Done():
					batchStatus.Lock()
					batchStatus.Errors = append(batchStatus.Errors, "batch stopped by user")
					batchStatus.Unlock()
					goto done
				default:
				}
				wg.Add(1)
				sem <- struct{}{}
				go func(bIdx int, batch []models.Candidate) {
					defer wg.Done()
					defer func() { <-sem }()
					if batchFunc != nil {
						analyses, err := batchFunc(ctx, batch)
						results <- batchResult{analyses: analyses, batch: batch, err: err, batchIdx: bIdx}
					} else {
						var analyses []*models.Analysis
						for i := range batch {
							select {
							case <-ctx.Done():
								return
							default:
							}
							a, err := analyzeFunc(ctx, &batch[i])
							if err != nil {
								results <- batchResult{err: err, batch: batch[i : i+1], batchIdx: bIdx}
								if isRateLimited(err) {
									return
								}
								continue
							}
							analyses = append(analyses, a)
						}
						results <- batchResult{analyses: analyses, batch: batch, err: nil, batchIdx: bIdx}
					}
				}(idx, b)
			}

		done:
			// Close results after all goroutines finish
			go func() {
				wg.Wait()
				close(results)
			}()

			processedCount := 0
			for r := range results {
				if r.err != nil {
					if isRateLimited(r.err) {
						rateLimitHit = true
						batchStatus.Lock()
						batchStatus.Errors = append(batchStatus.Errors, fmt.Sprintf("rate limit hit at batch %d: %v", r.batchIdx, r.err))
						batchStatus.Unlock()
					} else {
						// Fall back: process individually
						for i := range r.batch {
							select {
							case <-ctx.Done():
								processedCount++
								continue
							default:
							}
							a, err2 := analyzeFunc(ctx, &r.batch[i])
							processedCount++
							if err2 != nil {
								batchStatus.Lock()
								batchStatus.Errors = append(batchStatus.Errors, fmt.Sprintf("candidate %d: %v", r.batch[i].ID, err2))
								batchStatus.Unlock()
								continue
							}
							if err2 = SaveAnalysis(ctx, pool, a); err2 != nil {
								batchStatus.Lock()
								batchStatus.Errors = append(batchStatus.Errors, fmt.Sprintf("candidate %d save: %v", r.batch[i].ID, err2))
								batchStatus.Unlock()
							}
						}
					}
				} else {
					for j, a := range r.analyses {
						processedCount++
						if a == nil || j >= len(r.batch) {
							continue
						}
						if err2 := SaveAnalysis(ctx, pool, a); err2 != nil {
							batchStatus.Lock()
							batchStatus.Errors = append(batchStatus.Errors, fmt.Sprintf("candidate %d save: %v", r.batch[j].ID, err2))
							batchStatus.Unlock()
						}
					}
				}
				batchStatus.Lock()
				batchStatus.Processed = processedCount
				batchStatus.Unlock()
			}
		}()

		c.JSON(202, gin.H{"message": "batch analysis started", "count": len(candidates)})
	}
}

// AITextGenerators maps provider names to a raw text-generation function
type AITextGenerators map[string]func(ctx context.Context, systemPrompt, userMsg string) (string, error)

// AIRecommendRequest is the body for POST /candidates/ai-recommend
type AIRecommendRequest struct {
	CandidateIDs []int `json:"candidate_ids" binding:"required,min=2"`
	SelectCount  int   `json:"select_count" binding:"required,min=1"`
}

// RecommendCandidates picks the best N candidates using AI
func RecommendCandidates(pool *pgxpool.Pool, textGens AITextGenerators, defaultProvider string) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req AIRecommendRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(400, gin.H{"error": err.Error()})
			return
		}
		if req.SelectCount >= len(req.CandidateIDs) {
			c.JSON(400, gin.H{"error": "select_count must be less than total candidates"})
			return
		}

		ctx := c.Request.Context()

		type candRow struct {
			ID         int
			Name       string
			FinalScore float64
			Category   string
			Summary    string
			Strengths  []string
			RedFlags   []string
		}
		var cands []candRow
		for _, id := range req.CandidateIDs {
			var r candRow
			r.ID = id
			err := pool.QueryRow(ctx,
				`SELECT c.full_name, COALESCE(a.final_score,0), COALESCE(a.category,''),
				        COALESCE(a.summary,''), COALESCE(a.key_strengths,ARRAY[]::text[]), COALESCE(a.red_flags,ARRAY[]::text[])
				 FROM candidates c LEFT JOIN analyses a ON c.id=a.candidate_id WHERE c.id=$1`, id,
			).Scan(&r.Name, &r.FinalScore, &r.Category, &r.Summary, &r.Strengths, &r.RedFlags)
			if err != nil {
				c.JSON(404, gin.H{"error": fmt.Sprintf("candidate %d not found", id)})
				return
			}
			cands = append(cands, r)
		}

		providerName := c.Query("provider")
		if providerName == "" {
			providerName = defaultProvider
		}
		genFn, ok := textGens[providerName]
		if !ok {
			c.JSON(400, gin.H{"error": "unknown AI provider: " + providerName})
			return
		}

		sysPrompt := fmt.Sprintf(
			"You are an expert admissions committee member for inVision U. "+
				"Select the best %d candidates from the list provided. "+
				"Respond ONLY with a valid JSON object — no text before or after:\n"+
				`{"selected":[{"index":<0-based>,"reason":"<2 sentences>"}],"overall_reasoning":"<2-3 sentences>"}`,
			req.SelectCount)

		var sb strings.Builder
		sb.WriteString(fmt.Sprintf("From the %d candidates below, pick the top %d:\n\n", len(cands), req.SelectCount))
		for i, r := range cands {
			sb.WriteString(fmt.Sprintf("[%d] %s | Score: %.1f | %s\n    %s\n", i, r.Name, r.FinalScore, r.Category, r.Summary))
			if len(r.Strengths) > 0 {
				sb.WriteString(fmt.Sprintf("    Strengths: %s\n", strings.Join(r.Strengths, "; ")))
			}
			if len(r.RedFlags) > 0 {
				sb.WriteString(fmt.Sprintf("    Red flags: %s\n", strings.Join(r.RedFlags, "; ")))
			}
		}

		responseText, err := genFn(ctx, sysPrompt, sb.String())
		if err != nil {
			c.JSON(500, gin.H{"error": "AI failed: " + err.Error()})
			return
		}

		var parsed struct {
			Selected []struct {
				Index  int    `json:"index"`
				Reason string `json:"reason"`
			} `json:"selected"`
			OverallReasoning string `json:"overall_reasoning"`
		}
		if err := json.Unmarshal([]byte(responseText), &parsed); err != nil {
			c.JSON(500, gin.H{"error": "failed to parse AI response"})
			return
		}

		type outItem struct {
			ID     int     `json:"id"`
			Name   string  `json:"name"`
			Score  float64 `json:"score"`
			Reason string  `json:"reason"`
		}
		var out []outItem
		for _, s := range parsed.Selected {
			if s.Index < 0 || s.Index >= len(cands) {
				continue
			}
			r := cands[s.Index]
			out = append(out, outItem{ID: r.ID, Name: r.Name, Score: r.FinalScore, Reason: s.Reason})
		}
		c.JSON(200, gin.H{"selected": out, "overall_reasoning": parsed.OverallReasoning})
	}
}

func StopBatch() gin.HandlerFunc {
	return func(c *gin.Context) {
		batchStatus.Lock()
		running := batchStatus.Running
		batchStatus.Unlock()
		if !running {
			c.JSON(200, gin.H{"message": "no batch running"})
			return
		}
		if batchCancel != nil {
			batchCancel()
		}
		c.JSON(200, gin.H{"message": "batch stop requested"})
	}
}

func GetBatchStatus() gin.HandlerFunc {
	return func(c *gin.Context) {
		batchStatus.Lock()
		defer batchStatus.Unlock()
		c.JSON(200, gin.H{
			"running":   batchStatus.Running,
			"processed": batchStatus.Processed,
			"total":     batchStatus.Total,
			"errors":    batchStatus.Errors,
		})
	}
}
