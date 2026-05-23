/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_API_BASE_URL: string
  readonly VITE_ARMS_ENABLED?: 'true' | 'false'
  readonly VITE_ARMS_ENDPOINT?: string
  readonly VITE_ARMS_ENV?: 'prod' | 'gray' | 'pre' | 'daily' | 'local'
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
