import { api, setToken, type ApiRequestInit } from './client'

export type User = {
  id: string
  email: string
  name: string
  created_at?: string
}

export type AuthResp = { token: string; user: User }

export function login(body: { email: string; password: string }, init?: ApiRequestInit) {
  return api<AuthResp>('/auth/login', {
    method: 'POST',
    body: JSON.stringify(body),
    skipToast: true,
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
    skipToast: true,
    ...init,
  }).then((data) => {
    setToken(data.token)
    return data
  })
}

export function logout() {
  return api<{ message: string }>('/auth/logout', { method: 'POST', skipToast: true }).finally(
    () => setToken(null),
  )
}

export function me() {
  return api<{ user: User }>('/auth/me')
}

export { listMyWorkspaces as listWorkspaces } from './workspaces'
