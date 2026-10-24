import { useEffect, useMemo, useState } from 'react'
import { Link, Outlet, useLocation, useNavigate, useParams } from 'react-router-dom'
import { Layout, Menu, Typography, Button, Space, Spin, Select, theme } from 'antd'
import {
  DashboardOutlined,
  UnorderedListOutlined,
  FieldTimeOutlined,
  ScheduleOutlined,
  BookOutlined,
  ClockCircleOutlined,
  SettingOutlined,
  RobotOutlined,
  LogoutOutlined,
  BulbOutlined,
  MoonOutlined,
  CrownOutlined,
} from '@ant-design/icons'
import { listMyWorkspaces, type Workspace } from '../api/workspaces'
import { listProjects, type Project } from '../api/projects'
import type { User } from '../api/auth'
import { THEME_MODE_STORAGE_KEY, resolveThemeMode, type ThemeMode } from '../theme'
import { TeamPoolSelect } from '../components/common/TeamPoolSelect'
import { CurrentTeamProvider, useCurrentTeam } from '../hooks/useCurrentTeam'

const { Header, Sider, Content } = Layout

function AppHeaderTeam() {
  const { wsId } = useParams()
  const { teams, teamId, setTeamId, loading, reload } = useCurrentTeam(wsId || '')
  return (
    <span className="acb-header-team">
      <TeamPoolSelect
        teams={teams}
        teamId={teamId}
        onChange={(id) => setTeamId(id)}
        disabled={loading || !wsId}
        workspaceId={wsId}
        onTeamCreated={() => {
          void reload()
        }}
      />
    </span>
  )
}

function AppLayoutShell({ user, onLogout }: { user: User | null; onLogout: () => void }) {
  const { wsId, projectId: routeProjectId } = useParams()
  const nav = useNavigate()
  const loc = useLocation()
  const { token: t } = theme.useToken()
  const [ws, setWs] = useState<Workspace | null>(null)
  const [projects, setProjects] = useState<Project[]>([])
  const [projectId, setProjectId] = useState<string | undefined>(routeProjectId)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    if (!wsId) return
    localStorage.setItem('workspaceId', wsId)
    ;(async () => {
      try {
        const [wsRes, projRes] = await Promise.all([
          listMyWorkspaces(),
          listProjects(wsId).catch(() => ({ projects: [] as Project[] })),
        ])
        setWs(wsRes.workspaces.find((w) => w.id === wsId) || null)
        setProjects(projRes.projects || [])
        setProjectId(routeProjectId || projRes.projects?.[0]?.id)
      } finally {
        setLoading(false)
      }
    })()
  }, [wsId, routeProjectId])

  const selected = useMemo(() => {
    const p = loc.pathname
    if (p.includes('/issues') || p.includes('/list')) return ['list']
    if (p.includes('/timeline')) return ['timeline']
    if (p.includes('/cycles')) return ['cycles']
    if (p.includes('/wiki')) return ['wiki']
    if (p.includes('/time')) return ['time']
    if (p.includes('/settings/bots')) return ['bots']
    if (p.includes('/settings')) return ['settings']
    return ['board']
  }, [loc.pathname])

  const base = projectId ? `/w/${wsId}/projects/${projectId}` : `/w/${wsId}`
  const themeMode = resolveThemeMode()
  const toggleTheme = () => {
    const next: ThemeMode = themeMode === 'dark' ? 'light' : 'dark'
    localStorage.setItem(THEME_MODE_STORAGE_KEY, next)
    document.documentElement.setAttribute('data-theme', next)
    document.documentElement.style.colorScheme = next
    location.reload()
  }

  if (loading) {
    return (
      <div style={{ height: '100%', display: 'grid', placeItems: 'center' }}>
        <Spin size="small" />
      </div>
    )
  }

  return (
    <Layout style={{ height: '100%', overflow: 'hidden' }}>
      <Sider
        width={240}
        breakpoint="lg"
        collapsedWidth={64}
        theme="light"
        className="acb-sider"
        style={{
          borderRight: `1px solid ${t.colorBorderSecondary}`,
          height: '100%',
          overflow: 'hidden',
        }}
      >
        <div
          style={{
            height: 48,
            flexShrink: 0,
            display: 'flex',
            alignItems: 'center',
            padding: '0 14px',
            borderBottom: `1px solid ${t.colorBorderSecondary}`,
          }}
        >
          <Typography.Text strong ellipsis style={{ fontSize: 13 }}>
            {ws?.name || '工作区'}
          </Typography.Text>
        </div>
        <Menu
          mode="inline"
          selectedKeys={selected}
          className="acb-sider-menu"
          style={{ border: 'none', marginTop: 4 }}
          items={[
            { key: 'board', icon: <DashboardOutlined />, label: <Link to={projectId ? `${base}/board` : `/w/${wsId}/board`}>看板</Link> },
            { key: 'list', icon: <UnorderedListOutlined />, label: <Link to={projectId ? `${base}/issues` : `/w/${wsId}/list`}>列表</Link> },
            { key: 'cycles', icon: <FieldTimeOutlined />, label: <Link to={projectId ? `${base}/cycles` : `/w/${wsId}/cycles`}>周期</Link> },
            { key: 'timeline', icon: <ScheduleOutlined />, label: <Link to={projectId ? `${base}/timeline` : `/w/${wsId}/timeline`}>排期</Link> },
            { key: 'wiki', icon: <BookOutlined />, label: <Link to={`/w/${wsId}/wiki`}>Wiki</Link> },
            { key: 'time', icon: <ClockCircleOutlined />, label: <Link to={`/w/${wsId}/time`}>工时</Link> },
            { type: 'divider' },
            { key: 'settings', icon: <SettingOutlined />, label: <Link to={`/w/${wsId}/settings`}>设置</Link> },
            { key: 'bots', icon: <RobotOutlined />, label: <Link to={`/w/${wsId}/settings/bots`}>Bot</Link> },
          ]}
        />
        {user?.is_platform_admin ? (
          <div className="acb-sider-foot">
            <Link
              to="/admin"
              className={`acb-sider-foot-link${loc.pathname.startsWith('/admin') ? ' is-active' : ''}`}
              title="平台管理（仅超管）"
            >
              <CrownOutlined style={{ fontSize: 12 }} />
              <span>平台管理</span>
            </Link>
          </div>
        ) : null}
      </Sider>
      <Layout style={{ height: '100%', minWidth: 0, overflow: 'hidden' }}>
        <Header
          style={{
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            borderBottom: `1px solid ${t.colorBorderSecondary}`,
            background: t.colorBgContainer,
            flexShrink: 0,
          }}
        >
          <Space className="acb-header-context" size={10} align="center">
            <Select
              size="small"
              variant="borderless"
              style={{ minWidth: 160 }}
              placeholder="项目"
              value={projectId}
              options={projects.map((p) => ({ value: p.id, label: `${p.key_prefix} · ${p.name}` }))}
              onChange={(id) => {
                setProjectId(id)
                const tab =
                  selected[0] === 'list'
                    ? 'issues'
                    : selected[0] === 'cycles'
                      ? 'cycles'
                      : selected[0] === 'timeline'
                        ? 'timeline'
                        : 'board'
                nav(`/w/${wsId}/projects/${id}/${tab}`)
              }}
            />
            <AppHeaderTeam />
          </Space>
          <Space size={2}>
            <Typography.Text type="secondary" style={{ fontSize: 12 }}>
              {user?.name || user?.email}
            </Typography.Text>
            <Button
              type="text"
              size="small"
              icon={themeMode === 'dark' ? <BulbOutlined /> : <MoonOutlined />}
              onClick={toggleTheme}
              title={themeMode === 'dark' ? '浅色' : '深色'}
            />
            <Button type="text" size="small" onClick={() => nav('/')}>
              工作区
            </Button>
            <Button type="text" size="small" icon={<LogoutOutlined />} onClick={onLogout} />
          </Space>
        </Header>
        <Content
          style={{
            flex: 1,
            minHeight: 0,
            overflow: 'auto',
            display: 'flex',
            flexDirection: 'column',
          }}
        >
          <Outlet />
        </Content>
      </Layout>
    </Layout>
  )
}

export default function AppLayout({ user, onLogout }: { user: User | null; onLogout: () => void }) {
  const { wsId } = useParams()
  if (!wsId) {
    return (
      <div style={{ height: '100%', display: 'grid', placeItems: 'center' }}>
        <Spin size="small" />
      </div>
    )
  }
  return (
    <CurrentTeamProvider workspaceId={wsId}>
      <AppLayoutShell user={user} onLogout={onLogout} />
    </CurrentTeamProvider>
  )
}
