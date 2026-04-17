---
phase: requirements
title: Requirements & Problem Understanding
description: Clarify the problem space, gather requirements, and define success criteria
feature: multi-language-support
status: draft
created: 2026-04-17
---

# Requirements & Problem Understanding

## Problem Statement
**What problem are we solving?**

- The web UI is currently single-language and creates friction for mixed-language users.
- Primary users are FC25 esport players and organizers in Vietnam; some users prefer English UI labels while others need Vietnamese.
- Current workaround is informal translation/explaining by other users, which is slow and error-prone.

## Goals & Objectives
**What do we want to achieve?**

- Primary goals
  - Add bilingual UI support for Vietnamese (`vi`) and English (`en`) in frontend.
  - Allow users to switch language without reloading the page.
  - Persist language choice across sessions.
- Secondary goals
  - Keep translation keys maintainable and scalable for additional locales in future.
  - Ensure all current core views (dashboard, users, matches, settlements, fund, config) are localized.
- Non-goals (what's explicitly out of scope)
  - Backend API localization.
  - Multi-currency conversion logic changes.
  - Adding third language in this phase.

## User Stories & Use Cases
**How will users interact with the solution?**

- As a Vietnamese-speaking player, I want to view all labels/messages in Vietnamese so that I can use the app quickly.
- As an English-speaking user, I want to switch to English so that I can understand actions and data.
- As an organizer, I want language preference remembered so that I do not need to re-select it every time.
- As a developer, I want consistent i18n key conventions so that adding new strings is low-risk.
- Edge cases to consider
  - Missing translation key should gracefully fall back to default locale.
  - Dynamic messages with variables (e.g., score, amount, dates) should render correctly in both locales.

## Success Criteria
**How will we know when we're done?**

- Measurable outcomes
  - 100% user-facing static text in scoped views is translated using i18n keys.
  - Locale switch takes effect immediately (no full-page reload).
  - Selected locale persists after refresh and browser restart.
- Acceptance criteria
  - Language switcher is visible and usable on desktop and mobile layouts.
  - Default language is Vietnamese (`vi`) unless user previously selected another locale.
  - API/error messages shown to users are mapped to localized friendly text where feasible.
- Performance benchmarks (if applicable)
  - Locale switch action completes within 200ms on local environment.
  - No material increase in initial bundle causing >10% page-load regression.

## Constraints & Assumptions
**What limitations do we need to work within?**

- Technical constraints
  - Frontend stack remains Vue 3 + TypeScript + Vite.
  - Existing routing/store architecture remains unchanged.
  - No backend schema or endpoint changes.
- Business constraints
  - Deliver in small incremental steps without blocking ongoing frontend feature work.
- Time/budget constraints
  - Implementation should be planning-ready with phased execution in 1-2 focused dev days.
- Assumptions we're making
  - Vietnamese and English content is available from product owner/team.
  - Current UI components can be updated safely to use translation keys.

## Questions & Open Items
**What do we still need to clarify?**

- Resolved decisions
  - Feature name is `multi-language-support`.
  - This phase is text-only localization (no locale-specific date/number/currency formatting changes).
  - Backend errors are localized by error-code mapping to i18n keys, with fallback to raw message.
  - Localization coverage includes all user-visible text: page labels, modal content, toast notifications, and validation messages.
  - Glossary approval owner is the user.
- Remaining open questions
  - None for requirements fundamentals.

## Approach Options (Brainstorm)

- Option A: Full app-wide Vue I18n from start
  - Trade-off: Clean long-term architecture, but larger first rollout.
- Option B: Incremental per-route localization
  - Trade-off: Faster initial delivery, but temporary mixed-key risk.
- Option C: Lightweight custom dictionary utility
  - Trade-off: Quick setup, but weaker ecosystem/tooling and harder scaling.

**Recommendation:** Option A with phased rollout by view. It gives consistent architecture and avoids rework while still allowing incremental delivery.
