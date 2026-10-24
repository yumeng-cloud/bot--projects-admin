import { Fragment } from 'react'
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
}: Props) {
  const { setNodeRef, isOver } = useDroppable({ id: `col-${status}`, data: { status } })
  const crossColumn =
    draggingStatus != null && draggingStatus !== status && dragOverId != null
  const title = (label && label.trim()) || STATUS_LABELS[status]

  return (
    <div ref={setNodeRef} className={`acb-col${isOver ? ' is-over' : ''}`}>
      <div className="acb-col-title">
        <span className={`acb-col-dot status-${status}`} aria-hidden />
        <span className="acb-col-name">{title}</span>
        <span className="acb-col-count">{issues.length}</span>
      </div>
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
