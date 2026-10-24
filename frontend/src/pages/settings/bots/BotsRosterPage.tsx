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
import { CopyOutlined, PlusOutlined, SendOutlined, UserAddOutlined } from '@ant-design/icons'
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
  createTeamInvite,
  getRoster,
  inviteStatusCn,
  inviteTail,
  isInviteServiceMissing,
  isSeatConflict,
  listTeamInvites,
  revokeTeamInvite,
  seatOccupancy,
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
  const [searchParams] = useSearchParams()
  const forceEmpty = searchParams.get('f1empty') === '1'
  const { teams, teamId, setTeamId, loading: teamLoading } = useCurrentTeam(wsId)
  const [loading, setLoading] = useState(true)
  const [bots, setBots] = useState<BotRosterItem[]>([])
  const [roles, setRoles] = useState<BotRole[]>([])
  const [keys, setKeys] = useState<ApiKeyRow[]>([])
  const [drawerRole, setDrawerRole] = useState<BotRole | null>(null)
  const [claiming, setClaiming] = useState<string | null>(null)
  const [joinOpen, setJoinOpen] = useState(false)
  const [createFromTplOpen, setCreateFromTplOpen] = useState(false)
  const [conflictOpen, setConflictOpen] = useState(false)
  const [conflictInfo, setConflictInfo] = useState<SeatConflictInfo | null>(null)
  /** 与向导同源的团队席位占用（N4 同步） */
  const [teamSeats, setTeamSeats] = useState<TeamSeat[]>([])
  const [issuedInvites, setIssuedInvites] = useState<TeamInvite[]>([])
  const [inviteBusy, setInviteBusy] = useState(false)
  /** 签发当次明文，只存活到复制成功后的一次 Toast；列表只留末四 */
  const [lastPlainInvite, setLastPlainInvite] = useState<string | null>(null)
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

  useEffect(() => {
    void load()
  }, [wsId])

  /** 顶栏 CurrentTeam 换队：清空席/邀请缓存并 refetch；记忆墙由 TeamPinsPanel 按 teamId 换池（P5） */
  useEffect(() => {
    setTeamSeats([])
    setIssuedInvites([])
    if (!wsId || !teamId) return
    let cancelled = false
    void (async () => {
      try {
        const rost = await getRoster(wsId, teamId)
        if (!cancelled) setTeamSeats(rost.roster || [])
      } catch {
        if (!cancelled) setTeamSeats([])
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

  const onTeamCreatedFromTemplate = (team: Team, _roster: TeamSeat[]) => {
    setTeamId(team.id)
    void load()
  }

  const currentTeamName = teams.find((t) => t.id === teamId)?.name
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
  const hasAnyBinding = useMemo(() => {
    if (forceEmpty) return false
    if (roles.some((r) => Boolean(r.claim?.member_id))) return true
    return teamSeats.some((s) => seatOccupancy(s) === 'active')
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


  /** D5：与向导同源读 team_seats.status（active/empty/left）；无席时再回落 claim */
  const seatOccForRole = (role: BotRole) => {
    const claimMid = role.claim?.member_id || null
    const seat =
      teamSeats.find((s) => s.bot_role_id && s.bot_role_id === role.id) ||
      teamSeats.find((s) => s.role_slug && s.role_slug === role.slug)
    const occ: 'active' | 'empty' | 'left' | null = seat ? seatOccupancy(seat) : null
    let kind: 'active' | 'empty' | 'left'
    if (occ) kind = occ
    else kind = claimMid ? 'active' : 'empty'
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


  const onRelease = async (role: BotRole) => {
    try {
      await releaseBotRole(wsId, role.id, crypto.randomUUID())
      message.success(`已释放席位「${roleDisplay(role)}」`)
      void load()
    } catch {
      /* api client toasts */
    }
  }

  const onClaim = async (role: BotRole, memberId: string) => {
    setClaiming(role.id)
    try {
      await claimBotRole(wsId, role.id, memberId, crypto.randomUUID(), { skipToast: true })
      message.success(`已将「${botName(memberId)}」绑到席位「${roleDisplay(role)}」`)
      void load()
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

  const onCopyInvite = async () => {
    if (!teamId) {
      message.warning('尚无团队，请先拉班底 / seed')
      return
    }
    setInviteBusy(true)
    try {
      const { invite } = await createTeamInvite(
        wsId,
        teamId,
        { allow_expand_seat: true },
        crypto.randomUUID(),
      )
      const plain = invite.code
      if (plain) {
        await navigator.clipboard.writeText(plain)
        setLastPlainInvite(plain)
        message.success(`已复制入场码（末四 ·${inviteTail(invite)}），请妥善转发`)
        // 之后列表只留末四：剥明文
        setIssuedInvites((prev) => [
          {
            ...invite,
            code: undefined,
            code_tail: inviteTail(invite),
            code_prefix: invite.code_prefix || `ymt_····${inviteTail(invite)}`,
          },
          ...prev.filter((x) => x.id !== invite.id),
        ])
        // 明文仅 Toast 一次；状态里不长期存
        window.setTimeout(() => setLastPlainInvite(null), 0)
      } else {
        message.success('已签发入场码（服务未返回明文，请查已签发列表末四）')
        setIssuedInvites((prev) => [invite, ...prev])
      }
    } catch (e: unknown) {
      if (isInviteServiceMissing(e)) {
        message.warning('入场码服务尚未开通')
      } else {
        message.error(e instanceof ApiError ? e.message : '签发失败')
      }
    } finally {
      setInviteBusy(false)
    }
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
            <Tooltip title="Admin 签发团队入场码（ymt_）">
              <Button size="small" icon={<CopyOutlined />} loading={inviteBusy} onClick={() => void onCopyInvite()}>
                复制入场码
              </Button>
            </Tooltip>
          </Space>
        </div>
        <Typography.Paragraph type="secondary" className="acb-bots-lead">
          空手上场：贴身份钥匙与团队入场码，进空席或申请新席。活跃席不可顶替。
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
          onClose={() => setJoinOpen(false)}
          wsId={wsId}
          onEntered={({ memberId }) => {
            void load()
            const role = roles.find((r) => r.claim?.member_id === memberId) || null
            if (role) setDrawerRole(role)
            else {
              // 刷新后再开交接：按 member 找
              void listBotRoles(wsId).then((r) => {
                const hit = (r.bot_roles || []).find((x) => x.claim?.member_id === memberId)
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
          <Tooltip title="Admin：一次明文 + Toast，之后只见末四">
            <Button size="small" icon={<CopyOutlined />} loading={inviteBusy} onClick={() => void onCopyInvite()}>
              复制入场码
            </Button>
          </Tooltip>
        </Space>
      </div>
      <Typography.Paragraph type="secondary" className="acb-bots-lead">
        席位 = 角色坑。换机只换演员；续命靠服务端交接摘要，不靠本机文件。
      </Typography.Paragraph>

      {migrateCollapse}

      <TeamPinsPanel wsId={wsId} teamId={teamId} />

      {roles.length ? (
        <div className="acb-bots-seat-grid">
          {roles.map((role) => {
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
                  {busy && role.claim?.member_id ? (
                    <Button size="small" type="link" danger style={{ padding: 0 }} onClick={() => void onRelease(role)}>
                      释放
                    </Button>
                  ) : !busy ? (
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
              </Card>
            )
          })}
        </div>
      ) : (
        <div className="acb-empty-panel">
          <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无席位">
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
                          title: 'Key',
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
      />

      <CreateTeamFromTemplateModal
        open={createFromTplOpen}
        wsId={wsId}
        onClose={() => setCreateFromTplOpen(false)}
        onCreated={onTeamCreatedFromTemplate}
      />
      <JoinTeamWizard
        open={joinOpen}
        onClose={() => setJoinOpen(false)}
        wsId={wsId}
        onEntered={({ memberId }) => {
          void load().then(() => {
            void listBotRoles(wsId).then((r) => {
              const hit = (r.bot_roles || []).find((x) => x.claim?.member_id === memberId)
              if (hit) setDrawerRole(hit)
            })
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
