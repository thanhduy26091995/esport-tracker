---
phase: testing
title: WC Google OAuth Login — Testing Strategy
description: Test cases for hybrid auth (password + mandatory Google link), Google-only new player flow, and profile management
---

# Testing Strategy

## Scope

- **Unit tests:** Google token verification, GoogleLoginOrCreate, LinkGoogleToAccount, profile update validation, WcGoogleLinkedMiddleware
- **Integration tests (HTTP):** All new/changed auth endpoints; profile endpoints; middleware enforcement
- **Manual E2E:** Full Google OAuth popup flows (requires real browser + Google account)

## Test Files

| File | Layer | Status | Coverage Target |
|---|---|---|---|
| `backend/internal/service/wc_auth_service_test.go` | Service | ✅ Written | Google login/create, link, error cases |
| `backend/internal/service/wc_profile_service_test.go` | Service | ✅ Written | GetProfile, UpdateProfile, name conflict |
| `backend/internal/middleware/wc_auth_test.go` | Middleware | ✅ Written | WcGoogleLinkedMiddleware (linked / not linked / admin) |
| `backend/internal/api/wc_auth_handler_test.go` | HTTP handler | ⏭ Deferred | All auth endpoints + updated login response |
| `backend/internal/api/wc_profile_handler_test.go` | HTTP handler | ⏭ Deferred | GET/PUT profile |

**Implementation note:** `WcAuthService.verifyGoogle` is now injectable via `.withVerifier()`, so service tests never call the real Google API.

## Unit Tests

### `WcAuthService`

```
TestVerifyGoogleToken_ValidToken     → returns payload with correct sub, name, picture
TestVerifyGoogleToken_Expired        → returns ErrInvalidGoogleToken
TestVerifyGoogleToken_WrongAudience  → returns ErrInvalidGoogleToken

TestGoogleLoginOrCreate_ExistingLinked → returns existing WcUser (no INSERT)
TestGoogleLoginOrCreate_NewPlayer      → inserts WcUser + WcWallet, returns created user
TestGoogleLoginOrCreate_NameTaken      → appends suffix to ensure unique name, succeeds

TestLinkGoogleToAccount_OK             → UPDATE sets google_id; 1 row affected
TestLinkGoogleToAccount_AlreadyLinked  → 0 rows affected (user already has google_id) → ErrAlreadyLinked
TestLinkGoogleToAccount_GoogleTaken    → unique constraint violation → ErrGoogleAlreadyLinked
TestLinkGoogleToAccount_Concurrent     → two goroutines; only one succeeds

TestLogin_ExistingUserGoogleLinked     → response includes google_linked: true
TestLogin_ExistingUserNotLinked        → response includes google_linked: false
```

### `WcProfileService`

```
TestGetProfile_OK           → returns all fields
TestGetProfile_NotFound     → error
TestUpdateProfile_Name      → name updated; avatar unchanged
TestUpdateProfile_AvatarURL → avatar updated; name unchanged
TestUpdateProfile_Both      → both updated
TestUpdateProfile_NameTaken → ErrNameTaken
TestUpdateProfile_NameTooShort → validation error (< 2 chars)
TestUpdateProfile_Empty     → returns current profile unchanged
```

### `WcGoogleLinkedMiddleware`

```
TestMiddleware_LinkedUser  → c.Next() called; no abort
TestMiddleware_UnlinkedUser → 403 returned with {"error":"google_not_linked"}
TestMiddleware_AdminUser   → c.Next() called (admin bypass)
```

## Integration Tests (HTTP)

### `POST /api/v1/wc/auth/login` (updated)

| Case | Setup | Expected |
|---|---|---|
| Valid password, linked account | `google_id IS NOT NULL` | 200, `google_linked: true` |
| Valid password, unlinked account | `google_id IS NULL` | 200, `google_linked: false` |
| Wrong password | — | 401 |

### `POST /api/v1/wc/auth/google`

| Case | Setup | Expected |
|---|---|---|
| Valid token, existing linked account | Row with matching `google_id` | 200 + JWT, `google_linked: true` |
| Valid token, no account for this google_id | No matching row | 200 + JWT, new row created |
| Invalid token | — | 400 |
| Missing `id_token` | — | 400 |

### `POST /api/v1/wc/auth/google/link`

| Case | Setup | Expected |
|---|---|---|
| Valid JWT + valid token, user has no google_id | `google_id IS NULL` | 200, `google_linked: true` |
| Valid JWT + valid token, google_id already taken | Another user has this google_id | 409 |
| Valid JWT + valid token, user already linked | `google_id IS NOT NULL` | 409 |
| No JWT | — | 401 |
| Invalid Google token | — | 400 |

### `GET /api/v1/wc/profile`

| Case | Expected |
|---|---|
| Valid JWT, google linked | 200 + profile JSON |
| Valid JWT, not linked | 403 `google_not_linked` |
| No JWT | 401 |

### `PUT /api/v1/wc/profile`

| Case | Expected |
|---|---|
| Valid JWT + linked, valid body | 200 + updated profile |
| Name taken | 409 |
| Empty body | 200 + unchanged profile |
| Not linked | 403 |
| No JWT | 401 |

### Removed endpoints

| Endpoint | Expected |
|---|---|
| `POST /wc/auth/register` | 404 |
| `POST /wc/auth/reset-password` | 404 |

## Manual / E2E Checklist

**New player flow:**
- [ ] Visit `/world-cup/login` → click "Đăng nhập với Google" → Google popup → consent → redirected to `/world-cup/predict`
- [ ] Navbar shows Google display name and avatar
- [ ] Second login with same Google account → goes directly in (no duplicate account created)

**Existing player link flow:**
- [ ] Login with username/password → redirected to `/world-cup/link-google`
- [ ] Click "Liên kết với Google" → Google popup → consent → redirected to `/world-cup/predict`
- [ ] All existing predictions/balance/bets intact after link
- [ ] Next password login → goes directly to predict (no link redirect)

**Conflict case:**
- [ ] Two browsers: both try to link the same Google account → one succeeds, other sees error toast

**Profile management:**
- [ ] Navigate to `/world-cup/profile` → see current name and avatar
- [ ] Update name → saved → reflected in navbar and leaderboard
- [ ] Update avatar URL → live preview updates → saved → reflected in all 4 locations
- [ ] Enter broken image URL → broken URL is saved but img shows fallback icon everywhere
- [ ] Try to use another player's name → 409 toast shown

**Avatar display:**
- [ ] Navbar: avatar thumbnail visible
- [ ] Leaderboard: avatar column visible for all players (fallback for those without avatar)
- [ ] Bet/prediction history: small avatar in rows

**Admin:**
- [ ] Admin logs in with username/password → no `/world-cup/link-google` redirect → admin panel accessible
- [ ] `POST /wc/auth/register` → 404

## Test Data & Environments

**Seeds needed:**
- `wc_user` with `google_id IS NOT NULL` (already linked)
- `wc_user` with `google_id IS NULL` (old unlinked player)
- Admin `wc_user` (`is_admin = true`, `google_id IS NULL`)

**Mock strategy for unit tests:**
- Inject a mock `idtoken.Validator` interface — never call real Google API in automated tests
- Use a test DB (SQLite in-memory or test PostgreSQL) for integration tests

## Execution

```bash
cd backend

# Run all WC-related tests
go test ./internal/... -v -run "Wc"

# Specific new suites
go test ./internal/service/... -run "TestGoogleLogin|TestLink|TestProfile"
go test ./internal/middleware/... -run "TestWcGoogleLinked"
go test ./internal/api/... -run "TestWcAuth|TestWcProfile"

# With race detector (important for concurrent link test)
go test -race ./internal/service/... -run "TestLinkGoogleToAccount_Concurrent"

# Coverage
go test ./internal/... -coverprofile=coverage.out && go tool cover -html=coverage.out
```

## Coverage & Quality Gates

- New service files: ≥ 80% line coverage
- Critical paths (must be green before deploy):
  - [ ] `TestLinkGoogleToAccount_Concurrent` (race detector enabled)
  - [ ] `TestMiddleware_UnlinkedUser` → 403
  - [ ] `TestMiddleware_AdminUser` → bypass
  - [ ] `POST /wc/auth/google` — new account creation
  - [ ] `POST /wc/auth/google/link` — success + conflict cases
  - [ ] `POST /wc/auth/register` → 404

## Risks & Gaps

- Real Google OAuth popup cannot be automated in CI → covered by manual E2E checklist
- `WcGoogleLinkedMiddleware` does a DB query per request — acceptable for current load; add caching if needed at scale
- Name uniqueness suffix logic in `GoogleLoginOrCreate` needs edge-case testing (e.g. "Player", "Player1", "Player2" all taken)
