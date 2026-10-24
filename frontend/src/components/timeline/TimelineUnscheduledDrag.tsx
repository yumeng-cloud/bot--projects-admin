import type { DragEvent as ReactDragEvent, ReactNode, CSSProperties } from 'react'
import {
  ACB_UNSCHEDULED_MIME,
  UNSCHEDULED_DROP_END_OFFSET,
  getUnscheduledDragIssueId,
  rangeFromDropDay,
  setUnscheduledDragData,
  type ScheduleDropRange,
} from './scheduleDrop'
import { dayIndexFromClientX } from './axisDraw'
import './timeline.css'

export type TimelineUnscheduledDragSourceProps = {
  issueId: string
  disabled?: boolean
  className?: string
  style?: CSSProperties
  children: ReactNode
}

/**
 * S2 presentational drag source for an unscheduled bucket row.
 * Uses HTML5 DnD; page listens on axis drop target.
 */
export function TimelineUnscheduledDragSource({
  issueId,
  disabled = false,
  className,
  style,
  children,
}: TimelineUnscheduledDragSourceProps) {
  return (
    <div
      className={['acb-timeline-unscheduled-drag-source', className].filter(Boolean).join(' ')}
      style={style}
      draggable={!disabled}
      onDragStart={(e: ReactDragEvent) => {
        if (disabled) {
          e.preventDefault()
          return
        }
        setUnscheduledDragData(e.dataTransfer, issueId)
      }}
      data-issue-id={issueId}
    >
      {children}
    </div>
  )
}

export type TimelineAxisDayDropTargetProps = {
  /** Ordered axis days YYYY-MM-DD */
  days: string[]
  /**
   * S2: on drop emits issueId + range.
   * Duration default dropDay → dropDay+2 (see UNSCHEDULED_DROP_END_OFFSET).
   */
  onDropIssue?: (payload: { issueId: string } & ScheduleDropRange) => void
  /** Inject page TZ addDays */
  addDays: (day: string, n: number) => string
  /** Inclusive end offset; default 2 → 3 calendar days */
  dropSpanEndOffset?: number
  disabled?: boolean
  className?: string
  style?: CSSProperties
  children?: ReactNode
}

/**
 * S2 presentational axis drop target (wrap axis head or a row track).
 * Resolves drop column from pointer X; does not call API.
 */
export function TimelineAxisDayDropTarget({
  days,
  onDropIssue,
  addDays,
  dropSpanEndOffset = UNSCHEDULED_DROP_END_OFFSET,
  disabled = false,
  className,
  style,
  children,
}: TimelineAxisDayDropTargetProps) {
  const onDragOver = (e: ReactDragEvent) => {
    if (disabled) return
    const types = Array.from(e.dataTransfer.types || [])
    if (
      types.includes(ACB_UNSCHEDULED_MIME) ||
      types.includes('text/plain') ||
      types.includes('Text')
    ) {
      e.preventDefault()
      e.dataTransfer.dropEffect = 'move'
      e.currentTarget.classList.add('is-drop-over')
    }
  }

  const onDragLeave = (e: ReactDragEvent) => {
    e.currentTarget.classList.remove('is-drop-over')
  }

  const onDrop = (e: ReactDragEvent) => {
    e.currentTarget.classList.remove('is-drop-over')
    if (disabled) return
    e.preventDefault()
    const issueId = getUnscheduledDragIssueId(e.dataTransfer)
    if (!issueId || days.length === 0) return
    const idx = dayIndexFromClientX(e.clientX, e.currentTarget as HTMLElement, days.length)
    const dropDay = days[idx]
    if (!dropDay) return
    const range = rangeFromDropDay(dropDay, addDays, dropSpanEndOffset)
    onDropIssue?.({ issueId, ...range })
  }

  return (
    <div
      className={['acb-timeline-axis-drop-target', className].filter(Boolean).join(' ')}
      style={style}
      onDragOver={onDragOver}
      onDragLeave={onDragLeave}
      onDrop={onDrop}
      data-acb-drop-span={String(dropSpanEndOffset)}
    >
      {children}
    </div>
  )
}

export {
  UNSCHEDULED_DROP_END_OFFSET,
  rangeFromDropDay,
  ACB_UNSCHEDULED_MIME,
  setUnscheduledDragData,
  getUnscheduledDragIssueId,
} from './scheduleDrop'
export type { ScheduleDropRange } from './scheduleDrop'
