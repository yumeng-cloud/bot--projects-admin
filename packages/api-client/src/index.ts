/** @acb/api-client — 鉴权头 / 错误 / 类型 / session-context（禁 UI / 布局 / DnD） */

export {
  api,
  ApiError,
  getApiBase,
  getToken,
  setToken,
  resolveUrl,
  type ApiRequestInit,
} from './client'

export { login, register, logout, me } from './auth'

export {
  getSessionContext,
  putSessionContext,
  emptySessionContext,
  hasSessionScene,
  type SessionContext,
  type SessionContextPut,
} from './sessionContext'

export type { User, AuthResp, MeResp, ApiErrorBody } from './types'
