---
phase: requirements
title: WC Standalone Site (soc.sitenow.cloud)
description: Divorce the WC prediction feature from the core FC25 tracker by serving it as a separate build at soc.sitenow.cloud
---

# Requirements & Problem Understanding

## Problem Statement

Currently the WC prediction/betting game lives at `fifa.sitenow.cloud` alongside the core FC25 esport tracker. This forces two unrelated audiences to share the same URL:

- **Core players** — friends who track FC25 matches, debts, and scores
- **WC predictors** — friends who only want to bet on World Cup 2026 matches

WC-only users see the full sidebar (dashboard, match recording, fund, etc.), which is irrelevant and confusing to them. There is no clean entry point that says "this is just the WC prediction game."

**Goal:** Deploy a second domain `soc.sitenow.cloud` that serves an identical frontend build except WC-only routes are the only routes available. Core FC25 tracker routes are absent.

## Goals & Objectives

**Primary**
- `soc.sitenow.cloud` serves only WC prediction features (login, predict, schedule, leaderboard, champion prediction, admin)
- Core FC25 routes (`/`, `/dashboard`, `/match`, `/fund`, etc.) are not accessible at all
- Same backend, same PostgreSQL DB — no infra duplication

**Secondary**
- Landing on `/` at `soc.sitenow.cloud` automatically redirects to WC login
- Navbar/sidebar on `soc.sitenow.cloud` shows no FC25 links
- Sidebar logo title reads "World Cup 2026" (not "FC25 Tracker") in soc mode
- `userStore.fetchUsers()` is skipped in soc mode (no pointless FC25 user API call)

**Non-goals**
- Separate database or separate user accounts — WC users are the same `wc_users` table
- Full custom redesign/branding for `soc.sitenow.cloud`
- Backend changes — the API is already WC-isolated under `/api/v1/wc/`

## User Stories & Use Cases

| Story | Detail |
|-------|--------|
| WC predictor visits soc.sitenow.cloud | Lands on WC login page, never sees FC25 tracker UI |
| WC predictor tries to navigate to /dashboard | Route does not exist; stays on WC flow |
| Admin opens soc.sitenow.cloud | Gets WC admin panel at `/world-cup/admin` — no core admin routes |
| Core FC25 player visits fifa.sitenow.cloud | Unchanged — full app, WC widget still on dashboard |

## Success Criteria

- [ ] `VITE_SITE=soc npm run build` produces a build where only WC routes are registered in the Vue Router
- [ ] Visiting `/` on the soc build redirects to `/world-cup/login`
- [ ] Visiting any FC25 core URL (e.g. `/dashboard`) on the soc build returns a 404/not-found page
- [ ] `fifa.sitenow.cloud` behaviour is unchanged (VITE_SITE unset or `fifa`)
- [ ] The WC sidebar/nav on soc build contains no links to core esport features
- [ ] All existing WC routes and guards work correctly on the soc build
- [ ] Sidebar logo title shows "World Cup 2026" (not "FC25 Tracker") in soc build
- [ ] `userStore.fetchUsers()` is **not** called on soc build (verified via network tab)

## Constraints & Assumptions

- **Single codebase** — no monorepo split; one `frontend/` directory with build-time flag
- **Build-time env var** — `VITE_SITE=soc` (or `fifa` / unset for full app). Tree-shaking removes unused route imports
- **Same API origin** — both sites call the same backend; `VITE_API_BASE_URL` stays unchanged per deployment
- **No code duplication** — all WC components, stores, services remain in their current files
- **Vue Router** — only the route registration changes; guards themselves don't need modification

## Questions & Open Items

- [x] ~~Should `soc.sitenow.cloud` display a custom sidebar logo title?~~ → Yes, "World Cup 2026" via VITE_SITE check in MainLayout
- [x] ~~Should the soc build hide the WC upcoming-matches widget that appears on the core dashboard?~~ → Moot — core dashboard route won't exist on soc
- [x] ~~Should the 404 page "Back Home" link cause issues on soc?~~ → No, `to="/"` redirect chain works fine (`/` → `/world-cup/login`)
- [x] ~~Should `userStore.fetchUsers()` be skipped on soc?~~ → Yes, gate with `import.meta.env.VITE_SITE !== 'soc'` in App.vue
- [ ] CI/CD: who deploys the soc build? Deploy via `npm run build:soc` + point `soc.sitenow.cloud` to dist output (infra outside scope)
