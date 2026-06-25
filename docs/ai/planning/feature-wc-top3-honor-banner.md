---
phase: planning
title: WC Top-3 Honor Banner — Planning
description: Task breakdown and implementation order for the animated top-3 honor banner on WC pages
---

# Project Planning & Task Breakdown

## Milestones

- [x] **M1:** Banner component renders top-3 cards with CSS marquee animation
- [x] **M2:** Banner wired into WC layout (visible on all WC authenticated pages)
- [x] **M3:** Auto-refresh + current-user highlight + i18n + responsive polish

## Task Breakdown

### Phase 1: Component Shell

- [x] **1.1** Create `WcTop3BannerCard.vue`
  - Props: `entry: WcLeaderboardEntry`, `rank: 1|2|3`, `isCurrentUser: boolean`
  - Render: medal emoji, avatar (28px), name (truncated 16 chars), net_points (+/- color)
  - CSS: card width ~180px, flex, gap, border-radius, gold/silver/bronze accent border

- [x] **1.2** Create `WcTop3Banner.vue`
  - Reads `wcStore.leaderboard` (slice 0–2)
  - Duplicates entries for seamless loop (`[...top3, ...top3]`)
  - Applies CSS `@keyframes marquee` (`translateX(0) → translateX(-50%)`)
  - `animation-play-state: paused` on wrapper hover
  - `prefers-reduced-motion` → disables animation, cards shown statically
  - Skeleton placeholder (3 grey pill cards) while `wcStore.leaderboardLoading` is true
  - `v-if="top3.length >= 1"` guard (hide if no data)

- [x] **1.3** Add i18n keys
  - `vi`: `wc.top3Banner.label = "🏆 BXH Top 3"`, `wc.top3Banner.pts = "điểm"`
  - `en`: `wc.top3Banner.label = "🏆 Top 3"`, `wc.top3Banner.pts = "pts"`

### Phase 2: Layout Integration

- [x] **2.1** Check if a shared WC layout wrapper exists
  - If yes: add `<WcTop3Banner />` before `<router-view />`
  - If no: create `frontend/src/layouts/WcLayout.vue` with banner slot

- [x] **2.2** (no separate WcLayout needed — `MainLayout.vue` already uses `isWcRoute` pattern) Wire WC authenticated routes to use `WcLayout`
  - In `frontend/src/router/index.ts`, add `meta: { layout: 'WcLayout' }` to: `wc-predict`, `wc-schedule` (authenticated), `wc-wallet`, `wc-custom-bet`
  - Confirm `App.vue` dynamic layout switching works (pattern already used in this project)

- [x] **2.3** (`wcAuth.isLoggedIn` guard in `MainLayout.vue` + `v-if` in banner) Verify banner does not appear on public/unauthenticated WC routes (`/wc/schedule` public, `/wc/login`)
  - Use `wcAuthStore.isLoggedIn` in `WcTop3Banner.vue` as extra guard

### Phase 3: Data & Polish

- [x] **3.1** Auto-refresh in `WcTop3Banner.vue`
  - `onMounted`: call `wcStore.fetchLeaderboard()` if `leaderboard.length === 0`
  - `setInterval(() => wcStore.fetchLeaderboard(), 5 * 60 * 1000)`
  - `onUnmounted`: `clearInterval`

- [x] **3.2** Current-user highlight
  - Compare `wcAuthStore.user?.id` with each `entry.wc_user_id`
  - Add CSS class `banner-card--me` → gold glow ring: `box-shadow: 0 0 0 2px #d97706`

- [x] **3.3** Responsive adjustments
  - Mobile (< 640px): card width 140px, name max-width 80px, font-size 11px
  - Ensure banner height stays at 56px on all breakpoints

- [x] **3.4** Accessibility pass
  - `role="marquee"` or `role="region"` + `aria-label` on banner wrapper
  - Cards: `aria-label="Rank 1: PlayerName, +12.50 pts"` (screenreader-friendly)

## Dependencies

- `wcStore.leaderboard` and `wcStore.fetchLeaderboard()` — **already exist**, no backend work
- `wcAuthStore.user.id` — **already exists** in wc auth store
- `WcLeaderboardEntry` type — **already defined** in `frontend/src/types/wc.ts`
- Layout pattern (`meta.layout`) — verify it exists in `App.vue`; may need to add if not already wired

## Timeline & Estimates

| Phase | Effort |
|-------|--------|
| Phase 1 (components + i18n) | ~2–3h |
| Phase 2 (layout wiring) | ~1h |
| Phase 3 (refresh + polish + a11y) | ~1h |
| **Total** | **~4–5h** |

## Risks & Mitigation

| Risk | Likelihood | Mitigation |
|------|-----------|------------|
| No shared WC layout wrapper exists | Medium | Check `App.vue` and router; create `WcLayout.vue` if needed (~30 min) |
| Marquee speed feels wrong on wide/narrow screens | Low | Use `animation-duration` proportional to content width via CSS custom property |
| Leaderboard API slow on first load → visible skeleton flash | Low | Banner already piggybacks on store data populated by page load; skeleton shown <1s typically |
| Name truncation breaks layout | Low | Hard cap at 16 chars + CSS `text-overflow: ellipsis` |

## Resources Needed

- Existing: `wcStore`, `wcService`, `WcLeaderboardEntry` type, CSS variables, i18n setup
- New files: `WcTop3Banner.vue`, `WcTop3BannerCard.vue`, optionally `WcLayout.vue`
- Locale files: `frontend/src/locales/vi.json`, `frontend/src/locales/en.json`
