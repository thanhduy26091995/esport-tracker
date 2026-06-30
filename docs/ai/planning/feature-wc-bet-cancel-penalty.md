---
phase: planning
title: WC Bet Cancel Penalty — Planning
description: Cancel penalty + reduce stake penalty for all bet types (regular + custom)
---

# Project Planning & Task Breakdown

## Milestones

- [ ] **M1 — Backend foundation:** DB migrations, model changes, soft-cancel + penalty logic for regular and custom bets
- [ ] **M2 — Backend reduce stake:** Preview endpoint + UpdateBetStake penalty logic
- [ ] **M3 — Frontend config + cancel + reduce flows:** Admin config UI, cancel dialog, reduce stake popup
- [ ] **M4 — History tab:** Unified Lịch sử tab with cancelled + settled bets (regular + custom)

---

## Task Breakdown

### Phase 1: Backend — DB & Models

- [ ] **1.1** Add `cancelled_at TIMESTAMPTZ NULL`, `cancel_penalty INT NULL`, `original_stake INT NULL` to `WcBet` struct (`wc_match.go`)
- [ ] **1.2** Add same 3 fields to `WcCustomBetEntry` struct (`wc_custom_bet.go`)
- [ ] **1.3** Add 4 new fields to `WcConfig` struct (`wc_user.go`): `CancelPenaltyEnabled bool`, `CancelPenaltyPercent int`, `BetReduceMaxPercent int`, `BetReducePenaltyPercent int`
- [ ] **1.4** Change `WcWalletLog.AdminID` from `uuid.UUID` → `*uuid.UUID` (nullable); add explicit SQL migration `ALTER TABLE wc_wallet_logs ALTER COLUMN admin_id DROP NOT NULL`; update all `WcWalletLog{}` construction sites

### Phase 2: Backend — Repository

- [ ] **2.1** Add `SoftCancelBet(tx *gorm.DB, id, wcUserID uuid.UUID, penalty int) error` to `WcRepository` — sets `cancelled_at = NOW()`, `cancel_penalty = penalty`
- [ ] **2.2** Add `SoftCancelEntry(tx *gorm.DB, id, wcUserID uuid.UUID, penalty int) error` to `WcCustomBetRepository`
- [ ] **2.3** Add `CreateWalletLogSystem(tx *gorm.DB, log *model.WcWalletLog) error` to `WcRepository`
- [ ] **2.4** Add `GetWalletTx(tx *gorm.DB, wcUserID uuid.UUID) (*model.WcWallet, error)` to `WcRepository`
- [ ] **2.5** Add `ListBetHistoryForUser(wcUserID uuid.UUID) ([]*model.WcBet, error)` — `result IS NOT NULL OR cancelled_at IS NOT NULL`, ORDER BY `created_at DESC`
- [ ] **2.6** Add `ListCancelledOrSettledEntriesForUser(wcUserID uuid.UUID) ([]*model.WcCustomBetEntry, error)` to `WcCustomBetRepository`
- [ ] **2.7** Update `PlaceBet` (repo or service) to set `original_stake = stake` at creation
- [ ] **2.8** Update `PlaceEntry` (repo or service) to set `original_stake = stake` at creation
- [ ] **2.9** Audit + update all pending-bet queries that rely on `result IS NULL` to add `AND cancelled_at IS NULL` — specifically: `ListBetsForSettlement`, per-match open-bet query, `ListPendingBetsForUser` (used in BlockUser)

### Phase 3: Backend — Service

- [ ] **3.1** Rewrite `WcService.DeleteBet()`: soft-cancel + conditional cancel penalty (load config, compute `floor(stake × cancel_penalty_percent / 100)`, TX: SoftCancelBet + wallet deduction + wallet log if enabled && penalty > 0)
- [ ] **3.2** Update `WcCustomBetService.CancelEntry()`: same cancel penalty logic as 3.1
- [ ] **3.3** Add `WcService.PreviewReduceStake(wcUserID, betID uuid.UUID, newStake int) (penalty, excess, allowedMin int, err error)` — compute only, no writes
- [ ] **3.4** Update `WcService.UpdateBetStake()`: if `newStake < currentStake`, check excess vs `bet_reduce_max_percent`, apply penalty if over limit (TX: update stake + wallet deduction + wallet log)
- [ ] **3.5** Add `WcService.GetBetHistory(wcUserID uuid.UUID) ([]BetHistoryItem, error)` — merge `ListBetHistoryForUser` + `ListCancelledOrSettledEntriesForUser`, join match info, sort by `created_at DESC`
- [ ] **3.6** Update `UpdateConfig()` service to accept + validate all 4 new fields (`cancel_penalty_percent` 1–100, `bet_reduce_max_percent` 0–100, `bet_reduce_penalty_percent` 1–100)

### Phase 4: Backend — Handler & Router

- [ ] **4.1** Add `WcHandler.GetBetHistory` handler: `GET /wc/bets/history`
- [ ] **4.2** Add `WcHandler.PreviewReduceStake` handler: `GET /wc/bets/:id/reduce-preview?new_stake=N`
- [ ] **4.3** Update admin config handler (`PUT /admin/config`) to accept 4 new fields + return validation errors
- [ ] **4.4** Wire routes in `router.go`: `wcAuth.GET("/bets/history", ...)`, `wcAuth.GET("/bets/:id/reduce-preview", ...)`

### Phase 5: Frontend — Types & Service

- [ ] **5.1** Update `WcConfig` interface (`types/wc.ts`): add `cancel_penalty_enabled`, `cancel_penalty_percent`, `bet_reduce_max_percent`, `bet_reduce_penalty_percent`
- [ ] **5.2** Update `WcBet` interface: add `cancelled_at?`, `cancel_penalty?`, `original_stake?`
- [ ] **5.3** Add `BetHistoryItem` interface with `kind: 'regular' | 'custom'`, `bet_title?`, `cancel_penalty?`, etc.
- [ ] **5.4** Add `getBetHistory(): Promise<BetHistoryItem[]>` to `wcService.ts`
- [ ] **5.5** Add `previewReduceStake(betId, newStake): Promise<{penalty, excess, allowed_min_stake}>` to `wcService.ts`
- [ ] **5.6** Add i18n keys to `vi.json` + `en.json`: `cancelBetTitle`, `cancelPenaltyWarning`, `reduceStakeTitle`, `reducePenaltyWarning`, `betCancelledBadge`, `cancelPenaltyLabel`, `cancelPenaltyPercentLabel`, `reduceMaxPercentLabel`, `reducePenaltyPercentLabel`

### Phase 6: Frontend — Admin Config UI

- [ ] **6.1** Add cancel penalty toggle (`cancel_penalty_enabled`) + % input (`cancel_penalty_percent`, disabled when OFF) to `WcAdminPanel.vue` config form
- [ ] **6.2** Add reduce stake max % input (`bet_reduce_max_percent`, 0=no limit) + penalty % input (`bet_reduce_penalty_percent`, disabled when max=0) to config form
- [ ] **6.3** Ensure `updateConfig()` submit includes all 4 new fields

### Phase 7: Frontend — Cancel Flow

- [ ] **7.1** Regular bet cancel in `WcBettingView.vue`: compute `penalty = floor(stake × cancel_penalty_percent / 100)`; if `cancel_penalty_enabled && penalty > 0` → show `ElMessageBox.confirm` with penalty amount; then call `wcService.deleteBet()`
- [ ] **7.2** Custom bet entry cancel: same dialog logic before calling `wcService.cancelCustomBetEntry()`

### Phase 8: Frontend — Reduce Stake Flow

- [ ] **8.1** In bet stake edit component: when `newStake < currentStake`, call `wcService.previewReduceStake()` → if `preview.penalty > 0`, show `ElMessageBox.confirm` with penalty amount
- [ ] **8.2** On confirm: call `wcService.updateBetStake()` then refresh open bets

### Phase 9: Frontend — History Tab

- [ ] **9.1** Switch Lịch sử tab data source: call `wcService.getBetHistory()` (replaces per-match settled bet aggregation)
- [ ] **9.2** Update `WcBetHistoryList.vue` to accept `BetHistoryItem[]`; render `[Đã hủy]` badge + `cancel_penalty` display for cancelled items; show `bet_title` for custom items
- [ ] **9.3** Visual treatment for cancelled bets: muted/grey style, no payout display

---

## Dependencies

```
1.1, 1.2 → 2.1, 2.2, 2.5, 2.6, 2.7, 2.8, 2.9
1.3 → 3.1–3.6, 4.3, 6.1–6.3
1.4 → 2.3
2.1–2.9 → 3.1–3.5
3.1–3.6 → 4.1–4.4
4.1–4.4 → 5.4, 5.5
5.1–5.6 → 6.1–9.3
```

Backend phases (1–4) must complete before frontend phases (5–9).  
Within backend: models → repo → service → handler.

---

## Risks & Mitigation

| Risk | Likelihood | Mitigation |
|------|-----------|-----------|
| Pending-bet queries break after `cancelled_at` added | Medium | Task 2.9: audit all `result IS NULL` sites before merge |
| `WcWalletLog.AdminID` nullable breaks existing code | Low | Grep all `WcWalletLog{` construction sites; update to pointer |
| `original_stake = NULL` on existing bets breaks reduce logic | Low | Treat NULL as "no limit applies" in `PreviewReduceStake` and `UpdateBetStake` |
| History tab loses custom bet data during endpoint switch | Low | Test unified endpoint matches existing display before removing old aggregation |
