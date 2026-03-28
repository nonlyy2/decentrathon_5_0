package main

import (
	"log"
	"os"

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

	// Seed candidates if --seed flag
	for _, arg := range os.Args[1:] {
		if arg == "--seed" {
			if err := seed.SeedCandidates(pool); err != nil {
				log.Printf("Warning: failed to seed candidates: %v", err)
			}
			break
		}
	}

	// AI clients — create all available providers
	providers := make(handlers.AIProviders)

	if cfg.GeminiAPIKey != "" {
		geminiClient := gemini.NewClient(cfg.GeminiAPIKey)
		providers["gemini"] = geminiClient.AnalyzeCandidate
		log.Println("Gemini API client initialized")
	}

	ollamaClient := ollama.NewClient(cfg.OllamaURL, cfg.OllamaModel)
	providers["ollama"] = ollamaClient.AnalyzeCandidate
	log.Printf("Ollama client initialized (url=%s, model=%s)", cfg.OllamaURL, cfg.OllamaModel)

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

		protected.GET("/stats", handlers.GetDashboardStats(pool))

		protected.POST("/analyze-all", handlers.AnalyzeAllPending(pool, providers, defaultProvider))
		protected.GET("/analyze-all/status", handlers.GetBatchStatus())

		protected.GET("/ai-providers", handlers.GetAIProviders(providers, defaultProvider))
	}

	log.Printf("Server starting on port %s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
