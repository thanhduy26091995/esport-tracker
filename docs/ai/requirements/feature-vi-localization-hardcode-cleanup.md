---
phase: requirements
title: Requirements & Problem Understanding
description: Clarify the problem space, gather requirements, and define success criteria
---

# Requirements & Problem Understanding

## Problem Statement
**What problem are we solving?**

- The Vietnamese locale file is only partially complete and some wording is still literal or inconsistent with natural Vietnamese UI copy.
- Multiple frontend views still render user-facing text as hardcoded English strings instead of using `vue-i18n` keys.
- The result is a mixed-language experience, increased regression risk when copy changes, and duplicated translation effort across components.
- Primary users affected are Vietnamese-speaking players and organizers using the FC25 tracker on the frontend.
- Current workaround is manual copy edits inside Vue files, which bypasses locale conventions documented in `frontend/src/locales/README.md`.

## Goals & Objectives
**What do we want to achieve?**

- Review all entries in `frontend/src/locales/vi.json` and rewrite any awkward or overly literal copy into natural, product-appropriate Vietnamese.
- Ensure every user-facing string in all routed frontend views is driven by locale keys instead of hardcoded text.
- Add any missing keys to both `frontend/src/locales/vi.json` and `frontend/src/locales/en.json` in the same change.
- Keep existing translation keys stable where possible to avoid unnecessary churn.
- Establish a documented, repeatable audit command for detecting future hardcoded strings during review.

- Non-goals:
- Rewriting backend error payloads or API contracts.
- Redesigning tournament UX or changing business rules.
- Adding CI enforcement or a custom lint rule in this iteration.

## User Stories & Use Cases
**How will users interact with the solution?**

- As a Vietnamese-speaking user, I want all visible UI labels, empty states, confirmations, and buttons to read naturally in Vietnamese so the app feels coherent and trustworthy.
- As a product maintainer, I want all visible copy to live in locale files so I can update wording without scanning Vue templates manually.
- As a developer, I want missing translations and hardcoded strings to be easy to detect so future features do not regress localization quality.
- As a reviewer or QA contributor, I want a clear scoped audit target across all routed views so I can verify localization coverage objectively before release.

- Key workflows:
- Open tournament listing, detail, and creation screens without seeing English labels such as "Create Tournament", "Standings", or "Delete".
- Use dialogs and confirmations with translated, context-aware messaging.
- Navigate all routed views, including placeholder pages that are still reachable, without encountering hardcoded user-facing English strings.
- Review locale files without encountering inconsistent terminology across dashboard, players, matches, settlements, fund, config, and tournament modules.

- Edge cases:
- Tournament-specific enums such as `active`, `completed`, `1v1`, and `2v2` may need display mapping instead of exposing raw values.
- Dynamic messages with interpolation must remain grammatically natural in Vietnamese.
- Routed placeholder screens such as `SettingsView.vue` and `CreateMatchView.vue` must still be localized if they remain reachable from navigation or router configuration.

## Success Criteria
**How will we know when we're done?**

- `vi.json` is reviewed end-to-end and all revised strings read naturally for Vietnamese UI usage.
- No visible hardcoded user-facing strings remain in routed frontend views that are accessible in the app.
- Every new or revised translation key exists in both `vi.json` and `en.json`.
- Development mode shows no missing-key warnings for the audited routes.
- Manual walkthrough of tournaments, matches, settlements, fund, config, and players screens shows consistent terminology.

- Acceptance criteria:
- Tournament screens no longer contain hardcoded English copy.
- Routed placeholder screens no longer contain hardcoded visible copy if they remain reachable.
- Confirmation dialogs, table headers, buttons, empty states, and badges use `t(...)` or centralized translation helpers.
- Locale naming follows the existing domain-prefix convention from `frontend/src/locales/README.md`.
- Raw backend enum-like values are mapped to localized display labels through shared helpers or reusable computed mappings where repeated.

## Constraints & Assumptions
**What limitations do we need to work within?**

- The frontend uses Vue 3, TypeScript, Element Plus, and `vue-i18n` with `vi` as both default and fallback locale.
- Existing locale keys should be preserved where possible; update values first, add keys only when coverage is missing.
- Frontend currently has both completed flows and placeholder pages, and all routed views are in scope for this cleanup.
- This feature is being prepared to implementation readiness, so documents should define scope and sequencing clearly before code execution.

- Assumptions:
- Vietnamese is the primary runtime language and should be polished first.
- English remains the secondary locale and should maintain key parity and functional correctness.
- Vietnamese copy should use a neutral, product-style tone rather than informal club slang.
- Tournament-related screens are the highest-risk gap because they currently contain the most visible hardcoded strings, but the implementation scope covers all routed views.
- A documented grep-style audit command is sufficient as the guardrail for this iteration; CI automation can be handled later if still needed.

## Questions & Open Items
**What do we still need to clarify?**

- Are there any routed views outside the currently identified set that should be explicitly excluded from this cleanup for release timing reasons?
- Should the documented audit command be added only to feature documentation or also to frontend contributor docs after implementation?

- Approaches considered:
- Approach 1: Full frontend localization sweep in one pass. Trade-off: maximum coverage but higher regression risk and slower review.
- Approach 2: Tournament-first cleanup only. Trade-off: fastest visible improvement but leaves known hardcoded strings elsewhere.
- Approach 3: Staged audit with tournament-first implementation plus follow-up sweep and guardrails. Trade-off: slightly more planning effort, but best balance of risk, coverage, and maintainability.

- Recommendation:
- Use a hybrid of Approach 1 and Approach 3: keep tournament flows as the first execution priority, but complete the cleanup across all routed views in the same feature scope, and document a repeatable audit command for future reviews.