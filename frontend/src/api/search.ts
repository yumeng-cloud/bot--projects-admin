import { api } from './client'
import type { Issue } from '../types/issue'

export type SearchProjectHit = {
  id: string
  workspace_id: string
  name: string
  key?: string
  description?: string
}

export function searchWorkspace(workspaceId: string, q: string) {
  const qs = new URLSearchParams({ q })
  return api<{ q: string; issues: Issue[]; projects: SearchProjectHit[] }>(
    `/workspaces/${workspaceId}/search?${qs}`,
  )
}
