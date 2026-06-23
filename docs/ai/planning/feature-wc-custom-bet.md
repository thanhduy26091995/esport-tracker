---
phase: planning
title: WC Custom Bet — Planning
description: Task breakdown for custom proposition bets (kèo phụ)
---

# Project Planning & Task Breakdown

## Milestones

- [x] **M1:** Backend — CRUD + settlement API live
- [x] **M2:** Frontend admin — create, close, settle, void custom bets
- [x] **M3:** Frontend player — view and place bets on custom bets
- [ ] **M4:** Bet history integration

## Task Breakdown

### Phase 1: Backend

- [x] **1.1** — Create Go models in `backend/internal/model/wc_custom_bet.go`
  - `WcCustomBet`, `WcCustomBetOption`, `WcCustomBetEntry`, `WcCustomBetWithOptions`
  - Effort: XS

- [x] **1.2** — Register models in `database.go` AutoMigrate list
  - Effort: XS

- [x] **1.3** — Create `backend/internal/repository/wc_custom_bet_repository.go`
  - `Create(bet, options)` — creates bet + options in one transaction
  - `GetByID(id)` — returns bet + options + entry counts
  - `ListForMatch(matchID)` — all custom bets for a match
  - `ListForMatchWithMyEntry(matchID, userID)` — player view
  - `PlaceEntry(entry)` — inserts entry (debit handled by service)
  - `GetEntry(entryID)` — for cancel
  - `DeleteEntry(entryID)` — for cancel
  - `GetEntriesForBet(betID)` — for settlement
  - `UpdateBet(bet)` — update title/status/settled fields
  - `UpdateOptionWinner(optionID, isWinner)` — mark winning option
  - `UpdateEntryResult(entryID, status, payout)` — after settlement
  - Effort: M

- [x] **1.4** — Create `backend/internal/service/wc_custom_bet_service.go`
  - `CreateCustomBet(matchID, adminID, title, options)` — validates, creates
  - `UpdateCustomBet(betID, req)` — update title/odds/status; guard: cannot close if already settled/void
  - `PlaceEntry(betID, userID, optionID, stake)` — validates limits, deducts wallet, creates entry
  - `CancelEntry(entryID, userID)` — validates ownership + pending status, refunds wallet
  - `Settle(betID, winningOptionID, adminID)` — DB transaction: mark winner, calc payouts, credit winners, write wallet logs
  - `VoidBet(betID, adminID)` — DB transaction: refund all pending entries, set status=void
  - Effort: L

- [x] **1.5** — Create `backend/internal/api/wc_custom_bet_handler.go`
  - Admin: `CreateCustomBet`, `UpdateCustomBet`, `SettleCustomBet`, `VoidCustomBet`, `AdminListCustomBets`
  - Player: `ListCustomBets`, `PlaceEntry`, `CancelEntry`
  - Effort: M

- [x] **1.6** — Wire routes in `backend/internal/api/router.go`
  - Admin group: `POST /admin/matches/:id/custom-bets`, `PUT /admin/custom-bets/:id`, `POST /admin/custom-bets/:id/settle`, `PUT /admin/custom-bets/:id/void`
  - Player group: `GET /wc/matches/:id/custom-bets`, `POST /wc/custom-bets/:id/entry`, `DELETE /wc/custom-bet-entries/:id`
  - Effort: XS

### Phase 2: Frontend — Admin

- [x] **2.1** — Add `WcCustomBet`, `WcCustomBetOption`, `WcCustomBetEntry` TypeScript types to `frontend/src/types/wc.ts`
  - Effort: XS

- [x] **2.2** — Add custom bet API calls to `frontend/src/services/wcService.ts`
  - `createCustomBet`, `updateCustomBet`, `settleCustomBet`, `voidCustomBet`, `adminListCustomBets`
  - Effort: XS

- [x] **2.3** — Create `WcAdminCustomBetPanel.vue` component (used inside WcAdminPanel per match)
  - Create form: title input + dynamic options list (label + odds + add/remove row)
  - List existing bets for the match: status badge, settle/void buttons
  - Settle dialog: select winning option → confirm
  - Effort: L

- [x] **2.4** — Integrate `WcAdminCustomBetPanel` into `WcAdminPanel.vue` per-match section
  - Add "Kèo phụ" toggle/button in the match management row
  - Effort: S

### Phase 3: Frontend — Player

- [x] **3.1** — Add player-facing API calls to `wcService.ts`
  - `listCustomBets`, `placeCustomBetEntry`, `cancelCustomBetEntry`
  - Effort: XS

- [x] **3.2** — Create `WcCustomBetCard.vue` component
  - Shows: title, status badge, list of options with odds
  - Open: stake input + place button per option (only one selectable at a time)
  - Closed/settled: show result, highlight winning option, show my entry result
  - Effort: M

- [x] **3.3** — Add custom bet section to match detail page (wherever other bet forms are shown)
  - Fetch and render `WcCustomBetCard` list for the match
  - Effort: S

### Phase 4: History & Labels

- [x] **4.1** — Register `'custom'` in `WC_BET_TYPES` and add `"betTypeCustom": "Kèo phụ"` to `vi.json`
  - File: `frontend/src/utils/wcBetType.ts`, `frontend/src/locales/vi.json`
  - Effort: XS

- [ ] **4.2** — Add custom bet entries to player bet history view
  - Either extend `WcBetHistoryList.vue` or create `WcCustomBetHistoryList.vue`
  - Show: match name, bet title, chosen option, stake, payout, status
  - Effort: S

## Dependencies

- Phase 2 + 3 depend on Phase 1 (backend must be live)
- Phase 3.3 depends on knowing which page/component hosts match-level bet forms
- Phase 4.1 is independent (label registration)

## Timeline & Estimates

| Phase | Effort | Notes |
|-------|--------|-------|
| Phase 1 (Backend) | ~5h | Models + repo + service + handler + router |
| Phase 2 (Admin UI) | ~3h | Create form + settle dialog + integration |
| Phase 3 (Player UI) | ~3h | Card component + match page integration |
| Phase 4 (History) | ~1h | Label + history list |
| **Total** | **~12h** | Backend is the critical path |

## Risks & Mitigation

| Risk | Mitigation |
|------|-----------|
| Settlement transaction fails mid-way | Wrap entire settlement in `db.Transaction()`; return error and let caller retry |
| Wallet balance goes negative (race condition) | Check balance before deducting using `SELECT FOR UPDATE` or atomic update with WHERE balance >= stake |
| Player bets on a just-closed bet | `status == 'open'` check in service layer; also enforce in UI |
| Odds changed after entries exist | Odds snapshotted at entry time — admin change only affects future entries |
| Match page shows too many custom bets | Collapsible list; show max 3 by default, expand on click |
