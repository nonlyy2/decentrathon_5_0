package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/assylkhan/invisionu-backend/internal/config"
	"github.com/assylkhan/invisionu-backend/internal/database"
	"github.com/assylkhan/invisionu-backend/internal/gemini"
	"github.com/assylkhan/invisionu-backend/internal/handlers"
	"github.com/assylkhan/invisionu-backend/internal/middleware"
	"github.com/assylkhan/invisionu-backend/internal/ollama"
	"github.com/assylkhan/invisionu-backend/internal/seed"
	"github.com/assylkhan/invisionu-backend/internal/telegram_bot"
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	// Database
	if cfg.DatabaseURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set. Set it to the Railway Postgres connection string.")
	}
	pool, err := database.Connect(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer pool.Close()

	if err := database.RunMigrations(pool); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	// Seed default admin
	if err := seed.SeedAdminUser(pool); err != nil {
		log.Printf("Warning: failed to seed admin user: %v", err)
	}
	// Upgrade legacy admin role to superadmin
	if err := seed.EnsureSuperAdminRole(pool); err != nil {
		log.Printf("Warning: failed to upgrade admin role: %v", err)
	}

	// Seed candidates if --seed / --force-seed / --seed-only / --force-seed-only flag
	seedOnly := false
	for _, arg := range os.Args[1:] {
		switch arg {
		case "--seed":
			if err := seed.SeedCandidates(pool, false); err != nil {
				log.Printf("Warning: failed to seed candidates: %v", err)
			}
		case "--force-seed":
			if err := seed.SeedCandidates(pool, true); err != nil {
				log.Printf("Warning: failed to seed candidates: %v", err)
			}
		case "--seed-only":
			if err := seed.SeedCandidates(pool, false); err != nil {
				log.Fatalf("Seed failed: %v", err)
			}
			seedOnly = true
		case "--force-seed-only":
			if err := seed.SeedCandidates(pool, true); err != nil {
				log.Fatalf("Seed failed: %v", err)
			}
			seedOnly = true
		}
	}
	if seedOnly {
		log.Println("Seed-only mode: exiting.")
		return
	}

	// AI clients
	aiProviders := make(handlers.AIProviders)
	batchProviders := make(handlers.AIBatchProviders)
	textGens := make(handlers.AITextGenerators)

	if cfg.GeminiAPIKey != "" {
		geminiClient := gemini.NewClient(cfg.GeminiAPIKey)
		aiProviders["gemini"] = geminiClient.AnalyzeCandidate
		batchProviders["gemini"] = geminiClient.AnalyzeBatch
		textGens["gemini"] = geminiClient.GenerateText
		log.Printf("Gemini initialized (model=%s)", gemini.ModelName)
	}

	ollamaClient := ollama.NewClient(cfg.OllamaURL, cfg.OllamaModel)
	aiProviders["ollama"] = ollamaClient.AnalyzeCandidate
	batchProviders["ollama"] = ollamaClient.AnalyzeBatch
	textGens["ollama"] = ollamaClient.GenerateText
	if isOllamaAvailable(cfg.OllamaURL) {
		log.Printf("Ollama initialized (url=%s, model=%s)", cfg.OllamaURL, cfg.OllamaModel)
	} else {
		log.Printf("Ollama registered but not reachable at %s", cfg.OllamaURL)
	}

	defaultProvider := cfg.AIProvider
	if _, ok := aiProviders[defaultProvider]; !ok {
		for k := range aiProviders {
			defaultProvider = k
			break
		}
	}
	log.Printf("Default AI provider: %s", defaultProvider)

	// Email service
	emailSvc := handlers.NewEmailService(cfg, pool)
	if emailSvc.Enabled() {
		log.Printf("Email service initialized (SMTP: %s:%d)", cfg.SMTPHost, cfg.SMTPPort)
	} else {
		log.Printf("Email service disabled (SMTP not configured)")
	}

	// Ensure upload directory
	if err := os.MkdirAll(cfg.UploadDir, 0755); err != nil {
		log.Printf("Warning: could not create upload dir %s: %v", cfg.UploadDir, err)
	}

	// Router
	router := gin.Default()
	router.Use(middleware.CORSMiddleware(cfg.AllowOrigins))
	router.Use(middleware.NoCacheMiddleware())

	// Serve uploaded files
	router.Static("/uploads", cfg.UploadDir)

	api := router.Group("/api")

	// Public routes
	api.GET("/health", func(c *gin.Context) { c.JSON(200, gin.H{"status": "ok"}) })
	api.POST("/apply", handlers.SubmitApplication(pool, cfg.WhisperAPIKey, cfg.WhisperProvider, emailSvc))
	api.POST("/auth/register", handlers.Register(pool))
	api.POST("/auth/login", handlers.Login(pool, cfg.JWTSecret))
	api.GET("/majors", handlers.GetMajors())
	// Photo upload is public so the apply page (no JWT) can attach a photo right after submission
	api.POST("/candidates/:id/photo", handlers.UploadCandidatePhoto(pool, cfg.UploadDir, cfg.GeminiAPIKey))
	// Document uploads (english cert, certificate, additional docs) — public for apply page
	api.POST("/candidates/:id/document/:docType", handlers.UploadCandidateDocument(pool, cfg.UploadDir))

	// Public Telegram Mini App status endpoint (validated by initData)
	api.GET("/tma/status", handlers.GetTMAStatusByChatID(pool, cfg.TelegramBotToken))

	// Protected routes
	protected := api.Group("/", middleware.AuthRequired(cfg.JWTSecret))
	{
		// Candidates — manager+
		protected.GET("/candidates", handlers.ListCandidates(pool))
		protected.POST("/candidates", handlers.CreateCandidate(pool, cfg.WhisperAPIKey, cfg.WhisperProvider))
		protected.GET("/candidates/:id", handlers.GetCandidate(pool))
		protected.PATCH("/candidates/:id", handlers.UpdateCandidate(pool, cfg.WhisperAPIKey, cfg.WhisperProvider))
		protected.DELETE("/candidates/:id", handlers.DeleteCandidate(pool))
		protected.PATCH("/candidates/:id/status", handlers.UpdateCandidateStatus(pool))
		protected.POST("/candidates/:id/fetch-transcript", handlers.FetchTranscriptManually(pool, cfg.WhisperAPIKey, cfg.WhisperProvider))

		// Analysis
		protected.GET("/candidates/:id/analysis", handlers.GetAnalysis(pool))
		protected.GET("/candidates/:id/analysis-status", handlers.GetCandidateAnalysisStatus())
		protected.DELETE("/candidates/:id/analysis", handlers.DeleteAnalysis(pool))
		protected.POST("/candidates/:id/analyze", middleware.TechAdminRestricted(), handlers.AnalyzeSingleCandidate(pool, aiProviders, defaultProvider))
		protected.DELETE("/analyses", handlers.DeleteAllAnalyses(pool))

		// Decisions (tech-admin cannot vote)
		protected.POST("/candidates/:id/decision", middleware.TechAdminRestricted(), handlers.MakeDecision(pool))
		protected.GET("/candidates/:id/decisions", handlers.GetDecisions(pool))

		// Comments (tech-admin cannot comment)
		protected.GET("/candidates/:id/comments", handlers.GetComments(pool))
		protected.POST("/candidates/:id/comments", middleware.TechAdminRestricted(), handlers.AddComment(pool))
		protected.DELETE("/comments/:commentId", handlers.DeleteComment(pool))

		// Stats / Export / Bulk
		protected.GET("/stats", handlers.GetDashboardStats(pool))
		protected.POST("/analyze-all", middleware.TechAdminRestricted(), handlers.AnalyzeAllPending(pool, aiProviders, batchProviders, defaultProvider))
		protected.POST("/reanalyze-all", middleware.TechAdminRestricted(), handlers.AnalyzeAllCandidates(pool, aiProviders, batchProviders, defaultProvider))
		protected.POST("/analyze-all/stop", handlers.StopBatch())
		protected.GET("/analyze-all/status", handlers.GetBatchStatus())
		protected.POST("/candidates/ai-recommend", handlers.RecommendCandidates(pool, textGens, defaultProvider))
		protected.GET("/ai-providers", handlers.GetAIProviders(aiProviders, defaultProvider))
		protected.GET("/candidates/export/csv", handlers.ExportCandidatesCSV(pool))
		protected.POST("/candidates/import/csv", handlers.ImportCandidatesCSV(pool))
		protected.POST("/candidates/bulk-decision", handlers.BulkDecision(pool))
		protected.POST("/candidates/auto-accept", handlers.AutoAcceptTopN(pool))
		protected.GET("/candidates/:id/similar", handlers.GetSimilarCandidates(pool))

		// Partner schools (auditor+ can view)
		protected.GET("/partner-schools", middleware.AuditorOrAbove(), handlers.GetPartnerSchools(pool))

		// Analytics distribution endpoints
		protected.GET("/analytics/city-distribution", handlers.GetCityDistribution(pool))
		protected.GET("/analytics/major-distribution", handlers.GetMajorDistribution(pool))

		// Auditor analytics — manager performance (auditor+)
		protected.GET("/auditor/manager-performance", middleware.AuditorOrAbove(), handlers.GetManagerPerformance(pool))
		protected.GET("/auditor/analysis-variance", handlers.GetAnalysisVariance(pool))

		// AI Assistant (manager gets full data context; regular users get FAQ mode)
		protected.POST("/ai/assistant", handlers.AssistantChat(pool, textGens, defaultProvider))

		// Committee War Room
		protected.GET("/war-room/feed", handlers.GetActivityFeed(pool))
		protected.POST("/war-room/feed", middleware.TechAdminRestricted(), handlers.PostActivityFeed(pool))
		protected.GET("/war-room/discussion", handlers.GetDiscussionCandidates(pool))
		protected.POST("/candidates/:id/discuss", handlers.MarkForDiscussion(pool))
		protected.GET("/notifications", handlers.GetNotifications(pool))
		protected.POST("/notifications/read", handlers.MarkNotificationsRead(pool))

		// User management — auditor can read, tech-admin+ can modify
		protected.GET("/users", middleware.AuditorOrAbove(), handlers.ListUsers(pool))
		protected.GET("/users/:id", middleware.AuditorOrAbove(), handlers.GetUser(pool))
		protected.PATCH("/users/:id", middleware.TechAdminOrAbove(), handlers.UpdateUser(pool))
		protected.DELETE("/users/:id", middleware.SuperAdminOnly(), handlers.DeleteUser(pool))
		protected.POST("/users/:id/reset-password", middleware.SuperAdminOnly(), handlers.ResetUserPassword(pool, emailSvc))

		// Private Notes
		protected.GET("/candidates/:id/notes", handlers.GetPrivateNote(pool))
		protected.PUT("/candidates/:id/notes", handlers.SavePrivateNote(pool))
		protected.DELETE("/candidates/:id/notes", handlers.DeletePrivateNote(pool))

		// Review Tasks
		protected.GET("/tasks", handlers.GetTasks(pool))
		protected.POST("/tasks/assign", handlers.AssignNewTasks(pool))
		protected.PATCH("/tasks/:taskId", handlers.UpdateTaskStatus(pool))

		// Analysis History
		protected.GET("/candidates/:id/analysis-history", handlers.GetAnalysisHistory(pool))

		// Recalculate review complexity for all candidates
		protected.POST("/candidates/recalc-complexity", handlers.RecalcComplexity(pool))

		// Profile
		protected.GET("/profile", handlers.GetProfile(pool))
		protected.PATCH("/profile", handlers.UpdateProfile(pool))

		// Email (manual trigger) — manager+
		protected.POST("/candidates/:id/email/shortlist", func(c *gin.Context) {
			id := c.Param("id")
			var row struct {
				Email    string
				FullName string
			}
			if err := pool.QueryRow(c.Request.Context(),
				`SELECT email, full_name FROM candidates WHERE id = $1`, id).
				Scan(&row.Email, &row.FullName); err != nil {
				c.JSON(404, gin.H{"error": "candidate not found"})
				return
			}
			if emailSvc.Enabled() {
				idInt := 0
				fmt.Sscan(id, &idInt)
				go emailSvc.SendShortlistNotification(row.Email, row.FullName, idInt)
				c.JSON(200, gin.H{"message": "email queued"})
			} else {
				c.JSON(503, gin.H{"error": "email service not configured"})
			}
		})
	}

	// Telegram bot (Stage 2 Interview)
	botUsername := ""
	if cfg.TelegramBotToken != "" {
		var botTextGen telegram_bot.TextGenerator
		botModelName := "unknown"
		if gen, ok := textGens["gemini"]; ok {
			botTextGen = gen
			botModelName = gemini.ModelName
		} else if gen, ok := textGens["ollama"]; ok {
			botTextGen = gen
			botModelName = "ollama/" + cfg.OllamaModel
		}

		if botTextGen != nil {
			bot, err := telegram_bot.New(cfg, pool, botTextGen, botModelName)
			if err != nil {
				log.Printf("Warning: failed to init Telegram bot: %v", err)
			} else {
				botUsername = bot.Username()
				go bot.Start(context.Background())
				log.Printf("Telegram bot started (@%s)", botUsername)
			}
		}
	}

	// Interview routes
	protected.POST("/candidates/:id/telegram-invite", handlers.CreateTelegramInvite(pool, botUsername))
	protected.GET("/candidates/:id/interview", handlers.GetInterviewStatus(pool, botUsername))
	protected.GET("/candidates/:id/interview/messages", handlers.GetInterviewTranscript(pool))

	// Force evaluate / re-evaluate interview
	var evaluateFn func(interviewID, candidateID int) error
	if cfg.TelegramBotToken != "" {
		var evalTextGen telegram_bot.TextGenerator
		evalModelName := "unknown"
		if gen, ok := textGens["gemini"]; ok {
			evalTextGen = gen
			evalModelName = gemini.ModelName
		} else if gen, ok := textGens["ollama"]; ok {
			evalTextGen = gen
			evalModelName = "ollama/" + cfg.OllamaModel
		}
		if evalTextGen != nil {
			evaluator := telegram_bot.NewEvaluator(pool, evalTextGen, evalModelName)
			evaluateFn = func(interviewID, candidateID int) error {
				return evaluator.EvaluateFromDB(context.Background(), interviewID, candidateID)
			}
		}
	}
	if evaluateFn != nil {
		protected.POST("/candidates/:id/interview/evaluate", handlers.ForceEvaluateInterview(pool, evaluateFn))
		protected.POST("/candidates/:id/interview/re-evaluate", handlers.ReEvaluateInterview(pool, evaluateFn))
		protected.DELETE("/candidates/:id/interview/analysis", handlers.DeleteInterviewAnalysis(pool))
		protected.POST("/interviews/evaluate-all-pending", handlers.EvaluateAllPendingInterviews(pool, evaluateFn))
	}

	log.Printf("Server starting on port %s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func isOllamaAvailable(baseURL string) bool {
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(baseURL + "/api/version")
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == 200
}
