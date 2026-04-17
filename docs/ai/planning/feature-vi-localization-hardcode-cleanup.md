---
phase: planning
title: Project Planning & Task Breakdown
description: Break down work into actionable tasks and estimate timeline
---

# Project Planning & Task Breakdown

## Milestones
**What are the major checkpoints?**

- [x] Milestone 1: Complete localization audit and finalize new translation key inventory
- [ ] Milestone 2: Refactor tournament flows and locale files with no missing key warnings
- [ ] Milestone 3: Sweep remaining routed/stub frontend files and add regression guardrails

## Task Breakdown
**What specific work needs to be done?**

### Phase 1: Foundation
- [x] Task 1.1: Review `vi.json` end-to-end and flag awkward, inconsistent, or literal Vietnamese copy
- [x] Task 1.2: Compare `vi.json` and `en.json` coverage for tournament and shared UI domains
- [x] Task 1.3: Run a grep-based audit of visible hardcoded strings across `frontend/src`
- [x] Task 1.4: Define new locale keys needed for tournaments, shared labels, and placeholder screens in scope

### Phase 2: Core Features
- [x] Task 2.1: Refactor `TournamentsView.vue` to use translation keys for all visible copy
- [x] Task 2.2: Refactor `TournamentDetailView.vue` to use translated labels, badges, buttons, dialogs, and standings headers
- [x] Task 2.3: Refactor `CreateTournamentView.vue` to remove hardcoded copy and use natural localized wording
- [x] Task 2.4: Update `vi.json` and `en.json` together with new keys and revised Vietnamese phrasing
- [x] Task 2.5: Add helper mappings for tournament status, match type, and score-affect labels if repeated in multiple places

### Phase 3: Integration & Polish
- [x] Task 3.1: Sweep remaining routed support components such as `MatchForm.vue`, `MatchList.vue`, `RecentMatches.vue`, settlement dialogs, dashboard fallbacks, and config point-unit labels
- [ ] Task 3.2: Run manual route walkthrough in both locales and verify no missing-key warnings (currently blocked pending interactive browser and console verification)
- [x] Task 3.3: Add a documented hardcode-audit command or CI validation recommendation for future regressions

## Dependencies
**What needs to happen in what order?**

- Locale key inventory must be finalized before broad component refactors to avoid repeated renames.
- Tournament views should be updated before secondary placeholder screens because they carry the highest user impact.
- Manual QA depends on both locale file updates and component refactors being complete.
- Any optional CI check depends on agreeing what files and patterns count as valid exceptions.

## Timeline & Estimates
**When will things be done?**

- Phase 1: 0.5 day
- Phase 2: 0.5 to 1 day
- Phase 3: 0.5 day
- Total estimated effort: 1.5 to 2 days including review and copy refinement

- Buffer:
- Add 0.5 day if broader placeholder/stub screens must also be fully localized.

## Risks & Mitigation
**What could go wrong?**

- Risk: Hidden hardcoded strings remain in less-traveled components.
- Mitigation: Use grep audit plus manual navigation of routed pages.

- Risk: Locale keys become fragmented or duplicated.
- Mitigation: Add keys by domain and prefer extending existing namespaces.

- Risk: Vietnamese wording becomes technically correct but unnatural to end users.
- Mitigation: Review revised copy as product text, not literal translation.

- Risk: Raw enum values still leak into badges or tables.
- Mitigation: Centralize display mapping for repeated tournament values.

## Resources Needed
**What do we need to succeed?**

- One frontend engineer familiar with Vue 3 and `vue-i18n`
- Access to the running frontend for manual walkthrough
- Product/copy review input if tone preference needs confirmation
- Existing locale convention document in `frontend/src/locales/README.md`