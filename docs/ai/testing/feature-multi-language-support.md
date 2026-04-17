---
phase: testing
title: Testing Strategy & Validation
description: Define test scope, coverage, and validation criteria for the feature
feature: multi-language-support
status: draft
created: 2026-04-17
---

# Testing Strategy

## Scope
**What should be tested for this feature?**

- Unit test scope
  - Locale utility behavior (default locale, persisted locale parsing, fallback).
  - Error translation mapping helper.
- Integration test scope
  - i18n plugin wiring with app bootstrap.
  - Language switcher updates visible text in mounted components.
- End-to-end test scope (if applicable)
  - User switches vi/en and sees major pages translated.

## Test Files
**Which files contain tests?**

| File | Package/Layer | Coverage Target |
|------|---------------|----------------|
| `frontend/src/utils/__tests__/locale.test.ts` | frontend utils | locale util core branches >= 90% |
| `frontend/src/plugins/__tests__/i18n.test.ts` | frontend plugin | initialization/fallback behavior >= 85% |
| `frontend/src/components/common/__tests__/LanguageSwitcher.test.ts` | frontend component | switch + persistence behavior >= 90% |

## Unit Tests
**What logic should be validated in isolation?**

- Restores `vi` as default when no preference exists.
- Accepts only supported locales (`vi`, `en`) from storage.
- Rejects invalid stored locale and falls back safely.
- Error key mapping returns localized string for known codes.
- Unknown error code returns safe fallback message.

## Integration Tests
**What cross-component behaviors should be validated?**

- Changing locale updates text bound by `t(...)` in active view.
- Language switch persists and survives page refresh.
- Missing translation key falls back to default locale string.

## Manual QA Checklist
**What should be manually verified on desktop and mobile?**

- Global
  - Switch locale from `vi` to `en` and back using the main language switcher.
  - Refresh the page and confirm the selected locale persists.
  - Navigate across all major routes and confirm text updates without requiring a full reload.
  - Verify no raw translation keys are shown in the UI.
  - Verify common buttons, empty states, dialogs, badges, tooltips, toasts, and validation messages are translated.

- Desktop
  - Check sidebar and top navigation labels in both locales.
  - Open Dashboard and verify cards, recent activity blocks, empty states, and transaction labels.
  - Open Players and verify table columns, filters, dialogs, settlement actions, and confirmation messages.
  - Open Matches and verify filters, warnings, forms, delete flow, and recent match labels.
  - Open Settlements and verify info banner, cards, detail dialog, and manual settlement dialog.
  - Open Fund and verify balance hero, transaction history, filter labels, and deposit/withdraw dialogs.
  - Open Settings and verify all config cards, hints, previews, summary panel, and reset/save toasts.

- Mobile
  - Verify the language switcher remains usable in the mobile header.
  - Check that long English strings do not overflow cards, tables, dialogs, or badges.
  - Confirm dialogs remain readable and buttons stay visible without clipping.
  - Verify list cards in Dashboard, Settlements, and Fund remain scannable in both locales.
  - Confirm config previews and settlement explanations wrap correctly on narrow screens.

- Locale-aware formatting
  - In `vi`, verify currency, numbers, dates, and relative time appear in Vietnamese formatting.
  - In `en`, verify currency, numbers, dates, and relative time switch to English formatting.
  - Confirm the same underlying values render consistently across Dashboard, Fund, Settlements, Users, and Tournaments.

- Error handling
  - Trigger client-side validation errors in user, match, fund, config, and tournament forms.
  - Trigger a known server-side error path and confirm the mapped localized error message is shown.
  - Simulate or force a network failure and confirm the localized network error toast appears.

- Pass criteria
  - No visible untranslated strings in completed feature areas.
  - No layout breakage in `vi` or `en` on desktop or mobile.
  - Locale persists after refresh and route changes.
  - Dynamic interpolation values such as `{name}`, `{amount}`, `{count}`, `{percent}`, and `{threshold}` render correctly.

## Test Data & Environments
**What data and setup are required?**

- Fixtures/seeds needed
  - Minimal mock dictionaries for `vi` and `en`.
- Environment variables/config
  - Test runtime with jsdom for component tests.
- Local vs CI considerations
  - CI should run locale tests in same suite as frontend unit tests.

## Execution
**How do we run and verify tests?**

- Commands to run tests
  - `cd frontend && npm test`
- Commands to run coverage
  - `cd frontend && npm test -- --coverage`
- Expected pass criteria
  - All new i18n tests pass, no regression in existing test suites.

## Coverage & Quality Gates
**What quality bar must be met?**

- Coverage targets
  - New i18n utilities/components >= 85% line coverage.
- Critical paths that must be green
  - Locale switch, persistence, fallback, and error localization mapping.
- Regression checks
  - Existing view rendering and API interaction tests remain passing.

## Risks & Gaps
**What is not fully covered yet?**

- Deferred tests
  - Full end-to-end bilingual UI smoke in real browser.
- Known blind spots
  - Copy quality/terminology correctness requires human review.
- Follow-up actions
  - Add route-level translation completeness check script.
