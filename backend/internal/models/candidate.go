package models

import "time"

type Candidate struct {
	ID                  int       `json:"id"`
	FullName            string    `json:"full_name"`
	FirstName           *string   `json:"first_name"`
	LastName            *string   `json:"last_name"`
	Patronymic          *string   `json:"patronymic"`
	Email               string    `json:"email"`
	Phone               *string   `json:"phone"`
	Telegram            *string   `json:"telegram"`
	Age                 *int      `json:"age"`
	DateOfBirth         *string   `json:"date_of_birth"`
	Gender              *string   `json:"gender"`
	City                *string   `json:"city"`
	HomeCountry         *string   `json:"home_country"`
	School              *string   `json:"school"`
	GraduationYear      *int      `json:"graduation_year"`
	Nationality         *string   `json:"nationality"`
	IIN                 *string   `json:"iin"`
	IdentityDocType     *string   `json:"identity_doc_type"`
	Instagram           *string   `json:"instagram"`
	WhatsApp            *string   `json:"whatsapp"`
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
	ExamType            *string   `json:"exam_type"`
	IELTSScore          *float64  `json:"ielts_score"`
	TOEFLScore          *int      `json:"toefl_score"`
	EnglishCertURL      *string   `json:"english_cert_url"`
	CertificateType     *string   `json:"certificate_type"`
	CertificateURL      *string   `json:"certificate_url"`
	AdditionalDocsURL   *string   `json:"additional_docs_url"`
	PersonalityAnswers  *string   `json:"personality_answers"`
	ReviewComplexity    *float64  `json:"review_complexity"`
	UNTScore            *int      `json:"unt_score"`
	NISGrade            *string   `json:"nis_grade"`
	PartnerSchool       *string   `json:"partner_school"`
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
	Age              *int       `json:"age"`
	NetScore         *int       `json:"net_score"`
	ReviewComplexity *float64   `json:"review_complexity"`
}

type CandidateDetail struct {
	Candidate
	Analysis  *Analysis  `json:"analysis"`
	Decisions []Decision `json:"decisions"`
}

type CreateCandidateRequest struct {
	FullName            string   `json:"full_name" binding:"required"`
	FirstName           string   `json:"first_name"`
	LastName            string   `json:"last_name"`
	Patronymic          string   `json:"patronymic"`
	Email               string   `json:"email" binding:"required,email"`
	Phone               string   `json:"phone" binding:"required"`
	Telegram            string   `json:"telegram" binding:"required"`
	Age                 int      `json:"age"`
	DateOfBirth         string   `json:"date_of_birth"`
	Gender              string   `json:"gender"`
	City                string   `json:"city" binding:"required"`
	HomeCountry         string   `json:"home_country"`
	School              string   `json:"school"`
	GraduationYear      int      `json:"graduation_year"`
	Nationality         string   `json:"nationality"`
	IIN                 string   `json:"iin"`
	IdentityDocType     string   `json:"identity_doc_type"`
	Instagram           string   `json:"instagram"`
	WhatsApp            string   `json:"whatsapp"`
	Achievements        string   `json:"achievements"`
	Extracurriculars    string   `json:"extracurriculars"`
	Essay               string   `json:"essay" binding:"required"`
	MotivationStatement string   `json:"motivation_statement"`
	Disability          *string  `json:"disability"`
	Major               *string  `json:"major"`
	YouTubeURL          string   `json:"youtube_url"`
	ExamType            string   `json:"exam_type"`
	IELTSScore          *float64 `json:"ielts_score"`
	TOEFLScore          *int     `json:"toefl_score"`
	CertificateType     string   `json:"certificate_type"`
	PersonalityAnswers  string   `json:"personality_answers"`
	UNTScore            *int     `json:"unt_score"`
	NISGrade            string   `json:"nis_grade"`
}

// специальн��сти: tag → название (мультиязычно)
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
	{
		Tag:  "Foundation",
		En:   "Foundation Program",
		Ru:   "Подготовительная программа",
		Kk:   "Дайындық бағдарламасы",
	},
}

type MajorOption struct {
	Tag string `json:"tag"`
	En  string `json:"en"`
	Ru  string `json:"ru"`
	Kk  string `json:"kk"`
}
