---
phase: testing
title: WC Admin Block/Unblock User — Testing Strategy
description: Test scope for user block/unblock feature
---

# Testing Strategy

## Scope

- Unit: service BlockUser/UnblockUser + PlaceBet guard
- Integration: API block → bet blocked; unblock → bet works
- Manual: admin panel UI toggle

## Test Files

| File | Layer | Coverage Target |
|------|-------|----------------|
| `backend/internal/service/wc_service_test.go` | Service | BlockUser self-block guard, PlaceBet blocked check |

## Unit Tests

- `BlockUser(adminID, adminID)` → returns error (self-block guard)
- `BlockUser(adminID, otherID)` → success, `is_blocked = true`
- `UnblockUser(blockedUserID)` → success, `is_blocked = false`
- `PlaceBet` with blocked user → returns "user is blocked" error
- `PlaceBet` with non-blocked user → proceeds normally

## Integration Tests (Manual)

1. Block user via `PUT /admin/users/:id/block` → 200
2. Blocked user calls `POST /wc/matches/:id/bet` → 403 with `{"error": "user is blocked from placing bets"}`
3. Unblock user via `PUT /admin/users/:id/unblock` → 200
4. Same user calls bet → proceeds (no 403)
5. Admin tries to block themselves → 400

## Execution

```bash
cd backend && go test ./internal/service/... -run TestBlock -v
```

## Risks & Gaps

- `PlaceBet` handler error mapping: cần verify strings.Contains approach robust (hoặc dùng sentinel error)
- `fetchAllUsers` phải include `is_blocked` trong response — verify type từ `GET /admin/users`
