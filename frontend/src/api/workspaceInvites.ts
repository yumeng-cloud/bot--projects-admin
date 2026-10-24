import { api, ApiError, type ApiRequestInit } from './client'

/** 人类工作区邀请（ymw_）— 勿与队席 ymt_ / teams.ts 混用 */

export type WorkspaceInviteStatus = 'active' | 'revoked' | 'expired' | 'exhausted'

export type WorkspaceInvite = {
  id: string
  workspace_id: string
  email?: string | null
  role: string
  token_prefix?: string
  code_tail?: string
  max_uses?: number | null
  use_count?: number
  expires_at?: string | null
  revoked_at?: string | null
  status?: WorkspaceInviteStatus | string
  created_at?: string
  /** 仅签发当次明文；列表无此字段 */
  code?: string
}

export type CreateWorkspaceInviteBody = {
  email?: string
  role: 'admin' | 'member'
  expires_at?: string
  max_uses?: number
}

export type RedeemWorkspaceInviteResult = {
  workspace_id: string
  role?: string
  member_id?: string
  member?: {
    id: string
    workspace_id?: string
    user_id?: string | null
    role: string
    display_name?: string
    email?: string | null
    created_at?: string
  }
}

/** 签发邀请链接；响应含明文 code（仅此一次） */
export async function createWorkspaceInvite(
  workspaceId: string,
  body: CreateWorkspaceInviteBody,
  idempotencyKey?: string,
) {
  const payload: Record<string, unknown> = {
    role: body.role || 'member',
  }
  const email = body.email?.trim().toLowerCase()
  if (email) payload.email = email
  if (body.expires_at?.trim()) payload.expires_at = body.expires_at.trim()
  if (body.max_uses != null) payload.max_uses = body.max_uses

  const res = await api<{ invite: WorkspaceInvite; code?: string }>(
    `/workspaces/${workspaceId}/invites`,
    {
      method: 'POST',
      body: JSON.stringify(payload),
      idempotencyKey,
      skipToast: true,
    },
  )
  if (res.code && res.invite && !res.invite.code) {
    res.invite = { ...res.invite, code: res.code }
  }
  return res
}

export function listWorkspaceInvites(workspaceId: string) {
  return api<{ invites: WorkspaceInvite[] }>(`/workspaces/${workspaceId}/invites`, {
    skipToast: true,
  } as ApiRequestInit)
}

/** 吊销；已吊销幂等 200 */
export function revokeWorkspaceInvite(
  workspaceId: string,
  inviteId: string,
  idempotencyKey?: string,
) {
  return api<{ invite: WorkspaceInvite }>(
    `/workspaces/${workspaceId}/invites/${inviteId}/revoke`,
    {
      method: 'POST',
      body: JSON.stringify({}),
      idempotencyKey,
      skipToast: true,
    },
  )
}

/** 登录用户兑换 ymw_ 入席；拒绝 ymt_ 时后端返回中文 400 */
export function redeemWorkspaceInvite(code: string, idempotencyKey?: string) {
  return api<RedeemWorkspaceInviteResult>(`/workspace-invites/redeem`, {
    method: 'POST',
    body: JSON.stringify({ code: code.trim() }),
    idempotencyKey,
    skipToast: true,
  })
}

export function workspaceInviteStatusCn(status?: string | null): string {
  switch ((status || '').toLowerCase()) {
    case 'revoked':
      return '已吊销'
    case 'expired':
      return '已过期'
    case 'exhausted':
      return '已用尽'
    case 'active':
      return '有效'
    default:
      return status || '有效'
  }
}

export function workspaceInviteTail(inv: WorkspaceInvite): string {
  if (inv.code_tail) return inv.code_tail
  const raw = inv.code || inv.token_prefix || ''
  const clean = raw.replace(/^ymw_/i, '')
  return clean.slice(-4) || '····'
}

export function isWorkspaceInviteMissing(err: unknown): boolean {
  return err instanceof ApiError && (err.status === 404 || err.status === 501)
}
