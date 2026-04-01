package config

import (
	"os"
	"strconv"

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

	// Telegram bot
	TelegramBotToken string
	WhisperAPIKey    string
	WhisperProvider  string

	// Interview settings
	InterviewTimeoutMin   int
	InterviewMinQuestions int
	InterviewMaxQuestions int

	// Email (SMTP)
	SMTPHost     string
	SMTPPort     int
	SMTPUser     string
	SMTPPassword string
	SMTPFrom     string

	// File uploads
	UploadDir string
}

func Load() *Config {
	godotenv.Load()
	return &Config{
		Port: getEnv("PORT", "8080"),
		// Railway injects DATABASE_URL automatically when a Postgres service is linked.
		// Falls back to POSTGRES_URL / DATABASE_PUBLIC_URL for other Railway configurations.
		DatabaseURL:  getEnvMulti("DATABASE_URL", "POSTGRES_URL", "DATABASE_PUBLIC_URL", "postgres://postgres:postgres@localhost:5432/invisionu?sslmode=disable"),
		JWTSecret:    getEnv("JWT_SECRET", "dev-secret-change-in-prod"),
		GeminiAPIKey: getEnv("GEMINI_API_KEY", ""),
		AllowOrigins: getEnv("ALLOW_ORIGINS", "*"),
		AIProvider:   getEnv("AI_PROVIDER", "gemini"),
		OllamaURL:    getEnv("OLLAMA_URL", "http://localhost:11434"),
		OllamaModel:  getEnv("OLLAMA_MODEL", "mistral:7b"),

		TelegramBotToken: getEnv("TELEGRAM_BOT_TOKEN", ""),
		WhisperAPIKey:    getEnvMulti("ALEM_STT_API_KEY", "WHISPER_API_KEY", ""),
		WhisperProvider:  getEnv("WHISPER_PROVIDER", "openai"),

		InterviewTimeoutMin:   getEnvInt("INTERVIEW_TIMEOUT_MIN", 30),
		InterviewMinQuestions: getEnvInt("INTERVIEW_MIN_QUESTIONS", 8),
		InterviewMaxQuestions: getEnvInt("INTERVIEW_MAX_QUESTIONS", 15),

		SMTPHost:     getEnv("SMTP_HOST", ""),
		SMTPPort:     getEnvInt("SMTP_PORT", 587),
		SMTPUser:     getEnv("SMTP_USER", ""),
		SMTPPassword: getEnvMulti("SMTP_PASSWORD", "SMTP_PASS", ""),
		SMTPFrom:     getEnv("SMTP_FROM", "noreply@invisionu.kz"),

		UploadDir: getEnv("UPLOAD_DIR", "./uploads"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvMulti(keys ...string) string {
	fallback := keys[len(keys)-1]
	for _, k := range keys[:len(keys)-1] {
		if v := os.Getenv(k); v != "" {
			return v
		}
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}
