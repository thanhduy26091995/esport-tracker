---
phase: design
title: Design – Player Personalization & Dynamic Global Theme
description: Architecture, data models, API, and component breakdown
---

# System Design & Architecture

## Architecture Overview

```mermaid
graph TD
  FE[Frontend Vue App] -->|PUT /users/:id/avatar multipart| API
  FE -->|PUT /users/:id/club| API
  FE -->|GET /users/leaderboard| API
  API --> UserService
  UserService --> UserRepo
  UserRepo --> DB[(PostgreSQL users table)]
  UserService -->|write/delete file| FS[(Local filesystem ./uploads/avatars/)]
  FE -->|GET /uploads/avatars/:file| StaticServe[Gin Static Route]
  StaticServe --> FS
  FE --> ThemeEngine[Vue useGlobalTheme composable]
  ThemeEngine -->|watch leaderboard[0].favorite_club| CSSVars[document.documentElement.style]
  FE -->|click player row| Modal[UserEditModal]
  Modal -->|upload / club change| FE
```

**Key components:**
- **Upload endpoint** — `PUT /api/v1/users/:id/avatar` — multipart/form-data, returns `avatar_url`.
- **Club endpoint** — `PUT /api/v1/users/:id/club` — JSON body `{ "favorite_club": "real-madrid" }`.
- **Static file route** — Gin serves `/uploads/avatars/` as a static directory.
- **ThemeEngine composable** — watches `leaderboard[0].favorite_club`, maps to CSS vars, writes to `document.documentElement.style`.

## Data Models

### `users` table — new columns

| Column | Type | Nullable | Notes |
|---|---|---|---|
| `avatar_url` | `varchar(255)` | YES | Relative path: `/uploads/avatars/{uuid}.{ext}` |
| `favorite_club` | `varchar(50)` | YES | Slug e.g. `real-madrid`, `man-city` |

Go model addition:
```go
AvatarURL     *string `gorm:"type:varchar(255)" json:"avatar_url,omitempty"`
FavoriteClub  *string `gorm:"type:varchar(50)"  json:"favorite_club,omitempty"`
```

GORM AutoMigrate will add these columns automatically.

### Club theme config (frontend-only, no DB)

Each club gets a full identity system — not just 2 colours. The CSS variable set drives buttons, glows, gradients, and background tints so the whole UI feels branded.

```ts
// src/config/clubs.ts
export interface ClubTheme {
  slug: string
  name: string
  primary: string      // main brand color — buttons, badges, active nav
  secondary: string    // second brand color — used in gradients & stripes
  accent: string       // highlight color — rank #1 badge, hover states
  bg: string           // subtle page background tint (very low opacity)
  gradient: string     // CSS linear-gradient — used in hero banners & headers
  glow: string         // rgba — box-shadow glow for cards & focused elements
  text: 'light' | 'dark'  // text color on top of primary background
}

export const CLUBS: ClubTheme[] = [
  {
    slug: 'real-madrid', name: 'Real Madrid',
    primary: '#FEBE10', secondary: '#00529F', accent: '#ffffff',
    bg: 'rgba(254,190,16,0.06)',
    gradient: 'linear-gradient(135deg, #00529F 0%, #002c6e 50%, #FEBE10 100%)',
    glow: 'rgba(254,190,16,0.45)',
    text: 'dark',
  },
  {
    slug: 'barcelona', name: 'Barcelona',
    primary: '#A50044', secondary: '#004D98', accent: '#EDBB00',
    bg: 'rgba(165,0,68,0.06)',
    gradient: 'linear-gradient(135deg, #A50044 0%, #004D98 50%, #A50044 100%)',
    glow: 'rgba(165,0,68,0.45)',
    text: 'light',
  },
  {
    slug: 'man-city', name: 'Manchester City',
    primary: '#6CABDD', secondary: '#1C2C5B', accent: '#ffffff',
    bg: 'rgba(108,171,221,0.07)',
    gradient: 'linear-gradient(135deg, #1C2C5B 0%, #6CABDD 100%)',
    glow: 'rgba(108,171,221,0.5)',
    text: 'dark',
  },
  {
    slug: 'liverpool', name: 'Liverpool',
    primary: '#C8102E', secondary: '#00B2A9', accent: '#F6EB61',
    bg: 'rgba(200,16,46,0.06)',
    gradient: 'linear-gradient(135deg, #8B0000 0%, #C8102E 60%, #F6EB61 100%)',
    glow: 'rgba(200,16,46,0.45)',
    text: 'light',
  },
  {
    slug: 'man-utd', name: 'Manchester Utd',
    primary: '#DA291C', secondary: '#FBE122', accent: '#ffffff',
    bg: 'rgba(218,41,28,0.06)',
    gradient: 'linear-gradient(135deg, #7a0a05 0%, #DA291C 70%, #FBE122 100%)',
    glow: 'rgba(218,41,28,0.45)',
    text: 'light',
  },
  {
    slug: 'chelsea', name: 'Chelsea',
    primary: '#034694', secondary: '#DBA111', accent: '#ffffff',
    bg: 'rgba(3,70,148,0.07)',
    gradient: 'linear-gradient(135deg, #011f45 0%, #034694 65%, #DBA111 100%)',
    glow: 'rgba(219,161,17,0.45)',
    text: 'light',
  },
  {
    slug: 'arsenal', name: 'Arsenal',
    primary: '#EF0107', secondary: '#063672', accent: '#ffffff',
    bg: 'rgba(239,1,7,0.06)',
    gradient: 'linear-gradient(135deg, #8B0000 0%, #EF0107 100%)',
    glow: 'rgba(239,1,7,0.4)',
    text: 'light',
  },
  {
    slug: 'psg', name: 'PSG',
    primary: '#003370', secondary: '#ED1C24', accent: '#C6A84B',
    bg: 'rgba(0,51,112,0.08)',
    gradient: 'linear-gradient(135deg, #001529 0%, #003370 55%, #C6A84B 100%)',
    glow: 'rgba(198,168,75,0.45)',
    text: 'light',
  },
  {
    slug: 'bayern', name: 'Bayern München',
    primary: '#DC052D', secondary: '#0066B2', accent: '#ffffff',
    bg: 'rgba(220,5,45,0.06)',
    gradient: 'linear-gradient(135deg, #7a0010 0%, #DC052D 100%)',
    glow: 'rgba(220,5,45,0.45)',
    text: 'light',
  },
  {
    slug: 'juventus', name: 'Juventus',
    primary: '#111111', secondary: '#ffffff', accent: '#aaaaaa',
    bg: 'rgba(0,0,0,0.08)',
    gradient: 'linear-gradient(135deg, #000000 0%, #1a1a1a 50%, #333333 100%)',
    glow: 'rgba(200,200,200,0.3)',
    text: 'light',
  },
  {
    slug: 'atletico', name: 'Atlético Madrid',
    primary: '#CE3524', secondary: '#272E61', accent: '#ffffff',
    bg: 'rgba(206,53,36,0.06)',
    gradient: 'linear-gradient(135deg, #272E61 0%, #CE3524 100%)',
    glow: 'rgba(206,53,36,0.4)',
    text: 'light',
  },
  {
    slug: 'dortmund', name: 'Borussia Dortmund',
    primary: '#FDE100', secondary: '#1a1a1a', accent: '#000000',
    bg: 'rgba(253,225,0,0.08)',
    gradient: 'linear-gradient(135deg, #1a1a1a 0%, #3d3400 50%, #FDE100 100%)',
    glow: 'rgba(253,225,0,0.5)',
    text: 'dark',
  },
  {
    slug: 'inter', name: 'Inter Milan',
    primary: '#010E80', secondary: '#000000', accent: '#4c87c8',
    bg: 'rgba(1,14,128,0.08)',
    gradient: 'linear-gradient(135deg, #000000 0%, #010E80 100%)',
    glow: 'rgba(76,135,200,0.4)',
    text: 'light',
  },
  {
    slug: 'ac-milan', name: 'AC Milan',
    primary: '#FB090B', secondary: '#000000', accent: '#ffffff',
    bg: 'rgba(251,9,11,0.06)',
    gradient: 'linear-gradient(135deg, #000000 0%, #3d0000 50%, #FB090B 100%)',
    glow: 'rgba(251,9,11,0.45)',
    text: 'light',
  },
  {
    slug: 'napoli', name: 'Napoli',
    primary: '#12A0C3', secondary: '#ffffff', accent: '#005f7a',
    bg: 'rgba(18,160,195,0.07)',
    gradient: 'linear-gradient(135deg, #005a70 0%, #12A0C3 100%)',
    glow: 'rgba(18,160,195,0.45)',
    text: 'dark',
  },
  {
    slug: 'porto', name: 'Porto',
    primary: '#003087', secondary: '#ffffff', accent: '#C8A951',
    bg: 'rgba(0,48,135,0.07)',
    gradient: 'linear-gradient(135deg, #001040 0%, #003087 70%, #C8A951 100%)',
    glow: 'rgba(200,169,81,0.4)',
    text: 'light',
  },
  {
    slug: 'benfica', name: 'Benfica',
    primary: '#CC0000', secondary: '#ffffff', accent: '#ffcccc',
    bg: 'rgba(204,0,0,0.06)',
    gradient: 'linear-gradient(135deg, #6b0000 0%, #CC0000 100%)',
    glow: 'rgba(204,0,0,0.4)',
    text: 'light',
  },
  {
    slug: 'ajax', name: 'Ajax',
    primary: '#CC0000', secondary: '#ffffff', accent: '#ff6666',
    bg: 'rgba(204,0,0,0.06)',
    gradient: 'linear-gradient(135deg, #CC0000 0%, #ffffff 50%, #CC0000 100%)',
    glow: 'rgba(204,0,0,0.4)',
    text: 'light',
  },
  {
    slug: 'flamengo', name: 'Flamengo',
    primary: '#CC0000', secondary: '#111111', accent: '#FF6600',
    bg: 'rgba(204,0,0,0.06)',
    gradient: 'linear-gradient(135deg, #000000 0%, #CC0000 60%, #FF6600 100%)',
    glow: 'rgba(255,102,0,0.45)',
    text: 'light',
  },
  {
    slug: 'none', name: '— Không chọn —',
    primary: '#16a34a', secondary: '#14532d', accent: '#4ade80',
    bg: 'rgba(22,163,74,0.06)',
    gradient: 'linear-gradient(135deg, #14532d 0%, #16a34a 100%)',
    glow: 'rgba(22,163,74,0.35)',
    text: 'light',
  },
]
```

### CSS variables applied to `<html>`

```css
/* Applied dynamically by useGlobalTheme composable */
--theme-primary        /* main brand color */
--theme-secondary      /* second brand color */
--theme-accent         /* highlight / gold / glow accent */
--theme-bg             /* subtle background tint */
--theme-gradient       /* header / hero gradient */
--theme-glow           /* box-shadow color */
--theme-text-on-primary /* #fff or #000 */
```

Usage examples in component CSS:
```css
.btn-primary        { background: var(--theme-primary); color: var(--theme-text-on-primary); }
.rank-1-card        { box-shadow: 0 0 24px var(--theme-glow); border-color: var(--theme-accent); }
.page-header        { background: var(--theme-gradient); }
.nav-item--active   { color: var(--theme-primary); border-bottom: 2px solid var(--theme-accent); }
body                { background-color: var(--theme-bg); /* very subtle */ }
```

## API Design

### `PUT /api/v1/users/:id/avatar`
- **Auth:** None — open access (private group app, any client can update any player's avatar).
- **Content-Type:** `multipart/form-data`
- **Field:** `avatar` (file)
- **Constraints:** max 2 MB, accepted types: `image/jpeg`, `image/png`, `image/gif`, `image/webp`. SVG rejected.
- **Response:** `{ "avatar_url": "/uploads/avatars/abc123.gif" }`
- **Side-effect:** Deletes previous avatar file from disk before writing new one.

### `DELETE /api/v1/users/:id/avatar`
- **Auth:** None — open access.
- **Response:** `204 No Content`
- **Side-effect:** Deletes file from disk, sets `avatar_url = NULL`.

### `PUT /api/v1/users/:id/club`
- **Auth:** None — open access.
- **Body:** `{ "favorite_club": "real-madrid" }`
- **Validation (backend):** Must be one of the hardcoded valid slugs or empty string. Hardcoded list mirrors `src/config/clubs.ts` slugs. Returns `400` for unknown slugs.
- **Response:** `{ "favorite_club": "real-madrid" }`

### `GET /api/v1/users/leaderboard` (existing — extend response)
- Add `avatar_url` and `favorite_club` to `UserWithStats` response.

## Component Breakdown

### Backend
- `user_handler.go` — add `UploadAvatar`, `DeleteAvatar`, `UpdateClub` handlers.
- `user_service.go` — add `UploadAvatar(userID, file)`, `DeleteAvatar(userID)`, `UpdateClub(userID, club)`.
- `user_repository.go` — add `UpdateAvatarURL(id, url)`, `ClearAvatarURL(id)`, `UpdateFavoriteClub(id, club)`.
- `router.go` — register new routes + `router.Static("/uploads", "./uploads")`.
- `model/user.go` — add `AvatarURL *string`, `FavoriteClub *string`.

### Frontend
- `src/config/clubs.ts` — club list + theme map (new file).
- `src/composables/useGlobalTheme.ts` — watches leaderboard rank #1, writes CSS vars.
- `src/components/shared/UserAvatar.vue` — reusable avatar `<img>` with fallback initials circle.
- `src/components/user/UserEditModal.vue` — existing edit modal (extend): add avatar upload widget + club picker `<el-select>`.
- `src/components/user/UserTable.vue` — existing table (extend): add edit button per row that opens `UserEditModal`, show `UserAvatar` in each row.
- No new route/view needed. All editing happens via modal triggered from `UserTable` or leaderboard.

## Design Decisions

1. **Local filesystem over S3**: Simplest for a self-hosted app. Serve via Gin static route. Easy to migrate to S3 later by swapping `UploadAvatar` implementation.
2. **Club list in frontend config, not DB**: The list is stable and small (~20). DB adds complexity with no benefit. Admin can deploy a code update to add clubs.
3. **GIF animation**: Animate GIFs everywhere for simplicity. If performance degrades on the leaderboard table, fall back to `image-rendering: auto` with a `?thumb` variant later.
4. **Theme via CSS custom properties**: Cleanest approach — no Vue reactivity needed in child components. Single `useGlobalTheme` composable sets vars on `<html>`, all components inherit via `var(--theme-primary)`.
5. **Theme update cadence**: Poll leaderboard every 30s OR react to existing `store.fetchLeaderboard()` calls. No websocket needed for MVP.

## Non-Functional Requirements

- **File storage path:** `./uploads/avatars/{uuid}.{ext}` relative to binary working directory.
- **Max upload size:** 2 MB enforced at Gin middleware level (`router.MaxMultipartMemory = 2 << 20`).
- **Security:** Validate MIME type server-side (read magic bytes, not just extension). Reject SVG (XSS risk). Store by UUID, not original filename (path traversal prevention).
- **Theme fallback:** If rank #1 has no club or club is `none`, revert to default green theme (`#16a34a`).
