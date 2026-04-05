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

		// AI-recommended major (may differ from candidate's chosen major)
		`ALTER TABLE analyses ADD COLUMN IF NOT EXISTS recommended_major VARCHAR(100)`,
		`ALTER TABLE analyses ADD COLUMN IF NOT EXISTS major_reason_note TEXT`,

		// Upvote/downvote system: extend decision CHECK constraint
		`DO $$ BEGIN
			ALTER TABLE committee_decisions DROP CONSTRAINT IF EXISTS committee_decisions_decision_check;
			ALTER TABLE committee_decisions ADD CONSTRAINT committee_decisions_decision_check
				CHECK (decision IN ('shortlist','reject','waitlist','review','upvote','downvote'));
		EXCEPTION WHEN OTHERS THEN NULL;
		END $$`,

		// War Room: activity feed
		`CREATE TABLE IF NOT EXISTS activity_feed (
			id SERIAL PRIMARY KEY,
			user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
			candidate_id INTEGER REFERENCES candidates(id) ON DELETE CASCADE,
			action_type VARCHAR(50) NOT NULL,
			content TEXT,
			mentions JSONB DEFAULT '[]',
			created_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_activity_feed_created ON activity_feed(created_at DESC)`,
		`CREATE INDEX IF NOT EXISTS idx_activity_feed_candidate ON activity_feed(candidate_id)`,

		// War Room: notifications/mentions
		`CREATE TABLE IF NOT EXISTS notifications (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			from_user_id INTEGER REFERENCES users(id) ON DELETE SET NULL,
			candidate_id INTEGER REFERENCES candidates(id) ON DELETE CASCADE,
			activity_id INTEGER REFERENCES activity_feed(id) ON DELETE CASCADE,
			type VARCHAR(30) NOT NULL DEFAULT 'mention',
			read BOOLEAN NOT NULL DEFAULT false,
			created_at TIMESTAMP DEFAULT NOW()
		)`,
		`CREATE INDEX IF NOT EXISTS idx_notifications_user ON notifications(user_id, read)`,

		// War Room: "Needs Discussion" flag on candidates
		`ALTER TABLE candidates ADD COLUMN IF NOT EXISTS needs_discussion BOOLEAN DEFAULT false`,
		`ALTER TABLE candidates ADD COLUMN IF NOT EXISTS discussion_note TEXT`,
		`ALTER TABLE candidates ADD COLUMN IF NOT EXISTS discussion_by INTEGER REFERENCES users(id)`,
		`ALTER TABLE candidates ADD COLUMN IF NOT EXISTS discussion_at TIMESTAMP`,

		// ─── New fields: personal info expansion ──────────────────────────────────
		`ALTER TABLE candidates ADD COLUMN IF NOT EXISTS first_name VARCHAR(255)`,
		`ALTER TABLE candidates ADD COLUMN IF NOT EXISTS last_name VARCHAR(255)`,
		`ALTER TABLE candidates ADD COLUMN IF NOT EXISTS patronymic VARCHAR(255)`,
		`ALTER TABLE candidates ADD COLUMN IF NOT EXISTS date_of_birth DATE`,
		`ALTER TABLE candidates ADD COLUMN IF NOT EXISTS gender VARCHAR(20)`,
		`ALTER TABLE candidates ADD COLUMN IF NOT EXISTS nationality VARCHAR(100)`,
		`ALTER TABLE candidates ADD COLUMN IF NOT EXISTS iin VARCHAR(12)`,
		`ALTER TABLE candidates ADD COLUMN IF NOT EXISTS identity_doc_type VARCHAR(50)`,
		`ALTER TABLE candidates ADD COLUMN IF NOT EXISTS instagram VARCHAR(255)`,
		`ALTER TABLE candidates ADD COLUMN IF NOT EXISTS whatsapp VARCHAR(50)`,
		`ALTER TABLE candidates ADD COLUMN IF NOT EXISTS home_country VARCHAR(100)`,

		// ─── English proficiency ──────────────────────────────────────────────────
		`ALTER TABLE candidates ADD COLUMN IF NOT EXISTS exam_type VARCHAR(10)`,
		`ALTER TABLE candidates ADD COLUMN IF NOT EXISTS ielts_score NUMERIC(3,1)`,
		`ALTER TABLE candidates ADD COLUMN IF NOT EXISTS toefl_score INTEGER`,
		`ALTER TABLE candidates ADD COLUMN IF NOT EXISTS english_cert_url VARCHAR(512)`,

		// ─── Certificate ──────────────────────────────────────────────────────────
		`ALTER TABLE candidates ADD COLUMN IF NOT EXISTS certificate_type VARCHAR(50)`,
		`ALTER TABLE candidates ADD COLUMN IF NOT EXISTS certificate_url VARCHAR(512)`,

		// ─── Additional documents ─────────────────────────────────────────────────
		`ALTER TABLE candidates ADD COLUMN IF NOT EXISTS additional_docs_url VARCHAR(512)`,

		// ─── Personality test ─────────────────────────────────────────────────────
		`ALTER TABLE candidates ADD COLUMN IF NOT EXISTS personality_answers JSONB DEFAULT '[]'`,

		// ─── Review complexity scoring ────────────────────────────────────────────
		`ALTER TABLE candidates ADD COLUMN IF NOT EXISTS review_complexity NUMERIC(5,2)`,

		// ─── Private notes (per-user per-candidate) ──────────────────────────────
		`CREATE TABLE IF NOT EXISTS private_notes (
			id SERIAL PRIMARY KEY,
			user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
			candidate_id INTEGER NOT NULL REFERENCES candidates(id) ON DELETE CASCADE,
			content TEXT NOT NULL DEFAULT '',
			updated_at TIMESTAMP DEFAULT NOW(),
			UNIQUE(user_id, candidate_id)
		)`,
		`CREATE INDEX IF NOT EXISTS idx_private_notes_user_candidate ON private_notes(user_id, candidate_id)`,

		// ─── Tasks (round-robin assignment) ───────────────────────────────────────
		`CREATE TABLE IF NOT EXISTS review_tasks (
			id SERIAL PRIMARY KEY,
			candidate_id INTEGER UNIQUE NOT NULL REFERENCES candidates(id) ON DELETE CASCADE,
			assigned_to INTEGER REFERENCES users(id) ON DELETE SET NULL,
			status VARCHAR(20) NOT NULL DEFAULT 'pending' CHECK (status IN ('pending','in_progress','completed')),
			created_at TIMESTAMP DEFAULT NOW(),
			completed_at TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_review_tasks_assigned ON review_tasks(assigned_to, status)`,

		// ─── Analysis history (keep old analyses instead of overwriting) ──────────
		`CREATE TABLE IF NOT EXISTS analysis_history (
			id SERIAL PRIMARY KEY,
			candidate_id INTEGER NOT NULL REFERENCES candidates(id) ON DELETE CASCADE,
			score_leadership INTEGER NOT NULL,
			score_motivation INTEGER NOT NULL,
			score_growth INTEGER NOT NULL,
			score_vision INTEGER NOT NULL,
			score_communication INTEGER NOT NULL,
			final_score NUMERIC(5,2) NOT NULL,
			category VARCHAR(30) NOT NULL,
			ai_generated_risk VARCHAR(10) DEFAULT 'low',
			ai_generated_score INTEGER DEFAULT 0,
			summary TEXT,
			key_strengths TEXT[],
			red_flags TEXT[],
			model_used VARCHAR(50),
			analyzed_at TIMESTAMP DEFAULT NOW(),
			duration_ms INTEGER DEFAULT 0
		)`,
		`CREATE INDEX IF NOT EXISTS idx_analysis_history_candidate ON analysis_history(candidate_id, analyzed_at DESC)`,

		// ─── Update status constraint for new stages ──────────────────────────────
		`DO $$ BEGIN
			ALTER TABLE candidates DROP CONSTRAINT IF EXISTS candidates_status_check;
			ALTER TABLE candidates ADD CONSTRAINT candidates_status_check
				CHECK (status IN ('pending','in_progress','initial_screening','application_review','interview_stage','committee_review','decision','analyzed','shortlisted','rejected','waitlisted'));
		EXCEPTION WHEN OTHERS THEN NULL;
		END $$`,

		// ─── UNT / NIS score fields ──────────────────────────────────────────────
		`ALTER TABLE candidates ADD COLUMN IF NOT EXISTS unt_score INTEGER`,
		`ALTER TABLE candidates ADD COLUMN IF NOT EXISTS nis_grade VARCHAR(5)`,

		// ─── Partner school tag ──────────────────────────────────────────────────
		`ALTER TABLE candidates ADD COLUMN IF NOT EXISTS partner_school VARCHAR(255)`,

		// ─── Partner schools table ───────────────────────────────────────────────
		`CREATE TABLE IF NOT EXISTS partner_schools (
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL,
			city VARCHAR(100),
			contact_email VARCHAR(255),
			contact_phone VARCHAR(50),
			graduates_per_year INTEGER DEFAULT 0,
			created_at TIMESTAMP DEFAULT NOW()
		)`,

		// ─── Analysis history: add explanation columns ────────────────────────────
		`ALTER TABLE analysis_history ADD COLUMN IF NOT EXISTS explanation_leadership TEXT`,
		`ALTER TABLE analysis_history ADD COLUMN IF NOT EXISTS explanation_motivation TEXT`,
		`ALTER TABLE analysis_history ADD COLUMN IF NOT EXISTS explanation_growth TEXT`,
		`ALTER TABLE analysis_history ADD COLUMN IF NOT EXISTS explanation_vision TEXT`,
		`ALTER TABLE analysis_history ADD COLUMN IF NOT EXISTS explanation_communication TEXT`,
		`ALTER TABLE analysis_history ADD COLUMN IF NOT EXISTS incomplete_flag BOOLEAN DEFAULT false`,
		`ALTER TABLE analysis_history ADD COLUMN IF NOT EXISTS recommended_major VARCHAR(100)`,
		`ALTER TABLE analysis_history ADD COLUMN IF NOT EXISTS major_reason_note TEXT`,

		// ─── Notes: add created_at ────────────────────────────────────────────────
		`ALTER TABLE private_notes ADD COLUMN IF NOT EXISTS created_at TIMESTAMP DEFAULT NOW()`,
	}

	for _, q := range queries {
		if _, err := pool.Exec(ctx, q); err != nil {
			return fmt.Errorf("migration failed [%.80s]: %w", q, err)
		}
	}

	log.Println("Database migrations completed successfully")
	return nil
}
