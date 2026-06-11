---
phase: implementation
title: WC2026 Upcoming Matches Dashboard Widget — Implementation Guide
description: Step-by-step implementation notes, code snippets, and integration details
---

# Implementation Guide

## Development Setup

No new dependencies needed. Uses existing Go and Vue tooling.

## Code Structure

```
backend/internal/
  repository/wc_repository.go   — add DateFrom/DateTo to MatchFilter + ListMatches WHERE clause
  api/wc_handler.go             — parse date_from/date_to query params

frontend/src/
  types/wc.ts                   — add date_from/date_to to WcMatchFilter
  services/
    wcPublicApi.ts              — NEW: bare axios instance (no error interceptor) + listMatchesPublic()
  components/wc/
    WcUpcomingWidget.vue        — NEW: header + horizontal scroll cards + hasMore trailing card
  views/
    DashboardView.vue           — fetch via wcPublicApi + render widget above champion-banner
```

## Implementation Notes

### Backend: `internal/repository/wc_repository.go`

**Step 1 — extend MatchFilter:**
```go
type MatchFilter struct {
    Status   string
    Stage    string
    Group    string
    Date     string // "YYYY-MM-DD" exact day (existing)
    DateFrom string // ISO8601 UTC, e.g. "2026-06-11T10:00:00Z" (new)
    DateTo   string // ISO8601 UTC (new)
}
```

**Step 2 — extend ListMatches query:**
```go
func (r *WcRepository) ListMatches(f MatchFilter) ([]*model.WcMatch, error) {
    q := r.db.Order("match_date ASC")
    if f.Status != "" { q = q.Where("status = ?", f.Status) }
    if f.Stage != ""  { q = q.Where("stage = ?", f.Stage) }
    if f.Group != ""  { q = q.Where("group_name = ?", f.Group) }
    if f.Date != ""   { q = q.Where("DATE(match_date AT TIME ZONE 'UTC') = ?", f.Date) }
    if f.DateFrom != "" { q = q.Where("match_date >= ?", f.DateFrom) }  // new
    if f.DateTo != ""   { q = q.Where("match_date <= ?", f.DateTo) }    // new
    var matches []*model.WcMatch
    return matches, q.Find(&matches).Error
}
```

### Backend: `internal/api/wc_handler.go`

**In `ListMatches` handler:**
```go
f := repository.MatchFilter{
    Status:   c.Query("status"),
    Stage:    c.Query("stage"),
    Group:    c.Query("group"),
    Date:     c.Query("date"),
    DateFrom: c.Query("date_from"),  // new
    DateTo:   c.Query("date_to"),    // new
}
```

### Frontend: `src/types/wc.ts`

```ts
export interface WcMatchFilter {
  status?: WcMatchStatus
  stage?: WcStage
  group?: string
  date?: string
  date_from?: string  // new
  date_to?: string    // new
}
```

### New: `src/services/wcPublicApi.ts`

```ts
import axios from 'axios'
import type { WcMatch, WcMatchFilter } from '@/types/wc'

const BASE = (import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1') + '/wc'

export const wcPublicApi = axios.create({
  baseURL: BASE,
  timeout: 10000,
  headers: { 'Content-Type': 'application/json' },
})
// No error interceptor — callers handle errors silently via try/catch

export async function listMatchesPublic(filter: WcMatchFilter = {}): Promise<WcMatch[]> {
  const r = await wcPublicApi.get<WcMatch[]>('/matches', { params: filter })
  return r.data
}
```

### Frontend: `src/components/wc/WcUpcomingWidget.vue`

```vue
<script setup lang="ts">
import { useRouter } from 'vue-router'
import type { WcMatch, WcStage } from '@/types/wc'

defineProps<{
  matches: WcMatch[]
  hasMore: boolean
}>()
const router = useRouter()

const STAGE_LABELS: Record<WcStage, string> = {
  group: 'Vòng bảng',
  r32: 'Vòng 32',
  r16: 'Vòng 16',
  qf: 'Tứ kết',
  sf: 'Bán kết',
  final: 'Chung kết',
  third_place: 'Tranh hạng 3',
}

function formatMatchTime(iso: string): string {
  return new Date(iso).toLocaleString('vi-VN', {
    timeZone: 'Asia/Ho_Chi_Minh',
    weekday: 'short',
    day: '2-digit',
    month: '2-digit',
    hour: '2-digit',
    minute: '2-digit',
  })
}

function stageLabel(stage: WcStage, groupName?: string): string {
  return groupName || STAGE_LABELS[stage] || stage
}
</script>

<template>
  <div class="wc-upcoming-widget">
    <div class="wc-upcoming-header">
      <span class="wc-upcoming-title">⚽ WC2026 — Sắp diễn ra</span>
      <router-link to="/world-cup/schedule" class="view-all-link">Xem lịch đầy đủ →</router-link>
    </div>
    <div class="wc-upcoming-list">
      <div
        v-for="m in matches"
        :key="m.id"
        class="wc-upcoming-card"
        :class="{ 'wc-upcoming-card--live': m.status === 'live' }"
        @click="router.push('/world-cup/schedule')"
      >
        <div class="wc-upcoming-meta">{{ stageLabel(m.stage, m.group_name) }}</div>
        <div class="wc-upcoming-teams">
          <span class="team">{{ m.home_team_code || m.home_team }}</span>
          <span class="vs">vs</span>
          <span class="team">{{ m.away_team_code || m.away_team }}</span>
        </div>
        <div class="wc-upcoming-time" :class="{ live: m.status === 'live' }">
          {{ m.status === 'live' ? '🔴 LIVE' : formatMatchTime(m.match_date) }}
        </div>
      </div>
      <!-- Trailing card when total > 5 -->
      <div
        v-if="hasMore"
        class="wc-upcoming-card wc-upcoming-card--more"
        @click="router.push('/world-cup/schedule')"
      >
        <span>Xem thêm →</span>
      </div>
    </div>
  </div>
</template>

<style scoped>
.wc-upcoming-widget {
  background: var(--color-bg-card, #1e2130);
  border-radius: 12px;
  padding: 12px 16px;
  border: 1px solid rgba(255, 255, 255, 0.08);
}
.wc-upcoming-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
}
.wc-upcoming-title {
  font-weight: 600;
  font-size: 0.9rem;
  color: var(--color-text-primary, #e2e8f0);
}
.view-all-link {
  font-size: 0.8rem;
  color: var(--color-primary, #4f8ef7);
  text-decoration: none;
}
.wc-upcoming-list {
  display: flex;
  gap: 10px;
  overflow-x: auto;
  padding-bottom: 4px;
  scrollbar-width: thin;
}
.wc-upcoming-card {
  flex: 0 0 auto;
  min-width: 140px;
  background: rgba(255,255,255,0.04);
  border: 1px solid rgba(255,255,255,0.08);
  border-radius: 8px;
  padding: 10px 12px;
  cursor: pointer;
  transition: border-color 0.2s, box-shadow 0.2s;
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.wc-upcoming-card:hover {
  border-color: rgba(79, 142, 247, 0.5);
}
.wc-upcoming-card--live {
  border-color: #16a34a;
  box-shadow: 0 0 0 2px rgba(22,163,74,0.25), 0 0 12px rgba(22,163,74,0.2);
  animation: glow-live 2s ease-in-out infinite alternate;
}
@keyframes glow-live {
  from { box-shadow: 0 0 0 2px rgba(22,163,74,0.2), 0 0 6px rgba(22,163,74,0.1); }
  to   { box-shadow: 0 0 0 2px rgba(22,163,74,0.4), 0 0 16px rgba(22,163,74,0.25); }
}
.wc-upcoming-meta {
  font-size: 0.7rem;
  color: var(--color-text-muted, #94a3b8);
  text-transform: uppercase;
  letter-spacing: 0.03em;
}
.wc-upcoming-teams {
  display: flex;
  align-items: center;
  gap: 6px;
  font-size: 0.85rem;
  font-weight: 600;
  color: var(--color-text-primary, #e2e8f0);
}
.vs { color: var(--color-text-muted, #94a3b8); font-weight: 400; font-size: 0.75rem; }
.wc-upcoming-time {
  font-size: 0.78rem;
  color: var(--color-text-secondary, #cbd5e1);
}
.wc-upcoming-time.live {
  color: #22c55e;
  font-weight: 600;
}
</style>
```

### Frontend: `DashboardView.vue`

**Imports (add):**
```ts
import { listMatchesPublic } from '@/services/wcPublicApi'  // bare axios, no toast interceptor
import type { WcMatch } from '@/types/wc'
import WcUpcomingWidget from '@/components/wc/WcUpcomingWidget.vue'
import { onUnmounted } from 'vue'  // add to existing vue import
```

**Reactive state (add):**
```ts
const upcomingMatches = ref<WcMatch[]>([])
const hasMoreUpcoming = ref(false)

async function fetchUpcomingWcMatches() {
  try {
    const now = new Date()
    const in48h = new Date(now.getTime() + 48 * 3600 * 1000)
    const all = await listMatchesPublic({
      date_from: now.toISOString(),
      date_to: in48h.toISOString(),
    })
    const filtered = all.filter(m => m.status === 'scheduled' || m.status === 'live')
    hasMoreUpcoming.value = filtered.length > 5
    upcomingMatches.value = filtered.slice(0, 5)
  } catch {
    // silent — no toast (wcPublicApi has no interceptor); widget stays hidden
  }
}

let refreshInterval: ReturnType<typeof setInterval> | null = null
```

**onMounted (update):**
```ts
onMounted(async () => {
  await Promise.all([
    // ... existing calls unchanged ...
  ])
  // WC widget — non-blocking, after main data
  fetchUpcomingWcMatches()
  refreshInterval = setInterval(fetchUpcomingWcMatches, 5 * 60 * 1000)
})

onUnmounted(() => {
  if (refreshInterval) clearInterval(refreshInterval)
})
```

**Template (add before `.champion-banner`, inside `.dashboard-grid`):**
```vue
<!-- WC upcoming: spans all columns, hidden when empty or loading -->
<WcUpcomingWidget
  v-if="upcomingMatches.length > 0"
  :matches="upcomingMatches"
  :has-more="hasMoreUpcoming"
  class="dashboard-full-width"
/>
```

**CSS (add to DashboardView.vue `<style>` if not already present):**
```css
.dashboard-full-width {
  grid-column: 1 / -1;
}
```

## Integration Points

- `listMatchesPublic` (new) — plain axios, no error interceptor; used only from Dashboard for silent background fetches.
- `GET /api/v1/wc/matches` is public and exempt from `WcFeatureMiddleware` — confirmed in router.

## Error Handling

- Backend: invalid ISO8601 string passed as `date_from`/`date_to` causes PostgreSQL error → 500; client catches silently (no toast).
- Frontend: any error (network, 500) → `upcomingMatches = []` → widget stays hidden. No ElMessage toast because `wcPublicApi` has no error interceptor.

## Security Notes

- Endpoint is already public (no auth required on `GET /matches`). No new security surface.
- `date_from`/`date_to` are used in parameterized queries (GORM `Where("match_date >= ?", ...)`) — safe from SQL injection.
