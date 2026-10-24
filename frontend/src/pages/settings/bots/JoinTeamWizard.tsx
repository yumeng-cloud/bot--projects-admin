import { useEffect, useMemo, useState } from 'react'
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
  isRefillNeedsExplicitBind,
  humanizeRefillBindError,
  rosterHasRefillableSeat,
  SEAT_REFILL_GUIDE_TITLE,
  SEAT_REFILL_GUIDE_DESC,
  maskKeyTail,
  redeemTeamInvite,
  seatDisplayName,
  seatOccupancy,
  type RedeemInviteResult,
  type TeamSeat,
} from '../../../api/teams'
import IdentityKeyRevealModal from '../../../components/common/IdentityKeyRevealModal'
import SeatConflictPanel, { type SeatConflictInfo } from './SeatConflictPanel'

type StepKey = 'code' | 'roster' | 'conflict'

type Props = {
  open: boolean
  onClose: () => void
  wsId: string
  /** 进席成功后回调：打开交接摘要 */
  onEntered: (opts: { memberId: string; seatName: string; roleId?: string | null }) => void
  /** C1：深链 /join?code= 预填入场码 */
  initialCode?: string | null
  /** C1：深链 role= 席位 slug，redeem 时带上 */
  initialRole?: string | null
  /** C1：深链 seat= 席位 id，优先于 role */
  initialSeat?: string | null
}

const YMK_RE = /^ymk_[A-Za-z0-9_-]+$/i
const YMT_RE = /^ymt_[A-Za-z0-9_-]+$/i

export default function JoinTeamWizard({
  open,
  onClose,
  wsId,
  onEntered,
  initialCode,
  initialRole,
  initialSeat,
}: Props) {
  const [step, setStep] = useState<StepKey>('code')
  const [ymk, setYmk] = useState('')
  const [ymt, setYmt] = useState('')
  const [busy, setBusy] = useState(false)
  const [redeem, setRedeem] = useState<RedeemInviteResult | null>(null)
  const [roster, setRoster] = useState<TeamSeat[]>([])
  const [allowNew, setAllowNew] = useState(false)
  const [conflict, setConflict] = useState<SeatConflictInfo | null>(null)
  const [serviceHint, setServiceHint] = useState<string | null>(null)
  const [revealToken, setRevealToken] = useState<string | null>(null)
  const [showOptionalKey, setShowOptionalKey] = useState(false)
  /** Soft† V2G：通码未绑 + 有停工席 → 人话引导（禁静默补） */
  const [refillGuide, setRefillGuide] = useState(false)
  /** 深链带入的 role/seat，仅用于 redeem body */
  const [prefRole, setPrefRole] = useState('')
  const [prefSeat, setPrefSeat] = useState('')


  const reset = () => {
    setStep('code')
    setYmk('')
    setYmt('')
    setBusy(false)
    setRedeem(null)
    setRoster([])
    setAllowNew(false)
    setConflict(null)
    setServiceHint(null)
    setRevealToken(null)
    setShowOptionalKey(false)
    setRefillGuide(false)
    setPrefRole('')
    setPrefSeat('')
  }

  const handleClose = () => {
    reset()
    onClose()
  }

  // C1：打开时预填深链 code/role/seat（主路径只贴 ymt_）
  useEffect(() => {
    if (!open) return
    const c = (initialCode || '').trim()
    if (c) setYmt(c)
    const r = (initialRole || '').trim()
    const s = (initialSeat || '').trim()
    setPrefRole(r)
    setPrefSeat(s)
  }, [open, initialCode, initialRole, initialSeat])


  const ymkMasked = useMemo(() => {
    const t = ymk.trim()
    if (!t) return ''
    return maskKeyTail(t)
  }, [ymk])

  const validateCode = (): string | null => {
    const c = ymt.trim()
    if (!c) return '请粘贴团队入场码（ymt_…）'
    if (!YMT_RE.test(c)) return '团队入场码须以 ymt_ 开头，勿贴裸 UUID'
    const k = ymk.trim()
    if (k && !YMK_RE.test(k)) return '身份钥匙须以 ymk_ 开头，勿贴裸 UUID'
    return null
  }

  const finishEntered = (res: RedeemInviteResult, seats: TeamSeat[]) => {
    const memberId = res.member_id
    if (!memberId) return
    let seatName = res.role_slug || res.seat_slug || res.role_hint || '席位'
    if (res.seat_id) {
      const hit = seats.find((s) => s.id === res.seat_id)
      if (hit) seatName = seatDisplayName(hit)
    }
    onEntered({
      memberId,
      seatName: typeof seatName === 'string' ? seatName : '席位',
      roleId: res.seat_id || null,
    })
  }

  const loadSeatsAfterRedeem = async (res: RedeemInviteResult): Promise<TeamSeat[]> => {
    let seats = res.roster || []
    if (!seats.length && res.team_id) {
      try {
        const r = await getRoster(wsId, res.team_id)
        seats = r.roster || []
      } catch {
        /* roster 可后拉 */
      }
    }
    return seats
  }

  /** V2G-1：通码成功未绑席 + 有停工/离职可补位席 → 人话引导，禁静默补 */
  const applyUnboundRefillGuide = (res: RedeemInviteResult, seats: TeamSeat[]): boolean => {
    if (res.seat_id) {
      setRefillGuide(false)
      return false
    }
    if (!rosterHasRefillableSeat(seats)) {
      setRefillGuide(false)
      return false
    }
    setRefillGuide(true)
    setRoster(seats)
    setStep('roster')
    return true
  }

  const onRedeem = async () => {
    const err = validateCode()
    if (err) {
      message.warning(err)
      return
    }
    setBusy(true)
    setServiceHint(null)
    setRefillGuide(false)
    try {
      const botKey = ymk.trim() || null
      const res = await redeemTeamInvite(ymt.trim(), {
        idempotencyKey: crypto.randomUUID(),
        botKey,
        role: prefRole || null,
        seat: prefSeat || null,
      })
      if (res.workspace_id && res.workspace_id !== wsId) {
        message.error('入场码属于其他工作区，请切换工作区后再试')
        return
      }
      setRedeem(res)
      setAllowNew(Boolean(res.allow_expand_seat))

      const issued = Boolean(res.token_issued && res.token && res.token.startsWith('ymk_'))
      if (issued && res.token) {
        setRevealToken(res.token)
        message.success('入场成功，已自动获得身份钥匙（仅此一次）')
        const seats = await loadSeatsAfterRedeem(res)
        setRoster(seats)
        if (applyUnboundRefillGuide(res, seats)) {
          finishEntered(res, seats)
          return
        }
        finishEntered(res, seats)
        return
      }

      // 续队（持钥）已定位/进席：无新 token
      if (botKey && res.member_id) {
        const seats = await loadSeatsAfterRedeem(res)
        setRoster(seats)
        if (applyUnboundRefillGuide(res, seats)) {
          message.success('续队成功（未重发身份钥匙）')
          finishEntered(res, seats)
          return
        }
        message.success('续队成功（未重发身份钥匙）')
        finishEntered(res, seats)
        handleClose()
        return
      }

      let seats = await loadSeatsAfterRedeem(res)
      if (!seats.length && res.team_id) {
        try {
          const ensured = await ensureRoster(wsId, res.team_id, undefined, crypto.randomUUID())
          seats = ensured.roster || seats
        } catch {
          /* ensure 可选 */
        }
      }
      setRoster(seats)
      if (applyUnboundRefillGuide(res, seats)) {
        return
      }
      setStep('roster')
    } catch (e: unknown) {
      if (isInviteServiceMissing(e)) {
        setServiceHint('入场码服务尚未开通')
        message.warning('入场码服务尚未开通')
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
            setRedeem({ workspace_id: wsId, team_id: tid, roster: seats, allow_expand_seat: false })
            setAllowNew(false)
            setRoster(seats)
            setStep('roster')
            message.info('已打开本区班底供演示（入场码服务未开通）')
          }
        } catch {
          /* 保持在入场码步 */
        }
        return
      }
      // V2G-2：409「补位须走显式绑席…」→ 人话，勿甩技术句
      const human = humanizeRefillBindError(e)
      if (human || isRefillNeedsExplicitBind(e)) {
        setRefillGuide(true)
        message.warning(human || `${SEAT_REFILL_GUIDE_TITLE}。${SEAT_REFILL_GUIDE_DESC}`)
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
    if (!redeem?.team_id) {
      message.error('缺少团队上下文，请重新兑换入场码')
      return
    }
    setBusy(true)
    try {
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
      const human = humanizeRefillBindError(e)
      if (human || isRefillNeedsExplicitBind(e)) {
        setRefillGuide(true)
        message.warning(human || `${SEAT_REFILL_GUIDE_TITLE}。${SEAT_REFILL_GUIDE_DESC}`)
        return
      }
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
  const stepIndex = step === 'code' ? 0 : step === 'roster' ? 1 : 2

  return (
    <>
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
          items={[{ title: '入场码' }, { title: '选席' }, { title: '结果' }]}
        />

        {step === 'code' ? (
          <Space direction="vertical" size="middle" style={{ width: '100%' }}>
            <div>
              <Typography.Text strong>团队入场码</Typography.Text>
              <Typography.Paragraph type="secondary" style={{ margin: '2px 0 6px', fontSize: 12 }}>
                主路径：只贴总指挥签发的 ymt_…。空手入场后自动获得身份钥匙（仅此一次）。
              </Typography.Paragraph>
              <Input
                value={ymt}
                onChange={(e) => setYmt(e.target.value)}
                placeholder="ymt_…"
                autoComplete="off"
              />
              {prefRole || prefSeat ? (
                <Typography.Paragraph type="secondary" style={{ margin: '6px 0 0', fontSize: 12 }}>
                  深链已带
                  {prefRole ? <> 角色 <Typography.Text code>{prefRole}</Typography.Text></> : null}
                  {prefRole && prefSeat ? ' ·' : null}
                  {prefSeat ? <> 席位 <Typography.Text code>{prefSeat}</Typography.Text></> : null}
                  ，兑换时自动带上。
                </Typography.Paragraph>
              ) : null}
            </div>
            <div>
              <Button
                type="link"
                size="small"
                style={{ padding: 0, height: 'auto' }}
                onClick={() => setShowOptionalKey((v) => !v)}
              >
                {showOptionalKey ? '收起：已有身份钥匙（续队）' : '可选：已有身份钥匙（续队）'}
              </Button>
              {showOptionalKey ? (
                <div style={{ marginTop: 8 }}>
                  <Typography.Paragraph type="secondary" style={{ margin: '0 0 6px', fontSize: 12 }}>
                    续队/换机持原 ymk_… 再兑同一入场码；不会重发钥匙。勿贴裸 UUID。
                  </Typography.Paragraph>
                  <Input.Password
                    value={ymk}
                    onChange={(e) => setYmk(e.target.value)}
                    placeholder="ymk_…（可选）"
                    visibilityToggle
                    autoComplete="off"
                  />
                  {ymk.trim() ? (
                    <Typography.Text type="secondary" style={{ fontSize: 12, display: 'block', marginTop: 4 }}>
                      掩码 · {ymkMasked}
                    </Typography.Text>
                  ) : null}
                </div>
              ) : null}
            </div>
            {serviceHint ? <Alert type="info" showIcon message={serviceHint} /> : null}
            {refillGuide ? (
              <Alert
                type="warning"
                showIcon
                message={SEAT_REFILL_GUIDE_TITLE}
                description={SEAT_REFILL_GUIDE_DESC}
              />
            ) : null}
            <Button type="primary" block loading={busy} onClick={() => void onRedeem()}>
              兑换入场
            </Button>
          </Space>
        ) : null}

        {step === 'roster' ? (
          <div>
            {refillGuide ? (
              <Alert
                type="warning"
                showIcon
                style={{ marginBottom: 12 }}
                message={SEAT_REFILL_GUIDE_TITLE}
                description={SEAT_REFILL_GUIDE_DESC}
              />
            ) : null}
            <Typography.Paragraph type="secondary" style={{ marginBottom: 10, fontSize: 13 }}>
              空席可选进入；占用中灰显不可选（不能顶替）。无空席时申请新席（码未允许则失败）。
              {refillGuide
                ? ' 停工/离职席请持管理员签发的补位入场码入场，普通通码不能自动顶。'
                : ''}
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
                      disabled={active || busy}
                      tabIndex={active ? -1 : undefined}
                      aria-disabled={active || busy}
                      aria-label={
                        active
                          ? `${seatDisplayName(seat)}，占用中，不可选`
                          : `${seatDisplayName(seat)}`
                      }
                      onClick={() => {
                        if (active || busy) return
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
                        opacity: active ? 0.55 : 1,
                        cursor: active || busy ? 'not-allowed' : 'pointer',
                        pointerEvents: active ? 'none' : undefined,
                      }}
                      title={seatDisplayName(seat)}
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
                  setStep('code')
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

      <IdentityKeyRevealModal
        open={Boolean(revealToken)}
        token={revealToken}
        inviteCode={ymt.trim() || null}
        onClose={() => {
          setRevealToken(null)
          // V2G：通码未绑且有停工席时留在选席步展示引导，勿关向导
          if (refillGuide) return
          handleClose()
        }}
      />
    </>
  )
}
