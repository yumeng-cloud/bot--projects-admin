import { useCallback, useEffect, useMemo, useState, type Dispatch, type SetStateAction } from 'react'
import { listIssues, listProjectMembers } from '../api/issues'
import { listCycles } from '../api/cycles'
import { listRelations } from '../api/relations'
import {
  collaboratorsFromIssues,
  type PackCollaborator,
} from '../components/board/BoardPackStrip'
import type { Cycle, Issue, MemberRef } from '../types/issue'
import { isIssueArchived } from '../types/issue'
import { dayjs, TZ } from '../utils/time'

export type UseBoardPackStatsArgs = {
  workspaceId: string
  projectId: string
  /** Current team pool (T6I I4); omit/empty → no team filter */
  teamId?: string | null
  /** false while current-team context is still resolving */
  enabled?: boolean
}

export type UseBoardPackStatsResult = {
  loading: boolean
  issues: Issue[]
  /** Local mutations (create / drag / drawer save) — keeps pack stats in sync. */
  setIssues: Dispatch<SetStateAction<Issue[]>>
  done: number
  total: number
  activeCycle: Cycle | null
  collaborators: PackCollaborator[]
  blockerCount: number
  reload: () => Promise<void>
}

function isTerminal(status: Issue['status']): boolean {
  return status === 'done' || status === 'canceled'
}

/** Normalize to Asia/Shanghai calendar day YYYY-MM-DD. */
function calendarDay(isoOrYmd: string): string {
  if (/^\d{4}-\d{2}-\d{2}$/.test(isoOrYmd)) return isoOrYmd
  const d = dayjs.utc(isoOrYmd).tz(TZ)
  if (!d.isValid()) return ''
  return d.format('YYYY-MM-DD')
}

function todayYmd(): string {
  return dayjs().tz(TZ).format('YYYY-MM-DD')
}

/**
 * Project pack strip stats (PRODUCT_SUITE_P0 §1.1).
 * Self-fetches issues + members + active cycle; aggregates done/total,
 * collaborators, and blockerCount (unresolved blocked_by ∪ overdue).
 */
export function useBoardPackStats({
  workspaceId,
  projectId,
  teamId,
  enabled = true,
}: UseBoardPackStatsArgs): UseBoardPackStatsResult {
  const [loading, setLoading] = useState(true)
  const [issues, setIssues] = useState<Issue[]>([])
  const [members, setMembers] = useState<MemberRef[]>([])
  const [activeCycle, setActiveCycle] = useState<Cycle | null>(null)
  const [blockedIds, setBlockedIds] = useState<Set<string>>(() => new Set())

  const reload = useCallback(async () => {
    if (!enabled) {
      setIssues([])
      setLoading(true)
      return
    }
    setLoading(true)
    setIssues([]) // I4: clear pool before fetch — no cross-team residue
    try {
      const [{ issues: list }, mem, cyclesRes] = await Promise.all([
        listIssues(workspaceId, {
          project_id: projectId,
          team_id: teamId || undefined,
        }),
        listProjectMembers(workspaceId).catch(() => ({ members: [] as MemberRef[] })),
        listCycles(workspaceId, projectId).catch(() => ({ cycles: [] as Cycle[] })),
      ])
      // C2/C3：默认 list 不传 include_archived；若偶发带回归档则本地剔掉，避免虚增统计
      setIssues(list.filter((i) => !isIssueArchived(i)))
      setMembers(mem.members || [])
      const active =
        (cyclesRes.cycles || []).find((c) => c.status === 'active') || null
      setActiveCycle(active)

      // Soft-load blocked_by for open issues (capped); overdue counted separately from due_at
      const openIssues = list.filter((i) => !isTerminal(i.status)).slice(0, 40)
      if (openIssues.length === 0) {
        setBlockedIds(new Set())
      } else {
        const results = await Promise.allSettled(
          openIssues.map((i) => listRelations(workspaceId, i.id)),
        )
        const blocked = new Set<string>()
        const byId = new Map(list.map((i) => [i.id, i]))
        results.forEach((r, idx) => {
          if (r.status !== 'fulfilled') return
          const issue = openIssues[idx]
          const rels = r.value.relations || []
          const hasUnresolvedBlocker = rels.some((rel) => {
            if (rel.relation_type !== 'blocked_by') return false
            // blocked_by: from_issue is blocked by to_issue
            const blockerId =
              rel.from_issue_id === issue.id ? rel.to_issue_id : rel.from_issue_id
            const blocker = byId.get(blockerId)
            if (!blocker) return true // unknown → count as unresolved
            return !isTerminal(blocker.status)
          })
          if (hasUnresolvedBlocker) blocked.add(issue.id)
        })
        setBlockedIds(blocked)
      }
    } catch {
      setIssues([])
      setBlockedIds(new Set())
      // 4xx toast from client; keep empty pool (no white screen)
    } finally {
      setLoading(false)
    }
  }, [workspaceId, projectId, teamId, enabled])

  useEffect(() => {
    void reload()
  }, [reload])

  const { done, total, blockerCount, collaborators } = useMemo(() => {
    // C3：Pack 统计排除已归档（与 BE 一致；不本地虚增）
    const visible = issues.filter((i) => !isIssueArchived(i))
    const countable = visible.filter((i) => i.status !== 'canceled')
    const doneCount = countable.filter((i) => i.status === 'done').length
    const totalCount = countable.length
    const today = todayYmd()
    const overdueIds = new Set<string>()
    for (const i of visible) {
      if (isTerminal(i.status)) continue
      const due = i.due_at
      if (!due) continue
      const day = calendarDay(due)
      if (day && day < today) overdueIds.add(i.id)
    }
    const blockerSet = new Set<string>([...blockedIds, ...overdueIds])
    const collabs = collaboratorsFromIssues(
      visible.map((i) => i.assignee_id),
      members,
    )
    return {
      done: doneCount,
      total: totalCount,
      blockerCount: blockerSet.size,
      collaborators: collabs,
    }
  }, [issues, members, blockedIds])

  return {
    loading,
    issues,
    setIssues,
    done,
    total,
    activeCycle,
    collaborators,
    blockerCount,
    reload,
  }
}
