# Project Overview

## What This Is

FC25 Esport Score Tracker — a private app for managing FC25 matches among a friend group. The app has two distinct feature domains:

1. **Core esport tracker** — match recording, scoring, debt settlement, fund management
2. **World Cup 2026 (WC2026)** — prediction game, betting with virtual wallets, Google OAuth login

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go 1.21+, Gin (HTTP), GORM (ORM), PostgreSQL 14 |
| Frontend | Vue 3, TypeScript, Vite, Pinia, Vue Router, Element Plus, Tailwind CSS |
| Auth | JWT (7-day), Google Identity Services (GSI) for WC feature |
| Decimal math | `shopspring/decimal` — always use for money/point calculations |
| DB hosting | PostgreSQL local / Supabase (prod) |

## Directory Structure

```
esport/
├── backend/
│   ├── cmd/server/          # main.go entry point
│   └── internal/
│       ├── api/             # Gin handlers + router.go (all route wiring)
│       ├── service/         # Business logic
│       ├── repository/      # Data access (GORM)
│       ├── model/           # DB models (+ TableName() overrides)
│       ├── database/        # Connect(), AutoMigrate, seed
│       └── middleware/      # JWT, admin gate, Google-link gate, feature flag
├── frontend/
│   └── src/
│       ├── views/           # Page components (WcXxxView.vue prefix = WC feature)
│       ├── components/wc/   # WC-scoped reusable components
│       ├── stores/          # Pinia stores (wcAuthStore.ts for WC auth)
│       ├── services/        # Thin axios wrappers (wcAuthService.ts, wcProfileService.ts)
│       ├── types/wc.ts      # All WC TypeScript types
│       └── router/index.ts  # Route guards and meta-driven access control
└── docs/ai/
    ├── requirements/        # Feature requirement docs
    ├── design/              # Architecture/API design docs
    ├── planning/            # Task breakdown docs
    ├── implementation/      # Implementation notes
    ├── testing/             # Test plans
    └── knowledge/           # ← this folder: persistent AI context
```

## Naming Conventions

- **Backend** — WC-domain files are prefixed `wc_`: `wc_user.go`, `wc_auth_handler.go`, `wc_repository.go`
- **Frontend** — WC views: `WcXxxView.vue`; WC components: `components/wc/WcXxx.vue`; WC services: `wcXxxService.ts`
- **DB tables** — WC tables prefixed `wc_`: `wc_users`, `wc_matches`, `wc_bets`, `wc_predictions`, `wc_wallets`

## Feature Flag

The WC2026 feature is controlled by `wc_config.is_enabled` (single DB row, id=1).  
When disabled, all non-public WC routes return 404 via `WcFeatureMiddleware`.  
Frontend route guard also checks `/api/v1/wc/config` and redirects to schedule if feature is off.

## Key Business Domains

See the feature docs in `docs/ai/` for full details on each:

- **Debt settlement** — score threshold triggers debt-to-VND conversion, 50/50 split to fund and creditors
- **WC predictions** — handicap / exact score / over-under bets with wallet-based payouts
- **WC champion prediction** — single pick of tournament winner, settled at end
- **Google OAuth gate** — all non-admin WC players must link a Google account to play
