import type { Issue, IssuePriority, IssueStatus } from '../types/issue'

/** Flat issue row for Ant Design Table tree (`children` omitted on leaves). */
export type IssueTreeRow = Issue & { children?: IssueTreeRow[] }

export type IssueTreeFilters = {
  status?: IssueStatus
  priority?: IssuePriority
  assigneeId?: string
  labelId?: string
  q?: string
  /** due_at < today (Asia/Shanghai calendar day); done/canceled excluded */
  overdue?: boolean
}

function todayShanghaiYmd(): string {
  // Box/user zone is Asia/Shanghai; use local calendar day (same as mock overdue).
  const today = new Date()
  const y = today.getFullYear()
  const m = String(today.getMonth() + 1).padStart(2, '0')
  const d = String(today.getDate()).padStart(2, '0')
  return `${y}-${m}-${d}`
}

export function issueMatchesFilters(issue: Issue, f: IssueTreeFilters): boolean {
  if (f.status && issue.status !== f.status) return false
  if (f.priority && issue.priority !== f.priority) return false
  if (f.assigneeId && issue.assignee_id !== f.assigneeId) return false
  if (f.labelId && !(issue.labels || []).some((l) => l.id === f.labelId)) return false
  if (f.q) {
    const s = f.q.toLowerCase()
    if (!issue.title.toLowerCase().includes(s) && !issue.identifier.toLowerCase().includes(s)) {
      return false
    }
  }
  if (f.overdue) {
    const due = issue.due_at
    if (!due || typeof due !== 'string') return false
    const day = due.slice(0, 10)
    if (day >= todayShanghaiYmd()) return false
    if (issue.status === 'done' || issue.status === 'canceled') return false
  }
  return true
}

export function hasActiveFilters(f: IssueTreeFilters): boolean {
  return Boolean(f.status || f.priority || f.assigneeId || f.labelId || f.q || f.overdue)
}

/**
 * Filter: keep matches + ancestor chain (auto-expand those paths).
 * If a match's parent is missing from the loaded set, lift it to root (caller may show 「子项」).
 */
export function selectIssuesWithAncestors(
  all: Issue[],
  filters: IssueTreeFilters,
): { visible: Issue[]; expandIds: string[]; matchIds: Set<string> } {
  const byId = new Map(all.map((i) => [i.id, i]))
  const matchIds = new Set<string>()
  for (const i of all) {
    if (issueMatchesFilters(i, filters)) matchIds.add(i.id)
  }

  if (!hasActiveFilters(filters)) {
    return {
      visible: all.slice(),
      expandIds: [],
      matchIds,
    }
  }

  const keep = new Set<string>()
  const expandIds: string[] = []
  for (const id of matchIds) {
    keep.add(id)
    let pid = byId.get(id)?.parent_id ?? null
    while (pid) {
      if (!byId.has(pid)) break
      if (!keep.has(pid)) {
        keep.add(pid)
        expandIds.push(pid)
      } else if (!expandIds.includes(pid)) {
        expandIds.push(pid)
      }
      pid = byId.get(pid)?.parent_id ?? null
    }
  }

  return {
    visible: all.filter((i) => keep.has(i.id)),
    expandIds,
    matchIds,
  }
}

const sortFn = (a: Issue, b: Issue) =>
  a.board_order - b.board_order ||
  a.number - b.number ||
  a.identifier.localeCompare(b.identifier)

/** Build Ant Design Table tree from flat issues via `parent_id` (missing parent → root). */
export function buildIssueTree(issues: Issue[]): IssueTreeRow[] {
  const map = new Map<string, IssueTreeRow>()
  for (const i of issues) {
    map.set(i.id, { ...i })
  }
  const roots: IssueTreeRow[] = []
  for (const row of map.values()) {
    const pid = row.parent_id
    if (pid && map.has(pid)) {
      const parent = map.get(pid)!
      if (!parent.children) parent.children = []
      parent.children.push(row)
    } else {
      roots.push(row)
    }
  }
  const sortWalk = (nodes: IssueTreeRow[]) => {
    nodes.sort(sortFn)
    for (const n of nodes) {
      if (n.children?.length) sortWalk(n.children)
      else delete n.children
    }
  }
  sortWalk(roots)
  return roots
}

/** True when row sits at root only because its parent_id was not in the visible set. */
export function isOrphanSubitem(row: Issue, visibleIds: Set<string>): boolean {
  return Boolean(row.parent_id && !visibleIds.has(row.parent_id))
}
