/**
 * S1 empty-axis draw helpers (no API).
 * Click / minimal drag → 1-day bar; drag across days → inclusive range.
 * // S1: click = 1-day bar; drag = range
 */

export type AxisDrawRange = {
  startDay: string
  endDay: string
}

/** Map clientX inside a track element to a day index (clamped). */
export function dayIndexFromClientX(
  clientX: number,
  trackEl: HTMLElement,
  dayCount: number,
): number {
  if (dayCount <= 0) return 0
  const rect = trackEl.getBoundingClientRect()
  const w = rect.width || 1
  const x = clientX - rect.left
  const idx = Math.floor((x / w) * dayCount)
  return Math.max(0, Math.min(dayCount - 1, idx))
}

/** Inclusive ordered range from two day indices. */
export function rangeFromDayIndices(
  days: string[],
  a: number,
  b: number,
): AxisDrawRange | null {
  if (!days.length) return null
  const i0 = Math.max(0, Math.min(days.length - 1, Math.min(a, b)))
  const i1 = Math.max(0, Math.min(days.length - 1, Math.max(a, b)))
  const startDay = days[i0]
  const endDay = days[i1]
  if (!startDay || !endDay) return null
  return { startDay, endDay }
}

/**
 * Given pointer-down / pointer-up day indices, resolve the draw result.
 * Same index (click or tiny drag) → 1-day; otherwise inclusive span.
 */
export function resolveAxisDraw(
  days: string[],
  downIdx: number,
  upIdx: number,
): AxisDrawRange | null {
  return rangeFromDayIndices(days, downIdx, upIdx)
}
