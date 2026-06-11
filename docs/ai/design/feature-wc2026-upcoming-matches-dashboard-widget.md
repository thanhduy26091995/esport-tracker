---
phase: design
title: WC2026 Upcoming Matches Dashboard Widget — System Design
description: Architecture, API changes, and component breakdown for the dashboard upcoming-matches widget
---

# System Design & Architecture

## Architecture Overview

```mermaid
graph TD
    DashboardView -->|on mount + every 5min| wcPublicApi.get
    wcPublicApi.get -->|GET /api/v1/wc/matches?date_from=&date_to=| WcHandler
    WcHandler -->|MatchFilter{DateFrom, DateTo}| WcService
    WcService --> WcRepository
    WcRepository -->|WHERE match_date >= date_from AND match_date <= date_to| DB[(PostgreSQL wc_matches)]
    DB --> WcRepository --> WcService --> WcHandler -->|[]WcMatch| DashboardView
    DashboardView -->|filter scheduled+live, slice 0-5| upcomingMatches
    DashboardView -->|v-if upcoming.length > 0| WcUpcomingWidget
    WcUpcomingWidget -->|v-for matches, card per match| MatchCards
```

**No new tables or services.** Changes touch:
1. `MatchFilter` struct (backend repository) — add `DateFrom`, `DateTo`
2. `ListMatches` query (backend repository) — apply date range filter
3. `ListMatches` handler (backend API) — parse `date_from` / `date_to` query params
4. `WcMatchFilter` type (frontend types/wc.ts) — add `date_from`, `date_to`
5. New file `src/services/wcPublicApi.ts` — bare axios instance (no error interceptor) for public WC calls from Dashboard
6. New Vue component `WcUpcomingWidget.vue`
7. `DashboardView.vue` — fetch + render widget above champion-banner

---

## Data Models

No schema changes. Uses existing `wc_matches` table.

Relevant fields for the widget:
```
wc_matches.id
wc_matches.home_team, away_team
wc_matches.home_team_code, away_team_code  (for flag emoji or code badge)
wc_matches.match_date                      (UTC, displayed as Asia/Ho_Chi_Minh)
wc_matches.status                          ('scheduled' | 'live' | 'completed' | 'cancelled')
wc_matches.group_name                      ('Group A' … or 'Quarter-final')
wc_matches.stage
wc_matches.home_score, away_score          (populated when live/completed)
```

---

## API Design

### Extended: `GET /api/v1/wc/matches`

**New optional query params:**

| Param | Type | Description |
|---|---|---|
| `date_from` | `string` ISO8601 UTC | Match `match_date >= date_from` |
| `date_to` | `string` ISO8601 UTC | Match `match_date <= date_to` |

**Existing params unchanged:** `status`, `stage`, `group`, `date` (single day).

**Example call from dashboard widget:**
```
GET /api/v1/wc/matches?date_from=2026-06-11T10:00:00Z&date_to=2026-06-13T10:00:00Z
```

Widget computes:
- `date_from = new Date(now - 4h).toISOString()` — looks back 4h so live matches (whose `match_date` is their kickoff time, in the past) are included
- `date_to = new Date(now + 48h).toISOString()`

No `status` filter sent — backend returns all matches in the window; frontend filters to `scheduled | live` and removes `completed`/`cancelled` client-side.

**Response:** unchanged — `[]WcMatch` sorted by `match_date ASC`.

### Backend changes — `MatchFilter` struct

```go
// internal/repository/wc_repository.go
type MatchFilter struct {
    Status   string
    Stage    string
    Group    string
    Date     string // "YYYY-MM-DD" — exact day filter (existing)
    DateFrom string // ISO8601 UTC — range start (new)
    DateTo   string // ISO8601 UTC — range end (new)
}
```

`ListMatches` query additions:
```go
if f.DateFrom != "" {
    q = q.Where("match_date >= ?", f.DateFrom)
}
if f.DateTo != "" {
    q = q.Where("match_date <= ?", f.DateTo)
}
```

### Backend changes — handler

```go
// internal/api/wc_handler.go — ListMatches
f := repository.MatchFilter{
    Status:   c.Query("status"),
    Stage:    c.Query("stage"),
    Group:    c.Query("group"),
    Date:     c.Query("date"),
    DateFrom: c.Query("date_from"),  // new
    DateTo:   c.Query("date_to"),    // new
}
```

### Frontend changes — `WcMatchFilter` type

```ts
// src/types/wc.ts
export interface WcMatchFilter {
  status?: WcMatchStatus
  stage?: WcStage
  group?: string
  date?: string
  date_from?: string  // new — ISO8601 UTC string
  date_to?: string    // new — ISO8601 UTC string
}
```

### New: `src/services/wcPublicApi.ts`

A bare axios instance with **no error interceptor** — used by `DashboardView` for silent background fetches that must not show toasts on failure. `wcApi` (with full error handling) remains unchanged for WC pages.

```ts
// src/services/wcPublicApi.ts
import axios from 'axios'
import type { WcMatch, WcMatchFilter } from '@/types/wc'

const BASE = (import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1') + '/wc'

// Bare instance — no error interceptors; callers handle errors silently
export const wcPublicApi = axios.create({
  baseURL: BASE,
  timeout: 10000,
  headers: { 'Content-Type': 'application/json' },
})

export async function listMatchesPublic(filter: WcMatchFilter = {}): Promise<WcMatch[]> {
  const r = await wcPublicApi.get<WcMatch[]>('/matches', { params: filter })
  return r.data
}
```

`DashboardView` imports `listMatchesPublic` directly — never touches `wcService` or `wcApi` for this call.

---

## Component Breakdown

### `WcUpcomingWidget.vue` (new)

**Location:** `src/components/wc/WcUpcomingWidget.vue`

**Props:**
```ts
defineProps<{
  matches: WcMatch[]
  hasMore: boolean   // true when total filtered > 5; shows "Xem thêm →" after card list
}>()
```

**Responsibilities:**
- Render a horizontal scrollable list of match cards (max 5 — sliced by parent).
- Each card shows: group/stage label, team codes, match time in VN timezone or 🔴 LIVE badge.
- Header always shows "Xem lịch đầy đủ →" link to `/world-cup/schedule`.
- When `hasMore = true`: renders a trailing "+ more" card or "Xem thêm →" text link after the card list.
- Clicking any card (or "Xem thêm") navigates to `/world-cup/schedule`.
- No loading state — parent controls visibility via `v-if`.

**Template sketch:**
```vue
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
        @click="$router.push('/world-cup/schedule')"
      >
        <div class="wc-upcoming-meta">{{ m.group_name || stageLabel(m.stage) }}</div>
        <div class="wc-upcoming-teams">
          <span class="team">{{ m.home_team_code || m.home_team }}</span>
          <span class="vs">vs</span>
          <span class="team">{{ m.away_team_code || m.away_team }}</span>
        </div>
        <div class="wc-upcoming-time" :class="{ live: m.status === 'live' }">
          {{ m.status === 'live' ? '🔴 LIVE' : formatMatchTime(m.match_date) }}
        </div>
      </div>
      <!-- Trailing "more" card when capped -->
      <div v-if="hasMore" class="wc-upcoming-card wc-upcoming-card--more" @click="$router.push('/world-cup/schedule')">
        <span>Xem thêm →</span>
      </div>
    </div>
  </div>
</template>
```

**`formatMatchTime`:**
```ts
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
```

---

### `DashboardView.vue` changes

1. **Imports** (add):
   ```ts
   import { listMatchesPublic } from '@/services/wcPublicApi'  // bare axios, no toast
   import type { WcMatch } from '@/types/wc'
   import WcUpcomingWidget from '@/components/wc/WcUpcomingWidget.vue'
   import { onUnmounted } from 'vue'
   ```
2. **Add refs:**
   ```ts
   const upcomingMatches = ref<WcMatch[]>([])
   const hasMoreUpcoming = ref(false)
   ```
3. **On `onMounted`**: fetch with 48h window, set up auto-refresh:
   ```ts
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
       // silent — no toast (wcPublicApi has no interceptor)
     }
   }

   let refreshInterval: ReturnType<typeof setInterval> | null = null
   onMounted(() => {
     fetchUpcomingWcMatches()
     refreshInterval = setInterval(fetchUpcomingWcMatches, 5 * 60 * 1000)
   })
   onUnmounted(() => {
     if (refreshInterval) clearInterval(refreshInterval)
   })
   ```
4. **Template** — insert before `.champion-banner` (inside `.dashboard-grid`):
   ```vue
   <!-- WC upcoming matches: spans all columns, hidden when empty or loading -->
   <WcUpcomingWidget
     v-if="upcomingMatches.length > 0"
     :matches="upcomingMatches"
     :has-more="hasMoreUpcoming"
     class="dashboard-full-width"
   />
   ```
   `.dashboard-full-width` → `grid-column: 1 / -1` (same pattern as `.champion-banner`; add to `<style>` if not already present).

---

## Design Decisions

| Decision | Choice | Rationale |
|---|---|---|
| Widget always visible (no WC flag check) | `GET /matches` is exempt from WcFeatureMiddleware; silent fail on any error | Schedule info is public; confirmed in router — matches route outside wcFeature group |
| Status filter: scheduled + live | Client-side filter after fetch | Live matches are the most urgent; excluding them would miss the key use case |
| `completed`/`cancelled` excluded | Client filter `.filter(m => m.status === 'scheduled' \|\| m.status === 'live')` | Non-goal: don't show completed matches |
| Cap at 5 matches | `.slice(0, 5)` + "Xem thêm" link | Keeps widget compact; group stage can have 8 matches/day |
| Auto-refresh every 5 minutes | `setInterval` + `clearInterval` on unmount | Match status changes (scheduled→live) within the 48h window; 5min is timely without hammering API |
| Silent fail — use `wcPublicApi` not `wcApi` | Bare axios instance without error interceptor for Dashboard calls | `wcApi` interceptor calls `ElMessage.error()` automatically; background 5-min refresh would show toasts on network blips — unacceptable on Dashboard |
| Silent fail on API error | `try/catch` → `upcomingMatches = []` → widget hidden | Dashboard should never break if WC backend is off |
| `date_from`/`date_to` added to backend | Scales correctly; avoids over-fetching | WC2026 has 64+ matches; fetching all to filter client-side wastes bandwidth |
| Widget placed in same `dashboard-grid` | Uses existing `grid-column: 1/-1` pattern | Consistent with champion-banner layout; no new layout wrapper needed |
| Loading state: invisible (not skeleton) | `v-if="upcomingMatches.length > 0"` controls render | Simple; avoids layout shift; widget "appears" once data arrives |

---

## Non-Functional Requirements

- **Performance:** Single query with indexed `match_date` column; < 50ms response.
- **Resilience:** Widget fails silently — Dashboard never shows an error state for this widget.
- **Timezone:** All times displayed in `Asia/Ho_Chi_Minh` via `Intl.DateTimeFormat`.
- **Responsiveness:** Horizontal scroll on mobile; grid/wrap on desktop.
