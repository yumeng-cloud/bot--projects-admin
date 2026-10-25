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
export type TeamSeatStatus = 'active' | 'empty' | 'open' | 'left' | 'stopped' | 'departed'

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

export type InviteStatus = 'active' | 'revoked' | 'expired' | 'exhausted'

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
  /** 派生态，不落库 */
  status?: InviteStatus | string
  created_at?: string
  /** 签发时锚定的席位 slug（列表分组用） */
  role_hint?: string | null
  seat_slug?: string | null
}

export type RedeemInviteResult = {
  workspace_id: string
  team_id: string
  team?: Team
  roster?: TeamSeat[]
  allow_expand_seat?: boolean
  role_hint?: string | null
  seat_slug?: string | null
  seat_id?: string
  member_id?: string
  role_slug?: string
  /** bootstrap 当次明文 ymk_；续队省略 */
  token?: string
  token_issued?: boolean
}

export function listTeams(workspaceId: string) {
  return api<{ teams: Team[] }>(`${ws(workspaceId)}/teams`)
}

export function getTeam(workspaceId: string, teamId: string) {
  return api<{ team: Team }>(`${ws(workspaceId)}/teams/${teamId}`)
}

/** T7-R：本 WS 新建团队；可选 template_key（Go 挂模板时可能顺带建空席） */
export function createTeam(
  workspaceId: string,
  body: { name: string; slug?: string; template_key?: string },
  idempotencyKey?: string,
) {
  return api<{ team: Team; roster?: TeamSeat[] }>(`${ws(workspaceId)}/teams`, {
    method: 'POST',
    body: JSON.stringify(body),
    idempotencyKey,
    skipToast: true,
  })
}

/** S-R1：Admin 改名（Go S-G1 PATCH …/teams/{teamId}）；可选 slug 预留 */
export function patchTeam(
  workspaceId: string,
  teamId: string,
  body: { name: string; slug?: string },
) {
  return api<{ team: Team }>(`${ws(workspaceId)}/teams/${teamId}`, {
    method: 'PATCH',
    body: JSON.stringify(body),
    skipToast: true,
  })
}

/** 预览/ensure 用席位行 */
export type TeamTemplateSeat = {
  slug: string
  name: string
  /** 建议挂载；ensure 入参为字符串（JSON {repo,ref,path} 或 path） */
  playbook_ref: string
}

export type TeamTemplateCatalogItem = {
  key: string
  name: string
  description?: string
  seats: TeamTemplateSeat[]
  source: 'api' | 'builtin'
}

/** FE 内置「默认班底」—— GET 模板目录未就绪时仅作预览 */
export const BUILTIN_TEAM_TEMPLATES: TeamTemplateCatalogItem[] = [
  {
    key: 'default',
    name: '默认班底',
    description: '五席：前端 / 产品经理 / QA / 后端 / 知识库；创建后复制入场包即可',
    source: 'builtin',
    seats: [
      { slug: 'frontend', name: '前端', playbook_ref: 'playbook/frontend.md' },
      { slug: 'pm', name: '产品经理', playbook_ref: 'playbook/pm.md' },
      { slug: 'qa', name: 'QA', playbook_ref: 'playbook/qa.md' },
      { slug: 'backend', name: '后端', playbook_ref: 'playbook/backend.md' },
      { slug: 'wiki', name: '知识库', playbook_ref: 'playbook/wiki.md' },
    ],
  },
]

/** 预览 path → ensure 用的 JSON 字符串（兼容校验 playbook_ref 的 Go） */
export function playbookRefForEnsure(pathOrJson: string): string {
  const raw = (pathOrJson || '').trim()
  if (!raw) return ''
  if (raw.startsWith('{')) return raw
  return JSON.stringify({
    repo: 'bot-team-template',
    ref: 'main',
    path: raw.replace(/^\/+/, ''),
  })
}

function normalizePlaybookRef(v: unknown): string {
  if (v == null) return ''
  if (typeof v === 'string') {
    const s = v.trim()
    if (!s) return ''
    if (s.startsWith('{')) {
      try {
        const o = JSON.parse(s) as { path?: string }
        return (o.path || s).trim()
      } catch {
        return s
      }
    }
    return s
  }
  if (typeof v === 'object') {
    const o = v as { path?: string; repo?: string; ref?: string }
    if (o.path) return String(o.path)
    try {
      return JSON.stringify(v)
    } catch {
      return ''
    }
  }
  return String(v)
}

function normalizeTemplateItem(raw: unknown, source: 'api' | 'builtin'): TeamTemplateCatalogItem | null {
  if (!raw || typeof raw !== 'object') return null
  const o = raw as Record<string, unknown>
  const key = String(o.key || o.id || o.template_key || '').trim().toLowerCase()
  if (!key) return null
  const name = String(o.name || key).trim() || key
  const description = o.description != null ? String(o.description) : undefined
  const seatsRaw = Array.isArray(o.seats) ? o.seats : []
  const seats: TeamTemplateSeat[] = seatsRaw
    .map((row) => {
      if (!row || typeof row !== 'object') return null
      const s = row as Record<string, unknown>
      const slug = String(s.slug || s.role_slug || '').trim().toLowerCase()
      if (!slug) return null
      const seatName = String(s.name || s.display_name || slug).trim() || slug
      const pref = normalizePlaybookRef(s.playbook_ref)
      return { slug, name: seatName, playbook_ref: pref || `playbook/${slug}.md` }
    })
    .filter((x): x is TeamTemplateSeat => Boolean(x))
  if (!seats.length) return null
  return { key, name, description, seats, source }
}

/**
 * 拉模板目录：优先 GET …/team-templates 与 …/bot-team-templates；
 * 404/未就绪 → 内置「默认班底」。
 */
export async function listTeamTemplates(workspaceId: string): Promise<{
  templates: TeamTemplateCatalogItem[]
  fromApi: boolean
}> {
  const paths = [
    `${ws(workspaceId)}/team-templates`,
    `${ws(workspaceId)}/bot-team-templates`,
  ]
  for (const p of paths) {
    try {
      const res = await api<{ templates?: unknown[]; items?: unknown[] }>(p, {
        skipToast: true,
      } as ApiRequestInit)
      const list = res.templates || res.items || []
      const templates = list
        .map((x) => normalizeTemplateItem(x, 'api'))
        .filter((x): x is TeamTemplateCatalogItem => Boolean(x))
      if (templates.length) return { templates, fromApi: true }
    } catch (e: unknown) {
      if (e instanceof ApiError && (e.status === 404 || e.status === 501)) continue
      if (e instanceof ApiError && e.status >= 500) continue
      // 其它错误也降级内置，避免挡创建
      break
    }
  }
  return { templates: BUILTIN_TEAM_TEMPLATES.map((t) => ({ ...t, seats: t.seats.map((s) => ({ ...s })) })), fromApi: false }
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
  init?: Pick<ApiRequestInit, 'skipToast'>,
) {
  return api<{ roster: TeamSeat[] }>(`${ws(workspaceId)}/teams/${teamId}/roster/ensure`, {
    method: 'POST',
    body: JSON.stringify(body || {}),
    idempotencyKey,
    skipToast: init?.skipToast,
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
  return api<{ seat: TeamSeat; message?: string }>(
    `${ws(workspaceId)}/teams/${teamId}/seats/${seatId}/unbind`,
    {
      method: 'POST',
      body: JSON.stringify(body || {}),
      idempotencyKey,
      skipToast: true,
    },
  )
}

/** Admin 停工：清绑 + status=stopped（可补位）；与 unbind→left 对仗 */
export function stopSeat(
  workspaceId: string,
  teamId: string,
  seatId: string,
  idempotencyKey?: string,
) {
  return api<{ seat: TeamSeat; message?: string }>(
    `${ws(workspaceId)}/teams/${teamId}/seats/${seatId}/stop`,
    {
      method: 'POST',
      body: '{}',
      idempotencyKey,
      skipToast: true,
    },
  )
}

/** 签发入场码（Go G3）；404 → 人话由调用方处理 */
export async function createTeamInvite(
  workspaceId: string,
  teamId: string,
  body?: {
    allow_expand_seat?: boolean
    max_uses?: number
    expires_at?: string
    /** 席位 slug 提示（ONE_TAP；URL/提示词仍以 FE 拼装为准） */
    role_hint?: string
    seat_slug?: string
  },
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

/** 兑换入场码 → workspace+team(+roster)；可回 token/token_issued（空手 bootstrap） */
export function redeemTeamInvite(
  code: string,
  opts?: { idempotencyKey?: string; botKey?: string | null; role?: string | null; seat?: string | null },
) {
  const botKey = (opts?.botKey || '').trim()
  const body: Record<string, string> = { code: code.trim() }
  const role = (opts?.role || '').trim()
  const seat = (opts?.seat || '').trim()
  if (role) body.role = role
  if (seat) body.seat = seat
  const headers: Record<string, string> = {}
  if (botKey) {
    headers.Authorization = `Bearer ${botKey}`
  }
  return api<RedeemInviteResult>(`/team-invites/redeem`, {
    method: 'POST',
    body: JSON.stringify(body),
    headers,
    idempotencyKey: opts?.idempotencyKey,
    skipToast: true,
    // 无 ymk_ 时不带人类 JWT，走无 Bearer bootstrap；有 ymk_ 则显式 Bearer
    skipAuth: !botKey,
  })
}

export function isInviteServiceMissing(err: unknown): boolean {
  return err instanceof ApiError && (err.status === 404 || err.status === 501)
}

export function isSeatConflict(err: unknown): boolean {
  if (!(err instanceof ApiError) || err.status !== 409) return false
  return /占用|顶替|已绑定|conflict/i.test(err.message) || err.code === 'conflict'
}

/** UNBIND_LOOP U5：脏绑已清理人话 */
export const SEAT_STALE_BIND_CLEARED = '席位绑定已失效，已清理'

/** UNBIND_LOOP U6：他人存活占用人话 */
export function seatOccupiedByHuman(displayName: string): string {
  const name = (displayName || '').trim() || '其他成员'
  return `席位由「${name}」占用；请对方离席，或让管理员停工/离职解绑`
}

export function isSeatStaleBindClearedMessage(msg?: string | null): boolean {
  if (!msg) return false
  return /席位绑定已失效|已失效.*清理|绑定已失效/.test(msg)
}

/** U5/U4：错误体实为「已解除」→ 按成功处理并刷新 */
export function isSeatLeaveEffectiveSuccess(err: unknown): boolean {
  if (!(err instanceof ApiError)) return false
  const m = err.message || ''
  if (isSeatStaleBindClearedMessage(m)) return true
  // 无 bound 幂等：偶发 4xx 仍带「已解绑/无绑定」
  if (/无\s*bound|无绑定|已解绑|already\s*unbound|not\s*bound/i.test(m)) return true
  return false
}

/**
 * UNBIND_LOOP U-R2：stop/unbind/leave 错误人话。
 * API 已回新人话则透传；否则对 403/冲突改写为 U6。
 */
export function humanizeSeatLeaveError(
  err: unknown,
  opts?: { occupantDisplayName?: string | null },
): string {
  if (!(err instanceof ApiError)) {
    return err instanceof Error ? err.message : '操作失败'
  }
  const m = (err.message || '').trim()
  if (isSeatStaleBindClearedMessage(m)) return SEAT_STALE_BIND_CLEARED
  // 已是 U6 句式 → 透传
  if (/席位由「.+」占用/.test(m)) return m
  if (err.status === 403 || err.status === 409 || err.code === 'forbidden' || err.code === 'conflict') {
    if (/只能解绑|自己绑定|无权解绑|不能解绑|占用中|已被.*占用|seat.*bound|forbidden/i.test(m) || !m) {
      const name = (opts?.occupantDisplayName || '').trim()
      if (name) return seatOccupiedByHuman(name)
      // 无显示名时仍给可操作指引，勿甩技术腔
      return '席位由其他成员占用；请对方离席，或让管理员停工/离职解绑'
    }
  }
  return m || '操作失败'
}

/** 设置·移除成员后：班底页乐观清绑 / 立刻 reload（U-R1） */
export const MEMBER_REMOVED_EVENT = 'acb:member-removed'

export type MemberRemovedDetail = { workspaceId: string; memberId: string }

export function emitMemberRemoved(workspaceId: string, memberId: string) {
  if (typeof window === 'undefined') return
  window.dispatchEvent(
    new CustomEvent<MemberRemovedDetail>(MEMBER_REMOVED_EVENT, {
      detail: { workspaceId, memberId },
    }),
  )
}


/** Soft† V2G：通码踩停工席 → 人话引导（禁静默补） */
export const SEAT_REFILL_GUIDE_TITLE = '普通入场码不能自动顶停工席'
export const SEAT_REFILL_GUIDE_DESC =
  '请向管理员要该席的补位入场码（班底页 · 停工席 · 复制补位码）；持补位码入场即占该席'

/** roster 是否存在 stopped|left|departed 可补位席 */
export function rosterHasRefillableSeat(seats: TeamSeat[]): boolean {
  return seats.some((s) => seatOccupancy(s) === 'left')
}

/** 409「补位须走显式绑席…」类（BE invite_join）；勿原文甩技术句 */
export function isRefillNeedsExplicitBind(err: unknown): boolean {
  if (!(err instanceof ApiError) || err.status !== 409) return false
  const m = err.message || ''
  return /补位须走显式绑席|不自动补位|显式绑席/i.test(m)
}

/** 改写为 V2G 人话；非此类 409 返回 null */
export function humanizeRefillBindError(err: unknown): string | null {
  if (!isRefillNeedsExplicitBind(err)) return null
  return `${SEAT_REFILL_GUIDE_TITLE}。${SEAT_REFILL_GUIDE_DESC}`
}

/** 推断席位态：显式 status 优先；否则 claim/bound → active，否则 empty */
export function seatOccupancy(seat: TeamSeat): 'active' | 'empty' | 'left' {
  const st = (seat.status || '').toLowerCase()
  if (st === 'left' || st === 'stopped' || st === 'departed') return 'left'
  if (st === 'active') return 'active'
  if (st === 'empty' || st === 'open') return 'empty'
  if (seat.claim?.member_id) return 'active'
  if (seat.bound_member_id) return 'active'
  return 'empty'
}

export function seatDisplayName(seat: TeamSeat): string {
  const name = (seat.display_name || '').trim()
  if (name && !/^team:/i.test(name) && !/[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}/i.test(name)) {
    return name
  }
  const slug = (seat.role_slug || '').trim()
  if (slug && !/^team:/i.test(slug) && !/[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}/i.test(slug)) {
    return slug
  }
  // team:uuid:wiki → wiki
  if (/^team:/i.test(slug)) {
    const last = slug.split(':').pop()?.trim() || ''
    if (last && !/[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}/i.test(last)) {
      return last
    }
  }
  return '未命名席位'
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
