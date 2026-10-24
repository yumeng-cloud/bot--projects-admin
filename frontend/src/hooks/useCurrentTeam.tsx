import {
  createContext,
  useCallback,
  useContext,
  useEffect,
  useMemo,
  useState,
  type ReactNode,
} from 'react'
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

const CurrentTeamContext = createContext<UseCurrentTeamResult | null>(null)

/**
 * Workspace-scoped current team (T6I I4 / TEAM_SWITCH).
 * One provider under AppLayout so header + list/board/cycles/… share state.
 * localStorage key: acb.team.<wsId>
 */
export function CurrentTeamProvider({
  workspaceId,
  children,
}: {
  workspaceId: string
  children: ReactNode
}) {
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

  const value = useMemo(
    () => ({ teams, teamId, setTeamId, loading }),
    [teams, teamId, setTeamId, loading],
  )

  return (
    <CurrentTeamContext.Provider value={value}>{children}</CurrentTeamContext.Provider>
  )
}

/**
 * Read current team from nearest CurrentTeamProvider (AppLayout).
 * `workspaceId` kept for call-site clarity; provider is the source of truth.
 */
export function useCurrentTeam(_workspaceId: string): UseCurrentTeamResult {
  const ctx = useContext(CurrentTeamContext)
  if (!ctx) {
    throw new Error('useCurrentTeam must be used within CurrentTeamProvider')
  }
  return ctx
}
