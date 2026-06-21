---
phase: implementation
title: Bug Fix Batch — 21 June 2026 — Implementation Guide
description: Exact file locations, code patterns, and implementation notes for all 8 fixes
---

# Implementation Guide

## Code Structure

All changes are in existing files — no new files need to be created except the DB migration script.

```
frontend/src/views/
  WcScheduleView.vue          ← bugs 1, 2
  WcPredictView.vue           ← bug 3

frontend/src/components/wc/
  WcChampionPanel.vue         ← bugs 5, 6 (frontend)

frontend/src/types/
  wc.ts                       ← bug 6 (TS types)

frontend/src/services/
  wcService.ts                ← bug 6 (API client)

backend/internal/model/
  wc_match.go                 ← bug 8 (JSON tags)
  wc_champion.go              ← bug 6 (GORM tag, model)

backend/internal/service/
  wc_service.go               ← bugs 4, 7

backend/internal/cron/
  wc_sync.go                  ← bug 7

backend/internal/repository/
  wc_champion_repository.go   ← bug 6

backend/internal/service/
  wc_champion_service.go      ← bug 6

backend/internal/api/
  wc_champion_handler.go      ← bug 6
  router.go                   ← bug 6 (route change)

database/migrations/
  YYYYMMDDHHMMSS_champion_multi_pick.sql  ← bug 6
```

---

## Implementation Notes

### Bug 1 — Auto-redirect (`WcScheduleView.vue`)

In `onMounted`, after `featureEnabled.value` is set from the config API, add:

```typescript
import { useRouter } from 'vue-router'
const router = useRouter()

// inside onMounted, after featureEnabled is set:
if (featureEnabled.value) {
  if (wcAuthStore.isLoggedIn) {
    router.replace({ name: 'wc-predict' })
  } else {
    router.replace({ name: 'wc-login' })
  }
}
```

Use `router.replace` (not `push`) so the schedule page is not added to history — back button won't loop.

**Guard check**: The router guard for `wc-predict` is `requiresWcAuth`. Because we only redirect if `wcAuthStore.isLoggedIn`, the guard will pass. No circular redirect risk.

---

### Bug 2 — Scroll to next match (`WcScheduleView.vue`)

Add `nextTick` import and a scroll helper:

```typescript
import { ref, computed, onMounted, watch, nextTick } from 'vue'

async function scrollToFirstMatchGroup() {
  await nextTick()
  const groups = document.querySelectorAll('.wc-date-group')
  if (groups.length > 0) {
    groups[0].scrollIntoView({ behavior: 'smooth', block: 'start' })
  }
}

// In onMounted, after selectedFilter is set:
selectedFilter.value = computeDefaultFilter(store.matches)
scrollToFirstMatchGroup()
```

Since `selectedFilter` change triggers `store.fetchMatches(filter)` via the `watch`, add a slight wait or call scroll after the fetch resolves. The `nextTick` handles DOM re-render.

---

### Bug 3 — Per-match collapse (`WcPredictView.vue`)

Replace these lines:

```typescript
// Remove:
const expandedMatchId = ref<string | null>(null)

// Add:
const expandedMatchIds = ref<Set<string>>(new Set())
const matchPredictionsMap = ref<Record<string, WcPrediction[]>>({})
```

Update `toggleMatchPredictions`:

```typescript
async function toggleMatchPredictions(matchId: string) {
  const ids = new Set(expandedMatchIds.value)
  if (ids.has(matchId)) {
    ids.delete(matchId)
    expandedMatchIds.value = ids
    return
  }
  ids.add(matchId)
  expandedMatchIds.value = ids
  if (!matchPredictionsMap.value[matchId]) {
    await store.fetchMatchPredictions(matchId)
    matchPredictionsMap.value[matchId] = [...store.matchPredictions]
    matchPredictionCounts.value[matchId] = store.matchPredictions.length
  }
}
```

Update template:

```html
<!-- Replace v-if="expandedMatchId === match.id" with: -->
<div v-if="expandedMatchIds.has(match.id)" class="wc-match-bets-panel">
  <WcMatchPredictionList :predictions="matchPredictionsMap[match.id] ?? []" />
</div>
```

Add the `WcPrediction` import if not already present:

```typescript
import type { WcMatchWithOdds, WcScoreMultiplier, WcMatch, WcPrediction } from '@/types/wc'
```

---

### Bug 4 — P&L guard (`wc_service.go`)

**Scope**: Guard applies to `finalize-match` and `finalize-all` only. `refinalize-all` shows full totals (all matches).

Add an `excludeSettled bool` parameter to `buildPreviewResult`:

```go
func buildPreviewResult(
    matches []*model.WcMatch,
    getPredictions func(uuid.UUID) ([]*model.WcPrediction, error),
    getUserName func(uuid.UUID) string,
    excludeSettled bool,  // ← new parameter
) (*model.FinalizePreviewResult, error) {
    // ...
    for _, m := range matches {
        // ...
        countInSummary := !excludeSettled || m.SettledAt == nil
        for _, bet := range bets {
            row := buildPreviewRow(bet, *m.HomeScore, *m.AwayScore)
            row.UserName = getUserName(bet.WcUserID)
            pm.Predictions = append(pm.Predictions, row)
            if countInSummary {
                result.HouseSummary.TotalStaked += float64(bet.Points)
                result.HouseSummary.TotalPaidOut += row.NewPointsEarned
                result.HouseSummary.PredictionCount++
            }
        }
        result.Matches = append(result.Matches, pm)
        if countInSummary {
            result.HouseSummary.MatchCount++
        }
    }
}
```

Callers:
```go
// PreviewFinalizeMatch  → buildPreviewResult(..., true)
// PreviewFinalizeAll    → buildPreviewResult(..., true)
// PreviewRefinalizeAll  → buildPreviewResult(..., false)
```

---

### Bug 5 — Champion responsive (`WcChampionPanel.vue`)

CSS changes in `<style scoped>`:

```css
/* Lower the two-column → single-column breakpoint */
@media (max-width: 700px) {
  .champion-main { grid-template-columns: 1fr; }
}
/* Remove or adjust the old 900px breakpoint if still present */

/* Team pick grid: single column on very small screens */
@media (max-width: 420px) {
  .teams-pick-grid { grid-template-columns: 1fr; }
}
```

Template change — wrap the predictions `<el-table>` in a scrollable div:

```html
<!-- Before -->
<el-table :data="allPredictions" size="small" max-height="300">

<!-- After -->
<div style="overflow-x: auto; -webkit-overflow-scrolling: touch;">
  <el-table :data="allPredictions" size="small" max-height="300">
    ...
  </el-table>
</div>
```

---

### Bug 6 — Multi-pick champion (full stack)

#### DB migration

Create `database/migrations/YYYYMMDDHHMMSS_champion_multi_pick.sql`:

```sql
-- Drop old single-user UNIQUE constraint
ALTER TABLE wc_champion_predictions
  DROP CONSTRAINT IF EXISTS wc_champion_predictions_user_id_key;

-- Add composite unique (user_id, team_id) — one pick per team per user
ALTER TABLE wc_champion_predictions
  ADD CONSTRAINT uq_champion_prediction_user_team UNIQUE (user_id, team_id);
```

#### `wc_champion.go` model

```go
type WcChampionPrediction struct {
    ID             uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
    // Change uniqueIndex to composite:
    UserID         uuid.UUID  `gorm:"type:uuid;not null;index;uniqueIndex:uq_champion_user_team" json:"user_id"`
    TeamID         uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:uq_champion_user_team" json:"team_id"`
    // rest unchanged...
}
```

Response types — add `ID` to the response struct used by the API (or use the model directly):

```go
// If there is a separate WcChampionPredictionResponse struct, add:
ID string `json:"id"`
```

#### `wc_champion_repository.go`

```go
// Old:
func (r *WcChampionRepository) GetMyPrediction(userID uuid.UUID) (*model.WcChampionPrediction, error)

// New:
func (r *WcChampionRepository) GetMyPredictions(userID uuid.UUID) ([]*model.WcChampionPrediction, error) {
    var preds []*model.WcChampionPrediction
    err := r.db.Where("user_id = ?", userID).Order("created_at ASC").Find(&preds).Error
    return preds, err
}

// Old: CreateOrUpdatePrediction (upsert)
// New:
func (r *WcChampionRepository) CreatePrediction(p *model.WcChampionPrediction) error {
    return r.db.Create(p).Error
    // DB enforces uq_champion_user_team; returns unique violation error if same team picked twice
}

// Old: DeletePrediction(userID uuid.UUID) — deletes all for user
// New:
func (r *WcChampionRepository) DeletePredictionByID(predictionID, userID uuid.UUID) error {
    return r.db.Where("id = ? AND user_id = ?", predictionID, userID).Delete(&model.WcChampionPrediction{}).Error
}
```

#### `wc_champion_service.go`

```go
func (s *WcChampionService) GetMyPredictions(userID uuid.UUID) ([]*model.WcChampionPrediction, error) {
    return s.repo.GetMyPredictions(userID)
}

func (s *WcChampionService) PlaceChampionPrediction(userID, teamID uuid.UUID, points int) (*model.WcChampionPrediction, error) {
    // validate config open, points range, etc. — same as before
    // changed: call CreatePrediction instead of CreateOrUpdatePrediction
    p := &model.WcChampionPrediction{...}
    if err := s.repo.CreatePrediction(p); err != nil {
        // on unique violation → return friendly error "already predicted this team"
    }
    return p, nil
}

func (s *WcChampionService) DeleteChampionPrediction(userID, predictionID uuid.UUID) error {
    return s.repo.DeletePredictionByID(predictionID, userID)
}

// SettleChampion — updated to:
// 1. Pay winners (add points × odds to wallet) — same as before
// 2. NEW: Deduct losers (subtract points_wagered from wallet)
// 3. Track settled_user_count (distinct users) alongside settled_count (total predictions)
func (s *WcChampionService) SettleChampion(adminID, winnerTeamID uuid.UUID) (*model.WcChampionSettleResult, error) {
    // ... same idempotency check and config validation ...

    preds, err := s.repo.ListPredictionsForSettle()
    // settled_user_count: count distinct WcUserIDs
    userSet := make(map[uuid.UUID]struct{})
    for _, p := range preds { userSet[p.WcUserID] = struct{}{} }

    result := &model.WcChampionSettleResult{
        Winner:           winnerTeam.Name,
        SettledCount:     len(preds),
        SettledUserCount: len(userSet),  // ← new field
    }

    db.Transaction(func(tx *gorm.DB) error {
        for _, p := range preds {
            isCorrect := p.TeamID == winnerTeamID
            if isCorrect {
                // Add payout to wallet (existing behavior)
                pointsEarned := int(float64(p.Points) * p.OddsSnapshot)
                s.wcRepo.UpdateWalletBalance(tx, p.WcUserID, float64(pointsEarned))
                // log wallet change ...
                result.CorrectCount++
                result.TotalPointsAwarded += pointsEarned
            } else {
                // NEW: Deduct stake from losers
                s.wcRepo.UpdateWalletBalance(tx, p.WcUserID, -float64(p.Points))
                // log wallet change ...
            }
            s.repo.SettlePrediction(tx, p.ID, resStr, pointsEarned)
        }
        return s.repo.MarkSettled(winnerTeamID)
    })
}
```

#### `wc_champion_handler.go`

```go
// GetMyPredictions handler
func (h *WcChampionHandler) GetMyPredictions(c *gin.Context) {
    userID := c.MustGet(middleware.WcUserIDKey).(uuid.UUID)
    preds, err := h.svc.GetMyPredictions(userID)
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch predictions"})
        return
    }
    // Return array (empty array if no predictions)
    if preds == nil { preds = []*model.WcChampionPrediction{} }
    c.JSON(http.StatusOK, preds)
}

// DeleteChampionPrediction handler
func (h *WcChampionHandler) DeleteChampionPrediction(c *gin.Context) {
    userID := c.MustGet(middleware.WcUserIDKey).(uuid.UUID)
    predID, err := uuid.Parse(c.Param("id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid prediction ID"})
        return
    }
    if err := h.svc.DeleteChampionPrediction(userID, predID); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete prediction"})
        return
    }
    c.JSON(http.StatusOK, gin.H{"ok": true})
}
```

#### `router.go` route change

```go
// Old:
wcPlayer.GET("/champion/my-prediction", wcChampionHandler.GetMyPrediction)
wcPlayer.DELETE("/champion/predict", wcChampionHandler.DeleteChampionPrediction)

// New:
wcPlayer.GET("/champion/my-predictions", wcChampionHandler.GetMyPredictions)
wcPlayer.DELETE("/champion/predictions/:id", wcChampionHandler.DeleteChampionPrediction)
```

#### Frontend `wc.ts` types

```typescript
export interface WcChampionPredictionMine {
  id: string          // ← add this
  team_name: string
  flag_emoji: string
  odds_snapshot: number
  points: number
  payout_if_correct: number
  result?: string
  points_earned?: number
}
```

#### Frontend `wcService.ts`

```typescript
async getMyChampionPredictions(): Promise<WcChampionPredictionMine[]> {
  const res = await wcApi.get('/champion/my-predictions')
  return res.data ?? []
},

async deleteChampionPrediction(predictionId: string): Promise<void> {
  await wcApi.delete(`/champion/predictions/${predictionId}`)
},
```

#### Frontend `WcChampionPanel.vue`

```typescript
// Before:
const myPrediction = ref<WcChampionPredictionMine | null>(null)

// After:
const myPredictions = ref<WcChampionPredictionMine[]>([])
```

Template — replace the single "My prediction card" with a `v-for`:

```html
<!-- My predictions list -->
<el-card
  v-for="pred in myPredictions"
  :key="pred.id"
  class="champion-card"
  shadow="never"
>
  <div class="my-pred-header">
    <span class="my-pred-label">Dự đoán của bạn</span>
    <el-button
      v-if="config?.is_open && !config.settled_at"
      size="small" type="danger" text
      @click="handleDelete(pred.id)"
    >Xóa</el-button>
  </div>
  <!-- ... same body as before ... -->
</el-card>

<!-- Place form condition changes: show if window open, regardless of existing predictions -->
<el-card v-if="config?.is_open && !config.settled_at" class="champion-card" shadow="never">
  <div class="form-title">🏆 Thêm dự đoán Vô địch</div>
  <!-- ... same form ... -->
</el-card>
```

```typescript
async function handleDelete(predictionId: string) {
  await ElMessageBox.confirm('Xóa dự đoán này?', 'Xác nhận', { type: 'warning' })
  try {
    await wcService.deleteChampionPrediction(predictionId)
    myPredictions.value = await wcService.getMyChampionPredictions()
    allPredictions.value = await wcService.getChampionPredictions()
    ElMessage.success('Đã xóa dự đoán')
  } catch { /* handled by interceptor */ }
}

async function handlePlace() {
  if (!selectedTeamId.value) { ElMessage.warning('Vui lòng chọn đội'); return }
  placing.value = true
  try {
    await wcService.placeChampionPrediction(selectedTeamId.value, selectedPoints.value)
    myPredictions.value = await wcService.getMyChampionPredictions()
    allPredictions.value = await wcService.getChampionPredictions()
    selectedTeamId.value = ''
    ElMessage.success('Đã đặt dự đoán!')
  } catch { /* handled by interceptor */ }
  finally { placing.value = false }
}
```

#### Frontend `WcChampionAdminPanel.vue`

Add `settled_user_count` to the settlement result type and update the display:

```typescript
// In wc.ts, update WcChampionSettleResult:
export interface WcChampionSettleResult {
  winner: string
  settled_count: number
  settled_user_count: number  // ← new
  correct_count: number
  total_points_awarded: number
}
```

In the admin panel template, replace:
```html
<!-- Before -->
<span>{{ settleResult.settled_count }} dự đoán đã tất toán</span>

<!-- After -->
<span>{{ settleResult.settled_count }} dự đoán từ {{ settleResult.settled_user_count }} người</span>
```

---

### Bug 7 — Smart cron (`wc_service.go` + `wc_sync.go`)

#### Add `GetMatchScheduleSummary` to `WcService`

```go
type MatchScheduleSummary struct {
    LiveCount           int
    NextScheduledAt     *time.Time
}

func (s *WcService) GetMatchScheduleSummary() (MatchScheduleSummary, error) {
    var liveCount int64
    if err := s.repo.DB().Model(&model.WcMatch{}).Where("status = ?", "live").Count(&liveCount).Error; err != nil {
        return MatchScheduleSummary{}, err
    }
    var nextMatch model.WcMatch
    err := s.repo.DB().Where("status = ? AND match_date > ?", "scheduled", time.Now()).
        Order("match_date ASC").First(&nextMatch).Error
    if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
        return MatchScheduleSummary{LiveCount: int(liveCount)}, err
    }
    sum := MatchScheduleSummary{LiveCount: int(liveCount)}
    if err == nil {
        sum.NextScheduledAt = &nextMatch.MatchDate
    }
    return sum, nil
}
```

Note: `s.repo.DB()` assumes the repository exposes the underlying `*gorm.DB`. If not, add a `CountMatchesByStatus(status string) (int64, error)` and `GetNextScheduledMatchDate() (*time.Time, error)` to the repository interface instead.

#### Update `wc_sync.go`

```go
func StartWcMatchSync(svc *service.WcService) {
    liveIntervalMinutes := 5
    if v := os.Getenv("WC_LIVE_SYNC_INTERVAL_MINUTES"); v != "" {
        if n, err := strconv.Atoi(v); err == nil && n > 0 {
            liveIntervalMinutes = n
        }
    }
    preMatchIntervalMinutes := 10
    if v := os.Getenv("WC_PRE_MATCH_INTERVAL_MINUTES"); v != "" {
        if n, err := strconv.Atoi(v); err == nil && n > 0 {
            preMatchIntervalMinutes = n
        }
    }
    idleIntervalMinutes := 30
    if v := os.Getenv("WC_SYNC_INTERVAL_MINUTES"); v != "" {
        if n, err := strconv.Atoi(v); err == nil && n > 0 {
            idleIntervalMinutes = n
        }
    }

    sync := func() {
        count, err := svc.SyncMatches()
        if err != nil {
            log.Printf("⚠️  WC sync failed: %v", err)
            return
        }
        log.Printf("✅ WC sync: %d matches upserted", count)
    }

    computeInterval := func() time.Duration {
        summary, err := svc.GetMatchScheduleSummary()
        if err != nil {
            return time.Duration(idleIntervalMinutes) * time.Minute
        }
        if summary.LiveCount > 0 {
            return time.Duration(liveIntervalMinutes) * time.Minute
        }
        if summary.NextScheduledAt != nil {
            if time.Until(*summary.NextScheduledAt) <= 2*time.Hour {
                return time.Duration(preMatchIntervalMinutes) * time.Minute
            }
        }
        return time.Duration(idleIntervalMinutes) * time.Minute
    }

    go func() {
        sync()
        for {
            interval := computeInterval()
            log.Printf("🔄 WC sync: next run in %v", interval)
            time.Sleep(interval)
            sync()
        }
    }()
}
```

---

### Bug 8 — Settlement user name (`wc_match.go`)

Two one-line changes:

```go
// WcSettlementDetailWithUser
type WcSettlementDetailWithUser struct {
    WcSettlementDetail
    Name string `json:"user_name"`  // was: `json:"name"`
}

// WcSettlementPreviewRow
type WcSettlementPreviewRow struct {
    WcUserID  uuid.UUID `json:"wc_user_id"`
    Name      string    `json:"user_name"`  // was: `json:"name"`
    Balance   float64   `json:"balance"`
    Direction string    `json:"direction"`
    Amount    float64   `json:"amount"`
}
```

---

## Error Handling

- **Bug 6 duplicate team error**: When `CreatePrediction` hits the `uq_champion_user_team` constraint, the DB returns a `duplicate key` error. The service should wrap this: `if isDuplicateKeyError(err) { return nil, errors.New("already predicted this team") }`. The handler returns HTTP 409.
- **Bug 6 delete ownership**: If `DeletePredictionByID` deletes 0 rows (wrong userID or wrong predictionID), return HTTP 404.
- **Bug 7 summary query fails**: On error, fall back to idle interval to avoid tight loops.

## Security Notes

- **Bug 6 delete**: Ownership check is enforced in the DB query (`WHERE id = ? AND user_id = ?`) — no separate fetch-then-delete pattern needed.
- No new auth surfaces introduced by any of the 8 fixes.
