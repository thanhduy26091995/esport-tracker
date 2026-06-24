---
phase: testing
title: Testing Strategy
description: Define testing approach, test cases, and quality assurance
---

# Testing Strategy — WC Betting Activity Feed

## Test Coverage Goals

- Unit tests: hub logic, event formatting for all bet/prediction types, MockHub verification
- Integration: nil hub graceful degradation verified via existing `wc_integration_test.go`
- Manual: toast display, reconnection behavior, Nginx WS proxy on production

## Unit Tests

### WsHub (`backend/internal/ws/hub_test.go`)

- [x] `Register` adds client to hub client map
- [x] `Unregister` removes client and closes send channel
- [x] `Broadcast` sends message to all registered clients
- [x] `Broadcast` with zero clients does not panic
- [ ] `Run` goroutine exits cleanly when hub is stopped (deferred — requires context cancellation refactor)

### ActivityEvent formatting — PlaceBetRequest (`backend/internal/service/wc_activity_test.go`)

- [x] Handicap home: `selection` = home team name
- [x] Handicap away: `selection` = away team name
- [x] Over/Under over: `selection` = `"Tài"`
- [x] Over/Under under: `selection` = `"Xỉu"`
- [x] Exact score: `selection` = `"X - Y"` format

### ActivityEvent formatting — SubmitPredictionRequest (`backend/internal/service/wc_activity_test.go`)

- [x] Handicap home prediction: `selection` = home team name
- [x] Handicap away prediction: `selection` = away team name
- [x] Over/Under over prediction: `selection` = `"Tài"`
- [x] Over/Under under prediction: `selection` = `"Xỉu"`
- [x] Exact score prediction: `selection` = `"X - Y"` format

### MockHub + nil hub (`backend/internal/service/wc_activity_test.go`)

- [x] `MockHub.Broadcast` records calls with correct fields
- [x] Nil `HubBroadcaster` interface — guard prevents panic (documents nil-safety contract)

## Integration Tests

- [x] `wc_integration_test.go` — existing suite passes with `hub=nil` (nil hub = no broadcast, no crash)
- [ ] WS client receives `ActivityEvent` after `SubmitPrediction` — requires httptest WS setup (deferred)
- [ ] Two clients: client A bets → client B receives, client A does not (self-suppression) — manual only

## End-to-End Tests (Manual)

- [ ] Open two browser tabs on `/world-cup/predict` with different accounts. Place prediction on tab A. Confirm toast appears on tab B.
- [ ] Place prediction on tab A. Confirm tab A does NOT show its own toast (self-suppression).
- [ ] Disconnect from internet, reconnect — WS reconnects automatically within 3s.
- [ ] Place prediction while WS is disconnected — prediction saves normally, no error shown.
- [ ] Nginx proxy: confirm WSS upgrade works on `soc.sitenow.cloud` (DevTools Network → WS → status 101).
- [ ] Custom bet: toast shows `"[Tiêu đề kèo] - [Lựa chọn]"` format.
- [ ] Champion prediction: toast shows team name as selection.
- [ ] Toast auto-dismisses after 5 seconds.

## Test Files

| File | Tests |
|------|-------|
| `backend/internal/ws/hub_test.go` | 3 hub tests (register, unregister, broadcast) |
| `backend/internal/service/wc_activity_test.go` | 12 unit tests (event formatting + MockHub) |
| `backend/internal/service/wc_integration_test.go` | existing suite — nil hub regression |
| `backend/internal/service/wc_custom_bet_service_test.go` | existing suite — nil hub regression |

## Test Results (2026-06-24)

```
ok  github.com/duyb/esport-score-tracker/internal/ws      — 3 tests PASS (race-clean)
ok  github.com/duyb/esport-score-tracker/internal/service — 12 new tests PASS (race-clean)
```

## Manual Testing

- Verify `ElNotification` stacks correctly when multiple bets fire rapidly
- Test on mobile viewport — toast at bottom-right should not obscure bet form
- Verify no console errors when WS server is not running

## Performance Testing

- `go test -race ./internal/ws/...` — passes clean, no race conditions detected

## Bug Tracking

- Regression: existing integration test suites pass unchanged with `hub=nil`
