import {
  useCallback,
  useRef,
  useState,
  type PointerEvent as ReactPointerEvent,
  type ReactNode,
  type CSSProperties,
} from 'react'
import {
  dayIndexFromClientX,
  resolveAxisDraw,
  type AxisDrawRange,
} from './axisDraw'
import './timeline.css'

export type TimelineAxisDrawTrackProps = {
  /** Ordered axis calendar days YYYY-MM-DD */
  days: string[]
  /**
   * Fired on pointer-up after a press-drag (or click) on blank track.
   * // S1: click = 1-day bar; drag = range
   */
  onDrawComplete?: (range: AxisDrawRange) => void
  /** Live preview while drawing; null on cancel/up */
  onDrawPreview?: (range: AxisDrawRange | null) => void
  disabled?: boolean
  className?: string
  style?: CSSProperties
  children?: ReactNode
}

type DrawState = {
  pointerId: number
  downIdx: number
  preview: AxisDrawRange
}

/**
 * S1 presentational empty-axis draw track.
 * Emits {startDay,endDay} only — page owns patchIssue.
 * Ignores pointer-down that starts on a bar / handle / menu (data-handle, .acb-timeline-bar).
 */
export function TimelineAxisDrawTrack({
  days,
  onDrawComplete,
  onDrawPreview,
  disabled = false,
  className,
  style,
  children,
}: TimelineAxisDrawTrackProps) {
  const trackRef = useRef<HTMLDivElement | null>(null)
  const drawRef = useRef<DrawState | null>(null)
  const [preview, setPreview] = useState<AxisDrawRange | null>(null)

  const clearPreview = useCallback(() => {
    drawRef.current = null
    setPreview(null)
    onDrawPreview?.(null)
  }, [onDrawPreview])

  const onPointerDown = (e: ReactPointerEvent<HTMLDivElement>) => {
    if (disabled || e.button !== 0) return
    const t = e.target as HTMLElement | null
    if (t?.closest?.('.acb-timeline-bar, [data-handle], .acb-timeline-bar-menu-btn, .acb-timeline-bar-menu')) {
      return
    }
    const el = trackRef.current
    if (!el || days.length === 0) return
    e.preventDefault()
    el.setPointerCapture?.(e.pointerId)
    const downIdx = dayIndexFromClientX(e.clientX, el, days.length)
    const range = resolveAxisDraw(days, downIdx, downIdx)
    if (!range) return
    drawRef.current = { pointerId: e.pointerId, downIdx, preview: range }
    setPreview(range)
    onDrawPreview?.(range)
  }

  const onPointerMove = (e: ReactPointerEvent<HTMLDivElement>) => {
    const d = drawRef.current
    const el = trackRef.current
    if (!d || !el || e.pointerId !== d.pointerId) return
    const upIdx = dayIndexFromClientX(e.clientX, el, days.length)
    const range = resolveAxisDraw(days, d.downIdx, upIdx)
    if (!range) return
    if (range.startDay === d.preview.startDay && range.endDay === d.preview.endDay) return
    drawRef.current = { ...d, preview: range }
    setPreview(range)
    onDrawPreview?.(range)
  }

  const finish = (e: ReactPointerEvent<HTMLDivElement>) => {
    const d = drawRef.current
    if (!d || e.pointerId !== d.pointerId) return
    const el = trackRef.current
    let range = d.preview
    if (el) {
      const upIdx = dayIndexFromClientX(e.clientX, el, days.length)
      range = resolveAxisDraw(days, d.downIdx, upIdx) || d.preview
    }
    clearPreview()
    onDrawComplete?.(range)
  }

  const pctLeft =
    preview && days.length
      ? (days.indexOf(preview.startDay) / days.length) * 100
      : 0
  const pctWidth =
    preview && days.length
      ? ((days.indexOf(preview.endDay) - days.indexOf(preview.startDay) + 1) / days.length) *
        100
      : 0

  return (
    <div
      ref={trackRef}
      className={['acb-timeline-axis-draw-track', className].filter(Boolean).join(' ')}
      style={style}
      role="presentation"
      onPointerDown={onPointerDown}
      onPointerMove={onPointerMove}
      onPointerUp={finish}
      onPointerCancel={clearPreview}
    >
      {children}
      {preview ? (
        <div
          className="acb-timeline-axis-draw-preview"
          style={{ left: `${pctLeft}%`, width: `${Math.max(pctWidth, 100 / Math.max(days.length, 1))}%` }}
          aria-hidden
        />
      ) : null}
    </div>
  )
}

export type { AxisDrawRange }
