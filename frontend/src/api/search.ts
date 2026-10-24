import { api } from './client'
import type { Issue } from '../types/issue'

export type SearchProjectHit = {
  id: string
  workspace_id: string
  name: string
  key?: string
  description?: string
}

export type SearchWorkspaceOpts = {
  /** T6I pool filter — same as ListIssues ?team_id= */
  team_id?: string
  /** Explicit unassigned pool (team_id IS NULL) */
  unassigned?: boolean
}

export function searchWorkspace(workspaceId: string, q: string, opts?: SearchWorkspaceOpts) {
  const qs = new URLSearchParams({ q })
  if (opts?.team_id) qs.set('team_id', opts.team_id)
  if (opts?.unassigned) qs.set('unassigned', '1')
  return api<{ q: string; issues: Issue[]; projects: SearchProjectHit[] }>(
    `/workspaces/${workspaceId}/search?${qs}`,
  )
}
