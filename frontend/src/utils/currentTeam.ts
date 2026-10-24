/** Workspace-scoped current team — shared by list/board + Bot roster (I4). */

export function teamStorageKey(workspaceId: string): string {
  return `acb.team.${workspaceId}`
}

export function readStoredTeamId(workspaceId: string): string | null {
  if (!workspaceId) return null
  try {
    return localStorage.getItem(teamStorageKey(workspaceId))
  } catch {
    return null
  }
}

export function writeStoredTeamId(workspaceId: string, teamId: string | null): void {
  if (!workspaceId) return
  try {
    if (teamId) localStorage.setItem(teamStorageKey(workspaceId), teamId)
    else localStorage.removeItem(teamStorageKey(workspaceId))
  } catch {
    /* ignore quota / private mode */
  }
}

/** Prefer saved id if still in list; else first team. */
export function resolveTeamId(
  teams: { id: string }[],
  preferred: string | null | undefined,
): string | null {
  if (!teams.length) return null
  if (preferred && teams.some((t) => t.id === preferred)) return preferred
  return teams[0]?.id ?? null
}
