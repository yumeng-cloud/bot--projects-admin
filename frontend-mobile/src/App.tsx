import { useCallback, useEffect, useMemo, useState } from 'react'
import { Navigate, Route, Routes, useNavigate } from 'react-router-dom'
import {
  getToken,
  me,
  getSessionContext,
  emptySessionContext,
  putSessionContext,
  type User,
  type SessionContext,
  type SessionContextPut,
} from '@acb/api-client'
import { SessionProvider } from './context/SessionCtx'
import { pathFromSession } from './lib/sessionNav'
import AppShell from './components/AppShell'
import LoginPage from './pages/LoginPage'
import MePage from './pages/MePage'
import ResumePage from './pages/ResumePage'
import CrewWorkspacePage from './pages/CrewWorkspacePage'
import CrewTeamsPage from './pages/CrewTeamsPage'
import CrewSeatsPage from './pages/CrewSeatsPage'
import IssueListPage from './pages/IssueListPage'
import IssueDetailPage from './pages/IssueDetailPage'
import WikiPage from './pages/WikiPage'

/**
 * MV2-R0/R1/R2：登录 · hydrate · 底栏 Tab（队席/Issue/Wiki/我的）+ 列表→详情栈
 */
export default function App() {
  const navigate = useNavigate()
  const [ready, setReady] = useState(false)
  const [user, setUser] = useState<User | null>(null)
  const [session, setSession] = useState<SessionContext>(emptySessionContext())
  const [hydrateNote, setHydrateNote] = useState<string | null>(null)

  const hydrateSession = useCallback(async (): Promise<SessionContext> => {
    try {
      const ctx = await getSessionContext()
      setSession(ctx)
      setHydrateNote(null)
      return ctx
    } catch {
      const empty = emptySessionContext()
      setSession(empty)
      setHydrateNote('现场同步暂不可用，已用安静默认')
      return empty
    }
  }, [])

  const refreshMe = useCallback(
    async (opts?: { resumeNav?: boolean }) => {
      if (!getToken()) {
        setUser(null)
        setSession(emptySessionContext())
        setReady(true)
        return
      }
      try {
        const { user: u } = await me()
        setUser(u)
        const ctx = await hydrateSession()
        if (opts?.resumeNav) {
          navigate(pathFromSession(ctx), { replace: true })
        }
      } catch {
        setUser(null)
        setSession(emptySessionContext())
      } finally {
        setReady(true)
      }
    },
    [hydrateSession, navigate],
  )

  useEffect(() => {
    void refreshMe({ resumeNav: true })
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const onAuthed = useCallback(async () => {
    setReady(false)
    await refreshMe({ resumeNav: true })
  }, [refreshMe])

  const onLogout = useCallback(() => {
    setUser(null)
    setSession(emptySessionContext())
    setHydrateNote(null)
  }, [])

  const onSessionChange = useCallback((ctx: SessionContext) => {
    setSession(ctx)
    setHydrateNote(null)
  }, [])

  const saveSession = useCallback(
    async (patch: SessionContextPut) => {
      try {
        const next = await putSessionContext({
          wsId: patch.wsId ?? session.wsId,
          teamId: patch.teamId ?? session.teamId,
          projectId: patch.projectId ?? session.projectId,
          issueId: patch.issueId ?? session.issueId,
          view: patch.view ?? session.view,
        })
        onSessionChange(next)
        return next
      } catch {
        return null
      }
    },
    [session, onSessionChange],
  )

  const sessionValue = useMemo(
    () =>
      user
        ? {
            user,
            session,
            hydrateNote,
            setSession: onSessionChange,
            saveSession,
            onLogout,
            onRehydrate: hydrateSession,
          }
        : null,
    [user, session, hydrateNote, onSessionChange, saveSession, onLogout, hydrateSession],
  )

  if (!ready) {
    return (
      <div className="m-shell m-shell--boot">
        <p className="m-muted">加载中…</p>
      </div>
    )
  }

  if (!user || !sessionValue) {
    return (
      <Routes>
        <Route path="/login" element={<LoginPage onAuthed={onAuthed} />} />
        <Route path="*" element={<Navigate to="/login" replace />} />
      </Routes>
    )
  }

  return (
    <SessionProvider value={sessionValue}>
      <Routes>
        <Route path="/login" element={<Navigate to="/crew" replace />} />
        <Route element={<AppShell />}>
          <Route path="/" element={<Navigate to="/crew" replace />} />
          <Route path="/crew" element={<CrewWorkspacePage />} />
          <Route path="/crew/:wsId" element={<CrewTeamsPage />} />
          <Route path="/crew/:wsId/teams/:teamId" element={<CrewSeatsPage />} />
          <Route path="/issues" element={<IssueListPage />} />
          <Route path="/w/:wsId/issues" element={<IssueListPage />} />
          <Route path="/w/:wsId/issues/:issueId" element={<IssueDetailPage />} />
          <Route path="/wiki" element={<WikiPage />} />
          <Route path="/w/:wsId/wiki" element={<WikiPage />} />
          <Route path="/resume" element={<ResumePage />} />
          <Route path="/me" element={<MePage />} />
        </Route>
        <Route path="*" element={<Navigate to="/crew" replace />} />
      </Routes>
    </SessionProvider>
  )
}
