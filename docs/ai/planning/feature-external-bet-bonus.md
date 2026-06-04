---
phase: planning
title: "External Bet Bonus Points — Planning"
description: Task breakdown and implementation order for the external bet bonus feature
---

# Project Planning & Task Breakdown

## Milestones

- [ ] Milestone 1: DB migration and backend service complete
- [ ] Milestone 2: API endpoints wired and tested
- [ ] Milestone 3: Frontend dialog and Fund page integration

## Task Breakdown

### Phase 1: Backend — Data Layer

- [ ] **1.1** Create `internal/model/score_bonus.go` with `ScoreBonus` struct and `TableName()`
- [ ] **1.2** Create `internal/repository/score_bonus_repository.go` with `Create`, `GetAll`, `GetByID`, `Delete`
- [ ] **1.3** Add DB migration: `score_bonuses` table (UUID PK, user_id FK, points, fund_amount, description, recorded_by, bonus_date, created_at)
- [ ] **1.4** Register `ScoreBonus` in GORM AutoMigrate (or apply raw SQL migration)

### Phase 2: Backend — Service & Handler

- [ ] **2.1** Create `internal/service/score_bonus_service.go`
  - `CreateBonus`: validate → begin tx → insert bonus → update user score → commit → tier recalc
  - `DeleteBonus`: get bonus → begin tx → revert score → delete bonus → commit → tier recalc
  - `GetAll`: pagination
  - `GetByUserID`: for player history endpoint
- [ ] **2.2** Create `internal/api/score_bonus_handler.go` (Create, Delete only — no GET)
- [ ] **2.3** Update `internal/api/match_handler.go`:
  - `GetAll`: inject `ScoreBonusService`, fetch bonuses, merge + sort by date, return unified feed with `"type"` field
  - `GetRecent`: same merge approach (bonuses appear in Dashboard recent feed)
  - `GetStats`: no change (counts `matches` table only)
- [ ] **2.4** Register routes in `router.go`:
  ```
  POST   /api/v1/score-bonuses
  DELETE /api/v1/score-bonuses/:id
  ```
  (GET /api/v1/matches already exists — no new route for reading)
- [ ] **2.5** Wire `ScoreBonusService` into both `bonusHandler` and `matchHandler` in router setup

### Phase 3: Frontend

- [ ] **3.1** Create `frontend/src/types/scoreBonus.ts` (`ScoreBonus`, `CreateScoreBonusRequest`)
- [ ] **3.2** Create `frontend/src/services/scoreBonusService.ts` (API calls)
- [ ] **3.3** Create `ScoreBonusForm.vue` dialog component
  - Fields: Player (searchable dropdown), Points (number), Description (text), Date (optional)
  - Validation: player required, points > 0
- [ ] **3.4** Add "Cộng điểm cá cược" button (Fund page or Dashboard) that opens the dialog
- [ ] **3.5** On submit: call service → refresh leaderboard/dashboard
- [ ] **3.6b** Add `MatchFeedItem` discriminated union type to `frontend/src/types/match.ts`; update `matchStore` from `Match[]` → `MatchFeedItem[]`
- [ ] **3.6c** `MatchList` type filter: add `[Cược]` option; filter logic hides bonus when type filter active, hides matches when `[Cược]` selected
- [ ] **3.6d** `MatchList` bonus card rendering: `[Cược]` tag, single-row player+pts+description layout, delete → `DELETE /api/v1/score-bonuses/:id`
- [ ] **3.6e** `matchStore`: no change to fetch — `/matches` returns unified feed; refresh after bonus create/delete
- [ ] **3.6** i18n keys for all new UI strings in `vi.json` and `en.json`

## Dependencies

- Phase 2 depends on Phase 1 (model + repo must exist before service)
- Phase 3.2+ depends on Phase 2 (API must exist before service calls)
- Phase 3.3–3.5 can be built against a mock/stub if backend not yet deployed

## Timeline & Estimates

| Task | Estimate |
|---|---|
| 1.1–1.4 Model + migration | 30 min |
| 2.1 Service logic | 45 min |
| 2.2–2.4 Handler + routes | 20 min |
| 3.1–3.2 Types + service | 15 min |
| 3.3–3.5 Dialog + integration | 45 min |
| 3.6 i18n | 15 min |
| **Total** | **~2.5 hours** |

## Risks & Mitigation

| Risk | Mitigation |
|---|---|
| Duplicate bonus recording | Idempotency not required — description field provides human audit trail |
| Player history page doesn't exist yet | May need to create the page or add bonus section to existing user view |

## Resources Needed

- Backend: Go, GORM, PostgreSQL migration
- Frontend: Vue 3, TypeScript, Element Plus
