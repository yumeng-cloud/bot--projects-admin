import { useCallback, useEffect, useState } from 'react'
import { Navigate, Route, Routes, useNavigate } from 'react-router-dom'
import {
  getToken,
  me,
  getSessionContext,
  emptySessionContext,
  hasSessionScene,
  type User,
  type SessionContext,
} from '@acb/api-client'
import LoginPage from './pages/LoginPage'
import MePage from './pages/MePage'
import ResumePage from './pages/ResumePage'

/**
 * MV2-R0：脚手架 + 登录 + /auth/me。
 * MV2-R1：进壳 hydrate session-context → 有值续现场 / 空安静默认；禁踢下线。
 * R2 底栏 Tab · 列表详情 — 本刀不做（空壳注释位）。
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
      // 禁踢下线：session-context 失败只安静默认，不清 token
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
          if (hasSessionScene(ctx)) {
            navigate('/resume', { replace: true })
          } else {
            navigate('/', { replace: true })
          }
        }
      } catch {
        // /auth/me 失败：登出态（token 可能已被 client 清）；不因 session 失败踢人
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
    // 仅首屏
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

  const onSessionSaved = useCallback((ctx: SessionContext) => {
    setSession(ctx)
    setHydrateNote(null)
  }, [])

  if (!ready) {
    return (
      <div className="m-shell m-shell--boot">
        <p className="m-muted">加载中…</p>
      </div>
    )
  }

  return (
    <Routes>
      <Route
        path="/login"
        element={user ? <Navigate to="/" replace /> : <LoginPage onAuthed={onAuthed} />}
      />
      <Route
        path="/"
        element={
          user ? (
            <MePage
              user={user}
              session={session}
              hydrateNote={hydrateNote}
              onLogout={onLogout}
              onSessionChange={onSessionSaved}
            />
          ) : (
            <Navigate to="/login" replace />
          )
        }
      />
      <Route
        path="/resume"
        element={
          user ? (
            <ResumePage
              session={session}
              hydrateNote={hydrateNote}
              onSessionChange={onSessionSaved}
              onRehydrate={hydrateSession}
            />
          ) : (
            <Navigate to="/login" replace />
          )
        }
      />
      {/* R2：底栏 Tab / 列表→详情 路由位 — 本刀留空 */}
      <Route path="*" element={<Navigate to={user ? '/' : '/login'} replace />} />
    </Routes>
  )
}
