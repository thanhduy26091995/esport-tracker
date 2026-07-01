---
phase: planning
title: WC Prediction Cancel & Reduce Penalty — Planning
description: Task breakdown và implementation order
---

# Project Planning & Task Breakdown

## Milestones

- [x] M1: Backend penalty logic + DB migrations
- [x] M2: API trả về penalty info + ListPredictions bao gồm cancelled
- [x] M3: Frontend confirm popup + Lịch sử hiển thị đúng

---

## Task Breakdown

### Phase 1: DB & Model

- [x] **1.1** Migration: `cancel_penalty` → `NUMERIC(10,2)` (wc_predictions, wc_bets, wc_custom_bet_entries)
- [x] **1.2** Migration: thêm `reduce_penalty NUMERIC(10,2) NOT NULL DEFAULT 0` vào `wc_predictions`
- [x] **1.3** Migration: drop standard unique index → tạo partial index `WHERE cancelled_at IS NULL`
- [x] **1.4** Go model `WcPrediction`: `CancelPenalty *float64`, `ReducePenalty float64`, `OriginalPoints *int`, `CancelledAt *time.Time`

### Phase 2: Backend Penalty Logic

- [x] **2.1** `wc_penalty.go`: `computeCancelPenalty` trả `float64` (không floor)
- [x] **2.2** `wc_penalty.go`: `computeReducePenalty` trả `(float64, int, int)` — chỉ penalty là float
- [x] **2.3** `SubmitPrediction`: set `original_points = points` khi tạo mới
- [x] **2.4** `SoftCancelPrediction`: nhận `penalty float64`, set `cancelled_at` + `cancel_penalty`
- [x] **2.5** `DeletePrediction` service: soft-cancel + wallet deduction (float) + wallet log
- [x] **2.6** `UpdatePredictionPointsWithPenalty` repo: update `points` + `reduce_penalty += penalty` trong 1 UPDATE
- [x] **2.7** `UpdatePredictionPoints` service: returns `(float64, error)`, gọi repo with penalty

### Phase 3: API

- [x] **3.1** `DELETE /predictions/:id` response: thêm `penalty_applied`
- [x] **3.2** `PUT /predictions/:id` response: thêm `penalty_applied`
- [x] **3.3** `GET /predictions/:id/reduce-preview`: trả `penalty` (float), `excess`, `allowed_min_stake`
- [x] **3.4** `GET /predictions`: **bỏ** filter `cancelled_at IS NULL` — trả cả active + cancelled
- [x] **3.5** `GET /wc/config`: expose 4 penalty fields trong response

### Phase 4: Analytics & Settlement Guards

- [x] **4.1** `GetLeaderboard`: thêm `AND b.cancelled_at IS NULL` trong pred subquery
- [x] **4.2** `wc_analytics_repository.go`: tất cả prediction queries thêm `AND p.cancelled_at IS NULL` (10 chỗ)
- [x] **4.3** Verify `ListPredictionsForMatch` (settlement) đã có filter `cancelled_at IS NULL`

### Phase 5: Frontend

- [x] **5.1** `wc.ts` types: thêm `cancelled_at`, `cancel_penalty`, `reduce_penalty`, `original_points` vào `WcPredictionWithMatch`
- [x] **5.2** `WcPredictionHistoryList.vue`: hiện prediction đã huỷ với badge "Đã huỷ" + dòng "Bị phạt: -N điểm"
- [x] **5.3** `WcPredictionHistoryList.vue`: confirm popup trước khi huỷ (khi `cancel_penalty_enabled && penalty > 0`)
- [x] **5.4** `WcPredictionForm.vue`: `confirmReduceIfNeeded()` helper — preview → confirm popup → abort nếu cancel
- [x] **5.5** `WcAdminPanel.vue`: 4 penalty config fields
- [x] **5.6** `wcStore`: expose `cancelPenaltyEnabled`, `cancelPenaltyPercent`, `betReduceMaxPercent`, `betReducePenaltyPercent`
- [x] **5.7** i18n: thêm keys `cancelPenaltyWarning`, `reducePenaltyWarning`, `reduceStakeTitle`, `cancelBetTitle`, `betCancelledBadge`

### Phase 6: Testing

- [x] **6.1** Unit test `computeCancelPenalty`: float output, enabled/disabled cases
- [x] **6.2** Unit test `computeReducePenalty`: float penalty, excess, allowedMin
- [x] **6.3** Verify partial unique index: cancel → re-predict cùng pick không bị duplicate error
- [x] **6.4** Verify settlement không bao gồm cancelled prediction
- [x] **6.5** Verify analytics queries không count cancelled prediction

---

## Dependencies

- Phase 2 phụ thuộc Phase 1 (model/migration phải có trước)
- Phase 3 phụ thuộc Phase 2 (service logic)
- Phase 5 phụ thuộc Phase 3 (API shape)
- Phase 4 có thể làm song song với Phase 3

---

## Risks & Mitigation

| Risk | Mitigation |
|------|-----------|
| Quên thêm `cancelled_at IS NULL` vào một query analytics | Grep toàn bộ `wc_predictions` usages trước khi done |
| `original_points` NULL trên prediction cũ → không tính reduce penalty | Backfill: `UPDATE wc_predictions SET original_points = points WHERE original_points IS NULL` |
| GORM tạo lại standard unique index sau khi ta drop | Đặt migration DROP+CREATE trước AutoMigrate, dùng cùng tên index |
| User thấy `reduce_penalty: 0` sau update nhưng thực ra bị trừ | Verify `UpdatePredictionPointsWithPenalty` được gọi đúng khi penalty > 0 |
