package models

import "time"

type Candidate struct {
	ID                  int       `json:"id"`
	FullName            string    `json:"full_name"`
	Email               string    `json:"email"`
	Phone               *string   `json:"phone"`
	Telegram            *string   `json:"telegram"`
	Age                 *int      `json:"age"`
	City                *string   `json:"city"`
	School              *string   `json:"school"`
	GraduationYear      *int      `json:"graduation_year"`
	Achievements        *string   `json:"achievements"`
	Extracurriculars    *string   `json:"extracurriculars"`
	Essay               string    `json:"essay"`
	MotivationStatement *string   `json:"motivation_statement"`
	Disability          *string   `json:"disability"`
	CreatedAt           time.Time `json:"created_at"`
	Status              string    `json:"status"`
}

type CandidateListItem struct {
	ID         int        `json:"id"`
	FullName   string     `json:"full_name"`
	Email      string     `json:"email"`
	City       *string    `json:"city"`
	School     *string    `json:"school"`
	Status     string     `json:"status"`
	CreatedAt  time.Time  `json:"created_at"`
	FinalScore *float64   `json:"final_score"`
	Category   *string    `json:"category"`
	AnalyzedAt *time.Time `json:"analyzed_at"`
	ModelUsed  *string    `json:"model_used"`
}

type CandidateDetail struct {
	Candidate
	Analysis  *Analysis  `json:"analysis"`
	Decisions []Decision `json:"decisions"`
}

type CreateCandidateRequest struct {
	FullName            string  `json:"full_name" binding:"required"`
	Email               string  `json:"email" binding:"required,email"`
	Phone               string  `json:"phone" binding:"required"`
	Telegram            string  `json:"telegram" binding:"required"`
	Age                 int     `json:"age" binding:"required"`
	City                string  `json:"city" binding:"required"`
	School              string  `json:"school" binding:"required"`
	GraduationYear      int     `json:"graduation_year" binding:"required"`
	Achievements        string  `json:"achievements" binding:"required"`
	Extracurriculars    string  `json:"extracurriculars" binding:"required"`
	Essay               string  `json:"essay" binding:"required"`
	MotivationStatement string  `json:"motivation_statement" binding:"required"`
	Disability          *string `json:"disability"`
}
