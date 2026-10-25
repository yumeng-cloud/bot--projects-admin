import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
import { Alert, Button, Checkbox, Modal, Space, Typography, message } from 'antd'
import { CopyOutlined } from '@ant-design/icons'
import { seatDisplayName, type TeamSeat } from '../../api/teams'
import {
  emptySeatsForPack,
  issueCrewPack,
  type CrewPackResult,
  type CrewPackSeatFail,
} from '../../utils/autoCrewPack'

export type CrewPackRevealModalProps = {
  open: boolean
  onClose: () => void
  workspaceId: string
  teamId: string
  teamName: string
  /** 当前队 roster；勾选空席确认后再签 */
  seats: TeamSeat[]
  /** 签发成功后回写已签发列表（末四） */
  onIssued?: (result: CrewPackResult) => void
  /** 建队成功自动弹出时标题区分（现已不再自动弹；保留兼容） */
  afterCreate?: boolean
}

/**
 * INVITE_ON_DEMAND I-R4：全班入场包二级挡层。
 * 禁止 useEffect 一 open 就 issueCrewPack；勾选空席 → 用户确认后再签。
 * 禁嵌 ymk_。
 */
export default function CrewPackRevealModal({
  open,
  onClose,
  workspaceId,
  teamId,
  teamName,
  seats,
  onIssued,
  afterCreate,
}: CrewPackRevealModalProps) {
  const [busy, setBusy] = useState(false)
  const [result, setResult] = useState<CrewPackResult | null>(null)
  const [copied, setCopied] = useState(false)
  const [selected, setSelected] = useState<string[]>([])
  const runIdRef = useRef(0)

  const emptySeats = useMemo(() => emptySeatsForPack(seats), [seats])

  const reset = useCallback(() => {
    setBusy(false)
    setResult(null)
    setCopied(false)
    setSelected([])
  }, [])

  // 打开时重置；默认不勾选（要一把才开一把）；不自动签发
  useEffect(() => {
    if (!open) {
      reset()
      runIdRef.current += 1
      return
    }
    setResult(null)
    setCopied(false)
    setBusy(false)
    setSelected([])
  }, [open, reset])

  const runIssue = useCallback(async () => {
    if (!workspaceId || !teamId) return
    if (selected.length === 0) {
      message.warning('请先勾选要签发的空席')
      return
    }
    const myRun = ++runIdRef.current
    setBusy(true)
    setResult(null)
    setCopied(false)
    try {
      const res = await issueCrewPack({
        workspaceId,
        teamId,
        teamName: teamName || '未命名团队',
        seats,
        selectedSlugs: selected,
      })
      if (myRun !== runIdRef.current) return
      setResult(res)
      onIssued?.(res)
      if (res.ok.length === 0 && res.failed.length > 0) {
        message.error(
          `全班入场包签发失败：${res.failed.map((f) => f.displayName).join('、')}`,
        )
      } else if (res.failed.length > 0) {
        message.warning(
          `部分席签发失败（${res.failed.map((f) => f.displayName).join('、')}）；其余已可复制`,
        )
      }
    } catch (e: unknown) {
      if (myRun !== runIdRef.current) return
      message.error(e instanceof Error ? e.message : '签发全班入场包失败')
      setResult({ teamName, markdown: '', ok: [], failed: [] })
    } finally {
      if (myRun === runIdRef.current) setBusy(false)
    }
  }, [workspaceId, teamId, teamName, seats, selected, onIssued])

  const failSummary = (fails: CrewPackSeatFail[]) =>
    fails.map((f) => `「${f.displayName}」${f.reason ? `：${f.reason}` : ''}`).join('；')

  const onCopy = async () => {
    if (!result?.markdown) {
      message.warning('暂无可复制内容')
      return
    }
    try {
      await navigator.clipboard.writeText(result.markdown)
      setCopied(true)
      const n = result.ok.length
      const failHint =
        result.failed.length > 0
          ? `；失败席：${result.failed.map((f) => f.displayName).join('、')}`
          : ''
      message.success(`已复制全班入场包（${n} 席）。勿发公开群${failHint}`)
    } catch {
      message.warning('复制失败，请手动选中下方文本复制')
    }
  }

  const title = afterCreate
    ? `班底已就绪 · 勾选空席后签发入场包`
    : `全班入场包（二级）`

  const showPicker = !result && !busy
  const allSlugs = emptySeats.map((s) => (s.role_slug || '').trim()).filter(Boolean)
  const allChecked = allSlugs.length > 0 && allSlugs.every((s) => selected.includes(s))

  return (
    <Modal
      title={title}
      open={open}
      onCancel={onClose}
      footer={null}
      destroyOnHidden
      centered
      width={720}
      maskClosable={!busy}
    >
      <Space direction="vertical" size="middle" style={{ width: '100%' }}>
        <Alert
          type="info"
          showIcon
          message={
            afterCreate
              ? `团队「${teamName}」已创建；按需勾选空席再签发（勿默认全开）`
              : `为当前队「${teamName}」勾选空席后确认签发（仅 ymt_ + 席位）`
          }
          description="入场码是钥匙不是海报——要一把才开一把。勿发公开群。主路径请用席卡「入场链接 / 入场提示词」。"
        />

        {showPicker ? (
          <div>
            <div style={{ display: 'flex', alignItems: 'center', gap: 12, marginBottom: 8 }}>
              <Typography.Text strong style={{ fontSize: 13 }}>
                勾选要签发的空席
              </Typography.Text>
              {allSlugs.length > 0 ? (
                <Button
                  type="link"
                  size="small"
                  style={{ padding: 0 }}
                  onClick={() => setSelected(allChecked ? [] : [...allSlugs])}
                >
                  {allChecked ? '清空' : '全选空席'}
                </Button>
              ) : null}
            </div>
            {emptySeats.length === 0 ? (
              <Alert type="warning" showIcon message="当前没有空席可签" />
            ) : (
              <Checkbox.Group
                style={{ width: '100%' }}
                value={selected}
                onChange={(vals) => setSelected(vals as string[])}
              >
                <Space direction="vertical" size={4} style={{ width: '100%' }}>
                  {emptySeats.map((seat) => {
                    const slug = (seat.role_slug || '').trim()
                    if (!slug) return null
                    return (
                      <Checkbox key={slug} value={slug}>
                        {seatDisplayName(seat)}
                        <Typography.Text type="secondary" style={{ marginLeft: 6, fontSize: 12 }}>
                          （空席）
                        </Typography.Text>
                      </Checkbox>
                    )
                  })}
                </Space>
              </Checkbox.Group>
            )}
          </div>
        ) : null}

        {busy ? (
          <Typography.Text type="secondary">正在为勾选席签发入场码…</Typography.Text>
        ) : null}

        {result && result.failed.length > 0 ? (
          <Alert
            type="warning"
            showIcon
            message="部分席签发失败"
            description={failSummary(result.failed)}
          />
        ) : null}

        {result && result.ok.length === 0 && !busy ? (
          <Alert
            type="error"
            showIcon
            message="未能签发任何入场码"
            description={
              result.failed.length
                ? failSummary(result.failed)
                : '未勾选空席，或签发接口未返回明文'
            }
          />
        ) : null}

        {result?.markdown ? (
          <div>
            <Typography.Text type="secondary" style={{ fontSize: 12, display: 'block', marginBottom: 6 }}>
              全班入场包（{result.ok.length} 席 · Markdown）
            </Typography.Text>
            <Typography.Paragraph
              style={{
                marginBottom: 0,
                maxHeight: 360,
                overflow: 'auto',
                whiteSpace: 'pre-wrap',
                wordBreak: 'break-word',
                fontSize: 12,
                fontFamily: 'var(--acb-font-mono, ui-monospace, monospace)',
                padding: '10px 12px',
                background: 'var(--ant-color-fill-quaternary, #f5f5f5)',
                borderRadius: 6,
              }}
            >
              {result.markdown}
            </Typography.Paragraph>
          </div>
        ) : null}

        <Space wrap>
          {showPicker ? (
            <Button
              type="primary"
              loading={busy}
              disabled={selected.length === 0 || emptySeats.length === 0}
              onClick={() => void runIssue()}
            >
              确认签发选中席（{selected.length}）
            </Button>
          ) : null}
          {result?.markdown ? (
            <Button
              type="primary"
              icon={<CopyOutlined />}
              loading={busy}
              onClick={() => void onCopy()}
            >
              复制全班入场包
            </Button>
          ) : null}
          {result && !busy ? (
            <Button
              onClick={() => {
                setResult(null)
                setCopied(false)
              }}
            >
              重新勾选
            </Button>
          ) : null}
          <Button onClick={onClose} disabled={busy}>
            {copied ? '完成' : '关闭'}
          </Button>
        </Space>
      </Space>
    </Modal>
  )
}
