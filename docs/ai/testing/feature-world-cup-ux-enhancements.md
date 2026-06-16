---
phase: testing
title: World Cup 2026 UX Enhancements — Testing Strategy
description: Manual and automated test scenarios for sidebar branding, smart redirect, and default filter
---

# Testing Strategy

## Scope

This feature is pure frontend UX — no new business logic, no new API endpoints. Testing is primarily manual/visual with a few integration checks for the router guard.

## Test Scenarios

### Sidebar Branding

| Scenario | Expected |
|---|---|
| Open app and view sidebar | World Cup nav item has green tinted background and `⚽` badge |
| Active state on World Cup routes | `.nav-item--active` styles override WC styles (green left border, theme bg) |
| Other nav items | No visual change |
| Mobile drawer | Same WC styling applies |

### Smart Redirect — `/world-cup`

| Scenario | Setup | Expected |
|---|---|---|
| Feature disabled | `is_enabled: false` from API | `/world-cup` loads `WcScheduleView` normally |
| Feature enabled, not logged in | Feature on, no `wc_token` in localStorage | Redirected to `/world-cup/login` |
| Feature enabled, logged in (regular user) | Feature on, `wc_token` set, `wc_user.isAdmin = false` | Redirected to `/world-cup/predict` |
| Feature enabled, logged in (admin) | Feature on, `wc_token` set, `wc_user.isAdmin = true` | Redirected to `/world-cup/admin` |
| API error on feature check | Network failure | Stays on schedule page (guard returns `true`) |
| Direct navigation to `/world-cup/predict` | Feature on, logged in | No change — route works as before |

### Default Filter on Predict Page

| Scenario | Expected |
|---|---|
| Open `/world-cup/predict` | "Mở dự đoán" filter pill is active (highlighted green) on first render |
| Matches with `status: open / predictions_open: true` | Visible immediately without any user action |
| Manual filter switch | Other filter pills still work correctly |
| No open matches | Empty state shows correctly under "Mở dự đoán" filter |

## Execution

Test manually against local dev server:
```bash
cd frontend && npm run dev
```

1. Toggle feature flag via admin panel or directly via `PATCH /wc/config`.
2. Clear/set `wc_token` / `wc_user` in DevTools → Application → Local Storage.
3. Navigate to `http://localhost:5173/world-cup` and verify redirects.
4. Navigate to `http://localhost:5173/world-cup/predict` and verify default filter.

## Risks & Gaps

- No automated unit tests for the `beforeEnter` guard — manual testing covers all cases.
- CSS visual regression not automated — review sidebar visually in both light and dark themes.
