---
phase: planning
title: Planning – Player Personalization & Dynamic Global Theme
description: Task breakdown, dependencies, and implementation order
---

# Project Planning & Task Breakdown

## Milestones

- [x] **M1 — Avatar upload working end-to-end** (backend + frontend)
- [x] **M2 — Club selection + leaderboard display** (avatar shown in tables)
- [x] **M3 — Dynamic global theme live** (rank #1 club drives CSS vars)

## Task Breakdown

### Phase 1: Backend foundation

- [x] **1.1** Add `avatar_url *string` and `favorite_club *string` to `model/user.go`
- [x] **1.2** Update `database.go` AutoMigrate (GORM will add columns automatically)
- [x] **1.3** Add `UpdateAvatarURL`, `ClearAvatarURL`, `UpdateFavoriteClub` to `user_repository.go`
- [x] **1.4** Add `UploadAvatar(userID uuid, file multipart.File, header *multipart.FileHeader)` to `user_service.go`
  - Validate MIME type (magic bytes), max 2 MB
  - Generate UUID filename, write to `./uploads/avatars/`
  - Delete previous avatar file if exists
  - Return relative URL `/uploads/avatars/{uuid}.{ext}`
- [x] **1.5** Add `DeleteAvatar(userID uuid)` to `user_service.go`
- [x] **1.6** Add `UpdateClub(userID uuid, club string)` to `user_service.go` with slug validation
- [x] **1.7** Add `UploadAvatar`, `DeleteAvatar`, `UpdateClub` handlers to `user_handler.go`
- [x] **1.8** Register routes in `router.go`:
  - `PUT /api/v1/users/:id/avatar`
  - `DELETE /api/v1/users/:id/avatar`
  - `PUT /api/v1/users/:id/club`
  - `router.Static("/uploads", "./uploads")`
- [x] **1.9** Extend leaderboard and `GetAll` responses to include `avatar_url` and `favorite_club`

### Phase 2: Frontend — avatar upload & club picker

- [x] **2.1** Create `src/config/clubs.ts` with `CLUBS` array and `ClubTheme` interface
- [x] **2.2** Create `src/components/shared/UserAvatar.vue`
  - Props: `avatarUrl?: string`, `name: string`, `size?: 'sm'|'md'|'lg'`
  - Shows image if `avatarUrl` is set, otherwise renders initials circle
  - Supports GIF (just `<img>` with no special handling needed)
- [x] **2.3** Add `uploadAvatar(userId, file)` and `updateClub(userId, club)` to `userService.ts`
- [x] **2.4** Extend `UserTable.vue` existing edit modal (`UserEditModal` or equivalent):
  - Avatar upload widget: click-to-upload `<input type="file">`, preview before save
  - Club picker: `<el-select>` populated from `CLUBS` config
  - No new route/view needed
- [x] **2.5** Add `UserAvatar` to `UserTable.vue` rows (show avatar + name in same cell)
- [x] **2.6** Add `UserAvatar` to leaderboard component rows
- [x] **2.7** Add `UserAvatar` to recent matches list (home & away player names)

### Phase 3: Dynamic global theme

- [x] **3.1** Create `src/composables/useGlobalTheme.ts`
  - Import `CLUBS` config
  - Watch `leaderboard[0]?.favorite_club`
  - On change: find matching `ClubTheme`, write `--theme-primary`, `--theme-accent`, `--theme-text-on-primary` to `document.documentElement.style`
  - Fallback to default green if no club
- [x] **3.2** Wire `useGlobalTheme` in `App.vue` (call once, runs for the session)
- [x] **3.3** Update global CSS to use `var(--theme-primary)` and `var(--theme-accent)` in key places:
  - Primary buttons
  - Active nav item
  - Leaderboard rank #1 highlight
  - Tier badge accents (optional)
- [x] **3.4** Add a subtle "🏆 Rank #1: {name} — {club}" banner or tooltip so players can see whose theme is active

## Dependencies

- 1.1 → 1.3 → 1.4 → 1.7 → 1.8 (strict chain)
- 1.9 → 2.5, 2.6, 2.7 (need `avatar_url` in API response before rendering)
- 2.1 → 2.2 → 2.5, 2.6 (need `UserAvatar` component before using in tables)
- 2.1 → 3.1 (need `CLUBS` config)
- 3.1 → 3.2 → 3.3 (theme composable must exist before wiring in App.vue)

## Timeline & Estimates

| Phase | Effort |
|---|---|
| Phase 1 (backend) | ~3–4h |
| Phase 2 (avatar + club UI) | ~3–4h |
| Phase 3 (dynamic theme) | ~2h |
| **Total** | **~8–10h** |

## Risks & Mitigation

| Risk | Mitigation |
|---|---|
| GIF files can be large (>2 MB) | Enforce 2 MB limit server-side; show clear error message |
| MIME spoofing (rename .exe to .gif) | Read first 512 bytes and check magic bytes with `http.DetectContentType` |
| Disk fills up if old avatars not deleted | Delete previous file on new upload; add `DELETE /avatar` for admin cleanup |
| Rank #1 changes frequently — theme flickers | Only update CSS vars when club *changes*, not on every leaderboard poll |
| SVG upload XSS | Explicitly reject `image/svg+xml` in MIME check |
