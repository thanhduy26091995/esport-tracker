---
phase: design
title: System Design & Architecture
description: Define the technical architecture, components, and data models
feature: multi-language-support
status: draft
created: 2026-04-17
---

# System Design & Architecture

## Architecture Overview
**What is the high-level system structure?**

```mermaid
graph TD
  User -->|toggle locale| LanguageSwitcher
  LanguageSwitcher --> I18nPlugin[Vue I18n Plugin]
  I18nPlugin --> LocaleStore[Locale State + Persistence]
  I18nPlugin --> MessagesVI[locales/vi.json]
  I18nPlugin --> MessagesEN[locales/en.json]
  Views[Vue Views + Components] -->|t('key')| I18nPlugin
  APIErrors[API Error Codes] --> ErrorMapper[Localized Error Map]
  ErrorMapper --> I18nPlugin
```

- Key components and responsibilities
  - Vue I18n plugin provides translation engine and fallback behavior.
  - Pinia locale store manages current locale and persists to localStorage.
  - Language switcher component updates locale globally.
  - Locales folder stores translation dictionaries (`vi`, `en`).
  - Error mapper translates backend error codes into localized UI strings.
- Technology stack choices and rationale
  - Use `vue-i18n` (industry standard for Vue 3, TS-friendly, robust fallback and pluralization).

## Data Models
**What data do we need to manage?**

- Core entities and their relationships
  - `LocaleCode`: `'vi' | 'en'`
  - `TranslationMessages`: nested key-value dictionary per locale.
  - `LocalePreference`: persisted setting for active locale.
- Data schemas/structures

```ts
export type LocaleCode = 'vi' | 'en'

export interface I18nState {
  currentLocale: LocaleCode
  fallbackLocale: LocaleCode
}

export type TranslationMessages = Record<string, string | TranslationMessages>
```

- Data flow between components
  - On app startup: read persisted locale -> validate -> set active locale.
  - On switch: update Pinia locale state + i18n global locale + persist preference + rerender bound text.

## API Design
**How do components communicate?**

- External APIs (if applicable)
  - None (frontend-only feature).
- Internal interfaces
  - `useLocaleStore(): { locale, setLocale, getLocale }`
  - `setLocale(locale: LocaleCode): void`
  - `getLocale(): LocaleCode`
  - `t(key: string, params?: Record<string, unknown>): string`
- Request/response formats
  - N/A for remote API; local dictionaries use JSON files.
- Authentication/authorization approach
  - N/A.

## Component Breakdown
**What are the major building blocks?**

- Frontend components (if applicable)
  - `components/common/LanguageSwitcher.vue`
  - `stores/localeStore.ts`
  - Updates across all text-bearing components to use i18n keys.
- Backend services/modules
  - None.
- Database/storage layer
  - Browser localStorage key: `app.locale`.
- Third-party integrations
  - `vue-i18n` package.

## Design Decisions
**Why did we choose this approach?**

- Key architectural decisions and trade-offs
  - Centralized i18n plugin to prevent fragmented translation logic.
  - Fallback locale `vi` to protect current user base.
  - Keep `vi` and `en` bundled in this phase to minimize implementation complexity.
  - Stable key naming (`domain.section.label`) for maintainability.
  - Enforce domain-prefixed keys with a translation check script.
- Alternatives considered
  - Custom dictionary helper (rejected for scalability/tooling concerns).
  - Per-component translation constants (rejected due to duplication risk).
  - Lazy-loaded locale chunks in this phase (rejected for now to reduce rollout risk).
- Patterns and principles applied
  - Single source of truth for locale state.
  - Progressive enhancement: localize core pages first, then remaining surfaces.

## Non-Functional Requirements
**How should the system perform?**

- Performance targets
  - Locale switching should be near-instant and not trigger full app reload.
- Scalability considerations
  - Structure locale files to allow adding new language with minimal refactor.
- Security requirements
  - Treat translation strings as static trusted assets; no user-generated HTML interpolation.
- Reliability/availability needs
  - Missing keys should fall back to `vi` and log warning in development.

## Design Review Decisions

- Locale state ownership: Pinia store is the single state owner for locale.
- Locale loading strategy: bundle `vi` and `en` in this phase.
- Error localization contract: map by backend error code only, with raw message fallback handled in UI.
- Translation governance: require domain-prefixed keys and run translation consistency checks before merge.
