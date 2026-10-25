import { useMemo, useState, type ReactNode } from 'react'
import { Button, Collapse, Empty, Popconfirm, Space, Tag, Typography } from 'antd'
import {
  inviteStatusCn,
  inviteTail,
  type TeamInvite,
  type TeamSeat,
} from '../../../api/teams'
import {
  workspaceInviteStatusCn,
  workspaceInviteTail,
  type WorkspaceInvite,
} from '../../../api/workspaceInvites'

type Props = {
  teamInvites: TeamInvite[]
  wsInvites: WorkspaceInvite[]
  teamSeats: TeamSeat[]
  revokingId: string | null
  onRevokeTeam: (inv: TeamInvite) => void
  onRevokeWs?: (inv: WorkspaceInvite) => void
  /** I-R3 空态主 CTA：按席发一把 */
  onIssuePrimary?: () => void
  /** I-R3 空态次要：全班包（二级确认） */
  onOpenCrewPack?: () => void
  hasEmptySeats?: boolean
}

const PREVIEW = 2

function isRevokedTeam(inv: TeamInvite): boolean {
  const st = (inv.status || (inv.revoked_at ? 'revoked' : 'active')).toLowerCase()
  return st === 'revoked'
}

function isRevokedWs(inv: WorkspaceInvite): boolean {
  const st = (inv.status || (inv.revoked_at ? 'revoked' : 'active')).toLowerCase()
  return st === 'revoked'
}

function seatGroupKey(inv: TeamInvite): string {
  return (inv.role_hint || inv.seat_slug || '').trim() || '__none__'
}

/** INVITE_ON_DEMAND I-R3：脏串 / 对不上席表 →「未命名席」；永不渲染 team:uuid:slug */
function isDirtySeatKey(key: string): boolean {
  if (!key || key === '__none__') return false
  if (/^team:/i.test(key)) return true
  if (/[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}/i.test(key)) {
    return true
  }
  return false
}

function seatGroupLabel(key: string, seats: TeamSeat[]): string {
  if (key === '__none__') return '未指定席'
  if (isDirtySeatKey(key)) return '未命名席'
  const hit = seats.find((s) => s.role_slug === key)
  if (!hit) return '未命名席'
  return hit.display_name || hit.role_slug || '未命名席'
}

function InviteRow(props: {
  tail: string
  statusLabel: string
  revoked: boolean
  busy?: boolean
  onRevoke?: () => void
}) {
  return (
    <div className="acb-bots-invite-row">
      <Typography.Text className="acb-bots-invite-tail">末四 ·{props.tail}</Typography.Text>
      <Tag
        color={props.revoked ? 'default' : 'green'}
        style={{ margin: 0, flexShrink: 0 }}
      >
        {props.statusLabel}
      </Tag>
      {!props.revoked && props.onRevoke ? (
        <Popconfirm
          title="吊销此入场码？"
          description="吊销后无法再用该码加入；已进席不受影响。"
          okText="吊销"
          cancelText="取消"
          okButtonProps={{ danger: true }}
          onConfirm={() => props.onRevoke?.()}
        >
          <Button
            size="small"
            type="link"
            danger
            className="acb-bots-invite-revoke"
            style={{ padding: 0 }}
            loading={props.busy}
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
}

function TruncatedList<T>({
  items,
  render,
}: {
  items: T[]
  render: (item: T, i: number) => ReactNode
}) {
  const [open, setOpen] = useState(false)
  const shown = open ? items : items.slice(0, PREVIEW)
  const rest = items.length - PREVIEW
  return (
    <div className="acb-bots-invite-group-list">
      {shown.map((item, i) => render(item, i))}
      {!open && rest > 0 ? (
        <Button
          type="link"
          size="small"
          className="acb-bots-invite-more"
          style={{ padding: 0, height: 'auto' }}
          onClick={() => setOpen(true)}
        >
          还有 {rest} 条
        </Button>
      ) : null}
      {open && items.length > PREVIEW ? (
        <Button
          type="link"
          size="small"
          style={{ padding: 0, height: 'auto' }}
          onClick={() => setOpen(false)}
        >
          收起
        </Button>
      ) : null}
    </div>
  )
}

export default function IssuedInvitesPanel({
  teamInvites,
  wsInvites,
  teamSeats,
  revokingId,
  onRevokeTeam,
  onRevokeWs,
  onIssuePrimary,
  onOpenCrewPack,
  hasEmptySeats,
}: Props) {
  const teamActive = useMemo(() => teamInvites.filter((i) => !isRevokedTeam(i)), [teamInvites])
  const teamRevoked = useMemo(() => teamInvites.filter((i) => isRevokedTeam(i)), [teamInvites])
  const wsActive = useMemo(() => wsInvites.filter((i) => !isRevokedWs(i)), [wsInvites])
  const wsRevoked = useMemo(() => wsInvites.filter((i) => isRevokedWs(i)), [wsInvites])

  const ymtGroups = useMemo(() => {
    const map = new Map<string, { key: string; label: string; active: TeamInvite[]; revoked: TeamInvite[] }>()
    for (const inv of teamInvites) {
      const key = seatGroupKey(inv)
      let g = map.get(key)
      if (!g) {
        g = { key, label: seatGroupLabel(key, teamSeats), active: [], revoked: [] }
        map.set(key, g)
      }
      if (isRevokedTeam(inv)) g.revoked.push(inv)
      else g.active.push(inv)
    }
    return [...map.values()].sort((a, b) => {
      const aScore = a.active.length ? 0 : 1
      const bScore = b.active.length ? 0 : 1
      if (aScore !== bScore) return aScore - bScore
      return a.label.localeCompare(b.label, 'zh')
    })
  }, [teamInvites, teamSeats])

  const total = teamInvites.length + wsInvites.length

  // I-R3：空态勿 null；主 CTA=按席发一把；全班包次要
  if (total === 0) {
    return (
      <div className="acb-bots-invite-empty">
        <Empty
          image={Empty.PRESENTED_IMAGE_SIMPLE}
          description={
            <span>
              尚无入场码 · 入场码是钥匙不是海报
              <br />
              <Typography.Text type="secondary" style={{ fontSize: 12 }}>
                要一把才开一把
              </Typography.Text>
            </span>
          }
        >
          <Space wrap>
            <Button type="primary" size="small" onClick={() => onIssuePrimary?.()}>
              按席发一把
            </Button>
            {hasEmptySeats && onOpenCrewPack ? (
              <Button size="small" onClick={() => onOpenCrewPack()}>
                全班入场包（需确认）
              </Button>
            ) : null}
          </Space>
        </Empty>
      </div>
    )
  }

  const activeCount = teamActive.length + wsActive.length
  const summary =
    activeCount > 0
      ? `已签发 ${total} · 有效 ${activeCount}`
      : `已签发 ${total} · 暂无有效`

  const ymtRevokedAll = teamRevoked
  const ymwRevokedAll = wsRevoked

  return (
    <Collapse
      size="small"
      className="acb-bots-invite-list"
      items={[
        {
          key: 'issued',
          label: `已签发入场码 · ${summary}`,
          children: (
            <div className="acb-bots-invite-body">
              {teamInvites.length ? (
                <div className="acb-bots-invite-kind">
                  <Typography.Text strong className="acb-bots-invite-kind-title">
                    队席码 ymt_
                  </Typography.Text>
                  {ymtGroups.map((g) =>
                    g.active.length ? (
                      <div key={g.key} className="acb-bots-invite-group">
                        <Typography.Text type="secondary" className="acb-bots-invite-group-label">
                          {g.label}
                        </Typography.Text>
                        <TruncatedList
                          items={g.active}
                          render={(inv) => (
                            <InviteRow
                              key={inv.id}
                              tail={inviteTail(inv)}
                              statusLabel={inviteStatusCn(
                                (inv.status || 'active').toLowerCase(),
                              )}
                              revoked={false}
                              busy={revokingId === inv.id}
                              onRevoke={() => onRevokeTeam(inv)}
                            />
                          )}
                        />
                      </div>
                    ) : null,
                  )}
                  {ymtRevokedAll.length ? (
                    <Collapse
                      size="small"
                      className="acb-bots-invite-revoked"
                      items={[
                        {
                          key: 'ymt-revoked',
                          label: `已吊销 ${ymtRevokedAll.length}`,
                          children: (
                            <TruncatedList
                              items={ymtRevokedAll}
                              render={(inv) => (
                                <InviteRow
                                  key={inv.id}
                                  tail={inviteTail(inv)}
                                  statusLabel={inviteStatusCn('revoked')}
                                  revoked
                                />
                              )}
                            />
                          ),
                        },
                      ]}
                    />
                  ) : null}
                </div>
              ) : null}

              {wsInvites.length ? (
                <div className="acb-bots-invite-kind">
                  <Typography.Text strong className="acb-bots-invite-kind-title">
                    工作区码 ymw_
                  </Typography.Text>
                  {wsActive.length ? (
                    <div className="acb-bots-invite-group">
                      <TruncatedList
                        items={wsActive}
                        render={(inv) => (
                          <InviteRow
                            key={inv.id}
                            tail={workspaceInviteTail(inv)}
                            statusLabel={workspaceInviteStatusCn(
                              (inv.status || 'active').toLowerCase(),
                            )}
                            revoked={false}
                            busy={revokingId === inv.id}
                            onRevoke={onRevokeWs ? () => onRevokeWs(inv) : undefined}
                          />
                        )}
                      />
                    </div>
                  ) : null}
                  {ymwRevokedAll.length ? (
                    <Collapse
                      size="small"
                      className="acb-bots-invite-revoked"
                      items={[
                        {
                          key: 'ymw-revoked',
                          label: `已吊销 ${ymwRevokedAll.length}`,
                          children: (
                            <TruncatedList
                              items={ymwRevokedAll}
                              render={(inv) => (
                                <InviteRow
                                  key={inv.id}
                                  tail={workspaceInviteTail(inv)}
                                  statusLabel={workspaceInviteStatusCn('revoked')}
                                  revoked
                                />
                              )}
                            />
                          ),
                        },
                      ]}
                    />
                  ) : null}
                </div>
              ) : null}
            </div>
          ),
        },
      ]}
    />
  )
}
