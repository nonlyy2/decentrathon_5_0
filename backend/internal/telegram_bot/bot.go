package telegram_bot

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/assylkhan/invisionu-backend/internal/config"
	"github.com/assylkhan/invisionu-backend/internal/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// функция генерации текста через LLM
type TextGenerator func(ctx context.Context, systemPrompt, userMessage string) (string, error)

// Telegram interview бот
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

// создаёт и конфигурирует бота
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

	if cfg.WhisperAPIKey != "" {
		bot.whisper = NewAlemSTTClient(cfg.WhisperAPIKey)
	}

	return bot, nil
}

// username бота
func (b *Bot) Username() string {
	return b.api.Self.UserName
}

// запуск long-polling (блокирует до отмены ctx; ручной polling из-за 409 конфликтов)
func (b *Bot) Start(ctx context.Context) {
	log.Printf("[TG-BOT] Starting @%s (long polling)", b.api.Self.UserName)

	// восстанавливаем сессии из БД
	b.restoreSessions(ctx)

	// повторяем зависшие оценки (краш сервера в середине)
	b.retryStuckEvaluations(ctx)

	// сбрасываем очередь апдейтов
	clearURL := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates?offset=-1&timeout=0", b.api.Token)
	resp, err := http.Get(clearURL)
	if err != nil {
		log.Printf("[TG-BOT] Warning: failed to clear pending updates: %v", err)
	} else {
		resp.Body.Close()
	}
	time.Sleep(2 * time.Second)

	// таймаут интервью
	go b.timeoutChecker(ctx)

	offset := 0
	conflictErrors := 0
	const maxConflictErrors = 5

	for {
		select {
		case <-ctx.Done():
			log.Println("[TG-BOT] Shutting down")
			return
		default:
		}

		u := tgbotapi.NewUpdate(offset)
		u.Timeout = 30

		updates, err := b.api.GetUpdates(u)
		if err != nil {
			if strings.Contains(err.Error(), "Conflict") || strings.Contains(err.Error(), "409") {
				conflictErrors++
				if conflictErrors >= maxConflictErrors {
					log.Printf("[TG-BOT] Too many 409 conflicts (%d), another bot instance is running. Disabling polling — HTTP server continues.", conflictErrors)
					return
				}
				log.Printf("[TG-BOT] Conflict (%d/%d): another bot instance detected, retrying...", conflictErrors, maxConflictErrors)
				time.Sleep(3 * time.Second)
				continue
			}
			log.Printf("[TG-BOT] Failed to get updates: %v", err)
			time.Sleep(3 * time.Second)
			continue
		}

		conflictErrors = 0
		for _, update := range updates {
			if update.UpdateID >= offset {
				offset = update.UpdateID + 1
			}
			go b.handleUpdate(ctx, update)
		}
	}
}

// загружаем активные интервью из БД в память
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

		// количество вопросов по темам
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

// повтор оценок, застрявших при краше сервера (topic='evaluating', status='active')
func (b *Bot) retryStuckEvaluations(ctx context.Context) {
	rows, err := b.pool.Query(ctx, `
		SELECT i.id, i.candidate_id, i.language, i.essay_summary, i.conversation_context,
		       i.telegram_chat_id, c.full_name, i.questions_asked
		FROM interviews i
		JOIN candidates c ON c.id = i.candidate_id
		WHERE i.status = 'active' AND i.current_topic = 'evaluating'`)
	if err != nil {
		log.Printf("[TG-BOT] Failed to query stuck evaluations: %v", err)
		return
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var (
			interviewID    int
			candidateID    int
			lang           string
			essaySummary   string
			contextJSON    []byte
			chatID         int64
			candidateName  string
			questionsAsked int
		)
		if err := rows.Scan(&interviewID, &candidateID, &lang, &essaySummary, &contextJSON,
			&chatID, &candidateName, &questionsAsked); err != nil {
			log.Printf("[TG-BOT] Failed to scan stuck evaluation: %v", err)
			continue
		}

		var conversation []models.ConversationMessage
		if err := json.Unmarshal(contextJSON, &conversation); err != nil {
			log.Printf("[TG-BOT] Failed to unmarshal conversation for stuck interview %d: %v", interviewID, err)
			continue
		}

		session := &activeSession{
			InterviewID:   interviewID,
			CandidateID:   candidateID,
			CandidateName: candidateName,
			Language:      lang,
			State:         StateEvaluating,
			QuestionsAsked: questionsAsked,
			EssaySummary:  essaySummary,
			Conversation:  conversation,
		}

		count++
		go func(iid int, cid int64, s *activeSession) {
			log.Printf("[TG-BOT] Retrying stuck evaluation for interview %d", iid)
			if err := b.evaluator.EvaluateInterview(context.Background(), s); err != nil {
				log.Printf("[TG-BOT] Stuck evaluation retry failed for interview %d: %v", iid, err)
			} else {
				log.Printf("[TG-BOT] Stuck evaluation completed for interview %d", iid)
				doneMsg := map[string]string{
					"en": "Your interview has been evaluated! Thank you for participating.",
					"ru": "Ваше интервью оценено! Спасибо за участие.",
					"kz": "Сұхбатыңыз бағаланды! Қатысқаныңыз үшін рахмет.",
				}
				if m := doneMsg[s.Language]; m != "" {
					b.sendMessage(cid, m)
				}
			}
		}(interviewID, chatID, session)
	}

	if count > 0 {
		log.Printf("[TG-BOT] Found %d stuck evaluations, retrying...", count)
	}
}

// периодически закрывает зависшие интервью по таймауту
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

// завершение интервью по таймауту
func (b *Bot) handleTimeout(ctx context.Context, chatID int64, session *activeSession) {
	session.mu.Lock()
	session.State = StateEvaluating
	interviewID := session.InterviewID
	session.mu.Unlock()

	// уведомляем кандидата
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

	b.pool.Exec(ctx, `UPDATE interviews SET status = 'timeout', completed_at = NOW() WHERE id = $1`, interviewID)
	b.pool.Exec(ctx, `UPDATE candidates SET interview_status = 'completed' WHERE id = $1`, session.CandidateID)

	// оцениваем если есть достаточно данных
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

// отправка текстового сообщения
func (b *Bot) sendMessage(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = ""
	if _, err := b.api.Send(msg); err != nil {
		log.Printf("[TG-BOT] Failed to send message to %d: %v", chatID, err)
	}
}

// сообщение с markdown
func (b *Bot) sendMarkdown(chatID int64, text string) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ParseMode = "Markdown"
	if _, err := b.api.Send(msg); err != nil {
		log.Printf("[TG-BOT] Failed to send markdown message to %d: %v", chatID, err)
	}
}

// сообщение с inline клавиатурой
func (b *Bot) sendMessageWithKeyboard(chatID int64, text string, keyboard tgbotapi.InlineKeyboardMarkup) {
	msg := tgbotapi.NewMessage(chatID, text)
	msg.ReplyMarkup = keyboard
	if _, err := b.api.Send(msg); err != nil {
		log.Printf("[TG-BOT] Failed to send message with keyboard to %d: %v", chatID, err)
	}
}

// сохраняем контекст диалога в БД
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

// сохраняем сообщение для аудита
func (b *Bot) saveMessage(ctx context.Context, interviewID int, role, content, msgType string, voiceDur, responseTime *int) {
	_, err := b.pool.Exec(ctx, `
		INSERT INTO interview_messages (interview_id, role, content, message_type, voice_duration_sec, response_time_sec)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		interviewID, role, content, msgType, voiceDur, responseTime)
	if err != nil {
		log.Printf("[TG-BOT] Failed to save message: %v", err)
	}
}
