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
	}

	for _, q := range queries {
		if _, err := pool.Exec(ctx, q); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	log.Println("Database migrations completed successfully")
	return nil
}
