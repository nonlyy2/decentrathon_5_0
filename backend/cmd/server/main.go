package main

import (
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
	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	// Database
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

	// AI clients — create all available providers
	providers := make(handlers.AIProviders)
	batchProviders := make(handlers.AIBatchProviders)
	textGens := make(handlers.AITextGenerators)

	if cfg.GeminiAPIKey != "" {
		geminiClient := gemini.NewClient(cfg.GeminiAPIKey)
		providers["gemini"] = geminiClient.AnalyzeCandidate
		batchProviders["gemini"] = geminiClient.AnalyzeBatch
		textGens["gemini"] = geminiClient.GenerateText
		log.Printf("Gemini initialized (model=%s, tier-1, no rate limit)", gemini.ModelName)
	}

	ollamaClient := ollama.NewClient(cfg.OllamaURL, cfg.OllamaModel)
	providers["ollama"] = ollamaClient.AnalyzeCandidate
	batchProviders["ollama"] = ollamaClient.AnalyzeBatch
	textGens["ollama"] = ollamaClient.GenerateText
	if isOllamaAvailable(cfg.OllamaURL) {
		log.Printf("Ollama initialized (url=%s, model=%s)", cfg.OllamaURL, cfg.OllamaModel)
	} else {
		log.Printf("Ollama registered but not reachable at %s — will show in UI, analysis will fail if not running", cfg.OllamaURL)
	}

	defaultProvider := cfg.AIProvider
	if _, ok := providers[defaultProvider]; !ok {
		// fallback to whatever is available
		for k := range providers {
			defaultProvider = k
			break
		}
	}
	log.Printf("Default AI provider: %s", defaultProvider)

	// Router
	router := gin.Default()
	router.Use(middleware.CORSMiddleware(cfg.AllowOrigins))
	router.Use(middleware.NoCacheMiddleware())

	api := router.Group("/api")

	// Public routes
	api.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
	api.POST("/apply", handlers.SubmitApplication(pool))
	api.POST("/auth/register", handlers.Register(pool))
	api.POST("/auth/login", handlers.Login(pool, cfg.JWTSecret))

	// Protected routes
	protected := api.Group("/", middleware.AuthRequired(cfg.JWTSecret))
	{
		protected.GET("/candidates", handlers.ListCandidates(pool))
		protected.POST("/candidates", handlers.CreateCandidate(pool))
		protected.GET("/candidates/:id", handlers.GetCandidate(pool))
		protected.PATCH("/candidates/:id/status", handlers.UpdateCandidateStatus(pool))

		protected.GET("/candidates/:id/analysis", handlers.GetAnalysis(pool))
		protected.GET("/candidates/:id/analysis-status", handlers.GetCandidateAnalysisStatus())
		protected.DELETE("/candidates/:id/analysis", handlers.DeleteAnalysis(pool))
		protected.POST("/candidates/:id/analyze", handlers.AnalyzeSingleCandidate(pool, providers, defaultProvider))
		protected.DELETE("/analyses", handlers.DeleteAllAnalyses(pool))

		protected.POST("/candidates/:id/decision", handlers.MakeDecision(pool))
		protected.GET("/candidates/:id/decisions", handlers.GetDecisions(pool))

		protected.GET("/candidates/:id/comments", handlers.GetComments(pool))
		protected.POST("/candidates/:id/comments", handlers.AddComment(pool))
		protected.DELETE("/comments/:commentId", handlers.DeleteComment(pool))

		protected.GET("/stats", handlers.GetDashboardStats(pool))

		protected.POST("/analyze-all", handlers.AnalyzeAllPending(pool, providers, batchProviders, defaultProvider))
		protected.POST("/analyze-all/stop", handlers.StopBatch())
		protected.GET("/analyze-all/status", handlers.GetBatchStatus())
		protected.POST("/candidates/ai-recommend", handlers.RecommendCandidates(pool, textGens, defaultProvider))

		protected.GET("/ai-providers", handlers.GetAIProviders(providers, defaultProvider))
		protected.GET("/candidates/export/csv", handlers.ExportCandidatesCSV(pool))
		protected.POST("/candidates/bulk-decision", handlers.BulkDecision(pool))
	}

	log.Printf("Server starting on port %s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// isOllamaAvailable checks if the Ollama service is reachable
func isOllamaAvailable(baseURL string) bool {
	client := &http.Client{Timeout: 3 * time.Second}
	resp, err := client.Get(baseURL + "/api/version")
	if err != nil {
		return false
	}
	resp.Body.Close()
	return resp.StatusCode == 200
}
