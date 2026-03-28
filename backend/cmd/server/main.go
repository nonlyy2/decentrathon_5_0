package main

import (
	"context"
	"log"
	"os"

	"github.com/assylkhan/invisionu-backend/internal/config"
	"github.com/assylkhan/invisionu-backend/internal/database"
	"github.com/assylkhan/invisionu-backend/internal/gemini"
	"github.com/assylkhan/invisionu-backend/internal/handlers"
	"github.com/assylkhan/invisionu-backend/internal/models"
	"github.com/assylkhan/invisionu-backend/internal/middleware"
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

	// Gemini client
	var geminiClient *gemini.Client
	if cfg.GeminiAPIKey != "" {
		geminiClient = gemini.NewClient(cfg.GeminiAPIKey)
		log.Println("Gemini API client initialized")
	} else {
		log.Println("Warning: GEMINI_API_KEY not set, analysis endpoint will return 503")
	}

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
		var analyzeFunc func(ctx context.Context, candidate *models.Candidate) (*models.Analysis, error)
		if geminiClient != nil {
			analyzeFunc = geminiClient.AnalyzeCandidate
		}
		protected.POST("/candidates/:id/analyze", handlers.AnalyzeSingleCandidate(pool, analyzeFunc))

		protected.POST("/candidates/:id/decision", handlers.MakeDecision(pool))
		protected.GET("/candidates/:id/decisions", handlers.GetDecisions(pool))

		protected.GET("/stats", handlers.GetDashboardStats(pool))

		protected.POST("/analyze-all", handlers.AnalyzeAllPending(pool, analyzeFunc))
		protected.GET("/analyze-all/status", handlers.GetBatchStatus())
	}

	log.Printf("Server starting on port %s", cfg.Port)
	if err := router.Run(":" + cfg.Port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
