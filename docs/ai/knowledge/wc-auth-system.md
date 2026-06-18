# WC Auth System

## Two Auth Flows

### Flow A — Existing player (password + Google link gate)

1. POST `/wc/auth/login` with `{ name, password }` → JWT + `{ google_linked: bool }`
2. If `google_linked = false` → frontend redirects to `/world-cup/link-google` (blocking)
3. User clicks link button → Google GSI popup → id_token sent to POST `/wc/auth/google/link`
4. Backend: `UPDATE wc_users SET google_id=sub WHERE id=userID AND google_id IS NULL`
5. Frontend stores updated user state, redirects to `/world-cup/predict`

### Flow B — New player (Google login → auto-create)

1. User clicks "Đăng nhập với Google" → GSI popup → id_token
2. POST `/wc/auth/google` with `{ id_token }` (no existing JWT needed)
3. Backend: look up by `google_id = sub`
   - Found → return JWT + user info
   - Not found → INSERT into `wc_users`, INSERT `wc_wallets` (balance=0), return JWT
4. Frontend → redirect to `/world-cup/predict`

## JWT Details

- Custom claims type `WcClaims` — contains `wc_user_id` (UUID), `name`, `isAdmin`
- 7-day expiry, HMAC signed
- Stored in localStorage as `wc_token`; user object stored as `wc_user` (JSON)
- Extracted in `WcJWTMiddleware`, injected into Gin context as `WcUserIDKey`, `WcUserNameKey`, `WcIsAdminKey`

## Middleware Stack

```
/api/v1/wc/
  ├── [public]  GET /config, GET /matches, GET /schedule         (no auth)
  ├── [auth]    POST /auth/login, POST /auth/google, /auth/google/link (no auth needed to log in)
  ├── [jwt]     WcJWTMiddleware → handler                        (any logged-in user)
  ├── [linked]  WcJWTMiddleware → WcGoogleLinkedMiddleware → handler  (must have Google linked)
  └── [admin]   WcJWTMiddleware → WcAdminMiddleware → handler    (admin only)
```

**`WcGoogleLinkedMiddleware`** — queries DB for `google_id`, blocks non-admin users without it:
- Returns 403 `{ "error": "google_not_linked" }`
- Admins bypass this gate entirely

## Frontend Route Guards

All enforced in `router/index.ts` `beforeEach`:

| Meta flag | Behaviour |
|-----------|-----------|
| `requiresWcAuth` | Validate `wc_token` in localStorage, redirect to `/world-cup/login` if missing |
| `requiresGoogleLink` | Check `user.googleLinked`, redirect to `/world-cup/link-google` if false |
| `requiresAdmin` | Check `user.isAdmin`, redirect to `/world-cup/predict` if false |
| `requiresWcFeature` | Fetch `/wc/config`, redirect to `/world-cup` schedule if feature off |
| `skipGoogleLinkCheck` | Used on `/link-google` itself to prevent redirect loop |

## DB Schema — `wc_users`

```sql
id            UUID PRIMARY KEY DEFAULT gen_random_uuid()
name          VARCHAR(100) NOT NULL UNIQUE
password_hash VARCHAR(255)          -- nullable (new Google-only accounts have no password)
google_id     VARCHAR(100) UNIQUE   -- nullable (sparse unique index)
avatar_url    VARCHAR(500)          -- nullable, pulled from Google profile picture
is_admin      BOOLEAN DEFAULT false
created_at    TIMESTAMPTZ
updated_at    TIMESTAMPTZ
```

`GoogleLinked()` method on `WcUser` model: returns `google_id != nil`.

## Key Service Errors

| Sentinel | Meaning |
|----------|---------|
| `ErrInvalidGoogleToken` | Google token verification failed |
| `ErrGoogleAlreadyLinked` | Another account already uses this Google ID |
| `ErrAlreadyLinked` | This user already has a Google account linked |

## Files

| File | Role |
|------|------|
| `backend/internal/service/wc_auth_service.go` | Login, Google verify/create/link, JWT sign/verify |
| `backend/internal/middleware/wc_auth.go` | JWT extract + Google-link gate |
| `backend/internal/api/wc_auth_handler.go` | HTTP handlers for auth routes |
| `frontend/src/stores/wcAuthStore.ts` | Pinia store, localStorage persistence |
| `frontend/src/services/wcAuthService.ts` | Axios wrappers for auth endpoints |
| `frontend/src/views/WcLoginView.vue` | Login page (password + Google button) |
| `frontend/src/views/WcLinkGoogleView.vue` | Mandatory Google-link page |
| `frontend/src/google-gsi.d.ts` | TypeScript types for Google GSI library |
