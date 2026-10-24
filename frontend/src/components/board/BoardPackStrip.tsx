import type { Cycle, MemberRef } from '../../types/issue'
import { memberLabel } from '../../types/issue'
import './board.css'

export type PackCollaborator = {
  id: string
  name: string
}

export type BoardPackStripProps = {
  done: number
  total: number
  activeCycle?: Cycle | null
  collaborators: PackCollaborator[]
  blockerCount: number
  onCycleClick?: () => void
  onBlockersClick?: () => void
}

function initialOf(name: string): string {
  const t = name.trim()
  if (!t) return '?'
  return t.slice(0, 1).toUpperCase()
}

/** §1.1 global pack strip — progress / cycle / collaborators / blockers on board header. */
export function BoardPackStrip({
  done,
  total,
  activeCycle,
  collaborators,
  blockerCount,
  onCycleClick,
  onBlockersClick,
}: BoardPackStripProps) {
  const pct = total > 0 ? Math.round((done / total) * 100) : 0
  const shown = collaborators.slice(0, 5)
  const extra = Math.max(0, collaborators.length - 5)
  const barDone = total > 0 && done >= total

  return (
    <div className="acb-pack-strip" role="group" aria-label="项目包装">
      <div className="acb-pack-progress" title={`${done}/${total} · ${pct}%`}>
        <div className="acb-pack-progress-meta">
          <span className="acb-pack-progress-label">进度</span>
          <span className="acb-pack-progress-frac">
            {done}/{total}
          </span>
        </div>
        <div
          className={`acb-pack-progress-track${barDone ? ' is-complete' : ''}`}
          role="progressbar"
          aria-valuenow={pct}
          aria-valuemin={0}
          aria-valuemax={100}
        >
          <div
            className="acb-pack-progress-fill"
            style={{ width: `${Math.min(100, Math.max(0, pct))}%` }}
          />
        </div>
      </div>

      {activeCycle ? (
        <button
          type="button"
          className="acb-pack-chip acb-pack-cycle acb-pack-secondary"
          onClick={onCycleClick}
          title={`当前 Cycle：${activeCycle.name}`}
        >
          <span className="acb-pack-chip-dot" aria-hidden />
          <span className="acb-pack-chip-text">{activeCycle.name}</span>
        </button>
      ) : null}

      {collaborators.length > 0 ? (
        <div
          className="acb-pack-avatars acb-pack-secondary"
          title={`${collaborators.length} 位协作人`}
          aria-label="协作人"
        >
          {shown.map((c) => (
            <span key={c.id} className="acb-pack-avatar" title={c.name}>
              {initialOf(c.name)}
            </span>
          ))}
          {extra > 0 ? <span className="acb-pack-avatar acb-pack-avatar-more">+{extra}</span> : null}
        </div>
      ) : null}

      <button
        type="button"
        className={`acb-pack-chip acb-pack-blockers${blockerCount > 0 ? ' has-blockers' : ''}`}
        onClick={onBlockersClick}
        disabled={blockerCount <= 0}
        title={blockerCount > 0 ? `${blockerCount} 个卡点（阻塞或逾期）` : '无卡点'}
      >
        <span className="acb-pack-chip-text">卡点 {blockerCount}</span>
      </button>
    </div>
  )
}

/** Build collaborator list from issue assignees + member map. */
export function collaboratorsFromIssues(
  assigneeIds: (string | null | undefined)[],
  members: MemberRef[],
): PackCollaborator[] {
  const byId = new Map(members.map((m) => [m.id, m]))
  const seen = new Set<string>()
  const out: PackCollaborator[] = []
  for (const id of assigneeIds) {
    if (!id || seen.has(id)) continue
    seen.add(id)
    const m = byId.get(id)
    out.push({ id, name: m ? memberLabel(m) : '成员' })
  }
  return out
}
