import { useMemo, useState } from 'react'
import {
  Alert,
  Button,
  Drawer,
  Empty,
  Input,
  Space,
  Spin,
  Steps,
  Tag,
  Typography,
  message,
} from 'antd'
import { ApiError } from '../../../api/client'
import {
  bindSeat,
  ensureRoster,
  getRoster,
  listTeams,
  isInviteServiceMissing,
  isSeatConflict,
  maskKeyTail,
  redeemTeamInvite,
  seatDisplayName,
  seatOccupancy,
  type RedeemInviteResult,
  type TeamSeat,
} from '../../../api/teams'
import SeatConflictPanel, { type SeatConflictInfo } from './SeatConflictPanel'

type StepKey = 'keys' | 'roster' | 'conflict'

type Props = {
  open: boolean
  onClose: () => void
  wsId: string
  /** 进席成功后回调：打开交接摘要 */
  onEntered: (opts: { memberId: string; seatName: string; roleId?: string | null }) => void
}

const YMK_RE = /^ymk_[A-Za-z0-9_-]+$/i
const YMT_RE = /^ymt_[A-Za-z0-9_-]+$/i

export default function JoinTeamWizard({ open, onClose, wsId, onEntered }: Props) {
  const [step, setStep] = useState<StepKey>('keys')
  const [ymk, setYmk] = useState('')
  const [ymt, setYmt] = useState('')
  const [busy, setBusy] = useState(false)
  const [redeem, setRedeem] = useState<RedeemInviteResult | null>(null)
  const [roster, setRoster] = useState<TeamSeat[]>([])
  const [allowNew, setAllowNew] = useState(false)
  const [conflict, setConflict] = useState<SeatConflictInfo | null>(null)
  const [serviceHint, setServiceHint] = useState<string | null>(null)

  const reset = () => {
    setStep('keys')
    setYmk('')
    setYmt('')
    setBusy(false)
    setRedeem(null)
    setRoster([])
    setAllowNew(false)
    setConflict(null)
    setServiceHint(null)
  }

  const handleClose = () => {
    reset()
    onClose()
  }

  const ymkMasked = useMemo(() => {
    const t = ymk.trim()
    if (!t) return ''
    if (t.length < 8) return maskKeyTail(t)
    return maskKeyTail(t)
  }, [ymk])

  const validateKeys = (): string | null => {
    const k = ymk.trim()
    const c = ymt.trim()
    if (!k) return '请粘贴身份钥匙（ymk_…）'
    if (!YMK_RE.test(k)) return '身份钥匙须以 ymk_ 开头，勿贴裸 UUID'
    if (!c) return '请粘贴团队入场码（ymt_…）'
    if (!YMT_RE.test(c)) return '团队入场码须以 ymt_ 开头，勿贴裸 UUID'
    return null
  }

  const onRedeem = async () => {
    const err = validateKeys()
    if (err) {
      message.warning(err)
      return
    }
    setBusy(true)
    setServiceHint(null)
    try {
      const res = await redeemTeamInvite(ymt.trim(), crypto.randomUUID())
      if (res.workspace_id && res.workspace_id !== wsId) {
        message.error('入场码属于其他工作区，请切换工作区后再试')
        return
      }
      setRedeem(res)
      setAllowNew(Boolean(res.allow_new_seat))
      let seats = res.roster || []
      if (!seats.length && res.team_id) {
        try {
          const r = await getRoster(wsId, res.team_id)
          seats = r.roster || []
        } catch {
          /* roster 可后拉 */
        }
        try {
          const ensured = await ensureRoster(wsId, res.team_id, undefined, crypto.randomUUID())
          seats = ensured.roster || seats
        } catch {
          /* ensure 可选 */
        }
      }
      setRoster(seats)
      setStep('roster')
    } catch (e: unknown) {
      if (isInviteServiceMissing(e)) {
        setServiceHint('入场码服务尚未开通')
        message.warning('入场码服务尚未开通')
        // 演示：拉本工作区已有班底，便于选席 / T3 专面（非真兑换）
        try {
          const { teams } = await listTeams(wsId)
          const tid = teams?.[0]?.id
          if (tid) {
            let seats: TeamSeat[] = []
            try {
              const ensured = await ensureRoster(wsId, tid, undefined, crypto.randomUUID())
              seats = ensured.roster || []
            } catch {
              const r = await getRoster(wsId, tid)
              seats = r.roster || []
            }
            setRedeem({ workspace_id: wsId, team_id: tid, roster: seats, allow_new_seat: false })
            setAllowNew(false)
            setRoster(seats)
            setStep('roster')
            message.info('已打开本区班底供演示（入场码服务未开通）')
          }
        } catch {
          /* 保持在钥匙步 */
        }
        return
      }
      const msg = e instanceof ApiError ? e.message : '兑换失败'
      message.error(msg)
    } finally {
      setBusy(false)
    }
  }

  const onPickSeat = async (seat: TeamSeat) => {
    const occ = seatOccupancy(seat)
    if (occ === 'active') {
      setConflict({
        seatName: seatDisplayName(seat),
        occupantName: seat.bound_display_name || null,
        boundAt: seat.bound_at,
        claimedAt: seat.claim?.claimed_at,
        message: '不能顶替活跃席',
      })
      setStep('conflict')
      return
    }
    // left/stopped/empty → 正常进席（不进 T3）
    if (!redeem?.team_id) {
      message.error('缺少团队上下文，请重新兑换入场码')
      return
    }
    setBusy(true)
    try {
      // 身份钥匙仅本机持有；看板侧 bind 走当前登录成员（Admin 代绑需选 Bot）
      // 产品：Bot 持 ymk_ 调用；Web 向导用当前会话 + 校验钥匙格式后进席
      const { seat: bound } = await bindSeat(
        wsId,
        redeem.team_id,
        seat.id,
        { force: false },
        crypto.randomUUID(),
      )
      message.success(`已进入席位「${seatDisplayName(bound)}」`)
      const memberId = bound.bound_member_id || bound.member_id || bound.claim?.member_id
      if (memberId) {
        onEntered({
          memberId,
          seatName: seatDisplayName(bound),
          roleId: bound.bot_role_id,
        })
      }
      handleClose()
    } catch (e: unknown) {
      if (isSeatConflict(e)) {
        setConflict({
          seatName: seatDisplayName(seat),
          occupantName: seat.bound_display_name || null,
          boundAt: seat.bound_at,
          claimedAt: seat.claim?.claimed_at,
          message: e instanceof ApiError ? e.message : '该席位占用中，禁止顶替',
        })
        setStep('conflict')
        return
      }
      message.error(e instanceof ApiError ? e.message : '进席失败')
    } finally {
      setBusy(false)
    }
  }

  const onRequestNewSeat = async () => {
    if (!redeem?.team_id) return
    if (!allowNew) {
      message.warning('此入场码未允许申请新席，请联系总指挥扩编或换空席')
      return
    }
    setBusy(true)
    try {
      const slug = `seat-${Date.now().toString(36).slice(-4)}`
      const res = await ensureRoster(
        wsId,
        redeem.team_id,
        { seats: [{ slug, name: '新席位' }] },
        crypto.randomUUID(),
      )
      setRoster(res.roster || [])
      setStep('roster')
      message.success('已申请新席，请选择空席进入')
    } catch (e: unknown) {
      if (isInviteServiceMissing(e) || (e instanceof ApiError && e.status === 403)) {
        message.warning('入场码未允许申请新席，或服务尚未开通')
      } else {
        message.error(e instanceof ApiError ? e.message : '申请新席失败')
      }
    } finally {
      setBusy(false)
    }
  }

  const emptySeats = roster.filter((s) => seatOccupancy(s) !== 'active')
  const stepIndex = step === 'keys' ? 0 : step === 'roster' ? 1 : 2

  return (
    <Drawer
      title="加入班底"
      width={440}
      open={open}
      onClose={handleClose}
      destroyOnClose
      className="acb-bots-join-drawer"
    >
      <Steps
        size="small"
        current={stepIndex}
        style={{ marginBottom: 16 }}
        items={[{ title: '钥匙' }, { title: '选席' }, { title: '结果' }]}
      />

      {step === 'keys' ? (
        <Space direction="vertical" size="middle" style={{ width: '100%' }}>
          <div>
            <Typography.Text strong>身份钥匙</Typography.Text>
            <Typography.Paragraph type="secondary" style={{ margin: '2px 0 6px', fontSize: 12 }}>
              本机 Bot Key（ymk_…）。界面只显末四，勿贴裸 UUID。
            </Typography.Paragraph>
            <Input.Password
              value={ymk}
              onChange={(e) => setYmk(e.target.value)}
              placeholder="ymk_…"
              visibilityToggle
              autoComplete="off"
            />
            {ymk.trim() ? (
              <Typography.Text type="secondary" style={{ fontSize: 12, display: 'block', marginTop: 4 }}>
                掩码 · {ymkMasked}
              </Typography.Text>
            ) : null}
          </div>
          <div>
            <Typography.Text strong>团队入场码</Typography.Text>
            <Typography.Paragraph type="secondary" style={{ margin: '2px 0 6px', fontSize: 12 }}>
              总指挥/Admin 签发的 ymt_…，定位工作区与班底。
            </Typography.Paragraph>
            <Input
              value={ymt}
              onChange={(e) => setYmt(e.target.value)}
              placeholder="ymt_…"
              autoComplete="off"
            />
          </div>
          {serviceHint ? <Alert type="info" showIcon message={serviceHint} /> : null}
          <Button type="primary" block loading={busy} onClick={() => void onRedeem()}>
            兑换并查看班底
          </Button>
        </Space>
      ) : null}

      {step === 'roster' ? (
        <div>
          <Typography.Paragraph type="secondary" style={{ marginBottom: 10, fontSize: 13 }}>
            空席可选；占用中灰显不可点。无空席时申请新席（码未允许则失败，无顶替）。
          </Typography.Paragraph>
          {busy && !roster.length ? (
            <div style={{ textAlign: 'center', padding: 24 }}>
              <Spin />
            </div>
          ) : roster.length ? (
            <div className="acb-bots-seat-grid" style={{ display: 'flex', flexDirection: 'column', gap: 6 }}>
              {roster.map((seat) => {
                const occ = seatOccupancy(seat)
                const active = occ === 'active'
                const left = occ === 'left'
                return (
                  <button
                    key={seat.id}
                    type="button"
                    className="acb-bots-seat-card"
                    aria-disabled={active || busy}
                    onClick={() => {
                      if (busy) return
                      void onPickSeat(seat)
                    }}
                    style={{
                      display: 'flex',
                      alignItems: 'center',
                      gap: 8,
                      minHeight: 36,
                      padding: '6px 10px',
                      textAlign: 'left',
                      borderRadius: 6,
                      border: '1px solid var(--acb-border, #e5e7eb)',
                      background: active
                        ? 'var(--acb-bg-layout, #f5f5f5)'
                        : 'var(--acb-bg-container, #fff)',
                      opacity: active ? 0.65 : 1,
                      cursor: active ? 'not-allowed' : 'pointer',
                    }}
                    title={seat.role_slug ? `CLI · ${seat.role_slug}` : undefined}
                  >
                    <Typography.Text strong className="acb-bots-lead" style={{ flex: 1, minWidth: 0 }}>
                      {seatDisplayName(seat)}
                    </Typography.Text>
                    {active ? (
                      <Tag style={{ margin: 0 }}>占用中</Tag>
                    ) : left ? (
                      <Tag color="orange" style={{ margin: 0 }}>
                        可补位
                      </Tag>
                    ) : (
                      <Tag color="green" style={{ margin: 0 }}>
                        空席
                      </Tag>
                    )}
                    {active && seat.bound_display_name ? (
                      <Typography.Text type="secondary" style={{ fontSize: 12, maxWidth: 88 }} ellipsis>
                        {seat.bound_display_name}
                      </Typography.Text>
                    ) : null}
                  </button>
                )
              })}
            </div>
          ) : (
            <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="班底暂无席位" />
          )}
          <Space style={{ marginTop: 14 }} wrap>
            {!emptySeats.length || allowNew ? (
              <Button loading={busy} onClick={() => void onRequestNewSeat()}>
                申请新席
              </Button>
            ) : null}
            <Button
              onClick={() => {
                setStep('keys')
                setConflict(null)
              }}
            >
              返回
            </Button>
          </Space>
        </div>
      ) : null}

      {step === 'conflict' && conflict ? (
        <SeatConflictPanel
          info={conflict}
          showRequestNew
          onPickOther={() => {
            setConflict(null)
            setStep('roster')
          }}
          onRequestNew={() => void onRequestNewSeat()}
          onBack={() => {
            setConflict(null)
            setStep('roster')
          }}
        />
      ) : null}
    </Drawer>
  )
}
