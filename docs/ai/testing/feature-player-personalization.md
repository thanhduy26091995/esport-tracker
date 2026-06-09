---
phase: testing
title: Testing – Player Personalization & Dynamic Global Theme
description: Test scope, cases, and validation for avatar upload, club selection, and dynamic theme
---

# Testing Strategy

## Scope

- **Unit:** Service-layer logic (MIME validation, filename generation, club slug validation, theme composable logic).
- **Integration:** API endpoints (upload, delete, club update) with real DB and real filesystem in test.
- **Manual / visual:** Avatar rendering in leaderboard, GIF animation, CSS theme switching.

## Test Files

| File | Layer | Status | Coverage |
|---|---|---|---|
| `backend/internal/service/user_service_test.go` | Service (unit + integration) | **Written** | MIME map, slug map, size guard, DB flows |
| `backend/internal/api/user_handler_test.go` | Handler/Integration | Deferred — no auth layer to test | — |
| `frontend/src/composables/useGlobalTheme.test.ts` | Composable | Deferred — vitest not installed | — |

## Unit Tests

### Avatar upload service
- Valid JPEG upload → returns `/uploads/avatars/{uuid}.jpg`
- Valid GIF upload → accepted
- SVG upload → rejected with error
- File > 2 MB → rejected (test at service level with byte limit mock)
- Second upload → previous file deleted, new URL returned
- User with no existing avatar → no delete attempted, no error

### Club validation
- Valid slug (`real-madrid`) → no error
- Empty string → no error (unset club is allowed)
- Unknown slug (`fake-club`) → error returned
- `none` slug → accepted, treated as "no club"

### `useGlobalTheme` composable
- Rank #1 has club `liverpool` → `--theme-primary` set to `#C8102E`
- Rank #1 has no club → `--theme-primary` set to default `#16a34a`
- Club changes from `barcelona` to `man-city` → CSS vars updated
- Club unchanged (same value on re-render) → CSS vars NOT re-written (no flicker)

## Integration Tests

### `PUT /api/v1/users/:id/avatar`
- Valid image upload by correct user → `200` + `{ avatar_url: "..." }`
- Valid image upload by different (non-admin) user → `403`
- File too large (>2 MB) → `413` or `400`
- Unsupported file type (PDF, SVG) → `400`
- User not found → `404`

### `DELETE /api/v1/users/:id/avatar`
- Admin deletes avatar → `204`, file removed from disk, `avatar_url` set to null
- Non-admin attempts delete → `403`

### `PUT /api/v1/users/:id/club`
- Valid club slug → `200` + updated user
- Invalid club slug → `400`
- Empty string (clear club) → `200`

### `GET /api/v1/users/leaderboard`
- Response includes `avatar_url` and `favorite_club` fields (even if null)

## Test Data & Environments

- Use a temp directory (`t.TempDir()`) for file writes in tests — never write to production `./uploads/`.
- Seed one test user with `is_active = true`.
- For MIME tests, create minimal valid file headers in memory (no real image files needed).

## Execution

```bash
# Backend — no-DB unit tests (always runnable)
cd backend && go test ./internal/service/... -run "TestValidClubSlugs|TestAllowedAvatarMIME|TestUpdateClub_Unknown|TestUpdateClub_Malformed|TestUploadAvatar_File|TestUploadAvatar_Exactly|TestUploadAvatar_Unsupported" -v

# Backend — integration tests (require Postgres)
export TEST_DATABASE_URL="host=localhost port=5432 user=postgres password=secret dbname=esport_test sslmode=disable"
cd backend && go test ./internal/service/... -run "TestUpdateClub_Valid|TestUpdateClub_Empty|TestUpdateClub_NewLeague|TestUpdateClub_Unknown|TestUploadAvatar_Valid|TestUploadAvatar_Second|TestDeleteAvatar" -v

# Frontend — vitest not installed; add with: npm install -D vitest @vue/test-utils happy-dom
```

## Coverage & Quality Gates

- MIME validation path: must be covered (security-critical).
- Auth check (user can only update own avatar): must be covered.
- CSS var fallback to default theme: must be covered.
- GIF animation: manual visual check only (no automated test).

## Risks & Gaps

- **GIF performance on leaderboard:** Not automated — requires manual check with a large animated GIF.
- **File deletion race condition:** If two uploads happen simultaneously, the old file deletion could race. Low risk for MVP (single-user profile updates). Deferred.
- **Theme transition animation:** CSS transition on `--theme-primary` requires browser support for `@property` (Chrome 85+). Visual check needed.
