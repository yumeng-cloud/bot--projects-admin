import type { CSSProperties, MouseEvent, ReactNode } from 'react'
import { CaretDownOutlined, CaretRightOutlined } from '@ant-design/icons'
import './timeline.css'

export type TimelineTreeRowProps = {
  /** Nesting depth; indent = level × CSS indent step */
  level?: number
  /** Show expand chevron (parent with children). Hide slot when leaf. */
  expandable?: boolean
  expanded?: boolean
  onToggleExpand?: (e: MouseEvent) => void
  identifier: string
  title: string
  /** Optional leaf marker; `true` → default 「叶」 */
  leafBadge?: ReactNode | boolean
  /** Muted styling for issues without start/due */
  unscheduled?: boolean
  selected?: boolean
  onClick?: () => void
  className?: string
  style?: CSSProperties
  /** Extra trailing content (status / assignee chip) — page owns data */
  trailing?: ReactNode
}

/**
 * Presentational Timeline left-tree row (Linear Roadmap density).
 * No data fetching / no tree state — parent owns expand + selection.
 */
export function TimelineTreeRow({
  level = 0,
  expandable = false,
  expanded = false,
  onToggleExpand,
  identifier,
  title,
  leafBadge,
  unscheduled = false,
  selected = false,
  onClick,
  className,
  style,
  trailing,
}: TimelineTreeRowProps) {
  const depth = Math.max(0, level)
  const cls = [
    'acb-timeline-tree-row',
    selected ? 'is-selected' : '',
    unscheduled ? 'is-unscheduled' : '',
    className || '',
  ]
    .filter(Boolean)
    .join(' ')

  return (
    <div
      className={cls}
      style={style}
      role="row"
      aria-selected={selected || undefined}
      aria-expanded={expandable ? expanded : undefined}
      onClick={onClick}
    >
      <span
        className="acb-timeline-tree-indent"
        style={{ width: depth * 16 }}
        aria-hidden
      />
      {expandable ? (
        <button
          type="button"
          className={[
            'acb-timeline-tree-chevron',
            expanded ? 'is-expanded' : '',
          ]
            .filter(Boolean)
            .join(' ')}
          aria-label={expanded ? '折叠' : '展开'}
          onClick={(e) => {
            e.stopPropagation()
            onToggleExpand?.(e)
          }}
        >
          {expanded ? <CaretDownOutlined /> : <CaretRightOutlined />}
        </button>
      ) : (
        <span className="acb-timeline-tree-chevron is-leaf-slot" aria-hidden />
      )}
      <span className="acb-timeline-tree-id" title={identifier}>
        {identifier}
      </span>
      <span className="acb-timeline-tree-title" title={title}>
        {title}
      </span>
      {leafBadge != null && leafBadge !== false ? (
        <span className="acb-timeline-leaf-badge">{leafBadge === true ? '叶' : leafBadge}</span>
      ) : null}
      {trailing}
    </div>
  )
}
