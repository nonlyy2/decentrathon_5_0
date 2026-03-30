package telegram_bot

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/assylkhan/invisionu-backend/internal/config"
	"github.com/assylkhan/invisionu-backend/internal/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TextGenerator is a function that sends a prompt to an LLM and returns the response.
type TextGenerator func(ctx context.Context, systemPrompt, userMessage string) (string, error)

// Bot is the Telegram interview bot.
type Bot struct {
	api       *tgbotapi.BotAPI
	pool      *pgxpool.Pool
	cfg       *config.Config
	engine    *InterviewEngine
	evaluator *Evaluator
	whisper   *AlemSTTClient
	sessions  sync.Map // chatID (int64) -> *activeSession
	genText   TextGenerator
	modelName string
}

// New creates and configures the Telegram bot.
func New(cfg *config.Config, pool *pgxpool.Pool, textGen TextGenerator, modelName string) (*Bot, error) {
	botAPI, err := tgbotapi.NewBotAPI(cfg.TelegramBotToken)
	if err != nil {
		return nil, fmt.Errorf("telegram bot init failed: %w", err)
	}

	bot := &Bot{
		api:       botAPI,
		pool:      pool,
		cfg:       cfg,
		engine:    NewInterviewEngine(textGen),
		evaluator: NewEvaluator(pool, textGen, modelName),
		genText:   textGen,
		modelName: modelName,
	}

	if cfg.AlemSTTAPIKey != "" {
		bot.whisper = NewAlemSTTClient(cfg.AlemSTTAPIKey)
	}

	return bot, nil
}

// Username returns the bot's Telegram username.
func (b *Bot) Username() string {
	return b.api.Self.UserName
}

// Start begins the long-polling loop. Blocks until ctx is cancelled.
func (b *Bot) Start(ctx context.Context) {
	log.Printf("[TG-BOT] Starting @%s (long polling)", b.api.Self.UserName)

	// Restore active sessions from DB on startup
	b.restoreSessions(ctx)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	// Interview timeout checker
	go b.timeoutChecker(ctx)

	for {
		select {
		case <-ctx.Done():
			log.Println("[TG-BOT] Shutting down")
			b.api.StopReceivingUpdates()
			return
		case update := <-updates:
			go b.handleUpdate(ctx, update)
		}
	}
}

// restoreSessions loads active interviews from DB into memory.
func (b *Bot) restoreSessions(ctx context.Context) {
	rows, err := b.pool.Query(ctx, `
		SELECT i.id, i.candidate_id, i.telegram_chat_id, i.status, i.language,
			   i.current_topic, i.questions_asked, i.essay_summary, i.conversation_context,
			   c.full_name
		FROM interviews i
		JOIN candidates c ON c.id = i.candidate_id
		WHERE i.status = 'active'`)
	if err != nil {
		log.Printf("[TG-BOT] Failed to restore sessions: %v", err)
		return
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var (
			interviewID    int
			candidateID    int
			chatID         int64
			status         string
			lang           string
			topic          string
			questionsAsked int
			essaySummary   string
			contextJSON    []byte
			candidateName  string
		)
		if err := rows.Scan(&interviewID, &candidateID, &chatID, &status, &lang,
			&topic, &questionsAsked, &essaySummary, &contextJSON, &candidateName); err != nil {
			log.Printf("[TG-BOT] Failed to scan session: %v", err)
			continue
		}

		var conversation []models.ConversationMessage
		if err := json.Unmarshal(contextJSON, &conversation); err != nil {
			log.Printf("[TG-BOT] Failed to unmarshal conversation for interview %d: %v", interviewID, err)
			conversation = []models.ConversationMessage{}
		}

		// Count questions per topic
		topicQ := make(map[string]int)
		for _, msg := range conversation {
			if msg.Role == "bot" && msg.Topic != "" {
				topicQ[msg.Topic]++
			}
		}

		session := &activeSession{
			InterviewID:     interviewID,
			CandidateID:     candidateID,
			CandidateName:   candidateName,
			Language:        lang,
			State:           InterviewState(topic),
			CurrentTopic:    topic,
			QuestionsAsked:  questionsAsked,
			TopicQuestions:  topicQ,
			EssaySummary:    essaySummary,
			EssayHighlights: extractEssayHighlights(essaySummary),
			Conversation:    conversation,
			LastBotMsgTime:  time.Now(),
		}

		b.sessions.Store(chatID, session)
		count++
	}

	if count > 0 {
		log.Printf("[TG-BOT] Restored %d active sessions", count)
	}
}

// timeoutChecker periodically marks stale interviews as timed out.
func (b *Bot) timeoutChecker(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()

	timeout := time.Duration(b.cfg.InterviewTimeoutMin) * time.Minute

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			b.sessions.Range(func(key, value any) bool {
				chatID := key.(int64)
				session := value.(*activeSession)

				session.mu.Lock()
				elapsed := time.Since(session.LastBotMsgTime)
				state := session.State
				session.mu.Unlock()

				if state != StateCompleted && state != StateEvaluating && elapsed > timeout {
					log.Printf("[TG-BOT] Interview timeout for chat %d (idle %v)", chatID, elapsed)
					b.handleTimeout(ctx, chatID, session)
				}
				return true
			})
		}
	}
}

// handleTimeout marks an interview as timed out and triggers evaluation.
func (b *Bot) handleTimeout(ctx context.Context, chatID int64, session *activeSession) {
	session.mu.Lock()
	session.State = StateEvaluating
	interviewID := session.InterviewID
	session.mu.Unlock()

	// Notify candidate
	timeoutMsg := map[string]string{
		"en": "The interview has timed out due to inactivity. Your responses so far will still be evaluated. Thank you!",
		"ru": "Интервью завершено из-за неактивности. Твои ответы всё равно будут оценены. Спасибо!",
		"kz": "Сұхбат белсенділік болмағандықтан аяқталды. Сенің жауаптарың бәрібір бағаланады. Рахмет!",
	}
	lang := session.Language
	msg := timeoutMsg[lang]
	if msg == "" {
		msg = timeoutMsg["en"]
	}
	b.sendMessage(chatID, msg)

	// Update DB
	b.pool.Exec(ctx, `UPDATE interviews SET status = 'timeout', completed_at = NOW() WHERE id = $1`, interviewID)
	b.pool.Exec(ctx, `UPDATE candidates SET interview_status = 'completed' WHERE id = $1`, session.CandidateID)

	// Evaluate if we have enough data
	if session.QuestionsAsked >= 3 {
		go func() {
			if err := b.evaluator.EvaluateInterview(context.Background(), session); err != nil {
				log.Printf("[TG-BOT] Timeout evaluation failed for interview %d: %v", interviewID, err)
			} else {
				log.Printf("[TG-BOT] Timeout evaluation completed for interview %d", interviewID)
			}
		}()
	}

	b.sessions.Delete(chatID)
}

// sendMessage sends a text message to a Telegram chat.
func (b *Bot) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = ""
	if _, err := b.api.Send(msg); err != nil {
		log.Printf("[TG-BOT] Failed to send message to %d: %v", chatID, err)
	}
}

// sendMessageWithKeyboard sends a message with an inline keyboard.
func (b *Bot) sendMessageWithKeyboard(chatID int64, text string, keyboard tgbotapi.InlineKeyboardMarkup) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	if _, err := b.api.Send(msg); err != nil {
		log.Printf("[TG-BOT] Failed to send message with keyboard to %d: %v", chatID, err)
	}
}

// persistConversation saves the current conversation context to DB.
func (b *Bot) persistConversation(ctx context.Context, session *activeSession) {
	session.mu.Lock()
	data, _ := json.Marshal(session.Conversation)
	interviewID := session.InterviewID
	topic := string(session.State)
	questionsAsked := session.QuestionsAsked
	session.mu.Unlock()

	_, err := b.pool.Exec(ctx, `
		UPDATE interviews SET conversation_context = $1, current_topic = $2, questions_asked = $3
		WHERE id = $4`,
		data, topic, questionsAsked, interviewID)
	if err != nil {
		log.Printf("[TG-BOT] Failed to persist conversation for interview %d: %v", interviewID, err)
	}
}

// saveMessage inserts a message row for audit.
func (b *Bot) saveMessage(ctx context.Context, interviewID int, role, content, msgType string, voiceDur, responseTime *int) {
	_, err := b.pool.Exec(ctx, `
		INSERT INTO interview_messages (interview_id, role, content, message_type, voice_duration_sec, response_time_sec)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		interviewID, role, content, msgType, voiceDur, responseTime)
	if err != nil {
		log.Printf("[TG-BOT] Failed to save message: %v", err)
	}
}
