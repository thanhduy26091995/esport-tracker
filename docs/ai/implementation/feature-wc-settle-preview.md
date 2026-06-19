---
phase: implementation
title: WC Settlement Preview — Implementation Guide
description: Technical notes for implementing the settle preview popup
---

# Implementation Guide

## Code Structure

```
backend/internal/model/wc_match.go          ← thêm FinalizePreview* DTOs
backend/internal/service/wc_service.go      ← thêm Preview* methods (reuse evaluate funcs)
backend/internal/api/wc_handler.go          ← thêm Preview* handlers
backend/internal/api/router.go              ← đăng ký routes (thứ tự!)

frontend/src/types/wc.ts                    ← thêm FinalizePreview* types
frontend/src/services/wcService.ts          ← thêm preview* methods
frontend/src/components/wc/
  WcFinalizePreviewDialog.vue               ← mới
  WcAdminPanel.vue                          ← refactor 3 button handlers
```

## Implementation Notes

### Core Features

**1. Preview Service Methods**

`PreviewFinalizeMatch` cần xử lý trường hợp trận chưa có điểm số:
```go
if m.HomeScore == nil || m.AwayScore == nil {
    return nil, fmt.Errorf("match score not set — cannot preview")
}
```

`PreviewRefinalizeAll` cần populate old values:
```go
pred := FinalizePreviewPrediction{
    OldResult:       bet.Result,        // *string từ WcPrediction
    OldPointsEarned: bet.PointsEarned,  // *float64 từ WcPrediction
    NewResult:       newResult,
    NewPointsEarned: newPointsEarned,
    NetDelta:        newPointsEarned - float64(bet.Points),
}
```

**2. Route Registration Order**

QUAN TRỌNG — trong `router.go`, routes phải theo thứ tự này:
```go
// Bulk routes TRƯỚC :id để tránh Gin parse "finalize-all-preview" như UUID param
adminGroup.GET("/matches/finalize-all-preview", wcHandler.PreviewFinalizeAll)
adminGroup.GET("/matches/refinalize-all-preview", wcHandler.PreviewRefinalizeAll)
adminGroup.GET("/matches/:id/finalize-preview", wcHandler.PreviewFinalizeMatch)
// existing routes...
adminGroup.POST("/matches/:id/finalize", wcHandler.FinalizeMatch)
```

**3. WcFinalizePreviewDialog Layout**

Dùng `el-collapse` cho match list (collapsible per match):
```html
<el-collapse>
  <el-collapse-item
    v-for="m in preview.matches"
    :key="m.match_id"
    :name="m.match_id"
  >
    <template #title>
      {{ m.home_team }} {{ m.home_score }} - {{ m.away_score }} {{ m.away_team }}
      <span class="wc-preview-pcount">({{ m.predictions.length }} dự đoán)</span>
    </template>
    <!-- prediction rows -->
  </el-collapse-item>
</el-collapse>
```

Result badge helper:
```ts
function resultBadge(result: string): string {
  if (result === 'correct' || result === 'win' || result === 'win_half') return '✅'
  if (result === 'push') return '➡️'
  if (result === 'lose_half') return '⬇️'
  return '❌'
}
```

Net delta formatting:
```ts
function fmtDelta(d: number): string {
  return (d >= 0 ? '+' : '') + d.toFixed(2)
}
```

**4. WcAdminPanel Integration**

`pendingAction` pattern để tránh duplicate code trong confirm handler:
```ts
type PendingAction = 'finalize-match' | 'finalize-all' | 'refinalize-all'
```

Khi mở preview dialog, set loading state trước khi fetch để UI responsive ngay:
```ts
previewDialogVisible.value = true   // mở dialog với skeleton
previewLoading.value = true
previewData.value = null
try {
  previewData.value = await wcService.previewFinalizeMatch(matchId)
} finally {
  previewLoading.value = false
}
```

## Integration Points

- Preview service reuses `evaluateHandicapPrediction`, `evaluateExactScorePrediction`, `evaluateOverUnderPrediction` — những functions này đã trong cùng package `service`
- `ListUnfinalizedScoredMatches()` và `ListAllScoredMatches()` đã có trong `wc_repository.go`
- Frontend gọi preview API qua `wcService` (cùng axios instance, có auth header)

## Error Handling

- Preview endpoint trả 404 nếu match không tồn tại
- Preview endpoint trả 422 nếu match chưa có điểm số (single match preview)
- Nếu 0 predictions → trả empty `matches` array, không phải lỗi
- Frontend: nếu preview API fail → đóng dialog, hiện ElMessage error

## Security Notes

- Tất cả preview endpoints đều dưới `WcAdminMiddleware` — chỉ admin gọi được
- Không có write operations trong preview paths
