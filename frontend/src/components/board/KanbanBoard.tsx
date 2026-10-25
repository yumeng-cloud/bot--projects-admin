import {
  DndContext,
  DragOverlay,
  PointerSensor,
  closestCorners,
  defaultDropAnimationSideEffects,
  useSensor,
  useSensors,
  type DragEndEvent,
  type DragOverEvent,
  type DragStartEvent,
  type DropAnimation,
} from '@dnd-kit/core'
import { useCallback, useEffect, useMemo, useState } from 'react'
import { Modal, message } from 'antd'
import { patchIssue, listProjectMembers } from '../../api/issues'
import type { Issue, IssueStatus } from '../../types/issue'
import { memberLabel } from '../../types/issue'
import { BOARD_COLUMNS } from '../../types/issue'
import { BoardColumn } from './BoardColumn'
import { BoardCardView } from './BoardCard'
import { columnLabel, type BoardColumnDef } from './boardTemplates'
import './board.css'

/** Snappy settle without overshoot; keep source ghosted during overlay return (cancel/drop). */
const dropAnimation: DropAnimation = {
  duration: 160,
  easing: 'cubic-bezier(0.2, 0, 0, 1)',
  sideEffects: defaultDropAnimationSideEffects({
    styles: {
      active: { opacity: '0.28' },
    },
  }),
}

type Props = {
  workspaceId: string
  issues: Issue[]
  selectedId?: string | null
  onChange: (issues: Issue[]) => void
  onOpen: (issue: Issue) => void
  /** Ordered visible columns from prefs; defaults to BOARD_COLUMNS */
  columns?: BoardColumnDef[]
  /**
   * C1：归档一批 done Issue（调用方发 POST …/archive，成功后本地去卡）。
   * 返回 Promise；失败应抛错以便本组件回滚勾选态。
   */
  onArchiveIssues?: (ids: string[]) => Promise<void>
  archiving?: boolean
}

function statusFromOver(overId: string | number, issues: Issue[]): IssueStatus | null {
  const id = String(overId)
  if (id.startsWith('col-')) return id.slice(4) as IssueStatus
  const hit = issues.find((i) => i.id === id)
  return hit?.status ?? null
}

export function KanbanBoard({
  workspaceId,
  issues,
  selectedId,
  onChange,
  onOpen,
  columns,
  onArchiveIssues,
  archiving,
}: Props) {
  const [activeId, setActiveId] = useState<string | null>(null)
  const [overId, setOverId] = useState<string | null>(null)
  const [assigneeNames, setAssigneeNames] = useState<Record<string, string>>({})
  const [selectedIds, setSelectedIds] = useState<Set<string>>(() => new Set())
  // MOBILE_RESPONSIVE：抬高激活距，窄屏点按优先开卡、不挡横滑
  const sensors = useSensors(
    useSensor(PointerSensor, {
      activationConstraint: {
        distance: typeof window !== 'undefined' && window.matchMedia('(max-width: 639px)').matches ? 14 : 6,
      },
    }),
  )

  const visibleCols = useMemo(() => {
    if (columns && columns.length > 0) {
      return columns.filter((c) => !c.hidden)
    }
    return BOARD_COLUMNS.map((status) => ({ status }))
  }, [columns])

  const statusOrder = useMemo(() => visibleCols.map((c) => c.status), [visibleCols])

  useEffect(() => {
    let cancelled = false
    listProjectMembers(workspaceId)
      .then(({ members }) => {
        if (cancelled) return
        const map: Record<string, string> = {}
        for (const m of members) map[m.id] = memberLabel(m)
        setAssigneeNames(map)
      })
      .catch(() => {})
    return () => {
      cancelled = true
    }
  }, [workspaceId])

  const byStatus = useMemo(() => {
    const map: Record<string, Issue[]> = {}
    for (const s of statusOrder) map[s] = []
    for (const i of issues) {
      if (!map[i.status]) continue
      map[i.status].push(i)
    }
    for (const s of statusOrder) {
      map[s].sort((a, b) => a.board_order - b.board_order)
    }
    return map
  }, [issues, statusOrder])

  const active = issues.find((i) => i.id === activeId) || null

  // 卡片被移除（归档）后清掉失效勾选
  useEffect(() => {
    setSelectedIds((prev) => {
      if (prev.size === 0) return prev
      let changed = false
      const next = new Set<string>()
      const idSet = new Set(issues.map((i) => i.id))
      for (const id of prev) {
        if (idSet.has(id)) next.add(id)
        else changed = true
      }
      return changed ? next : prev
    })
  }, [issues])

  const onToggleSelect = useCallback((issue: Issue, next: boolean) => {
    if (issue.status !== 'done') return
    setSelectedIds((prev) => {
      const n = new Set(prev)
      if (next) n.add(issue.id)
      else n.delete(issue.id)
      return n
    })
  }, [])

  const archiveIds = useCallback(
    async (ids: string[]) => {
      const doneIds = ids.filter((id) => {
        const iss = issues.find((i) => i.id === id)
        return iss?.status === 'done'
      })
      if (doneIds.length === 0) {
        message.warning('请选择已完成的 Issue')
        return
      }
      if (!onArchiveIssues) {
        message.error('归档接口尚未就绪')
        return
      }
      await onArchiveIssues(doneIds)
      setSelectedIds((prev) => {
        const n = new Set(prev)
        for (const id of doneIds) n.delete(id)
        return n
      })
    },
    [issues, onArchiveIssues],
  )

  const onArchiveSelected = useCallback(() => {
    void archiveIds([...selectedIds])
  }, [archiveIds, selectedIds])

  const onArchiveColumn = useCallback(() => {
    const colIds = (byStatus['done'] || []).map((i) => i.id)
    if (colIds.length === 0) return
    Modal.confirm({
      title: '归档本列',
      content: `确认将完成列中的 ${colIds.length} 条已完成 Issue 归档？归档后默认列表与看板不再显示。`,
      okText: '归档',
      cancelText: '取消',
      onOk: () => archiveIds(colIds),
    })
  }, [archiveIds, byStatus])

  const onDragStart = (e: DragStartEvent) => {
    setActiveId(String(e.active.id))
    setOverId(null)
  }

  const onDragOver = (e: DragOverEvent) => {
    setOverId(e.over ? String(e.over.id) : null)
  }

  const onDragCancel = () => {
    setActiveId(null)
    setOverId(null)
  }

  const onDragEnd = async (e: DragEndEvent) => {
    setActiveId(null)
    setOverId(null)
    const { active: a, over } = e
    if (!over) return
    const issue = issues.find((i) => i.id === a.id)
    if (!issue) return
    const nextStatus = statusFromOver(over.id, issues)
    if (!nextStatus || nextStatus === issue.status) {
      // 同列重排
      if (nextStatus === issue.status && String(over.id) !== issue.id) {
        const col = [...(byStatus[issue.status] || [])]
        const from = col.findIndex((i) => i.id === issue.id)
        const to = col.findIndex((i) => i.id === over.id)
        if (from < 0 || to < 0) return
        col.splice(from, 1)
        col.splice(to, 0, issue)
        const orders = col.map((item, idx) => ({ ...item, board_order: (idx + 1) * 100 }))
        const prev = issues
        onChange(issues.map((i) => orders.find((o) => o.id === i.id) || i))
        try {
          const moved = orders.find((o) => o.id === issue.id)!
          const { issue: saved } = await patchIssue(workspaceId, issue.id, {
            version: issue.version,
            board_order: moved.board_order,
            status: issue.status,
          })
          onChange(
            issues
              .map((i) => (i.id === saved.id ? saved : i))
              .map((i) => {
                const o = orders.find((x) => x.id === i.id)
                return i.id === saved.id
                  ? saved
                  : o && i.id !== saved.id
                    ? { ...i, board_order: o.board_order }
                    : i
              }),
          )
        } catch (err) {
          onChange(prev)
          message.error((err as Error).message || '排序失败')
        }
      }
      return
    }

    // Only allow drop onto currently visible columns
    if (!statusOrder.includes(nextStatus)) return

    const col = byStatus[nextStatus] || []
    const overIdx = col.findIndex((i) => i.id === over.id)
    const board_order =
      overIdx >= 0
        ? (col[overIdx].board_order + (col[overIdx - 1]?.board_order ?? 0)) / 2 ||
          (overIdx + 1) * 100
        : (col.length + 1) * 100

    const optimistic: Issue = { ...issue, status: nextStatus, board_order }
    const prev = issues
    onChange(issues.map((i) => (i.id === issue.id ? optimistic : i)))
    try {
      const { issue: saved } = await patchIssue(workspaceId, issue.id, {
        version: issue.version,
        status: nextStatus,
        board_order,
      })
      onChange(prev.map((i) => (i.id === saved.id ? saved : i)))
    } catch {
      onChange(prev)
      // client 已 Toast（含 version_conflict / conflict）
    }
  }

  return (
    <DndContext
      sensors={sensors}
      collisionDetection={closestCorners}
      autoScroll={{ threshold: { x: 0.12, y: 0.12 }, acceleration: 18, interval: 5 }}
      onDragStart={onDragStart}
      onDragOver={onDragOver}
      onDragCancel={onDragCancel}
      onDragEnd={onDragEnd}
    >
      <div className="acb-board">
        {visibleCols.map((col) => (
          <BoardColumn
            key={col.status}
            status={col.status}
            label={columnLabel(col)}
            issues={byStatus[col.status] || []}
            selectedId={selectedId}
            assigneeNames={assigneeNames}
            onOpen={onOpen}
            dragOverId={overId}
            draggingStatus={active?.status ?? null}
            selectedIds={selectedIds}
            onToggleSelect={onToggleSelect}
            onArchiveSelected={onArchiveSelected}
            onArchiveColumn={onArchiveColumn}
            archiving={archiving}
          />
        ))}
      </div>
      <DragOverlay dropAnimation={dropAnimation}>
        {active ? (
          <BoardCardView
            issue={active}
            overlay
            assigneeName={active.assignee_id ? assigneeNames[active.assignee_id] : null}
          />
        ) : null}
      </DragOverlay>
    </DndContext>
  )
}
