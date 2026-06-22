---
phase: planning
title: WC Standalone Site (soc.sitenow.cloud) — Task Breakdown
description: Minimal task list to ship VITE_SITE=soc build with WC-only routes, nav, and branding
---

# Project Planning & Task Breakdown

## Milestones

- [x] **M1: Build config** — soc mode builds without errors; page title is "WC Prediction 2026"
- [x] **M2: Routing** — soc build has only WC routes; / redirects correctly; tree-shaking verified
- [x] **M3: Navigation + branding** — sidebar in soc build shows WC nav only + "World Cup 2026" logo title
- [x] **M4: App cleanup** — `userStore.fetchUsers()` skipped in soc mode
- [ ] **M5: Verification** — full smoke test on soc + regression check on fifa build

## Task Breakdown

### Phase 1: Build Config + TypeScript ✅

- [x] **1.1** Created `frontend/.env.soc` with VITE_SITE=soc, VITE_SITE_TITLE, VITE_API_BASE_URL, VITE_GOOGLE_CLIENT_ID
- [x] **1.2** Updated `frontend/.env` with VITE_SITE=fifa and VITE_SITE_TITLE=FC25 Tracker
- [x] **1.3** Added `"build:soc": "vue-tsc -b && vite build --mode soc"` and `"dev:soc": "vite --mode soc"` to `frontend/package.json`
- [x] **1.4** Created `frontend/src/vite-env.d.ts` with full `ImportMetaEnv` extension
- [x] **1.5** Updated `frontend/index.html`: `<title>%VITE_SITE_TITLE%</title>`
- [x] **1.6** `npm run build:soc` succeeds; dist/index.html title = "WC Prediction 2026" ✓

### Phase 2: Router — WC-only route set ✅

- [x] **2.1** Replaced flat routes array with inline ternary `isSocSite ? [...wc] : [...all]` in `frontend/src/router/index.ts`
- [x] **2.2** soc branch: WC paths + `{ path: '/', redirect: '/world-cup/login' }` + not-found
- [x] **2.3** fifa branch: existing full route list verbatim
- [x] **2.4** `vue-tsc --noEmit` — no errors; core view chunks absent from soc build ✓

### Phase 3: Sidebar Navigation + Logo Branding ✅

- [x] **3.1** Updated `frontend/src/layouts/MainLayout.vue`: added `isSocSite`, `appTitle`, `appSubtitle`; navigation is now conditional; both sidebar instances updated
- [x] **3.2** Added i18n keys to `en.json` (`appNameSoc`, `sidebarSubtitleSoc`) and `vi.json` (`Dự Đoán WC`, `World Cup 2026`)

### Phase 4: App.vue — Skip FC25 user fetch ✅

- [x] **4.1** Gated `useGlobalTheme` + `onMounted(fetchUsers)` behind `if (!isSocSite)` in `frontend/src/App.vue`; `userStore` kept at top level for safe Vue reactivity

### Phase 5: Smoke Test

- [ ] **5.1** `npm run dev:soc` → visit `/` → confirm redirect to `/world-cup/login`
- [ ] **5.2** Visit `/dashboard` → confirm 404/not-found page
- [ ] **5.3** Check browser tab title shows "WC Prediction 2026"
- [ ] **5.4** Check sidebar logo title shows "World Cup 2026" / "Dự Đoán WC"
- [ ] **5.5** WC full flow: login → predict → leaderboard — all guards working
- [ ] **5.6** Network tab: confirm no `GET /users` call on soc build
- [ ] **5.7** `npm run dev` → confirm full app unchanged, no regression

## Dependencies

- No backend changes required
- `.env.soc` `VITE_API_BASE_URL` must match the production backend URL
- Deployment of `soc.sitenow.cloud` pointing to same Go server (infra outside this task)

## Timeline & Estimates

| Phase | Estimate | Actual |
|-------|----------|--------|
| 1 — Build config + TypeScript | 30 min | ✅ |
| 2 — Router | 30 min | ✅ |
| 3 — Nav + branding | 25 min | ✅ |
| 4 — App.vue cleanup | 10 min | ✅ |
| 5 — Smoke test | 15 min | pending |

## Risks & Mitigation

| Risk | Status |
|------|--------|
| Tree-shaking doesn't eliminate core routes from soc bundle | ✅ Resolved — inline ternary used; verified no DashboardView/MatchesView in soc dist |
| `VITE_SITE` not typed correctly | ✅ Resolved — `vite-env.d.ts` created |
| `%VITE_SITE_TITLE%` not replaced | ✅ Resolved — verified in dist/index.html |
| WC `requiresWcFeature` guard redirects to `/world-cup` when feature is off | Acceptable — public page |
| `useGlobalTheme` called conditionally | ✅ Resolved — `userStore` stays top-level, only watch/onMounted gated |

## Resources Needed

- One frontend developer (or AI-assisted via `/execute-plan`)
- Hosting config to point `soc.sitenow.cloud` to same backend with the soc build artifact served as the frontend
