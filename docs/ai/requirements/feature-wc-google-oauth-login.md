---
phase: requirements
title: WC Google OAuth Login
description: Mandate Google account linking for existing WC players; replace self-registration with Google-only account creation; add profile management
---

# Requirements & Problem Understanding

## Problem Statement

The current WC (World Cup) login/register system uses only a username + password with **no identity verification** — any player can register with any name and impersonate another. There is no email, no identity proof, and the password reset resets to a predictable pattern (`{name}_@123`).

**Who is affected:** All WC players. Impersonation undermines leaderboard integrity and trust in the betting/prediction system.

**Current workaround:** None — players rely on honour. Admin has no way to verify identity.

## Goals & Objectives

**Primary goals:**
1. **Bind each WC player account to a verified Google identity** — preventing impersonation without removing the existing username/password login for current players.
2. **Remove self-service registration** — new players join only via Google OAuth (account auto-created on first Google login).
3. **Block access for unlinked accounts** — after deployment, existing players who have not linked a Google account are blocked until they complete the linking step.
4. **Provide a profile management page** — players can update their display name and avatar URL.

**Secondary goals:**
- Show user avatar in the WC UI (navbar, leaderboard, profile page, bet/prediction history).
- Clean up the unused password-reset and register endpoints.

**Non-goals:**
- Removing username/password login for existing linked accounts — it stays.
- Google OAuth for the main app (non-WC) users — out of scope.
- Admin accounts are not affected; they remain on username/password login.
- Apple Sign-In, email OTP, or any other OAuth provider.
- Bulk admin migration of accounts — players link themselves.

## User Stories & Use Cases

### New player (no existing account)
> As someone joining WC for the first time, I want to click "Login with Google" and get a WC account immediately using my Google identity, so I don't need to create a username/password.

- Visits `/world-cup/login`.
- Clicks "Đăng nhập với Google" → Google consent.
- Backend sees no account for this Google ID → auto-creates a new WcUser (name pre-filled from Google display name, avatar from Google picture).
- Gets a WC JWT, redirected to `/world-cup/predict`.

### Existing player — linking Google account (mandatory post-deploy)
> As an existing WC player, I want to link my Google account to my WC profile so I can keep all my existing data (balance, predictions, bets) and continue playing.

Steps:
1. Logs in with username/password as usual → receives WC JWT.
2. Backend (or frontend) detects `google_id IS NULL` on the account.
3. **Blocked immediately** — redirected to a "Liên kết tài khoản Google" page before accessing any other WC page.
4. Clicks "Liên kết với Google" → Google consent screen.
5. Backend receives Google ID token, links `google_id` to the authenticated WcUser (must not already be used by another account).
6. Player is now unblocked; redirected to `/world-cup/predict` with all previous data intact.

### Existing player (already linked) — normal login
> As a player who has already linked Google, I can log in with username/password and access the app with no interruption.

- Logs in with username/password → JWT → `google_id IS NOT NULL` → normal access.
- Can also log in via "Login with Google" directly if preferred (Google ID found in DB → issues JWT).

### Old player who never links
> As an existing player who ignores the migration, I will be blocked after every login until I complete the Google linking step.

### Profile management
> As a logged-in WC player, I want to navigate to "My Profile", update my display name and set an avatar URL, so my leaderboard entry looks personalised.

### Admin
> Admin is unaffected. Admin accounts continue to use username/password login, and the `is_admin` flag continues to control access to `/world-cup/admin`. No Google linking required for admin.

## Success Criteria

- [ ] "Đăng nhập với Google" creates a new WC account on first use and issues a JWT.
- [ ] A Google account can be linked to at most one WC player account (DB unique constraint on `google_id`).
- [ ] Existing player logs in with username/password → if `google_id IS NULL` → blocked at `/world-cup/link-google` before any other page.
- [ ] After linking, player's existing balance, predictions, bets, and champion predictions are fully intact.
- [ ] Player can log in via Google (if already linked) in ≤ 3 s (P95).
- [ ] Player can update `name` and `avatar_url` from the profile page; name uniqueness is enforced.
- [ ] Player avatar is displayed in: navbar/header, leaderboard/ranking table, profile page, and bet/prediction history.
- [ ] Self-service `/world-cup/register` page is removed (URL returns 404 or redirects to login).
- [ ] Password-reset endpoint is removed or disabled for players.
- [ ] All existing protected routes still work for linked accounts (JWT format unchanged).
- [ ] Admin username/password login continues to function.

## Constraints & Assumptions

**Technical constraints:**
- Google Client ID: `685662087046-1u7mm9ov2pjbfhhstjib4s02su7sdjt3.apps.googleusercontent.com`
- Backend: Go + Gin + GORM + PostgreSQL (existing stack).
- Frontend: Vue 3 + Pinia + Element Plus (existing stack).
- Google Identity Services (GSI) frontend library used for OAuth popup; backend verifies the signed ID token.
- JWT claims format and `WcJWTMiddleware` remain unchanged — the google-link check is done at the frontend routing layer or as a dedicated middleware check on the `google_id` field, not by changing the token format.

**Business constraints:**
- Only WC players are in scope. Main-app users are unchanged.
- Google linking is self-service; no admin action required per player.
- One Google account ↔ one WC player account (enforced by unique index). A Google account that is already linked cannot claim another WC account.
- The existing player list is considered "fixed" for the season — the removal of the password-based registration page is intentional.

**Assumptions:**
- All WC players have a personal Google account they are willing to use.
- The `name` field remains as the display name. After auto-creation via Google, the player can update it via the profile page.
- Admin accounts (`is_admin = true`) are managed out-of-band and do not require Google linking.
- No "unlink Google" feature is needed — once linked, the binding is permanent (or requires admin DB action to undo).

## Questions & Open Items

All questions resolved:

- ~~Should there be a grace period for old password login?~~ → **Resolved: Immediate block after deploy. No grace period.**
- ~~How do new players join?~~ → **Resolved: Login with Google auto-creates an account.**
- ~~When is Google linking required?~~ → **Resolved: Blocking on first login after deploy if `google_id IS NULL`.**
- ~~What name does a re-linker get?~~ → **Resolved: Keep old WC username. Google name is used only for auto-created new accounts.**
- ~~Where does the avatar appear?~~ → **Resolved: Navbar, leaderboard, profile page, bet/prediction history.**
- ~~Should name changes be rate-limited?~~ → **Resolved: No rate-limit for MVP. Deferred to post-launch.**
- ~~Show linked Google email on profile page?~~ → **Resolved: No — keep it simple for MVP. Email is private.**
- ~~What if Google account is deleted/suspended?~~ → **Resolved: Player can still login with username/password — unaffected.**
