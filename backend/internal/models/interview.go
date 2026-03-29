package models

import (
	"encoding/json"
	"time"
)

// TelegramInvite links a candidate to a Telegram deep-link token.
type TelegramInvite struct {
	ID             int        `json:"id"`
	CandidateID    int        `json:"candidate_id"`
	Token          string     `json:"token"`
	TelegramChatID *int64     `json:"telegram_chat_id"`
	Status         string     `json:"status"` // pending, linked, interview_active, completed, expired
	CreatedAt      time.Time  `json:"created_at"`
	LinkedAt       *time.Time `json:"linked_at"`
	ExpiresAt      time.Time  `json:"expires_at"`
}

// Interview holds the state of a single Telegram interview session.
type Interview struct {
	ID                  int              `json:"id"`
	CandidateID         int              `json:"candidate_id"`
	TelegramChatID      int64            `json:"telegram_chat_id"`
	Status              string           `json:"status"` // active, completed, abandoned, timeout
	Language            string           `json:"language"`
	CurrentTopic        string           `json:"current_topic"`
	QuestionsAsked      int              `json:"questions_asked"`
	StartedAt           time.Time        `json:"started_at"`
	CompletedAt         *time.Time       `json:"completed_at,omitempty"`
	EssaySummary        string           `json:"essay_summary"`
	ConversationContext json.RawMessage  `json:"conversation_context"`
}

// ConversationMessage is one turn in the interview dialogue, stored as JSONB.
type ConversationMessage struct {
	Role           string `json:"role"`            // bot | candidate
	Content        string `json:"content"`
	Topic          string `json:"topic,omitempty"`
	QuestionType   string `json:"question_type,omitempty"` // star_situation, star_task, star_action, star_result, followup, probe, transition, verify, warmup
	MessageType    string `json:"message_type,omitempty"`  // text | voice
	ResponseTimeSec int   `json:"response_time_sec,omitempty"`
	Timestamp      string `json:"timestamp"`
}

// InterviewMessage is an individual message row for audit.
type InterviewMessage struct {
	ID               int        `json:"id"`
	InterviewID      int        `json:"interview_id"`
	Role             string     `json:"role"` // bot | candidate
	Content          string     `json:"content"`
	MessageType      string     `json:"message_type"` // text | voice
	VoiceDurationSec *int       `json:"voice_duration_sec,omitempty"`
	ResponseTimeSec  *int       `json:"response_time_sec,omitempty"`
	CreatedAt        time.Time  `json:"created_at"`
}

// InterviewAnalysis stores the LLM evaluation of a completed interview.
type InterviewAnalysis struct {
	ID               int       `json:"id"`
	InterviewID      int       `json:"interview_id"`
	CandidateID      int       `json:"candidate_id"`
	ScoreLeadership  int       `json:"score_leadership"`
	ScoreGrit        int       `json:"score_grit"`
	ScoreAuthenticity int      `json:"score_authenticity"`
	ScoreMotivation  int       `json:"score_motivation"`
	ScoreVision      int       `json:"score_vision"`
	FinalScore       float64   `json:"final_score"`
	Category         string    `json:"category"`
	ConsistencyScore int       `json:"consistency_score"`
	StyleMatchScore  int       `json:"style_match_score"`
	SuspicionFlags   []string  `json:"suspicion_flags"`
	Summary          string    `json:"summary"`
	Strengths        []string  `json:"strengths"`
	Concerns         []string  `json:"concerns"`
	AnalyzedAt       time.Time `json:"analyzed_at"`
	ModelUsed        string    `json:"model_used"`
}
