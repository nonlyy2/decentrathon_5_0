# inVision U — AI Admissions Platform

AI-powered two-stage candidate screening platform for inVision U (inDrive's 100% scholarship university in Kazakhstan), built for **Decentrathon 5.0**.

## Overview

inVision U receives hundreds of scholarship applications. This platform automates screening through a two-stage pipeline:

1. **Stage 1 — Essay Analysis**: AI evaluates written applications across 5 dimensions, detects AI-generated content, and categorizes candidates
2. **Stage 2 — Telegram Interview**: AI-powered behavioral interview bot using STAR method with voice support, anti-cheat detection, and automated evaluation

The admissions committee reviews AI recommendations and makes final decisions through the dashboard.

### Key Features

- **Public application portal** — candidates submit essays and background info (English required)
- **Dual AI providers** — Gemini 2.5 Flash (cloud/fast) or Ollama/mistral:7b (local/privacy)
- **Stage 1: AI essay screening** — scores across 5 criteria with AI-writing detection
- **Stage 2: Telegram AI interview** — STAR-method behavioral interview with voice transcription
- **Anti-cheat system** — response timing analysis, style shift detection, essay fact verification
- **Disability accommodation** — candidates declare needs; hidden from Stage 1, visible to Stage 2
- **Depersonalized evaluation** — PII (name, email, age, city) excluded from AI prompts
- **Batch analysis** — analyze up to 5 candidates per batch with cancellation support
- **Human-in-the-loop** — AI recommends, committee decides (shortlist / waitlist / reject)
- **Decision audit trail** — full datetime + user email logged for every decision
- **Dashboard** — real-time stats, score distribution, category breakdown
- **Multilingual UI** — English, Russian, Kazakh interface

### Scoring Model

**Stage 1 (Essay — 60% of combined score)**

| Dimension | Weight | Description |
|-----------|--------|-------------|
| Leadership | 25% | Evidence of initiative and impact |
| Motivation | 25% | Clarity of purpose and alignment with mission |
| Growth | 20% | Learning mindset and resilience |
| Vision | 15% | Concrete goals and societal impact |
| Communication | 15% | Clarity, depth, and authenticity of writing |

**Stage 2 (Interview — 40% of combined score)**

| Dimension | Description |
|-----------|-------------|
| Leadership | Initiative and team influence demonstrated in conversation |
| Grit | Perseverance and resilience under challenge |
| Authenticity | Consistency with essay, genuine personal voice |
| Motivation | Depth of purpose beyond surface-level answers |
| Vision | Clarity of future plans and impact awareness |

**Categories:** 80-100 Strong Recommend, 65-79 Recommend, 50-64 Borderline, 0-49 Not Recommended

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go 1.22 + Gin |
| Database | PostgreSQL 16 (pgx connection pool) |
| AI (Cloud) | Google Gemini 2.5 Flash |
| AI (Local) | Ollama + mistral:7b |
| Voice | OpenAI Whisper API |
| Telegram | go-telegram-bot-api v5 (long polling) |
| Frontend | Next.js 14 (App Router) + TypeScript + Tailwind CSS |
| UI Components | shadcn/ui |
| Charts | Recharts |
| Auth | JWT (72h expiry) + bcrypt |

## Project Structure

```
decentrathon/
├── backend/
│   ├── cmd/server/main.go              # Entry point
│   ├── internal/
│   │   ├── config/                     # Env config
│   │   ├── database/                   # DB pool + migrations
│   │   ├── gemini/                     # Gemini client, prompt, parser
│   │   ├── ollama/                     # Ollama client + prompt
│   │   ├── handlers/                   # HTTP handlers (candidates, analysis, decisions, interview)
│   │   ├── middleware/                 # Auth + CORS
│   │   ├── models/                     # Data models (candidate, analysis, interview)
│   │   ├── seed/                       # Admin user + 25 demo candidates
│   │   └── telegram_bot/              # Stage 2 interview bot
│   │       ├── bot.go                  # Bot lifecycle, long polling, session recovery
│   │       ├── handler.go              # Message routing, /start deep link, interview flow
│   │       ├── interview.go            # LLM question generation (STAR method)
│   │       ├── evaluator.go            # Post-interview LLM evaluation
│   │       ├── anticheat.go            # Response timing, style shift detection
│   │       ├── voice.go                # Whisper API transcription
│   │       └── state.go                # Interview state machine
│   ├── .env.example                    # Environment template
│   ├── Dockerfile
│   └── go.mod
├── frontend/
│   ├── app/
│   │   ├── (dashboard)/                # Protected admin pages
│   │   │   ├── dashboard/              # Stats overview
│   │   │   ├── candidates/             # Candidate list + detail
│   │   │   └── analytics/              # Analytics view
│   │   ├── apply/                      # Public application form
│   │   └── page.tsx                    # Login page
│   ├── components/                     # Shared UI components
│   ├── lib/                            # API client, auth, hooks, types, i18n
│   └── .env.local
├── docker-compose.yml
└── .gitignore
```

## Quick Start

### Prerequisites

- Go 1.22+
- Node.js 20+
- PostgreSQL 16
- Gemini API key ([aistudio.google.com](https://aistudio.google.com)) OR Ollama installed locally

### 1. Database

```bash
createdb invisionu
```

### 2. Backend

```bash
cd backend
cp .env.example .env
# Edit .env: set GEMINI_API_KEY (or AI_PROVIDER=ollama)

go mod download
go run ./cmd/server
```

Backend starts at `http://localhost:8080`. Default admin: `admin@invisionu.kz` / `admin123`

### 3. Frontend

```bash
cd frontend
npm install
npm run dev
```

Frontend starts at `http://localhost:3000`.

### 4. Telegram Bot (optional)

1. Create a bot via [@BotFather](https://t.me/BotFather)
2. Set `TELEGRAM_BOT_TOKEN` in `.env`
3. For voice support: get an OpenAI API key and set `WHISPER_API_KEY`
4. Restart the backend — bot starts automatically

### Docker

```bash
docker compose up -d
```

### Environment Variables

**Backend (`backend/.env`)**

```env
DATABASE_URL=postgres://postgres:postgres@localhost:5432/invisionu?sslmode=disable
PORT=8080
JWT_SECRET=change-me-in-production
GEMINI_API_KEY=your-gemini-key
AI_PROVIDER=gemini                    # "gemini" or "ollama"
OLLAMA_URL=http://localhost:11434
OLLAMA_MODEL=mistral:7b
TELEGRAM_BOT_TOKEN=                   # Optional: Stage 2 interview bot
WHISPER_API_KEY=                      # Optional: voice message transcription
ALLOW_ORIGINS=http://localhost:3000
```

## API Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/api/apply` | — | Submit application |
| POST | `/api/auth/login` | — | Committee login |
| GET | `/api/candidates` | JWT | List with filters & sort |
| GET | `/api/candidates/:id` | JWT | Detail + analysis + decisions |
| POST | `/api/candidates/:id/analyze` | JWT | Trigger AI analysis |
| POST | `/api/candidates/:id/decision` | JWT | Record committee decision |
| POST | `/api/analyze-all` | JWT | Batch analyze pending |
| POST | `/api/analyze-all/stop` | JWT | Cancel batch analysis |
| GET | `/api/analyze-all/status` | JWT | Batch progress |
| POST | `/api/candidates/:id/telegram-invite` | JWT | Generate Telegram interview link |
| GET | `/api/candidates/:id/interview` | JWT | Interview status + results |
| GET | `/api/candidates/:id/interview/messages` | JWT | Interview transcript |
| GET | `/api/stats` | JWT | Dashboard statistics |
| GET | `/api/ai-providers` | JWT | Available AI providers |

## Team

Built for Decentrathon 5.0

- **Assylkhan** — Go backend, AI integration, database, Telegram bot
- **Zhanibek** — Next.js frontend, UI/UX
