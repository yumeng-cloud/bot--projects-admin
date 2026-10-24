import { api, ApiError, type ApiRequestInit } from './client'

const ws = (id: string) => `/workspaces/${id}`

export type Team = {
  id: string
  workspace_id: string
  name: string
  slug?: string | null
  template_key: string
  created_at: string
  updated_at: string
}

export type TeamSeatClaim = {
  member_id: string
  claimed_at: string
}

/** 席位态：有 claim/绑定且活跃 → active；无绑 → empty；产品 left/stopped 可显式标 */
export type TeamSeatStatus = 'active' | 'empty' | 'left' | 'stopped' | 'departed'

export type TeamSeat = {
  id: string
  team_id: string
  workspace_id: string
  role_slug: string
  display_name: string
  playbook_ref?: string | null
  bot_role_id?: string | null
  bound_member_id?: string | null
  member_id?: string | null
  bound_at?: string | null
  bound_display_name?: string | null
  /** Go G3 未齐时可能缺；FE 用 claim/bound 推断 */
  status?: TeamSeatStatus | string | null
  left_at?: string | null
  created_at: string
  updated_at: string
  claim?: TeamSeatClaim | null
}

export type TeamInvite = {
  id: string
  team_id: string
  workspace_id: string
  /** 仅签发当次明文；之后列表只有末四 */
  code?: string
  code_prefix?: string
  token_prefix?: string
  code_tail?: string
  allow_expand_seat?: boolean
  max_uses?: number | null
  uses?: number
  use_count?: number
  expires_at?: string | null
  revoked_at?: string | null
  /** active | revoked | expired | exhausted */
  status?: string
  created_at?: string
}

export type RedeemInviteResult = {
  workspace_id: string
  team_id: string
  team?: Team
  roster?: TeamSeat[]
  allow_expand_seat?: boolean
}

export function listTeams(workspaceId: string) {
  return api<{ teams: Team[] }>(`${ws(workspaceId)}/teams`)
}

export function getTeam(workspaceId: string, teamId: string) {
  return api<{ team: Team }>(`${ws(workspaceId)}/teams/${teamId}`)
}

export function getRoster(workspaceId: string, teamId: string) {
  return api<{ roster: TeamSeat[] }>(`${ws(workspaceId)}/teams/${teamId}/roster`)
}

export function ensureRoster(
  workspaceId: string,
  teamId: string,
  body?: {
    template_key?: string
    seats?: { slug: string; name?: string; playbook_ref?: string }[]
  },
  idempotencyKey?: string,
) {
  return api<{ roster: TeamSeat[] }>(`${ws(workspaceId)}/teams/${teamId}/roster/ensure`, {
    method: 'POST',
    body: JSON.stringify(body || {}),
    idempotencyKey,
  })
}

export function bindSeat(
  workspaceId: string,
  teamId: string,
  seatId: string,
  body?: { member_id?: string; force?: boolean },
  idempotencyKey?: string,
) {
  return api<{ seat: TeamSeat }>(`${ws(workspaceId)}/teams/${teamId}/seats/${seatId}/bind`, {
    method: 'POST',
    body: JSON.stringify(body || {}),
    idempotencyKey,
    skipToast: true,
  })
}

export function unbindSeat(
  workspaceId: string,
  teamId: string,
  seatId: string,
  body?: { member_id?: string },
  idempotencyKey?: string,
) {
  return api<{ seat: TeamSeat }>(`${ws(workspaceId)}/teams/${teamId}/seats/${seatId}/unbind`, {
    method: 'POST',
    body: JSON.stringify(body || {}),
    idempotencyKey,
  })
}

/** 签发入场码（Go G3）；404 → 人话由调用方处理 */
export async function createTeamInvite(
  workspaceId: string,
  teamId: string,
  body?: { allow_expand_seat?: boolean; max_uses?: number; expires_at?: string },
  idempotencyKey?: string,
) {
  const res = await api<{ invite: TeamInvite; code?: string }>(
    `${ws(workspaceId)}/teams/${teamId}/invites`,
    {
      method: 'POST',
      body: JSON.stringify(body || {}),
      idempotencyKey,
      skipToast: true,
    },
  )
  // BE 同时返回顶层 code 与 invite.code；明文仅此一次
  if (res.code && res.invite && !res.invite.code) {
    res.invite = { ...res.invite, code: res.code }
  }
  return res
}

export function listTeamInvites(workspaceId: string, teamId: string) {
  return api<{ invites: TeamInvite[] }>(`${ws(workspaceId)}/teams/${teamId}/invites`, {
    skipToast: true,
  } as ApiRequestInit)
}

/** Admin 吊销入场码；幂等已吊销仍 200 */
export function revokeTeamInvite(
  workspaceId: string,
  teamId: string,
  inviteId: string,
  idempotencyKey?: string,
) {
  return api<{ invite: TeamInvite }>(
    `${ws(workspaceId)}/teams/${teamId}/invites/${inviteId}/revoke`,
    {
      method: 'POST',
      body: JSON.stringify({}),
      idempotencyKey,
      skipToast: true,
    },
  )
}

export function inviteStatusCn(status?: string | null): string {
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

/** 兑换入场码 → workspace+team(+roster) */
export function redeemTeamInvite(code: string, idempotencyKey?: string) {
  return api<RedeemInviteResult>(`/team-invites/redeem`, {
    method: 'POST',
    body: JSON.stringify({ code: code.trim() }),
    idempotencyKey,
    skipToast: true,
  })
}

export function isInviteServiceMissing(err: unknown): boolean {
  return err instanceof ApiError && (err.status === 404 || err.status === 501)
}

export function isSeatConflict(err: unknown): boolean {
  if (!(err instanceof ApiError) || err.status !== 409) return false
  return /占用|顶替|已绑定|conflict/i.test(err.message) || err.code === 'conflict'
}

/** 推断席位态：显式 status 优先；否则 claim/bound → active，否则 empty */
export function seatOccupancy(seat: TeamSeat): 'active' | 'empty' | 'left' {
  const st = (seat.status || '').toLowerCase()
  if (st === 'left' || st === 'stopped' || st === 'departed') return 'left'
  if (st === 'active') return 'active'
  if (st === 'empty') return 'empty'
  if (seat.claim?.member_id) return 'active'
  if (seat.bound_member_id) return 'active'
  return 'empty'
}

export function seatDisplayName(seat: TeamSeat): string {
  return seat.display_name || seat.role_slug || '未命名席位'
}

export function inviteTail(inv: TeamInvite): string {
  if (inv.code_tail) return inv.code_tail
  const raw = inv.code || inv.code_prefix || inv.token_prefix || ''
  const clean = raw.replace(/^ymt_/i, '')
  return clean.slice(-4) || '····'
}

export function maskKeyTail(key: string): string {
  const clean = key.replace(/^ymk_/i, '').replace(/[^a-zA-Z0-9]/g, '')
  const tail = clean.slice(-4)
  return tail ? `ymk_····${tail}` : 'ymk_····'
}
