# FC25 Esport Score Tracker

FC25 match tracker with automatic debt settlement and a World Cup 2026 prediction/betting game for a friend group.

## Quick Start

### Prerequisites

- Go 1.21+
- Node.js 18+
- PostgreSQL 14+

### 1. Backend

```bash
cd backend
cp .env.example .env   # fill in values
go run cmd/server/main.go
```

Runs on `http://localhost:8080`

### 2. Frontend

```bash
cd frontend
npm install
npm run dev
```

Runs on `http://localhost:5173`

---

## Environment Variables

### Backend (`backend/.env`)

```env
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=esport_tracker
DB_SSLMODE=disable

# Server
PORT=8080
CORS_ORIGINS=http://localhost:5173

# WC2026 auth
JWT_SECRET=your-secret-here
GOOGLE_CLIENT_ID=your-google-client-id

# WC2026 match sync (football-data.org)
FOOTBALL_DATA_API_KEY=your-key-here
WC_SYNC_INTERVAL_MINUTES=30   # idle interval; auto-switches to 5 min when matches are live

# WC2026 odds (optional)
ODDSAPI_KEY=your-key-here

# Seed first WC admin on startup (optional)
WC_ADMIN_NAME=admin
WC_ADMIN_PASSWORD=secret
```

### Frontend (`frontend/.env`)

```env
VITE_API_BASE_URL=http://localhost:8080/api/v1
VITE_GOOGLE_CLIENT_ID=your-google-client-id
```

---

## Project Structure

```
esport-tracker/
├── backend/
│   ├── cmd/server/          # Entry point
│   └── internal/
│       ├── api/             # HTTP handlers & router
│       ├── cron/            # Background jobs (WC match sync)
│       ├── database/        # DB connection & migrations
│       ├── middleware/       # JWT, admin, feature-flag guards
│       ├── model/           # GORM models
│       ├── repository/      # Data access layer
│       └── service/         # Business logic
│
├── frontend/
│   └── src/
│       ├── components/      # Reusable Vue components
│       ├── composables/     # Shared composition functions
│       ├── router/          # Vue Router + route guards
│       ├── services/        # Axios API clients
│       ├── stores/          # Pinia stores
│       ├── types/           # TypeScript types
│       └── views/           # Page-level components
│
└── docs/ai/                 # Feature documentation (requirements, design, planning)
```

---

## Features

### Core Esport System

- **Players** — CRUD, avatar upload, FC club badge, tier (Noob / Normal / Pro)
- **Matches** — 1v1, 2v2, 1v2 (one vs two); automatic score & point updates
- **Score bonuses** — extra points attached to a match
- **Debt settlement** — auto-triggers when a player's score hits the threshold; splits payment between fund and recent opponents
- **Fund** — deposit / withdraw / balance tracking
- **Tournaments** — round-robin group stage + top-4 knockout bracket; random team assignment scheduler

### World Cup 2026 System

- **Match schedule** — synced from football-data.org on an adaptive cron (5 min when live, configurable otherwise)
- **Auth** — password login + Google OAuth; JWT session; admin vs player roles
- **Predictions** — handicap, exact score, over/under bet types; locking at kickoff
- **Champion prediction** — multi-pick (one pick per team); window open/close; settlement pays winners and deducts losers
- **Leaderboard** — ranked by wallet balance
- **House P&L** — admin dashboard showing stake vs payout across all settled bets
- **Settlements** — VND cash settlement based on final wallet balances
- **Admin panel** — manage matches, odds, users (block/unblock), wallet top-up, finalize/re-finalize results

---

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend language | Go 1.21 |
| HTTP framework | Gin |
| ORM | GORM |
| Database | PostgreSQL 14 |
| Money arithmetic | shopspring/decimal |
| Frontend framework | Vue 3 + TypeScript |
| Build tool | Vite |
| UI library | Element Plus |
| State management | Pinia |
| HTTP client | Axios |
| Auth | JWT + Google OAuth |

---

## API Overview

### Core (`/api/v1`)

```
GET|POST        /users
GET             /users/leaderboard
GET|PUT|DELETE  /users/:id
PUT             /users/:id/avatar
GET             /matches
POST            /matches
DELETE          /matches/:id
GET|PUT         /config
GET|POST        /fund/...
GET|POST        /settlements/...
GET|POST|DELETE /tournaments/...
```

### WC2026 (`/api/v1/wc`)

```
# Public
GET  /config
GET  /matches
GET  /matches/:id
GET  /leaderboard
GET  /champion/config
GET  /champion/teams
GET  /champion/predictions

# Auth (JWT required)
POST /auth/login
POST /auth/google
GET|PUT /wallet
GET|POST|DELETE|PUT /predictions
GET|POST|PUT|DELETE /bets
GET|POST|DELETE /champion/predict

# Admin
POST /admin/sync
PUT  /admin/matches/:id
POST /admin/matches/:id/finalize
POST /admin/matches/:id/settle
GET  /admin/house-pnl
GET|POST /admin/settlements
GET|PUT  /admin/users
POST /admin/champion/settle
...
```

---

## Development

```bash
# Backend
cd backend && go run cmd/server/main.go
cd backend && go test ./...

# Frontend
cd frontend && npm run dev
cd frontend && npm run build
```
