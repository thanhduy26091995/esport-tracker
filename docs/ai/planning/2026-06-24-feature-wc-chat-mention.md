---
phase: planning
title: WC Chat @Mention — Task Plan
description: Ordered task breakdown for implementing the @mention feature in Live Chat.
---

# Task Plan: WC Chat @Mention

## Section A — Backend Foundation

- [ ] A1: Add `WcChatMention` model + AutoMigrate
- [ ] A2: Add `SaveMentions`, `UnreadMentionCount`, `MarkMentionsRead` to `WcChatRepository`
- [ ] A3: Add `SendToUser` method to `ws.Hub`; add `ChatMentionEvent` struct; extend `ChatSendFrame` with `Mentions []string`
- [ ] A4: Update `WcChatService.SendMessage` to accept `mentions []uuid.UUID`; save rows; call `hub.SendToUser`
- [ ] A5: Add `GetWcUsers` API handler + `GET /wc/users` route
- [ ] A6: Add `GetUnreadMentionCount` + `MarkMentionsRead` handlers; wire routes

## Section B — Frontend

- [ ] B1: Extend `ChatSendFrame` type; add `ChatMentionEvent` type in `types/chat.ts`
- [ ] B2: Add `unreadMentionCount`, `fetchUnreadMentions`, `markMentionsRead` to `chatStore`
- [ ] B3: Add `getWcUsers`, `getUnreadMentionCount`, `markMentionsRead` to `wcService`
- [ ] B4: Handle `chat_mention` WS frame in `useChatWs` → `ElNotification` + increment count
- [ ] B5: `WcChatPanel` — add `@` autocomplete, pass `mentions[]` on send, render `@name` highlighted, mark-read on open
- [ ] B6: `WcChatButton` — badge = `unreadCount + unreadMentionCount`

## Dependencies

- A3 before A4 (hub method needed by service)
- A1 + A2 before A4 (repo needed by service)
- A4 before A6 (service signature change propagates to handler)
- B1 before B4, B5 (types needed)
- B2 + B3 before B5 (store + service needed by panel)
