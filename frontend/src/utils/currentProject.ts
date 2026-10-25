/** Workspace-scoped current project — X-R1 acb.project.<wsId>（替换全局 ai-collab-board:projectId） */

export function projectStorageKey(workspaceId: string): string {
  return `acb.project.${workspaceId}`
}

export function readStoredProjectId(workspaceId: string): string | null {
  if (!workspaceId) return null
  try {
    return localStorage.getItem(projectStorageKey(workspaceId))
  } catch {
    return null
  }
}

export function writeStoredProjectId(workspaceId: string, projectId: string | null): void {
  if (!workspaceId) return
  try {
    if (projectId) localStorage.setItem(projectStorageKey(workspaceId), projectId)
    else localStorage.removeItem(projectStorageKey(workspaceId))
  } catch {
    /* ignore quota / private mode */
  }
}

/** Prefer saved id if still in list; else first project. */
export function resolveProjectId(
  projects: { id: string }[],
  preferred: string | null | undefined,
): string | null {
  if (!projects.length) return null
  if (preferred && projects.some((p) => p.id === preferred)) return preferred
  return projects[0]?.id ?? null
}
