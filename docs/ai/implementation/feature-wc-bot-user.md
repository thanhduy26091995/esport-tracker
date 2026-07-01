---
phase: implementation
title: WC Bot User Flag — Implementation Guide
description: Technical notes for is_bot flag
---

# Implementation Guide

## Key Implementation Notes

### Backend: `GetLeaderboard` query change

Add `u.is_bot` to the SELECT in `wc_repository.go`:

```sql
SELECT
    u.id          AS wc_user_id,
    u.name,
    u.avatar_url,
    u.is_bot,                          -- ADD THIS
    COALESCE(w.balance, 0) AS net_points,
    ...
```

### Backend: `SetUserBot` repo method

```go
func (r *WcRepository) SetUserBot(tx *gorm.DB, userID uuid.UUID, isBot bool) error {
    db := r.db
    if tx != nil { db = tx }
    return db.Model(&model.WcUser{}).
        Where("id = ?", userID).
        Update("is_bot", isBot).Error
}
```

### Backend: Handler pattern (follows `SetUserRole`)

```go
func (h *WcHandler) SetUserBot(c *gin.Context) {
    userID, err := uuid.Parse(c.Param("wc_user_id"))
    if err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user ID"})
        return
    }
    var req struct {
        IsBot bool `json:"is_bot"`
    }
    if err := c.ShouldBindJSON(&req); err != nil {
        c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
        return
    }
    if err := h.svc.SetUserBot(userID, req.IsBot); err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusOK, gin.H{"ok": true})
}
```

### Frontend: Banner fix

```ts
// WcTop3Banner.vue
const top3 = computed(() =>
  wcStore.leaderboard.filter(e => !e.is_bot).slice(0, 3)
)
```

`displayEntries` spreads `top3` twice for the infinite marquee loop — no change needed there.

### Frontend: Bot badge in `WcLeaderboard.vue`

Add a small inline badge after the user name:
```html
<span v-if="entry.is_bot" class="wc-bot-badge">Bot</span>
```

CSS: small pill, neutral color (grey/blue), similar to the existing admin badge pattern.

### Frontend: Admin panel toggle

In the user list row, add a button next to "Block/Unblock":
```html
<el-button size="small" text @click="toggleBot(user)">
  {{ user.is_bot ? 'Unbot' : 'Bot' }}
</el-button>
```

Or a more explicit label: "Đánh dấu Bot" / "Bỏ Bot".

## Integration Points

- `PUT /api/v1/wc/admin/users/:id/set-bot` — admin-only, guarded by `WcAdminMiddleware`
- `GET /api/v1/wc/leaderboard` — now includes `is_bot` in each entry

## Error Handling

- If user not found in `SetUserBot`, GORM returns no rows affected — handle gracefully (treat as no-op or 404).
- Client-side: if `is_bot` is missing on a leaderboard entry (old API), `undefined` is falsy — `filter(e => !e.is_bot)` still works correctly.
