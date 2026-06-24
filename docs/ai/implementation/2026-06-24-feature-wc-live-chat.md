---
phase: implementation
title: Implementation Guide
description: Technical implementation notes, patterns, and code guidelines
---

# Implementation Guide — WC Live Chat

## Files Changed / Added

### Backend
| File | Status | Notes |
|------|--------|-------|
| `backend/internal/model/wc_chat_message.go` | NEW | GORM model; table `wc_chat_messages` |
| `backend/internal/repository/wc_chat_repository.go` | NEW | `Save()` + `ListLast100()` (fetches DESC, reverses in Go) |
| `backend/internal/database/database.go` | MODIFIED | `WcChatMessage` added to `AutoMigrate` |
| `backend/internal/ws/types.go` | MODIFIED | Added `ChatSendFrame`, `ChatMessageEvent`, `ChatErrorFrame`; `client` struct gained `userID`, `userName`, `avatarURL` fields |
| `backend/internal/ws/hub.go` | MODIFIED | Added `BroadcastChat(event ChatMessageEvent)` |
| `backend/internal/ws/chat_handler.go` | NEW | WS upgrade, JWT via `?token=` query param, bidirectional read pump, adapter types `WcAuthTokenVerifier` / `WcUserAvatarFetcher` |
| `backend/internal/service/wc_chat_service.go` | NEW | `SendMessage()` + `ListHistory()`; `ChatHubBroadcaster` interface |
| `backend/internal/api/wc_chat_handler.go` | NEW | `GET /wc/chat/messages` REST handler |
| `backend/internal/api/router.go` | MODIFIED | `wcChatRepo`, `wcChatService`, `chatHub`, adapters wired; routes `GET /ws/chat` + `GET /api/v1/wc/chat/messages` registered |

### Frontend
| File | Status | Notes |
|------|--------|-------|
| `frontend/src/types/chat.ts` | NEW | `ChatMessage`, `ChatSendFrame`, `ChatMessageEvent`, `ChatErrorFrame` interfaces |
| `frontend/src/stores/chatStore.ts` | NEW | Pinia store: `messages`, `unreadCount`, `isPanelOpen`; deduplication in `appendMessage` |
| `frontend/src/composables/useChatWs.ts` | NEW | WS lifecycle with auto-reconnect; exposes `isConnected`, `sendMessage` |
| `frontend/src/components/wc/WcChatPanel.vue` | NEW | Message list, input box, history load on mount; `useChatWs` scoped to panel lifecycle |
| `frontend/src/components/wc/WcChatButton.vue` | NEW | Fixed-position FAB with unread badge |
| `frontend/src/layouts/MainLayout.vue` | MODIFIED | Imports + mounts `WcChatButton` + `WcChatPanel` when `isSocSite || isWcRoute` |
| `frontend/src/composables/useActivityFeed.ts` | MODIFIED | Reads `chatStore.isPanelOpen`; switches notification position to `top-right` when panel open |
| `frontend/src/locales/vi.json` | MODIFIED | Added `wc.chat.*` keys |
| `frontend/src/locales/en.json` | MODIFIED | Added `wc.chat.*` keys |

## Key Decisions Made During Implementation

### Avatar URL in WS handler
`WcClaims` JWT does not carry `avatar_url`. `ChatHandler` does a one-time `wcUserRepo.GetByID()` on WS upgrade to fetch the avatar URL. Cost: one DB query per connection (not per message) — acceptable.

### Adapter pattern for ChatHandler
`ChatHandler` depends on two interfaces (`TokenVerifier`, `UserAvatarFetcher`) wired in `router.go` via `WcAuthTokenVerifier` and `WcUserAvatarFetcher` closure adapters. Keeps `ws` package free of `service`/`repository` imports.

### `useChatWs` scoped to `WcChatPanel`
The WS connection opens when `WcChatPanel` mounts (when the panel is first opened) rather than at app startup. This avoids an idle persistent connection for users who never open chat. Trade-off: ~200ms connect latency on first panel open.

### History loaded per-panel-open (when empty)
`WcChatPanel` loads history only when `messages.length === 0`. Subsequent opens reuse in-memory messages and receive new ones via WS. If the user refreshes, history is re-fetched.

## Nginx Configuration Required (Production)

```nginx
location /ws/chat {
    proxy_pass http://localhost:8080;
    proxy_http_version 1.1;
    proxy_set_header Upgrade $http_upgrade;
    proxy_set_header Connection "upgrade";
    proxy_read_timeout 86400;
}
```

## Testing Notes

- Backend: `go test ./internal/ws/... ./internal/service/... ./internal/repository/...` — all pass
- Frontend: `npx vue-tsc --noEmit` — no errors
- Manual: see `docs/ai/testing/2026-06-24-feature-wc-live-chat.md`
