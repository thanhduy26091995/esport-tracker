---
phase: planning
title: World Cup 2026 UX Enhancements — Planning
description: Task breakdown and implementation order for sidebar branding, smart redirect, and default filter
---

# Project Planning & Task Breakdown

## Milestones

- [x] Milestone 1: Sidebar World Cup branding live
- [x] Milestone 2: Prominent CTA on schedule page when feature enabled
- [x] Milestone 3: Default "Mở dự đoán" filter on predict page

## Task Breakdown

### Phase 1: Sidebar Branding (MainLayout.vue)
- [x] Task 1.1: Add `highlight: 'wc'` property to the World Cup nav item in the `navigation` array
- [x] Task 1.2: Update `router-link` template to apply `.nav-item--wc` class and render inline `<span class="nav-item-badge">2026</span>` when `highlight === 'wc'`
- [x] Task 1.3: Add `.nav-item--wc` CSS (green gradient background + border + green icon)
- [x] Task 1.4: Add `.nav-item-badge` CSS for the `2026` pill
- [x] Task 1.5: Ensure `.nav-item--active` rule appears after `.nav-item--wc` so active state wins

### Phase 2: CTA Button on Schedule Page (WcScheduleView.vue)
- [x] Task 2.1: Fetch `/wc/config` on mount and store `featureEnabled` ref
- [x] Task 2.2: When `featureEnabled`, replace existing small header links with a prominent "🏆 Vào trang dự đoán" CTA button
- [x] Task 2.3: Button destination: `wcAuthStore.isLoggedIn` → `wc-predict`, else → `wc-login`
- [x] Task 2.4: Add CSS for `.wc-cta-btn` with green gradient, shadow, hover lift effect

### Phase 3: Default Filter (WcPredictView.vue)
- [x] Task 3.1: Change `useMatchFilter(storeMatches, 'incoming')` to `useMatchFilter(storeMatches, 'open')`

## Dependencies

- Phase 2 depends on the existing `isWcFeatureEnabled()` function — no changes needed there.
- Phase 3 depends on `useMatchFilter` composable supporting `'open'` as initial value — it already does.
- All three phases are independent and can be implemented in any order.

## Timeline & Estimates

| Task | Estimate |
|---|---|
| Phase 1 — Sidebar branding | ~30 min |
| Phase 2 — Smart redirect | ~20 min |
| Phase 3 — Default filter | ~5 min |
| **Total** | **~1 hour** |

## Risks & Mitigation

| Risk | Mitigation |
|---|---|
| `featureEnabled` fetch in WcScheduleView adds a small flicker (button appears after mount) | Acceptable — same pattern as existing loading states in the app |
| Changing default filter surprises users who expected "Sắp tới" | "Mở dự đoán" is strictly more useful; users can still manually switch |
| CSS specificity conflict between `.nav-item--wc` and `.nav-item--active` | Ensure `.nav-item--active` rule appears after `.nav-item--wc` in stylesheet |

## Resources Needed

- 1 frontend developer
- No backend changes
- No new dependencies
