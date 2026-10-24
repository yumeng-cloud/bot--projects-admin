import type { ReactNode } from 'react'
import { Dropdown, type MenuProps } from 'antd'
import { MoreOutlined } from '@ant-design/icons'
import './timeline.css'

export type TimelineBarMenuProps = {
  /** 打开详情（抽屉） */
  onOpenDetail?: () => void
  /** 清除排期（清空 start_at + due_at）— page owns PATCH */
  onClearSchedule?: () => void
  /** 排到本周 — page owns date math + PATCH */
  onScheduleThisWeek?: () => void
  disabled?: boolean
  /**
   * Optional wrap target: right-click opens the same menu.
   * If omitted, only the ⋯ button is rendered.
   */
  children?: ReactNode
  className?: string
}

function buildItems(props: TimelineBarMenuProps): MenuProps['items'] {
  return [
    {
      key: 'open',
      label: '打开详情',
      disabled: !props.onOpenDetail,
      onClick: () => props.onOpenDetail?.(),
    },
    {
      key: 'clear',
      label: '清除排期',
      disabled: !props.onClearSchedule,
      onClick: () => props.onClearSchedule?.(),
    },
    {
      key: 'this-week',
      label: '排到本周',
      disabled: !props.onScheduleThisWeek,
      onClick: () => props.onScheduleThisWeek?.(),
    },
  ]
}

/**
 * S5 presentational bar menu — Dropdown/⋯ only; no API calls.
 * Wire callbacks from TimelinePage (patchIssue / open drawer).
 */
export function TimelineBarMenu({
  onOpenDetail,
  onClearSchedule,
  onScheduleThisWeek,
  disabled = false,
  children,
  className,
}: TimelineBarMenuProps) {
  const items = buildItems({
    onOpenDetail,
    onClearSchedule,
    onScheduleThisWeek,
  })

  const btn = (
    <button
      type="button"
      className="acb-timeline-bar-menu-btn"
      aria-label="条带快捷菜单"
      disabled={disabled}
      onClick={(e) => {
        e.stopPropagation()
      }}
      onPointerDown={(e) => {
        // keep bar move/resize from capturing the ⋯ click
        e.stopPropagation()
      }}
    >
      <MoreOutlined />
    </button>
  )

  const clickMenu = (
    <Dropdown menu={{ items }} trigger={['click']} disabled={disabled} placement="bottomRight">
      {btn}
    </Dropdown>
  )

  if (!children) {
    return <span className={['acb-timeline-bar-menu', className].filter(Boolean).join(' ')}>{clickMenu}</span>
  }

  return (
    <Dropdown menu={{ items }} trigger={['contextMenu']} disabled={disabled}>
      <div className={['acb-timeline-bar-menu-wrap', className].filter(Boolean).join(' ')}>
        {children}
        <span className="acb-timeline-bar-menu">{clickMenu}</span>
      </div>
    </Dropdown>
  )
}
