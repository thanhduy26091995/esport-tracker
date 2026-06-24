---
phase: requirements
title: Requirements & Problem Understanding
description: Clarify the problem space, gather requirements, and define success criteria
---

# Requirements & Problem Understanding — WC Live Chat

## Problem Statement

The friend group currently has no shared channel inside the platform to discuss matches, bets, and World Cup events in real-time. Social commentary is fragmented across external apps (Zalo, Messenger). A built-in live chat would keep the social engagement inside the app and make the betting experience more lively.

**Who is affected:** All WC users (logged-in players) and casual visitors who want to follow the conversation.

**Current workaround:** External messaging apps; no in-app context when checking odds or bets.

## Goals & Objectives

**Primary goals:**
- Provide a single global chat room visible on all WC pages
- Messages delivered in real-time via WebSocket
- Persist the last 100 messages so new joiners see context

**Secondary goals:**
- Unread badge on the floating button when panel is closed
- Graceful notification coexistence (activity feed toasts must not overlap the open chat panel)

**Non-goals (explicit out of scope):**
- Per-match chat rooms
- Private / direct messaging
- Message editing or deletion by users
- File or image uploads
- Reactions / emoji reactions
- Push notifications (mobile / browser)
- Admin message moderation beyond future scope

## User Stories & Use Cases

- As a **logged-in WC user**, I can open the chat panel, type a message, and send it to the global room so that everyone online sees it immediately.
- As a **guest (not logged in)**, I can open the chat panel and read all messages, but the input box is disabled with a "Log in to chat" prompt.
- As **any visitor**, when I open the chat panel, the last 100 messages load so I have context of the ongoing conversation.
- As **any visitor**, new messages from other users arrive in real-time without a page refresh.
- As a **logged-in user**, when the chat panel is closed, a badge on the floating button shows how many new messages have arrived since I last opened it.
- As **any visitor**, when the chat panel is open, activity feed notifications (ElNotification bottom-right toasts) switch to `top-right` so the chat panel is not covered.

## Success Criteria

- A logged-in user can type and send a message; it appears for all connected clients within 2 seconds.
- On panel open, last 100 messages load from the REST endpoint before the WS stream continues.
- Unread badge increments correctly when the panel is closed and resets to zero on open.
- When the chat panel is open, activity notifications render at `top-right` (no overlap with bottom-right panel).
- Chat continues to work normally if the WS connection drops (auto-reconnect within 3 seconds).
- Sending a message fails gracefully with an error message if the user is unauthenticated or the WS is disconnected.

## Constraints & Assumptions

**Technical:**
- Must reuse `backend/internal/ws/Hub` — the chat hub is a second `Hub` instance alongside the activity hub.
- JWT token passed as query param `?token=<JWT>` on WS upgrade (browser WS API does not support custom headers).
- Single-server VPS deployment, < 50 concurrent users — no Redis pub/sub needed.
- All UI strings go through `vue-i18n` (no hardcoded text in components).
- Message max length: 500 characters (enforced client + server).

**Assumptions:**
- WC auth (Google OAuth + JWT) is already in place; chat re-uses `WcJWTMiddleware` logic.
- `wc_users` table has `avatar_url` or `google_picture` column for display in chat.
- A separate `/ws/chat` endpoint and hub are created; existing `/ws` activity feed is untouched.

## Questions & Open Items

- **Avatar source**: Confirm `wc_users.google_picture` is the field to use for user avatar in chat messages. *(Assumed yes — matches the auth design.)*
- **Message deletion**: No user-facing delete in this scope. Admin delete can be added later.
- **Rate limiting**: No per-user send rate limit in this scope. Acceptable for a private friend group.
