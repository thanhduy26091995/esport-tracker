---
phase: testing
title: Refactored WC2026 — Testing Strategy
description: Test scope and validation criteria for player filter, VN timezone display, live glow, and admin route.
---

# Testing Strategy

## Scope

- **Feature 1 (player filter)**: Integration test on the repository method + handler-level test.
- **Feature 2 (timezone + live glow)**: Manual visual check (no unit tests for CSS/locale formatting).
- **Feature 3 (admin route)**: Router guard logic + manual navigation test.

---

## Test Files

| File | Layer | Coverage Target |
|------|-------|----------------|
| `backend/internal/service/wc_integration_test.go` (existing) | Integration | Reference only — no changes |
| `backend/internal/repository/match_repository_test.go` (new or existing) | Repository | `GetAllFiltered` with and without `playerID` |
| `backend/internal/api/match_handler_test.go` (existing/extend) | Handler | `player_id` query param parsing |

---

## Unit / Integration Tests

### Feature 1 — GetAllFiltered

| Case | Input | Expected |
|------|-------|----------|
| No player filter | `playerID = nil` | Returns all matches (same as `GetAll`) |
| Valid player filter | `playerID = <uuid of player who played 3 matches>` | Returns only those 3 matches |
| Player with no matches | `playerID = <uuid of player with 0 matches>` | Returns empty slice, no error |
| Invalid UUID in handler | `?player_id=not-a-uuid` | HTTP 400, `{ "error": "invalid player_id" }` |
| Valid UUID not in DB | `?player_id=<valid uuid, no matches>` | HTTP 200, empty array |

### Feature 3 — Router guard

| Case | Token state | Expected redirect |
|------|-------------|------------------|
| No WC token | `localStorage` empty | `→ wc-login` (via `requiresWcAuth`) |
| WC token, `is_admin: false` | Valid JWT, non-admin | `→ wc-schedule` |
| WC token, `is_admin: true` | Valid JWT, admin | Page loads |
| Malformed JWT | Corrupt base64 | `→ wc-schedule` (catch block) |

---

## Manual Validation Checklist

### Feature 1 — Player filter UI
- [ ] Open `/matches`; player dropdown visible above list.
- [ ] Select a player → list updates to show only that player's matches; all cards include the selected player in their participants list.
- [ ] Select "All players" → full list returns.
- [ ] Select a player with no matches → list shows empty state (not error).

### Feature 2 — Timezone
- [ ] Change OS/browser timezone to UTC or UTC+5. Open `/world-cup`. Verify kickoff time shown matches UTC+7 (e.g., a 13:00 UTC match shows 20:00 on the card).
- [ ] Live match card has a glowing green border (animate between two glow intensities).
- [ ] Scheduled and completed cards have no glow; the static status badge still appears.

### Feature 3 — Admin route
- [ ] Logged-out user navigates to `/world-cup/admin` → redirected to `/world-cup/login`.
- [ ] Non-admin WC user navigates to `/world-cup/admin` → redirected to `/world-cup`.
- [ ] WC admin navigates to `/world-cup/admin` → `WcAdminPanel` renders with all sections (feature toggle, sync, match management, settlement, top-up).
- [ ] Open `/world-cup/predict` → no "Admin" tab visible (for any user role).

---

## Test Data & Environments

- Use a local Postgres DB seeded with at least 2 players and 5 matches (3 for player A, 2 for player B).
- WC feature must be enabled in `wc_config` for the admin route guard test.
- Two WC user accounts needed: one with `is_admin = true`, one with `is_admin = false`.

---

## Execution

```bash
# Backend tests
cd backend && go test ./internal/repository/... ./internal/api/...

# Frontend — no automated tests; use manual checklist above
cd frontend && npm run dev
```

---

## Coverage & Quality Gates

- `GetAllFiltered` must pass all 5 table-driven cases.
- Handler must return 400 for invalid UUID and 200 for valid.
- All 4 router guard cases verified manually.
- No regression: existing `/matches` page (no filter) still works.
- No regression: existing WC predict/betting flow unaffected after admin tab removal.

---

## Risks & Gaps

- No automated E2E tests for the Vue router guard — manual check is the gate.
- CSS `animation` for glow cannot be unit-tested; visual QA required.
- If the JWT library changes the payload structure, the `atob` parse in the guard may break — low risk for this codebase.
