---
phase: design
title: World Cup 2026 UX Enhancements — System Design
description: Technical design for sidebar branding, smart redirect, and default prediction filter
---

# System Design & Architecture

## Architecture Overview

All changes are confined to the Vue 3 frontend. No backend changes required.

```mermaid
graph TD
  User -->|visits /world-cup| WcScheduleView
  WcScheduleView -->|fetches /wc/config| isWcEnabled{Feature enabled?}
  isWcEnabled -->|No| ScheduleOnly[Schedule page only]
  isWcEnabled -->|Yes| ShowCTA[Show prominent CTA button]
  ShowCTA -->|logged in| WcPredictView[/world-cup/predict]
  ShowCTA -->|not logged in| WcLoginView[/world-cup/login]

  User -->|sidebar| MainLayout
  MainLayout -->|wc nav item| WcNavItem[Styled WC nav item — green gradient + 2026 badge]
```

## Component Breakdown

### 1. `MainLayout.vue` — Sidebar World Cup branding

**Change:** Add a special class/flag to the World Cup navigation item so it renders with distinct styling.

**Approach:**
- Add a `highlight: 'wc'` flag to the World Cup navigation item config.
- Apply a dedicated CSS class (`.nav-item--wc`) when the flag is set.
- Style with green gradient background + a small `2026` pill badge rendered via `::after` pseudo-element or an inline `<span>`.

```ts
// navigation array — add highlight flag
{ navKey: 'nav.worldCup', href: '/world-cup', icon: Promotion, highlight: 'wc' }
```

```css
/* Special WC nav item */
.nav-item--wc {
  background: linear-gradient(90deg, rgba(22,163,74,0.15) 0%, transparent 100%);
  border: 1px solid rgba(22,163,74,0.25);
  color: #16a34a;
}
.nav-item--wc .nav-icon { opacity: 1; color: #16a34a; }

/* 2026 pill badge */
.nav-item-badge {
  margin-left: auto;
  font-size: 10px;
  font-weight: 700;
  background: rgba(22, 163, 74, 0.18);
  color: #16a34a;
  padding: 1px 6px;
  border-radius: 10px;
  letter-spacing: 0.02em;
  flex-shrink: 0;
}
```

The badge is rendered as an inline `<span class="nav-item-badge">2026</span>` inside the `router-link` (not `::after`) so it participates in the flex layout naturally alongside the icon and label.

### 2. `WcScheduleView.vue` — Prominent CTA when feature is enabled

**Change:** No router changes. The schedule view itself detects the feature flag and renders a prominent CTA button in the header.

**Approach:**
- Call `wcService.getConfig()` (or reuse the store if available) on mount to check `is_enabled`.
- If enabled: show a large, styled CTA button in the header area.
  - Logged in → `router-link` to `wc-predict`
  - Not logged in → `router-link` to `wc-login`
- If disabled: button hidden, page unchanged.

```ts
// WcScheduleView.vue <script setup>
import { ref, onMounted } from 'vue'
const featureEnabled = ref(false)

onMounted(async () => {
  try {
    const res = await fetch(`${WC_API_BASE}/config`)
    const data = await res.json()
    featureEnabled.value = !!data.is_enabled
  } catch { /* ignore */ }
})
```

```html
<!-- Header right slot — replaces existing small links -->
<div v-if="featureEnabled" class="wc-cta">
  <router-link
    :to="wcAuthStore.isLoggedIn ? { name: 'wc-predict' } : { name: 'wc-login' }"
    class="wc-cta-btn"
  >
    🏆 Vào trang dự đoán
  </router-link>
</div>
```

```css
.wc-cta-btn {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 10px 20px;
  background: linear-gradient(135deg, #16a34a, #15803d);
  color: #fff;
  font-size: 14px;
  font-weight: 700;
  border-radius: 10px;
  text-decoration: none;
  box-shadow: 0 4px 14px rgba(22, 163, 74, 0.35);
  transition: box-shadow 0.15s, transform 0.1s;
}
.wc-cta-btn:hover {
  box-shadow: 0 6px 18px rgba(22, 163, 74, 0.45);
  transform: translateY(-1px);
}
```

### 3. `WcPredictView.vue` — Default filter "Mở dự đoán"

**Change:** One-line change — pass `'open'` as the initial filter to `useMatchFilter`.

```ts
// Before
const { search: betSearch, activeFilter: betFilter, filtered: betFiltered, counts: betCounts } =
  useMatchFilter(storeMatches, 'incoming')

// After
const { search: betSearch, activeFilter: betFilter, filtered: betFiltered, counts: betCounts } =
  useMatchFilter(storeMatches, 'open')
```

No "Lịch thi đấu" back-button needed — schedule is always accessible via `/world-cup` (no navigation loop).

## Design Decisions

| Decision | Choice | Rationale |
|---|---|---|
| Redirect logic placement | `beforeEnter` guard on `/world-cup` route | Keeps routing logic in the router, consistent with existing `beforeEach` guards |
| Sidebar highlight approach | CSS class + inline `<span>` badge | `::after` can't participate in flex layout; inline span positions badge naturally |
| Badge content | `2026` pill | Cleaner than emoji; branded without being noisy |
| Default filter | Change initial value from `'incoming'` to `'open'` | `useMatchFilter` already supports this; one-line, zero risk |
| WC entry point | CTA button in `WcScheduleView`, not router guard | Avoids navigation loop; schedule always accessible; simpler code |
| CTA destination | Logged in → predict, not logged in → login | Mirrors existing small links already in schedule header |

## Non-Functional Requirements

- **Performance:** `beforeEnter` makes one extra `fetch` to `/wc/config` only when `/world-cup` is navigated to directly — same cost as existing `beforeEach` guards for protected routes.
- **Security:** No new auth surface; reuses existing token/flag checks.
- **Regression risk:** Low — changes are additive (new CSS class) or redirect-only (guard), with a clear `!enabled → stay` fallback.
