/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_API_BASE_URL?: string
  /** @deprecated 使用 VITE_API_BASE_URL */
  readonly VITE_API_BASE?: string
  readonly VITE_USE_MOCK?: string
  /** 写入入场链接/Bot 提示词的公开 API 根；默认 https://bot.api.sn.cn */
  readonly VITE_PUBLIC_API_BASE?: string
}

interface ImportMeta {
  readonly env: ImportMetaEnv
}
