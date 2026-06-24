---
phase: planning
title: Project Planning & Task Breakdown
description: Break down work into actionable tasks and estimate timeline
---

# Project Planning & Task Breakdown — WC Live Chat

## Milestones

- [x] M1: Backend data layer (model, repo, migration)
- [x] M2: WebSocket chat infrastructure (types, hub, handler, service)
- [x] M3: REST history endpoint + router wiring
- [x] M4: Frontend state & WS composable
- [x] M5: Frontend UI (chat panel, floating button)
- [x] M6: Notification overlap fix
- [x] M7: Tests, i18n, nginx, polish

---

## Task Breakdown

### M1 — Backend Data Layer

- [x] **T1.1** Create `backend/internal/model/wc_chat_message.go` with `WcChatMessage` GORM model (id UUID, user_id, user_name, avatar_url, message, created_at).
  - _Validates:_ unit test T-repo-1, T-repo-2

- [x] **T1.2** Create `backend/internal/repository/wc_chat_repository.go` with `SaveMessage(msg *WcChatMessage) error` and `ListLast100() ([]WcChatMessage, error)`.
  - _Validates:_ unit tests T-repo-1 through T-repo-3

- [x] **T1.3** Add `WcChatMessage` to `AutoMigrate` call in `backend/internal/database/database.go`.
  - _Depends on:_ T1.1

---

### M2 — WebSocket Chat Infrastructure

- [x] **T2.1** Extend `backend/internal/ws/types.go`: add `ChatSendFrame` (inbound) and `ChatMessageEvent` (outbound) structs.
  - _Validates:_ design D6

- [x] **T2.2** Add `BroadcastChat(event ChatMessageEvent)` method to `backend/internal/ws/hub.go` (serialise to JSON → push to `h.broadcast` channel).
  - _Depends on:_ T2.1
  - _Validates:_ unit test T-hub-broadcast

- [x] **T2.3** Create `backend/internal/service/wc_chat_service.go` — `WcChatService` with `SendMessage(userID uuid.UUID, userName, avatarURL, text string) error`: validate length (1–500 chars), save via repo, broadcast via `ChatHubBroadcaster` interface.
  - _Depends on:_ T1.2, T2.2
  - _Validates:_ unit tests T-svc-1 through T-svc-4

- [x] **T2.4** Create `backend/internal/ws/chat_handler.go` — `ChatHandler` Gin handler: upgrades HTTP→WS, parses `?token=<JWT>` query param (validates via `WcAuthService.VerifyToken`), attaches user identity to `client` struct, reads inbound `chat_send` frames and calls `WcChatService.SendMessage`; guest clients (no token) reject send with `{"type":"error","message":"unauthenticated"}`.
  - _Depends on:_ T2.3
  - _Validates:_ unit tests T-handler-1 through T-handler-3, integration test T-int-1, T-int-2

- [x] **T2.5** Extend `client` struct in `backend/internal/ws/` to carry `userID`, `userName`, `avatarURL` fields for authenticated chat clients.
  - _Depends on:_ T2.4

---

### M3 — REST History Endpoint + Router Wiring

- [x] **T3.1** Add `GET /wc/chat/messages` handler (in `wc_handler.go` or a new `wc_chat_handler.go`): calls `repo.ListLast100()`, returns `{ messages: [] }` JSON; no auth required.
  - _Depends on:_ T1.2
  - _Validates:_ integration tests T-int-3, T-int-4

- [x] **T3.2** Wire everything in `backend/internal/api/router.go`: create `chatHub` (`ws.NewHub()`), start `go chatHub.Run()`, create `wcChatRepo`, `wcChatService`, `chatHandler`; register `GET /ws/chat` and `GET /wc/chat/messages`.
  - _Depends on:_ T2.4, T3.1
  - _Validates:_ integration test T-int-5 (end-to-end WS flow)

---

### M4 — Frontend State & WS Composable

- [x] **T4.1** Create `frontend/src/types/chat.ts`: `ChatMessage`, `ChatSendFrame`, `ChatMessageEvent` TypeScript interfaces.

- [x] **T4.2** Create `frontend/src/stores/chatStore.ts` (Pinia): state `messages: ChatMessage[]`, `unreadCount: number`, `isPanelOpen: boolean`; actions `openPanel()` (resets unread), `closePanel()`, `appendMessage(msg)` (increments unread when panel closed).
  - _Depends on:_ T4.1

- [x] **T4.3** Create `frontend/src/composables/useChatWs.ts`: connects to `${VITE_WS_BASE_URL}/ws/chat?token=<JWT>` (no token for guests), handles inbound `chat_message` frames → calls `chatStore.appendMessage`, sends `chat_send` frames; auto-reconnects on close (3 s delay); exposes `sendMessage(text: string)` and `isConnected: Ref<boolean>`.
  - _Depends on:_ T4.1, T4.2

---

### M5 — Frontend UI

- [x] **T5.1** Create `frontend/src/components/wc/WcChatPanel.vue`: message list (scrolled to bottom on new message), each row shows avatar, user name, timestamp, message text; input box (maxlength=500) + send button; disabled input with i18n "login to chat" prompt for guests; calls `useChatWs.sendMessage` on submit; fetches history via `GET /wc/chat/messages` on `onMounted`.
  - _Depends on:_ T4.2, T4.3

- [x] **T5.2** Create `frontend/src/components/wc/WcChatButton.vue`: fixed-position floating button (bottom-right, above bet activity area); shows unread badge (`chatStore.unreadCount`) when > 0; toggles `chatStore.openPanel()` / `chatStore.closePanel()`.
  - _Depends on:_ T4.2

- [x] **T5.3** Mount `<WcChatButton>` and `<WcChatPanel>` in the WC layout component (whichever layout wraps all WC pages — `WcLayout.vue` or `MainLayout.vue`).
  - _Depends on:_ T5.1, T5.2

---

### M6 — Notification Overlap Fix

- [x] **T6.1** Modify `frontend/src/composables/useActivityFeed.ts`: import `storeToRefs` + `useChatStore`; read `isPanelOpen`; pass `position: isPanelOpen.value ? 'top-right' : 'bottom-right'` to each `ElNotification` call.
  - _Depends on:_ T4.2
  - _Validates:_ manual test "notification position switches when panel is open/closed"

---

### M7 — Tests, i18n, Nginx, Polish

- [x] **T7.1** Write unit tests for `WcChatService` (T-svc-1 through T-svc-4) and `WcChatMessageRepository` (T-repo-1 through T-repo-3).

- [x] **T7.2** Write integration test: authenticated WS client sends frame → DB row created → broadcast received by second client.

- [x] **T7.3** Write integration test: `GET /wc/chat/messages` returns ≤ 100 rows ordered oldest → newest.

- [x] **T7.4** Add i18n keys to `frontend/src/locales/vi.json` and `en.json` for all chat UI strings (placeholder, send button, login prompt, unread label, error message).
  - _Depends on:_ T5.1, T5.2

- [x] **T7.5** Add `/ws/chat` Nginx proxy block to deployment docs (`docs/ai/deployment/2026-06-24-feature-wc-live-chat.md`) and VPS config.

- [x] **T7.6** Run manual testing checklist from `docs/ai/testing/2026-06-24-feature-wc-live-chat.md`.

---

## Dependencies

```
T1.1 → T1.3
T1.2 → T2.3, T3.1
T2.1 → T2.2 → T2.3 → T2.4 → T2.5
T2.4 → T3.2
T3.1 → T3.2
T4.1 → T4.2 → T4.3
T4.2 → T5.1, T5.2, T6.1
T5.1 → T5.3
T5.2 → T5.3
T5.1, T5.2 → T7.4
```

## Sequencing Notes

- Start with M1 + M2.1–T2.2 in parallel (no frontend dependency).
- M4 (frontend state) can start as soon as the API contract is agreed — no backend must be running.
- T6.1 (notification fix) is a 5-line change; do it immediately after T4.2 lands.
- T7.5 (Nginx) must land before production deployment.

## Risks & Mitigation

| Risk | Likelihood | Mitigation |
|------|-----------|------------|
| JWT in query param logged in Nginx access logs | Low | App is internal/friend group; no sensitive escalation. Acceptable. |
| Race: REST history fetch + WS open → duplicate messages | Medium | Deduplicate by `id` in `chatStore.appendMessage` |
| `wc_users.google_picture` column missing | Low | Verify column name before T2.4; design doc assumes it exists from WC auth system |
| Unread badge not resetting | Low | `openPanel()` action explicitly zeroes `unreadCount`; covered by manual test |

## Resources Needed

- Existing `WsHub` infrastructure (`backend/internal/ws/`)
- Existing `WcJWTMiddleware` / `WcAuthService.VerifyToken` for token validation in `ChatHandler`
- Element Plus `ElNotification` for notification position switch
- `gorilla/websocket` already in `go.mod`
