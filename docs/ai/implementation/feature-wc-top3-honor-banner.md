---
phase: implementation
title: WC Top-3 Honor Banner — Implementation Guide
description: Code structure, CSS animation approach, and integration notes for the top-3 honor banner
---

# Implementation Guide

## Development Setup

No new dependencies. Uses existing:
- `wcStore` (`frontend/src/stores/wcStore.ts`)
- `wcAuthStore` (`frontend/src/stores/wcAuthStore.ts`)
- `WcLeaderboardEntry` (`frontend/src/types/wc.ts`)
- `vue-i18n` for strings
- Existing CSS custom properties

## Code Structure

```
frontend/src/
  components/wc/
    WcTop3Banner.vue         ← new: outer marquee wrapper
    WcTop3BannerCard.vue     ← new: single player card
  layouts/
    WcLayout.vue             ← new (if not exists): banner + <router-view>
  locales/
    vi.json                  ← add wc.top3Banner.* keys
    en.json                  ← add wc.top3Banner.* keys
  router/
    index.ts                 ← add meta.layout = 'WcLayout' to WC auth routes
```

## Implementation Notes

### CSS Marquee Animation

```css
/* Banner track: double-width so clone fills the gap */
.banner-track {
  display: flex;
  gap: 12px;
  width: max-content;
  animation: marquee 20s linear infinite;
}

@keyframes marquee {
  from { transform: translateX(0); }
  to   { transform: translateX(-50%); } /* -50% = exactly the original 3-card width */
}

.banner-track-wrapper:hover .banner-track {
  animation-play-state: paused;
}

@media (prefers-reduced-motion: reduce) {
  .banner-track { animation: none; }
}
```

The DOM renders `[...top3, ...top3]` (6 cards). At `-50%` translateX the visual resets to card 1 seamlessly.

### WcTop3Banner.vue — key logic

```typescript
const wcStore = useWcStore()
const wcAuthStore = useWcAuthStore()

const top3 = computed(() => wcStore.leaderboard.slice(0, 3))
// Clone for seamless marquee loop
const displayEntries = computed(() => [...top3.value, ...top3.value])

const currentUserId = computed(() => wcAuthStore.user?.id)

let refreshTimer: ReturnType<typeof setInterval> | null = null

onMounted(() => {
  if (wcStore.leaderboard.length === 0) wcStore.fetchLeaderboard()
  refreshTimer = setInterval(() => wcStore.fetchLeaderboard(), 5 * 60 * 1000)
})

onUnmounted(() => {
  if (refreshTimer) clearInterval(refreshTimer)
})
```

### Medal mapping

```typescript
const MEDALS: Record<number, string> = { 1: '🥇', 2: '🥈', 3: '🥉' }
// rank is 1-indexed (entry index + 1)
```

### Net points display

Reuse same formatter as `WcLeaderboard.vue`:
```typescript
function fmtPts(v: number): string {
  return parseFloat(v.toFixed(2)).toString()
}
// Display: "+12.5" or "-3.25"
const sign = entry.net_points >= 0 ? '+' : ''
```

### Layout wiring (App.vue dynamic layout)

Check `App.vue` — if it has a `<component :is="layout">` pattern, add `'WcLayout'` to the layout map. If the project uses a simpler route `component:` approach, nest WC routes under a layout component instead:

```typescript
// router/index.ts
{
  path: '/wc',
  component: WcLayout,           // <-- layout shell
  children: [
    { path: 'predict', component: WcPredictView, ... },
    { path: 'schedule', component: WcScheduleView, ... },
    // ...
  ]
}
```

Verify which pattern is already in use before implementing.

## Integration Points

- **`wcStore.fetchLeaderboard()`** — must return sorted entries. Confirm the API response is already sorted descending by `net_points`.
- **`wcAuthStore.user.id`** — check exact field name: may be `wcAuthStore.user?.wc_user_id` depending on how auth user is typed.

## Error Handling

- If `fetchLeaderboard()` fails: banner stays with stale data (last successful fetch). No error UI shown — failure is silent (same pattern as upcoming-matches widget).
- If leaderboard is empty: banner is hidden (`v-if="top3.length >= 1"`).

## Performance Considerations

- CSS animation runs on the compositor thread — zero main-thread cost.
- One `setInterval` per WcLayout mount (not per page navigation, since layout persists).
- No watchers on leaderboard array needed — `computed` reacts automatically.
