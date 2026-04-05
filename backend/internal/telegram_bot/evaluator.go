package telegram_bot

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/assylkhan/invisionu-backend/internal/gemini"
	"github.com/assylkhan/invisionu-backend/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

// финальная LLM-оценка завершённого интервью
type Evaluator struct {
	pool         *pgxpool.Pool
	generateText func(ctx context.Context, systemPrompt, userMessage string) (string, error)
	modelName    string
}

// создаёт evaluator
func NewEvaluator(pool *pgxpool.Pool, genFn func(ctx context.Context, systemPrompt, userMessage string) (string, error), modelName string) *Evaluator {
	return &Evaluator{pool: pool, generateText: genFn, modelName: modelName}
}

type evalResponse struct {
	ScoreLeadership        int      `json:"score_leadership"`
	ScoreGrit              int      `json:"score_grit"`
	ScoreAuthenticity      int      `json:"score_authenticity"`
	ScoreMotivation        int      `json:"score_motivation"`
	ScoreVision            int      `json:"score_vision"`
	ExplanationLeadership  string   `json:"explanation_leadership"`
	ExplanationGrit        string   `json:"explanation_grit"`
	ExplanationAuthenticity string  `json:"explanation_authenticity"`
	ExplanationMotivation  string   `json:"explanation_motivation"`
	ExplanationVision      string   `json:"explanation_vision"`
	ConsistencyScore       int      `json:"consistency_score"`
	StyleMatchScore        int      `json:"style_match_score"`
	SuspicionFlags         []string `json:"suspicion_flags"`
	Summary                string   `json:"summary"`
	Strengths              []string `json:"strengths"`
	Concerns               []string `json:"concerns"`
}

const evalSystemPrompt = `You are an expert evaluator for inVision U scholarship interviews.
You will receive:
1. The candidate's essay summary from Stage 1 (written application)
2. The full interview transcript from Stage 2 (Telegram conversation)
3. Behavioral signals (response times, message types, flags)

Evaluate the candidate on these dimensions (0-100 each).
For EACH dimension, provide a score AND a short explanation (2-3 sentences) justifying the score with specific evidence from the interview:
- score_leadership + explanation_leadership: Evidence of initiative, leading others, taking responsibility. Look for concrete examples.
- score_grit + explanation_grit: Persistence, resilience, handling failure, determination. Did they overcome obstacles?
- score_authenticity + explanation_authenticity: How genuine are their responses? Do they match the essay? Natural conversation vs scripted?
- score_motivation + explanation_motivation: Depth of passion, personal connection, intrinsic vs extrinsic motivation.
- score_vision + explanation_vision: Clarity of future goals, understanding of impact, realistic ambition.

Also evaluate:
- consistency_score (0-100): How well do interview answers match claims made in the essay? High = consistent, Low = contradictions found.
- style_match_score (0-100): Linguistic similarity between essay writing style and chat style. Consider that chat is naturally more casual. 70+ is normal for genuine authors.
- suspicion_flags: Array of specific concerns. Use these codes when applicable:
  "too_fast" - multiple responses under 5 seconds with substantial content
  "style_shift" - dramatic change in writing style mid-conversation
  "factual_inconsistency" - contradicts facts stated in essay
  "generic_answers" - responses are vague, could apply to anyone
  "copy_paste" - responses appear pre-written or pasted
  "failed_verification" - could not answer questions about their own essay
  Empty array [] if no concerns.

Return ONLY a valid JSON object with all fields filled.`

// финальная оценка после завершения интервью
func (e *Evaluator) EvaluateInterview(ctx context.Context, session *activeSession) error {
	session.mu.Lock()
	interviewID := session.InterviewID
	candidateID := session.CandidateID
	essaySummary := session.EssaySummary
	conversation := make([]models.ConversationMessage, len(session.Conversation))
	copy(conversation, session.Conversation)
	session.mu.Unlock()

	// собираем античит-сигналы
	report := collectAntiCheatSignals(conversation)

	// строим промпт
	var sb strings.Builder
	sb.WriteString("=== ESSAY SUMMARY (Stage 1) ===\n")
	sb.WriteString(essaySummary)
	sb.WriteString("\n\n")

	sb.WriteString(buildAntiCheatSection(report))
	sb.WriteString("\n")

	sb.WriteString("=== INTERVIEW TRANSCRIPT ===\n")
	for _, msg := range conversation {
		role := "Interviewer"
		if msg.Role == "candidate" {
			role = "Candidate"
			if msg.MessageType == "voice" {
				role = "Candidate [VOICE]"
			}
			if msg.ResponseTimeSec > 0 {
				role += fmt.Sprintf(" (responded in %ds)", msg.ResponseTimeSec)
			}
		}
		sb.WriteString(fmt.Sprintf("%s: %s\n\n", role, msg.Content))
	}

	sb.WriteString("\nEvaluate this interview and return the JSON result.")

	result, err := e.generateText(ctx, evalSystemPrompt, sb.String())
	if err != nil {
		return fmt.Errorf("LLM evaluation failed: %w", err)
	}

	// парсим ответ
	result = strings.TrimSpace(result)
	if idx := strings.Index(result, "{"); idx > 0 {
		result = result[idx:]
	}
	if idx := strings.LastIndex(result, "}"); idx >= 0 && idx < len(result)-1 {
		result = result[:idx+1]
	}

	var eval evalResponse
	if err := json.Unmarshal([]byte(result), &eval); err != nil {
		return fmt.Errorf("failed to parse evaluation: %w (raw: %.300s)", err, result)
	}

	// ограничиваем диапазон
	eval.ScoreLeadership = clamp(eval.ScoreLeadership)
	eval.ScoreGrit = clamp(eval.ScoreGrit)
	eval.ScoreAuthenticity = clamp(eval.ScoreAuthenticity)
	eval.ScoreMotivation = clamp(eval.ScoreMotivation)
	eval.ScoreVision = clamp(eval.ScoreVision)
	eval.ConsistencyScore = clamp(eval.ConsistencyScore)
	eval.StyleMatchScore = clamp(eval.StyleMatchScore)

	if eval.SuspicionFlags == nil {
		eval.SuspicionFlags = []string{}
	}
	if eval.Strengths == nil {
		eval.Strengths = []string{}
	}
	if eval.Concerns == nil {
		eval.Concerns = []string{}
	}

	// итоговый балл (взвешенный)
	finalScore := math.Round((float64(eval.ScoreLeadership)*0.25+
		float64(eval.ScoreGrit)*0.20+
		float64(eval.ScoreAuthenticity)*0.20+
		float64(eval.ScoreMotivation)*0.20+
		float64(eval.ScoreVision)*0.15)*100) / 100
	category := gemini.ScoreToCategory(finalScore)

	// сохраняем в БД
	flagsJSON, _ := json.Marshal(eval.SuspicionFlags)

	_, err = e.pool.Exec(ctx, `
		INSERT INTO interview_analyses
			(interview_id, candidate_id, score_leadership, score_grit, score_authenticity,
			 score_motivation, score_vision, final_score, category,
			 consistency_score, style_match_score, suspicion_flags,
			 summary, strengths, concerns, model_used,
			 explanation_leadership, explanation_grit, explanation_authenticity,
			 explanation_motivation, explanation_vision)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19,$20,$21)
		ON CONFLICT (candidate_id) DO UPDATE SET
			interview_id=$1, score_leadership=$3, score_grit=$4, score_authenticity=$5,
			score_motivation=$6, score_vision=$7, final_score=$8, category=$9,
			consistency_score=$10, style_match_score=$11, suspicion_flags=$12,
			summary=$13, strengths=$14, concerns=$15, model_used=$16, analyzed_at=NOW(),
			explanation_leadership=$17, explanation_grit=$18, explanation_authenticity=$19,
			explanation_motivation=$20, explanation_vision=$21`,
		interviewID, candidateID,
		eval.ScoreLeadership, eval.ScoreGrit, eval.ScoreAuthenticity,
		eval.ScoreMotivation, eval.ScoreVision, finalScore, category,
		eval.ConsistencyScore, eval.StyleMatchScore, string(flagsJSON),
		eval.Summary, eval.Strengths, eval.Concerns, e.modelName,
		eval.ExplanationLeadership, eval.ExplanationGrit, eval.ExplanationAuthenticity,
		eval.ExplanationMotivation, eval.ExplanationVision,
	)
	if err != nil {
		return fmt.Errorf("save interview analysis: %w", err)
	}

	// комбинированный балл: этап1 (60%) + этап2 (40%)
	var stage1Score *float64
	e.pool.QueryRow(ctx, `SELECT final_score FROM analyses WHERE candidate_id = $1`, candidateID).Scan(&stage1Score)

	var combinedScore float64
	if stage1Score != nil {
		combinedScore = math.Round((*stage1Score*0.60+finalScore*0.40)*100) / 100
	} else {
		combinedScore = finalScore
	}

	// обновляем кандидата
	now := time.Now()
	_, err = e.pool.Exec(ctx, `
		UPDATE candidates SET interview_status = 'completed', combined_score = $1 WHERE id = $2`,
		combinedScore, candidateID)
	if err != nil {
		return fmt.Errorf("update candidate combined score: %w", err)
	}

	// завершаем интервью
	_, err = e.pool.Exec(ctx, `
		UPDATE interviews SET status = 'completed', completed_at = $1 WHERE id = $2`,
		now, interviewID)
	if err != nil {
		return fmt.Errorf("update interview status: %w", err)
	}

	// закрываем приглашение
	_, err = e.pool.Exec(ctx, `
		UPDATE telegram_invites SET status = 'completed' WHERE candidate_id = $1`,
		candidateID)
	if err != nil {
		return fmt.Errorf("update invite status: %w", err)
	}

	return nil
}

// загружает интервью из БД и запускает оценку (force/re-evaluate)
func (e *Evaluator) EvaluateFromDB(ctx context.Context, interviewID, candidateID int) error {
	var lang, essaySummary, candidateName string
	var contextJSON []byte
	var questionsAsked int

	err := e.pool.QueryRow(ctx, `
		SELECT i.language, i.essay_summary, i.conversation_context, i.questions_asked, c.full_name
		FROM interviews i
		JOIN candidates c ON c.id = i.candidate_id
		WHERE i.id = $1`, interviewID).Scan(&lang, &essaySummary, &contextJSON, &questionsAsked, &candidateName)
	if err != nil {
		return fmt.Errorf("failed to load interview %d: %w", interviewID, err)
	}

	var conversation []models.ConversationMessage
	if err := json.Unmarshal(contextJSON, &conversation); err != nil {
		return fmt.Errorf("failed to unmarshal conversation: %w", err)
	}

	session := &activeSession{
		InterviewID:    interviewID,
		CandidateID:    candidateID,
		CandidateName:  candidateName,
		Language:       lang,
		State:          StateEvaluating,
		QuestionsAsked: questionsAsked,
		EssaySummary:   essaySummary,
		Conversation:   conversation,
	}

	return e.EvaluateInterview(ctx, session)
}

func clamp(v int) int {
	if v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return v
}
