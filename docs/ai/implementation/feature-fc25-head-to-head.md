---
phase: implementation
title: Implementation Guide
description: Technical implementation notes, patterns, and code guidelines
---

# Implementation Guide — FC25 Head-to-Head Popup

> To be filled in during `/execute-plan`. Seeded with the intended approach below.

## Development Setup
**How do we get started?**

- Backend: `cd backend && go run cmd/server/main.go`; tests `go test ./...`.
- Frontend: `cd frontend && npm run dev`; types `npm run type-check`.
- No new env vars, no migration (read-only feature over existing tables).

## Code Structure
**How is the code organized?**

Backend (touch points):
- `backend/internal/repository/match_repository.go` — add `GetHeadToHeadMatches`.
- `backend/internal/services/user_service.go` — add `GetHeadToHead` + DTOs + `ErrSamePlayer`.
- `backend/internal/api/user_handler.go` — add `GetHeadToHead` handler.
- `backend/internal/api/router.go` — register route in `/users` group.

Frontend (touch points):
- `frontend/src/services/userService.ts` — add `getHeadToHead`.
- `frontend/src/types/*` — add `HeadToHeadResponse` type (snake_case).
- `frontend/src/components/user/HeadToHeadModal.vue` — new modal.
- `frontend/src/views/DashboardView.vue` — add the "So tài" card.
- `frontend/src/locales/vi.*`, `en.*` — `dashboard.headToHead.*` keys.

## Implementation Notes
**Key technical details to remember:**

### Core Features
- **Opponents-only aggregation**: enforced by the join condition `mp2.team_number <> mp1.team_number`. Never post-filter in Go.
- **Win orientation**: `player1_won = (p1_team == winner_team)`; `player2_wins = total - player1_wins` (no draws).
- **Form**: map ordered-by-date-DESC rows to `"W"/"L"` from P1 POV; take first 10 (most-recent first).
- **Streak**: walk `form` from index 0; count the leading run of identical results; `PlayerID` = P1 if that run is "W", else P2; `nil` when total is 0.
- **Recent matches**: first 10 rows, already date-DESC.

### Patterns & Best Practices
- Repository → Service → Handler; handlers never touch the repo.
- Cache-aside via `cache.GetOrFetch[*HeadToHeadResponse]`, key `esport:h2h:v{version}:{p1}:{p2}` where `version, _ := s.cache.GetInt("esport:matches:version")`.
- Sentinel error `ErrSamePlayer` at service layer; handler maps to 400.
- Frontend: `<script setup lang="ts">`, Element Plus `el-dialog`, Tailwind for layout, all strings via `t('dashboard.headToHead.*')`.
- Chart: register chart.js modules once in the component (`ChartJS.register(ArcElement, Tooltip, Legend)`), follow `components/wc/analytics/*`.

## Integration Points
**How do pieces connect?**

- Modal calls `userService.getHeadToHead(p1, p2)` → `GET /api/v1/users/head-to-head`.
- Dashboard sources the selectable players from the existing `userStore` (active users) — no new fetch.

## Error Handling
**How do we handle failures?**

- Backend: `player1 == player2` → `ErrSamePlayer` → 400; missing/inactive player → not-found → 404; bad UUID → 400.
- No-history is **not** an error → 200 with zeroed response + empty arrays.
- Frontend: show loading spinner while fetching, an error alert on non-2xx, and an empty-state block when `total_matches === 0`.

## Performance Considerations
**How do we keep it fast?**

- Single indexed join; result set small. Cache-aside + singleflight.
- Lists capped at 10; no pagination.
- Cache auto-invalidates through the shared `esport:matches:version` counter on any match write.

## Security Notes
**What security measures are in place?**

- Read-only; no mutation; no new auth surface (mirrors existing public `/users` reads).
- Validate both UUID params; reject same-player.
- No secrets, no PII beyond existing public player fields.
</content>
