/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_SITE?: string
  readonly VITE_SITE_TITLE?: string
  readonly VITE_API_BASE_URL?: string
  readonly VITE_GOOGLE_CLIENT_ID?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
