import type { ApiErrorBody } from './types'

/**
 * 共享 HTTP 客户端：鉴权头 / 错误 / URL 解析。
 * 无 UI 依赖（禁 antd toast）；错误由调用方处理。
 */

const TOKEN_KEY = 'token'

export function getApiBase(): string {
  const raw =
    (typeof import.meta !== 'undefined' &&
      (import.meta as ImportMeta & { env?: Record<string, string> }).env &&
      ((import.meta as ImportMeta & { env?: Record<string, string> }).env!.VITE_API_BASE_URL ||
        (import.meta as ImportMeta & { env?: Record<string, string> }).env!.VITE_API_BASE)) ||
    '/api/v1'
  return String(raw).replace(/\/$/, '')
}

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

export type ApiRequestInit = RequestInit & {
  /** 不附带 localStorage JWT */
  skipAuth?: boolean
  /** 写操作幂等键 */
  idempotencyKey?: string
  /** 401 时是否清除 token（默认 true；登录失败等可关） */
  clearOnUnauthorized?: boolean
}

export function getToken(): string | null {
  if (typeof localStorage === 'undefined') return null
  return localStorage.getItem(TOKEN_KEY)
}

export function setToken(token: string | null) {
  if (typeof localStorage === 'undefined') return
  if (token) localStorage.setItem(TOKEN_KEY, token)
  else localStorage.removeItem(TOKEN_KEY)
}

/** 兼容 `/auth/login` 与历史 `/api/v1/auth/login` */
export function resolveUrl(path: string, apiBase = getApiBase()): string {
  if (/^https?:\/\//i.test(path)) return path
  let p = path.startsWith('/') ? path : `/${path}`
  if (apiBase.endsWith('/api/v1') && p.startsWith('/api/v1')) {
    p = p.slice('/api/v1'.length) || '/'
  }
  if (!apiBase) return p
  return `${apiBase}${p}`
}

function parseError(data: unknown, status: number, statusText: string): ApiErrorBody {
  if (data && typeof data === 'object') {
    const o = data as Record<string, unknown>
    if (typeof o.message === 'string') {
      return {
        code: typeof o.code === 'string' ? o.code : 'http',
        message: o.message,
      }
    }
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

export async function api<T>(path: string, init: ApiRequestInit = {}): Promise<T> {
  const { skipAuth, idempotencyKey, clearOnUnauthorized = true, ...req } = init
  const headers = new Headers(req.headers)
  if (!headers.has('Content-Type') && req.body) {
    headers.set('Content-Type', 'application/json')
  }
  const explicitAuth = headers.has('Authorization')
  const token = getToken()
  if (!skipAuth && token && !explicitAuth) {
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
    const err = new ApiError(body.message, body.code, res.status)
    if (res.status === 401 && clearOnUnauthorized && !skipAuth && !explicitAuth) {
      setToken(null)
    }
    throw err
  }

  return data as T
}
