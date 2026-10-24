import { useCallback, useEffect, useState } from 'react'
import { listTeams, type Team } from '../api/teams'
import {
  readStoredTeamId,
  resolveTeamId,
  writeStoredTeamId,
} from '../utils/currentTeam'

export type UseCurrentTeamResult = {
  teams: Team[]
  teamId: string | null
  /** Persist + update; call before clearing issue pools on switch. */
  setTeamId: (id: string | null) => void
  loading: boolean
}

/**
 * Minimal workspace current-team context (T6I I4).
 * localStorage key: acb.team.<wsId> — same as Bot roster page.
 */
export function useCurrentTeam(workspaceId: string): UseCurrentTeamResult {
  const [teams, setTeams] = useState<Team[]>([])
  const [teamId, setTeamIdState] = useState<string | null>(() =>
    readStoredTeamId(workspaceId),
  )
  const [loading, setLoading] = useState(true)

  const setTeamId = useCallback(
    (id: string | null) => {
      setTeamIdState(id)
      writeStoredTeamId(workspaceId, id)
    },
    [workspaceId],
  )

  useEffect(() => {
    let cancelled = false
    setLoading(true)
    // Re-hydrate from storage when ws changes
    const saved = readStoredTeamId(workspaceId)
    setTeamIdState(saved)
    ;(async () => {
      try {
        const { teams: list } = await listTeams(workspaceId)
        if (cancelled) return
        const teamList = list || []
        setTeams(teamList)
        const next = resolveTeamId(teamList, saved)
        setTeamIdState(next)
        if (next) writeStoredTeamId(workspaceId, next)
      } catch {
        if (!cancelled) {
          setTeams([])
          // keep stored id; list/board may still pass it when API recovers
        }
      } finally {
        if (!cancelled) setLoading(false)
      }
    })()
    return () => {
      cancelled = true
    }
  }, [workspaceId])

  return { teams, teamId, setTeamId, loading }
}
