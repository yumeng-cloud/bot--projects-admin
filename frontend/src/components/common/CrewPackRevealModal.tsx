import { useCallback, useEffect, useRef, useState } from 'react'
import { Alert, Button, Modal, Space, Typography, message } from 'antd'
import { CopyOutlined } from '@ant-design/icons'
import type { TeamSeat } from '../../api/teams'
import {
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
  /** 当前队 roster；打开时对空席签发 */
  seats: TeamSeat[]
  /** 签发成功后回写已签发列表（末四） */
  onIssued?: (result: CrewPackResult) => void
  /** 建队成功自动弹出时标题区分 */
  afterCreate?: boolean
}

/**
 * AUTO_CREW C1/C3：全班入场包挡层——签发空席 ymt_ + 展示 Markdown + 一键复制。
 * 部分席失败标明席名，其余仍可复制。禁嵌 ymk_。
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
  const runIdRef = useRef(0)

  const reset = useCallback(() => {
    setBusy(false)
    setResult(null)
    setCopied(false)
  }, [])

  const runIssue = useCallback(async () => {
    if (!workspaceId || !teamId) return
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
  }, [workspaceId, teamId, teamName, seats, onIssued])

  useEffect(() => {
    if (!open) {
      reset()
      runIdRef.current += 1
      return
    }
    void runIssue()
  }, [open]) // eslint-disable-line react-hooks/exhaustive-deps -- 仅随 open 触发一次签发

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
    ? `班底已就绪 · 复制全班入场包`
    : `复制全班入场包`

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
              ? `团队「${teamName}」已创建；一次复制即可把各空席入场材料发给对应 Agent`
              : `为当前队「${teamName}」的每个空席签发带 role 的入场码（仅 ymt_，不含身份钥匙）`
          }
          description="勿发公开群。身份钥匙（ymk_）仍走设置·接入，勿与入场码混用。"
        />

        {busy ? (
          <Typography.Text type="secondary">正在为各空席签发入场码…</Typography.Text>
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
                : '当前没有空席，或签发接口未返回明文'
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
          <Button
            type="primary"
            icon={<CopyOutlined />}
            loading={busy}
            disabled={!result?.markdown}
            onClick={() => void onCopy()}
          >
            复制全班入场包
          </Button>
          {!busy && result && result.ok.length === 0 ? (
            <Button onClick={() => void runIssue()}>重试签发</Button>
          ) : null}
          <Button onClick={onClose} disabled={busy}>
            {copied ? '完成' : '关闭'}
          </Button>
        </Space>
      </Space>
    </Modal>
  )
}
