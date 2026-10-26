import { api, type ApiRequestInit } from './client'

/** 与 Go MV2-G1 / MOBILE_V2 债 camelCase 对齐 */
export type SessionContext = {
  wsId: string
  teamId: string
  projectId: string
  issueId: string
  view: string
  updatedAt: string
}

export type SessionContextPut = {
  wsId?: string
  teamId?: string
  projectId?: string
  issueId?: string
  view?: string
}

export function emptySessionContext(): SessionContext {
  return { wsId: '', teamId: '', projectId: '', issueId: '', view: '', updatedAt: '' }
}

/** 是否有可续的现场（任一关键字段非空） */
export function hasSessionScene(ctx: SessionContext | null | undefined): boolean {
  if (!ctx) return false
  return Boolean(ctx.wsId || ctx.teamId || ctx.projectId || ctx.issueId || ctx.view)
}

export function getSessionContext(init?: ApiRequestInit) {
  return api<SessionContext>('/me/session-context', init)
}

export function putSessionContext(body: SessionContextPut, init?: ApiRequestInit) {
  return api<SessionContext>('/me/session-context', {
    method: 'PUT',
    body: JSON.stringify(body),
    ...init,
  })
}
