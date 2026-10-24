/**
 * S2 unscheduled → axis drop helpers (no API).
 * Default schedule span: dropDay → dropDay+2 (3 calendar days inclusive).
 */

/** Inclusive end offset from drop day (2 → start=drop, end=drop+2). */
export const UNSCHEDULED_DROP_END_OFFSET = 2

export type ScheduleDropRange = {
  startDay: string
  endDay: string
  dropDay: string
}

/**
 * Build default S2 range from a drop calendar day.
 * @param dropDay YYYY-MM-DD
 * @param addDays (day, n) => YYYY-MM-DD — inject page TZ helper
 * @param endOffset inclusive end offset; default UNSCHEDULED_DROP_END_OFFSET (2)
 */
export function rangeFromDropDay(
  dropDay: string,
  addDays: (day: string, n: number) => string,
  endOffset: number = UNSCHEDULED_DROP_END_OFFSET,
): ScheduleDropRange {
  const startDay = dropDay
  const endDay = addDays(dropDay, endOffset)
  return { dropDay, startDay, endDay }
}

/** HTML5 DnD mime for unscheduled issue id */
export const ACB_UNSCHEDULED_MIME = 'application/x-acb-unscheduled-issue'

export function setUnscheduledDragData(dt: DataTransfer, issueId: string): void {
  dt.setData(ACB_UNSCHEDULED_MIME, issueId)
  dt.setData('text/plain', issueId)
  dt.effectAllowed = 'move'
}

export function getUnscheduledDragIssueId(dt: DataTransfer): string | null {
  const id = dt.getData(ACB_UNSCHEDULED_MIME) || dt.getData('text/plain')
  return id ? id.trim() : null
}
