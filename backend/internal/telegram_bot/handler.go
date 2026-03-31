package telegram_bot

import (
	"context"
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/assylkhan/invisionu-backend/internal/models"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// handleUpdate routes an incoming Telegram update to the right handler.
func (b *Bot) handleUpdate(ctx context.Context, update tgbotapi.Update) {
	if update.Message == nil {
		return
	}

	msg := update.Message
	chatID := msg.Chat.ID

	// Handle /start command
	if msg.IsCommand() && msg.Command() == "start" {
		b.handleStart(ctx, chatID, msg.CommandArguments(), msg.From)
		return
	}

	// Check if there's an active session
	val, ok := b.sessions.Load(chatID)
	if !ok {
		b.sendMessage(chatID, "No active interview session. If you received an invite link, please use it to start.")
		return
	}

	session := val.(*activeSession)

	// Handle voice message
	if msg.Voice != nil {
		b.handleVoiceMessage(ctx, chatID, session, msg)
		return
	}

	// Handle text message
	if msg.Text != "" {
		b.handleTextMessage(ctx, chatID, session, msg.Text)
		return
	}
}

// normalizeTelegram strips leading @ and lowercases for comparison.
func normalizeTelegram(tg string) string {
	tg = strings.TrimSpace(tg)
	tg = strings.TrimPrefix(tg, "@")
	return strings.ToLower(tg)
}

// handleStart processes the /start command with a deep-link token.
func (b *Bot) handleStart(ctx context.Context, chatID int64, token string, from *tgbotapi.User) {
	token = strings.TrimSpace(token)

	if token == "" {
		b.sendMessage(chatID, "Welcome to inVision U Interview Bot!\n\nIf you have an invite link, please use the full link to start your interview.")
		return
	}

	// Check if this chat already has an active session
	if _, ok := b.sessions.Load(chatID); ok {
		b.sendMessage(chatID, "You already have an active interview session. Please continue answering questions.")
		return
	}

	// Look up the invite token
	var invite struct {
		ID          int
		CandidateID int
		Status      string
		ExpiresAt   time.Time
	}
	err := b.pool.QueryRow(ctx, `
		SELECT id, candidate_id, status, expires_at
		FROM telegram_invites WHERE token = $1`, token).Scan(
		&invite.ID, &invite.CandidateID, &invite.Status, &invite.ExpiresAt)

	if err != nil {
		log.Printf("[TG-BOT] Invalid token %s: %v", token, err)
		b.sendMessage(chatID, "Invalid or expired invite link. Please contact the admissions team.")
		return
	}

	if invite.Status != "pending" {
		b.sendMessage(chatID, "This invite link has already been used.")
		return
	}

	if time.Now().After(invite.ExpiresAt) {
		b.pool.Exec(ctx, `UPDATE telegram_invites SET status = 'expired' WHERE id = $1`, invite.ID)
		b.sendMessage(chatID, "This invite link has expired. Please contact the admissions team for a new one.")
		return
	}

	// Check candidate exists and fetch info including disability and telegram handle
	var candidate struct {
		FullName   string
		Summary    string
		Disability *string
		Telegram   *string
	}
	err = b.pool.QueryRow(ctx, `
		SELECT c.full_name, COALESCE(a.summary, ''), c.disability, c.telegram
		FROM candidates c
		LEFT JOIN analyses a ON a.candidate_id = c.id
		WHERE c.id = $1`, invite.CandidateID).Scan(&candidate.FullName, &candidate.Summary, &candidate.Disability, &candidate.Telegram)
	if err != nil {
		log.Printf("[TG-BOT] Candidate %d not found: %v", invite.CandidateID, err)
		b.sendMessage(chatID, "Error looking up your application. Please contact the admissions team.")
		return
	}

	// Verify the Telegram username matches the candidate's application
	if candidate.Telegram != nil && *candidate.Telegram != "" && from != nil {
		expectedTG := normalizeTelegram(*candidate.Telegram)
		actualTG := normalizeTelegram(from.UserName)
		if actualTG == "" || expectedTG != actualTG {
			log.Printf("[TG-BOT] Username mismatch for candidate %d: expected @%s, got @%s", invite.CandidateID, expectedTG, actualTG)
			b.sendMessage(chatID, "Your Telegram username does not match the one in your application. Please use the same Telegram account you registered with, or contact the admissions team.")
			return
		}
	}

	// Link the invite
	now := time.Now()
	_, err = b.pool.Exec(ctx, `
		UPDATE telegram_invites SET telegram_chat_id = $1, status = 'linked', linked_at = $2
		WHERE id = $3`, chatID, now, invite.ID)
	if err != nil {
		log.Printf("[TG-BOT] Failed to link invite %d: %v", invite.ID, err)
		b.sendMessage(chatID, "Technical error. Please try again.")
		return
	}

	// Update candidate record
	_, err = b.pool.Exec(ctx, `
		UPDATE candidates SET telegram_chat_id = $1, interview_status = 'in_progress'
		WHERE id = $2`, chatID, invite.CandidateID)
	if err != nil {
		log.Printf("[TG-BOT] Failed to update candidate %d: %v", invite.CandidateID, err)
	}

	// Determine disability accommodation
	hasDisability := candidate.Disability != nil && strings.TrimSpace(*candidate.Disability) != ""
	disabilityInfo := ""
	if hasDisability {
		disabilityInfo = *candidate.Disability
	}

	// Create interview record in DB (English only)
	var interviewID int
	contextJSON, _ := json.Marshal([]models.ConversationMessage{})
	err = b.pool.QueryRow(ctx, `
		INSERT INTO interviews (candidate_id, telegram_chat_id, language, essay_summary, conversation_context)
		VALUES ($1, $2, 'en', $3, $4)
		ON CONFLICT (candidate_id) DO UPDATE SET
			telegram_chat_id = $2, language = 'en', status = 'active',
			started_at = NOW(), completed_at = NULL, essay_summary = $3,
			conversation_context = $4, current_topic = 'warmup', questions_asked = 0
		RETURNING id`,
		invite.CandidateID, chatID, candidate.Summary, contextJSON).Scan(&interviewID)
	if err != nil {
		log.Printf("[TG-BOT] Failed to create interview: %v", err)
		b.sendMessage(chatID, "Technical error starting the interview. Please try again.")
		return
	}

	// Update invite status
	b.pool.Exec(ctx, `UPDATE telegram_invites SET status = 'interview_active' WHERE candidate_id = $1`, invite.CandidateID)

	// Build session
	session := &activeSession{
		InterviewID:    interviewID,
		CandidateID:    invite.CandidateID,
		CandidateName:  candidate.FullName,
		Language:       "en",
		State:          StateWarmUp,
		TopicQuestions: make(map[string]int),
		EssaySummary:   candidate.Summary,
		EssayHighlights: extractEssayHighlights(candidate.Summary),
		HasDisability:  hasDisability,
		DisabilityInfo: disabilityInfo,
	}
	b.sessions.Store(chatID, session)

	// Send welcome message (with disability accommodation if needed)
	var welcome string
	if hasDisability {
		welcome = GenerateWelcomeMessageWithDisability(candidate.FullName, disabilityInfo)
	} else {
		welcome = GenerateWelcomeMessage(candidate.FullName)
	}
	b.sendMessage(chatID, welcome)

	// Add welcome to conversation context
	session.mu.Lock()
	session.Conversation = append(session.Conversation, models.ConversationMessage{
		Role:      "bot",
		Content:   welcome,
		Topic:     "warmup",
		Timestamp: time.Now().Format(time.RFC3339),
	})
	session.LastBotMsgTime = time.Now()
	session.mu.Unlock()

	b.saveMessage(ctx, interviewID, "bot", welcome, "text", nil, nil)

	// Generate and send the first warm-up question
	b.generateAndSendQuestion(ctx, chatID, session)
}

// handleTextMessage processes a text message from a candidate.
func (b *Bot) handleTextMessage(ctx context.Context, chatID int64, session *activeSession, text string) {
	session.mu.Lock()
	state := session.State
	if state == StateLanguageSelect || state == StateClosing || state == StateEvaluating || state == StateCompleted {
		session.mu.Unlock()
		return
	}

	// Calculate response time
	responseTime := int(time.Since(session.LastBotMsgTime).Seconds())
	interviewID := session.InterviewID

	// Add candidate message to conversation
	session.Conversation = append(session.Conversation, models.ConversationMessage{
		Role:            "candidate",
		Content:         text,
		MessageType:     "text",
		ResponseTimeSec: responseTime,
		Timestamp:       time.Now().Format(time.RFC3339),
	})
	session.mu.Unlock()

	// Save message
	b.saveMessage(ctx, interviewID, "candidate", text, "text", nil, &responseTime)

	// Persist and generate next question
	b.persistConversation(ctx, session)
	b.processAndAdvance(ctx, chatID, session)
}

// handleVoiceMessage processes a voice message from a candidate.
func (b *Bot) handleVoiceMessage(ctx context.Context, chatID int64, session *activeSession, msg *tgbotapi.Message) {
	session.mu.Lock()
	state := session.State
	if state == StateLanguageSelect || state == StateClosing || state == StateEvaluating || state == StateCompleted {
		session.mu.Unlock()
		return
	}
	lang := session.Language
	interviewID := session.InterviewID
	responseTime := int(time.Since(session.LastBotMsgTime).Seconds())
	session.mu.Unlock()

	// Check if Alem STT is available
	if b.whisper == nil {
		b.sendMessage(chatID, "Voice messages are not supported at the moment. Please send a text message.")
		return
	}

	// Download voice file
	audioData, err := downloadVoiceFile(b.api, msg.Voice.FileID)
	if err != nil {
		log.Printf("[TG-BOT] Failed to download voice: %v", err)
		b.sendMessage(chatID, "Failed to process voice message. Please try again or send a text message.")
		return
	}

	// Transcribe
	transcription, err := b.whisper.Transcribe(ctx, audioData, lang)
	if err != nil {
		log.Printf("[TG-BOT] Transcription failed: %v", err)
		b.sendMessage(chatID, "Failed to transcribe voice message. Please try again or send a text message.")
		return
	}

	if strings.TrimSpace(transcription) == "" {
		b.sendMessage(chatID, "Could not understand the voice message. Please try again or send a text message.")
		return
	}

	voiceDur := msg.Voice.Duration

	session.mu.Lock()
	session.Conversation = append(session.Conversation, models.ConversationMessage{
		Role:            "candidate",
		Content:         transcription,
		MessageType:     "voice",
		ResponseTimeSec: responseTime,
		Timestamp:       time.Now().Format(time.RFC3339),
	})
	session.mu.Unlock()

	b.saveMessage(ctx, interviewID, "candidate", transcription, "voice", &voiceDur, &responseTime)

	b.persistConversation(ctx, session)
	b.processAndAdvance(ctx, chatID, session)
}

// processAndAdvance checks whether to advance topics and generates the next question.
func (b *Bot) processAndAdvance(ctx context.Context, chatID int64, session *activeSession) {
	session.mu.Lock()
	totalQ := session.QuestionsAsked
	maxQ := b.cfg.InterviewMaxQuestions
	session.mu.Unlock()

	// Check if we've hit the max questions
	if totalQ >= maxQ {
		b.finishInterview(ctx, chatID, session)
		return
	}

	b.generateAndSendQuestion(ctx, chatID, session)
}

// generateAndSendQuestion uses the LLM to generate the next question and sends it.
func (b *Bot) generateAndSendQuestion(ctx context.Context, chatID int64, session *activeSession) {
	qr, err := b.engine.GenerateNextQuestion(ctx, session)
	if err != nil {
		log.Printf("[TG-BOT] Question generation failed: %v", err)
		qr = &questionResponse{
			Question:     "Could you tell me more about that?",
			QuestionType: "followup",
		}
	}

	session.mu.Lock()
	currentTopic := session.State

	// Check if we should move to the next topic
	if qr.MoveToNextTopic {
		next := nextTopic(currentTopic)
		session.State = next
		session.CurrentTopic = string(next)
		currentTopic = next
	}

	// Check if we've completed all topics
	if currentTopic == StateClosing {
		session.mu.Unlock()
		b.finishInterview(ctx, chatID, session)
		return
	}

	session.QuestionsAsked++
	session.TopicQuestions[string(currentTopic)]++
	session.Conversation = append(session.Conversation, models.ConversationMessage{
		Role:         "bot",
		Content:      qr.Question,
		Topic:        string(currentTopic),
		QuestionType: qr.QuestionType,
		Timestamp:    time.Now().Format(time.RFC3339),
	})
	session.LastBotMsgTime = time.Now()
	interviewID := session.InterviewID
	session.mu.Unlock()

	b.sendMessage(chatID, qr.Question)
	b.saveMessage(ctx, interviewID, "bot", qr.Question, "text", nil, nil)
	b.persistConversation(ctx, session)
}

// finishInterview sends the closing message and triggers async evaluation.
func (b *Bot) finishInterview(ctx context.Context, chatID int64, session *activeSession) {
	session.mu.Lock()
	if session.State == StateEvaluating || session.State == StateCompleted {
		session.mu.Unlock()
		return
	}
	session.State = StateEvaluating
	interviewID := session.InterviewID
	session.mu.Unlock()

	// Send closing message
	closing := GenerateClosingMessage()
	b.sendMessage(chatID, closing)
	b.saveMessage(ctx, interviewID, "bot", closing, "text", nil, nil)

	// Update DB status
	b.pool.Exec(ctx, `UPDATE interviews SET current_topic = 'evaluating' WHERE id = $1`, interviewID)

	// Trigger async evaluation
	go func() {
		log.Printf("[TG-BOT] Starting evaluation for interview %d", interviewID)
		if err := b.evaluator.EvaluateInterview(context.Background(), session); err != nil {
			log.Printf("[TG-BOT] Evaluation failed for interview %d: %v", interviewID, err)
		} else {
			log.Printf("[TG-BOT] Evaluation completed for interview %d", interviewID)
			session.mu.Lock()
			session.State = StateCompleted
			session.mu.Unlock()
		}
		b.sessions.Delete(chatID)
	}()
}
