---
phase: implementation
title: WC Admin Block/Unblock User — Implementation Guide
description: Technical notes for implementing the user block/unblock feature
---

# Implementation Guide

## Code Structure

```
backend/internal/model/wc_user.go               ← thêm IsBlocked field
backend/internal/repository/wc_user_repository.go ← thêm SetBlocked method
backend/internal/service/wc_service.go           ← thêm BlockUser/UnblockUser, sửa PlaceBet
backend/internal/api/wc_handler.go              ← thêm BlockUser/UnblockUser handlers, sửa PlaceBet error code
backend/internal/api/router.go                  ← đăng ký routes

frontend/src/types/wc.ts                        ← thêm is_blocked field
frontend/src/services/wcService.ts              ← thêm blockUser/unblockUser
frontend/src/components/wc/WcAdminPanel.vue     ← thêm block/unblock UI
```

## Implementation Notes

### Core Features

**1. Model Update**

File: `backend/internal/model/wc_user.go`
```go
type WcUser struct {
    // ...existing fields...
    IsBlocked bool `gorm:"not null;default:false" json:"is_blocked"`
}
```

GORM `AutoMigrate` sẽ tự `ADD COLUMN is_blocked BOOLEAN NOT NULL DEFAULT FALSE` nếu `WcUser{}` đã có trong `AutoMigrate()` call. Kiểm tra `database.go` để xác nhận.

**2. PlaceBet Guard**

Hiện tại `WcService.PlaceBet` không fetch user info — cần thêm:
```go
func (s *WcService) PlaceBet(betterID uuid.UUID, matchID uuid.UUID, ...) (*model.WcBet, error) {
    // Thêm đầu function:
    user, err := s.userRepo.GetByID(betterID)
    if err != nil {
        return nil, fmt.Errorf("user not found")
    }
    if user.IsBlocked {
        return nil, fmt.Errorf("user is blocked from placing bets")
    }
    // ...rest of existing logic...
}
```

**3. Handler 403 Mapping**

Trong `WcHandler.PlaceBet`, cần phân biệt blocked error:
```go
bet, err := h.svc.PlaceBet(...)
if err != nil {
    if strings.Contains(err.Error(), "blocked") {
        c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
        return
    }
    c.JSON(http.StatusUnprocessableEntity, gin.H{"error": err.Error()})
    return
}
```

**4. Frontend Self-Block Disable**

Cần lấy current user ID từ WcAuth store:
```ts
import { useWcAuthStore } from '@/stores/wcAuthStore'
const authStore = useWcAuthStore()
const currentUserId = computed(() => authStore.user?.id)
```

Disable button:
```html
:disabled="user.id === currentUserId"
```

**5. Badge Style**

Thêm CSS cho `wc-blocked-tag` tương tự `wc-admin-tag`:
```css
.wc-blocked-tag {
  font-size: 10px;
  font-weight: 700;
  background: rgba(239, 68, 68, 0.12);
  color: #ef4444;
  padding: 1px 6px;
  border-radius: 4px;
}
```

## Integration Points

- `WcUserRepository.GetByID` phải trả đầy đủ WcUser kể cả `is_blocked`
- `store.fetchAllUsers()` reload users sau toggle — đảm bảo `WcUser` response từ `GET /admin/users` include `is_blocked`

## Error Handling

- Self-block → 400 Bad Request
- User not found → 404
- DB error → 500

## Security Notes

- Block/unblock chỉ qua WcAdminMiddleware
- Blocked user vẫn authenticate được (JWT valid) — chỉ bị chặn ở action PlaceBet
- Service-layer check, không phải middleware — chặt hơn vì không thể bypass
