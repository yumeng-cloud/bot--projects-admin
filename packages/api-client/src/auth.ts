import { api, setToken, type ApiRequestInit } from './client'
import type { AuthResp, MeResp } from './types'

export function login(body: { email: string; password: string }, init?: ApiRequestInit) {
  return api<AuthResp>('/auth/login', {
    method: 'POST',
    body: JSON.stringify(body),
    clearOnUnauthorized: false,
    ...init,
  }).then((data) => {
    setToken(data.token)
    return data
  })
}

export function register(
  body: { email: string; password: string; name?: string },
  init?: ApiRequestInit,
) {
  return api<AuthResp>('/auth/register', {
    method: 'POST',
    body: JSON.stringify(body),
    clearOnUnauthorized: false,
    ...init,
  }).then((data) => {
    setToken(data.token)
    return data
  })
}

export function logout() {
  return api<{ message: string }>('/auth/logout', {
    method: 'POST',
    clearOnUnauthorized: false,
  }).finally(() => setToken(null))
}

export function me() {
  return api<MeResp>('/auth/me')
}
