import { Fragment } from 'react'
import { Button, Space } from 'antd'
import { useDroppable } from '@dnd-kit/core'
import { SortableContext, verticalListSortingStrategy } from '@dnd-kit/sortable'
import type { Issue, IssueStatus } from '../../types/issue'
import { STATUS_LABELS } from '../../types/issue'
import { BoardCard } from './BoardCard'
import './board.css'

type Props = {
  status: IssueStatus
  /** Display name override (from column prefs) */
  label?: string
  issues: Issue[]
  selectedId?: string | null
  assigneeNames?: Record<string, string>
  onOpen: (issue: Issue) => void
  /** Cross-column drop target id (card id or `col-${status}`) while dragging */
  dragOverId?: string | null
  /** Status of the card being dragged — used to show insertion line only across columns */
  draggingStatus?: IssueStatus | null
  /** C1：完成列多选 id 集合 */
  selectedIds?: Set<string>
  onToggleSelect?: (issue: Issue, next: boolean) => void
  /** 归档当前勾选（仅 done） */
  onArchiveSelected?: () => void
  /** 归档本列全部已完成 */
  onArchiveColumn?: () => void
  archiving?: boolean
}

export function BoardColumn({
  status,
  label,
  issues,
  selectedId,
  assigneeNames,
  onOpen,
  dragOverId,
  draggingStatus,
  selectedIds,
  onToggleSelect,
  onArchiveSelected,
  onArchiveColumn,
  archiving,
}: Props) {
  const { setNodeRef, isOver } = useDroppable({ id: `col-${status}`, data: { status } })
  const crossColumn =
    draggingStatus != null && draggingStatus !== status && dragOverId != null
  const title = (label && label.trim()) || STATUS_LABELS[status]
  const isDoneCol = status === 'done'
  const selectedCount = isDoneCol
    ? issues.filter((i) => selectedIds?.has(i.id)).length
    : 0

  return (
    <div ref={setNodeRef} className={`acb-col${isOver ? ' is-over' : ''}`}>
      <div className="acb-col-title">
        <span className={`acb-col-dot status-${status}`} aria-hidden />
        <span className="acb-col-name">{title}</span>
        <span className="acb-col-count">{issues.length}</span>
      </div>
      {isDoneCol && issues.length > 0 ? (
        <div className="acb-col-archive-bar">
          <Space size={4} wrap>
            <Button
              type="link"
              size="small"
              className="acb-col-archive-btn"
              disabled={selectedCount === 0 || archiving}
              loading={archiving && selectedCount > 0}
              onClick={() => onArchiveSelected?.()}
            >
              归档{selectedCount > 0 ? `（${selectedCount}）` : ''}
            </Button>
            <Button
              type="link"
              size="small"
              className="acb-col-archive-btn"
              disabled={archiving || issues.length === 0}
              onClick={() => onArchiveColumn?.()}
            >
              归档本列
            </Button>
          </Space>
        </div>
      ) : null}
      <div className="acb-col-body">
        <SortableContext items={issues.map((i) => i.id)} strategy={verticalListSortingStrategy}>
          <div className="acb-col-cards">
            {issues.map((issue) => (
              <Fragment key={issue.id}>
                {crossColumn && dragOverId === issue.id && (
                  <div className="acb-drop-indicator" aria-hidden />
                )}
                <BoardCard
                  issue={issue}
                  selected={issue.id === selectedId}
                  assigneeName={
                    issue.assignee_id ? assigneeNames?.[issue.assignee_id] ?? null : null
                  }
                  onOpen={onOpen}
                  selectable={isDoneCol}
                  checked={isDoneCol ? !!selectedIds?.has(issue.id) : false}
                  onToggleSelect={onToggleSelect}
                />
              </Fragment>
            ))}
            {crossColumn && dragOverId === `col-${status}` && issues.length > 0 && (
              <div className="acb-drop-indicator" aria-hidden />
            )}
          </div>
        </SortableContext>
        {issues.length === 0 && (
          <div className={`acb-col-empty${isOver ? ' is-over' : ''}`}>
            <span className="acb-col-empty-cta">拖入到此列</span>
          </div>
        )}
      </div>
    </div>
  )
}
