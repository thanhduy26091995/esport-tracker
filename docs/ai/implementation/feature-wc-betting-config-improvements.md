---
phase: implementation
title: WC Betting Config Improvements — Implementation Guide
description: Implementation notes for configurable bet limits, handicap display, and consistent bet-type labels
---

# Implementation Guide

## Development Setup

No new dependencies. All changes use existing Go/GORM, Vue 3, Pinia, vue-i18n, and Element Plus stack.

Run backend: `cd backend && go run cmd/server/main.go`  
Run frontend: `cd frontend && npm run dev`  
Type-check: `cd frontend && npm run type-check`

## Code Structure

```
backend/
  internal/
    model/wc_user.go              ← WcConfig struct lives here — add MinPoints, MaxPoints
    api/wc_handler.go             ← extend GetPublicConfig response + UpdateConfig request
    service/wc_service.go         ← dynamic points validation for PlaceBet + UpdateBetStake
    service/wc_champion_service.go ← same for champion predictions
    repository/wc_repository.go   ← GetConfig() already exists — no change needed

frontend/src/
  stores/wcStore.ts              ← add minPoints, maxPoints to state (store already exists)
  components/wc/
    WcPredictionForm.vue         ← bind :min/:max to wcStore values (3 inputs)
    WcChampionPanel.vue          ← bind :min/:max on champion points input
    WcAdminPanel.vue             ← add "Giới hạn dự đoán" config UI section
    WcHandicapLine.vue           ← NEW component
    WcPredictionHistoryList.vue  ← integrate WcHandicapLine via wcStore match lookup
  utils/wcBetType.ts             ← NEW — canonical bet-type labels utility
```

## Implementation Notes

### Phase 1: Backend Configurable Limits

**Migration** — add columns with safe defaults so existing row is unchanged:
```sql
ALTER TABLE wc_config
  ADD COLUMN IF NOT EXISTS min_points INTEGER NOT NULL DEFAULT 1,
  ADD COLUMN IF NOT EXISTS max_points INTEGER NOT NULL DEFAULT 5;
```

**Model** (`backend/internal/model/wc_user.go` — WcConfig is defined here):
```go
// Add to existing WcConfig struct:
MinPoints int `gorm:"column:min_points;not null;default:1" json:"min_points"`
MaxPoints int `gorm:"column:max_points;not null;default:5" json:"max_points"`
```

**Handler DTOs** (`wc_handler.go`):
- `GetPublicConfig`: currently returns only `{"is_enabled": bool}` — extend to also return `min_points` and `max_points`.
- `UpdateConfig` request struct currently only has `IsEnabled bool` — extend with optional `MinPoints *int` / `MaxPoints *int`.

```go
// UpdateConfig request — extend existing struct
type updateWcConfigRequest struct {
    IsEnabled *bool `json:"is_enabled"`
    MinPoints *int  `json:"min_points"`
    MaxPoints *int  `json:"max_points"`
}
```

**Service validation** — three places to update in `wc_service.go`:
1. `PlacePrediction()` — remove Gin binding `max=5`; add:
2. `PlaceBet()` — replace `if req.Stake > 5` hardcoded check
3. `UpdateBetStake()` — same hardcoded `> 5` check

And one place in `wc_champion_service.go`. Pattern for all:
```go
cfg, err := s.repo.GetConfig()   // method is GetConfig(), not Get()
if err != nil { return nil, err }
if points < cfg.MinPoints || points > cfg.MaxPoints {
    return nil, fmt.Errorf("điểm cược phải từ %d đến %d", cfg.MinPoints, cfg.MaxPoints)
}
```

Audit all hardcoded max checks:
```bash
grep -rn 'max=5\|Stake > 5\|> 5' backend/internal/
```

### Phase 2: Frontend Dynamic Limits

**Pinia store** — add to existing `wcStore` (`frontend/src/stores/wcStore.ts`):
```ts
// Add to state:
minPoints: 1,  // default before fetch
maxPoints: 5,

// In fetchConfig() action (already fetches is_enabled), also read:
this.minPoints = res.data.min_points ?? 1
this.maxPoints = res.data.max_points ?? 5
```

**Form binding** — `WcPredictionForm.vue` (3 inputs) and `WcChampionPanel.vue` (1 input):
```vue
<el-input-number
  v-model="points"
  :min="wcStore.minPoints"
  :max="wcStore.maxPoints"
/>
```

**Admin UI** — in `WcAdminPanel.vue`, add a card section:
```vue
<el-card header="Giới hạn dự đoán">
  <el-form inline>
    <el-form-item label="Min điểm">
      <el-input-number v-model="betLimitForm.minPoints" :min="1" />
    </el-form-item>
    <el-form-item label="Max điểm">
      <el-input-number v-model="betLimitForm.maxPoints" :min="betLimitForm.minPoints" />
    </el-form-item>
    <el-button type="primary" @click="saveBetLimits">Lưu</el-button>
  </el-form>
</el-card>
```

### Phase 3: Handicap Display

**`WcHandicapLine.vue`** — core rendering logic:
```ts
const handicapText = computed(() => {
  if (props.homeHandicap && props.homeHandicap > 0) {
    return `${props.homeTeam} chấp ${props.awayTeam} ${props.homeHandicap} trái`
  }
  if (props.awayHandicap && props.awayHandicap > 0) {
    return `${props.awayTeam} chấp ${props.homeTeam} ${props.awayHandicap} trái`
  }
  return null  // level match or not yet configured — render nothing
})
```

**Integration in `WcPredictionHistoryList.vue`** — `WcPredictionWithMatch` does NOT include handicap fields. Look up the match from `wcStore` (already populated) by `match_id`:

```ts
// in script setup
const match = computed(() =>
  wcStore.matches.find(m => m.id === prediction.match_id)
)
```

```vue
<WcHandicapLine
  v-if="prediction.type === 'handicap' && match"
  :homeTeam="match.home_team"
  :awayTeam="match.away_team"
  :homeHandicap="match.home_handicap"
  :awayHandicap="match.away_handicap"
/>
```

No backend change required.

### Phase 4: Label Consistency — `wcBetType.ts`

Create `frontend/src/utils/wcBetType.ts` as the **single source of truth** for all bet type wording. Adding a future type (e.g. "Kèo góc") = edit this one file + one i18n key.

```ts
// frontend/src/utils/wcBetType.ts
import { useI18n } from 'vue-i18n'

export const WC_BET_TYPES = ['handicap', 'exact_score', 'over_under', 'corner'] as const
export type WcBetType = typeof WC_BET_TYPES[number]

export const BET_TYPE_I18N_KEYS: Record<WcBetType, string> = {
  handicap:    'betTypeHandicap',
  exact_score: 'betTypeExactScore',
  over_under:  'betTypeOverUnder',
  corner:      'betTypeCorner',      // pre-registered; used when "Kèo góc" ships
}

export function useWcBetTypeLabel() {
  const { t } = useI18n()
  return (type: WcBetType) => t(BET_TYPE_I18N_KEYS[type] ?? type)
}
```

**`vi.json`** — add the new canonical keys (and remove old `predictionType*` keys after replacing all usages):
```json
"betTypeHandicap":    "Kèo Handicap",
"betTypeExactScore":  "Kèo tỉ số",
"betTypeOverUnder":   "Kèo tài xỉu",
"betTypeCorner":      "Kèo góc"
```

**Usage in any component:**
```vue
<script setup>
import { useWcBetTypeLabel } from '@/utils/wcBetType'
const betTypeLabel = useWcBetTypeLabel()
</script>
<template>
  <span>{{ betTypeLabel(prediction.type) }}</span>
</template>
```

**Fix all 6 files — replace every label rendering with `betTypeLabel(type)`:**

```bash
# Find all hits to replace
grep -rn 'predictionTypeHandicap\|predictionTypeExactScore\|predictionTypeOverUnder\|betTypeHandicap\|betTypeExactScore\|"Tài Xỉu"' frontend/src --include="*.vue"
```

Per-file changes:

**`WcPredictionForm.vue`** — 3 tab labels:
```vue
<!-- Before -->
<el-tab-pane :label="t('wc.predictionTypeHandicap')" name="handicap">
<el-tab-pane label="Tài Xỉu" name="over_under">          <!-- hardcoded! -->
<el-tab-pane :label="t('wc.predictionTypeExactScore')" name="exact_score">

<!-- After -->
<el-tab-pane :label="betTypeLabel('handicap')" name="handicap">
<el-tab-pane :label="betTypeLabel('over_under')" name="over_under">
<el-tab-pane :label="betTypeLabel('exact_score')" name="exact_score">
```

**`WcBetForm.vue`** — 2 tab labels (currently "Chấp" / "Tỉ số"):
```vue
<!-- After -->
<el-tab-pane :label="betTypeLabel('handicap')" name="handicap">
<el-tab-pane :label="betTypeLabel('exact_score')" name="exact_score">
```

**`WcPredictionHistoryList.vue`** — badge (ternary → composable):
```vue
<!-- Before: ternary 3-branch -->
{{ pred.prediction_type === 'handicap' ? t('...') : pred.prediction_type === 'over_under' ? t('...') : t('...') }}

<!-- After -->
{{ betTypeLabel(pred.prediction_type) }}
```

**`WcMatchPredictionList.vue`** — badge (ternary 2-branch, **missing over_under — this is a bug**):
```vue
<!-- After — bug fixed, all 3 types handled -->
{{ betTypeLabel(pred.prediction_type) }}
```

**`WcMatchBetList.vue`** — badge (currently "Chấp" / "Tỉ số"):
```vue
<!-- After -->
{{ betTypeLabel(bet.bet_type) }}
```

**`WcBetHistoryList.vue`** — badge (currently "Chấp" / "Tỉ số"):
```vue
<!-- After -->
{{ betTypeLabel(bet.bet_type) }}
```

**`vi.json` cleanup** — remove old keys after all usages replaced:
```json
// Remove these:
"predictionTypeHandicap": "Kèo Handicap",
"predictionTypeExactScore": "Kèo tỉ số",
"predictionTypeOverUnder": "Kèo tài xỉu",
"betTypeHandicap": "Chấp",
"betTypeExactScore": "Tỉ số"

// Keep only the new canonical set:
"betTypeHandicap": "Kèo Handicap",
"betTypeExactScore": "Kèo tỉ số",
"betTypeOverUnder": "Kèo tài xỉu",
"betTypeCorner": "Kèo góc"
```

**TypeScript exhaustiveness:** `BET_TYPE_I18N_KEYS` is `Record<WcBetType, string>` — adding a type without a label is a compile error. Run `npm run type-check` to verify after changes.

## Integration Points

- `wcStore.fetchConfig()` (already called in router guard/app mount) — verify it also saves `min_points`/`max_points` after extending `GetPublicConfig`.
- `wcStore.matches` must be loaded before `WcPredictionHistoryList` renders — verify the parent view (`WcPredictView.vue`) fetches matches before rendering history tab. If lazy-loaded, `wcStore.matches.find()` will return `undefined` and `WcHandicapLine` will simply not render (graceful degradation).
- `wc_repository.go GetConfig()` is reused by service layer — confirm the method signature accepts no args and returns `(*model.WcConfig, error)`.

## Error Handling

- Config fetch failure: store keeps defaults (1, 5). Log warning. Do not block page render.
- Points out of range: backend 400 with message; frontend `el-input-number` constraints prevent this client-side anyway.

## Security Notes

- `PUT /admin/wc/config` is behind `WcAdminMiddleware` — no change needed.
- Min/max values are integers with server-side validation (`min >= 1`, `max >= min`).
