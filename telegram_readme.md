Архитектура: Telegram AI Interviewer (Stage 2)
Общая схема пайплайна

Stage 1 (существующий)              Stage 2 (новый)
┌──────────────┐                   ┌──────────────────────┐
│  /apply form │──→ Essay Analysis │  Telegram Bot        │
│  candidates  │   (Gemini/Ollama) │  AI-интервью (STAR)  │
│  table       │──→ analyses table │  Voice + Text        │
└──────┬───────┘                   └──────────┬───────────┘
       │                                      │
       │    ┌─────────────────────┐           │
       └───→│ Dashboard (frontend)│←──────────┘
            │ Combined Score View │
            └─────────────────────┘
1. Связка с БД и дашбордом
Новые таблицы

-- Инвайт-токены для deep link матчинга
CREATE TABLE telegram_invites (
    id SERIAL PRIMARY KEY,
    candidate_id INTEGER UNIQUE NOT NULL REFERENCES candidates(id) ON DELETE CASCADE,
    token UUID NOT NULL UNIQUE DEFAULT gen_random_uuid(),
    telegram_chat_id BIGINT,          -- заполняется при /start
    status VARCHAR(20) DEFAULT 'pending'
        CHECK (status IN ('pending','linked','interview_active','completed','expired')),
    created_at TIMESTAMP DEFAULT NOW(),
    linked_at TIMESTAMP,
    expires_at TIMESTAMP DEFAULT NOW() + INTERVAL '7 days'
);

-- Сессия интервью
CREATE TABLE interviews (
    id SERIAL PRIMARY KEY,
    candidate_id INTEGER UNIQUE NOT NULL REFERENCES candidates(id) ON DELETE CASCADE,
    telegram_chat_id BIGINT NOT NULL,
    status VARCHAR(20) DEFAULT 'active'
        CHECK (status IN ('active','completed','abandoned','timeout')),
    current_topic VARCHAR(30),        -- leadership|motivation|growth|vision|communication
    questions_asked INTEGER DEFAULT 0,
    started_at TIMESTAMP DEFAULT NOW(),
    completed_at TIMESTAMP,
    essay_summary TEXT,               -- краткая сводка из Stage 1 для контекст-чекинга
    conversation_context JSONB DEFAULT '[]'  -- полная история для LLM
);

-- Каждое сообщение (для аудита и анализа)
CREATE TABLE interview_messages (
    id SERIAL PRIMARY KEY,
    interview_id INTEGER NOT NULL REFERENCES interviews(id) ON DELETE CASCADE,
    role VARCHAR(10) NOT NULL CHECK (role IN ('bot','candidate')),
    content TEXT NOT NULL,
    message_type VARCHAR(10) DEFAULT 'text' CHECK (message_type IN ('text','voice')),
    voice_duration_sec INTEGER,       -- длительность голосового
    response_time_sec INTEGER,        -- время ответа кандидата
    created_at TIMESTAMP DEFAULT NOW()
);

-- Результат интервью (аналог analyses для Stage 2)
CREATE TABLE interview_analyses (
    id SERIAL PRIMARY KEY,
    interview_id INTEGER NOT NULL REFERENCES interviews(id) ON DELETE CASCADE,
    candidate_id INTEGER UNIQUE NOT NULL REFERENCES candidates(id) ON DELETE CASCADE,
    score_leadership INTEGER CHECK (score_leadership BETWEEN 0 AND 100),
    score_grit INTEGER CHECK (score_grit BETWEEN 0 AND 100),
    score_authenticity INTEGER CHECK (score_authenticity BETWEEN 0 AND 100),
    score_motivation INTEGER CHECK (score_motivation BETWEEN 0 AND 100),
    score_vision INTEGER CHECK (score_vision BETWEEN 0 AND 100),
    final_score NUMERIC(5,2) NOT NULL,
    category VARCHAR(30) NOT NULL,
    consistency_score INTEGER DEFAULT 0,  -- совпадение с эссе (0-100)
    style_match_score INTEGER DEFAULT 0,  -- стилистическое совпадение (0-100)
    suspicion_flags JSONB DEFAULT '[]',   -- ["too_fast","style_shift","copy_paste"]
    summary TEXT,
    strengths TEXT[],
    concerns TEXT[],
    analyzed_at TIMESTAMP DEFAULT NOW(),
    model_used VARCHAR(50)
);
Обновление candidates

ALTER TABLE candidates ADD COLUMN IF NOT EXISTS telegram_chat_id BIGINT;
ALTER TABLE candidates ADD COLUMN IF NOT EXISTS interview_status VARCHAR(20) 
    DEFAULT 'not_invited' 
    CHECK (interview_status IN ('not_invited','invited','in_progress','completed'));
ALTER TABLE candidates ADD COLUMN IF NOT EXISTS combined_score NUMERIC(5,2);
combined_score = weighted average Stage 1 (60%) + Stage 2 (40%). Считается после завершения интервью.

Deep Link Flow

1. Админ нажимает "Send Telegram Invite" на дашборде
2. Backend: INSERT INTO telegram_invites (candidate_id) → получает UUID token
3. Backend генерирует ссылку: https://t.me/{bot_username}?start={token}
4. Фронтенд показывает ссылку + кнопка "Copy Link" (или отправка напрямую)
5. Кандидат открывает ссылку → Telegram → /start {token}
6. Бот: SELECT candidate_id FROM telegram_invites WHERE token = $1
7. Бот: UPDATE telegram_invites SET telegram_chat_id = chat_id, status = 'linked'
8. Бот: UPDATE candidates SET telegram_chat_id = chat_id
9. Начинается интервью
Сидеры
Обновить seedCandidate struct — добавить Phone и Telegram:


type seedCandidate struct {
    // ...existing fields...
    Phone    *string
    Telegram *string
}
Добавить реалистичные данные всем 20+ кандидатам:


Phone: ptr("+7 707 123 4567"), Telegram: ptr("@aigerim_s"),
Phone: ptr("+7 701 987 6543"), Telegram: ptr("@daulet_kz"),
// ... и так далее
2. Анти-чит система
Многослойная верификация

┌─────────────────────────────────────────────────────┐
│                  ANTI-CHEAT LAYERS                   │
├─────────────────────────────────────────────────────┤
│ Layer 1: Response Timing                            │
│   - Avg response time per message                   │
│   - Flag if < 3 sec for complex question (copy-paste)│
│   - Flag if > 600 sec (getting external help)       │
│   - Store per-message response_time_sec             │
│                                                     │
│ Layer 2: Style Consistency                          │
│   - Compare vocabulary complexity: essay vs chat    │
│   - Check for sudden formality shifts mid-dialog    │
│   - Flag if chat uses rare words never in essay     │
│   - Measure avg sentence length variance            │
│                                                     │
│ Layer 3: Content Verification                       │
│   - Ask about specific details from their essay     │
│   - "You mentioned [X] in your essay — tell me more"│
│   - Cross-reference facts (dates, names, places)    │
│   - Inconsistency = red flag                        │
│                                                     │
│ Layer 4: Behavioral Signals                         │
│   - Repetitive phrasing (ChatGPT patterns)          │
│   - Overly structured responses in casual chat      │
│   - Lack of filler words / hedging in voice msgs    │
│   - Voice vs text consistency if mixing modes       │
│                                                     │
│ Layer 5: Trap Questions (subtle)                    │
│   - Paraphrase something from essay with a mistake  │
│   - "You wrote that you founded X in 2023" (if 2024)│
│   - Genuine author corrects; cheater agrees          │
└─────────────────────────────────────────────────────┘
Реализация: все эти сигналы собираются во время диалога и передаются в финальный LLM-промпт оценки. Бот не принимает решение в реальном времени — только собирает данные. Финальный анализ — после завершения интервью.

3. Voice Recognition Pipeline

Candidate sends voice (.ogg) in Telegram
         │
         ▼
┌─────────────────────┐
│ Telegram Bot        │
│ Download .ogg file  │
│ via Bot API         │
└────────┬────────────┘
         │
         ▼
┌─────────────────────┐
│ Transcription       │
│ Option A: Whisper   │
│   via Ollama/local  │
│ Option B: Google    │
│   Speech-to-Text    │
│ Option C: OpenAI    │
│   Whisper API       │
└────────┬────────────┘
         │ text
         ▼
┌─────────────────────┐
│ LLM (Gemini/Ollama) │
│ Process as regular   │
│ text message with    │
│ [VOICE] tag          │
└─────────────────────┘
Рекомендация: OpenAI Whisper API
Почему:

Казахский и русский язык — Whisper large-v3 отлично поддерживает оба
$0.006/мин — дешево для интервью (10-15 мин = ~$0.09)
Простая интеграция — один HTTP-запрос
Альтернатива: локальный whisper.cpp через Ollama, но качество хуже для KZ/RU

// internal/speech/whisper.go
type Transcriber interface {
    Transcribe(ctx context.Context, audioData []byte, format string) (string, error)
}

type WhisperClient struct {
    apiKey     string
    httpClient *http.Client
}

func (w *WhisperClient) Transcribe(ctx context.Context, audioData []byte, format string) (string, error) {
    // POST https://api.openai.com/v1/audio/transcriptions
    // model: "whisper-1"
    // file: audioData
    // language: auto-detect (or "ru"/"kk")
}
Голосовые сообщения помечаются в interview_messages.message_type = 'voice' и хранятся как транскрибированный текст. В контексте LLM они маркируются [VOICE MESSAGE] — это важно для оценки естественности.

4. State Management и диалоговый движок
Структура пакета

backend/internal/telegram_bot/
├── bot.go              # Telegram Bot init, webhook/polling, message router
├── handler.go          # Обработка /start, текстовых, голосовых сообщений
├── interview.go        # STAR-движок: генерация вопросов, фоллоу-апы
├── state.go            # State machine: этапы интервью
├── evaluator.go        # Финальная оценка диалога через LLM
├── voice.go            # Download OGG + transcription
└── anticheat.go        # Сбор сигналов: timing, style, traps
State Machine

                    /start {token}
                         │
                         ▼
                  ┌──────────────┐
                  │   GREETING   │  Приветствие, объяснение формата
                  │              │  "Привет! Я ИИ-интервьюер inVision U..."
                  └──────┬───────┘
                         │
                         ▼
                  ┌──────────────┐
                  │  WARM_UP     │  1-2 лёгких вопроса для раскрепощения
                  │              │  "Расскажи, чем ты сейчас увлекаешься?"
                  └──────┬───────┘
                         │
              ┌──────────┼──────────┐
              ▼          ▼          ▼
        ┌──────────┐┌──────────┐┌──────────┐
        │LEADERSHIP││MOTIVATION││  GROWTH  │  Основные блоки
        │ 2-3 Q's  ││ 2-3 Q's  ││ 2-3 Q's │  по методу STAR
        └────┬─────┘└────┬─────┘└────┬─────┘
             │           │           │
              ┌──────────┼──────────┐
              ▼          ▼          ▼
        ┌──────────┐┌──────────┐
        │  VISION  ││ VERIFY   │  Верификация + финал
        │ 1-2 Q's  ││ 1-2 trap │
        └────┬─────┘└────┬─────┘
             │           │
             └─────┬─────┘
                   ▼
            ┌──────────────┐
            │   CLOSING    │  "Спасибо! Результаты скоро будут."
            └──────┬───────┘
                   │
                   ▼
            ┌──────────────┐
            │  EVALUATING  │  Async: отправка всего диалога в LLM
            │  (background)│  для финальной оценки
            └──────────────┘
Контекст беседы (JSONB в interviews.conversation_context)

[
  {
    "role": "bot",
    "content": "Расскажи о ситуации, когда тебе пришлось возглавить команду.",
    "topic": "leadership",
    "question_type": "star_situation",
    "timestamp": "2026-03-29T10:05:00Z"
  },
  {
    "role": "candidate",
    "content": "В 10 классе я организовал хакатон...",
    "message_type": "voice",
    "response_time_sec": 45,
    "timestamp": "2026-03-29T10:05:45Z"
  },
  {
    "role": "bot",
    "content": "Интересно! А какой был главный вызов в организации?",
    "topic": "leadership",
    "question_type": "star_action_followup",
    "timestamp": "2026-03-29T10:05:46Z"
  }
]
STAR-движок (динамическая генерация вопросов)
Бот НЕ берёт вопросы из жёсткого списка. Вместо этого:


func (e *InterviewEngine) GenerateNextQuestion(ctx context.Context, interview *Interview) (string, error) {
    // 1. Собрать conversation_context
    // 2. Определить текущий topic и сколько вопросов задано
    // 3. Сформировать промпт для LLM:
    
    systemPrompt := `You are an AI interviewer for inVision U scholarship.
You are conducting a behavioral interview using the STAR method.
Current topic: {topic}
Questions asked on this topic: {count}
Candidate's essay summary: {essay_summary}

Based on the conversation so far, generate the next natural follow-up question.
- If candidate gave a Situation, ask about Task/Action
- If candidate gave Action, ask about Result/Impact
- If answer was vague, probe deeper with "Can you give a specific example?"
- If topic is covered (2-3 good answers), move to next topic
- Keep tone warm and conversational, like a real interviewer
- Respond in the same language the candidate uses (Russian/Kazakh/English)

Return JSON: {"question": "...", "question_type": "star_situation|star_task|star_action|star_result|followup|probe|transition", "move_to_next_topic": false}`

    // 4. Вызвать LLM (Gemini preferred, Ollama fallback)
    // 5. Парсить JSON, обновить state
}
Финальная оценка (evaluator.go)
После завершения диалога (все топики покрыты или timeout):


func (e *Evaluator) EvaluateInterview(ctx context.Context, interview *Interview) (*InterviewAnalysis, error) {
    // 1. Собрать полный контекст: essay_summary + все сообщения
    // 2. Собрать античит-сигналы: timing, style metrics
    // 3. Отправить в LLM с промптом оценки:
    
    prompt := `Evaluate this interview transcript. The candidate previously wrote an essay (summary below).
    
ESSAY SUMMARY: {summary from Stage 1}

INTERVIEW TRANSCRIPT:
{full conversation}

BEHAVIORAL SIGNALS:
- Average response time: {avg_sec}s
- Fastest response: {min_sec}s  
- Messages flagged for timing: {count}
- Voice vs text ratio: {ratio}

Score each dimension 0-100:
- score_leadership: evidence of leading, initiative, responsibility
- score_grit: persistence, handling failure, resilience
- score_authenticity: genuineness, consistency with essay, natural responses
- score_motivation: depth of passion, personal connection to goals
- score_vision: clarity of future plans, understanding of impact

Also evaluate:
- consistency_score (0-100): how well interview answers match essay claims
- style_match_score (0-100): linguistic similarity between essay and chat
- suspicion_flags: array of concerns ["too_fast", "style_shift", "factual_inconsistency", "generic_answers"]`

    // 4. Парсить результат
    // 5. Записать в interview_analyses
    // 6. Вычислить combined_score и обновить candidates
}
5. Интеграция с фронтендом
Новые API эндпоинты

POST   /api/candidates/:id/telegram-invite    → генерация deep link
GET    /api/candidates/:id/interview          → статус и результат интервью
GET    /api/candidates/:id/interview/messages  → transcript (для комитета)
UI на дашборде (candidates/[id]/page.tsx)
В профиле кандидата добавляется секция "Stage 2: Interview":


┌─────────────────────────────────────────────┐
│ Stage 2: Telegram Interview                  │
├─────────────────────────────────────────────┤
│ Status: ● Completed                          │
│                                              │
│ [Score Radar: leadership/grit/authenticity/  │
│  motivation/vision]                          │
│                                              │
│ Interview Score: 78/100 — "Recommend"        │
│ Consistency with Essay: 85%                  │
│ Style Match: 72%                             │
│ Suspicion Flags: none                        │
│                                              │
│ [View Transcript]  [Re-evaluate]             │
├─────────────────────────────────────────────┤
│ Combined Score (60% essay + 40% interview)   │
│ ████████████████████░░░░  82.4 / 100         │
│ Category: Strong Recommend                   │
└─────────────────────────────────────────────┘
Кнопка "Send Telegram Invite" появляется только если:

Stage 1 analysis завершён
final_score >= 50 (прошёл первый фильтр)
Интервью ещё не начато
6. Технические решения и стек
Компонент	Решение	Почему
Telegram Bot	go-telegram-bot-api/telegram-bot-api v5	Самая зрелая Go-библиотека, long polling
Voice Transcription	OpenAI Whisper API	Лучшее качество KZ/RU, $0.006/мин
LLM для диалога	Gemini 2.5 Flash (primary)	Быстрый, дешёвый, хороший контекст
State Storage	PostgreSQL (JSONB)	Уже есть, надёжно, не нужен Redis
Bot Deployment	Goroutine в том же бинарнике	Минимум инфры, общий пул БД
Deep Links	UUID v4 токены	Простой, безопасный матчинг
Конфигурация (новые env vars)

TELEGRAM_BOT_TOKEN=123456:ABC-DEF...
WHISPER_API_KEY=sk-...              # OpenAI API key for Whisper
WHISPER_PROVIDER=openai             # openai | local
INTERVIEW_TIMEOUT_MIN=30            # макс длительность интервью
INTERVIEW_MIN_QUESTIONS=8           # мин. вопросов до завершения
INTERVIEW_MAX_QUESTIONS=15          # макс. вопросов
Запуск бота
В main.go добавляется:


if cfg.TelegramBotToken != "" {
    bot := telegram_bot.New(cfg, pool, providers, textGens, defaultProvider)
    go bot.Start(context.Background())  // отдельная горутина
    log.Printf("Telegram bot started (@%s)", bot.Username())
}
7. Порядок реализации (фазы)
Фаза 1 — Фундамент (DB + Deep Link + Bot skeleton)

Миграции: новые таблицы
Сидеры: phone + telegram для всех кандидатов
POST /candidates/:id/telegram-invite endpoint
Bot: /start {token} → матчинг → приветствие
Фронтенд: кнопка "Send Telegram Invite"
Фаза 2 — Диалоговый движок

State machine (greeting → topics → closing)
STAR question generation через LLM
Обработка текстовых сообщений
conversation_context в JSONB
Античит: timing tracking
Фаза 3 — Voice Pipeline

Download OGG из Telegram
Whisper API интеграция
Транскрипция + маркировка [VOICE]
Фаза 4 — Оценка и дашборд

Финальный evaluator (LLM-based)
interview_analyses запись
combined_score расчёт
UI: interview results в профиле кандидата
Transcript viewer