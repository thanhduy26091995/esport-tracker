---
phase: planning
title: WC Google OAuth Login — Planning
description: Task breakdown for hybrid auth model - mandatory Google link for existing players, Google-only creation for new players, profile management
---

# Project Planning & Task Breakdown

## Milestones

- [ ] **M1 — Backend:** Migration applied, new auth endpoints live, old endpoints removed, google-link middleware in place
- [ ] **M2 — Frontend:** Login page updated, blocking link-google page working, register removed, router guard active
- [ ] **M3 — Profile Management:** Profile page and avatar display across all 4 UI locations
- [ ] **M4 — QA & Deploy:** All flows verified, deployed with player notification

## Task Breakdown

### Phase 1: Backend

- [ ] **1.1** Add Go dependency: `google.golang.org/api/idtoken`
  ```bash
  cd backend && go get google.golang.org/api/idtoken
  ```

- [ ] **1.2** Write DB migration file
  - File: `backend/migrations/<timestamp>_wc_google_oauth.sql`
  - Add `google_id VARCHAR(100)` (nullable, unique partial index)
  - Add `avatar_url VARCHAR(500)` (nullable)
  - Make `password_hash` nullable (for new Google-only accounts)

- [ ] **1.3** Update `WcUser` model struct
  - File: `backend/internal/model/wc_user.go`
  - `GoogleID *string`, `AvatarURL *string`, `PasswordHash *string`

- [ ] **1.4** Implement Google auth service methods
  - File: `backend/internal/service/wc_auth_service.go`
  - `verifyGoogleToken(ctx, idToken) (*idtoken.Payload, error)`
  - `GoogleLoginOrCreate(ctx, idToken) (*WcUser, error)` — finds by google_id OR creates new account
  - `LinkGoogleToAccount(ctx, userID, idToken) error` — UPDATE WHERE id=userID AND google_id IS NULL
  - Remove: `Register()`, `ResetPassword()`

- [ ] **1.5** Add `google_linked` field to login response
  - File: `backend/internal/service/wc_auth_service.go` and `wc_auth_handler.go`
  - `Login()` returns now include `GoogleLinked bool` (derived from `user.GoogleID != nil`)

- [ ] **1.6** Implement profile service
  - File: `backend/internal/service/wc_profile_service.go` (new)
  - `GetProfile(ctx, userID) (*WcUser, error)`
  - `UpdateProfile(ctx, userID, name *string, avatarURL *string) (*WcUser, error)`

- [ ] **1.7** Update auth handler
  - File: `backend/internal/api/wc_auth_handler.go`
  - Add `HandleGoogleLoginOrCreate` → `POST /wc/auth/google`
  - Add `HandleGoogleLink` → `POST /wc/auth/google/link` (requires JWT, no google-link check)
  - Remove `HandleRegister`, `HandleResetPassword`
  - Update `HandleLogin` response to include `google_linked`

- [ ] **1.8** Create profile handler
  - File: `backend/internal/api/wc_profile_handler.go` (new)
  - `HandleGetProfile` → `GET /wc/profile`
  - `HandleUpdateProfile` → `PUT /wc/profile`

- [ ] **1.9** Add `WcGoogleLinkedMiddleware`
  - File: `backend/internal/middleware/wc_auth.go`
  - Reads `wc_user_id` from context (set by WcJWTMiddleware)
  - Queries DB: `SELECT google_id FROM wc_users WHERE id = ?`
  - If `google_id IS NULL` and user is not admin → returns `403 { "error": "google_not_linked" }`

- [ ] **1.10** Update router
  - File: `backend/internal/router/wc_router.go`
  - `POST /auth/google` — no auth middleware
  - `POST /auth/google/link` — WcJWTMiddleware only (not google-link check)
  - `GET /profile`, `PUT /profile` — WcJWTMiddleware + WcGoogleLinkedMiddleware
  - All other player routes — WcJWTMiddleware + WcGoogleLinkedMiddleware
  - Admin routes — WcJWTMiddleware + WcAdminMiddleware (no google-link check)
  - Remove `/auth/register`, `/auth/reset-password`

- [ ] **1.11** Add `GOOGLE_CLIENT_ID` env var
  - Update backend config struct and `.env.example`
  - Value: `685662087046-1u7mm9ov2pjbfhhstjib4s02su7sdjt3.apps.googleusercontent.com`

### Phase 2: Frontend — Login & Link Flows

- [ ] **2.1** Add GSI script and env var
  - `index.html`: `<script src="https://accounts.google.com/gsi/client" async defer></script>`
  - `.env` / `.env.production`: `VITE_GOOGLE_CLIENT_ID=685662087046-...`

- [ ] **2.2** Update types
  - File: `frontend/src/types/wc.ts`
  - Add `avatarUrl: string | null` and `googleLinked: boolean` to `WcAuthUser`
  - Add `WcProfile` interface: `{ id, name, avatarUrl, isAdmin, googleLinked, createdAt }`

- [ ] **2.3** Update `wcAuthService.ts`
  - Add `googleLogin(idToken)` → `POST /wc/auth/google`
  - Add `googleLink(idToken)` → `POST /wc/auth/google/link`
  - Remove `register()`, `resetPassword()`

- [ ] **2.4** Update `wcAuthStore.ts`
  - Keep `login()`, update to store `googleLinked` and `avatarUrl` from response
  - Add `loginWithGoogle(idToken)` action
  - Add `linkGoogle(idToken)` action
  - Remove `register()`, `resetPassword()`

- [ ] **2.5** Update `WcLoginView.vue`
  - Keep existing username/password form
  - Add "Đăng nhập với Google" button below the form (using GSI `renderButton`)
  - On Google credential callback: call `authStore.loginWithGoogle()` → if success, router guard handles redirect
  - Remove any link to register page

- [ ] **2.6** Create `WcLinkGoogleView.vue`
  - Displayed when `google_linked = false` after password login
  - Content: explanation text + "Liên kết tài khoản Google" button
  - On click: GSI popup → `authStore.linkGoogle(idToken)` → on success: redirect to `/world-cup/predict`
  - On 409: show "Google account này đã được liên kết với tài khoản khác"

- [ ] **2.7** Update router
  - File: `frontend/src/router/index.ts`
  - Add `/world-cup/link-google` route → `WcLinkGoogleView` (requires JWT, no google-link check)
  - Add `/world-cup/profile` route → `WcProfileView` (requires JWT + google-link)
  - Remove `/world-cup/register` route (redirect → `/world-cup/login`)
  - Update `beforeEach` guard:
    ```
    if requiresWcAuth AND not logged in → /world-cup/login
    if requiresWcAuth AND logged in AND !googleLinked AND route ≠ /world-cup/link-google → /world-cup/link-google
    ```

- [ ] **2.8** Delete `WcRegisterView.vue`

### Phase 3: Profile Management & Avatar

- [ ] **3.1** Create `wcProfileService.ts`
  - `getProfile()` → `GET /wc/profile`
  - `updateProfile({ name?, avatarUrl? })` → `PUT /wc/profile`

- [ ] **3.2** Create `WcProfileView.vue`
  - Show avatar (large `<img>` with onerror fallback), name, linked email (optional)
  - Form: name input (min 2 chars), avatar URL input with live preview
  - Save button with loading state; success/error toast

- [ ] **3.3** Add avatar to WC navbar/header
  - Show `<img :src="user.avatarUrl" onerror="...defaultIcon">` + player name chip
  - Link chip to `/world-cup/profile`

- [ ] **3.4** Add avatar to leaderboard/ranking table
  - Add avatar column (small, 32px) next to player name in ranking rows
  - Fallback: player initials in a coloured circle if no avatar or broken URL

- [ ] **3.5** Add avatar to bet/prediction history
  - Show small avatar thumbnail in each history row header

### Phase 4: QA & Deploy

- [ ] **4.1** Manual test: new player Google login → auto-account creation → access `/world-cup/predict`
- [ ] **4.2** Manual test: existing player password login → redirect to link page → Google link → access granted
- [ ] **4.3** Manual test: existing player (already linked) password login → no redirect → normal access
- [ ] **4.4** Manual test: existing player (already linked) Google login → normal access
- [ ] **4.5** Manual test: try to link same Google account to a second WC account → 409 shown
- [ ] **4.6** Manual test: profile update (name change, avatar URL, broken URL fallback)
- [ ] **4.7** Verify `POST /wc/auth/register` returns 404
- [ ] **4.8** Verify admin can still login with username/password without Google link
- [ ] **4.9** Send player notification (group chat): "Hệ thống WC yêu cầu liên kết tài khoản Google. Đăng nhập bằng mật khẩu → làm theo hướng dẫn để liên kết."
- [ ] **4.10** Deploy backend migration + new binary
- [ ] **4.11** Deploy frontend

## Dependencies

```
1.1 → 1.4, 1.5
1.2 → 1.3 → 1.4, 1.5, 1.6
1.4, 1.5 → 1.7
1.6 → 1.8
1.7, 1.8 → 1.10
1.9, 1.10, 1.11 → M1 complete

2.2 → 2.3 → 2.4 → 2.5, 2.6
2.1 → 2.5, 2.6
2.6 → 2.7
M1 needed for full integration test (can mock earlier)

M2 → 3.1 → 3.2, 3.3, 3.4, 3.5 → M3
M3 → Phase 4
```

## Timeline & Estimates

| Phase | Tasks | Effort |
|---|---|---|
| Phase 1 — Backend | 1.1–1.11 | ~1.5 days |
| Phase 2 — Frontend Login & Link | 2.1–2.8 | ~1 day |
| Phase 3 — Profile & Avatar | 3.1–3.5 | ~0.5 day |
| Phase 4 — QA & Deploy | 4.1–4.11 | ~0.5 day |
| **Total** | | **~3.5 days** |

## Risks & Mitigation

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| Google Client ID not authorised for dev/prod domain | Medium | High | Verify Google Console authorised origins before any frontend testing; add both `localhost:5173` and prod domain |
| Player can't access app after deploy (google link required but they're stuck) | Low | High | Link page is publicly accessible after JWT login; clear UX + player notification before deploy |
| Two concurrent link attempts for same google_id | Low | Medium | `UPDATE ... WHERE google_id IS NULL` is atomic; DB unique index is the final guard |
| Admin accidentally linked to Google (if admin tries Google login) | Low | Medium | `POST /wc/auth/google` auto-creates, not auto-links to existing. Admin won't be affected unless they explicitly link via the link endpoint |
| Avatar URL broken | Low | Low | `onerror` fallback on all `<img>` tags; no server-side fetch |
| Rollback needed post-deploy | Low | High | Migration is additive only (new nullable columns). Remove columns + revert code if needed. No data loss. |

## Resources Needed

- Google Cloud Console access (verify authorised origins for client ID)
- Backend env: `GOOGLE_CLIENT_ID=685662087046-1u7mm9ov2pjbfhhstjib4s02su7sdjt3.apps.googleusercontent.com`
- Frontend env: `VITE_GOOGLE_CLIENT_ID=685662087046-1u7mm9ov2pjbfhhstjib4s02su7sdjt3.apps.googleusercontent.com`
- Deploy window: any time — password login continues working during the migration; only new blocking gate appears post-deploy
