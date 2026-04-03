# inVision U — Intelligent Admissions Platform

> ML/Agent-powered two-stage candidate screening platform for **inVision U** (inDrive's 100%-scholarship university in Kazakhstan).
> Built for **Decentrathon 5.0** | Track: AI inDrive

**Live demo:** https://front-production-6189.up.railway.app/  
Deployed on [Railway](https://railway.app/).

---

## Table of Contents

1. [Problem Statement](#problem-statement)
2. [Solution Overview](#solution-overview)
3. [Killer Features](#killer-features)
4. [Core Features](#core-features)
5. [Usage Scenarios](#usage-scenarios)
6. [Roles & Permissions](#roles--permissions)
7. [Architecture](#architecture)
8. [Tech Stack](#tech-stack)
9. [Project Structure](#project-structure)
10. [Database Schema](#database-schema)
11. [Scoring Model](#scoring-model)
12. [API Reference](#api-reference)
13. [Getting Started](#getting-started)
14. [Environment Variables](#environment-variables)
15. [Deployment (Railway)](#deployment-railway)
16. [Dependencies](#dependencies)
17. [Limitations & Notes](#limitations--notes)
18. [Team](#team)

---

## Problem Statement

InVision U is an innovative university launched by inDrive in Kazakhstan, offering 100% scholarships to future leaders, entrepreneurs and project creators. The admissions committee manually reviews hundreds of applications, essays, and video presentations every cycle. This creates several issues:

- Talented candidates who lack self-promotion skills are overlooked — the process sees the application, not the person.
- Standard forms and essays fail to surface early leadership potential signals.
- Generative AI blurs the authentic voice in essays, making text-based evaluation less reliable.
- Manual screening cannot scale — quality inevitably drops as application volume grows.

---

## Solution Overview

A data-driven, two-stage candidate screening system that **assists** (not replaces) the admissions committee:

1. **Stage 1 — Essay & Application Analysis**: An ML model evaluates written applications across 5 weighted dimensions, detects AI-generated content, flags incomplete submissions, and categorizes candidates.
2. **Stage 2 — Telegram Chat-Interview**: An agent-driven behavioral interview bot conducts a STAR-method interview via Telegram (text + voice), with anti-cheat detection and automated evaluation.

Human managers always have the final say. The model provides transparent, citation-backed recommendations — the committee reviews and decides.

---

## Killer Features

| Feature | Description |
|---------|-------------|
| **CSV Export & Import with Analysis** | Export/import the full candidate database including ML analysis results and committee decisions. Supports EN/RU/KK column headers. |
| **Telegram Agent Chat-Interview (Stage 2)** | Fully automated behavioral interview via Telegram bot with voice message support. 8-15 adaptive questions using STAR methodology. |
| **Transparent ML Analysis with Citations** | Every score dimension includes direct quotes from the candidate's essay/achievements as evidence. No black-box decisions. |
| **Advanced User Management & Role Authorities** | 6 roles with strict hierarchy (Superadmin > Tech-Admin > Auditor > Manager). Role-based access control on every endpoint. |
| **100 Candidates in Under 3 Minutes** | Batch analysis with 5 concurrent threads + rate-limit-aware retry. Analyze hundreds of candidates in minutes, not days. |
| **ML Selection Assistant** | Ask the model to pick the best N candidates from M and it explains *why* it chose each one — backed by scoring data. |
| **Auto-Accept Top N** | One click to shortlist the top N candidates by score. Eliminates manual status updates for clear frontrunners. |
| **Non-English Content Detection** | If a candidate submits text not in English, it's flagged next to the ML score. >50% non-English triggers an incomplete flag. |
| **ML Recommended Major** | If a candidate's strengths suggest a better fit for a different program, the model flags it with a reason (based on essay + achievements). |
| **Multi-Manager Consensus Required** | 4+ committee members must vote before a decision is finalized — eliminates individual bias and subjective shortlisting. |
| **Similar Candidates & Comparison** | View similar candidates under any profile and compare them side-by-side across all dimensions. |
| **Dimension Distribution Dashboard** | See how candidates distribute across each scoring criterion (Leadership, Motivation, Growth, Vision, Communication) with interactive charts. |
| **Accessibility Mode** | Toggle for visually impaired users: 1.5x larger fonts + high-contrast black-and-white theme. |

---

## Core Features

| Feature | Description |
|---------|-------------|
| Dual ML Providers | Gemini 2.5 Flash (cloud, fast) or Ollama/Mistral:7b (local, private). Switch per-request. |
| Human Override | Committee decisions always override ML recommendations. |
| Leadership-Focused Interview | Stage 2 interview designed to surface leadership potential, not just technical knowledge. |
| YouTube Transcript Integration | Video presentation transcripts are extracted and fed into the ML analysis alongside the essay. |
| Bias Removal | Personal info (name, email, age, city) is excluded from ML prompts — evaluation is content-only. |
| Extended Dashboard | Real-time stats, score distributions, category breakdowns, admissions funnel, and prediction-ready analytics. |
| Advanced Score Breakdowns | Per-dimension scores with weighted final score, explanations, key strengths, and red flags. |
| Cheat & Incompleteness Detection | AI-generated content scoring (0-100), response timing analysis, style shift detection, essay fact verification. |
| Multi-Manager Final Decision | Multiple managers vote independently; consensus threshold (4+) determines outcome. |
| Filters & Sorting | Filter by score range, age range, major, status. Sort by any column. Full-text search on essays. |
| Bulk Status Changes | Select multiple candidates via checkboxes and apply a decision (shortlist/waitlist/reject) in one action. |
| Multilingual Interface | Full UI in English, Russian, and Kazakh. Language persists across sessions. |
| Light & Dark Theme | Toggle between light and dark mode. Respects system preference. |

---

## Usage Scenarios

### Scenario 1: Large-Scale Application Review

1. **Import candidates** — Upload a CSV with hundreds of applicant records.
2. **Batch analyze** — Click "Analyze All Pending." The system processes ~100 candidates in under 3 minutes using Gemini.
3. **Review dashboard** — Check score distributions, dimension charts, and the admissions funnel.
4. **Filter & sort** — Narrow down to "Recommend" and "Strong Recommend" categories.
5. **Auto-accept top N** — Shortlist the top 30 candidates automatically.
6. **Committee review** — 4+ managers independently vote on borderline cases.
7. **Export results** — Download the full dataset with scores and decisions as CSV.

### Scenario 2: Deep Candidate Evaluation

1. **Open candidate profile** — View essay, achievements, motivation, YouTube video, and ML analysis.
2. **Read ML citations** — Each dimension score includes direct quotes explaining why the candidate scored that way.
3. **Check ML major recommendation** — If the model suggests a different major, review the reasoning.
4. **Compare with similar candidates** — View candidates with scores within 3% and compare side-by-side.
5. **Use ML Selection Assistant** — Ask the model to pick the best 5 out of 20 borderline candidates with explanations.
6. **Vote & comment** — Record your decision with notes. See how other committee members voted.

### Scenario 3: Stage 2 Telegram Interview

1. **Invite candidate** — Generate a Telegram deep link (requires Stage 1 score >= 65).
2. **Candidate chats with bot** — The agent asks 8-15 adaptive STAR-method questions. Supports voice messages.
3. **Anti-cheat runs** — Response timing analysis, style consistency checks, essay fact verification.
4. **ML evaluates** — Scores on Leadership, Grit, Authenticity, Motivation, Vision + suspicion flags.
5. **Combined score** — 60% essay + 40% interview = final combined score.
6. **Review transcript** — Full interview conversation available in the dashboard.

---

## Roles & Permissions

The system has 6 roles with a strict hierarchy (higher level = more permissions):

| Role | Level | Permissions |
|------|-------|-------------|
| **Superadmin** | 4 | Full system access. Can delete users, assign any role, manage all data. |
| **Tech-Admin** | 3 | User management (create/edit/delete users except superadmins). Batch operations. Cannot assign superadmin role. |
| **Auditor** | 2 | Read-only access to all candidate data, analyses, and decisions. Cannot modify anything. |
| **Manager** | 1 | Core workflow: view candidates, trigger analyses, vote on decisions, leave comments. Cannot manage users. |
| **Admin** (legacy) | 4 | Same as Superadmin (backwards compatibility). |
| **Committee** (legacy) | 1 | Same as Manager (backwards compatibility). |

**Key rules:**
- Users can only edit/delete users of a strictly lower role level.
- Superadmins cannot delete themselves.
- 4+ independent manager votes are required for a decision to reach consensus.
- Auditors are restricted to GET requests only.

---

## Architecture

```
┌─────────────────────────────────────────────────────────────────────┐
│                         CLIENTS                                     │
│                                                                     │
│   ┌──────────────┐  ┌──────────────┐  ┌───────────────────────┐    │
│   │  Next.js SPA │  │ Telegram Bot │  │ Telegram Mini App     │    │
│   │  (Dashboard) │  │  (Interview) │  │ (Candidate Status)    │    │
│   └──────┬───────┘  └──────┬───────┘  └───────────┬───────────┘    │
└──────────┼─────────────────┼──────────────────────┼────────────────┘
           │ REST API        │ Long Polling          │ REST API
           ▼                 ▼                       ▼
┌─────────────────────────────────────────────────────────────────────┐
│                      GO BACKEND (Gin)                               │
│                                                                     │
│  ┌─────────┐ ┌───────────┐ ┌───────────┐ ┌─────────────────────┐  │
│  │  Auth   │ │ Candidate │ │ Analysis  │ │  Telegram Bot       │  │
│  │  (JWT)  │ │ Handlers  │ │ Pipeline  │ │  (Interview Agent)  │  │
│  └────┬────┘ └─────┬─────┘ └─────┬─────┘ └──────────┬──────────┘  │
│       │            │              │                   │             │
│  ┌────┴────────────┴──────────────┴───────────────────┴─────────┐  │
│  │                     MIDDLEWARE                                │  │
│  │  CORS · JWT Auth · Role Hierarchy · Auditor Read-Only        │  │
│  └──────────────────────────┬───────────────────────────────────┘  │
│                             │                                      │
│  ┌──────────────────────────┴───────────────────────────────────┐  │
│  │                     SERVICES                                  │  │
│  │                                                               │  │
│  │  ┌─────────┐  ┌────────┐  ┌──────────┐  ┌────────────────┐  │  │
│  │  │ Gemini  │  │ Ollama │  │ YouTube  │  │  Whisper/Alem  │  │  │
│  │  │ 2.5     │  │ Mistral│  │Transcript│  │  STT (Voice)   │  │  │
│  │  │ Flash   │  │ :7b    │  │ Fetcher  │  │                │  │  │
│  │  └─────────┘  └────────┘  └──────────┘  └────────────────┘  │  │
│  │                                                               │  │
│  │  ┌──────────┐  ┌───────────┐  ┌──────────┐  ┌────────────┐  │  │
│  │  │  Email   │  │   CSV     │  │  Photo   │  │ Anti-Cheat │  │  │
│  │  │  (SMTP)  │  │ Import/   │  │ Upload + │  │ (Timing,   │  │  │
│  │  │          │  │ Export    │  │ ML Check │  │  Style)     │  │  │
│  │  └──────────┘  └───────────┘  └──────────┘  └────────────┘  │  │
│  └───────────────────────────────────────────────────────────────┘  │
└────────────────────────────┬────────────────────────────────────────┘
                             │
                             ▼
                   ┌──────────────────┐
                   │  PostgreSQL 16   │
                   │  (pgx pool)      │
                   └──────────────────┘
```

---

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go 1.25 + Gin HTTP framework |
| Database | PostgreSQL 16 (pgx connection pool, max 10 conns) |
| ML (Cloud) | Google Gemini 2.5 Flash |
| ML (Local) | Ollama + Mistral:7b |
| Voice STT | OpenAI Whisper API / Alem STT API |
| Telegram | go-telegram-bot-api v5 (long polling) |
| Frontend | Next.js 14 (App Router) + TypeScript + Tailwind CSS |
| UI Components | shadcn/ui + Lucide icons |
| Charts | Recharts |
| Auth | JWT (72h expiry) + bcrypt |
| Deployment | Railway (PostgreSQL + Go backend + Next.js frontend) |
| Containerization | Docker multi-stage builds |

---

## Project Structure

```
decentrathon/
├── backend/
│   ├── cmd/server/main.go                 # Entry point, routing, flags (--seed, --force-seed)
│   ├── internal/
│   │   ├── config/config.go               # Environment variable loader
│   │   ├── database/
│   │   │   ├── db.go                      # PostgreSQL connection pool (pgx)
│   │   │   └── migrations.go             # Schema creation + index setup
│   │   ├── models/                        # Data models
│   │   │   ├── candidate.go               # Candidate, CandidateListItem
│   │   │   ├── analysis.go                # Stage 1 ML analysis
│   │   │   ├── interview.go               # Stage 2 interview + analysis
│   │   │   ├── decision.go                # Committee decisions
│   │   │   ├── user.go                    # User + roles
│   │   │   └── stats.go                   # Dashboard statistics
│   │   ├── handlers/                      # HTTP handlers
│   │   │   ├── candidate.go               # CRUD + search + filters
│   │   │   ├── analysis.go                # ML analysis trigger + batch
│   │   │   ├── decision.go                # Committee voting
│   │   │   ├── bulk.go                    # Bulk decisions, auto-accept, ML recommend
│   │   │   ├── interview.go               # Stage 2 management
│   │   │   ├── export.go                  # CSV export/import
│   │   │   ├── upload.go                  # Photo upload + ML detection
│   │   │   ├── comments.go                # Candidate comments
│   │   │   ├── users.go                   # User management (CRUD)
│   │   │   ├── auth.go                    # Login/register
│   │   │   ├── stats.go                   # Dashboard stats endpoint
│   │   │   ├── email.go                   # SMTP email service
│   │   │   └── tma.go                     # Telegram Mini App status
│   │   ├── gemini/                        # Gemini ML provider
│   │   │   ├── client.go                  # HTTP client + rate limiting
│   │   │   ├── analyzer.go                # Single + batch analysis
│   │   │   ├── prompt.go                  # System prompt + rubric
│   │   │   └── parser.go                  # JSON response parser
│   │   ├── ollama/                        # Ollama ML provider (local)
│   │   │   ├── client.go                  # HTTP client
│   │   │   ├── analyzer.go                # Analysis logic
│   │   │   └── prompt.go                  # Simplified prompt for local models
│   │   ├── youtube/
│   │   │   └── transcript.go             # YouTube URL validation + transcript fetching
│   │   ├── telegram_bot/                  # Stage 2 Interview Agent
│   │   │   ├── bot.go                     # Bot lifecycle, long polling, session recovery
│   │   │   ├── handler.go                 # Message routing, /start deep link
│   │   │   ├── interview.go               # LLM question generation (STAR method)
│   │   │   ├── evaluator.go               # Post-interview ML evaluation
│   │   │   ├── anticheat.go               # Response timing, style shift detection
│   │   │   ├── voice.go                   # Whisper/Alem STT transcription
│   │   │   └── state.go                   # Interview state machine
│   │   ├── middleware/
│   │   │   ├── auth.go                    # JWT + role hierarchy middleware
│   │   │   └── cors.go                    # CORS configuration
│   │   └── seed/
│   │       ├── admin.go                   # Default superadmin user
│   │       └── candidates.go             # Demo candidate data
│   ├── .env.example
│   ├── Dockerfile
│   └── go.mod
├── frontend/
│   ├── app/
│   │   ├── layout.tsx                     # Root layout (theme, a11y, toasts)
│   │   ├── page.tsx                       # Home — Login + Student FAQ
│   │   ├── apply/page.tsx                 # Public application form
│   │   ├── faq/page.tsx                   # FAQ & requirements page
│   │   ├── tma/page.tsx                   # Telegram Mini App status viewer
│   │   └── (dashboard)/                   # Protected admin area
│   │       ├── layout.tsx                 # Sidebar, providers, auth guard
│   │       ├── dashboard/page.tsx         # Analytics dashboard
│   │       ├── candidates/
│   │       │   ├── page.tsx               # Candidate list + bulk actions
│   │       │   └── [id]/page.tsx          # Candidate detail + ML analysis
│   │       ├── compare/page.tsx           # Side-by-side comparison
│   │       ├── users/page.tsx             # User management
│   │       ├── profile/page.tsx           # Current user profile
│   │       └── analytics/page.tsx         # → Redirects to dashboard
│   ├── components/                        # Shared UI components (shadcn/ui)
│   ├── lib/
│   │   ├── api.ts                         # Axios instance + interceptors
│   │   ├── auth.tsx                       # AuthContext (login/logout/JWT)
│   │   ├── theme.tsx                      # Light/dark theme provider
│   │   ├── accessibility.ts               # Accessibility mode toggle
│   │   ├── i18n.tsx                       # Translations (en/ru/kk, 150+ keys)
│   │   ├── aiProvider.tsx                 # Gemini/Ollama provider context
│   │   ├── hooks.ts                       # useFetch generic hook
│   │   ├── types.ts                       # TypeScript interfaces
│   │   └── utils.ts                       # Utility functions
│   ├── .env.local.example
│   ├── Dockerfile
│   └── package.json
├── docker-compose.yml
└── .gitignore
```

---

## Database Schema

### `candidates`

| Column | Type | Description |
|--------|------|-------------|
| id | SERIAL PK | Unique identifier |
| full_name | TEXT | Candidate's full name |
| email | TEXT (unique) | Email address |
| phone | TEXT | Phone number |
| telegram | TEXT | Telegram username |
| age | INT | Age |
| city | TEXT | City of residence |
| school | TEXT | School name |
| graduation_year | INT | Year of graduation |
| major | TEXT | Chosen major (Engineering / Tech / Society / Policy Reform / Art + Media) |
| achievements | TEXT | Academic and personal achievements |
| extracurriculars | TEXT | Extracurricular activities |
| essay | TEXT | Main essay |
| motivation_statement | TEXT | Motivation statement |
| disability | TEXT | Accessibility needs (optional, hidden from Stage 1 ML) |
| photo_url | TEXT | Uploaded photo path |
| photo_ai_flag | BOOLEAN | Whether photo detected as ML-generated |
| photo_ai_note | TEXT | ML detection details |
| youtube_url | TEXT | YouTube presentation video URL |
| youtube_transcript | TEXT | Extracted video transcript |
| youtube_url_valid | BOOLEAN | Whether URL is accessible |
| keywords | TEXT[] | ML-extracted keywords (GIN indexed) |
| status | TEXT | pending / analyzed / shortlisted / rejected / waitlisted |
| interview_status | TEXT | not_invited / invited / in_progress / completed |
| combined_score | FLOAT | Weighted Stage 1 (60%) + Stage 2 (40%) |
| created_at | TIMESTAMP | Application timestamp |

### `users`

| Column | Type | Description |
|--------|------|-------------|
| id | SERIAL PK | Unique identifier |
| email | TEXT (unique) | Login email |
| password_hash | TEXT | bcrypt hash |
| full_name | TEXT | Display name |
| role | TEXT | superadmin / tech-admin / auditor / manager / admin / committee |
| avatar_url | TEXT | Profile avatar |
| created_at | TIMESTAMP | Account creation time |

### `analyses` (Stage 1 ML Evaluation)

| Column | Type | Description |
|--------|------|-------------|
| id | SERIAL PK | Unique identifier |
| candidate_id | INT (unique FK) | One analysis per candidate |
| score_leadership | INT | 0-100, weight 25% |
| score_motivation | INT | 0-100, weight 25% |
| score_growth | INT | 0-100, weight 20% |
| score_vision | INT | 0-100, weight 15% |
| score_communication | INT | 0-100, weight 15% |
| final_score | FLOAT | Weighted average |
| category | TEXT | Strong Recommend / Recommend / Borderline / Not Recommended |
| ai_generated_score | INT | 0-100 probability of ML-generated essay |
| ai_generated_risk | TEXT | low / medium / high |
| incomplete_flag | BOOLEAN | True if >50% non-English content |
| explanation_leadership | TEXT | Score reasoning with citations |
| explanation_motivation | TEXT | Score reasoning with citations |
| explanation_growth | TEXT | Score reasoning with citations |
| explanation_vision | TEXT | Score reasoning with citations |
| explanation_communication | TEXT | Score reasoning with citations |
| summary | TEXT | 3-5 sentence overall assessment |
| key_strengths | TEXT[] | List of strengths |
| red_flags | TEXT[] | List of concerns |
| recommended_major | TEXT | ML-suggested major (may differ from chosen) |
| major_reason_note | TEXT | Explanation if recommendation differs |
| model_used | TEXT | gemini-2.5-flash / ollama/mistral:7b |
| analyzed_at | TIMESTAMP | Analysis timestamp |
| duration_ms | INT | Processing time |

### `interview_analyses` (Stage 2 ML Evaluation)

| Column | Type | Description |
|--------|------|-------------|
| id | SERIAL PK | Unique identifier |
| candidate_id | INT (unique FK) | One interview analysis per candidate |
| score_leadership | INT | 0-100 |
| score_grit | INT | 0-100 |
| score_authenticity | INT | 0-100 |
| score_motivation | INT | 0-100 |
| score_vision | INT | 0-100 |
| final_score | FLOAT | Weighted interview score |
| category | TEXT | Strong Recommend / Recommend / Borderline / Not Recommended |
| consistency_score | FLOAT | Stage 1 vs Stage 2 alignment |
| style_match_score | FLOAT | Language pattern analysis |
| suspicion_flags | TEXT[] | Anti-cheat signals |
| summary | TEXT | Interview assessment |
| strengths | TEXT[] | Demonstrated strengths |
| concerns | TEXT[] | Areas of concern |

### `committee_decisions`

| Column | Type | Description |
|--------|------|-------------|
| id | SERIAL PK | Unique identifier |
| candidate_id | INT FK | Target candidate |
| decision | TEXT | shortlist / reject / waitlist / review |
| notes | TEXT | Reviewer's notes |
| decided_by | INT FK | User who voted |
| decided_at | TIMESTAMP | Vote timestamp |

*Unique constraint on (candidate_id, decided_by) — one vote per user per candidate.*

### `telegram_invites`

| Column | Type | Description |
|--------|------|-------------|
| candidate_id | INT (unique FK) | Invited candidate |
| token | UUID (unique) | Deep link token |
| status | TEXT | pending / linked / interview_active / completed / expired |
| telegram_chat_id | BIGINT | Linked Telegram chat |
| expires_at | TIMESTAMP | Token expiry (7 days) |

### `interviews`

| Column | Type | Description |
|--------|------|-------------|
| id | SERIAL PK | Unique identifier |
| candidate_id | INT (unique FK) | Interviewed candidate |
| telegram_chat_id | BIGINT | Telegram chat ID |
| status | TEXT | active / completed / abandoned / timeout |
| language | TEXT | Detected language |
| current_topic | TEXT | Current interview topic |
| questions_asked | INT | Number of questions asked |
| conversation_context | JSONB | Full conversation state |
| started_at / completed_at | TIMESTAMP | Interview timestamps |

### `interview_messages`

| Column | Type | Description |
|--------|------|-------------|
| id | SERIAL PK | Unique identifier |
| interview_id | INT FK | Parent interview |
| role | TEXT | bot / candidate |
| content | TEXT | Message content |
| message_type | TEXT | text / voice |
| voice_duration_sec | FLOAT | Voice message duration |
| response_time_sec | FLOAT | Candidate response time |
| created_at | TIMESTAMP | Message timestamp |

### `comments`

| Column | Type | Description |
|--------|------|-------------|
| id | SERIAL PK | Unique identifier |
| candidate_id | INT FK | Target candidate |
| user_id | INT FK | Comment author |
| content | TEXT | Comment text |
| created_at | TIMESTAMP | Comment timestamp |

### `email_logs`

| Column | Type | Description |
|--------|------|-------------|
| id | SERIAL PK | Unique identifier |
| candidate_id | INT FK | Related candidate |
| recipient_email | TEXT | Recipient |
| subject | TEXT | Email subject |
| template | TEXT | Template used |
| status | TEXT | Send status |
| sent_at | TIMESTAMP | Send time |

---

## Scoring Model

### Stage 1 — Essay Analysis (60% of combined score)

| Dimension | Weight | Description |
|-----------|--------|-------------|
| Leadership | 25% | Evidence of initiative, impact, and taking responsibility |
| Motivation | 25% | Genuine passion, specific personal stories, unique perspective |
| Growth | 20% | Learning mindset, resilience, overcoming obstacles |
| Vision | 15% | Clarity of goals, realistic ambition, understanding of impact |
| Communication | 15% | Structure, coherent narrative, depth of expression |

### Stage 2 — Interview (40% of combined score)

| Dimension | Description |
|-----------|-------------|
| Leadership | Initiative and team influence demonstrated in conversation |
| Grit | Perseverance and resilience under challenge |
| Authenticity | Consistency with essay, genuine personal voice |
| Motivation | Depth of purpose beyond surface-level answers |
| Vision | Clarity of future plans and impact awareness |

### Categories

| Range | Category |
|-------|----------|
| 80-100 | Strong Recommend |
| 65-79 | Recommend |
| 50-64 | Borderline |
| 0-49 | Not Recommended |

### ML-Generated Content Detection

| Score | Interpretation |
|-------|---------------|
| 0-20 | Almost certainly human |
| 21-40 | Likely human with polished sections |
| 41-60 | Mixed signals |
| 61-80 | Likely ML-generated |
| 81-100 | Almost certainly ML-generated |

---

## API Reference

### Public Endpoints

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/apply` | Submit candidate application |
| POST | `/api/auth/login` | Login, returns JWT |
| POST | `/api/auth/register` | Register new user |
| GET | `/api/majors` | List available programs |
| GET | `/api/health` | Health check |
| GET | `/api/tma/status?token=` | Telegram Mini App candidate status |
| POST | `/api/candidates/:id/photo` | Upload candidate photo |

### Protected Endpoints (JWT Required)

**Candidates**

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/candidates` | List with filters, search, sort, pagination |
| POST | `/api/candidates` | Create candidate manually |
| GET | `/api/candidates/:id` | Full detail + analysis + decisions |
| PATCH | `/api/candidates/:id` | Update candidate info |
| DELETE | `/api/candidates/:id` | Delete candidate (cascades) |
| PATCH | `/api/candidates/:id/status` | Update status |
| GET | `/api/candidates/:id/similar` | Find similar candidates (within 3% score) |
| POST | `/api/candidates/:id/fetch-transcript` | Re-fetch YouTube transcript |

**ML Analysis**

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/candidates/:id/analyze` | Trigger ML analysis (?provider=gemini\|ollama) |
| GET | `/api/candidates/:id/analysis` | Get completed analysis |
| GET | `/api/candidates/:id/analysis-status` | Check async analysis progress |
| DELETE | `/api/candidates/:id/analysis` | Delete analysis |
| POST | `/api/analyze-all` | Batch analyze all pending candidates |
| POST | `/api/analyze-all/stop` | Stop batch analysis |
| GET | `/api/analyze-all/status` | Batch progress |
| GET | `/api/ai-providers` | List available ML providers |

**Committee Decisions**

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/candidates/:id/decision` | Vote on candidate |
| GET | `/api/candidates/:id/decisions` | Get all votes + consensus status |

**Bulk Operations**

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/candidates/bulk-decision` | Bulk apply decision |
| POST | `/api/candidates/auto-accept` | Auto-shortlist top N |
| POST | `/api/candidates/ai-recommend` | ML pick best N from M candidates |
| GET | `/api/candidates/export/csv` | Export CSV (?lang=en\|ru\|kk) |
| POST | `/api/candidates/import/csv` | Import candidates from CSV |

**Interview (Stage 2)**

| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/candidates/:id/telegram-invite` | Generate Telegram invite link |
| GET | `/api/candidates/:id/interview` | Interview status + results |
| GET | `/api/candidates/:id/interview/messages` | Full transcript |
| POST | `/api/candidates/:id/interview/evaluate` | Force evaluate interview |
| POST | `/api/candidates/:id/interview/re-evaluate` | Re-run evaluation |
| DELETE | `/api/candidates/:id/interview/analysis` | Delete interview analysis |
| POST | `/api/interviews/evaluate-all-pending` | Batch evaluate all pending |

**Users & Profile**

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/users` | List all users (tech-admin+) |
| GET | `/api/users/:id` | Get user detail |
| PATCH | `/api/users/:id` | Update user role/name |
| DELETE | `/api/users/:id` | Delete user (superadmin only) |
| GET | `/api/profile` | Current user profile |
| PATCH | `/api/profile` | Update own profile |

**Other**

| Method | Path | Description |
|--------|------|-------------|
| GET | `/api/stats` | Dashboard statistics |
| GET/POST/DELETE | `/api/candidates/:id/comments` | Candidate comments |
| POST | `/api/candidates/:id/email/shortlist` | Send shortlist email |

---

## Getting Started

### Prerequisites

- **Go** 1.22+ (or 1.25 for exact match)
- **Node.js** 20+
- **PostgreSQL** 16
- **Gemini API key** ([aistudio.google.com](https://aistudio.google.com/app/apikey)) — OR Ollama installed locally
- *(Optional)* Telegram Bot Token from [@BotFather](https://t.me/BotFather)
- *(Optional)* OpenAI API key for Whisper voice transcription

### Option 1: Docker Compose (Recommended)

```bash
# Clone the repository
git clone <repo-url>
cd decentrathon

# Set required env vars (or create .env in project root)
export GEMINI_API_KEY=your-gemini-key

# Start all services
docker compose up -d
```

Services:
- **Frontend**: http://localhost:3000
- **Backend**: http://localhost:8080
- **PostgreSQL**: localhost:5432

### Option 2: Manual Setup

#### 1. Database

```bash
# Create the database
createdb invisionu
```

#### 2. Backend

```bash
cd backend
cp .env.example .env
# Edit .env — at minimum set GEMINI_API_KEY

# Install dependencies and run
go mod download
go run ./cmd/server --seed
```

The `--seed` flag creates a default superadmin and demo candidates.

Backend starts at http://localhost:8080.

**Default admin credentials:**
- Email: `admin@invisionu.kz`
- Password: `admin123`

#### 3. Frontend

```bash
cd frontend
cp .env.local.example .env.local
# Edit .env.local if backend is not on localhost:8080

npm install
npm run dev
```

Frontend starts at http://localhost:3000.

#### 4. Telegram Bot (Optional)

1. Create a bot via [@BotFather](https://t.me/BotFather).
2. Set `TELEGRAM_BOT_TOKEN` in `backend/.env`.
3. For voice support: set `WHISPER_API_KEY` (OpenAI) or `ALEM_STT_API_KEY`.
4. Restart the backend — the bot starts automatically.

---

## Environment Variables

### Backend (`backend/.env`)

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `DATABASE_URL` | Yes | — | PostgreSQL connection string |
| `PORT` | No | 8080 | Server port |
| `JWT_SECRET` | Yes (prod) | dev-secret | JWT signing secret. Generate: `openssl rand -hex 32` |
| `ALLOW_ORIGINS` | No | * | CORS allowed origins (comma-separated) |
| `AI_PROVIDER` | No | gemini | Default ML provider: `gemini` or `ollama` |
| `GEMINI_API_KEY` | If gemini | — | Google Gemini API key |
| `OLLAMA_URL` | If ollama | http://localhost:11434 | Ollama server URL |
| `OLLAMA_MODEL` | No | mistral:7b | Ollama model name |
| `TELEGRAM_BOT_TOKEN` | No | — | Telegram bot token for Stage 2 |
| `WHISPER_API_KEY` | No | — | OpenAI Whisper API key (voice) |
| `ALEM_STT_API_KEY` | No | — | Alternative STT provider |
| `WHISPER_PROVIDER` | No | openai | STT provider: `openai` or `alem` |
| `INTERVIEW_TIMEOUT_MIN` | No | 30 | Interview timeout in minutes |
| `INTERVIEW_MIN_QUESTIONS` | No | 8 | Minimum interview questions |
| `INTERVIEW_MAX_QUESTIONS` | No | 15 | Maximum interview questions |
| `UPLOAD_DIR` | No | ./uploads | File upload directory |
| `SMTP_HOST` | No | — | SMTP server (enables email) |
| `SMTP_PORT` | No | 587 | SMTP port |
| `SMTP_USER` | No | — | SMTP username |
| `SMTP_PASS` | No | — | SMTP password |
| `SMTP_FROM` | No | noreply@invisionu.kz | Sender email address |

### Frontend (`frontend/.env.local`)

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `NEXT_PUBLIC_API_URL` | Yes | http://localhost:8080/api | Backend API URL (baked at build time) |

---

## Deployment (Railway)

The project is deployed at https://front-production-6189.up.railway.app/ using [Railway](https://railway.app/).

### Railway Setup

1. **PostgreSQL**: Add a PostgreSQL service. `DATABASE_URL` is auto-injected.
2. **Backend**: Deploy from `backend/` directory.
   - Set environment variables: `JWT_SECRET`, `GEMINI_API_KEY`, `ALLOW_ORIGINS` (= frontend URL).
   - Add a persistent volume for `/app/uploads`.
3. **Frontend**: Deploy from `frontend/` directory.
   - Set **Build Argument** (not env var): `NEXT_PUBLIC_API_URL=https://your-backend.up.railway.app/api`
   - `NEXT_PUBLIC_API_URL` is baked into the build — it must be a Build Argument in Railway.

> **Note**: Ollama does not work on Railway due to limited computational resources. It works on localhost only. Gemini is the recommended provider for deployment.

---

## Dependencies

### Backend (Go)

| Package | Purpose |
|---------|---------|
| `gin-gonic/gin` | HTTP web framework |
| `gin-contrib/cors` | CORS middleware |
| `jackc/pgx/v5` | PostgreSQL driver with connection pooling |
| `golang-jwt/jwt/v5` | JWT authentication |
| `go-telegram-bot-api/v5` | Telegram Bot API client |
| `go-playground/validator/v10` | Input validation |
| `joho/godotenv` | .env file loader |
| `golang.org/x/crypto` | bcrypt password hashing |

### Frontend (Node.js)

| Package | Purpose |
|---------|---------|
| `next` 14.2 | React framework (App Router) |
| `react` 18 | UI library |
| `typescript` 5 | Type safety |
| `tailwindcss` 3.4 | Utility-first CSS |
| `axios` | HTTP client |
| `recharts` | Chart components (dashboard) |
| `shadcn` + `lucide-react` | UI component library + icons |
| `sonner` | Toast notifications |
| `class-variance-authority` + `clsx` + `tailwind-merge` | CSS class utilities |

---

## Limitations & Notes

| Item | Details |
|------|---------|
| Ollama on deploy | Does not work on Railway/cloud due to GPU/compute requirements. Works on localhost only. |
| ML analysis consistency | Always gives +-3% error margins between runs on the same candidate. |
| Leadership = Technical | The scoring model places leadership potential at the same level as technical abilities (25% each). |
| Cost efficiency | Very cheap API usage — less than $1 for 400+ candidate analyses via Gemini. |
| Bias removal | Personal info (name, email, age, city) is never sent to the ML model. Evaluation is purely content-based. |
| Interview prerequisite | Stage 2 Telegram interview requires Stage 1 score >= 65 (overridable). |
| Consensus threshold | 4+ independent committee votes required to finalize a decision. |
| Voice messages | Require Whisper API key (OpenAI) or Alem STT API key for transcription. |

---

## Available Majors

| Tag | Full Name |
|-----|-----------|
| Engineering | Creative Engineering |
| Tech | Innovative IT Product Design and Development |
| Society | Sociology: Leadership and Innovation |
| Policy Reform | Public Policy and Development |
| Art + Media | Digital Media and Marketing |

---

## Team

Built for **Decentrathon 5.0** — Track: AI inDrive

- **Assylkhan** — Go backend, database architecture, Telegram bot
- **Zhanibek** — ML integration, Next.js frontend, UI/UX design
