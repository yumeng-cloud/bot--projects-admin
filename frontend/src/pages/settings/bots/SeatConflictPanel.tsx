import { Alert, Button, Space, Typography } from 'antd'
import { formatWikiTime } from '../../../utils/time'

export type SeatConflictInfo = {
  seatName: string
  occupantName?: string | null
  boundAt?: string | null
  claimedAt?: string | null
  message?: string
}

type Props = {
  info: SeatConflictInfo
  onPickOther: () => void
  onRequestNew?: () => void
  onBack: () => void
  showRequestNew?: boolean
}

/** T3 单活冲突专面：非一行 message.error；无 force */
export default function SeatConflictPanel({
  info,
  onPickOther,
  onRequestNew,
  onBack,
  showRequestNew,
}: Props) {
  const lastAt = info.claimedAt || info.boundAt
  return (
    <div className="acb-bots-conflict" style={{ padding: '8px 0' }}>
      <Alert
        type="warning"
        showIcon
        message="席位占用中，不能顶替"
        description="活跃席禁止第二路绑定。请换空席、申请新席，或等对方离席后再进。"
        style={{ marginBottom: 16 }}
      />
      <div className="acb-bots-conflict-meta" style={{ marginBottom: 16, fontSize: 13, lineHeight: 1.6 }}>
        <div>
          <Typography.Text type="secondary">席位</Typography.Text>
          <div>
            <Typography.Text strong>{info.seatName}</Typography.Text>
          </div>
        </div>
        <div style={{ marginTop: 10 }}>
          <Typography.Text type="secondary">占用者</Typography.Text>
          <div>{info.occupantName || '未知演员'}</div>
        </div>
        {lastAt ? (
          <div style={{ marginTop: 10 }}>
            <Typography.Text type="secondary">末活 / 绑定</Typography.Text>
            <div>{formatWikiTime(lastAt)}</div>
          </div>
        ) : null}
        {info.message ? (
          <Typography.Paragraph type="secondary" style={{ margin: '10px 0 0', fontSize: 12 }}>
            {info.message}
          </Typography.Paragraph>
        ) : null}
      </div>
      <Space wrap>
        <Button type="primary" onClick={onPickOther}>
          换空席
        </Button>
        {showRequestNew && onRequestNew ? (
          <Button onClick={onRequestNew}>申请新席</Button>
        ) : null}
        <Button onClick={onBack}>返回</Button>
      </Space>
    </div>
  )
}
