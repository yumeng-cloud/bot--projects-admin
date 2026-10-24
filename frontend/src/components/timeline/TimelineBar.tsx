import type { CSSProperties, ReactNode } from 'react'
import './timeline.css'

export type TimelineBarProps = {
  /**
   * Position as percent of the axis track.
   * Prefer leftPct + widthPct; if omitted and labels given, bar still renders full-ish muted.
   */
  leftPct?: number
  widthPct?: number
  /** Optional visible label inside the bar (e.g. "10-24 → 11-02") */
  label?: ReactNode
  startLabel?: string
  endLabel?: string
  /** No dates — muted unscheduled placeholder */
  unscheduled?: boolean
  muted?: boolean
  /**
   * Raw Issue.status (backlog|todo|in_progress|in_review|done|canceled).
   * Sets `data-status` + class `is-status-${status}` for UI视觉 (no colors here).
   */
  status?: string
  /**
   * ClassName hints for page-owned DnD (no listeners here):
   * - has-drag-hint / has-resize-hint toggle CSS cursors & handle hit areas
   */
  draggableHint?: boolean
  resizableHint?: boolean
  /** Optional ⋯ / menu slot (e.g. TimelineBarMenu) — page owns callbacks */
  menu?: ReactNode
  className?: string
  style?: CSSProperties
  onClick?: () => void
  children?: ReactNode
}

function clampPct(n: number | undefined, fallback: number): number {
  if (n == null || !Number.isFinite(n)) return fallback
  return Math.min(100, Math.max(0, n))
}

/**
 * Presentational Timeline bar (right axis).
 * Positioning via left%/width%; drag/resize are className hints only — React page owns DnD.
 */
export function TimelineBar({
  leftPct,
  widthPct,
  label,
  startLabel,
  endLabel,
  unscheduled = false,
  muted = false,
  status,
  draggableHint = false,
  resizableHint = false,
  menu,
  className,
  style,
  onClick,
  children,
}: TimelineBarProps) {
  const left = clampPct(leftPct, unscheduled ? 0 : 0)
  const width = clampPct(widthPct, unscheduled ? 12 : 8)

  const text =
    label ??
    (startLabel || endLabel
      ? [startLabel, endLabel].filter(Boolean).join(' → ')
      : null)

  const statusKey =
    status && /^[a-z_]+$/.test(status) ? status : undefined

  const cls = [
    'acb-timeline-bar',
    unscheduled ? 'is-unscheduled' : '',
    muted || unscheduled ? 'is-muted' : '',
    statusKey && !unscheduled ? `is-status-${statusKey}` : '',
    draggableHint ? 'has-drag-hint' : '',
    resizableHint ? 'has-resize-hint' : '',
    className || '',
  ]
    .filter(Boolean)
    .join(' ')

  const barStyle: CSSProperties = {
    left: `${left}%`,
    width: `${Math.max(width, unscheduled ? 8 : 2)}%`,
    ...style,
  }

  return (
    <div className="acb-timeline-bar-track" role="presentation">
      <div
        className={cls}
        style={barStyle}
        role="img"
        aria-label={typeof text === 'string' ? text : undefined}
        data-status={statusKey && !unscheduled ? statusKey : undefined}
        onClick={onClick}
      >
        <span className="acb-timeline-bar-handle is-start" aria-hidden data-handle="start" />
        {text ? <span className="acb-timeline-bar-label">{text}</span> : null}
        {children}
        {menu ? <span className="acb-timeline-bar-menu-slot">{menu}</span> : null}
        <span className="acb-timeline-bar-handle is-end" aria-hidden data-handle="end" />
      </div>
    </div>
  )
}
