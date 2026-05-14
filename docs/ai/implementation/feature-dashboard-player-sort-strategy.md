---
phase: implementation
title: Dashboard Player Sort Strategy — Implementation Guide
description: Step-by-step technical notes for implementing the player sort strategy feature
---

# Implementation Guide

## Development Setup

```bash
# Backend
cd backend && go run ./cmd/server

# Frontend
cd frontend && npm run dev
```

## Code Structure

Files to create/modify:

```
backend/internal/
  model/user.go                       ← MODIFY: add UserWithPaymentTotal struct
  repository/user_repository.go       ← MODIFY: add GetPaymentRanking()
  service/user_service.go             ← MODIFY: add GetPaymentRanking()
  api/user_handler.go                 ← MODIFY: add GetPaymentRanking handler
  api/router.go                       ← MODIFY: register GET /users/payment-ranking

frontend/src/
  types/user.ts                       ← MODIFY: add UserWithPaymentTotal interface
  services/userService.ts             ← MODIFY: add getPaymentRanking()
  utils/sort.ts                       ← NEW: sortByStrategy() for strategies 2 & 3
  components/shared/Leaderboard.vue   ← MODIFY: remove internal sort (pure display)
  views/DashboardView.vue             ← MODIFY: sort state, payment fetch, computed, UI
  views/UsersView.vue                 ← MODIFY: sort control + sort computed (scope expanded)
  locales/en.json                     ← MODIFY: add 3 sort label keys
  locales/vi.json                     ← MODIFY: add 3 sort label keys
```

## Implementation Notes

### Backend

**`backend/internal/model/user.go`** — append:
```go
type UserWithPaymentTotal struct {
    User
    TotalPaid float64 `json:"total_paid"`
}
```

**`backend/internal/repository/user_repository.go`** — new method:
```go
func (r *UserRepository) GetPaymentRanking() ([]*model.UserWithPaymentTotal, error) {
    var results []*model.UserWithPaymentTotal
    err := r.db.Raw(`
        SELECT u.*, COALESCE(SUM(s.money_amount), 0) AS total_paid
        FROM users u
        LEFT JOIN debt_settlements s ON u.id = s.debtor_id
        WHERE u.is_active = true
        GROUP BY u.id
        ORDER BY total_paid DESC, u.name ASC
    `).Scan(&results).Error
    return results, err
}
```

**`backend/internal/service/user_service.go`** — new method:
```go
func (s *UserService) GetPaymentRanking() ([]*model.UserWithPaymentTotal, error) {
    return s.userRepo.GetPaymentRanking()
}
```

**`backend/internal/api/user_handler.go`** — new handler:
```go
func (h *UserHandler) GetPaymentRanking(c *gin.Context) {
    users, err := h.userService.GetPaymentRanking()
    if err != nil {
        c.JSON(http.StatusInternalServerError, gin.H{
            "error": gin.H{"code": "INTERNAL_ERROR", "message": "Failed to fetch payment ranking"},
        })
        return
    }
    c.JSON(http.StatusOK, users)
}
```

**`backend/internal/api/router.go`** — register route before `/:id`:
```go
users.GET("/payment-ranking", userHandler.GetPaymentRanking)
users.GET("/:id", userHandler.GetByID)  // must come after
```

### Frontend Types & Service

**`frontend/src/types/user.ts`** — append:
```ts
export interface UserWithPaymentTotal extends User {
  total_paid: number
}
```

**`frontend/src/services/userService.ts`** — add to the object:
```ts
async getPaymentRanking(): Promise<UserWithPaymentTotal[]> {
  const response = await api.get<UserWithPaymentTotal[]>('/users/payment-ranking')
  return response.data
}
```

### Frontend Sort Utility

**`frontend/src/utils/sort.ts`** (new file):
```ts
import type { User } from '@/types/user'

export type PlayerSortStrategy = 'debt-first' | 'default' | 'winners-first'

// Handles strategies 2 and 3 only.
// Strategy 1 ('debt-first') uses backend-sorted data — do not pass it here.
export function sortByStrategy(users: User[], strategy: PlayerSortStrategy): User[] {
  const byName = (a: User, b: User) => a.name.localeCompare(b.name)
  const neg = [...users.filter(u => u.current_score < 0)].sort((a, b) => a.current_score - b.current_score || byName(a, b))
  const pos = [...users.filter(u => u.current_score > 0)].sort((a, b) => b.current_score - a.current_score || byName(a, b))
  const zer = [...users.filter(u => u.current_score === 0)].sort(byName)

  switch (strategy) {
    case 'winners-first': return [...pos, ...zer, ...neg]
    default:              return [...neg, ...pos, ...zer]  // 'default' + fallback
  }
}
```

### Frontend Leaderboard (pure display)

**`frontend/src/components/shared/Leaderboard.vue`** — replace `displayUsers` computed:
```ts
// Remove: sort((a, b) => b.current_score - a.current_score)
// Replace with:
const displayUsers = computed(() => {
  return props.limit ? props.users.slice(0, props.limit) : props.users
})
```

No new props needed. The component just renders what it receives.

### Frontend DashboardView

```ts
import type { UserWithPaymentTotal } from '@/types/user'
import { sortByStrategy, type PlayerSortStrategy } from '@/utils/sort'
import { userService } from '@/services/userService'

const SORT_KEY = 'dashboard-player-sort'

function readSort(): PlayerSortStrategy {
  try { return (localStorage.getItem(SORT_KEY) as PlayerSortStrategy) ?? 'debt-first' } catch { return 'debt-first' }
}

const sortStrategy = ref<PlayerSortStrategy>(readSort())
const paymentRankingUsers = ref<UserWithPaymentTotal[]>([])

function onSortChange(s: PlayerSortStrategy) {
  sortStrategy.value = s
  try { localStorage.setItem(SORT_KEY, s) } catch {}
}

// Extend onMounted parallel calls:
userService.getPaymentRanking().then(r => { paymentRankingUsers.value = r })

// Computed sorted list:
const sortedUsersForLeaderboard = computed<User[]>(() => {
  if (sortStrategy.value === 'debt-first') return paymentRankingUsers.value
  return sortByStrategy(userStore.users, sortStrategy.value)
})

const sortOptions = computed(() => [
  { value: 'debt-first' as PlayerSortStrategy,    label: t('dashboard.sortDebtFirst') },
  { value: 'default' as PlayerSortStrategy,       label: t('dashboard.sortDefault') },
  { value: 'winners-first' as PlayerSortStrategy, label: t('dashboard.sortWinnersFirst') },
])
```

Template — replace existing `<Leaderboard>` call:
```html
<Leaderboard :users="sortedUsersForLeaderboard" :debt-threshold="configStore.debtThreshold" :limit="10" compact />
```

Add sort buttons to leaderboard card header:
```html
<el-button-group size="small">
  <el-button
    v-for="s in sortOptions" :key="s.value"
    :type="sortStrategy === s.value ? 'primary' : 'default'"
    @click="onSortChange(s.value)"
  >{{ s.label }}</el-button>
</el-button-group>
```

### Frontend UsersView (scope expanded)

UsersView currently uses `UserTable`, not `Leaderboard`. Add the same sort control + computed sorted list there. UsersView does not need "Debt First" (no payment ranking needed on Users page) — only `default` and `winners-first` are offered.

> Detailed UsersView integration can be scoped as a follow-up or bundled here — confirm with user.

### i18n

`en.json` (inside `"dashboard"` object):
```json
"sortDebtFirst":    "Debt First",
"sortDefault":      "Default",
"sortWinnersFirst": "Winners First"
```

`vi.json`:
```json
"sortDebtFirst":    "Nợ Trước",
"sortDefault":      "Mặc Định",
"sortWinnersFirst": "Thắng Trước"
```

## Integration Points

- `userStore.users` is reactive — `sortedUsersForLeaderboard` auto-updates when users change.
- `paymentRankingUsers` is only updated on mount. If a new settlement occurs while the dashboard is open, refresh requires page reload or a manual re-fetch trigger.
- Backend route order: `/users/payment-ranking` MUST be registered before `/users/:id` to avoid the Gin router treating "payment-ranking" as an ID.

## Error Handling

- `localStorage` wrapped in try/catch; fallback to `'debt-first'` (the new default).
- `getPaymentRanking()` failure: `paymentRankingUsers` stays empty → "Debt First" shows an empty leaderboard until resolved. Consider showing a loading/error state.
- Invalid localStorage value: falls back to `'debt-first'` via `?? 'debt-first'`.

## Performance Considerations

- `GetPaymentRanking` SQL: LEFT JOIN on two small tables, simple SUM aggregation. < 10ms expected.
- `sortByStrategy` for strategies 2 and 3: O(n log n) on ≤ 50 users — imperceptible.
- Payment ranking is fetched once on mount and cached in a ref; no re-fetch on strategy switch.
