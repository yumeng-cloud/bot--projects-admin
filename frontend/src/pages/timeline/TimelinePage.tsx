import React, {
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
  type PointerEvent as ReactPointerEvent,
} from 'react'
import { Button, Empty, Spin, Typography, Tooltip, message } from 'antd'
import { listIssues, patchIssue } from '../../api/issues'
import { ApiError } from '../../api/client'
import {
  TimelineBar,
  TimelineTreeRow,
  TimelineBarMenu,
  TimelineAxisDrawTrack,
  TimelineUnscheduledDragSource,
  TimelineAxisDayDropTarget,
  TimelineWeekBar,
} from '../../components/timeline'
import { IssueDrawer } from '../../components/issue'
import { MockBanner } from '../../components/common/MockBanner'
import type { Issue } from '../../types/issue'
import { STATUS_LABELS } from '../../types/issue'
import { dayjs, TZ } from '../../utils/time'
import type { Dayjs } from 'dayjs'

export type TimelinePageProps = {
  workspaceId: string
  projectId: string
}

type TreeNode = {
  issue: Issue
  children: TreeNode[]
}

type FlatRow = {
  issue: Issue
  depth: number
  expandable: boolean
  isLeaf: boolean
}

type DragKind = 'move' | 'resize-start' | 'resize-end'

type DragState = {
  issueId: string
  kind: DragKind
  originX: number
  /** calendar days at pointer-down */
  startDay: string
  endDay: string
  /** live preview while dragging */
  previewStart: string
  previewEnd: string
}

const AXIS_DAYS = 14 // 默认两周（周视图密度）
const DATE_FMT = 'YYYY-MM-DD'

/** 归一为 Asia/Shanghai 日历日 YYYY-MM-DD */
function toCalendarDay(v?: string | null): string | null {
  if (!v) return null
  const s = String(v).trim()
  if (!s) return null
  if (/^\d{4}-\d{2}-\d{2}$/.test(s)) return s
  const d = dayjs.utc(s).tz(TZ)
  return d.isValid() ? d.format(DATE_FMT) : null
}

function resolveBarRange(issue: Issue): { start: string; end: string } | null {
  const start = toCalendarDay(issue.start_at)
  const end = toCalendarDay(issue.due_at)
  if (!start && !end) return null
  if (start && end) return start <= end ? { start, end } : { start: end, end: start }
  if (start) return { start, end: start }
  return { start: end!, end: end! }
}

function isUnscheduled(issue: Issue): boolean {
  return !toCalendarDay(issue.start_at) && !toCalendarDay(issue.due_at)
}

function addDays(day: string, n: number): string {
  return dayjs.tz(day, TZ).add(n, 'day').format(DATE_FMT)
}

function dayDiff(a: string, b: string): number {
  return dayjs.tz(a, TZ).startOf('day').diff(dayjs.tz(b, TZ).startOf('day'), 'day')
}

function buildForest(issues: Issue[]): TreeNode[] {
  const byParent = new Map<string | null, Issue[]>()
  const ids = new Set(issues.map((i) => i.id))
  for (const iss of issues) {
    let pid = iss.parent_id
    if (pid && !ids.has(pid)) pid = null
    const list = byParent.get(pid) || []
    list.push(iss)
    byParent.set(pid, list)
  }
  const sortFn = (a: Issue, b: Issue) =>
    a.board_order - b.board_order || a.number - b.number || a.identifier.localeCompare(b.identifier)

  const walk = (parentId: string | null): TreeNode[] => {
    const kids = (byParent.get(parentId) || []).slice().sort(sortFn)
    return kids.map((issue) => ({
      issue,
      children: walk(issue.id),
    }))
  }
  return walk(null)
}

function childCountMap(issues: Issue[]): Map<string, number> {
  const m = new Map<string, number>()
  for (const i of issues) {
    if (!i.parent_id) continue
    m.set(i.parent_id, (m.get(i.parent_id) || 0) + 1)
  }
  return m
}

function flattenVisible(
  forest: TreeNode[],
  expanded: Set<string>,
  counts: Map<string, number>,
): FlatRow[] {
  const out: FlatRow[] = []
  const walk = (nodes: TreeNode[], depth: number) => {
    for (const n of nodes) {
      const kids =
        typeof n.issue.child_count === 'number'
          ? n.issue.child_count
          : counts.get(n.issue.id) || n.children.length
      const expandable = kids > 0
      const isLeaf =
        typeof n.issue.is_leaf === 'boolean' ? n.issue.is_leaf : kids === 0
      out.push({ issue: n.issue, depth, expandable, isLeaf })
      if (expandable && expanded.has(n.issue.id)) {
        walk(n.children, depth + 1)
      }
    }
  }
  walk(forest, 0)
  return out
}

function pctForDay(day: string, rangeStart: string, days: number): number {
  const idx = dayDiff(day, rangeStart)
  if (days <= 0) return 0
  return (idx / days) * 100
}

function widthPctForRange(start: string, end: string, days: number): number {
  if (days <= 0) return 0
  const span = dayDiff(end, start) + 1
  return (Math.max(span, 1) / days) * 100
}

function formatBarLabel(start: string, end: string): string {
  const a = dayjs.tz(start, TZ).format('M-D')
  const b = dayjs.tz(end, TZ).format('M-D')
  return a === b ? a : `${a} → ${b}`
}

function formatTooltipRange(start: string, end: string): string {
  const a = dayjs.tz(start, TZ).format('MM/DD')
  const b = dayjs.tz(end, TZ).format('MM/DD')
  return a === b ? a : `${a} – ${b}`
}

function errorMessage(err: unknown, fallback: string): string {
  if (err instanceof ApiError) return err.message || fallback
  if (err instanceof Error && err.message) return err.message
  return fallback
}

export default function TimelinePage({ workspaceId, projectId }: TimelinePageProps) {
  const [loading, setLoading] = useState(true)
  const [issues, setIssues] = useState<Issue[]>([])
  const [expanded, setExpanded] = useState<Set<string>>(() => new Set())
  const [rangeStart, setRangeStart] = useState<string>(() =>
    dayjs().tz(TZ).startOf('week').format(DATE_FMT),
  )
  const [drawerOpen, setDrawerOpen] = useState(false)
  const [current, setCurrent] = useState<Issue | null>(null)
  const [drag, setDrag] = useState<DragState | null>(null)
  const [patchingId, setPatchingId] = useState<string | null>(null)

  const axisRef = useRef<HTMLDivElement | null>(null)
  const dragRef = useRef<DragState | null>(null)
  dragRef.current = drag
  /** S9: onUp clears drag before click; suppress openIssue briefly after real drag/resize */
  const suppressOpenRef = useRef(false)

  const load = useCallback(async () => {
    setLoading(true)
    try {
      const { issues: list } = await listIssues(workspaceId, { project_id: projectId })
      setIssues(list)
      // 默认展开有子节点的前几层，便于演示 ≥3 层
      const counts = childCountMap(list)
      const next = new Set<string>()
      for (const i of list) {
        if ((counts.get(i.id) || 0) > 0) next.add(i.id)
      }
      setExpanded(next)
    } catch (err) {
      message.error(errorMessage(err, '加载 Issue 失败'))
    } finally {
      setLoading(false)
    }
  }, [workspaceId, projectId])

  useEffect(() => {
    void load()
  }, [load])

  const counts = useMemo(() => childCountMap(issues), [issues])
  const scheduledIssues = useMemo(() => issues.filter((i) => !isUnscheduled(i)), [issues])
  const forest = useMemo(() => buildForest(scheduledIssues), [scheduledIssues])
  const rows = useMemo(
    () => flattenVisible(forest, expanded, counts),
    [forest, expanded, counts],
  )
  const unscheduled = useMemo(() => issues.filter(isUnscheduled), [issues])

  const axisDays = useMemo(() => {
    const days: Dayjs[] = []
    const start = dayjs.tz(rangeStart, TZ)
    for (let i = 0; i < AXIS_DAYS; i++) days.push(start.add(i, 'day'))
    return days
  }, [rangeStart])

  const axisDayStrs = useMemo(() => axisDays.map((d) => d.format(DATE_FMT)), [axisDays])

  /** S6: today column position for vertical line (UI 视觉 CSS) */
  const todayAxisStyle = useMemo(() => {
    const today = dayjs().tz(TZ).format(DATE_FMT)
    const idx = axisDays.findIndex((d) => d.format(DATE_FMT) === today)
    if (idx < 0) return null
    const left = (idx / AXIS_DAYS) * 100
    const width = 100 / AXIS_DAYS
    return {
      ['--acb-timeline-today-left' as string]: `${left}%`,
      ['--acb-timeline-today-width' as string]: `${width}%`,
    } as React.CSSProperties
  }, [axisDays])

  const rangeEnd = useMemo(
    () => addDays(rangeStart, AXIS_DAYS - 1),
    [rangeStart],
  )

  const issueById = useMemo(() => {
    const m = new Map<string, Issue>()
    for (const i of issues) m.set(i.id, i)
    return m
  }, [issues])

  const toggleExpand = (id: string) => {
    setExpanded((prev) => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }

  const openIssue = (issue: Issue) => {
    // S9: consume suppress from effective drag/resize (pure click never sets it)
    if (suppressOpenRef.current) {
      suppressOpenRef.current = false
      return
    }
    setCurrent(issue)
    setDrawerOpen(true)
  }

  const shiftAxis = (weeks: number) => {
    setRangeStart((s) => addDays(s, weeks * 7))
  }

  const applyLocalDates = (id: string, start: string, end: string, versionBump?: number) => {
    setIssues((prev) =>
      prev.map((i) =>
        i.id === id
          ? {
              ...i,
              start_at: start,
              due_at: end,
              version: versionBump != null ? versionBump : i.version,
            }
          : i,
      ),
    )
    setCurrent((c) =>
      c && c.id === id
        ? {
            ...c,
            start_at: start,
            due_at: end,
            version: versionBump != null ? versionBump : c.version,
          }
        : c,
    )
  }

  const commitDates = async (
    issue: Issue,
    start: string,
    end: string,
    opts?: { silentOk?: boolean },
  ) => {
    const prev = issue
    // 乐观展示
    applyLocalDates(issue.id, start, end)
    setPatchingId(issue.id)
    try {
      const { issue: saved } = await patchIssue(workspaceId, issue.id, {
        version: issue.version,
        start_at: start,
        due_at: end,
      })
      setIssues((list) => list.map((i) => (i.id === saved.id ? saved : i)))
      setCurrent((c) => (c && c.id === saved.id ? saved : c))
      if (!opts?.silentOk) message.success('已更新排期')
    } catch (err) {
      // 回滚；client 已 toast，再补一句中文说明
      setIssues((list) => list.map((i) => (i.id === prev.id ? prev : i)))
      setCurrent((c) => (c && c.id === prev.id ? prev : c))
      const msg = errorMessage(err, '改期失败')
      if (!(err instanceof ApiError)) {
        message.error(msg)
      } else if (/unknown|not allowed|no such|unsupported|未支持|字段/i.test(msg)) {
        message.error(`排期字段尚未接通后端：${msg}（本地已回滚）`)
      }
      // ApiError 时 client 已 toast；字段未通时再强调
    } finally {
      setPatchingId(null)
    }
  }

  const scheduleThisWeek = async (issue: Issue) => {
    const start = dayjs().tz(TZ).format(DATE_FMT)
    const end = addDays(start, 6)
    await commitDates(issue, start, end)
  }

  const clearSchedule = async (issue: Issue) => {
    const prev = issue
    setIssues((list) =>
      list.map((i) =>
        i.id === issue.id ? { ...i, start_at: null, due_at: null } : i,
      ),
    )
    setCurrent((c) =>
      c && c.id === issue.id ? { ...c, start_at: null, due_at: null } : c,
    )
    setPatchingId(issue.id)
    try {
      const { issue: saved } = await patchIssue(workspaceId, issue.id, {
        version: issue.version,
        start_at: null,
        due_at: null,
      })
      setIssues((list) => list.map((i) => (i.id === saved.id ? saved : i)))
      setCurrent((c) => (c && c.id === saved.id ? saved : c))
      message.success('已清除排期')
    } catch (err) {
      setIssues((list) => list.map((i) => (i.id === prev.id ? prev : i)))
      setCurrent((c) => (c && c.id === prev.id ? prev : c))
      if (!(err instanceof ApiError)) message.error(errorMessage(err, '清除排期失败'))
    } finally {
      setPatchingId(null)
    }
  }

  const dayWidthPx = (): number => {
    const el = axisRef.current
    if (!el) return 40
    return el.clientWidth / AXIS_DAYS
  }

  const onBarPointerDown = (
    e: ReactPointerEvent<HTMLDivElement>,
    issue: Issue,
    kindHint?: DragKind,
  ) => {
    if (e.button !== 0) return
    const range = resolveBarRange(issue)
    if (!range) return
    const handle = (e.target as HTMLElement | null)?.closest?.('[data-handle]') as
      | HTMLElement
      | null
    const handleKind = handle?.dataset?.handle
    let kind: DragKind = kindHint || 'move'
    if (handleKind === 'start') kind = 'resize-start'
    else if (handleKind === 'end') kind = 'resize-end'

    e.preventDefault()
    e.stopPropagation()
    ;(e.currentTarget as HTMLElement).setPointerCapture?.(e.pointerId)

    const next: DragState = {
      issueId: issue.id,
      kind,
      originX: e.clientX,
      startDay: range.start,
      endDay: range.end,
      previewStart: range.start,
      previewEnd: range.end,
    }
    setDrag(next)
  }

  useEffect(() => {
    if (!drag) return

    const onMove = (e: PointerEvent) => {
      const d = dragRef.current
      if (!d) return
      const delta = Math.round((e.clientX - d.originX) / dayWidthPx())
      let start = d.startDay
      let end = d.endDay
      if (d.kind === 'move') {
        start = addDays(d.startDay, delta)
        end = addDays(d.endDay, delta)
      } else if (d.kind === 'resize-start') {
        start = addDays(d.startDay, delta)
        if (start > end) start = end
      } else {
        end = addDays(d.endDay, delta)
        if (end < start) end = start
      }
      if (start === d.previewStart && end === d.previewEnd) return
      setDrag({ ...d, previewStart: start, previewEnd: end })
    }

    const onUp = () => {
      const d = dragRef.current
      // S9: only suppress drawer open when dates actually changed (pure click still opens)
      const moved = !!(
        d &&
        (d.previewStart !== d.startDay || d.previewEnd !== d.endDay)
      )
      if (moved) {
        suppressOpenRef.current = true
        window.setTimeout(() => {
          suppressOpenRef.current = false
        }, 300)
      }
      setDrag(null)
      if (!d) return
      if (!moved) return
      const issue = issueById.get(d.issueId)
      if (!issue) return
      void commitDates(issue, d.previewStart, d.previewEnd, { silentOk: true })
    }

    window.addEventListener('pointermove', onMove)
    window.addEventListener('pointerup', onUp)
    window.addEventListener('pointercancel', onUp)
    return () => {
      window.removeEventListener('pointermove', onMove)
      window.removeEventListener('pointerup', onUp)
      window.removeEventListener('pointercancel', onUp)
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [drag, issueById, workspaceId])

  const effectiveRange = (issue: Issue): { start: string; end: string } | null => {
    if (drag && drag.issueId === issue.id) {
      return { start: drag.previewStart, end: drag.previewEnd }
    }
    return resolveBarRange(issue)
  }

  return (
    <div className="acb-page acb-timeline">
      <MockBanner />
      <div className="acb-page-bar" style={{ marginBottom: 8 }}>
        <div>
          <Typography.Title level={4} style={{ margin: 0, fontSize: 15 }}>
            排期
          </Typography.Title>
          <Typography.Text type="secondary" style={{ fontSize: 12 }}>
            左树右轴 · 默认周视图 · 拖条改期（Asia/Shanghai 日历日）
          </Typography.Text>
        </div>
        <TimelineWeekBar
          onPrevWeek={() => shiftAxis(-1)}
          onThisWeek={() => setRangeStart(dayjs().tz(TZ).startOf('week').format(DATE_FMT))}
          onNextWeek={() => shiftAxis(1)}
          onRefresh={() => void load()}
          refreshDisabled={!!drag}
        />
      </div>

      {loading ? (
        <div style={{ textAlign: 'center', padding: 32 }}>
          <Spin size="small" />
        </div>
      ) : issues.length === 0 ? (
        <Empty style={{ marginTop: 48 }} description="还没有 Issue" />
      ) : (
        <>
          <div
            className="acb-timeline-shell"
            style={{ ['--acb-timeline-axis-days']: String(AXIS_DAYS) } as React.CSSProperties}
          >
            <div className="acb-timeline-tree-pane">
              <div className="acb-timeline-tree-head">Issue</div>
              {rows.map((row) => (
                <TimelineTreeRow
                  key={row.issue.id}
                  level={row.depth}
                  expandable={row.expandable}
                  expanded={expanded.has(row.issue.id)}
                  onToggleExpand={() => toggleExpand(row.issue.id)}
                  identifier={row.issue.identifier}
                  title={row.issue.title}
                  leafBadge={row.isLeaf}
                  selected={current?.id === row.issue.id}
                  onClick={() => openIssue(row.issue)}
                  trailing={
                    <span
                      style={{
                        fontSize: 10,
                        color: 'var(--acb-text-tertiary)',
                        flex: '0 0 auto',
                        maxWidth: 48,
                        overflow: 'hidden',
                        textOverflow: 'ellipsis',
                      }}
                      title={STATUS_LABELS[row.issue.status]}
                    >
                      {STATUS_LABELS[row.issue.status]}
                    </span>
                  }
                />
              ))}
            </div>
            <div className="acb-timeline-axis-pane">
              <TimelineAxisDayDropTarget
                days={axisDayStrs}
                addDays={addDays}
                onDropIssue={({ issueId, startDay, endDay }) => {
                  const iss = issueById.get(issueId)
                  if (iss) void commitDates(iss, startDay, endDay)
                }}
              >
                <div
                  className={
                    'acb-timeline-axis-inner' + (todayAxisStyle ? ' has-today-line' : '')
                  }
                  ref={axisRef}
                  style={todayAxisStyle || undefined}
                >
                  <div className="acb-timeline-axis-head">
                    {axisDays.map((d) => {
                      const isToday = d.format(DATE_FMT) === dayjs().tz(TZ).format(DATE_FMT)
                      const isWeekStart = d.day() === 1 || d.date() === 1
                      return (
                        <div
                          key={d.format(DATE_FMT)}
                          className={
                            'acb-timeline-axis-day' +
                            (isToday ? ' is-today' : '') +
                            (isToday || isWeekStart ? ' is-mark' : '')
                          }
                          title={d.format('YYYY-MM-DD ddd')}
                        >
                          {d.format('M/D')}
                        </div>
                      )
                    })}
                  </div>
                  {rows.map((row) => {
                    const range = effectiveRange(row.issue)!
                    const left = pctForDay(range.start, rangeStart, AXIS_DAYS)
                    const width = widthPctForRange(range.start, range.end, AXIS_DAYS)
                    const inView = range.end >= rangeStart && range.start <= rangeEnd
                    const tip = formatTooltipRange(range.start, range.end)
                    return (
                      <TimelineAxisDrawTrack
                        key={row.issue.id}
                        days={axisDayStrs}
                        onDrawComplete={({ startDay, endDay }) => {
                          void commitDates(row.issue, startDay, endDay)
                        }}
                      >
                        <div
                          onPointerDown={(e) => onBarPointerDown(e, row.issue)}
                          onClick={(e) => {
                            // S9: dragRef 在 onUp 已清空，还需看 suppressOpenRef
                            if (suppressOpenRef.current) {
                              suppressOpenRef.current = false
                              return
                            }
                            if (dragRef.current) return
                            if ((e.target as HTMLElement).closest?.('[data-handle]')) return
                            openIssue(row.issue)
                          }}
                          style={{
                            opacity: patchingId === row.issue.id ? 0.7 : inView ? 1 : 0.45,
                            cursor: 'grab',
                          }}
                        >
                          <Tooltip title={tip}>
                            <div>
                              <TimelineBar
                                leftPct={left}
                                widthPct={Math.max(width, 100 / AXIS_DAYS)}
                                label={formatBarLabel(range.start, range.end)}
                                startLabel={range.start}
                                endLabel={range.end}
                                status={row.issue.status}
                                draggableHint
                                resizableHint
                                menu={
                                  <TimelineBarMenu
                                    onOpenDetail={() => openIssue(row.issue)}
                                    onClearSchedule={() => void clearSchedule(row.issue)}
                                    onScheduleThisWeek={() => void scheduleThisWeek(row.issue)}
                                    disabled={patchingId === row.issue.id}
                                  />
                                }
                              />
                            </div>
                          </Tooltip>
                        </div>
                      </TimelineAxisDrawTrack>
                    )
                  })}
                </div>
              </TimelineAxisDayDropTarget>
            </div>
          </div>

          {/* 未排期桶 */}
          {unscheduled.length > 0 && (
            <div style={{ marginTop: 16 }}>
              <Typography.Text
                type="secondary"
                style={{ fontSize: 12, display: 'block', marginBottom: 6 }}
              >
                未排期（{unscheduled.length}）— 拖到上方轴松手落该日起 3 天；或「排到本周」/打开抽屉
              </Typography.Text>
              <div
                style={{
                  border: '1px solid var(--acb-border)',
                  borderRadius: 6,
                  overflow: 'hidden',
                  background: 'var(--acb-bg-container)',
                }}
              >
                {unscheduled.map((iss) => (
                  <TimelineUnscheduledDragSource
                    key={iss.id}
                    issueId={iss.id}
                    disabled={patchingId === iss.id}
                  >
                    <div
                      style={{
                        display: 'grid',
                        gridTemplateColumns: 'minmax(240px, 320px) 1fr',
                        alignItems: 'center',
                      }}
                    >
                      <TimelineTreeRow
                        identifier={iss.identifier}
                        title={iss.title}
                        unscheduled
                        leafBadge={(counts.get(iss.id) || 0) === 0}
                        onClick={() => openIssue(iss)}
                        trailing={
                          <Button
                            type="link"
                            size="small"
                            style={{ fontSize: 11, padding: 0, height: 'auto' }}
                            disabled={patchingId === iss.id}
                            onClick={(e) => {
                              e.stopPropagation()
                              void scheduleThisWeek(iss)
                            }}
                          >
                            排到本周
                          </Button>
                        }
                      />
                      <div onClick={() => openIssue(iss)} style={{ cursor: 'pointer' }}>
                        <TimelineBar unscheduled label="未排期" />
                      </div>
                    </div>
                  </TimelineUnscheduledDragSource>
                ))}
              </div>
            </div>
          )}
        </>
      )}

      <IssueDrawer
        open={drawerOpen}
        issue={current}
        workspaceId={workspaceId}
        projectId={projectId}
        onClose={() => setDrawerOpen(false)}
        onSaved={(saved) => {
          setIssues((prev) => prev.map((i) => (i.id === saved.id ? saved : i)))
          setCurrent(saved)
        }}
      />
    </div>
  )
}
