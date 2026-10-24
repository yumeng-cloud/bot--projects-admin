import { message } from 'antd'

/** 默认走同源 `/api/v1`（Vite 代理→127.0.0.1:8080），避免 127.0.0.1↔localhost CORS；可用 VITE_API_BASE(_URL) 覆盖直连 */
const RAW_BASE =
  (import.meta.env.VITE_API_BASE_URL as string | undefined) ||
  (import.meta.env.VITE_API_BASE as string | undefined) ||
  '/api/v1'

export const API_BASE = RAW_BASE.replace(/\/$/, '')

export type ApiErrorBody = { code: string; message: string }

export class ApiError extends Error {
  code: string
  status: number
  constructor(message: string, code: string, status: number) {
    super(message)
    this.name = 'ApiError'
    this.code = code
    this.status = status
  }
}

/** Issue 乐观锁：后端现为 conflict，Issue 域将对齐 version_conflict；409 均认 */
export function isVersionConflict(err: unknown): boolean {
  if (err instanceof ApiError) {
    if (err.code === 'version_conflict') return true
    return (
      err.status === 409 &&
      err.code === 'conflict' &&
      /乐观锁|已被他人更新|version/i.test(err.message)
    )
  }
  const e = err as { code?: string; status?: number; message?: string } | null
  if (!e) return false
  if (e.code === 'version_conflict') return true
  return (
    e.status === 409 &&
    e.code === 'conflict' &&
    /乐观锁|已被他人更新|version/i.test(e.message || '')
  )
}

export type ApiRequestInit = RequestInit & {
  /** 为 true 时不弹 Toast（页面自行处理，如登录表单） */
  skipToast?: boolean
  /** 写操作幂等键 */
  idempotencyKey?: string
}

const TOKEN_KEY = 'token'

export function getToken(): string | null {
  return localStorage.getItem(TOKEN_KEY)
}

export function setToken(token: string | null) {
  if (token) localStorage.setItem(TOKEN_KEY, token)
  else localStorage.removeItem(TOKEN_KEY)
}

/** 兼容 `/auth/login` 与历史 `/api/v1/auth/login` */
export function resolveUrl(path: string): string {
  if (/^https?:\/\//i.test(path)) return path
  let p = path.startsWith('/') ? path : `/${path}`
  if (API_BASE.endsWith('/api/v1') && p.startsWith('/api/v1')) {
    p = p.slice('/api/v1'.length) || '/'
  }
  if (!API_BASE) return p
  return `${API_BASE}${p}`
}

function parseError(data: unknown, status: number, statusText: string): ApiErrorBody {
  if (data && typeof data === 'object') {
    const o = data as Record<string, unknown>
    // 顶层 {code,message}（HandleError / writeErr）
    if (typeof o.message === 'string') {
      return {
        code: typeof o.code === 'string' ? o.code : 'http',
        message: o.message,
      }
    }
    // 嵌套 {error:{code,message}}（旧 stub / 文档写法）
    const nested = o.error
    if (nested && typeof nested === 'object') {
      const e = nested as Record<string, unknown>
      if (typeof e.message === 'string') {
        return {
          code: typeof e.code === 'string' ? e.code : 'http',
          message: e.message,
        }
      }
    }
  }
  return { code: 'http', message: statusText || `HTTP ${status}` }
}

function handleUnauthorized() {
  setToken(null)
  if (typeof window !== 'undefined' && !window.location.pathname.includes('login')) {
    // App 以 token 门禁；整页刷新回登录
    window.location.assign('/')
  }
}

export async function api<T>(path: string, init: ApiRequestInit = {}): Promise<T> {
  const { skipToast, idempotencyKey, ...req } = init
  const headers = new Headers(req.headers)
  if (!headers.has('Content-Type') && req.body) {
    headers.set('Content-Type', 'application/json')
  }
  const token = getToken()
  if (token && !headers.has('Authorization')) {
    headers.set('Authorization', `Bearer ${token}`)
  }
  if (idempotencyKey) {
    headers.set('Idempotency-Key', idempotencyKey)
  }

  const res = await fetch(resolveUrl(path), {
    ...req,
    headers,
    credentials: 'include',
  })

  const text = await res.text()
  let data: unknown = {}
  if (text) {
    try {
      data = JSON.parse(text)
    } catch {
      data = { message: text }
    }
  }

  if (!res.ok) {
    const body = parseError(data, res.status, res.statusText)
    // Issue 乐观锁：后端暂用 conflict，对齐后为 version_conflict；仅乐观锁文案归一，勿吞 Toast
    let code = body.code
    if (
      res.status === 409 &&
      (code === 'version_conflict' ||
        (code === 'conflict' && /乐观锁|已被他人更新|version/i.test(body.message)))
    ) {
      code = 'version_conflict'
    }
    const err = new ApiError(body.message, code, res.status)
    if (res.status === 401) {
      if (!skipToast) message.error(body.message || '请重新登录')
      handleUnauthorized()
    } else if (!skipToast) {
      message.error(body.message || '请求失败')
    }
    throw err
  }

  return data as T
}
