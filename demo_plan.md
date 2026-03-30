# inVision U — Demo Plan (Decentrathon 5.0)

## Demo Flow (5-7 minutes)

### 1. Problem Statement (30s)
- inVision U receives hundreds of applications
- Manual screening is slow, biased, and inconsistent
- Need: fair, fast, scalable evaluation with human oversight

### 2. Public Application Form (1 min)
- Show the apply page at `/apply`
- Fill out a sample application with essay
- Point out: English-only requirement, disability accommodation field
- Submit and show success message

### 3. Admin Dashboard (1 min)
- Login as `admin@invisionu.kz` / `admin123`
- Show dashboard stats: total candidates, score distribution, categories
- Show candidate list with filters and sorting

### 4. AI Analysis — Stage 1 (1.5 min)
- Select a pending candidate
- Show dual AI provider selector: Gemini (fast/cloud) vs Ollama (privacy/local)
- Run single analysis — show real-time scoring
- Show results: 5 dimension scores, AI-generated risk, summary, strengths/red flags
- Point out: depersonalized prompts (no PII sent to AI)
- Run batch analysis on remaining candidates (batch size = 5)
- Show stop button for batch cancellation

### 5. Committee Decision (30s)
- Make a decision on a candidate (shortlist/reject/waitlist)
- Show audit trail with full timestamp and user email

### 6. Telegram Interview — Stage 2 (1.5 min)
- Show a Strong Recommend candidate (score >= 65)
- Generate Telegram invite link
- Show the bot conversation flow:
  - English-only interview
  - STAR method questions across topics (leadership, motivation, growth, vision)
  - Voice message support with Whisper transcription
  - Verification questions from essay facts
- Show interview results: 5 scores + anti-cheat signals + combined score (60/40 split)
- Show full transcript in the dashboard

### 7. Key Differentiators (30s)
- Two-stage pipeline: essay + interview
- Anti-cheat: AI detection + response timing + style analysis + fact verification
- Privacy mode: fully local evaluation with Ollama
- Disability accommodation: hidden from Stage 1, accommodated in Stage 2
- Depersonalized: no bias from name, city, age, gender

## Technical Highlights to Mention
- Go backend with PostgreSQL, Gin framework
- Next.js 14 with App Router, TypeScript, Tailwind, shadcn/ui
- Dual AI: Gemini 2.5 Flash + Ollama/mistral:7b
- Telegram Bot API with long polling + session recovery
- OpenAI Whisper for voice transcription
- JWT auth, CORS, rate limiting
- Docker Compose for one-command setup

## Backup Plan
- If Gemini API is down: switch to Ollama provider
- If Telegram bot is not configured: show pre-recorded interview screenshots
- If DB is empty: auto-seeded 25 demo candidates with admin user
