package telegram_bot

import (
	"sync"
	"time"

	"github.com/assylkhan/invisionu-backend/internal/models"
)

// текущая фаза интервью
type InterviewState string

const (
	StateLanguageSelect InterviewState = "language_select"
	StateWarmUp         InterviewState = "warmup"
	StateLeadership     InterviewState = "leadership"
	StateMotivation     InterviewState = "motivation"
	StateGrowth         InterviewState = "growth"
	StateVision         InterviewState = "vision"
	StateVerify         InterviewState = "verify"
	StateClosing        InterviewState = "closing"
	StateEvaluating     InterviewState = "evaluating"
	StateCompleted      InterviewState = "completed"
)

// порядок тем интервью
var topicOrder = []InterviewState{
	StateWarmUp,
	StateLeadership,
	StateMotivation,
	StateGrowth,
	StateVision,
	StateVerify,
}

// лимит вопросов на тему
var maxQuestionsPerTopic = map[InterviewState]int{
	StateWarmUp:     2,
	StateLeadership: 3,
	StateMotivation: 3,
	StateGrowth:     2,
	StateVision:     2,
	StateVerify:     2,
}

// in-memory сост��яние активного интервью
type activeSession struct {
	mu              sync.Mutex
	InterviewID     int
	CandidateID     int
	CandidateName   string
	Language        string
	State           InterviewState
	CurrentTopic    string
	QuestionsAsked  int
	TopicQuestions  map[string]int
	EssaySummary    string
	EssayHighlights []string // факты из эссе для верификации
	HasDisability   bool
	DisabilityInfo  string
	Conversation    []models.ConversationMessage
	LastBotMsgTime  time.Time
}

// следующая тема или StateClosing
func nextTopic(current InterviewState) InterviewState {
	for i, t := range topicOrder {
		if t == current && i+1 < len(topicOrder) {
			return topicOrder[i+1]
		}
	}
	return StateClosing
}

// состояние → строка current_topic в БД
func topicToString(s InterviewState) string {
	return string(s)
}
