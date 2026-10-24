import { useEffect, useMemo, useState } from 'react'
import { Link, useParams, useSearchParams } from 'react-router-dom'
import {
  Alert,
  Button,
  Card,
  Collapse,
  Drawer,
  Empty,
  Checkbox,
  Dropdown,
  Input,
  Modal,
  Popconfirm,
  Select,
  Space,
  Spin,
  Table,
  Tag,
  Tooltip,
  Typography,
  message,
} from 'antd'
import {
  CopyOutlined,
  FileTextOutlined,
  LinkOutlined,
  MoreOutlined,
  PlusOutlined,
  SendOutlined,
  UserAddOutlined,
} from '@ant-design/icons'
import {
  claimBotRole,
  listBotRoles,
  listBots,
  releaseBotRole,
  type BotRole,
  type BotRosterItem,
} from '../../../api/bots'
import { ApiError, api } from '../../../api/client'
import {
  bindSeat,
  createTeamInvite,
  getRoster,
  inviteStatusCn,
  inviteTail,
  isInviteServiceMissing,
  isSeatConflict,
  listTeamInvites,
  revokeTeamInvite,
  seatOccupancy,
  stopSeat,
  unbindSeat,
  type Team,
  type TeamInvite,
  type TeamSeat,
} from '../../../api/teams'
import { createIssue, notifyIssue } from '../../../api/issues'
import { listProjects } from '../../../api/projects'
import { useCurrentTeam } from '../../../hooks/useCurrentTeam'
import { formatWikiTime } from '../../../utils/time'
import JoinTeamWizard from './JoinTeamWizard'
import CreateTeamFromTemplateModal from './CreateTeamFromTemplateModal'
import SeatConflictPanel, { type SeatConflictInfo } from './SeatConflictPanel'
import SeatHandoffDrawer from './SeatHandoffDrawer'
import TeamPinsPanel from './TeamPinsPanel'
import { roleDisplay } from './seatHandoffText'
import {
  buildBotJoinPrompt,
  buildJoinUrl,
  inviteCodeTail,
  parseJoinSearchParams,
} from '../../../utils/oneTapJoin'
import { emptySeatsForPack } from '../../../utils/autoCrewPack'
import type { CrewPackResult } from '../../../utils/autoCrewPack'
import CrewPackRevealModal from '../../../components/common/CrewPackRevealModal'
import './bots.css'

type ApiKeyRow = {
  id: string
  member_id?: string
  name: string
  key_prefix: string
  last_used_at?: string | null
  revoked_at?: string | null
}

/** 仅展示前缀尾四位，禁止完整 Key */
function keyTailHint(prefix?: string | null): string | undefined {
  if (!prefix) return undefined
  const clean = prefix.replace(/^ymk_/i, '')
  const tail = clean.slice(-4)
  if (!tail) return undefined
  return `末四 ·${tail}`
}

export default function BotsRosterPage() {
  const { wsId = '' } = useParams()
  const [searchParams, setSearchParams] = useSearchParams()
  const forceEmpty = searchParams.get('f1empty') === '1'
  const { teams, teamId, setTeamId, loading: teamLoading, reload } = useCurrentTeam(wsId)
  const [loading, setLoading] = useState(true)
  const [bots, setBots] = useState<BotRosterItem[]>([])
  const [roles, setRoles] = useState<BotRole[]>([])
  const [keys, setKeys] = useState<ApiKeyRow[]>([])
  const [drawerRole, setDrawerRole] = useState<BotRole | null>(null)
  const [claiming, setClaiming] = useState<string | null>(null)
  const [joinOpen, setJoinOpen] = useState(false)
  /** C1：深链预填（code/role/seat） */
  const [joinPrefill, setJoinPrefill] = useState<{
    code: string
    role: string
    seat: string
  } | null>(null)
  const [createFromTplOpen, setCreateFromTplOpen] = useState(false)
  const [conflictOpen, setConflictOpen] = useState(false)
  const [conflictInfo, setConflictInfo] = useState<SeatConflictInfo | null>(null)
  /** 与向导同源的团队席位占用（N4 同步） */
  const [teamSeats, setTeamSeats] = useState<TeamSeat[]>([])
  /** 建队瞬间 teams 列表尚未 reload 时顶栏/页眉仍显示新队名 */
  const [pendingCreatedTeam, setPendingCreatedTeam] = useState<Team | null>(null)
  const [issuedInvites, setIssuedInvites] = useState<TeamInvite[]>([])
  const [inviteBusy, setInviteBusy] = useState(false)
  /** 签发当次明文，只存活到复制成功后的一次 Toast；列表只留末四 */
  const [lastPlainInvite, setLastPlainInvite] = useState<string | null>(null)
  /** AUTO_CREW：全班入场包挡层（seats 快照防切队 refetch 冲掉） */
  const [crewPackOpen, setCrewPackOpen] = useState(false)
  const [crewPackAfterCreate, setCrewPackAfterCreate] = useState(false)
  const [crewPackSeats, setCrewPackSeats] = useState<TeamSeat[]>([])
  const [revokingId, setRevokingId] = useState<string | null>(null)
  /** N2：占用席下发任务 */
  const [dispatchOpen, setDispatchOpen] = useState(false)
  const [dispatchTitle, setDispatchTitle] = useState('')
  const [dispatchBusy, setDispatchBusy] = useState(false)
  /** N4：立即推送（默认关） */
  const [dispatchPush, setDispatchPush] = useState(false)
  const [dispatchTarget, setDispatchTarget] = useState<{
    memberId: string
    seatName: string
    botName: string
  } | null>(null)

  const load = async () => {
    setLoading(true)
    try {
      const [b, r, k] = await Promise.all([
        listBots(wsId),
        listBotRoles(wsId),
        api<{ api_keys: ApiKeyRow[] }>(`/workspaces/${wsId}/api-keys`).catch(() => ({
          api_keys: [] as ApiKeyRow[],
        })),
      ])
      setBots(b.bots || [])
      setRoles(r.bot_roles || [])
      setKeys((k.api_keys || []).filter((x) => !x.revoked_at))
    } catch (e: unknown) {
      message.error(e instanceof Error ? e.message : '加载失败')
    } finally {
      setLoading(false)
    }
  }

  // ONE_TAP C1：?join=1&code=ymt_… 或仅 ?code=ymt_… → 打开预填 JoinWizard，并清 query
  useEffect(() => {
    const parsed = parseJoinSearchParams(searchParams)
    const want =
      searchParams.get('join') === '1' ||
      (parsed.code.startsWith('ymt_') && parsed.code.length > 4)
    if (!want || !parsed.code) return
    setJoinPrefill({
      code: parsed.code,
      role: parsed.role,
      seat: parsed.seat,
    })
    setJoinOpen(true)
    const next = new URLSearchParams(searchParams)
    next.delete('join')
    next.delete('code')
    next.delete('role')
    next.delete('seat')
    next.delete('api')
    setSearchParams(next, { replace: true })
  }, [searchParams, setSearchParams])


  useEffect(() => {
    void load()
  }, [wsId])

  /** 顶栏 CurrentTeam 换队：邀请缓存清空并 refetch；保留同队已 seed 的 roster（建队 onCreated）防闪空 */
  useEffect(() => {
    setIssuedInvites([])
    if (!wsId || !teamId) {
      setTeamSeats([])
      return
    }
    setTeamSeats((prev) => prev.filter((s) => s.team_id === teamId))
    let cancelled = false
    void (async () => {
      try {
        const rost = await getRoster(wsId, teamId)
        if (!cancelled) setTeamSeats(rost.roster || [])
      } catch {
        // 保留 onCreated 已 seed 的本队席，避免 Toast 有席而主区闪「尚未绑定」
        if (!cancelled) {
          setTeamSeats((prev) => (prev.some((s) => s.team_id === teamId) ? prev : []))
        }
      }
      // Go 29f9404+ 已挂 GET invites（无明文）；失败则保留本页签发缓存
      try {
        const inv = await listTeamInvites(wsId, teamId)
        if (!cancelled) setIssuedInvites(inv.invites || [])
      } catch {
        /* 保留本地签发缓存 */
      }
    })()
    return () => {
      cancelled = true
    }
  }, [wsId, teamId])

  /** SEAT_VISIBILITY：立刻切新队 + 用返回名册 seed，避免只 Toast */
  /** AUTO_CREW C1：同帧弹出全班入场包挡层 */
  const onTeamCreatedFromTemplate = (team: Team, roster: TeamSeat[]) => {
    setPendingCreatedTeam(team)
    setTeamSeats(roster || [])
    setTeamId(team.id)
    void reload().finally(() => setPendingCreatedTeam(null))
    void load()
    if (emptySeatsForPack(roster || []).length > 0) {
      setCrewPackSeats(roster || [])
      setCrewPackAfterCreate(true)
      setCrewPackOpen(true)
    }
  }

  const emptySeats = useMemo(() => emptySeatsForPack(teamSeats), [teamSeats])

  const onCrewPackIssued = (result: CrewPackResult) => {
    if (!result.ok.length) return
    setIssuedInvites((prev) => {
      const next = [...prev]
      for (const row of result.ok) {
        const inv = row.invite
        const tail = inviteTail(inv)
        const stripped = {
          ...inv,
          code: undefined,
          code_tail: tail,
          code_prefix: inv.code_prefix || `ymt_····${tail}`,
        }
        const idx = next.findIndex((x) => x.id === inv.id)
        if (idx >= 0) next[idx] = stripped
        else next.unshift(stripped)
      }
      return next
    })
  }

  const openCrewPack = (afterCreate = false) => {
    if (!teamId) {
      message.warning('尚无团队，请先创建或加入班底')
      return
    }
    if (emptySeats.length === 0) {
      message.info('当前队没有空席，无需签发全班入场包')
      return
    }
    setCrewPackSeats(teamSeats)
    setCrewPackAfterCreate(afterCreate)
    setCrewPackOpen(true)
  }

  const currentTeamName =
    teams.find((t) => t.id === teamId)?.name ||
    (pendingCreatedTeam && pendingCreatedTeam.id === teamId ? pendingCreatedTeam.name : undefined)
  const teamLabel = currentTeamName ? (
    <Typography.Text type="secondary" style={{ fontSize: 12 }}>
      当前队 · {currentTeamName}
    </Typography.Text>
  ) : null

  const botName = (memberId?: string | null) =>
    memberId ? bots.find((b) => b.member_id === memberId)?.display_name || '已绑定' : null

  const keyHintForMember = (memberId?: string | null) => {
    if (!memberId) return undefined
    const row = keys
      .filter((k) => k.member_id === memberId)
      .sort((a, b) => String(b.last_used_at || '').localeCompare(String(a.last_used_at || '')))[0]
    return keyTailHint(row?.key_prefix)
  }

  const freeBots = useMemo(() => bots.filter((b) => b.bot_enabled), [bots])
  /** TEAM_SWITCH S2 + SEAT_VISIBILITY：按本队 roster 出卡；无匹配 bot_role 时用席合成，保证空席可见 */
  const visibleRoles = useMemo(() => {
    if (!teamId) return roles
    if (!teamSeats.length) return []
    const byId = new Map(roles.map((r) => [r.id, r]))
    const bySlug = new Map(roles.map((r) => [r.slug, r]))
    const out: BotRole[] = []
    const seen = new Set<string>()
    for (const s of teamSeats) {
      const hit =
        (s.bot_role_id && byId.get(s.bot_role_id)) ||
        (s.role_slug && bySlug.get(s.role_slug)) ||
        null
      if (hit) {
        if (seen.has(hit.id)) continue
        seen.add(hit.id)
        out.push(hit)
        continue
      }
      const sid = s.bot_role_id || `seat:${s.id}`
      if (seen.has(sid)) continue
      seen.add(sid)
      out.push({
        id: sid,
        workspace_id: s.workspace_id || wsId,
        name: s.display_name || s.role_slug || '未命名席位',
        slug: s.role_slug || s.id,
        responsibilities: '',
        skills: '',
        default_project_ids: '',
        created_at: s.created_at,
        updated_at: s.updated_at,
        claim: null,
      })
    }
    return out
  }, [teamId, teamSeats, roles, wsId])
  /** SEAT_VISIBILITY：roster 非空（含纯空席）即走出空态；禁「有 5 空席仍尚未绑定」 */
  const hasAnyBinding = useMemo(() => {
    if (forceEmpty) return false
    if (teamSeats.length > 0) return true
    if (roles.some((r) => Boolean(r.claim?.member_id))) return true
    return false
  }, [roles, teamSeats, forceEmpty])

  const openSeat = (role: BotRole) => setDrawerRole(role)

  const onRevokeInvite = async (inv: TeamInvite) => {
    if (!teamId) return
    setRevokingId(inv.id)
    try {
      const { invite } = await revokeTeamInvite(wsId, teamId, inv.id, crypto.randomUUID())
      setIssuedInvites((prev) =>
        prev.map((x) =>
          x.id === inv.id
            ? { ...x, ...invite, status: invite.status || 'revoked', revoked_at: invite.revoked_at || x.revoked_at || new Date().toISOString() }
            : x,
        ),
      )
      message.success(`已吊销入场码（末四 ·${inviteTail(invite.status ? invite : inv)}）`)
    } catch (e: unknown) {
      if (isInviteServiceMissing(e)) {
        message.warning('吊销接口尚未开通')
      } else {
        message.error(e instanceof ApiError ? e.message : '吊销失败')
      }
    } finally {
      setRevokingId(null)
    }
  }

  const openDispatch = (role: BotRole) => {
    const occ = seatOccForRole(role)
    const memberId = occ.seat?.bound_member_id || occ.occupantId
    if (!occ.occupied || !memberId) {
      message.warning('请先占用席位再下发')
      return
    }
    setDispatchTarget({
      memberId,
      seatName: roleDisplay(role),
      botName: occ.occupantLabel || botName(memberId) || 'Bot',
    })
    setDispatchTitle('')
    setDispatchPush(false)
    setDispatchOpen(true)
  }

  const onDispatchCreate = async () => {
    if (!dispatchTarget) return
    const title = dispatchTitle.trim() || `给「${dispatchTarget.seatName}」的任务`
    setDispatchBusy(true)
    try {
      const { projects } = await listProjects(wsId)
      const projectId = projects?.[0]?.id
      if (!projectId) {
        message.error('暂无项目，无法建 Issue')
        return
      }
      const push = dispatchPush
      const { issue } = await createIssue(wsId, {
        project_id: projectId,
        title,
        status: 'todo',
        assignee_id: dispatchTarget.memberId,
        skip_notify: true,
        team_id: teamId || undefined,
      })
      message.success(`已下发 ${issue.identifier} → ${dispatchTarget.botName}`)
      if (push) {
        try {
          const res = await notifyIssue(wsId, issue.id)
          if ((res.failed ?? 0) > 0) {
            message.warning(
              `推送部分失败（成功 ${res.delivered ?? 0} / 失败 ${res.failed}），Issue 已创建`,
            )
          } else if ((res.delivered ?? 0) === 0) {
            message.info('暂无可用 Webhook 接收推送（Issue 已创建）')
          } else {
            message.success(`已立即推送（${res.delivered} 路）`)
          }
        } catch (pe: unknown) {
          message.warning(
            pe instanceof ApiError
              ? `推送失败：${pe.message}（Issue 已创建）`
              : '推送失败（Issue 已创建）',
          )
        }
      }
      setDispatchOpen(false)
      setDispatchTarget(null)
      setDispatchPush(false)
    } catch (e: unknown) {
      message.error(e instanceof ApiError ? e.message : '下发失败')
    } finally {
      setDispatchBusy(false)
    }
  }


  /** D5 + V7：team_seats 优先；席 empty 但 role.claim 已有人时仍算占用（防只 Toast 卡仍等待绑定） */
  const seatOccForRole = (role: BotRole) => {
    const claimMid = role.claim?.member_id || null
    const seat =
      teamSeats.find((s) => s.bot_role_id && s.bot_role_id === role.id) ||
      teamSeats.find((s) => s.role_slug && s.role_slug === role.slug) ||
      (role.id.startsWith('seat:')
        ? teamSeats.find((s) => s.id === role.id.slice(5))
        : undefined)
    const occ: 'active' | 'empty' | 'left' | null = seat ? seatOccupancy(seat) : null
    let kind: 'active' | 'empty' | 'left'
    if (occ === 'left') kind = 'left'
    else if (occ === 'active' || claimMid) kind = 'active'
    else if (occ === 'empty') kind = 'empty'
    else kind = 'empty'
    const occupantId =
      (kind === 'active'
        ? claimMid ||
          seat?.bound_member_id ||
          seat?.member_id ||
          seat?.claim?.member_id
        : null) || null
    const occupantLabel =
      (occupantId ? botName(occupantId) : null) ||
      (kind === 'active' ? seat?.bound_display_name : null) ||
      null
    return { kind, occupied: kind === 'active', occupantId, occupantLabel, seat }
  }

  const refreshTeamRoster = async () => {
    if (!wsId || !teamId) return
    try {
      const rost = await getRoster(wsId, teamId)
      setTeamSeats(rost.roster || [])
    } catch {
      /* 保留乐观态 */
    }
  }

  const applyOptimisticBind = (role: BotRole, memberId: string, bound?: TeamSeat | null) => {
    const label = botName(memberId) || '已绑定'
    const now = new Date().toISOString()
    setRoles((prev) =>
      prev.map((r) =>
        r.id === role.id
          ? {
              ...r,
              claim: {
                id: r.claim?.id || `optimistic:${r.id}`,
                workspace_id: r.workspace_id,
                role_id: r.id,
                member_id: memberId,
                claimed_at: r.claim?.claimed_at || now,
              },
            }
          : r,
      ),
    )
    setTeamSeats((prev) => {
      if (bound) {
        return prev.map((s) => (s.id === bound.id ? { ...s, ...bound } : s))
      }
      return prev.map((s) => {
        const match =
          (s.bot_role_id && s.bot_role_id === role.id) ||
          (s.role_slug && s.role_slug === role.slug) ||
          (role.id.startsWith('seat:') && s.id === role.id.slice(5))
        if (!match) return s
        return {
          ...s,
          status: 'active',
          bound_member_id: memberId,
          member_id: memberId,
          bound_display_name: label,
          bound_at: now,
          claim: { member_id: memberId, claimed_at: now },
        }
      })
    })
  }


  /** SEAT_LEAVE：乐观清绑为可补位态（stopped|left 均走 seatOccupancy→left →「停工席 · 可补位」） */
  const applyOptimisticLeave = (role: BotRole, status: 'stopped' | 'left', bound?: TeamSeat | null) => {
    setRoles((prev) => prev.map((r) => (r.id === role.id ? { ...r, claim: null } : r)))
    setTeamSeats((prev) => {
      if (bound) {
        return prev.map((s) => (s.id === bound.id ? { ...s, ...bound } : s))
      }
      return prev.map((s) => {
        const match =
          (s.bot_role_id && s.bot_role_id === role.id) ||
          (s.role_slug && s.role_slug === role.slug) ||
          (role.id.startsWith('seat:') && s.id === role.id.slice(5))
        if (!match) return s
        return {
          ...s,
          status,
          bound_member_id: null,
          member_id: null,
          bound_display_name: null,
          bound_at: null,
          claim: null,
        }
      })
    })
  }

  /** L3/L4：停工 → stopped；Toast + roster「停工席 · 可补位」；勿混 ymt_/Bot 入场文案 */
  const onStop = async (role: BotRole) => {
    const seat = seatOccForRole(role).seat
    if (!teamId || !seat?.id) {
      message.warning('无团队席位，无法停工')
      return
    }
    try {
      const { seat: stopped } = await stopSeat(wsId, teamId, seat.id, crypto.randomUUID())
      applyOptimisticLeave(role, 'stopped', stopped)
      if (!role.id.startsWith('seat:')) {
        try {
          await releaseBotRole(wsId, role.id, crypto.randomUUID())
        } catch {
          /* stop 已释 claim；release 幂等失败可忽略 */
        }
      }
      message.success(`已将「${roleDisplay(role)}」标为停工 · 可补位`)
      void load()
      void refreshTeamRoster()
    } catch (e: unknown) {
      message.error(e instanceof ApiError ? e.message : '停工失败')
    }
  }

  /** L3/L4：离职解绑 → left；与 stop 对仗 */
  const onLeaveUnbind = async (role: BotRole) => {
    const seat = seatOccForRole(role).seat
    if (!teamId || !seat?.id) {
      message.warning('无团队席位，无法离职解绑')
      return
    }
    try {
      const { seat: unbound } = await unbindSeat(wsId, teamId, seat.id, undefined, crypto.randomUUID())
      applyOptimisticLeave(role, 'left', unbound)
      if (!role.id.startsWith('seat:')) {
        try {
          await releaseBotRole(wsId, role.id, crypto.randomUUID())
        } catch {
          /* unbind 已释 claim */
        }
      }
      message.success(`已将「${roleDisplay(role)}」离职解绑 · 可补位`)
      void load()
      void refreshTeamRoster()
    } catch (e: unknown) {
      message.error(e instanceof ApiError ? e.message : '离职解绑失败')
    }
  }

  /** V7：绑成功同帧更新席卡；team 下优先 bindSeat，并乐观刷新 roles+teamSeats */
  const onClaim = async (role: BotRole, memberId: string) => {
    setClaiming(role.id)
    try {
      const seat = seatOccForRole(role).seat
      let bound: TeamSeat | null = null
      if (teamId && seat?.id) {
        try {
          const res = await bindSeat(
            wsId,
            teamId,
            seat.id,
            { member_id: memberId },
            crypto.randomUUID(),
          )
          bound = res.seat
        } catch (e: unknown) {
          if (isSeatConflict(e) || (e instanceof ApiError && e.status === 409)) {
            throw e
          }
          /* bind 未开通则回落 claimBotRole */
        }
      }
      if (!role.id.startsWith('seat:')) {
        try {
          await claimBotRole(wsId, role.id, memberId, crypto.randomUUID(), { skipToast: true })
        } catch (e: unknown) {
          if (!bound && (isSeatConflict(e) || (e instanceof ApiError && e.status === 409))) {
            throw e
          }
          if (!bound) throw e
        }
      } else if (!bound) {
        message.error('无法绑到合成席，请刷新后重试')
        return
      }
      applyOptimisticBind(role, memberId, bound)
      message.success(`已将「${botName(memberId)}」绑到席位「${roleDisplay(role)}」`)
      void load()
      void refreshTeamRoster()
    } catch (e: unknown) {
      if (isSeatConflict(e) || (e instanceof ApiError && e.status === 409)) {
        setConflictInfo({
          seatName: roleDisplay(role),
          occupantName: botName(role.claim?.member_id) || role.claim?.member_id || null,
          claimedAt: role.claim?.claimed_at,
          message: e instanceof ApiError ? e.message : '该席位占用中，禁止顶替',
        })
        setConflictOpen(true)
        return
      }
      message.error(e instanceof ApiError ? e.message : '认领失败')
    } finally {
      setClaiming(null)
    }
  }

  type InviteRoleCtx = { slug?: string | null; displayName?: string | null }

  /** 签发入场码并缓存末四；明文仅返回一次。role 写入 role_hint（可选）。 */
  const issueInvitePlain = async (roleCtx?: InviteRoleCtx): Promise<string | null> => {
    if (!teamId) {
      message.warning('尚无团队，请先创建或加入班底')
      return null
    }
    const slug = (roleCtx?.slug || '').trim()
    const { invite } = await createTeamInvite(
      wsId,
      teamId,
      {
        allow_expand_seat: true,
        ...(slug ? { role_hint: slug, seat_slug: slug } : {}),
      },
      crypto.randomUUID(),
    )
    const plain = invite.code || null
    const tail = inviteTail(invite)
    setIssuedInvites((prev) => [
      {
        ...invite,
        code: undefined,
        code_tail: tail,
        code_prefix: invite.code_prefix || `ymt_····${tail}`,
      },
      ...prev.filter((x) => x.id !== invite.id),
    ])
    return plain
  }

  const rememberPlainBriefly = (plain: string) => {
    setLastPlainInvite(plain)
    window.setTimeout(() => setLastPlainInvite(null), 0)
  }

  /** B1：签发 + 剪贴板完整入场链接；Toast 含末四；提示勿发公开群 */
  const onCopyJoinLink = async (roleCtx?: InviteRoleCtx) => {
    setInviteBusy(true)
    try {
      const plain = await issueInvitePlain(roleCtx)
      if (!plain) {
        message.success('已签发入场码（服务未返回明文，请查已签发列表末四）')
        return
      }
      const role = (roleCtx?.slug || '').trim() || undefined
      const url = buildJoinUrl({ code: plain, role })
      await navigator.clipboard.writeText(url)
      rememberPlainBriefly(plain)
      const tail = inviteCodeTail(plain)
      const roleHint = role ? ` · 席 ${role}` : ' · 通用码（未绑席）'
      message.success(`已复制入场链接（末四 ·${tail}${roleHint}）。勿发公开群`)
    } catch (e: unknown) {
      if (isInviteServiceMissing(e)) {
        message.warning('入场码服务尚未开通')
      } else {
        message.error(e instanceof ApiError ? e.message : e instanceof Error ? e.message : '复制入场链接失败')
      }
    } finally {
      setInviteBusy(false)
    }
  }

  /** B2 / F5：同码同 role 入场提示词（仅 ymt_；禁嵌 ymk_） */
  const onCopyBotPrompt = async (roleCtx?: InviteRoleCtx) => {
    setInviteBusy(true)
    try {
      const plain = await issueInvitePlain(roleCtx)
      if (!plain) {
        message.success('已签发入场码（服务未返回明文，请查已签发列表末四）')
        return
      }
      const role = (roleCtx?.slug || '').trim() || undefined
      const prompt = buildBotJoinPrompt({
        code: plain,
        role,
        roleDisplayName: roleCtx?.displayName || undefined,
      })
      await navigator.clipboard.writeText(prompt)
      rememberPlainBriefly(plain)
      const tail = inviteCodeTail(plain)
      const roleHint = role ? ` · 席 ${role}` : ' · 通用码（未绑席）'
      message.success(`已复制入场提示词（末四 ·${tail}${roleHint}；仅含入场码 ymt_）。勿发公开群`)
    } catch (e: unknown) {
      if (isInviteServiceMissing(e)) {
        message.warning('入场码服务尚未开通')
      } else {
        message.error(e instanceof ApiError ? e.message : e instanceof Error ? e.message : '复制入场提示词失败')
      }
    } finally {
      setInviteBusy(false)
    }
  }

  /** B4：裸 ymt_ 降次要 */
  const onCopyBareInvite = async (roleCtx?: InviteRoleCtx) => {
    setInviteBusy(true)
    try {
      const plain = await issueInvitePlain(roleCtx)
      if (!plain) {
        message.success('已签发入场码（服务未返回明文，请查已签发列表末四）')
        return
      }
      await navigator.clipboard.writeText(plain)
      rememberPlainBriefly(plain)
      message.success(`已复制裸入场码（末四 ·${inviteCodeTail(plain)}）。建议改用入场链接/提示词；勿发公开群`)
    } catch (e: unknown) {
      if (isInviteServiceMissing(e)) {
        message.warning('入场码服务尚未开通')
      } else {
        message.error(e instanceof ApiError ? e.message : e instanceof Error ? e.message : '签发失败')
      }
    } finally {
      setInviteBusy(false)
    }
  }

  /** SEAT_REFILL_BIND B3：停工/可补位席签发 ymt_（allow_expand_seat=false）；文案「持补位码入场即占该席」；禁暗示须 force；禁扩编/ymk_ */
  const onCopyRefillInvite = async (roleCtx: InviteRoleCtx) => {
    setInviteBusy(true)
    try {
      if (!teamId) {
        message.warning('尚无团队，请先创建或加入班底')
        return
      }
      const slug = (roleCtx.slug || '').trim()
      if (!slug) {
        message.warning('该席无 slug，无法签发补位入场码')
        return
      }
      const { invite } = await createTeamInvite(
        wsId,
        teamId,
        {
          allow_expand_seat: false,
          role_hint: slug,
          seat_slug: slug,
        },
        crypto.randomUUID(),
      )
      const plain = invite.code || null
      const tail = inviteTail(invite)
      setIssuedInvites((prev) => [
        {
          ...invite,
          code: undefined,
          code_tail: tail,
          code_prefix: invite.code_prefix || `ymt_····${tail}`,
        },
        ...prev.filter((x) => x.id !== invite.id),
      ])
      if (!plain) {
        message.success('已签发补位入场码（服务未返回明文，请查已签发列表末四）')
        return
      }
      await navigator.clipboard.writeText(plain)
      rememberPlainBriefly(plain)
      const name = roleCtx.displayName || slug
      message.success(
        `已复制补位入场码（末四 ·${inviteCodeTail(plain)} · 席 ${name}）。持补位码入场即占该席；非扩编新席。勿发公开群`,
      )
    } catch (e: unknown) {
      if (isInviteServiceMissing(e)) {
        message.warning('入场码服务尚未开通')
      } else {
        message.error(e instanceof ApiError ? e.message : e instanceof Error ? e.message : '复制补位入场码失败')
      }
    } finally {
      setInviteBusy(false)
    }
  }

  /** 顶栏无席：通用码；席卡带 slug */
  const inviteMoreMenu = (roleCtx?: InviteRoleCtx) => ({
    items: [
      {
        key: 'bare',
        label: '仅复制裸入场码（次要）',
        icon: <CopyOutlined />,
        onClick: () => void onCopyBareInvite(roleCtx),
      },
    ],
  })

  const renderOneTapActions = (opts?: {
    roleCtx?: InviteRoleCtx
    size?: 'small' | 'middle'
    primary?: boolean
    /** 席卡内紧凑链接按钮 */
    compact?: boolean
  }) => {
    const size = opts?.size || 'small'
    const roleCtx = opts?.roleCtx
    const tipRole = roleCtx?.slug
      ? `带席位 ${roleCtx.slug}`
      : '通用码（未选席；Bot 入场时自动选空席）。也可先点席卡上的一键按钮'
    if (opts?.compact) {
      return (
        <Space size={4} wrap onClick={(e) => e.stopPropagation()}>
          <Tooltip title={`复制入场链接 · ${tipRole}；勿发公开群`}>
            <Button
              size="small"
              type="link"
              style={{ padding: 0 }}
              icon={<LinkOutlined />}
              loading={inviteBusy}
              onClick={() => void onCopyJoinLink(roleCtx)}
            >
              入场链接
            </Button>
          </Tooltip>
          <Tooltip title={`复制入场提示词（含入场码 ymt_，不含身份钥匙）· ${tipRole}`}>
            <Button
              size="small"
              type="link"
              style={{ padding: 0 }}
              icon={<FileTextOutlined />}
              loading={inviteBusy}
              onClick={() => void onCopyBotPrompt(roleCtx)}
            >
              入场提示词
            </Button>
          </Tooltip>
          <Dropdown menu={inviteMoreMenu(roleCtx)} trigger={['click']}>
            <Button size="small" type="link" style={{ padding: 0 }} icon={<MoreOutlined />} aria-label="更多入场选项" />
          </Dropdown>
        </Space>
      )
    }
    return (
      <Space size={8} wrap>
        <Tooltip title={`Admin 一键入场链接 · ${tipRole}；勿发公开群`}>
          <Button
            size={size}
            icon={<LinkOutlined />}
            loading={inviteBusy}
            onClick={() => void onCopyJoinLink(roleCtx)}
          >
            复制入场链接
          </Button>
        </Tooltip>
        <Tooltip title={`复制入场提示词（含入场码 ymt_，不含身份钥匙）· ${tipRole}；勿发公开群`}>
          <Button
            size={size}
            icon={<FileTextOutlined />}
            loading={inviteBusy}
            onClick={() => void onCopyBotPrompt(roleCtx)}
          >
            复制入场提示词
          </Button>
        </Tooltip>
        <Dropdown menu={inviteMoreMenu(roleCtx)} trigger={['click']}>
          <Tooltip title="次要：仅复制裸 ymt_ 码">
            <Button size={size} icon={<MoreOutlined />} aria-label="更多入场选项" />
          </Tooltip>
        </Dropdown>
      </Space>
    )
  }


  if (loading || teamLoading) {
    return (
      <div className="acb-page acb-bots" style={{ textAlign: 'center', paddingTop: 48 }}>
        <Spin />
      </div>
    )
  }

  const migrateCollapse = (
    <Collapse
      size="small"
      items={[
        {
          key: 'migrate',
          label: '换机怎么续',
          children: (
            <ol style={{ margin: '0', paddingLeft: 18, fontSize: 13 }}>
              <li>新机打开同一工作区</li>
              <li>
                绑定到同一席位（贴 Bot Key，或本机「席位进入」指定席位）
              </li>
              <li>拉取交接摘要：记忆 / Todo / 开放 Issue</li>
            </ol>
          ),
        },
      ]}
    />
  )

  /* F1.1：无席绑定时主空态只有「加入班底」 */
  if (!hasAnyBinding) {
    return (
      <div className="acb-page acb-bots">
        <div className="acb-page-bar">
          <Typography.Title level={4} style={{ margin: 0 }}>
            Bot 席位
          </Typography.Title>
          <Space size={8} wrap>
            {teamLabel}
            <Button size="small" icon={<PlusOutlined />} onClick={() => setCreateFromTplOpen(true)}>
              从模板创建
            </Button>
            {renderOneTapActions()}
          </Space>
        </div>
        <Typography.Paragraph type="secondary" className="acb-bots-lead">
          空手上场：复制入场包发给对方；对方入场时自动获得身份，进空席或申请新席。活跃席不可顶替。设置·接入单把签 Key 仅作旁路/救场。
        </Typography.Paragraph>
        {migrateCollapse}
        <TeamPinsPanel wsId={wsId} teamId={teamId} />
        <div className="acb-empty-panel" style={{ marginTop: 8 }}>
          <Empty
            image={Empty.PRESENTED_IMAGE_SIMPLE}
            description="尚未绑定席位"
          >
            <Button type="primary" icon={<UserAddOutlined />} onClick={() => setJoinOpen(true)}>
              加入班底
            </Button>
          </Empty>
        </div>
        {/* 空态不重复列邀请；有绑定时下方统一列表 */}

        <CreateTeamFromTemplateModal
          open={createFromTplOpen}
          wsId={wsId}
          onClose={() => setCreateFromTplOpen(false)}
          onCreated={onTeamCreatedFromTemplate}
        />
        <JoinTeamWizard
          open={joinOpen}
          onClose={() => {
            setJoinOpen(false)
            setJoinPrefill(null)
          }}
          wsId={wsId}
          initialCode={joinPrefill?.code}
          initialRole={joinPrefill?.role}
          initialSeat={joinPrefill?.seat}
          onEntered={({ memberId, roleId }) => {
            void load()
            void refreshTeamRoster()
            const role =
              (roleId ? roles.find((r) => r.id === roleId) : null) ||
              roles.find((r) => r.claim?.member_id === memberId) ||
              null
            if (role) setDrawerRole(role)
            else {
              void listBotRoles(wsId).then((r) => {
                const list = r.bot_roles || []
                const hit =
                  (roleId ? list.find((x) => x.id === roleId) : null) ||
                  list.find((x) => x.claim?.member_id === memberId)
                if (hit) setDrawerRole(hit)
              })
            }
          }}
        />
        <Drawer
          title="席位占用中"
          width={420}
          open={conflictOpen}
          onClose={() => setConflictOpen(false)}
          destroyOnClose
        >
          {conflictInfo ? (
            <SeatConflictPanel
              info={conflictInfo}
              onPickOther={() => {
                setConflictOpen(false)
                setJoinOpen(true)
              }}
              onRequestNew={() => {
                setConflictOpen(false)
                setJoinOpen(true)
              }}
              onBack={() => setConflictOpen(false)}
              showRequestNew
            />
          ) : null}
        </Drawer>
      </div>
    )
  }

  return (
    <div className="acb-page acb-bots">
      <div className="acb-page-bar">
        <Typography.Title level={4} style={{ margin: 0 }}>
          Bot 席位
        </Typography.Title>
        <Space size={8} wrap>
          {teamLabel}
          <Button size="small" icon={<PlusOutlined />} onClick={() => setCreateFromTplOpen(true)}>
            从模板创建
          </Button>
          <Button size="small" type="primary" icon={<UserAddOutlined />} onClick={() => setJoinOpen(true)}>
            加入班底
          </Button>
          {emptySeats.length > 0 ? (
            <Button
              size="small"
              type="primary"
              ghost
              icon={<CopyOutlined />}
              loading={inviteBusy}
              onClick={() => openCrewPack(false)}
            >
              复制全班入场包
            </Button>
          ) : null}
          {renderOneTapActions()}
        </Space>
      </div>
      <Typography.Paragraph type="secondary" className="acb-bots-lead">
        席位 = 角色坑。主路径：复制入场包（或「入场链接 / 入场提示词」，只含 ymt_ 与席位）；对方入场时自动获得身份。勿发公开群。设置·接入单把签 ymk_ 仅作旁路/救场，不是建队主路径。换机只换演员；续命靠服务端交接。
      </Typography.Paragraph>

      {migrateCollapse}

      <TeamPinsPanel wsId={wsId} teamId={teamId} />

      {visibleRoles.length ? (
        <div className="acb-bots-seat-grid">
          {visibleRoles.map((role) => {
            const occ = seatOccForRole(role)
            const mid = occ.occupantId
            const bound = occ.occupantLabel || botName(mid)
            const hint = keyHintForMember(mid)
            const busy = occ.occupied
            const left = occ.kind === 'left'
            const name = roleDisplay(role)
            return (
              <Card
                key={role.id}
                size="small"
                hoverable
                className="acb-bots-seat-card"
                onClick={() => openSeat(role)}
              >
                {/* S1：中文席名 + 占用态（与向导同步）+ 末四 */}
                <div className="acb-bots-seat-title-row" style={{ display: 'flex', alignItems: 'center', gap: 8, minHeight: 36 }}>
                  <Typography.Text
                    strong
                    className="acb-bots-seat-title"
                    style={{ flex: 1, minWidth: 0 }}
                    ellipsis
                    title={name}
                  >
                    {name}
                  </Typography.Text>
                  <Tag
                    color={busy ? 'blue' : left ? 'orange' : 'default'}
                    style={{ margin: 0, flexShrink: 0 }}
                  >
                    {busy ? '占用中' : left ? '可补位' : '空闲'}
                  </Tag>
                  {hint ? (
                    <Typography.Text type="secondary" style={{ fontSize: 12, flexShrink: 0 }}>
                      {hint}
                    </Typography.Text>
                  ) : null}
                </div>

                <div className="acb-bots-seat-actor">
                  {busy ? (
                    <span>
                      演员 · {bound || '已绑定'}
                      {role.claim?.claimed_at ? (
                        <Typography.Text type="secondary" style={{ marginLeft: 8, fontSize: 12 }}>
                          {formatWikiTime(role.claim.claimed_at)}
                        </Typography.Text>
                      ) : null}
                    </span>
                  ) : left ? (
                    <Typography.Text type="secondary">停工席 · 可补位</Typography.Text>
                  ) : (
                    <Typography.Text type="secondary">等待绑定</Typography.Text>
                  )}
                </div>

                <div
                  style={{ marginTop: 8, display: 'flex', flexWrap: 'wrap', gap: 6, alignItems: 'center' }}
                  onClick={(e) => e.stopPropagation()}
                >
                  <Button size="small" type="link" style={{ padding: 0 }} onClick={() => openSeat(role)}>
                    查看交接
                  </Button>
                  {left ? (
                    <>
                      <Tooltip title="签发该席 ymt_ 补位入场码；持码入场即占该席（非扩编）；禁嵌 ymk_；勿发公开群">
                        <Button
                          size="small"
                          type="link"
                          style={{ padding: 0 }}
                          icon={<CopyOutlined />}
                          loading={inviteBusy}
                          onClick={() =>
                            void onCopyRefillInvite({
                              slug: role.slug || undefined,
                              displayName: roleDisplay(role),
                            })
                          }
                        >
                          复制补位入场码
                        </Button>
                      </Tooltip>
                      <Select
                        size="small"
                        placeholder="指派 Bot"
                        style={{ minWidth: 120 }}
                        loading={claiming === role.id}
                        options={freeBots.map((b) => ({ value: b.member_id, label: b.display_name }))}
                        onChange={(v) => void onClaim(role, v)}
                        popupMatchSelectWidth={false}
                      />
                    </>
                  ) : (
                    renderOneTapActions({
                      compact: true,
                      roleCtx: {
                        slug: role.slug || undefined,
                        displayName: roleDisplay(role),
                      },
                    })
                  )}
                  {busy ? (
                    <Button
                      size="small"
                      type="link"
                      style={{ padding: 0 }}
                      icon={<SendOutlined />}
                      onClick={() => openDispatch(role)}
                    >
                      下发任务
                    </Button>
                  ) : null}
                  {busy && teamId && occ.seat?.id ? (
                    <>
                      <Popconfirm
                        title="停工此席？"
                        description="原演员解除占用，席位变为「停工席 · 可补位」。活跃禁顶替不变。"
                        okText="停工"
                        cancelText="取消"
                        onConfirm={() => void onStop(role)}
                      >
                        <Button
                          size="small"
                          type="link"
                          style={{ padding: 0 }}
                          onClick={(e) => e.stopPropagation()}
                        >
                          停工
                        </Button>
                      </Popconfirm>
                      <Popconfirm
                        title="离职解绑此席？"
                        description="原演员解除绑定，席位变为可补位（left）。勿与入场码/Bot 入场混淆。"
                        okText="离职解绑"
                        cancelText="取消"
                        okButtonProps={{ danger: true }}
                        onConfirm={() => void onLeaveUnbind(role)}
                      >
                        <Button
                          size="small"
                          type="link"
                          danger
                          style={{ padding: 0 }}
                          onClick={(e) => e.stopPropagation()}
                        >
                          离职解绑
                        </Button>
                      </Popconfirm>
                    </>
                  ) : !busy && !left ? (
                    <Select
                      size="small"
                      placeholder="指派 Bot"
                      style={{ minWidth: 120 }}
                      loading={claiming === role.id}
                      options={freeBots.map((b) => ({ value: b.member_id, label: b.display_name }))}
                      onChange={(v) => void onClaim(role, v)}
                      popupMatchSelectWidth={false}
                    />
                  ) : null}
                </div>
                {left ? (
                  <Typography.Text
                    type="secondary"
                    style={{ display: 'block', marginTop: 6, fontSize: 12 }}
                    onClick={(e) => e.stopPropagation()}
                  >
                    持补位码入场即占该席；活跃席仍禁顶替。补位 ≠ 扩编新席。
                  </Typography.Text>
                ) : null}
              </Card>
            )
          })}
        </div>
      ) : (
        <div className="acb-empty-panel">
          <Empty
            image={Empty.PRESENTED_IMAGE_SIMPLE}
            description={teamId ? '本队暂无席位' : '暂无席位'}
          >
            <Button type="primary" onClick={() => setJoinOpen(true)}>
              加入班底
            </Button>
          </Empty>
        </div>
      )}

      {issuedInvites.length ? (
        <Card size="small" className="acb-bots-invite-list" title="已签发入场码" style={{ marginBottom: 12 }}>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
            {issuedInvites.slice(0, 12).map((inv) => {
              const st = (inv.status || (inv.revoked_at ? 'revoked' : 'active')).toLowerCase()
              const revoked = st === 'revoked'
              return (
                <div
                  key={inv.id}
                  style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: 8,
                    minHeight: 32,
                    flexWrap: 'wrap',
                  }}
                >
                  <Typography.Text style={{ fontFamily: 'var(--acb-font-mono, monospace)', fontSize: 13 }}>
                    末四 ·{inviteTail(inv)}
                  </Typography.Text>
                  <Tag color={revoked ? 'default' : st === 'active' ? 'green' : 'orange'} style={{ margin: 0 }}>
                    {inviteStatusCn(st)}
                  </Tag>
                  {!revoked ? (
                    <Popconfirm
                      title="吊销此入场码？"
                      description="吊销后无法再用该码加入；已进席不受影响。"
                      okText="吊销"
                      cancelText="取消"
                      okButtonProps={{ danger: true }}
                      onConfirm={() => void onRevokeInvite(inv)}
                    >
                      <Button
                        size="small"
                        type="link"
                        danger
                        style={{ padding: 0 }}
                        loading={revokingId === inv.id}
                      >
                        吊销
                      </Button>
                    </Popconfirm>
                  ) : (
                    <Typography.Text type="secondary" style={{ fontSize: 12 }}>
                      —
                    </Typography.Text>
                  )}
                </div>
              )
            })}
          </div>
        </Card>
      ) : null}

      <Collapse
        size="small"
        items={[
          {
            key: 'bind-guide',
            label: '高级 · 绑定此机 / CLI',
            children: (
              <Typography.Paragraph style={{ marginBottom: 0, fontSize: 13 }}>
                看板无「本机设备」清单。新机：打开本工作区 → 配置同一身份钥匙，或本机席位进入（英文代号仅 CLI，界面用中文席位名）→ 再拉交接。
                「指派 Bot」仅供管理员在看板内把已有 Bot 挂到空席。
              </Typography.Paragraph>
            ),
          },
          {
            key: 'advanced',
            label: '高级 · Bot 成员与 Key',
            children: (
              <Space direction="vertical" style={{ width: '100%' }} size="middle">
                <Alert
                  type="warning"
                  showIcon
                  message="export / import 是可选备份"
                  description="日常换机：绑同一席位 → 拉交接。不要把 JSON 当主路径。"
                />
                <div>
                  <Typography.Text strong>Bot 成员</Typography.Text>
                  {bots.length ? (
                    <Table
                      className="acb-list-table"
                      style={{ marginTop: 8 }}
                      rowKey="member_id"
                      size="small"
                      pagination={false}
                      dataSource={bots}
                      columns={[
                        {
                          title: '名称',
                          dataIndex: 'display_name',
                          render: (v: string, row) => (
                            <Link to={`/w/${wsId}/settings/bots/${row.member_id}`}>{v}</Link>
                          ),
                        },
                        {
                          title: '当前席位',
                          render: (_, row) => row.claimed_role?.name || '—',
                        },
                        {
                          title: '身份钥匙',
                          width: 100,
                          render: (_, row) => keyHintForMember(row.member_id) || '—',
                        },
                      ]}
                    />
                  ) : (
                    <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无 Bot 成员" />
                  )}
                </div>
              </Space>
            ),
          },
        ]}
      />


      <Modal
        title="下发任务"
        open={dispatchOpen}
        onCancel={() => (!dispatchBusy ? setDispatchOpen(false) : undefined)}
        onOk={() => void onDispatchCreate()}
        okText="创建并指派"
        cancelText="取消"
        confirmLoading={dispatchBusy}
        destroyOnHidden
        centered
      >
        {dispatchTarget ? (
          <Space direction="vertical" style={{ width: '100%' }} size="middle">
            <Typography.Text type="secondary" style={{ fontSize: 13 }}>
              负责人预填：{dispatchTarget.botName}（席位「{dispatchTarget.seatName}」）
            </Typography.Text>
            <Input
              placeholder="Issue 标题（可空，将用默认文案）"
              value={dispatchTitle}
              onChange={(e) => setDispatchTitle(e.target.value)}
              onPressEnter={() => void onDispatchCreate()}
              autoFocus
            />
            <Checkbox checked={dispatchPush} onChange={(e) => setDispatchPush(e.target.checked)}>
              立即推送
            </Checkbox>
            <Typography.Text type="secondary" style={{ fontSize: 12 }}>
              勾选后经现有 Webhook 扇出；失败仅提示，不撤销建单。默认关闭。
            </Typography.Text>
          </Space>
        ) : null}
      </Modal>

      <SeatHandoffDrawer
        open={Boolean(drawerRole)}
        onClose={() => setDrawerRole(null)}
        wsId={wsId}
        role={drawerRole}
        bots={bots}
        keyHint={
          drawerRole
            ? keyHintForMember(seatOccForRole(drawerRole).occupantId || drawerRole.claim?.member_id)
            : undefined
        }
        occupancyKind={drawerRole ? seatOccForRole(drawerRole).kind : null}
        occupantMemberId={drawerRole ? seatOccForRole(drawerRole).occupantId : null}
        occupantLabel={drawerRole ? seatOccForRole(drawerRole).occupantLabel : null}
        canLeave={Boolean(drawerRole && teamId && seatOccForRole(drawerRole).seat?.id)}
        onStop={drawerRole ? () => void onStop(drawerRole) : undefined}
        onLeaveUnbind={drawerRole ? () => void onLeaveUnbind(drawerRole) : undefined}
        onCopyRefill={
          drawerRole
            ? () =>
                void onCopyRefillInvite({
                  slug: drawerRole.slug || undefined,
                  displayName: roleDisplay(drawerRole),
                })
            : undefined
        }
        refillBusy={inviteBusy}
      />


        <CrewPackRevealModal
          open={crewPackOpen}
          onClose={() => {
            setCrewPackOpen(false)
            setCrewPackAfterCreate(false)
            setCrewPackSeats([])
          }}
          workspaceId={wsId}
          teamId={teamId || ''}
          teamName={currentTeamName || '未命名团队'}
          seats={crewPackSeats.length ? crewPackSeats : teamSeats}
          afterCreate={crewPackAfterCreate}
          onIssued={onCrewPackIssued}
        />

      <CreateTeamFromTemplateModal
        open={createFromTplOpen}
        wsId={wsId}
        onClose={() => setCreateFromTplOpen(false)}
        onCreated={onTeamCreatedFromTemplate}
      />
      <JoinTeamWizard
        open={joinOpen}
        onClose={() => {
          setJoinOpen(false)
          setJoinPrefill(null)
        }}
        wsId={wsId}
        initialCode={joinPrefill?.code}
        initialRole={joinPrefill?.role}
        initialSeat={joinPrefill?.seat}
        onEntered={({ memberId, roleId }) => {
          void load()
          void refreshTeamRoster()
          void listBotRoles(wsId).then((r) => {
            const list = r.bot_roles || []
            const hit =
              (roleId ? list.find((x) => x.id === roleId) : null) ||
              list.find((x) => x.claim?.member_id === memberId)
            if (hit) setDrawerRole(hit)
          })
        }}
      />

      <Drawer
        title="席位占用中"
        width={420}
        open={conflictOpen}
        onClose={() => setConflictOpen(false)}
        destroyOnClose
      >
        {conflictInfo ? (
          <SeatConflictPanel
            info={conflictInfo}
            onPickOther={() => {
              setConflictOpen(false)
              setJoinOpen(true)
            }}
            onRequestNew={() => {
              setConflictOpen(false)
              setJoinOpen(true)
            }}
            onBack={() => setConflictOpen(false)}
            showRequestNew
          />
        ) : null}
      </Drawer>
    </div>
  )
}
