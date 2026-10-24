import { useSortable } from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'
import type { DraggableAttributes } from '@dnd-kit/core'
import type { SyntheticListenerMap } from '@dnd-kit/core/dist/hooks/utilities'
import type { CSSProperties } from 'react'
import type { Issue, IssuePriority } from '../../types/issue'
import { PRIORITY_LABELS } from '../../types/issue'
import './board.css'

const PRIO_CLASS: Record<IssuePriority, string> = {
  urgent: 'prio-urgent',
  high: 'prio-high',
  medium: 'prio-medium',
  low: 'prio-low',
  none: 'prio-none',
}

function initials(name: string): string {
  const t = name.trim()
  if (!t) return '?'
  return t.slice(0, 1)
}

type ViewProps = {
  issue: Issue
  assigneeName?: string | null
  selected?: boolean
  dragging?: boolean
  overlay?: boolean
  onOpen?: (issue: Issue) => void
  style?: CSSProperties
  innerRef?: (node: HTMLElement | null) => void
  attributes?: DraggableAttributes
  listeners?: SyntheticListenerMap
}

/** Pure card face — shared by sortable row and DragOverlay */
export function BoardCardView({
  issue,
  assigneeName,
  selected,
  dragging,
  overlay,
  onOpen,
  style,
  innerRef,
  attributes,
  listeners,
}: ViewProps) {
  const cls = [
    'acb-card',
    selected ? 'is-selected' : '',
    dragging ? 'is-dragging' : '',
    overlay ? 'is-overlay' : '',
  ]
    .filter(Boolean)
    .join(' ')

  const prio = issue.priority || 'none'
  const name = assigneeName || null
  const labels = (issue.labels || []).slice(0, 2)
  const extraLabelCount = Math.max(0, (issue.labels || []).length - labels.length)

  return (
    <div
      ref={innerRef}
      style={style}
      className={cls}
      {...attributes}
      {...listeners}
      onClick={() => onOpen?.(issue)}
    >
      <span className={`acb-prio-bar ${PRIO_CLASS[prio]}`} title={PRIORITY_LABELS[prio]} />
      <div className="acb-card-top">
        <div className="acb-card-id">{issue.identifier}</div>
        {name ? (
          <span className="acb-avatar" title={name}>
            {initials(name)}
          </span>
        ) : (
          <span className="acb-avatar is-empty" title="未指派">
            ·
          </span>
        )}
      </div>
      <div className="acb-card-title">{issue.title}</div>
      {(labels.length > 0 || issue.estimate != null) && (
        <div className="acb-card-meta">
          {labels.map((l) => (
            <span
              key={l.id}
              className="acb-chip acb-label-chip"
              style={{
                borderColor: `color-mix(in srgb, ${l.color} 35%, var(--acb-border))`,
                color: l.color,
                background: `color-mix(in srgb, ${l.color} 10%, var(--acb-bg-container))`,
              }}
            >
              {l.name}
            </span>
          ))}
          {extraLabelCount > 0 && (
            <span className="acb-chip acb-label-more" title={`另有 ${extraLabelCount} 个标签`}>
              +{extraLabelCount}
            </span>
          )}
          {issue.estimate != null && (
            <span className="acb-chip estimate" title="估时">
              {issue.estimate}pt
            </span>
          )}
        </div>
      )}
    </div>
  )
}

type Props = {
  issue: Issue
  assigneeName?: string | null
  selected?: boolean
  onOpen: (issue: Issue) => void
}

export function BoardCard({ issue, assigneeName, selected, onOpen }: Props) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({
    id: issue.id,
    data: { status: issue.status },
  })

  // Translate-only avoids scale jank during sortable reorder; freeze transition while dragging.
  const style: CSSProperties = {
    transform: CSS.Translate.toString(transform),
    transition: isDragging ? undefined : transition,
  }

  return (
    <BoardCardView
      issue={issue}
      assigneeName={assigneeName}
      selected={selected}
      dragging={isDragging}
      onOpen={onOpen}
      style={style}
      innerRef={setNodeRef}
      attributes={attributes}
      listeners={listeners}
    />
  )
}
