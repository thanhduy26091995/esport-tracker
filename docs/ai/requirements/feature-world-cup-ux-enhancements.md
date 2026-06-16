---
phase: requirements
title: World Cup 2026 UX Enhancements
description: Improve visibility of World Cup 2026 in sidebar, smart redirect on entry, and better default tab for predictions
---

# Requirements & Problem Understanding

## Problem Statement
**What problem are we solving?**

The World Cup 2026 section currently has three UX friction points:
1. The sidebar entry for World Cup 2026 looks identical to every other nav item — no branding, no visual differentiation — making it easy to miss despite being a featured, time-sensitive event.
2. When a logged-in user navigates to `/world-cup`, they land on the schedule page even when the feature flag is active and they are authenticated. They must manually click a second link to reach the predict page, adding unnecessary steps.
3. On the predict page (`/world-cup/predict`), the "Dự đoán" tab defaults to the "Sắp tới" (Upcoming) filter. Users who want to place predictions must find and click the "Mở dự đoán" (Open Predictions) filter themselves — creating friction at the most important action.

**Who is affected?**
All app users who participate in the World Cup 2026 prediction feature.

**Current workaround:**
Users manually navigate through extra clicks; the problem is pure friction, not a blocker.

## Goals & Objectives

**Primary goals:**
- Make the World Cup 2026 sidebar entry visually prominent and branded: green gradient background + a small `2026` pill badge.
- When `/world-cup` is visited and the WC feature flag is enabled, show a prominent "Vào trang dự đoán" CTA button so users can enter the predict flow in one click. No automatic redirect — user stays on schedule page but the action is unmissable.
- Remove the need for a "Lịch thi đấu" back-button in the predict page (no longer needed since schedule is always accessible at `/world-cup`).
- Change the default active filter inside the "Dự đoán" tab to "Mở dự đoán" (filter pill) so that matches open for prediction are visible immediately.
- Show empty state when no open matches exist — no automatic fallback to another filter.

**Non-goals:**
- Redesigning the entire sidebar layout.
- Changing the redirect behavior for users visiting `/world-cup/predict` directly.
- Adding a special admin redirect (admins access `/world-cup/admin` manually via URL).
- Adding backend changes — all changes are frontend only.

## User Stories & Use Cases

- As a user, I want the World Cup section in the sidebar to stand out so I can find it instantly without scanning through all nav items.
- As a logged-in user, I want visiting `/world-cup` to take me straight to the predict page (when the feature is live) so I don't need an extra click.
- As a not-logged-in user visiting `/world-cup` with the feature enabled, I want to be directed to login so I can participate.
- As a predictor, I want the "Mở dự đoán" filter to be selected by default so I immediately see matches I can bet on.

## Success Criteria

- [ ] World Cup nav item in sidebar has green gradient background and a `2026` pill badge, visually distinct from all other nav items.
- [ ] Visiting `/world-cup` with feature flag enabled → a prominent "Vào trang dự đoán" CTA button is visible on the schedule page.
- [ ] Visiting `/world-cup` with feature flag disabled → CTA button is hidden, schedule page loads normally (no regression).
- [ ] Schedule page always remains accessible at `/world-cup` — no navigation loop.
- [ ] Opening the predict page shows the "Mở dự đoán" filter pill active by default.
- [ ] When no matches are open, the empty state is shown (no fallback to another filter).
- [ ] All existing WC routes and auth guards continue to work.

## Constraints & Assumptions

- All changes are frontend-only (Vue 3 + Element Plus).
- Feature flag is checked via `GET /wc/config` → `is_enabled` (already implemented in router guards).
- Auth state is stored in `localStorage` (`wc_token`) and managed by `wcAuthStore`.
- The redirect on `/world-cup` should reuse the existing `isWcFeatureEnabled()` helper in the router.
- The `useMatchFilter` composable accepts the initial filter as a second parameter — changing it from `'incoming'` to `'open'` is the minimal change needed.

## Questions & Open Items

All questions resolved:
- **Sidebar badge:** Small `2026` pill badge with green styling.
- **Schedule page entry:** No automatic redirect. `WcScheduleView` detects feature flag and shows a prominent "Vào trang dự đoán" CTA button (logged in → predict, not logged in → login). Schedule always accessible at `/world-cup`.
- **Admin redirect:** No special case — all users navigate to admin panel via URL directly.
- **Empty state:** Keep as-is when no open matches exist, no automatic filter fallback.
