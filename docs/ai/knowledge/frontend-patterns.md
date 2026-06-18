# Frontend Patterns

## State Management (Pinia)

Each feature domain has one store. WC auth state lives entirely in `wcAuthStore.ts`.

- `token` and `user` are persisted to `localStorage` (`wc_token`, `wc_user`)
- Computed getters (`isLoggedIn`, `isAdmin`, `googleLinked`) derive from store state
- Always update localStorage through store actions — never write directly from components

```ts
// Reading auth state
const auth = useWcAuthStore()
if (!auth.googleLinked) router.push('/world-cup/link-google')
```

## API Service Layer

Thin wrappers around axios — no business logic here:

```ts
// wcAuthService.ts
export const wcAuthService = {
  login: (name: string, password: string) =>
    wcApi.post<WcLoginResponse>('/auth/login', { name, password }).then(r => r.data),
  googleLogin: (idToken: string) =>
    wcApi.post<WcLoginResponse>('/auth/google', { id_token: idToken }).then(r => r.data),
  googleLink: (idToken: string) =>
    wcApi.post<{ google_linked: boolean; avatar_url: string | null }>('/auth/google/link', { id_token: idToken }).then(r => r.data),
}
```

Base URL is `VITE_API_BASE_URL` from `.env`.

## Route Guards

All route access control is in `router/index.ts` `beforeEach`. Meta flags drive logic:

```ts
// Route definition
{
  path: '/world-cup/predict',
  meta: { requiresWcAuth: true, requiresGoogleLink: true, requiresWcFeature: true }
}
```

Guard order matters:
1. Check feature flag (redirect to schedule if off)
2. Check auth token (redirect to login if missing/expired)
3. Check admin if needed
4. Check Google link (redirect to link-google page if not linked — skip on that page itself)

## TypeScript Types

All WC-related types live in `src/types/wc.ts`. Keep snake_case for API response types (mirror JSON), camelCase for frontend-only objects.

```ts
// API response — snake_case
interface WcLoginResponse {
  token: string
  user_id: string
  name: string
  avatar_url: string | null
  is_admin: boolean
  google_linked: boolean
}

// Frontend object — camelCase
interface WcAuthUser {
  id: string
  name: string
  isAdmin: boolean
  avatarUrl: string | null
  googleLinked: boolean
}
```

## Google Identity Services (GSI)

The Google GSI library is loaded via `<script>` tag in `index.html`. Types are declared in `src/google-gsi.d.ts`.

```ts
// Trigger Google popup in a component
window.google.accounts.id.prompt((notification) => {
  if (notification.isNotDisplayed() || notification.isSkippedMoment()) {
    // Fall back to button flow
  }
})

// Or use credential response callback
window.google.accounts.id.initialize({
  client_id: import.meta.env.VITE_GOOGLE_CLIENT_ID,
  callback: handleCredentialResponse,
})
```

The `id_token` from the credential response is sent directly to the backend for verification.

## Component Conventions

- WC views: `src/views/WcXxxView.vue` — page-level components routed by Vue Router
- WC components: `src/components/wc/WcXxx.vue` — reusable within WC feature
- Use Element Plus (`el-*`) for UI primitives; Tailwind for layout/spacing
- `<script setup lang="ts">` syntax only (no Options API)

## Locale / i18n

Multi-language support via `src/locales/`. WC UI strings should use the i18n plugin, not hardcoded text. Vietnamese is the primary locale (`vi`).
