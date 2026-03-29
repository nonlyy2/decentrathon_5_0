package telegram_bot

import (
	"sync"
	"time"

	"github.com/assylkhan/invisionu-backend/internal/models"
)

// InterviewState represents the current phase of the interview.
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

// topicOrder defines the sequence of interview topics.
var topicOrder = []InterviewState{
	StateWarmUp,
	StateLeadership,
	StateMotivation,
	StateGrowth,
	StateVision,
	StateVerify,
}

// maxQuestionsPerTopic is the target number of questions per STAR topic.
var maxQuestionsPerTopic = map[InterviewState]int{
	StateWarmUp:     2,
	StateLeadership: 3,
	StateMotivation: 3,
	StateGrowth:     2,
	StateVision:     2,
	StateVerify:     2,
}

// activeSession holds in-memory state for a running interview.
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
	EssayHighlights []string // specific facts from essay for verification
	HasDisability   bool
	DisabilityInfo  string
	Conversation    []models.ConversationMessage
	LastBotMsgTime  time.Time
}

// nextTopic returns the topic that follows current, or StateClosing if done.
func nextTopic(current InterviewState) InterviewState {
	for i, t := range topicOrder {
		if t == current && i+1 < len(topicOrder) {
			return topicOrder[i+1]
		}
	}
	return StateClosing
}

// topicToString maps state to the DB current_topic value.
func topicToString(s InterviewState) string {
	return string(s)
}
