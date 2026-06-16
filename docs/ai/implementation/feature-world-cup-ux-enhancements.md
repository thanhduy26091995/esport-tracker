---
phase: implementation
title: World Cup 2026 UX Enhancements — Implementation Guide
description: Exact code changes needed for sidebar branding, smart redirect, and default filter
---

# Implementation Guide

## Files to Change

| File | Change |
|---|---|
| `frontend/src/layouts/MainLayout.vue` | Add highlight flag + `2026` badge + CSS for WC nav item |
| `frontend/src/views/WcScheduleView.vue` | Fetch feature flag on mount, show prominent CTA button |
| `frontend/src/views/WcPredictView.vue` | Change default filter from `'incoming'` to `'open'` |

## Implementation Notes

### 1. MainLayout.vue — Sidebar Branding

Add `highlight: 'wc'` to the World Cup entry:
```ts
const navigation = [
  // ... other items ...
  { navKey: 'nav.worldCup', href: '/world-cup', icon: Promotion, highlight: 'wc' },
  // ...
]
```

Update the `router-link` template to bind the extra class and render the badge span:
```html
<router-link
  ...
  :class="{ 'nav-item--active': isActiveRoute(item.href), 'nav-item--wc': item.highlight === 'wc' }"
>
  <el-icon :size="18" class="nav-icon"><component :is="item.icon" /></el-icon>
  <span>{{ t(item.navKey) }}</span>
  <span v-if="item.highlight === 'wc'" class="nav-item-badge">2026</span>
</router-link>
```

Add CSS after the `.nav-item--active` block so active state wins:
```css
.nav-item--wc {
  background: linear-gradient(90deg, rgba(22, 163, 74, 0.15) 0%, transparent 100%);
  border: 1px solid rgba(22, 163, 74, 0.2);
  color: #16a34a;
}

.nav-item--wc .nav-icon {
  opacity: 1;
  color: #16a34a;
}

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

### 2. WcScheduleView.vue — Prominent CTA Button

Add `featureEnabled` ref and fetch on mount. Replace the existing small links in the header-right area:

```ts
// Add to <script setup>
const featureEnabled = ref(false)

onMounted(async () => {
  store.fetchMatches()
  try {
    const apiBase = (import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api/v1') + '/wc'
    const res = await fetch(`${apiBase}/config`)
    const data = await res.json()
    featureEnabled.value = !!data.is_enabled
  } catch { /* ignore */ }
})
```

Replace header-right template:
```html
<div class="wc-schedule-header-right">
  <template v-if="featureEnabled">
    <router-link
      :to="wcAuthStore.isLoggedIn ? { name: 'wc-predict' } : { name: 'wc-login' }"
      class="wc-cta-btn"
    >
      🏆 Vào trang dự đoán
    </router-link>
  </template>
  <template v-else>
    <router-link v-if="wcAuthStore.isLoggedIn && wcAuthStore.isAdmin" :to="{ name: 'wc-admin' }" class="wc-admin-link">
      {{ t('wc.adminPanel') }}
    </router-link>
    <router-link v-else-if="wcAuthStore.isLoggedIn" :to="{ name: 'wc-predict' }" class="wc-predict-link">
      {{ t('wc.predicting') }}
    </router-link>
    <router-link v-else :to="{ name: 'wc-login' }" class="wc-login-link">
      {{ t('wc.login') }}
    </router-link>
  </template>
</div>
```

Add CSS:
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

### 3. WcPredictView.vue — Default Filter

One-line change in `<script setup>`:
```ts
// Change:
useMatchFilter(storeMatches, 'incoming')
// To:
useMatchFilter(storeMatches, 'open')
```

## Integration Points

- `/wc/config` fetch in `WcScheduleView` is independent — does not reuse the router helper (that helper is router-scoped).
- `useMatchFilter` composable at `frontend/src/composables/useMatchFilter.ts` — second param is initial filter value.
- `wcAuthStore.isLoggedIn` / `wcAuthStore.isAdmin` used in template for CTA destination.

## Error Handling

- Feature flag fetch in `WcScheduleView` is wrapped in try/catch — `featureEnabled` stays `false` on error, so CTA simply doesn't appear (safe default).
