package database

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

func RunMigrations(pool *pgxpool.Pool) error {
	ctx := context.Background()

	queries := []string{
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			role VARCHAR(20) NOT NULL DEFAULT 'committee' CHECK (role IN ('admin', 'committee')),
			created_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS candidates (
			id SERIAL PRIMARY KEY,
			full_name VARCHAR(255) NOT NULL,
			email VARCHAR(255) UNIQUE NOT NULL,
			age INTEGER,
			city VARCHAR(100),
			school VARCHAR(255),
			graduation_year INTEGER,
			achievements TEXT,
			extracurriculars TEXT,
			essay TEXT NOT NULL,
			motivation_statement TEXT,
			created_at TIMESTAMP DEFAULT NOW(),
			status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','analyzed','shortlisted','rejected','waitlisted'))
		)`,
		`CREATE TABLE IF NOT EXISTS analyses (
			id SERIAL PRIMARY KEY,
			candidate_id INTEGER UNIQUE NOT NULL REFERENCES candidates(id) ON DELETE CASCADE,
			score_leadership INTEGER NOT NULL CHECK (score_leadership BETWEEN 0 AND 100),
			score_motivation INTEGER NOT NULL CHECK (score_motivation BETWEEN 0 AND 100),
			score_growth INTEGER NOT NULL CHECK (score_growth BETWEEN 0 AND 100),
			score_vision INTEGER NOT NULL CHECK (score_vision BETWEEN 0 AND 100),
			score_communication INTEGER NOT NULL CHECK (score_communication BETWEEN 0 AND 100),
			final_score NUMERIC(5,2) NOT NULL,
			category VARCHAR(30) NOT NULL,
			ai_generated_risk VARCHAR(10) NOT NULL DEFAULT 'low' CHECK (ai_generated_risk IN ('low','medium','high')),
			incomplete_flag BOOLEAN NOT NULL DEFAULT false,
			explanation_leadership TEXT,
			explanation_motivation TEXT,
			explanation_growth TEXT,
			explanation_vision TEXT,
			explanation_communication TEXT,
			summary TEXT,
			key_strengths TEXT[],
			red_flags TEXT[],
			analyzed_at TIMESTAMP DEFAULT NOW(),
			model_used VARCHAR(50)
		)`,
		`CREATE TABLE IF NOT EXISTS committee_decisions (
			id SERIAL PRIMARY KEY,
			candidate_id INTEGER NOT NULL REFERENCES candidates(id) ON DELETE CASCADE,
			decision VARCHAR(20) NOT NULL CHECK (decision IN ('shortlist','reject','waitlist','review')),
			notes TEXT,
			decided_by INTEGER REFERENCES users(id),
			decided_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS comments (
			id SERIAL PRIMARY KEY,
			candidate_id INTEGER NOT NULL REFERENCES candidates(id) ON DELETE CASCADE,
			user_id INTEGER REFERENCES users(id),
			content TEXT NOT NULL,
			created_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_comments_candidate ON comments(candidate_id)`,
		`CREATE INDEX IF NOT EXISTS idx_candidates_status ON candidates(status)`,
		`CREATE INDEX IF NOT EXISTS idx_analyses_candidate ON analyses(candidate_id)`,
		`CREATE INDEX IF NOT EXISTS idx_analyses_final_score ON analyses(final_score DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_decisions_candidate ON committee_decisions(candidate_id)`,
		// New columns
		`ALTER TABLE candidates ADD COLUMN IF NOT EXISTS phone VARCHAR(50)`,
		`ALTER TABLE candidates ADD COLUMN IF NOT EXISTS telegram VARCHAR(100)`,
		`ALTER TABLE analyses ADD COLUMN IF NOT EXISTS ai_generated_score INTEGER DEFAULT 0`,

		`ALTER TABLE candidates ADD COLUMN IF NOT EXISTS disability TEXT`,

		// Stage 2: Telegram interview tables
		`ALTER TABLE candidates ADD COLUMN IF NOT EXISTS telegram_chat_id BIGINT`,
		`ALTER TABLE candidates ADD COLUMN IF NOT EXISTS interview_status VARCHAR(20) DEFAULT 'not_invited' CHECK (interview_status IN ('not_invited','invited','in_progress','completed'))`,
		`ALTER TABLE candidates ADD COLUMN IF NOT EXISTS combined_score NUMERIC(5,2)`,

		`CREATE TABLE IF NOT EXISTS telegram_invites (
			id SERIAL PRIMARY KEY,
			candidate_id INTEGER UNIQUE NOT NULL REFERENCES candidates(id) ON DELETE CASCADE,
			token UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(),
			telegram_chat_id BIGINT,
			status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','linked','interview_active','completed','expired')),
			created_at TIMESTAMP DEFAULT NOW(),
			linked_at TIMESTAMP,
			expires_at TIMESTAMP DEFAULT NOW() + INTERVAL '7 days'
		)`,

		`CREATE TABLE IF NOT EXISTS interviews (
			id SERIAL PRIMARY KEY,
			candidate_id INTEGER UNIQUE NOT NULL REFERENCES candidates(id) ON DELETE CASCADE,
			telegram_chat_id BIGINT NOT NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'active' CHECK (status IN ('active','completed','abandoned','timeout')),
			language VARCHAR(5) DEFAULT 'en',
			current_topic VARCHAR(30) DEFAULT 'warmup',
			questions_asked INTEGER DEFAULT 0,
			started_at TIMESTAMP DEFAULT NOW(),
			completed_at TIMESTAMP,
			essay_summary TEXT,
			conversation_context JSONB DEFAULT '[]'
		)`,

		`CREATE TABLE IF NOT EXISTS interview_messages (
			id SERIAL PRIMARY KEY,
			interview_id INTEGER NOT NULL REFERENCES interviews(id) ON DELETE CASCADE,
			role VARCHAR(10) NOT NULL CHECK (role IN ('bot','candidate')),
			content TEXT NOT NULL,
			message_type VARCHAR(10) DEFAULT 'text' CHECK (message_type IN ('text','voice')),
			voice_duration_sec INTEGER,
			response_time_sec INTEGER,
			created_at TIMESTAMP DEFAULT NOW()
		)`,

		`CREATE TABLE IF NOT EXISTS interview_analyses (
			id SERIAL PRIMARY KEY,
			interview_id INTEGER NOT NULL REFERENCES interviews(id) ON DELETE CASCADE,
			candidate_id INTEGER UNIQUE NOT NULL REFERENCES candidates(id) ON DELETE CASCADE,
			score_leadership INTEGER CHECK (score_leadership BETWEEN 0 AND 100),
			score_grit INTEGER CHECK (score_grit BETWEEN 0 AND 100),
			score_authenticity INTEGER CHECK (score_authenticity BETWEEN 0 AND 100),
			score_motivation INTEGER CHECK (score_motivation BETWEEN 0 AND 100),
			score_vision INTEGER CHECK (score_vision BETWEEN 0 AND 100),
			final_score NUMERIC(5,2) NOT NULL,
			category VARCHAR(30) NOT NULL,
			consistency_score INTEGER DEFAULT 0,
			style_match_score INTEGER DEFAULT 0,
			suspicion_flags JSONB DEFAULT '[]',
			summary TEXT,
			strengths TEXT[],
			concerns TEXT[],
			analyzed_at TIMESTAMP DEFAULT NOW(),
			model_used VARCHAR(50)
		)`,

		// Review system: unique constraint so a user can only vote once per candidate
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_decisions_user_candidate ON committee_decisions(candidate_id, decided_by)`,

		`CREATE INDEX IF NOT EXISTS idx_telegram_invites_token ON telegram_invites(token)`,
		`CREATE INDEX IF NOT EXISTS idx_telegram_invites_chat ON telegram_invites(telegram_chat_id)`,
		`CREATE INDEX IF NOT EXISTS idx_interviews_chat ON interviews(telegram_chat_id)`,
		`CREATE INDEX IF NOT EXISTS idx_interviews_candidate ON interviews(candidate_id)`,
		`CREATE INDEX IF NOT EXISTS idx_interview_messages_interview ON interview_messages(interview_id)`,
	}

	for _, q := range queries {
		if _, err := pool.Exec(ctx, q); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	log.Println("Database migrations completed successfully")
	return nil
}
