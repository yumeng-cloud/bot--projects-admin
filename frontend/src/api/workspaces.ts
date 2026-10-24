import { api } from './client'

export type Workspace = {
  id: string
  name: string
  slug: string
  role: string
  member_id: string
}

export function listMyWorkspaces() {
  return api<{ workspaces: Workspace[] }>('/workspaces')
}

export function getWorkspace(workspaceId: string) {
  return api<{ workspace: Workspace }>(`/workspaces/${workspaceId}`)
}
