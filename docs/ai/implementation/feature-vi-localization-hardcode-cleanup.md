---
phase: implementation
title: Implementation Guide
description: Technical implementation notes, patterns, and code guidelines
---

# Implementation Guide

## Development Setup
**How do we get started?**

- Work from the existing frontend workspace under `frontend/`.
- Ensure dependencies are installed and the Vite app can run locally.
- Use development mode so `vue-i18n` missing-key warnings remain visible.
- Run grep-based audits from the repo root before and after changes to confirm visible hardcodes are removed.

- Recommended audit command:
```sh
grep -RInE 'title="|label="|empty-text="|active-text="|inactive-text="|confirmButtonText:|cancelButtonText:|>[^<{]*[A-Za-z][^<{]*<' frontend/src --include='*.vue' --include='*.ts'
```

- Use the audit command as a candidate finder only. Review results manually to filter out false positives such as translated `t(...)` calls, CSS class names, route paths, and TypeScript identifiers.

## Code Structure
**How is the code organized?**

- Locale dictionaries:
- `frontend/src/locales/vi.json`
- `frontend/src/locales/en.json`

- Localization entry points:
- `frontend/src/plugins/i18n.ts`
- `frontend/src/utils/i18n.ts`

- Primary tournament views:
- `frontend/src/views/TournamentsView.vue`
- `frontend/src/views/TournamentDetailView.vue`
- `frontend/src/views/CreateTournamentView.vue`

- Secondary cleanup targets:
- `frontend/src/views/SettingsView.vue`
- `frontend/src/views/CreateMatchView.vue`
- `frontend/src/components/match/MatchForm.vue`
- `frontend/src/components/match/MatchList.vue`
- `frontend/src/components/match/RecentMatches.vue`

## Implementation Notes
**Key technical details to remember:**

### Core Features
- Feature 1: Rewrite Vietnamese copy in `vi.json` for natural phrasing while preserving stable key names wherever existing coverage is sufficient.
- Feature 2: Add missing tournament and shared UI keys to both locale files in the same change.
- Feature 3: Replace hardcoded labels, empty states, badge text, and confirmation dialog copy with `t(...)` calls.

### Patterns & Best Practices
- Follow `frontend/src/locales/README.md` conventions: domain-prefixed keys, stable key names, and mirrored locale updates.
- Do not hardcode visible UI strings in Vue templates or `ElMessageBox` configuration.
- Prefer computed display helpers when a raw value such as `completed` or `1v1` needs localized rendering in multiple places.
- Keep dynamic interpolation inside translation strings rather than concatenating translated fragments in components.
- If a string is only used once but still user-facing, it still belongs in locale files.

## Integration Points
**How do pieces connect?**

- Views call `useI18n()` and render translated copy with `t(...)`.
- Shared helpers such as `translate(...)` remain valid for toast messages and store-level errors.
- `i18n` fallback behavior remains `vi`, which helps expose missing `en` keys without breaking the default runtime locale.

## Error Handling
**How do we handle failures?**

- Missing locale keys should continue to log warnings in development via `frontend/src/plugins/i18n.ts`.
- Avoid introducing silent fallback strings inside components; missing translations should be fixed at the source locale file.
- Preserve existing store-level error handling and toast behavior unless a user-facing message needs localization refinement.

## Performance Considerations
**How do we keep it fast?**

- Keep translations static in JSON and avoid runtime-generated key paths when simple explicit keys work.
- Use lightweight computed mappings for repeated badge/status labels instead of repeated branching in templates.
- Limit refactors to copy rendering; do not restructure data-loading flows unless needed for localization.

## Security Notes
**What security measures are in place?**

- Continue rendering translated content as plain text.
- Do not introduce `v-html` for localized strings.
- Treat user-provided names in translated interpolations as escaped plain values.