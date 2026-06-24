---
phase: planning
title: Project Planning & Task Breakdown
description: Break down work into actionable tasks and estimate timeline
---

# Project Planning & Task Breakdown — WC Betting Activity Feed

## Milestones

- [ ] M1: Backend WebSocket infrastructure (hub + handler + router wiring)
- [ ] M2: PlaceBet broadcast integration (service + event formatting)
- [ ] M3: Frontend composable + toast display
- [ ] M4: Nginx config + production verification

## Task Breakdown

### M1: Backend WebSocket Infrastructure

- [ ] **T1.1** Create `backend/internal/ws/types.go` — define `ActivityEvent`, `Message`, `Client` structs and `HubBroadcaster` interface
  - Outcome: shared types available to hub, handler, and service
  - Validation: `go build ./...` passes

- [ ] **T1.2** Create `backend/internal/ws/hub.go` — implement `WsHub` with `Register()`, `Unregister()`, `Broadcast()`, `Run()` goroutine using `sync.RWMutex`
  - Outcome: hub safely fans out messages to all connected clients
  - Validation: unit tests in `hub_test.go` pass; `go test -race ./internal/ws/...` clean

- [ ] **T1.3** Create `backend/internal/ws/handler.go` — Gin handler that upgrades HTTP→WS (`gorilla/websocket`), registers client with hub, runs read/write goroutines with ping/pong heartbeat
  - Outcome: `/ws` endpoint accepts WS connections and keeps them alive
  - Dependencies: T1.1, T1.2
  - Validation: manual WS connection via browser DevTools shows 101 Switching Protocols

- [ ] **T1.4** Wire hub into `router.go` — create `WsHub`, start `hub.Run()`, register `GET /ws` route
  - Outcome: server exposes `/ws` endpoint on startup
  - Dependencies: T1.3
  - Validation: `go run cmd/server/main.go` starts without error; `/ws` route visible in logs

- [ ] **T1.5** Add `gorilla/websocket` dependency
  - Outcome: `go.mod` and `go.sum` updated
  - Validation: `go mod tidy` clean

### M2: PlaceBet Broadcast Integration

- [ ] **T2.1** Add `HubBroadcaster` interface field to `WcService` struct; update `NewWcService()` constructor to accept it (nil-safe)
  - Outcome: service is injectable with real hub or no-op mock
  - Dependencies: T1.1
  - Validation: existing `wc_integration_test.go` still passes (nil hub = no-op)

- [ ] **T2.2** Build `buildActivityEvent()` helper in `wc_service.go` — formats `selection` string based on bet type (Handicap: team name, O/U: Tài/Xỉu, Exact: "X - Y"), fills all `ActivityEvent` fields
  - Outcome: consistent human-readable event payload
  - Validation: unit tests covering all three bet types pass

- [ ] **T2.3** Call `hub.Broadcast(event)` at the end of successful `PlaceBet()` — after DB write, before return
  - Outcome: every successful bet triggers a WS broadcast
  - Dependencies: T2.1, T2.2
  - Testing scenarios: T2 unit tests — success calls broadcast; failure does not

- [ ] **T2.4** Update `router.go` to inject hub into `NewWcService()`
  - Outcome: production wiring complete
  - Dependencies: T1.4, T2.1
  - Validation: end-to-end: place bet → WS message received in browser console

### M3: Frontend Composable + Toast

- [ ] **T3.1** Create `frontend/src/types/activity.ts` — `ActivityEvent` TypeScript interface (mirrors Go struct)
  - Outcome: typed events throughout frontend
  - Validation: `npm run type-check` passes

- [ ] **T3.2** Create `frontend/src/composables/useActivityFeed.ts` — WS connection to `${VITE_WS_BASE_URL}/ws`, `onmessage` parsing, auto-reconnect (`setTimeout(connect, 3000)` on close), `onMounted`/`onUnmounted` lifecycle
  - Outcome: composable manages full WS lifecycle
  - Dependencies: T3.1
  - Validation: unit test with mock WS; type-check passes

- [ ] **T3.3** Add `showToast(event)` in composable — calls `ElNotification` with formatted Vietnamese message; suppresses own events via `wcAuthStore.user.id` check
  - Outcome: toast appears for other users' bets; own bets silently skipped
  - Dependencies: T3.2
  - Testing scenario: self-suppression manual test

- [ ] **T3.4** Add `VITE_WS_BASE_URL` env var to `.env` / `.env.production` (e.g., `wss://yourdomain.com`)
  - Outcome: WS URL configurable per environment
  - Validation: `npm run type-check` passes; dev WS connects to `ws://localhost:8080/ws`

- [ ] **T3.5** Mount `useActivityFeed()` in `MainLayout.vue` — single instance covers all WC pages
  - Outcome: toasts appear on any WC page
  - Dependencies: T3.3, T3.4
  - Validation: manual test — place bet in tab A, toast appears in tab B

### M4: Nginx + Production

- [ ] **T4.1** Add WebSocket proxy block to Nginx config on VPS (`/ws` location with `Upgrade`, `Connection`, `proxy_read_timeout 86400`)
  - Outcome: WS connections work through reverse proxy in production
  - Validation: browser DevTools shows 101 on production domain; no 502 errors

- [ ] **T4.2** Production smoke test — two browser sessions, place bet, confirm toast end-to-end
  - Outcome: feature verified live
  - Testing scenarios: E2E manual checklist from testing doc

## Dependencies

```
T1.5 → T1.2, T1.3
T1.1 → T1.2 → T1.3 → T1.4
T1.1 → T2.1 → T2.3
T2.2 → T2.3
T1.4 + T2.1 → T2.4
T3.1 → T3.2 → T3.3 → T3.5
T3.4 → T3.5
T2.4 + T3.5 → T4.1 → T4.2
```

## Risks & Mitigation

| Risk | Likelihood | Mitigation |
|------|-----------|-----------|
| Nginx `Upgrade` header not forwarded in current config | Medium | Test on staging first; add `proxy_set_header Connection "upgrade"` explicitly |
| `gorilla/websocket` ping/pong keeps connections alive but Go runtime crashes on server restart clears all clients | Low | Browser auto-reconnects; acceptable for friend group |
| Race condition in hub under concurrent register/broadcast | Low | Use `sync.RWMutex`; run `go test -race` |
| `wcAuthStore.user` is null when WS message arrives (not logged in) | Low | Guard: if user is nil, skip self-suppression check and always show toast |

## Timeline & Estimates

| Milestone | Estimate |
|-----------|---------|
| M1 Backend WS infra | ~3h |
| M2 PlaceBet integration | ~1h |
| M3 Frontend composable | ~2h |
| M4 Nginx + smoke test | ~30min |
| **Total** | **~6.5h** |

## Progress Summary

Not started. All tasks pending implementation.
