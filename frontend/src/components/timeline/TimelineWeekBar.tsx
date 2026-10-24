import { Button, Space } from 'antd'
import { LeftOutlined, ReloadOutlined, RightOutlined } from '@ant-design/icons'
import './timeline.css'

export type TimelineWeekBarProps = {
  onPrevWeek?: () => void
  onThisWeek?: () => void
  onNextWeek?: () => void
  onRefresh?: () => void
  /** S8: disable refresh while a bar drag is in progress */
  refreshDisabled?: boolean
  /** Disable all nav */
  disabled?: boolean
  className?: string
}

/**
 * S8 presentational week controls — labels 「上周｜本周｜下周」+ 刷新.
 * Page owns range state; pass refreshDisabled while dragRef is active.
 */
export function TimelineWeekBar({
  onPrevWeek,
  onThisWeek,
  onNextWeek,
  onRefresh,
  refreshDisabled = false,
  disabled = false,
  className,
}: TimelineWeekBarProps) {
  return (
    <Space size={8} className={['acb-timeline-week-bar', className].filter(Boolean).join(' ')}>
      <Button
        type="text"
        size="small"
        icon={<LeftOutlined />}
        disabled={disabled}
        onClick={() => onPrevWeek?.()}
      >
        上周
      </Button>
      <Button type="text" size="small" disabled={disabled} onClick={() => onThisWeek?.()}>
        本周
      </Button>
      <Button
        type="text"
        size="small"
        icon={<RightOutlined />}
        disabled={disabled}
        onClick={() => onNextWeek?.()}
      >
        下周
      </Button>
      <Button
        type="text"
        size="small"
        icon={<ReloadOutlined />}
        disabled={disabled || refreshDisabled}
        onClick={() => onRefresh?.()}
        title={refreshDisabled ? '拖拽进行中，请先松手再刷新' : '刷新'}
      >
        刷新
      </Button>
    </Space>
  )
}
