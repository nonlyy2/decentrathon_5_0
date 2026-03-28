package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port         string
	DatabaseURL  string
	JWTSecret    string
	GeminiAPIKey string
	AllowOrigins string
	AIProvider   string // "gemini" or "ollama"
	OllamaURL    string
	OllamaModel  string
}

func Load() *Config {
	godotenv.Load()
	return &Config{
		Port:         getEnv("PORT", "8080"),
		DatabaseURL:  getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/invisionu?sslmode=disable"),
		JWTSecret:    getEnv("JWT_SECRET", "dev-secret-change-in-prod"),
		GeminiAPIKey: getEnv("GEMINI_API_KEY", ""),
		AllowOrigins: getEnv("ALLOW_ORIGINS", "http://localhost:3000"),
		AIProvider:   getEnv("AI_PROVIDER", "gemini"),
		OllamaURL:    getEnv("OLLAMA_URL", "http://localhost:11434"),
		OllamaModel:  getEnv("OLLAMA_MODEL", "llama3.1:8b"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
