---
phase: requirements
title: WC Chat @Mention
description: Tag a user in live chat; they receive a notification and their name is highlighted in the message bubble.
---

# Requirements: WC Chat @Mention

## Problem Statement

Users in the Live Chat want to direct a message at a specific person. Currently all messages are broadcast to everyone with no way to signal who a message is intended for, and no one is notified when they're specifically addressed.

## Goals & Objectives

**Primary goals:**
1. Type `@` in the chat input → autocomplete dropdown of all WC users.
2. Mentioned user receives a real-time WS notification if online.
3. Unread mention count persisted so badge shows even after page reload.
4. `@username` highlighted (accent color) in all chat message bubbles.

**Secondary goals:**
5. Unread mention count folded into existing FAB unread badge.
6. All unread mentions cleared when user opens the chat panel.

**Non-goals:**
- Mobile / email push for offline users.
- Scroll-to-message when clicking a mention notification.
- Per-message read receipt or separate mention inbox.
- Online presence indicator in autocomplete.

## User Stories & Use Cases

1. **Sender** — I type `@` in the input, see a dropdown of all WC users filtered by what I type, select one, and the name is inserted. I send the message normally.
2. **Mentioned user (online)** — I receive an `ElNotification` popup "Alice nhắc đến bạn: Hey @Alice check this!" without needing to open the chat panel.
3. **Mentioned user (offline / reconnecting)** — When I open the site later I see an elevated number on the chat FAB. Opening the panel marks all mentions as read.
4. **Any user reading chat** — `@username` text is visually distinct (highlight color) inside every message bubble.

**Edge cases:**
- Sender tags themselves → still saved, no WS ping to self.
- Multiple `@` in one message → all valid user IDs get mention records.
- User with no WS connection at send time → mention saved, no WS send.
- Non-existent user ID in mentions list → silently skipped.

## Success Criteria

- [ ] Autocomplete dropdown appears on `@` keystroke, filtering in real time.
- [ ] Message saved to `wc_chat_messages`; mention rows saved to `wc_chat_mentions`.
- [ ] Online mentioned user receives `{type: "chat_mention"}` WS frame → `ElNotification` shown.
- [ ] Unread mention count survives page refresh (persisted in DB).
- [ ] Opening chat panel calls mark-read API; badge returns to 0.
- [ ] `@username` substring rendered with highlight color in all chat bubbles.
- [ ] Multi-mention (multiple `@` in one message) works correctly.

## Constraints & Assumptions

- Reuse existing `chatHub` (`/ws/chat`); no new WS endpoint needed.
- Hub needs `SendToUser(userID uuid.UUID, data []byte)` — safe with current `sync.RWMutex` design.
- Autocomplete source: `GET /api/v1/wc/users` (all WC registered users, fetched once on panel open).
- Client sends `mentions: ["uuid1", "uuid2"]` alongside message text; backend validates user existence.
- `@username` in text uses display name. Duplicate names are acceptable at friend-group scale.
- Mark-read trigger: when `isPanelOpen` becomes `true`.
- Badge: mention unread count added to existing `chatStore.unreadCount`.

## Questions & Open Items

All resolved during requirements session.

| # | Question | Answer |
|---|----------|--------|
| 1 | Offline notification? | Persistent unread badge (DB-backed). |
| 2 | Autocomplete user pool | All WC users via REST. |
| 3 | Badge UI | Merged into existing FAB unread count. |
| 4 | Mark-read trigger | On chat panel open. |
| 5 | Multi-mention per message? | Yes — multiple `@` allowed. |
