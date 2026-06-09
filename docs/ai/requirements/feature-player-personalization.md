---
phase: requirements
title: Requirements – Player Personalization & Dynamic Global Theme
description: Avatar upload, favorite club selection, and dynamic website theme driven by rank #1 player's club
---

# Requirements & Problem Understanding

## Problem Statement

Players currently have no visual identity in the app — every user looks the same (name + tier badge only). There is no personalisation and no community dynamic that makes being rank #1 feel special or visible to others.

**Who is affected:** All FC25 esport tracker players.
**Current situation:** No avatar, no club affiliation, no theme variation.

## Goals & Objectives

**Primary goals:**
- Let players upload a profile avatar (static image or GIF).
- Let players select a favourite football club from a curated list.
- Dynamically apply a club-inspired colour theme to the global site when the current rank #1 player has a club set.

**Secondary goals:**
- Increase engagement and friendly competition (players fight for #1 to "own" the site theme).
- Improve visual identity across leaderboards, match history, and recent matches.

**Non-goals:**
- Social follows, direct messaging, or comment features.
- Club logos from licensed APIs (use flat colour palettes only to avoid IP issues).
- Per-user theme customisation (only rank #1 drives the global theme).

## User Stories

- As a **player**, I want to upload a profile picture (including GIF) so I can stand out in the leaderboard.
- As a **player**, I want to pick my favourite football club so my identity is linked to a real team.
- As a **player**, I want the site to display my club's colours when I reach rank #1, as a visible reward.
- As any **viewer**, I want to see each player's avatar next to their name in leaderboards and match results.
- As an **admin**, I want to be able to clear an inappropriate avatar without deleting the user.

## Success Criteria

- [ ] Players can upload an avatar (JPG/PNG/GIF, max 2 MB) via their profile page.
- [ ] Avatar is shown in: leaderboard table, recent matches list, user detail page.
- [ ] Players can select a club from a fixed dropdown list (~20 clubs).
- [ ] When rank #1 player has a club set, the global CSS theme reflects that club's full palette (gradient, glow, primary, accent, background tint) with a 0.5s fade transition.
- [ ] When rank #1 has no club, site reverts to default green theme.
- [ ] Theme updates within 60 seconds of a leaderboard change (no manual refresh required).
- [ ] Anyone can upload or update any player's avatar (no auth guard — private group app).
- [ ] Anyone can reset any player's avatar to default.

## Constraints & Assumptions

- **Storage:** Local filesystem on the server (`./uploads/avatars/`). Served via Gin static route.
- **File size limit:** 2 MB per upload, enforced server-side.
- **Accepted formats:** JPG, PNG, GIF, WebP. SVG explicitly rejected (XSS risk).
- **GIF animation:** Animated everywhere (leaderboard, match history, profile). No thumbnail generation needed — group is small (<20 players), performance not a concern.
- **Crop/resize:** None. Display uses `object-fit: cover` in the avatar component. Accept and store as-is.
- **Club list:** Fixed set of ~20 popular clubs, stored in frontend config (not a DB table). No dynamic club management needed.
- **Theme:** CSS custom properties (7 vars) injected into `<html>` via `style` attribute. 0.5s fade transition on change.
- **Theme fallback:** If rank #1 has no club set (or club = `none`), revert to default green theme. No stale state.
- **Auth — avatar upload:** No auth guard. The app is a private friend group. Any client can upload or update any player's avatar. No ownership check required.
- **Auth — club selection:** Same as above — no auth guard.
- **Rank #1 definition:** Highest `current_score` in the `users` table (active users only). On tie: first by `created_at` asc (oldest player).

## Questions & Open Items

All questions resolved. No open items remaining.
