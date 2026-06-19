---
phase: testing
title: WC Settlement Preview — Testing Strategy
description: Test scope for settle preview popup feature
---

# Testing Strategy

## Scope

- Unit: preview service methods (compute correct results without DB writes)
- Integration: API endpoints return correct preview shape
- Manual: dialog UX, confirm executes actual settle

## Test Files

| File | Layer | Coverage Target |
|------|-------|----------------|
| `backend/internal/service/wc_service_test.go` | Service | PreviewFinalizeMatch compute logic |

## Unit Tests

- `PreviewFinalizeMatch` với trận đã có điểm số → trả đúng result + net_delta per prediction
- `PreviewFinalizeMatch` với trận chưa có điểm số → trả error
- `PreviewFinalizeAll` với 0 matches chưa settle → trả empty matches, house_net = 0
- `PreviewRefinalizeAll` với match đã settle → `old_result` và `old_points_earned` populated
- House summary: `house_net = total_stake - total_points_out` đúng

## Integration Tests (Manual)

- `GET /admin/matches/:id/finalize-preview` với valid match ID → 200 + FinalizePreviewResult
- `GET /admin/matches/:id/finalize-preview` với match chưa có điểm → 422
- `GET /admin/matches/finalize-all-preview` → không confuse Gin với `:id` param
- `GET /admin/matches/refinalize-all-preview` → tương tự

## Execution

```bash
cd backend && go test ./internal/service/... -run TestPreview -v
```

## Risks & Gaps

- Evaluate functions không có separate unit tests — preview test gián tiếp cover chúng
- Không test concurrent preview calls (không cần — read-only)
