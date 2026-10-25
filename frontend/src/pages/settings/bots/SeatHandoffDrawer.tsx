import { useEffect, useMemo, useState } from 'react'
import { Alert, Button, Collapse, Drawer, Empty, Popconfirm, Space, Spin, Tag, Typography, message } from 'antd'
import {
  CopyOutlined,
  ReloadOutlined,
  StopOutlined,
  UserDeleteOutlined,
} from '@ant-design/icons'
import {
  getHandoff,
  type BotHandoff,
  type BotRole,
  type BotRosterItem,
} from '../../../api/bots'
import { formatShanghai, formatWikiTime } from '../../../utils/time'
import {
  buildHandoffCardText,
  openTodos,
  pickLastFocus,
  roleDisplay,
  seatOccupancyTagLabel,
  statusCn,
  type SeatCardMeta,
} from './seatHandoffText'

type Props = {
  open: boolean
  onClose: () => void
  wsId: string
  /** X-R3：项目上下文下 handoff Issues 仅展示该 project_id */
  projectId?: string | null
  role: BotRole | null
  bots: BotRosterItem[]
  /** Key 提示，如「末四 ·7624」；勿传完整 Key */
  keyHint?: string
  /** 与名册卡同源：team_seats.status（D5）；缺省回落 claim */
  occupancyKind?: 'active' | 'empty' | 'left' | null
  /** 占用者 member_id（可来自 team seat，非仅 claim） */
  occupantMemberId?: string | null
  occupantLabel?: string | null
  /** SEAT_LEAVE L3：占用中可停工/离职（需有 team seat） */
  canLeave?: boolean
  onStop?: () => void
  onLeaveUnbind?: () => void
  /** SEAT_REFILL：可补位席复制补位入场码（ymt_；非扩编） */
  onCopyRefill?: () => void
  refillBusy?: boolean
  /** IQ1-4：在席查用已有身份钥匙（Lookup；禁 mint） */
  onLookupEntry?: () => void
  lookupBusy?: boolean
}

export default function SeatHandoffDrawer({
  open,
  onClose,
  wsId,
  projectId = null,
  role,
  bots,
  keyHint,
  occupancyKind = null,
  occupantMemberId = null,
  occupantLabel = null,
  canLeave = false,
  onStop,
  onLeaveUnbind,
  onCopyRefill,
  refillBusy = false,
  onLookupEntry,
  lookupBusy = false,
}: Props) {
  const [loading, setLoading] = useState(false)
  const [handoff, setHandoff] = useState<BotHandoff | null>(null)

  const claimMid = role?.claim?.member_id || null
  const kind =
    occupancyKind || (claimMid ? 'active' : 'empty')
  const memberId = (kind === 'active' ? occupantMemberId || claimMid : null) || null
  const boundBot = memberId ? bots.find((b) => b.member_id === memberId) : undefined
  const boundLabel =
    boundBot?.display_name ||
    occupantLabel ||
    (memberId ? '已绑定' : kind === 'left' ? '停工可补' : '空闲（从未占用）')

  const seatMeta: SeatCardMeta = useMemo(
    () => ({
      roleName: role ? roleDisplay(role) : '—',
      roleSlug: role?.slug,
      boundLabel,
      keyHint,
      claimedAt: role?.claim?.claimed_at,
    }),
    [role, boundLabel, keyHint],
  )

  const load = async () => {
    if (!memberId) {
      setHandoff(null)
      return
    }
    setLoading(true)
    try {
      const res = await getHandoff(wsId, memberId)
      setHandoff(res.handoff)
    } catch (e: unknown) {
      message.error(e instanceof Error ? e.message : '加载交接失败')
      setHandoff(null)
    } finally {
      setLoading(false)
    }
  }

  // X-R1/R2：切仓或换席先清空 handoff，禁止他仓残留；cancelled 防竞态
  useEffect(() => {
    let cancelled = false
    setHandoff(null)
    if (!(open && role && memberId)) {
      setLoading(false)
      return () => {
        cancelled = true
      }
    }
    setLoading(true)
    ;(async () => {
      try {
        const res = await getHandoff(wsId, memberId)
        if (cancelled) return
        setHandoff(res.handoff)
      } catch (e: unknown) {
        if (cancelled) return
        message.error(e instanceof Error ? e.message : '加载交接失败')
        setHandoff(null)
      } finally {
        if (!cancelled) setLoading(false)
      }
    })()
    return () => {
      cancelled = true
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [open, role?.id, memberId, wsId])

  const focus = pickLastFocus(handoff?.memories, role?.slug)
  const todos = openTodos(handoff?.todos)
  // X-R3 / B1：有项目上下文时仅展示当前项目 Issue，不混他项目 YM-*
  const issues = useMemo(() => {
    const all = handoff?.issues || []
    if (!projectId) return all
    return all.filter((i) => i.project_id === projectId)
  }, [handoff?.issues, projectId])

  const cardText = buildHandoffCardText({
    seat: seatMeta,
    handoff,
    focus,
    todos,
    issues,
  })

  const copyCard = async () => {
    try {
      await navigator.clipboard.writeText(cardText)
      message.success('已复制交接卡（不含 Key）')
    } catch {
      message.error('复制失败，请手动选择下方文本')
    }
  }

  return (
    <Drawer
      title={role ? `席位 · ${roleDisplay(role)}` : '席位交接'}
      width={440}
      open={open}
      onClose={onClose}
      destroyOnClose
      extra={
        <Space size={4}>
          <Button size="small" icon={<ReloadOutlined />} disabled={!memberId} onClick={() => void load()}>
            刷新
          </Button>
          <Button size="small" type="primary" icon={<CopyOutlined />} onClick={() => void copyCard()}>
            复制交接卡
          </Button>
        </Space>
      }
    >
      {!role ? (
        <Empty description="未选择席位" />
      ) : (
        <Space direction="vertical" size="middle" style={{ width: '100%' }}>
          <div>
            <Typography.Text type="secondary" style={{ fontSize: 12 }}>
              席位 = 角色坑；换机换演员，不换席位
            </Typography.Text>
            <div
              style={{ marginTop: 8, display: 'flex', flexWrap: 'wrap', gap: 6, alignItems: 'center' }}
              title={roleDisplay(role)}
            >
              <Tag color={kind === 'active' ? 'blue' : kind === 'left' ? 'orange' : 'default'}>
                {seatOccupancyTagLabel(kind, kind === 'active' ? boundLabel : null)}
              </Tag>
              {keyHint ? <Tag>{keyHint}</Tag> : null}
            </div>
            {kind === 'active' ? (
              <Space size={8} wrap style={{ marginTop: 10 }}>
                {onLookupEntry ? (
                  <Button
                    size="small"
                    loading={lookupBusy}
                    onClick={() => onLookupEntry()}
                  >
                    查用身份钥匙
                  </Button>
                ) : null}
                {canLeave ? (
                  <>
                    <Popconfirm
                      title="停工此席？"
                      description="原演员解除占用，席位变为「停工可补」。"
                      okText="停工"
                      cancelText="取消"
                      onConfirm={() => onStop?.()}
                    >
                      <Button size="small" icon={<StopOutlined />}>
                        停工
                      </Button>
                    </Popconfirm>
                    <Popconfirm
                      title="离职解绑此席？"
                      description="原演员解除绑定，席位变为停工可补（left）。"
                      okText="离职解绑"
                      cancelText="取消"
                      okButtonProps={{ danger: true }}
                      onConfirm={() => onLeaveUnbind?.()}
                    >
                      <Button size="small" danger icon={<UserDeleteOutlined />}>
                        离职解绑
                      </Button>
                    </Popconfirm>
                  </>
                ) : null}
              </Space>
            ) : null}
            {role.claim?.claimed_at ? (
              <Typography.Paragraph type="secondary" style={{ margin: '6px 0 0', fontSize: 12 }}>
                认领于 {formatWikiTime(role.claim.claimed_at)}（{formatShanghai(role.claim.claimed_at)}）
              </Typography.Paragraph>
            ) : null}
            {role.responsibilities ? (
              <Typography.Paragraph style={{ margin: '8px 0 0', fontSize: 13 }}>
                {role.responsibilities}
              </Typography.Paragraph>
            ) : null}
          </div>

          {kind === 'left' ? (
            <Alert
              type="warning"
              showIcon
              message="停工可补"
              description={
                <Space direction="vertical" size={8} style={{ width: '100%' }}>
                  <span>持补位码入场即占该席；活跃席仍禁顶替。补位 ≠ 扩编新席（勿混 ymk_）。也可指派已有 Bot。占席后才有交接摘要。</span>
                  {onCopyRefill ? (
                    <Button
                      size="small"
                      type="primary"
                      icon={<CopyOutlined />}
                      loading={refillBusy}
                      onClick={() => onCopyRefill()}
                    >
                      复制补位入场码
                    </Button>
                  ) : null}
                </Space>
              }
            />
          ) : !memberId ? (
            <Alert
              type="info"
              showIcon
              message="空闲（从未占用）"
              description={
                <span>
                  尚未有 Bot 进本席。绑定后再看交接。新机：打开同一工作区 → 贴 Key，或本机 seat enter 绑到席位「
                  {roleDisplay(role)}」→{' '}
                  <Typography.Text code>ymbot handoff get</Typography.Text>
                  <Typography.Text type="secondary" style={{ display: 'block', marginTop: 6, fontSize: 12 }}>
                    （高级 CLI 见名册「绑定此机」折叠区）
                  </Typography.Text>
                </span>
              }
            />
          ) : loading && !handoff ? (
            <div style={{ textAlign: 'center', padding: 24 }}>
              <Spin />
            </div>
          ) : (
            <>
              <section>
                <Typography.Text strong>当前焦点</Typography.Text>
                <div style={{ marginTop: 6, fontSize: 14 }}>
                  {focus || <Typography.Text type="secondary">暂无焦点记忆</Typography.Text>}
                </div>
              </section>

              <section>
                <Typography.Text strong>进行中 Todos（{todos.length}）</Typography.Text>
                <div style={{ marginTop: 6 }}>
                  {todos.length ? (
                    todos.map((t) => (
                      <div key={t.id} style={{ marginBottom: 4, fontSize: 13 }}>
                        <Tag style={{ marginRight: 6 }}>{statusCn(t.status)}</Tag>
                        {t.title}
                      </div>
                    ))
                  ) : (
                    <Typography.Text type="secondary" style={{ fontSize: 13 }}>
                      无开放待办
                    </Typography.Text>
                  )}
                </div>
              </section>

              <section>
                <Typography.Text strong>开放 Issues（{issues.length}）</Typography.Text>
                <div style={{ marginTop: 6 }}>
                  {issues.length ? (
                    issues.map((i) => (
                      <div key={i.id} style={{ marginBottom: 4, fontSize: 13 }}>
                        <Tag color="purple" style={{ marginRight: 6 }}>
                          {i.identifier}
                        </Tag>
                        {i.title}{' '}
                        <Typography.Text type="secondary">（{statusCn(i.status)}）</Typography.Text>
                      </div>
                    ))
                  ) : (
                    <Typography.Text type="secondary" style={{ fontSize: 13 }}>
                      无指派开放 Issue
                    </Typography.Text>
                  )}
                </div>
              </section>

              {handoff?.generated_at ? (
                <Typography.Text type="secondary" style={{ fontSize: 12 }}>
                  摘要生成于 {formatWikiTime(handoff.generated_at)}
                </Typography.Text>
              ) : null}
            </>
          )}

          <Collapse
            size="small"
            items={[
              {
                key: 'migrate',
                label: '换机怎么续',
                children: (
                  <ol style={{ margin: 0, paddingLeft: 18, fontSize: 13 }}>
                    <li>新机打开同一工作区</li>
                    <li>绑定到同一席位（贴 Key / 本机席位进入）</li>
                    <li>拉取交接摘要</li>
                  </ol>
                ),
              },
            ]}
          />

          <section>
            <Typography.Text strong>交接卡预览</Typography.Text>
            <Typography.Paragraph type="secondary" style={{ margin: '4px 0 8px', fontSize: 12 }}>
              可传给同事或另一 Bot；不含完整 Key
            </Typography.Paragraph>
            <pre
              style={{
                margin: 0,
                padding: 12,
                fontSize: 12,
                lineHeight: 1.55,
                background: 'var(--ant-color-fill-quaternary, #f5f5f5)',
                borderRadius: 6,
                whiteSpace: 'pre-wrap',
                wordBreak: 'break-word',
                maxHeight: 220,
                overflow: 'auto',
              }}
            >
              {cardText}
            </pre>
          </section>
        </Space>
      )}
    </Drawer>
  )
}
