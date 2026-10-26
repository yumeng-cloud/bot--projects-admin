import { useCallback, useEffect, useMemo, useRef, useState } from 'react'
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
 *
 * 续现场落点：hydrate 后按 session-context 进深链。已登录时 /login 必须落到
 * pathFromSession(session)，禁止写死 /crew——否则与 resumeNav 竞态时会闪深链再被顶回队席。
 */
export default function App() {
  const navigate = useNavigate()
  const [ready, setReady] = useState(false)
  const [user, setUser] = useState<User | null>(null)
  const [session, setSession] = useState<SessionContext>(emptySessionContext())
  const [hydrateNote, setHydrateNote] = useState<string | null>(null)
  /** 丢弃过期的 refreshMe（登录/登出叠呼时避免后到的结果把深链顶掉） */
  const navGenRef = useRef(0)

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
      const gen = ++navGenRef.current
      if (!getToken()) {
        if (gen !== navGenRef.current) return
        setUser(null)
        setSession(emptySessionContext())
        setReady(true)
        return
      }
      try {
        const { user: u } = await me()
        if (gen !== navGenRef.current) return

        let ctx: SessionContext
        let note: string | null = null
        try {
          ctx = await getSessionContext()
        } catch {
          ctx = emptySessionContext()
          note = '现场同步暂不可用，已用安静默认'
        }
        if (gen !== navGenRef.current) return

        // 同步写入再揭开路由树，保证已登录 /login 重定向读到有现场的 session
        setUser(u)
        setSession(ctx)
        setHydrateNote(note)
        if (opts?.resumeNav) {
          navigate(pathFromSession(ctx), { replace: true })
        }
      } catch {
        if (gen !== navGenRef.current) return
        setUser(null)
        setSession(emptySessionContext())
      } finally {
        if (gen === navGenRef.current) {
          setReady(true)
        }
      }
    },
    [navigate],
  )

  useEffect(() => {
    void refreshMe({ resumeNav: true })
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  const onAuthed = useCallback(async () => {
    // 收起路由树，避免仍停在 /login 时挂上错误的已登录重定向
    setReady(false)
    await refreshMe({ resumeNav: true })
  }, [refreshMe])

  const onLogout = useCallback(() => {
    navGenRef.current += 1
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

  const resumeTo = pathFromSession(session)

  return (
    <SessionProvider value={sessionValue}>
      <Routes>
        {/* 禁写死 /crew：与 resumeNav 竞态会闪深链再被顶回队席（MV2-Q1 B2） */}
        <Route path="/login" element={<Navigate to={resumeTo} replace />} />
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
