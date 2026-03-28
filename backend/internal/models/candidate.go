package models

import "time"

type Candidate struct {
	ID                  int       `json:"id"`
	FullName            string    `json:"full_name"`
	Email               string    `json:"email"`
	Age                 *int      `json:"age"`
	City                *string   `json:"city"`
	School              *string   `json:"school"`
	GraduationYear      *int      `json:"graduation_year"`
	Achievements        *string   `json:"achievements"`
	Extracurriculars    *string   `json:"extracurriculars"`
	Essay               string    `json:"essay"`
	MotivationStatement *string   `json:"motivation_statement"`
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
}

type CandidateDetail struct {
	Candidate
	Analysis  *Analysis  `json:"analysis"`
	Decisions []Decision `json:"decisions"`
}

type CreateCandidateRequest struct {
	FullName            string  `json:"full_name" binding:"required"`
	Email               string  `json:"email" binding:"required,email"`
	Age                 *int    `json:"age"`
	City                *string `json:"city"`
	School              *string `json:"school"`
	GraduationYear      *int    `json:"graduation_year"`
	Achievements        *string `json:"achievements"`
	Extracurriculars    *string `json:"extracurriculars"`
	Essay               string  `json:"essay" binding:"required"`
	MotivationStatement *string `json:"motivation_statement"`
}
