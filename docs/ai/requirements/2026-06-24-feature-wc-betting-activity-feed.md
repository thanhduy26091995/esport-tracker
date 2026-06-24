---
phase: requirements
title: Requirements & Problem Understanding
description: Clarify the problem space, gather requirements, and define success criteria
---

# Requirements & Problem Understanding — WC Betting Activity Feed

## Problem Statement

When users visit the WC betting pages, there is no social signal that others are actively placing bets. The site feels static — a user has no reason to stay, refresh, or feel urgency. The friend group needs a sense of "the game is live and people are betting now" to increase engagement and return visits.

**Current situation:** No real-time feedback. A user places a bet and sees nothing about what others are doing.

**Affected users:** All WC users (bettors), especially casual ones who need social pull to stay engaged.

## Goals & Objectives

**Primary goals:**
- Show live toast notifications when any user places a bet (e.g., "Ric Phan vừa đặt kèo Tài Xỉu, theo Bồ Đào Nha với 5 ly")
- Make the site feel alive and social in real time
- Build WebSocket infrastructure extensible to future features (live chat, settlement alerts)

**Secondary goals:**
- Display an activity feed list (recent bets) accessible on betting page
- Lay groundwork for live chat as a future extension

**Non-goals:**
- Live chat (future feature — infrastructure will support it, UI not in this scope)
- Notifications for bet cancellations (low engagement value)
- Push notifications / mobile notifications (out of scope)
- Admin-only events (settlement, void) broadcast to users
- Persistent notification history in DB

## User Stories & Use Cases

- As a **bettor**, I want to see a toast pop up when someone else places a bet, so I feel FOMO and stay on the page longer.
- As a **bettor**, I want to know what others are betting on a match, so I can gauge sentiment.
- As a **visitor** (not yet logged in), I can still see activity toasts (they're on the WC schedule page which is public).
- As a **user on any WC page**, toasts appear automatically — I don't need to do anything.

**Key workflow:**
1. UserA places a bet via `POST /wc/matches/:id/bet`
2. Within ~1 second, all other connected browsers see a toast: "UserA vừa đặt kèo [type], theo [selection] với [stake] ly"
3. Toast auto-dismisses after 5 seconds

## Success Criteria

- [ ] Bet placed by any user broadcasts to all connected WC clients within 2 seconds
- [ ] Toast notification displays correct player name, bet type label (in Vietnamese), selection, and stake
- [ ] Toast displays and dismisses cleanly on WcBettingView and WcScheduleView
- [ ] Bettor who placed the bet does NOT see their own toast (suppress self-notification)
- [ ] Page works normally when WebSocket is disconnected (graceful degradation — no error, just no toasts)
- [ ] WebSocket reconnects automatically after disconnect
- [ ] Nginx on VPS correctly proxies the `/ws` endpoint

## Constraints & Assumptions

**Technical constraints:**
- Backend: Go/Gin — no existing WebSocket infrastructure
- Frontend: Vue 3 + Element Plus — `ElNotification` available for toast
- VPS hosting with Nginx reverse proxy — requires 5-line Nginx config change
- No Redis or message broker — in-process hub only (acceptable for < 50 concurrent users)

**Assumptions:**
- Friend group size stays small (< 50 concurrent users) — no need for distributed hub
- Full name and exact stake are shown — no privacy masking (friend group context)
- Only `bet_placed` events are broadcast in this scope (not cancel, settle, void)
- The hub lives in-process in the Go binary — a server restart clears connected clients (acceptable)

## Questions & Open Items

- **Resolved:** Transport = WebSocket (chosen over SSE/polling to support future live chat)
- **Resolved:** Events scope = `bet_placed` only
- **Resolved:** Privacy = full name + exact stake (friend group)
- **Deferred:** Live chat UI (infrastructure built, chat message type deferred)
- **Deferred:** Settlement broadcast notification
