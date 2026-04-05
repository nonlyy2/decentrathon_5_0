package telegram_bot

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/assylkhan/invisionu-backend/internal/models"
)

// генерирует вопросы STAR-метода через LLM
type InterviewEngine struct {
	generateText func(ctx context.Context, systemPrompt, userMessage string) (string, error)
}

// создаёт движок с переданным LLM
func NewInterviewEngine(genFn func(ctx context.Context, systemPrompt, userMessage string) (string, error)) *InterviewEngine {
	return &InterviewEngine{generateText: genFn}
}

type questionResponse struct {
	Question        string `json:"question"`
	QuestionType    string `json:"question_type"`
	MoveToNextTopic bool   `json:"move_to_next_topic"`
}

const questionGenSystemPrompt = `You are an AI interviewer for inVision U, a 100% scholarship university.
You are conducting a behavioral interview using the STAR method (Situation, Task, Action, Result).
The entire interview MUST be conducted in English. Respond ONLY in English.

Your role:
- Ask ONE question at a time
- Be warm, conversational, and encouraging — like a real interviewer
- Follow up on the candidate's answers naturally
- If the candidate gave a Situation, ask about their specific Task or Action
- If they gave an Action, ask about the Result and Impact
- If their answer is vague, probe deeper: "Can you give a specific example?"
- If the topic is well-covered (2-3 meaningful answers), set move_to_next_topic to true
- Never repeat a question that was already asked
- Keep questions concise (1-2 sentences max)
- If the candidate responds in a non-English language, gently remind them to use English

Return ONLY a valid JSON object:
{"question": "your question here", "question_type": "star_situation|star_task|star_action|star_result|followup|probe|transition|verify", "move_to_next_topic": false}`

// следующий вопрос по истории диалога
func (e *InterviewEngine) GenerateNextQuestion(ctx context.Context, session *activeSession) (*questionResponse, error) {
	session.mu.Lock()
	topic := session.State
	conversation := make([]models.ConversationMessage, len(session.Conversation))
	copy(conversation, session.Conversation)
	essaySummary := session.EssaySummary
	topicQCount := session.TopicQuestions[string(topic)]
	maxQ := maxQuestionsPerTopic[topic]
	highlights := session.EssayHighlights
	hasDisability := session.HasDisability
	disabilityInfo := session.DisabilityInfo
	session.mu.Unlock()

	var sb strings.Builder
	sb.WriteString("=== INTERVIEW CONTEXT ===\n")
	sb.WriteString(fmt.Sprintf("Current topic: %s\n", string(topic)))
	sb.WriteString(fmt.Sprintf("Questions asked on this topic: %d / %d\n", topicQCount, maxQ))
	sb.WriteString(fmt.Sprintf("Essay summary: %s\n\n", essaySummary))

	if hasDisability {
		sb.WriteString(fmt.Sprintf("IMPORTANT: This candidate has a disability: %s. Accommodate accordingly — text-only responses are acceptable.\n\n", disabilityInfo))
	}

	if topic == StateVerify && len(highlights) > 0 {
		sb.WriteString("=== ESSAY FACTS TO VERIFY ===\n")
		sb.WriteString("Ask about specific details from the essay. You may slightly misstate a fact to test if the candidate corrects you (genuine authors will notice).\n")
		for _, h := range highlights {
			sb.WriteString(fmt.Sprintf("- %s\n", h))
		}
		sb.WriteString("\n")
	}

	if topic == StateWarmUp {
		sb.WriteString("This is the warm-up phase. Ask a light, friendly question to help the candidate relax. Examples: what they're currently excited about, a fun fact about themselves, or what they do in their free time.\n\n")
	}

	sb.WriteString("=== CONVERSATION SO FAR ===\n")
	for _, msg := range conversation {
		role := "Interviewer"
		if msg.Role == "candidate" {
			role = "Candidate"
			if msg.MessageType == "voice" {
				role = "Candidate [VOICE MESSAGE]"
			}
		}
		sb.WriteString(fmt.Sprintf("%s: %s\n\n", role, msg.Content))
	}

	sb.WriteString("Generate the next question in English.")

	result, err := e.generateText(ctx, questionGenSystemPrompt, sb.String())
	if err != nil {
		return nil, fmt.Errorf("LLM question generation failed: %w", err)
	}

	// парсим ответ
	result = strings.TrimSpace(result)
	if idx := strings.Index(result, "{"); idx > 0 {
		result = result[idx:]
	}
	if idx := strings.LastIndex(result, "}"); idx >= 0 && idx < len(result)-1 {
		result = result[:idx+1]
	}

	var qr questionResponse
	if err := json.Unmarshal([]byte(result), &qr); err != nil {
		return nil, fmt.Errorf("failed to parse question response: %w (raw: %.200s)", err, result)
	}

	return &qr, nil
}

// приветственное сообщение (только English)
func GenerateWelcomeMessage(candidateName string) string {
	return fmt.Sprintf(`Hi %s! Welcome to the inVision U interview.

I'm your AI interviewer. This conversation will help us get to know you better beyond your written application.

A few things to know:
- We strongly prefer voice messages, but text is also fine
- The entire interview is conducted in English
- There are no right or wrong answers — just be yourself
- The interview takes about 15-20 minutes
- I'll ask you questions about your experiences, goals, and motivations

Let's start with something easy!`, candidateName)
}

// приветствие с учётом инвалидности
func GenerateWelcomeMessageWithDisability(candidateName, disability string) string {
	return fmt.Sprintf(`Hi %s! Welcome to the inVision U interview.

I'm your AI interviewer. This conversation will help us get to know you better beyond your written application.

I understand you have indicated: %s — please feel completely comfortable responding in whatever way works best for you. Text messages are perfectly fine.

A few things to know:
- The entire interview is conducted in English
- There are no right or wrong answers — just be yourself
- The interview takes about 15-20 minutes
- I'll ask you questions about your experiences, goals, and motivations

Let's start with something easy!`, candidateName, disability)
}

// прощальное сообщение
func GenerateClosingMessage() string {
	return `Thank you so much for your time! That was a great conversation.

Your interview is now being evaluated. The results will appear on your candidate profile in the inVision U dashboard within a few minutes.

Good luck with your application! We appreciate your openness and honesty.`
}

// извлекаем верифицируемые факты из резюме эссе
func extractEssayHighlights(essaySummary string) []string {
	var highlights []string
	sentences := strings.Split(essaySummary, ". ")
	for _, s := range sentences {
		s = strings.TrimSpace(s)
		if len(s) > 20 {
			highlights = append(highlights, s)
		}
	}
	if len(highlights) > 5 {
		highlights = highlights[:5]
	}
	return highlights
}
