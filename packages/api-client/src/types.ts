/** 共享鉴权 / 用户类型（仅类型，不含 UI） */

export type User = {
  id: string
  email: string
  name: string
  created_at?: string
  /** 平台超管旗标 */
  is_platform_admin?: boolean
  avatar_url?: string | null
}

export type AuthResp = { token: string; user: User }

export type MeResp = { user: User }

export type ApiErrorBody = { code: string; message: string }
