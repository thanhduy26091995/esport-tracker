---
phase: planning
title: WC Betting Config Improvements — Planning
description: Task breakdown for configurable bet limits, handicap display, and consistent labels
---

# Project Planning & Task Breakdown

## Milestones

- [ ] **M1:** Backend — configurable min/max limits live
- [ ] **M2:** Frontend — bet forms respect dynamic limits; admin panel has config UI
- [ ] **M3:** Handicap display in pending/history tabs
- [ ] **M4:** Label consistency audit and fix

## Task Breakdown

### Phase 1: Backend — Configurable Bet Limits

- [ ] **1.1** — Add migration: `ALTER TABLE wc_config ADD COLUMN min_points INTEGER NOT NULL DEFAULT 1, ADD COLUMN max_points INTEGER NOT NULL DEFAULT 5`
  - File: new migration file in `backend/migrations/` (or wherever migrations live)
  - Effort: XS

- [ ] **1.2** — Update `WcConfig` Go model with `MinPoints int`, `MaxPoints int` fields
  - File: `backend/internal/model/wc_config.go`
  - Effort: XS

- [ ] **1.3** — Extend `GET /wc/config` response DTO to include `min_points`, `max_points`
  - File: `backend/internal/api/wc_handler.go` — `GetConfig` handler and response struct
  - Effort: XS

- [ ] **1.4** — Extend `PUT /admin/wc/config` to accept and persist `min_points`, `max_points`
  - File: `backend/internal/api/wc_handler.go` — `UpdateConfig` handler and request struct
  - Add service-layer validation: `min >= 1`, `max >= min`
  - Effort: S

- [ ] **1.5** — Move prediction points validation from Gin binding tag (`max=5`) to service layer
  - Files: `backend/internal/service/wc_service.go`, `backend/internal/service/wc_champion_service.go`
  - Replace hardcoded constant with `cfg.MaxPoints` / `cfg.MinPoints` fetched from config repo
  - Effort: S

### Phase 2: Frontend — Dynamic Limits

- [ ] **2.1** — Add `minPoints` / `maxPoints` to `wcConfigStore` Pinia state; populate on `fetchConfig()`
  - File: `frontend/src/stores/wcConfig.ts` (or equivalent)
  - Effort: XS

- [ ] **2.2** — Wire `:min` / `:max` on all prediction `<el-input-number>` inputs to store values
  - Files: `WcPredictionForm.vue` (3 inputs), champion prediction form
  - Effort: S

- [ ] **2.3** — Add "Giới hạn dự đoán" section to admin panel with min/max inputs and save button
  - File: `frontend/src/components/wc/WcAdminPanel.vue`
  - Effort: S

### Phase 3: Handicap Display in Predict Tabs

- [ ] **3.1** — Create `WcHandicapLine.vue` component
  - Props: `homeTeam: string`, `awayTeam: string`, `homeHandicap: number | null`, `awayHandicap: number | null`
  - Logic: determine which side gives handicap; render Vietnamese sentence
  - File: `frontend/src/components/wc/WcHandicapLine.vue`
  - Effort: S

- [ ] **3.2** — Integrate `WcHandicapLine` into `WcPredictionHistoryList.vue`
  - Show under each handicap-type prediction row, using match data available in parent/store
  - Effort: S

- [ ] **3.3** — Verify match data (home_handicap, away_handicap, team names) is accessible in the history list context
  - If not: ensure `wcMatchStore` or prediction response includes it
  - Effort: S (may require backend response augmentation)

### Phase 4: Label Consistency

- [ ] **4.1** — Create `frontend/src/utils/wcBetType.ts` with `WC_BET_TYPES as const`, `BET_TYPE_I18N_KEYS`, and `useWcBetTypeLabel()` composable
  - Pre-register `'corner'` type so "Kèo góc" only needs one i18n key to go live
  - Effort: XS

- [ ] **4.2** — Add canonical `betType*` i18n keys to `vi.json`; remove old `predictionType*` keys after replacing all usages
  - Keys: `betTypeHandicap`, `betTypeExactScore`, `betTypeOverUnder`, `betTypeCorner`
  - Effort: XS

- [ ] **4.3** — Audit and replace all template-context raw type strings in `.vue` files
  - Command: `grep -rn '"handicap"\|"exact_score"\|"over_under"' frontend/src --include="*.vue"`
  - Replace with `betTypeLabel(prediction.type)` using the new composable
  - Files: `WcPredictionHistoryList.vue`, `WcMatchPredictionList.vue`, `WcBetForm.vue`, possibly others
  - Verify with `npm run type-check` — exhaustiveness check on `BET_TYPE_I18N_KEYS` surfaces any missed type
  - Effort: M

## Dependencies

- Phase 2 depends on Phase 1 (backend must return limits before frontend can use them)
- Phase 3.3 may reveal a backend change needed before 3.2 can be completed
- Phase 4 is independent — can run in parallel with Phases 1–3

## Timeline & Estimates

| Phase | Effort | Notes |
|-------|--------|-------|
| Phase 1 (Backend limits) | ~2h | Migration + model + handler + service |
| Phase 2 (Frontend limits) | ~1.5h | Store + form bindings + admin UI |
| Phase 3 (Handicap display) | ~2h | New component + integration + data check |
| Phase 4 (Label consistency) | ~1.5h | Audit + replace across multiple files |
| **Total** | **~7h** | Can parallelize Phase 4 with Phases 1–3 |

## Risks & Mitigation

| Risk | Mitigation |
|------|-----------|
| Match handicap data not in prediction list response | Augment backend response or load from wcMatchStore |
| Gin binding `max=5` tag on champion handler missed | Search all WC handler files for `max=5` before closing task |
| Frontend config fetch timing — form mounts before config loaded | Initialize store with defaults (1, 5); update reactively |
| Fractional handicap values (0.5) display edge cases | Round-trip test with known fixtures |

## Resources Needed

- Access to DB for migration
- Existing admin panel (WcAdminPanel.vue) for UI integration
- i18n keys already in `vi.json` — no translator needed
