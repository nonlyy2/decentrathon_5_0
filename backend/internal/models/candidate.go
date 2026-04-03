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
	Major               *string   `json:"major"`
	PhotoURL            *string   `json:"photo_url"`
	PhotoAIFlag         bool      `json:"photo_ai_flag"`
	PhotoAINote         *string   `json:"photo_ai_note"`
	Keywords            []string  `json:"keywords"`
	CreatedAt           time.Time `json:"created_at"`
	Status              string    `json:"status"`
	YouTubeURL          *string   `json:"youtube_url"`
	YouTubeTranscript   *string   `json:"youtube_transcript"`
	YouTubeURLValid     *bool     `json:"youtube_url_valid"`
}

type CandidateListItem struct {
	ID         int        `json:"id"`
	FullName   string     `json:"full_name"`
	Email      string     `json:"email"`
	City       *string    `json:"city"`
	School     *string    `json:"school"`
	Major      *string    `json:"major"`
	Status     string     `json:"status"`
	CreatedAt  time.Time  `json:"created_at"`
	FinalScore *float64   `json:"final_score"`
	Category   *string    `json:"category"`
	AnalyzedAt *time.Time `json:"analyzed_at"`
	ModelUsed  *string    `json:"model_used"`
	PhotoURL   *string    `json:"photo_url"`
	PhotoAIFlag bool      `json:"photo_ai_flag"`
	Age        *int       `json:"age"`
	NetScore   *int       `json:"net_score"`
}

type CandidateDetail struct {
	Candidate
	Analysis  *Analysis  `json:"analysis"`
	Decisions []Decision `json:"decisions"`
}

type CreateCandidateRequest struct {
	FullName            string   `json:"full_name" binding:"required"`
	Email               string   `json:"email" binding:"required,email"`
	Phone               string   `json:"phone" binding:"required"`
	Telegram            string   `json:"telegram" binding:"required"`
	Age                 int      `json:"age" binding:"required"`
	City                string   `json:"city" binding:"required"`
	School              string   `json:"school" binding:"required"`
	GraduationYear      int      `json:"graduation_year" binding:"required"`
	Achievements        string   `json:"achievements" binding:"required"`
	Extracurriculars    string   `json:"extracurriculars" binding:"required"`
	Essay               string   `json:"essay" binding:"required"`
	MotivationStatement string   `json:"motivation_statement" binding:"required"`
	Disability          *string  `json:"disability"`
	Major               *string  `json:"major"`
	YouTubeURL          string   `json:"youtube_url" binding:"required"`
}

// Majors maps tag → display name (multilingual)
var Majors = []MajorOption{
	{
		Tag:  "Engineering",
		En:   "Creative Engineering",
		Ru:   "Креативная инженерия",
		Kk:   "Креативті инженерия",
	},
	{
		Tag:  "Tech",
		En:   "Innovative IT Product Design and Development",
		Ru:   "Инновационные цифровые продукты и сервисы",
		Kk:   "Инновациялық цифрлық өнімдер мен қызметтер",
	},
	{
		Tag:  "Society",
		En:   "Sociology: Leadership and Innovation",
		Ru:   "Социология инноваций и лидерства",
		Kk:   "Инновация және көшбасшылық социологиясы",
	},
	{
		Tag:  "Policy Reform",
		En:   "Public Policy and Development",
		Ru:   "Стратегии государственного управления и развития",
		Kk:   "Мемлекеттік басқару және даму стратегиялары",
	},
	{
		Tag:  "Art + Media",
		En:   "Digital Media and Marketing",
		Ru:   "Цифровые медиа и маркетинг",
		Kk:   "Цифрлық медиа және маркетинг",
	},
}

type MajorOption struct {
	Tag string `json:"tag"`
	En  string `json:"en"`
	Ru  string `json:"ru"`
	Kk  string `json:"kk"`
}
