import { api } from './client'
import { mockIssues, mockLabels, mockMembers, preferMock } from './mockData'
import type { Issue, IssueStatus, Label, MemberRef } from '../types/issue'

export type IssueListQuery = {
  status?: IssueStatus | IssueStatus[]
  assignee_id?: string
  label_id?: string
  q?: string
  cycle_id?: string
  /** 仅叶子（无子节点） */
  leaf_only?: boolean | '1' | 1
  /** 仅被未完成 blocker 挡住的 */
  blocked?: boolean | '1' | 1
  /** due_at < today（日历日 Asia/Shanghai） */
  overdue?: boolean | '1' | 1
}

export type IssuePatch = Partial<
  Pick<
    Issue,
    | 'title'
    | 'description'
    | 'status'
    | 'priority'
    | 'estimate'
    | 'assignee_id'
    | 'cycle_id'
    | 'board_order'
    | 'parent_id'
    | 'start_at'
    | 'due_at'
  >
> & { version: number; label_ids?: string[] }

export type IssueCreate = {
  project_id: string
  title: string
  status?: IssueStatus
  priority?: Issue['priority']
  description?: string
  estimate?: number | null
  assignee_id?: string | null
  cycle_id?: string | null
  parent_id?: string | null
  /** YYYY-MM-DD 日历日 */
  start_at?: string | null
  /** YYYY-MM-DD 日历日；若与 start_at 同传须 due_at >= start_at */
  due_at?: string | null
  label_ids?: string[]
}

function filterMock(list: Issue[], q: IssueListQuery): Issue[] {
  return list.filter((i) => {
    if (q.status) {
      const statuses = Array.isArray(q.status) ? q.status : [q.status]
      if (!statuses.includes(i.status)) return false
    }
    if (q.assignee_id && i.assignee_id !== q.assignee_id) return false
    if (q.cycle_id && i.cycle_id !== q.cycle_id) return false
    if (q.label_id && !(i.labels || []).some((l) => l.id === q.label_id)) return false
    if (q.q) {
      const s = q.q.toLowerCase()
      if (!i.title.toLowerCase().includes(s) && !i.identifier.toLowerCase().includes(s)) return false
    }
    if (q.leaf_only === true || q.leaf_only === '1' || q.leaf_only === 1) {
      const hasChild = list.some((c) => c.parent_id === i.id)
      if (hasChild) return false
    }
    if (q.overdue === true || q.overdue === '1' || q.overdue === 1) {
      const due = i.due_at
      if (!due || typeof due !== 'string') return false
      const day = due.slice(0, 10)
      const today = new Date()
      const y = today.getFullYear()
      const m = String(today.getMonth() + 1).padStart(2, '0')
      const d = String(today.getDate()).padStart(2, '0')
      const todayStr = `${y}-${m}-${d}`
      if (day >= todayStr) return false
      if (i.status === 'done' || i.status === 'canceled') return false
    }
    // blocked: mock 无 relation 图，忽略（真 API 负责）
    return true
  })
}

function qs(query: IssueListQuery): string {
  const p = new URLSearchParams()
  if (query.status) {
    const v = Array.isArray(query.status) ? query.status.join(',') : query.status
    p.set('status', v)
  }
  if (query.assignee_id) p.set('assignee_id', query.assignee_id)
  if (query.label_id) p.set('label_id', query.label_id)
  if (query.cycle_id) p.set('cycle_id', query.cycle_id)
  if (query.q) p.set('q', query.q)
  if (query.leaf_only === true || query.leaf_only === '1' || query.leaf_only === 1) p.set('leaf_only', '1')
  if (query.blocked === true || query.blocked === '1' || query.blocked === 1) p.set('blocked', '1')
  if (query.overdue === true || query.overdue === '1' || query.overdue === 1) p.set('overdue', '1')
  const s = p.toString()
  return s ? `?${s}` : ''
}

function applyMockPatch(id: string, body: IssuePatch): Issue {
  const idx = mockIssues.findIndex((i) => i.id === id)
  if (idx < 0) throw Object.assign(new Error('Issue 不存在'), { code: 'not_found', status: 404 })
  const cur = mockIssues[idx]
  if (body.version !== cur.version) {
    throw Object.assign(new Error('版本冲突，请刷新后重试'), {
      code: 'version_conflict',
      status: 409,
    })
  }
  const { version: _v, label_ids, ...rest } = body
  const next: Issue = {
    ...cur,
    ...rest,
    version: cur.version + 1,
    updated_at: new Date().toISOString(),
  }
  if (label_ids) {
    next.labels = mockLabels.filter((l) => label_ids.includes(l.id))
  }
  mockIssues[idx] = next
  return next
}

export async function listIssues(
  workspaceId: string,
  query: IssueListQuery & { project_id?: string } = {},
): Promise<{ issues: Issue[]; mock: boolean }> {
  if (preferMock()) {
    return { issues: filterMock([...mockIssues], query), mock: true }
  }
  const p = new URLSearchParams()
  if (query.project_id) p.set('project_id', query.project_id)
  const baseQs = qs(query)
  const extra = p.toString()
  const join = baseQs ? (extra ? `${baseQs}&${extra}` : baseQs) : extra ? `?${extra}` : ''
  const data = await api<{ issues: Issue[] }>(`/workspaces/${workspaceId}/issues${join}`)
  return { issues: data.issues || [], mock: false }
}

export async function getIssue(
  workspaceId: string,
  id: string,
): Promise<{ issue: Issue; mock: boolean }> {
  if (preferMock()) {
    const issue = mockIssues.find((i) => i.id === id)
    if (!issue) throw Object.assign(new Error('Issue 不存在'), { code: 'not_found', status: 404 })
    return { issue: { ...issue }, mock: true }
  }
  const data = await api<{ issue: Issue }>(`/workspaces/${workspaceId}/issues/${id}`)
  return { issue: data.issue, mock: false }
}

export async function patchIssue(
  workspaceId: string,
  id: string,
  body: IssuePatch,
): Promise<{ issue: Issue; mock: boolean }> {
  if (preferMock()) {
    return { issue: applyMockPatch(id, body), mock: true }
  }
  const data = await api<{ issue: Issue }>(`/workspaces/${workspaceId}/issues/${id}`, {
    method: 'PATCH',
    body: JSON.stringify(body),
  })
  return { issue: data.issue, mock: false }
}

export async function createIssue(
  workspaceId: string,
  body: IssueCreate,
): Promise<{ issue: Issue; mock: boolean }> {
  if (preferMock()) {
    const n = mockIssues.reduce((m, i) => Math.max(m, i.number), 0) + 1
    const now = new Date().toISOString()
    const issue: Issue = {
      id: `iss-${n}`,
      project_id: body.project_id,
      workspace_id: workspaceId,
      number: n,
      identifier: `DEMO-${n}`,
      title: body.title,
      description: body.description ?? '',
      status: body.status ?? 'todo',
      priority: body.priority ?? 'none',
      estimate: body.estimate ?? null,
      assignee_id: body.assignee_id ?? null,
      reporter_id: mockMembers[0]?.id ?? null,
      cycle_id: body.cycle_id ?? null,
      parent_id: null,
      board_order: (mockIssues.filter((i) => i.status === (body.status ?? 'todo')).length + 1) * 100,
      version: 1,
      labels: body.label_ids
        ? mockLabels.filter((l) => body.label_ids!.includes(l.id))
        : [],
      created_at: now,
      updated_at: now,
    }
    mockIssues.unshift(issue)
    return { issue, mock: true }
  }
  const data = await api<{ issue: Issue }>(`/workspaces/${workspaceId}/issues`, {
    method: 'POST',
    body: JSON.stringify(body),
  })
  return { issue: data.issue, mock: false }
}


export async function listProjectLabels(
  workspaceId: string,
): Promise<{ labels: Label[]; mock: boolean }> {
  if (preferMock()) return { labels: [...mockLabels], mock: true }
  const data = await api<{ labels: Label[] }>(`/workspaces/${workspaceId}/labels`)
  return { labels: data.labels || [], mock: false }
}

export async function listProjectMembers(
  workspaceId: string,
): Promise<{ members: MemberRef[]; mock: boolean }> {
  if (preferMock()) return { members: [...mockMembers], mock: true }
  const data = await api<{ members: MemberRef[] }>(`/workspaces/${workspaceId}/members`)
  return { members: data.members || [], mock: false }
}
