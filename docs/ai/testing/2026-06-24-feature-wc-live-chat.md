---
phase: testing
title: Testing Strategy
description: Define testing approach, test cases, and quality assurance
---

# Testing Strategy — WC Live Chat

## Test Coverage Goals

- Unit test coverage: 100% of new service and repository methods
- Integration tests: WS chat flow + REST history endpoint
- Manual: floating button, unread badge, notification overlap fix

## Unit Tests

### `WcChatService`
- [ ] `SendMessage` with valid input saves to DB and calls `hub.BroadcastChat`
- [ ] `SendMessage` rejects empty message (validation error)
- [ ] `SendMessage` rejects message exceeding 500 characters
- [ ] `SendMessage` with DB error returns error without broadcasting

### `WcChatMessageRepository`
- [ ] `SaveMessage` inserts row and returns populated `WcChatMessage`
- [ ] `ListLast100` returns at most 100 rows ordered oldest → newest
- [ ] `ListLast100` returns empty slice when table is empty

### `ChatHub.BroadcastChat`
- [ ] Serialises `ChatMessageEvent` to JSON and pushes to all registered clients
- [ ] Slow client (full send channel) drops message without blocking other clients

### `ChatHandler` (WS auth parsing)
- [ ] Valid JWT token → client registered with correct userID/userName/avatarURL
- [ ] Missing token → client registered as guest (no userID)
- [ ] Invalid/expired JWT token → WS connection rejected with 401

## Integration Tests

- [ ] `POST`-style WS flow: authenticated client sends `chat_send` frame → message saved in DB → `chat_message` event broadcast to all connected clients
- [ ] Guest client sends `chat_send` frame → server responds with `{"type":"error","message":"unauthenticated"}`, no DB write
- [ ] `GET /wc/chat/messages` returns last 100 messages ordered oldest → newest
- [ ] `GET /wc/chat/messages` returns empty array on clean DB
- [ ] WS reconnect: client disconnects and reconnects → hub registers new connection cleanly, no goroutine leak

## Manual Testing

- [ ] Floating button renders at bottom-right on all WC pages
- [ ] Badge shows correct unread count when panel is closed and new messages arrive
- [ ] Badge resets to 0 when panel is opened
- [ ] Panel opens/closes on button click
- [ ] Last 100 messages load when panel opens for the first time
- [ ] Sending a message appears immediately for the sender and for a second browser tab
- [ ] Guest sees messages in real-time but input is disabled with login prompt
- [ ] Activity feed notifications appear at `top-right` when chat panel is open
- [ ] Activity feed notifications return to `bottom-right` when chat panel is closed
- [ ] Sending very long message (>500 chars) is blocked at the input level
- [ ] Auto-reconnect works: kill/restart backend → frontend reconnects within 3s and chat continues

## Test Data

- Seed 5 `wc_users` with `google_picture` URLs for avatar display testing
- Seed 120 chat messages to verify `ListLast100` caps at 100 and returns correct order

## Coverage Commands

```bash
# Backend
cd backend && go test ./internal/service/... ./internal/repository/... ./internal/ws/...

# Frontend type-check
cd frontend && npm run type-check
```
