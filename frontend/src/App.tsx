import { useEffect, useMemo, useState } from 'react'
import { BrowserRouter, Navigate, Route, Routes } from 'react-router-dom'
import LoginPage from './pages/LoginPage'
import HomePage from './pages/HomePage'
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

  if (!authed) {
    return <LoginPage onAuth={() => setTok(getToken())} />
  }

  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<HomePage onLogout={onLogout} />} />
        <Route path="/admin" element={<PlatformAdminPage />} />
        <Route path="/admin/*" element={<PlatformAdminPage />} />
        <Route path="/w/:wsId" element={<AppLayout user={user} onLogout={onLogout} />}>
          <Route index element={<Navigate to="board" replace />} />
          {/* 短路径 → 首个项目 */}
          <Route path="board" element={<ProjectRedirect tab="board" />} />
          <Route path="list" element={<ProjectRedirect tab="issues" />} />
          <Route path="cycles" element={<ProjectRedirect tab="cycles" />} />
          <Route path="timeline" element={<ProjectRedirect tab="timeline" />} />
          {/* 真页面 */}
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
      </Routes>
    </BrowserRouter>
  )
}
