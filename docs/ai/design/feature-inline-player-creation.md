---
phase: design
title: System Design & Architecture
description: Inline player creation from MatchForm and CreateTournamentView
feature: inline-player-creation
---

# System Design & Architecture

## Architecture Overview

```mermaid
graph TD
  MatchForm -->|mounts| UserForm
  CreateTournamentView -->|mounts| UserForm
  UserForm -->|submit(name, tier, handicap_rate)| MatchForm
  UserForm -->|submit(name, tier, handicap_rate)| CreateTournamentView
  MatchForm -->|dispatch| useUserStore.createUser
  CreateTournamentView -->|dispatch| useUserStore.createUser
  useUserStore.createUser -->|POST /api/v1/users| Backend
  MatchForm -->|request refresh| Parent View
  Parent View -->|passes updated users| MatchForm
  CreateTournamentView -->|refresh users list| useUserStore.users
```

**Approach: Option A — Quick-Add trong dropdown/selector**

- `MatchForm`: Thêm nút "➕ Tạo player mới" dưới mỗi `el-select` team (dùng slot `#footer` của `el-select`)
- `CreateTournamentView`: Thêm nút "➕ Tạo player mới" bên dưới `el-checkbox-group` player list
- Cả hai context đều mở `UserForm` component (reuse nguyên vẹn)
- Sau khi tạo thành công, player mới phải xuất hiện ngay trong danh sách player của context hiện tại; không auto-select hoặc auto-check mặc định

## Data Models

Không thay đổi. Reuse:
- `User` type từ `@/types/user`
- `useUserStore().createUser(name, tier, handicapRate)` action đã có
- `useUserStore().fetchUsers()` để đồng bộ list

## API Design

Không có API mới. Endpoint hiện có:
```
POST /api/v1/users
Body: { name: string, tier: string, handicap_rate: number }
Response: User
```

Frontend store contract giữ nguyên theo codebase hiện tại:
```ts
await userStore.createUser(name, tier, handicapRate)
```

Không refactor sang object payload trong feature này để tránh mở rộng scope không cần thiết.

## Component Breakdown

### Thay đổi ở `MatchForm.vue`
- Import `UserForm`
- Thêm state: `showQuickCreatePlayer: boolean`, `quickCreateTarget: 'team1' | 'team2' | null`
- Thêm slot `#footer` trong cả 2 `el-select` → nút "➕ Tạo player mới"
- Handler `handleQuickCreatePlayer(target)`: set target + open dialog
- `MatchForm` tiếp tục nhận `users` từ parent, không chuyển sang store-driven hoàn toàn
- Handler `handlePlayerCreated(...)`: gọi `userStore.createUser(...)`, sau đó emit event yêu cầu parent refresh lại danh sách `users` để player mới hiển thị trong form
- `UserForm` dialog v-model: `showQuickCreatePlayer`
- Nếu `el-select #footer` không hoạt động ổn với version Element Plus hiện tại, fallback sang nút nhỏ cạnh label Team 1 / Team 2

### Thay đổi ở `CreateTournamentView.vue`
- Import `UserForm`
- Thêm state: `showQuickCreatePlayer: boolean`
- Thêm nút "➕ Tạo player mới" bên dưới checkbox group
- Handler `handlePlayerCreated(...)`: gọi `userStore.createUser(...)`, sau đó refresh list để player mới xuất hiện trong checkbox group
- Player mới không auto-check mặc định
- `UserForm` dialog v-model: `showQuickCreatePlayer`

### `UserForm.vue` — Không thay đổi
Reuse 100%, emit `submit` như hiện tại.

## Design Decisions

| Decision | Rationale |
|---|---|
| Reuse `UserForm` thay vì mini inline form | Đảm bảo data quality (tier, handicap), không duplicate validation logic |
| `el-select` slot `#footer` thay vì option đặc biệt | Tránh bị include vào filter text, không gây lẫn lộn với player options |
| Không auto-select hoặc auto-check mặc định | Giữ quyền kiểm soát cho user và tránh side effect bất ngờ |
| `MatchForm` vẫn nhận `users` từ parent | Giữ boundary hiện tại của component, tránh trộn props-driven với store-driven state |
| Giữ `userStore.createUser(name, tier, handicapRate)` như hiện tại | Tránh refactor ngoài scope của feature |
| Không tạo wrapper component mới | Feature đơn giản, không cần abstraction thêm |
| Fallback sang nút cạnh label nếu `#footer` không ổn | Giảm rủi ro phụ thuộc vào behavior của Element Plus |

## Non-Functional Requirements

- **Performance:** `fetchUsers()` đã có trong store, cost thêm 1 API call khi tạo player — chấp nhận được
- **UX:** Dialog `UserForm` xuất hiện đè lên, có thể close để quay lại form gốc
- **Accessibility:** Nút "➕ Tạo player mới" có aria-label rõ ràng
