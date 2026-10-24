import { api } from './client'

export type Project = {
  id: string
  workspace_id: string
  name: string
  key_prefix: string
  description: string
  archived_at: string | null
  created_at: string
  updated_at: string
}

export function listProjects(workspaceId: string, includeArchived = false) {
  const q = includeArchived ? '?include_archived=1' : ''
  return api<{ projects: Project[] }>(`/workspaces/${workspaceId}/projects${q}`)
}

export function createProject(
  workspaceId: string,
  body: { name: string; key_prefix: string; description?: string },
) {
  return api<{ project: Project }>(`/workspaces/${workspaceId}/projects`, {
    method: 'POST',
    body: JSON.stringify(body),
  })
}

export function getProject(workspaceId: string, projectId: string) {
  return api<{ project: Project }>(`/workspaces/${workspaceId}/projects/${projectId}`)
}

export function patchProject(
  workspaceId: string,
  projectId: string,
  body: { name?: string; description?: string; archive?: boolean },
) {
  return api<{ project: Project }>(`/workspaces/${workspaceId}/projects/${projectId}`, {
    method: 'PATCH',
    body: JSON.stringify(body),
  })
}
