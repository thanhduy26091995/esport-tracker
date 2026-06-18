---
phase: implementation
title: WC Google OAuth Login — Implementation Guide
description: Technical notes for hybrid Google-link auth and profile management for WC players
---

# Implementation Guide

## Development Setup

### New backend dependency
```bash
cd backend && go get google.golang.org/api/idtoken
```

### Environment variables

**Backend `.env`:**
```
GOOGLE_CLIENT_ID=685662087046-1u7mm9ov2pjbfhhstjib4s02su7sdjt3.apps.googleusercontent.com
WC_JWT_SECRET=<existing — unchanged>
```

**Frontend `.env`:**
```
VITE_GOOGLE_CLIENT_ID=685662087046-1u7mm9ov2pjbfhhstjib4s02su7sdjt3.apps.googleusercontent.com
VITE_API_BASE_URL=http://localhost:8080/api/v1
```

### Google Console setup
Authorised JavaScript origins for the client ID must include:
- `http://localhost:5173` (dev)
- `https://<prod-domain>` (production)

No redirect URIs needed (credential callback pattern, not redirect flow).

---

## Code Structure

```
backend/internal/
  model/
    wc_user.go              ← add GoogleID *string, AvatarURL *string; make PasswordHash *string
  service/
    wc_auth_service.go      ← add GoogleLoginOrCreate, LinkGoogleToAccount; remove Register, ResetPassword
    wc_profile_service.go   ← NEW
  api/
    wc_auth_handler.go      ← add HandleGoogleLoginOrCreate, HandleGoogleLink; update HandleLogin response
    wc_profile_handler.go   ← NEW
  middleware/
    wc_auth.go              ← add WcGoogleLinkedMiddleware
  router/
    wc_router.go            ← update route wiring
  migrations/
    <ts>_wc_google_oauth.sql ← NEW

frontend/src/
  views/
    WcLoginView.vue         ← add Google button, remove register link
    WcLinkGoogleView.vue    ← NEW (blocking link page)
    WcProfileView.vue       ← NEW
    WcRegisterView.vue      ← DELETE
  stores/
    wcAuthStore.ts          ← add loginWithGoogle, linkGoogle; keep login; remove register, resetPassword
  services/
    wcAuthService.ts        ← add googleLogin, googleLink; remove register, resetPassword
    wcProfileService.ts     ← NEW
  types/
    wc.ts                   ← add avatarUrl, googleLinked to WcAuthUser; add WcProfile
  router/
    index.ts                ← update guards and routes
```

---

## Implementation Notes

### 1. DB Migration

```sql
-- backend/migrations/<timestamp>_wc_google_oauth.sql
ALTER TABLE wc_users
  ADD COLUMN IF NOT EXISTS google_id  VARCHAR(100),
  ADD COLUMN IF NOT EXISTS avatar_url VARCHAR(500),
  ALTER COLUMN password_hash          DROP NOT NULL;

CREATE UNIQUE INDEX IF NOT EXISTS idx_wc_users_google_id
  ON wc_users (google_id)
  WHERE google_id IS NOT NULL;
```

### 2. WcUser model (`wc_user.go`)

```go
type WcUser struct {
  ID           uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
  Name         string    `gorm:"size:100;not null;uniqueIndex"`
  PasswordHash *string   `gorm:"size:255"`
  GoogleID     *string   `gorm:"size:100;uniqueIndex"`
  AvatarURL    *string   `gorm:"size:500"`
  IsAdmin      bool      `gorm:"default:false"`
  CreatedAt    time.Time
  UpdatedAt    time.Time
}

func (u *WcUser) GoogleLinked() bool { return u.GoogleID != nil }
```

### 3. Google token verification

```go
import "google.golang.org/api/idtoken"

func (s *WcAuthService) verifyGoogleToken(ctx context.Context, idToken string) (*idtoken.Payload, error) {
    return idtoken.Validate(ctx, idToken, s.googleClientID)
    // payload.Subject = stable Google user ID (use as google_id in DB)
    // payload.Claims["name"], ["picture"], ["email"]
}
```

Certificates are fetched from Google once and cached automatically by the library.

### 4. GoogleLoginOrCreate (`wc_auth_service.go`)

```go
func (s *WcAuthService) GoogleLoginOrCreate(ctx context.Context, idToken string) (*WcUser, error) {
    payload, err := s.verifyGoogleToken(ctx, idToken)
    if err != nil {
        return nil, ErrInvalidGoogleToken
    }

    var user WcUser
    err = s.db.WithContext(ctx).Where("google_id = ?", payload.Subject).First(&user).Error
    if err == nil {
        return &user, nil // existing linked account
    }
    if !errors.Is(err, gorm.ErrRecordNotFound) {
        return nil, err
    }

    // Auto-create new account
    googleName, _ := payload.Claims["name"].(string)
    googlePic, _  := payload.Claims["picture"].(string)
    if googleName == "" {
        googleName = "Player"
    }
    // Ensure name uniqueness (append suffix if taken)
    name := s.uniqueName(ctx, googleName)

    user = WcUser{
        GoogleID:  &payload.Subject,
        Name:      name,
        AvatarURL: &googlePic,
    }
    if err := s.db.Create(&user).Error; err != nil {
        return nil, err
    }
    // Create wallet
    s.db.Create(&WcWallet{WcUserID: user.ID})
    return &user, nil
}
```

### 5. LinkGoogleToAccount (`wc_auth_service.go`)

```go
func (s *WcAuthService) LinkGoogleToAccount(ctx context.Context, userID uuid.UUID, idToken string) error {
    payload, err := s.verifyGoogleToken(ctx, idToken)
    if err != nil {
        return ErrInvalidGoogleToken
    }

    result := s.db.WithContext(ctx).Model(&WcUser{}).
        Where("id = ? AND google_id IS NULL", userID).
        Updates(map[string]interface{}{
            "google_id":  payload.Subject,
            "avatar_url": payload.Claims["picture"],
        })
    if result.Error != nil {
        if isUniqueViolation(result.Error) {
            return ErrGoogleAlreadyLinked // this google_id belongs to another account
        }
        return result.Error
    }
    if result.RowsAffected == 0 {
        // Either user not found, or already has a google_id
        return ErrAlreadyLinked
    }
    return nil
}
```

### 6. WcGoogleLinkedMiddleware (`middleware/wc_auth.go`)

```go
func WcGoogleLinkedMiddleware(db *gorm.DB) gin.HandlerFunc {
    return func(c *gin.Context) {
        userID := c.MustGet("wc_user_id").(uuid.UUID)
        isAdmin := c.GetBool("wc_is_admin")
        if isAdmin {
            c.Next()
            return
        }
        var u struct{ GoogleID *string }
        db.Model(&WcUser{}).Select("google_id").Where("id = ?", userID).Scan(&u)
        if u.GoogleID == nil {
            c.AbortWithStatusJSON(403, gin.H{"error": "google_not_linked"})
            return
        }
        c.Next()
    }
}
```

### 7. Updated Login response (`wc_auth_handler.go`)

```go
// Extend the existing login response struct
type WcAuthResponse struct {
    Token        string    `json:"token"`
    UserID       uuid.UUID `json:"user_id"`
    Name         string    `json:"name"`
    AvatarURL    *string   `json:"avatar_url"`
    IsAdmin      bool      `json:"is_admin"`
    GoogleLinked bool      `json:"google_linked"`
}
```

### 8. Route wiring (`wc_router.go`)

```go
auth := wc.Group("/auth")
{
    auth.POST("/login",        handler.HandleLogin)         // kept
    auth.POST("/google",       handler.HandleGoogleLoginOrCreate)  // new
    // JWT only — no google-link check (this IS the link endpoint)
    auth.POST("/google/link",  jwtMW, handler.HandleGoogleLink)    // new
    // REMOVED: /register, /reset-password
}

player := wc.Group("/", jwtMW, googleLinkedMW)
{
    player.GET("/profile",  profileHandler.HandleGetProfile)
    player.PUT("/profile",  profileHandler.HandleUpdateProfile)
    // ... existing player routes ...
}

admin := wc.Group("/admin", jwtMW, adminMW)
{
    // existing admin routes — no googleLinkedMW
}
```

### 9. Google Sign-In button (frontend)

**`index.html`:**
```html
<script src="https://accounts.google.com/gsi/client" async defer></script>
```

**`WcLoginView.vue`:**
```typescript
onMounted(() => {
  window.google?.accounts.id.initialize({
    client_id: import.meta.env.VITE_GOOGLE_CLIENT_ID,
    callback: async ({ credential }) => {
      await authStore.loginWithGoogle(credential)
      // router guard handles redirect based on googleLinked
    },
    auto_select: false,
  })
  window.google?.accounts.id.renderButton(
    document.getElementById('google-signin-btn')!,
    { theme: 'outline', size: 'large', text: 'signin_with', locale: 'vi' }
  )
})
```

**`WcLinkGoogleView.vue`:**
```typescript
onMounted(() => {
  window.google?.accounts.id.initialize({
    client_id: import.meta.env.VITE_GOOGLE_CLIENT_ID,
    callback: async ({ credential }) => {
      const ok = await authStore.linkGoogle(credential)
      if (ok) router.push('/world-cup/predict')
    },
    auto_select: false,
  })
  window.google?.accounts.id.renderButton(
    document.getElementById('link-google-btn')!,
    { theme: 'filled_blue', size: 'large', text: 'continue_with', locale: 'vi' }
  )
})
```

### 10. Router guard

```typescript
router.beforeEach((to, _from) => {
  const auth = useWcAuthStore()

  if (to.meta.requiresWcAuth && !auth.isLoggedIn) {
    return { path: '/world-cup/login' }
  }

  const isLinkPage = to.path === '/world-cup/link-google'
  if (to.meta.requiresWcAuth && auth.isLoggedIn && !auth.user?.googleLinked && !isLinkPage) {
    return { path: '/world-cup/link-google' }
  }
})
```

### 11. Avatar fallback (all locations)

```html
<img
  :src="avatarUrl || defaultIcon"
  @error="(e) => (e.target as HTMLImageElement).src = defaultIcon"
  class="w-8 h-8 rounded-full object-cover"
/>
```

---

## Error Handling

| Scenario | Backend | Frontend message |
|---|---|---|
| Invalid/expired Google token | 400 | "Đăng nhập Google thất bại. Vui lòng thử lại." |
| Google account already linked to another WC account | 409 | "Tài khoản Google này đã được liên kết với người dùng khác." |
| Existing account already has a google_id (link attempt) | 409 | "Tài khoản của bạn đã được liên kết." |
| Profile name taken | 409 | "Tên này đã được sử dụng bởi người chơi khác." |
| Profile name too short | 422 | "Tên phải có ít nhất 2 ký tự." |
| 403 google_not_linked (backend middleware) | 403 | Frontend router guard redirects to /world-cup/link-google before this fires |

---

## Security Notes

- `idtoken.Validate` verifies both `aud` (client_id) and `exp` — no replay possible after token expiry (1 hour Google default).
- `google_id` (sub) is never included in API responses to clients.
- Link atomicity: `UPDATE WHERE google_id IS NULL` + DB unique index prevents two concurrent links.
- Admin accounts bypass `WcGoogleLinkedMiddleware` — they are never locked out by this feature.
- New Google-only accounts have `password_hash = NULL` — `bcrypt.CompareHashAndPassword` would fail if called, preventing password login for these accounts.
