---
phase: implementation
title: Implementation Guide – Player Personalization & Dynamic Global Theme
description: Technical implementation notes, patterns, and integration details
---

# Implementation Guide

## Development Setup

No new dependencies needed for MVP:
- Go: `multipart` is stdlib; no new packages.
- Frontend: no new npm packages needed (just `<img>` and `el-select`).
- Create uploads directory: `mkdir -p ./uploads/avatars` (or do it in service on startup).

## Code Structure

```
backend/
  internal/
    model/user.go           ← add AvatarURL, FavoriteClub fields
    repository/user_repository.go  ← add 3 new repo methods
    service/user_service.go  ← add UploadAvatar, DeleteAvatar, UpdateClub
    api/user_handler.go     ← add 3 new handlers
    api/router.go           ← register routes + static serve

frontend/
  src/
    config/clubs.ts         ← NEW: CLUBS array + ClubTheme interface
    composables/useGlobalTheme.ts  ← NEW: watches leaderboard, writes CSS vars
    components/shared/UserAvatar.vue  ← NEW: reusable avatar component
    services/userService.ts  ← extend: uploadAvatar, updateClub
    stores/userStore.ts      ← extend if needed for profile state
    views/ProfileView.vue    ← NEW or extend: avatar upload + club picker

uploads/
  avatars/                  ← served statically by Gin
```

## Implementation Notes

### Avatar upload (backend)

```go
// user_service.go
func (s *UserService) UploadAvatar(userID uuid.UUID, file multipart.File, header *multipart.FileHeader) (string, error) {
    // 1. Read first 512 bytes to detect MIME
    buf := make([]byte, 512)
    n, _ := file.Read(buf)
    mimeType := http.DetectContentType(buf[:n])
    allowed := map[string]string{
        "image/jpeg": "jpg", "image/png": "png",
        "image/gif": "gif", "image/webp": "webp",
    }
    ext, ok := allowed[mimeType]
    if !ok {
        return "", fmt.Errorf("unsupported file type: %s", mimeType)
    }
    // 2. Seek back to start
    file.Seek(0, io.SeekStart)
    // 3. Generate filename, write to disk
    filename := uuid.New().String() + "." + ext
    dst := filepath.Join("uploads", "avatars", filename)
    out, err := os.Create(dst)
    // ... copy file, handle error
    // 4. Delete old avatar
    user, _ := s.repo.GetByID(userID)
    if user.AvatarURL != nil {
        oldPath := "." + *user.AvatarURL  // strip leading /
        os.Remove(oldPath)
    }
    // 5. Persist URL
    avatarURL := "/uploads/avatars/" + filename
    return avatarURL, s.repo.UpdateAvatarURL(userID, avatarURL)
}
```

### Gin static + max memory

```go
// router.go
router.MaxMultipartMemory = 2 << 20  // 2 MB
router.Static("/uploads", "./uploads")
```

### Club slug validation (backend)

```go
var validClubs = map[string]bool{
    "real-madrid": true, "barcelona": true, "man-city": true,
    "liverpool": true, "man-utd": true, "chelsea": true,
    "arsenal": true, "psg": true, "bayern": true, "juventus": true,
    "atletico": true, "dortmund": true, "inter": true, "ac-milan": true,
    "napoli": true, "porto": true, "benfica": true, "ajax": true,
    "flamengo": true, "none": true, "": true,
}
```

### UserAvatar component (frontend)

```vue
<!-- src/components/shared/UserAvatar.vue -->
<template>
  <div class="user-avatar" :class="`user-avatar--${size}`">
    <img v-if="avatarUrl" :src="avatarUrl" :alt="name" />
    <span v-else class="user-avatar__initials">{{ initials }}</span>
  </div>
</template>
<script setup lang="ts">
const props = withDefaults(defineProps<{
  avatarUrl?: string | null
  name: string
  size?: 'sm' | 'md' | 'lg'
}>(), { size: 'md' })
const initials = computed(() => props.name.slice(0, 2).toUpperCase())
</script>
```

### useGlobalTheme composable

```ts
// src/composables/useGlobalTheme.ts
import { watch, type Ref } from 'vue'
import { CLUBS } from '@/config/clubs'

export function useGlobalTheme(leaderboard: Ref<{ favorite_club?: string }[]>) {
  let lastClub = ''
  watch(
    () => leaderboard.value[0]?.favorite_club,
    (club) => {
      if (club === lastClub) return
      lastClub = club ?? ''
      const theme = CLUBS.find(c => c.slug === club) ?? CLUBS.find(c => c.slug === 'none')!
      const root = document.documentElement.style
      root.setProperty('--theme-primary',          theme.primary)
      root.setProperty('--theme-secondary',         theme.secondary)
      root.setProperty('--theme-accent',            theme.accent)
      root.setProperty('--theme-bg',                theme.bg)
      root.setProperty('--theme-gradient',          theme.gradient)
      root.setProperty('--theme-glow',              theme.glow)
      root.setProperty('--theme-text-on-primary',   theme.text === 'dark' ? '#000' : '#fff')
    },
    { immediate: true }
  )
}
```

CSS transition for smooth club switches (add to `App.vue` or `main.css`):
```css
html {
  transition:
    background-color 0.6s ease,
    --theme-primary 0.4s ease;
}
/* Most browsers don't interpolate custom props natively — use a wrapper element transition instead */
.theme-transition-layer {
  transition: background 0.5s ease, box-shadow 0.5s ease;
}
```

## Integration Points

- `GET /api/v1/users/leaderboard` must return `avatar_url` and `favorite_club` — check `UserWithStats` includes them.
- `App.vue` calls `useGlobalTheme(store.leaderboard)` — leaderboard must be fetched on mount.
- Existing CSS variables (`--surface-card`, `--text-primary`, etc.) stay unchanged. New theme vars are additive.

## Error Handling

- Upload too large → Gin returns 413 before handler is called (due to `MaxMultipartMemory`).
- Invalid MIME → handler returns `400 Bad Request` with message.
- File write failure → return `500 Internal Server Error`, do not persist URL.
- Unknown club slug → return `400 Bad Request`.

## Security Notes

- **Path traversal:** Use `filepath.Join` with a fixed base dir; never use user-supplied filenames.
- **SVG XSS:** `http.DetectContentType` returns `text/plain` or `text/xml` for SVG — both rejected by the allowlist.
- **Auth:** Avatar upload/club update must check `userID == tokenUserID || isAdmin`. Add this check to handler.
- **Serving uploaded files:** Gin static route serves the directory as-is. No code execution risk as long as you only accept image MIME types.
