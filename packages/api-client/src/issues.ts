import { api } from './client'
import type { Issue, IssueStatus } from './boardTypes'

export type IssueListQuery = {
  team_id?: string
  project_id?: string
  status?: IssueStatus | string
  q?: string
}

function qs(query: IssueListQuery = {}): string {
  const p = new URLSearchParams()
  if (query.team_id) p.set('team_id', query.team_id)
  if (query.project_id) p.set('project_id', query.project_id)
  if (query.status) p.set('status', query.status)
  if (query.q) p.set('q', query.q)
  const s = p.toString()
  return s ? `?${s}` : ''
}

export function listIssues(workspaceId: string, query: IssueListQuery = {}) {
  return api<{ issues: Issue[] }>(`/workspaces/${workspaceId}/issues${qs(query)}`)
}

export function getIssue(workspaceId: string, issueId: string) {
  return api<{ issue: Issue }>(`/workspaces/${workspaceId}/issues/${issueId}`)
}

export function patchIssue(
  workspaceId: string,
  issueId: string,
  body: { version: number; status?: IssueStatus; title?: string; description?: string },
) {
  return api<{ issue: Issue }>(`/workspaces/${workspaceId}/issues/${issueId}`, {
    method: 'PATCH',
    body: JSON.stringify(body),
  })
}
