# CLAUDE.md — Esport Score Tracker

## Quick Context

FC25 match tracker + World Cup 2026 prediction/betting game for a friend group.  
Backend: Go/Gin/GORM/PostgreSQL. Frontend: Vue 3/TypeScript/Pinia.

## Knowledge Docs

Read the relevant file before making changes in that domain:

### Project-Wide
| Topic | File |
|-------|------|
| Project structure, tech stack, naming conventions | [`docs/ai/knowledge/project-overview.md`](docs/ai/knowledge/project-overview.md) |
| Go backend patterns (DI, models, repository, errors) | [`docs/ai/knowledge/backend-patterns.md`](docs/ai/knowledge/backend-patterns.md) |
| Vue frontend patterns (Pinia, services, route meta, GSI) | [`docs/ai/knowledge/frontend-patterns.md`](docs/ai/knowledge/frontend-patterns.md) |
| i18n / multi-language (vue-i18n, locale files, conventions) | [`docs/ai/knowledge/frontend-i18n.md`](docs/ai/knowledge/frontend-i18n.md) |
| **Full database schema** (all tables, columns, FK relationships, enums, seeds) | [`docs/ai/knowledge/database-schema.md`](docs/ai/knowledge/database-schema.md) |

### Core Esport System
| Topic | File |
|-------|------|
| Match types (1v1/2v2/1v2), debt settlement, fund, score bonuses, tiers, personalization | [`docs/ai/knowledge/core-esport-system.md`](docs/ai/knowledge/core-esport-system.md) |
| Tournament creation, 2v2 scheduler, round-robin | [`docs/ai/knowledge/tournament-system.md`](docs/ai/knowledge/tournament-system.md) |

### World Cup 2026 (WC) System
| Topic | File |
|-------|------|
| WC base system: tables, feature flag, wallet, match sync | [`docs/ai/knowledge/wc-core-system.md`](docs/ai/knowledge/wc-core-system.md) |
| WC auth: Google OAuth, JWT, middleware, route guards | [`docs/ai/knowledge/wc-auth-system.md`](docs/ai/knowledge/wc-auth-system.md) |
| WC betting: handicap, exact score, O/U, payout, settlement, house P&L | [`docs/ai/knowledge/wc-betting-system.md`](docs/ai/knowledge/wc-betting-system.md) |
| WC champion prediction | [`docs/ai/knowledge/wc-champion-prediction.md`](docs/ai/knowledge/wc-champion-prediction.md) |

## Feature Design Docs

Each feature has full design + requirements + planning + implementation docs in `docs/ai/`:

| Feature | Design Doc |
|---------|-----------|
| WC Google OAuth login | `docs/ai/design/feature-wc-google-oauth-login.md` |
| WC champion prediction | `docs/ai/design/feature-wc-champion-prediction.md` |
| WC betting refinements (payout type, VND display) | `docs/ai/design/feature-betting-refinements.md` |
| WC house P&L dashboard | `docs/ai/design/feature-house-pnl-dashboard.md` |
| WC StatsAPI odds import + Poisson | `docs/ai/design/feature-statsapi-odds-import.md` |
| WC UX enhancements (sidebar, CTA, filter) | `docs/ai/design/feature-world-cup-ux-enhancements.md` |
| WC2026 upcoming matches widget | `docs/ai/design/feature-wc2026-upcoming-matches-dashboard-widget.md` |
| WC refactored (player filter, timezone, admin page) | `docs/ai/design/feature-refactored-wc2026.md` |
| WC base system | `docs/ai/design/feature-world-cup-2026.md` |
| Dynamic 2v2 scheduler | `docs/ai/design/feature-dynamic-2v2-scheduler.md` |
| Random tournament | `docs/ai/design/feature-random-tournament.md` |
| 1v2 match type | `docs/ai/design/feature-1v2-match-type.md` |
| External bet bonus | `docs/ai/design/feature-external-bet-bonus.md` |
| Win rate & tier evaluation | `docs/ai/design/feature-win-rate-tier-evaluation.md` |
| Dashboard player sort strategy | `docs/ai/design/feature-dashboard-player-sort-strategy.md` |
| Player personalization (avatar, club, theme) | `docs/ai/design/feature-player-personalization.md` |
| Inline player creation | `docs/ai/design/feature-inline-player-creation.md` |
| Multi-language support | `docs/ai/design/feature-multi-language-support.md` |
| VI localization hardcode cleanup | `docs/ai/design/feature-vi-localization-hardcode-cleanup.md` |
| Esport score tracker (base system) | `docs/ai/design/feature-esport-score-tracker.md` |
| Frontend integration | `docs/ai/design/feature-frontend-integration.md` |
| Tournament round-robin + top 4 knockout format | `docs/ai/design/feature-tournament-round-robin-knockout.md` |
| WC settle preview popup (Tính kết quả / Tính điểm toàn bộ / Tính lại toàn bộ) | `docs/ai/design/feature-wc-settle-preview.md` |
| WC admin block/unblock user | `docs/ai/design/feature-wc-user-block.md` |
| Bug fix batch 21-Jun-2026 (8 bugs: redirect, scroll, collapse, P&L, responsive, multi-pick champion, smart cron, settlement name) | `docs/ai/design/feature-fix-bug-21-june-2026.md` |
| WC group standings table (W/D/L, GF/GA/GD, Points, Form) on /world-cup schedule page | `docs/ai/design/feature-wc-group-standings.md` |
| WC standalone site (soc.sitenow.cloud) — build-time VITE_SITE=soc flag, WC-only routes + nav | `docs/ai/design/feature-wc-soc-site.md` |

## Dev Commands

```bash
# Backend
cd backend && go run cmd/server/main.go
cd backend && go test ./...

# Frontend
cd frontend && npm run dev
cd frontend && npm run type-check
```

## Hard Rules

- Use `shopspring/decimal` for all money/point arithmetic — never `float64`
- Never call repositories from handlers — always go through the service layer
- All WC route wiring lives in `backend/internal/api/router.go`
- All frontend route guards live in `frontend/src/router/index.ts` `beforeEach`
- WC auth state must only be modified through `wcAuthStore` actions
- WC and core esport systems never write to each other's tables
- All new UI strings must go through `vue-i18n` — no hardcoded text in components
