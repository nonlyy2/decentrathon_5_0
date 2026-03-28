# inVision U — Admissions AI

AI-powered candidate screening platform for inVision U (inDrive's 100% scholarship university in Kazakhstan), built for **Decentrathon 5.0**.

## Overview

inVision U receives hundreds of scholarship applications. This platform automates the initial screening using Google Gemini 2.0 Flash — scoring essays across 5 dimensions and surfacing AI-generated content risks — so the admissions committee can focus on final decisions rather than manual review.

### Key Features

- **Applicant portal** — public form for candidates to submit essays and background info
- **AI screening** — Gemini 2.0 Flash scores each application across 5 criteria
- **Human-in-the-loop** — AI recommends, committee decides (shortlist / waitlist / reject)
- **Batch analysis** — analyze all pending candidates in one click
- **AI-writing detection** — flags low/medium/high risk of AI-generated essays
- **Dashboard** — real-time stats, score distribution, category breakdown
- **Decision audit trail** — every committee decision is logged with optional notes

### Scoring Model

| Dimension | Weight | Description |
|-----------|--------|-------------|
| Leadership | 25% | Evidence of initiative and impact |
| Motivation | 25% | Clarity of purpose and alignment with inVision U mission |
| Growth | 20% | Learning mindset and resilience |
| Vision | 15% | Concrete goals and societal impact |
| Communication | 15% | Clarity, depth, and authenticity of writing |

**Categories:**
- 80–100 → Strong Recommend
- 65–79 → Recommend
- 50–64 → Borderline
- 0–49 → Not Recommended

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go 1.22 + Gin |
| Database | PostgreSQL 16 |
| AI | Google Gemini 2.0 Flash (REST API) |
| Frontend | Next.js 14 (App Router) + Tailwind CSS v3 |
| UI Components | shadcn/ui v4 |
| Charts | Recharts |
| Auth | JWT (72h expiry) + bcrypt |

## Project Structure

```
decentrathon/
├── backend/
│   ├── cmd/server/main.go          # Entry point
│   ├── internal/
│   │   ├── config/                 # Env config
│   │   ├── database/               # DB pool + migrations
│   │   ├── gemini/                 # Gemini client, prompt, parser
│   │   ├── handlers/               # HTTP handlers
│   │   ├── middleware/             # Auth + CORS
│   │   ├── models/                 # Data models
│   │   └── seed/                   # Admin + demo candidates
│   ├── .env                        # Local environment vars
│   ├── Dockerfile                  # Production build
│   └── go.mod
├── frontend/
│   ├── app/
│   │   ├── (dashboard)/            # Protected admin pages
│   │   │   ├── dashboard/          # Stats overview
│   │   │   ├── candidates/         # Candidate list + detail
│   │   │   └── analytics/          # Analytics view
│   │   ├── apply/                  # Public application form
│   │   └── page.tsx                # Login page
│   ├── components/                 # Shared UI components
│   ├── lib/                        # API client, auth, hooks, types
│   └── .env.local                  # Frontend env vars
└── tasks/                          # Project management task files
```

## Quick Start

### Prerequisites

- Go 1.22+
- Node.js 20+
- PostgreSQL 16 (running locally)
- Google Gemini API key (free at [aistudio.google.com](https://aistudio.google.com))

### 1. Database Setup

```bash
createdb invisionu
```

### 2. Backend

```bash
cd backend

# Copy and configure environment
cp .env.example .env
# Edit .env: set DATABASE_URL and GEMINI_API_KEY

# Install dependencies
go mod download

# Run server (auto-runs migrations + seeds admin user)
go run ./cmd/server
```

Backend starts at `http://localhost:8080`.

Default admin credentials: `admin@invisionu.kz` / `admin123`

### 3. Frontend

```bash
cd frontend

# Install dependencies
npm install

# Copy and configure environment
cp .env.local.example .env.local
# Edit .env.local if backend is not on :8080

# Start dev server
npm run dev
```

Frontend starts at `http://localhost:3000`.

### Environment Variables

**Backend (`backend/.env`)**

```env
DATABASE_URL=postgres://YOUR_USER@localhost:5432/invisionu?sslmode=disable
PORT=8080
JWT_SECRET=change-me-in-production
GEMINI_API_KEY=your-gemini-api-key
ALLOW_ORIGINS=http://localhost:3000
```

**Frontend (`frontend/.env.local`)**

```env
NEXT_PUBLIC_API_URL=http://localhost:8080/api
```

## API Endpoints

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/api/apply` | — | Submit application |
| POST | `/api/auth/login` | — | Committee login |
| GET | `/api/candidates` | JWT | List with filters |
| GET | `/api/candidates/:id` | JWT | Detail + analysis |
| POST | `/api/candidates/:id/analyze` | JWT | Trigger AI analysis |
| POST | `/api/candidates/:id/decision` | JWT | Record decision |
| POST | `/api/analyze-all` | JWT | Batch analyze pending |
| GET | `/api/analyze-all/status` | JWT | Batch progress |
| GET | `/api/stats` | JWT | Dashboard statistics |

## Rate Limiting

The Gemini integration is rate-limited to 15 RPM (1 request per 4 seconds) to stay within the free tier. Batch analysis queues candidates sequentially with exponential backoff on errors.

## Deployment

The backend includes a multi-stage `Dockerfile` and `railway.json` for one-click deployment to [Railway](https://railway.app).

```bash
cd backend
railway up
```

## Team

Built for Decentrathon 5.0 — Stage #1 deadline: March 29, 2025

- **Assylkhan** — Go backend, AI integration, database
- **Zhanibek** — React frontend, UI/UX
