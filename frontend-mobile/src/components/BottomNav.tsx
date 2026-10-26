import { NavLink, useLocation } from 'react-router-dom'
import { useSession } from '../context/SessionCtx'
import { issuesPath, wikiPath } from '../lib/sessionNav'

export default function BottomNav() {
  const { session } = useSession()
  const loc = useLocation()
  const ws = session.wsId || undefined

  const tabs: { key: string; label: string; to: string; active: boolean }[] = [
    {
      key: 'crew',
      label: '队席',
      to: '/crew',
      active: loc.pathname.startsWith('/crew'),
    },
    {
      key: 'issues',
      label: 'Issue',
      to: issuesPath(ws),
      active:
        loc.pathname === '/issues' ||
        /^\/w\/[^/]+\/issues/.test(loc.pathname),
    },
    {
      key: 'wiki',
      label: 'Wiki',
      to: wikiPath(ws),
      active: loc.pathname === '/wiki' || /^\/w\/[^/]+\/wiki/.test(loc.pathname),
    },
    {
      key: 'me',
      label: '我的',
      to: '/me',
      active: loc.pathname === '/me' || loc.pathname === '/resume',
    },
  ]

  return (
    <nav className="m-tabbar" aria-label="主导航">
      {tabs.map((t) => (
        <NavLink
          key={t.key}
          to={t.to}
          className={`m-tabbar__item${t.active ? ' is-active' : ''}`}
        >
          {t.label}
        </NavLink>
      ))}
    </nav>
  )
}
