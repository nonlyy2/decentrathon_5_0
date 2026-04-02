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
		// ─── Core tables ──────────────────────────────────────────────────────────
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			email VARCHAR(255) UNIQUE NOT NULL,
			password_hash VARCHAR(255) NOT NULL,
			full_name VARCHAR(255),
			role VARCHAR(20) NOT NULL DEFAULT 'manager' CHECK (role IN ('superadmin','tech-admin','auditor','manager','admin','committee')),
			avatar_url VARCHAR(512),
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

		// ─── Stage 2: Telegram interview ──────────────────────────────────────────
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

		// ─── Candidate info-change requests ────────────────────────────────────────
		`CREATE TABLE IF NOT EXISTS candidate_change_requests (
			id SERIAL PRIMARY KEY,
			candidate_id INTEGER NOT NULL REFERENCES candidates(id) ON DELETE CASCADE,
			field_name VARCHAR(100) NOT NULL,
			old_value TEXT,
			new_value TEXT NOT NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','approved','rejected')),
			requested_at TIMESTAMP DEFAULT NOW(),
			reviewed_by INTEGER REFERENCES users(id),
			reviewed_at TIMESTAMP,
			review_note TEXT
		)`,

		// ─── Email log ────────────────────────────────────────────────────────────
		`CREATE TABLE IF NOT EXISTS email_logs (
			id SERIAL PRIMARY KEY,
			candidate_id INTEGER REFERENCES candidates(id) ON DELETE SET NULL,
			recipient_email VARCHAR(255) NOT NULL,
			subject VARCHAR(512) NOT NULL,
			template VARCHAR(50) NOT NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'sent' CHECK (status IN ('sent','failed')),
			sent_at TIMESTAMP DEFAULT NOW()
		)`,

		// ─── Indexes ──────────────────────────────────────────────────────────────
		`CREATE INDEX IF NOT EXISTS idx_comments_candidate ON comments(candidate_id)`,
		`CREATE INDEX IF NOT EXISTS idx_candidates_status ON candidates(status)`,
		`CREATE INDEX IF NOT EXISTS idx_analyses_candidate ON analyses(candidate_id)`,
		`CREATE INDEX IF NOT EXISTS idx_analyses_final_score ON analyses(final_score DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_decisions_candidate ON committee_decisions(candidate_id)`,
		`CREATE INDEX IF NOT EXISTS idx_change_requests_candidate ON candidate_change_requests(candidate_id)`,
		`CREATE INDEX IF NOT EXISTS idx_change_requests_status ON candidate_change_requests(status)`,
		`CREATE INDEX IF NOT EXISTS idx_email_logs_candidate ON email_logs(candidate_id)`,

		// ─── ALTER TABLE additions (idempotent) ───────────────────────────────────
		`ALTER TABLE candidates ADD COLUMN IF NOT EXISTS phone VARCHAR(50)`,
		`ALTER TABLE candidates ADD COLUMN IF NOT EXISTS telegram VARCHAR(100)`,
		`ALTER TABLE analyses ADD COLUMN IF NOT EXISTS ai_generated_score INTEGER DEFAULT 0`,
		`ALTER TABLE candidates ADD COLUMN IF NOT EXISTS disability TEXT`,

		// Photo + AI detection
		`ALTER TABLE candidates ADD COLUMN IF NOT EXISTS photo_url VARCHAR(512)`,
		`ALTER TABLE candidates ADD COLUMN IF NOT EXISTS photo_ai_flag BOOLEAN DEFAULT false`,
		`ALTER TABLE candidates ADD COLUMN IF NOT EXISTS photo_ai_note TEXT`,

		// Major selection
		`ALTER TABLE candidates ADD COLUMN IF NOT EXISTS major VARCHAR(100)`,

		// Keywords (extracted by AI from essay + achievements for fast search)
		`ALTER TABLE candidates ADD COLUMN IF NOT EXISTS keywords TEXT[]`,

		// Stage 2 interview fields
		`ALTER TABLE candidates ADD COLUMN IF NOT EXISTS telegram_chat_id BIGINT`,
		`ALTER TABLE candidates ADD COLUMN IF NOT EXISTS interview_status VARCHAR(20) DEFAULT 'not_invited' CHECK (interview_status IN ('not_invited','invited','in_progress','completed'))`,
		`ALTER TABLE candidates ADD COLUMN IF NOT EXISTS combined_score NUMERIC(5,2)`,

		// users: role constraint updated (drop old check, add new)
		`DO $$ BEGIN
			ALTER TABLE users DROP CONSTRAINT IF EXISTS users_role_check;
			ALTER TABLE users ADD CONSTRAINT users_role_check CHECK (role IN ('superadmin','tech-admin','auditor','manager','admin','committee'));
		EXCEPTION WHEN OTHERS THEN NULL;
		END $$`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS full_name VARCHAR(255)`,
		`ALTER TABLE users ADD COLUMN IF NOT EXISTS avatar_url VARCHAR(512)`,

		// Review system: unique constraint so a user can only vote once per candidate
		`CREATE UNIQUE INDEX IF NOT EXISTS idx_decisions_user_candidate ON committee_decisions(candidate_id, decided_by)`,

		// Telegram invite indexes
		`CREATE INDEX IF NOT EXISTS idx_telegram_invites_token ON telegram_invites(token)`,
		`CREATE INDEX IF NOT EXISTS idx_telegram_invites_chat ON telegram_invites(telegram_chat_id)`,
		`CREATE INDEX IF NOT EXISTS idx_interviews_chat ON interviews(telegram_chat_id)`,
		`CREATE INDEX IF NOT EXISTS idx_interviews_candidate ON interviews(candidate_id)`,
		`CREATE INDEX IF NOT EXISTS idx_interview_messages_interview ON interview_messages(interview_id)`,

		// GIN index for keyword search
		`CREATE INDEX IF NOT EXISTS idx_candidates_keywords ON candidates USING GIN(keywords)`,

		// GIN index for full-text search on essay+achievements
		`CREATE INDEX IF NOT EXISTS idx_candidates_fts ON candidates USING GIN(to_tsvector('english', coalesce(essay,'') || ' ' || coalesce(achievements,'')))`,

		// Analysis duration tracking
		`ALTER TABLE analyses ADD COLUMN IF NOT EXISTS duration_ms INTEGER DEFAULT 0`,

		// Interview analysis explanations
		`ALTER TABLE interview_analyses ADD COLUMN IF NOT EXISTS explanation_leadership TEXT`,
		`ALTER TABLE interview_analyses ADD COLUMN IF NOT EXISTS explanation_grit TEXT`,
		`ALTER TABLE interview_analyses ADD COLUMN IF NOT EXISTS explanation_authenticity TEXT`,
		`ALTER TABLE interview_analyses ADD COLUMN IF NOT EXISTS explanation_motivation TEXT`,
		`ALTER TABLE interview_analyses ADD COLUMN IF NOT EXISTS explanation_vision TEXT`,

		// YouTube video presentation
		`ALTER TABLE candidates ADD COLUMN IF NOT EXISTS youtube_url TEXT`,
		`ALTER TABLE candidates ADD COLUMN IF NOT EXISTS youtube_transcript TEXT`,
		`ALTER TABLE candidates ADD COLUMN IF NOT EXISTS youtube_url_valid BOOLEAN`,
	}

	for _, q := range queries {
		if _, err := pool.Exec(ctx, q); err != nil {
			return fmt.Errorf("migration failed [%.80s]: %w", q, err)
		}
	}

	log.Println("Database migrations completed successfully")
	return nil
}
