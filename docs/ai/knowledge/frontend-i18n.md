# Frontend i18n & Localization

## Overview

Multi-language support via `vue-i18n`. Two locales: Vietnamese (`vi`, primary) and English (`en`). Locale preference persisted in localStorage.

## Setup

- Plugin: `vue-i18n` registered in `src/plugins/`
- Translation files: `src/locales/vi.json`, `src/locales/en.json`
- Pinia store: `localeStore` — manages `currentLocale` state and localStorage persistence
- Fallback locale: `vi` (always falls back to Vietnamese for missing keys)

## Key Conventions

- **Domain-prefixed keys**: `domain.section.label` — e.g., `nav.dashboard`, `match.type.1v2`, `wc.bet.handicap`
- **Missing-key warnings**: logged in dev mode — do not leave missing keys in production
- **Backend enum values** → locale keys via a helper function. Never display raw enum strings to users.
- **No hardcoded strings** in Vue components — always use `t('key')` from `useI18n()`

## Locale Toggle

UI element (usually in settings or header) switches between `vi` and `en` by calling `localeStore.setLocale()`.

## VI Localization Hardcode Cleanup

A dedicated cleanup pass replaced hardcoded English strings across all components. Locale dictionaries are grouped by domain:

```
common.*        → shared labels (save, cancel, loading...)
nav.*           → sidebar/header navigation
dashboard.*     → dashboard page
match.*         → match recording and history
tournament.*    → tournament views
wc.*            → World Cup feature strings
admin.*         → admin panel
```

## Adding New Strings

1. Add key to both `vi.json` and `en.json` under the appropriate domain prefix
2. Use `const { t } = useI18n()` in `<script setup>` and `{{ t('key') }}` in template
3. For dynamic values: `t('key', { param: value })`
