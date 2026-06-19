---
phase: planning
title: WC Settlement Preview — Planning
description: Task breakdown and implementation order for settle preview popup
---

# Project Planning & Task Breakdown

## Milestones

- [ ] **M1: Backend preview endpoints** — 3 GET endpoints trả dry-run result
- [ ] **M2: Frontend dialog component** — `WcFinalizePreviewDialog.vue` hiển thị đúng
- [ ] **M3: Integration** — 3 button trong `WcAdminPanel` mở dialog, confirm thực thi

## Task Breakdown

### Phase 1: Backend

- [ ] **1.1** Thêm DTOs vào `internal/model/wc_match.go`:
  - `FinalizePreviewPrediction`, `FinalizePreviewMatch`, `FinalizePreviewHouse`, `FinalizePreviewResult`

- [ ] **1.2** Thêm `PreviewFinalizeMatch(matchID)` vào `wc_service.go`:
  - Fetch match → validate score exists
  - Fetch predictions via `ListPredictionsForMatch`
  - Gọi existing evaluate functions (no DB write)
  - Build và return `FinalizePreviewResult`

- [ ] **1.3** Thêm `PreviewFinalizeAll()` vào `wc_service.go`:
  - Dùng `ListUnfinalizedScoredMatches()` (existing)
  - Loop qua từng match, build preview rows
  - Aggregate house summary

- [ ] **1.4** Thêm `PreviewRefinalizeAll()` vào `wc_service.go`:
  - Dùng `ListAllScoredMatches()` (existing)
  - Populate `OldResult` / `OldPointsEarned` cho predictions đã settle

- [ ] **1.5** Thêm 3 handlers vào `wc_handler.go`:
  - `PreviewFinalizeMatch`, `PreviewFinalizeAll`, `PreviewRefinalizeAll`

- [ ] **1.6** Đăng ký routes vào `router.go` (đúng thứ tự: bulk trước `:id`)

### Phase 2: Frontend

- [ ] **2.1** Thêm types vào `frontend/src/types/wc.ts`:
  - `FinalizePreviewPrediction`, `FinalizePreviewMatch`, `FinalizePreviewHouse`, `FinalizePreviewResult`

- [ ] **2.2** Thêm 3 service methods vào `wcService.ts`:
  - `previewFinalizeMatch`, `previewFinalizeAll`, `previewRefinalizeAll`

- [ ] **2.3** Tạo `WcFinalizePreviewDialog.vue`:
  - House summary section
  - Collapsible match list (el-collapse)
  - Per-prediction rows với result badge (✅/❌) và net delta
  - Empty state khi 0 predictions
  - Footer: Hủy + Xác nhận (confirm disabled khi loading)

### Phase 3: Integration

- [ ] **3.1** Sửa `WcAdminPanel.vue`:
  - Thêm `previewDialogVisible`, `previewData`, `previewLoading`, `previewConfirming`, `pendingAction`, `pendingMatchId`
  - Refactor `handleSettle`, `handleFinalizeAll`, `handleRefinalizeAll` → flow preview trước
  - Thêm `handleConfirmPreview` gọi actual action
  - Mount `<WcFinalizePreviewDialog>`

- [ ] **3.2** Manual testing:
  - Test single match preview (trận có điểm, trận không có điểm)
  - Test finalize-all với nhiều trận
  - Test refinalize-all (hiện old vs new)
  - Test confirm path thực sự thực thi
  - Test cancel path không thực thi gì

## Dependencies

- Task 1.2-1.4 depend on 1.1 (DTOs)
- Task 1.5-1.6 depend on 1.2-1.4
- Task 2.2 depends on 2.1 (types)
- Task 2.3 depends on 2.1
- Task 3.1 depends on 2.2, 2.3
- `ListUnfinalizedScoredMatches` và `ListAllScoredMatches` đã có trong repo — không cần thêm

## Timeline & Estimates

| Phase | Estimate |
|-------|----------|
| Phase 1 (Backend) | 2–3 giờ |
| Phase 2 (Frontend component) | 2–3 giờ |
| Phase 3 (Integration + test) | 1–2 giờ |
| **Total** | **5–8 giờ** |

## Risks & Mitigation

| Risk | Mitigation |
|------|-----------|
| Route conflict `finalize-all-preview` vs `:id` | Đăng ký bulk routes trước trong router.go |
| evaluate functions private (lowercase) | Chúng đều trong package `service` — preview methods cùng package nên gọi được |
| Preview diverge từ actual logic | Không duplicate logic — gọi lại cùng evaluate functions |
