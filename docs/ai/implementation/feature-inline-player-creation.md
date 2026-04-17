---
phase: implementation
title: Implementation Guide
description: Implementation notes for inline player creation
feature: inline-player-creation
---

# Implementation Guide

## Code Structure

Files cần chỉnh sửa:
- `frontend/src/components/match/MatchForm.vue`
- `frontend/src/views/CreateTournamentView.vue`
- `frontend/src/locales/vi.json`
- `frontend/src/locales/en.json`

Files reuse không đổi:
- `frontend/src/components/user/UserForm.vue`
- `frontend/src/stores/userStore.ts`

## Implementation Notes

### MatchForm.vue — Pattern

```vue
<!-- Thêm vào el-select team1 và team2 -->
<el-select ...>
  <el-option ... />
  <template #footer>
    <el-button
      text
      type="primary"
      size="small"
      class="w-full"
      @click="handleQuickCreate('team1')"
    >
      ➕ {{ t('players.quickCreate') }}
    </el-button>
  </template>
</el-select>

<!-- UserForm dialog (thêm 1 lần ở cuối template) -->
<UserForm
  v-model="showQuickCreatePlayer"
  :loading="quickCreateLoading"
  @submit="handlePlayerCreated"
  @cancel="showQuickCreatePlayer = false"
/>
```

```ts
// Script additions
import UserForm from '@/components/user/UserForm.vue'

const showQuickCreatePlayer = ref(false)
const quickCreateTarget = ref<'team1' | 'team2' | null>(null)
const quickCreateLoading = ref(false)

function handleQuickCreate(target: 'team1' | 'team2') {
  quickCreateTarget.value = target
  showQuickCreatePlayer.value = true
}

async function handlePlayerCreated(data: { name: string; tier: string; handicap_rate: number }) {
  quickCreateLoading.value = true
  try {
    await userStore.createUser(data.name, data.tier, data.handicap_rate)
    showQuickCreatePlayer.value = false
    emit('request-users-refresh')
  } finally {
    quickCreateLoading.value = false
  }
}
```

`MatchForm` tiếp tục nhận `users` từ parent. Sau khi tạo player thành công, component chỉ cần phát event để parent refresh source data; player mới sẽ xuất hiện lại trong form sau khi prop `users` cập nhật. Nếu slot `#footer` không khả dụng, đặt nút quick-create cạnh label Team 1 / Team 2.

### CreateTournamentView.vue — Pattern

```vue
<!-- Bên dưới el-checkbox-group -->
<el-button
  text
  type="primary"
  size="small"
  @click="showQuickCreatePlayer = true"
>
  ➕ {{ t('players.quickCreate') }}
</el-button>

<UserForm
  v-model="showQuickCreatePlayer"
  :loading="quickCreateLoading"
  @submit="handlePlayerCreated"
  @cancel="showQuickCreatePlayer = false"
/>
```

```ts
async function handlePlayerCreated(data: { name: string; tier: string; handicap_rate: number }) {
  quickCreateLoading.value = true
  try {
    await userStore.createUser(data.name, data.tier, data.handicap_rate)
    await userStore.fetchUsers()
    showQuickCreatePlayer.value = false
  } finally {
    quickCreateLoading.value = false
  }
}
```

Player mới chỉ cần xuất hiện trong list sau khi refresh; không auto-check mặc định.

### I18n Keys

```json
// Thêm vào cả vi.json và en.json
"players": {
  "quickCreate": "Tạo player mới"   // vi
  "quickCreate": "Create new player" // en
}
```

> **Note:** Nếu namespace `players` chưa có, tạo mới. Nếu đã có `users` namespace thì thêm vào đó.

## Integration Points

- `userStore.createUser(name, tier, handicapRate)` tạo player mới theo contract hiện tại của store
- `userStore.fetchUsers()` refresh `userStore.users` reactive array cho `CreateTournamentView`
- `MatchForm` cần parent refresh lại prop `users` sau khi nhận event từ component
- `UserForm` emit `submit` với `{ name, tier, handicap_rate }` — match API payload

## Error Handling

- Nếu `createUser` throw error: `UserForm` có loading prop, `finally` sẽ reset loading; không cần catch thêm ở caller
- Nếu `fetchUsers` fail: form vẫn hoạt động vì `createUser` đã thành công, chỉ không refresh list
