import { useEffect, useMemo, useState } from 'react'
import { BrowserRouter, Navigate, Route, Routes, useLocation } from 'react-router-dom'
import LoginPage from './pages/LoginPage'
import Landing from './pages/Landing'
import HomePage from './pages/HomePage'
import JoinPage from './pages/JoinPage'
import AppLayout from './layouts/AppLayout'
import WikiPage from './pages/wiki/WikiPage'
import TimePage from './pages/time/TimePage'
import ProjectRedirect from './pages/ProjectRedirect'
import { BoardRoute, ListRoute, CyclesRoute, TimelineRoute } from './pages/ProjectScoped'
import { getToken, setToken } from './api/client'
import { logout, me, type User } from './api/auth'
import SettingsPage from './pages/settings/SettingsPage'
import { BotsRosterPage, BotDetailPage } from './pages/settings/bots'
import PlatformAdminPage from './pages/admin/PlatformAdminPage'

/** Unauthed hit on protected path → /login, keep deep-link intent. */
function RedirectToLogin() {
  const location = useLocation()
  const from = `${location.pathname}${location.search}${location.hash}`
  return <Navigate to="/login" replace state={{ from }} />
}

export default function App() {
  const [token, setTok] = useState(() => getToken())
  const [user, setUser] = useState<User | null>(null)
  const authed = useMemo(() => !!token, [token])

  useEffect(() => {
    if (!token) {
      setUser(null)
      return
    }
    me()
      .then((r) => setUser(r.user))
      .catch(() => {
        setToken(null)
        setTok(null)
      })
  }, [token])

  const onLogout = () => {
    logout().catch(() => {})
    setToken(null)
    setTok(null)
  }

  const onAuth = () => setTok(getToken())

  return (
    <BrowserRouter>
      <Routes>
        {/* ONE_TAP C1：同源 /join 深链（登录→预填向导；未登录→短引导） */}
        <Route path="/join" element={<JoinPage authed={authed} onAuth={onAuth} />} />
        {!authed ? (
          <>
            <Route path="/" element={<Landing />} />
            <Route path="/login" element={<LoginPage onAuth={onAuth} />} />
            <Route path="*" element={<RedirectToLogin />} />
          </>
        ) : (
          <>
            <Route path="/" element={<HomePage onLogout={onLogout} />} />
            <Route path="/login" element={<Navigate to="/" replace />} />
            <Route path="/admin" element={<PlatformAdminPage />} />
            <Route path="/admin/*" element={<PlatformAdminPage />} />
            <Route path="/w/:wsId" element={<AppLayout user={user} onLogout={onLogout} />}>
              <Route index element={<Navigate to="board" replace />} />
              <Route path="board" element={<ProjectRedirect tab="board" />} />
              <Route path="list" element={<ProjectRedirect tab="issues" />} />
              <Route path="cycles" element={<ProjectRedirect tab="cycles" />} />
              <Route path="timeline" element={<ProjectRedirect tab="timeline" />} />
              <Route path="projects/:projectId/board" element={<BoardRoute />} />
              <Route path="projects/:projectId/issues" element={<ListRoute />} />
              <Route path="projects/:projectId/cycles" element={<CyclesRoute />} />
              <Route path="projects/:projectId/timeline" element={<TimelineRoute />} />
              <Route path="wiki/*" element={<WikiPage />} />
              <Route path="time/*" element={<TimePage />} />
              <Route path="settings" element={<SettingsPage />} />
              <Route path="settings/bots" element={<BotsRosterPage />} />
              <Route path="settings/bots/:memberId" element={<BotDetailPage />} />
            </Route>
            <Route path="*" element={<Navigate to="/" replace />} />
          </>
        )}
      </Routes>
    </BrowserRouter>
  )
}
