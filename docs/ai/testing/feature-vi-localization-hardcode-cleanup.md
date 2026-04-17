---
phase: testing
title: Testing Strategy & Validation
description: Define test scope, coverage, and validation criteria for the feature
---

# Testing Strategy

## Scope
**What should be tested for this feature?**

- Unit test scope:
- Any new helper that maps tournament enum/status values to localized labels.

- Integration test scope:
- Tournament list, create, and detail screens rendering translated strings correctly.
- Dialogs, empty states, and table headers resolving locale keys without missing-key warnings.

- End-to-end test scope:
- Manual smoke walkthrough of routed frontend pages in Vietnamese and English.

## Test Files
**Which files contain tests?**

| File | Package/Layer | Coverage Target |
|------|---------------|----------------|
| frontend/src/views/TournamentsView.vue | View | No hardcoded visible strings remain |
| frontend/src/views/TournamentDetailView.vue | View | All controls and labels are localized |
| frontend/src/views/CreateTournamentView.vue | View | Form labels, hints, and buttons are localized |
| frontend/src/locales/vi.json | Locale | Natural Vietnamese phrasing and key completeness |
| frontend/src/locales/en.json | Locale | Matching key coverage with `vi.json` |

## Unit Tests
**What logic should be validated in isolation?**

- Display mapping for tournament status labels
- Display mapping for match type labels if introduced
- Display mapping for score-affect badges if introduced

## Integration Tests
**What cross-component behaviors should be validated?**

- Tournament pages render using translated keys and not raw English strings
- `ElMessageBox` confirmations use translated title and action buttons
- New locale keys interpolate values correctly for counts, names, and amounts
- Locale fallback behavior does not hide missing key mismatches during development

## Test Data & Environments
**What data and setup are required?**

- Seed data with at least one active tournament, one completed tournament, and multiple participants
- Development environment with Vite frontend running and browser console visible
- Ability to switch locale between `vi` and `en`

## Execution
**How do we run and verify tests?**

- Run the frontend locally and navigate tournament routes manually
- Use grep audit commands to identify hardcoded visible strings under `frontend/src`
- Verify browser console shows no missing-key warnings for audited screens

- Expected pass criteria:
- No visible hardcoded English strings remain in scoped frontend files
- Both locale JSON files contain matching keys for the implemented feature scope
- Vietnamese copy reads naturally in context

## Coverage & Quality Gates
**What quality bar must be met?**

- 100% key parity for the touched locale namespaces
- 0 missing-key warnings for tournament routes in development
- 0 untranslated user-facing strings in audited active views

## Risks & Gaps
**What is not fully covered yet?**

- Placeholder screens outside the chosen scope may still contain hardcoded copy
- Grep-based checks may produce false positives for technical literals such as route paths or enum values
- Natural-language quality still benefits from human copy review after technical localization is complete