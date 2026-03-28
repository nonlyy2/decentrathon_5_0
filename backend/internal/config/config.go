package config

import (
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port         string
	DatabaseURL  string
	JWTSecret    string
	GeminiAPIKey  string
	GeminiAPIKeys string // comma-separated list for round-robin
	AllowOrigins string
	AIProvider   string // "gemini", "ollama", or "groq"
	OllamaURL    string
	OllamaModel  string
	GroqAPIKey   string
	GroqModel    string
}

func Load() *Config {
	godotenv.Load()
	return &Config{
		Port:         getEnv("PORT", "8080"),
		DatabaseURL:  getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/invisionu?sslmode=disable"),
		JWTSecret:    getEnv("JWT_SECRET", "dev-secret-change-in-prod"),
		GeminiAPIKey:  getEnv("GEMINI_API_KEY", ""),
		GeminiAPIKeys: getEnv("GEMINI_API_KEYS", ""),
		AllowOrigins: getEnv("ALLOW_ORIGINS", "http://localhost:3000"),
		AIProvider:   getEnv("AI_PROVIDER", "gemini"),
		OllamaURL:    getEnv("OLLAMA_URL", "http://localhost:11434"),
		OllamaModel:  getEnv("OLLAMA_MODEL", "llama3.1:8b"),
		GroqAPIKey:   getEnv("GROQ_API_KEY", ""),
		GroqModel:    getEnv("GROQ_MODEL", "llama-3.3-70b-versatile"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
