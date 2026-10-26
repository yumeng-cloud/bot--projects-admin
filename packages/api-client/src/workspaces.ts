import { api } from './client'
import type { Workspace } from './boardTypes'

export function listMyWorkspaces() {
  return api<{ workspaces: Workspace[] }>('/workspaces')
}

export function getWorkspace(workspaceId: string) {
  return api<{
    workspace: { id: string; name: string; slug: string }
    role?: string
    member_id?: string
  }>(`/workspaces/${workspaceId}`)
}
