import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { listMyWorkspaces, type Workspace } from '@acb/api-client'
import { useSession } from '../context/SessionCtx'

export default function CrewWorkspacePage() {
  const { saveSession, session } = useSession()
  const [list, setList] = useState<Workspace[]>([])
  const [err, setErr] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    let cancel = false
    ;(async () => {
      setLoading(true)
      try {
        const { workspaces } = await listMyWorkspaces()
        if (!cancel) setList(workspaces || [])
      } catch (e) {
        if (!cancel) setErr(e instanceof Error ? e.message : '加载工作区失败')
      } finally {
        if (!cancel) setLoading(false)
      }
    })()
    return () => {
      cancel = true
    }
  }, [])

  const openWs = async (ws: Workspace) => {
    await saveSession({
      wsId: ws.id,
      teamId: '',
      projectId: '',
      issueId: '',
      view: 'crew',
    })
  }

  return (
    <div className="m-page">
      <header className="m-page__head">
        <h1 className="m-page__title">队席</h1>
        <p className="m-muted">工作区 → 团队 → 席位</p>
      </header>
      {loading ? <p className="m-muted">加载中…</p> : null}
      {err ? <p className="m-error">{err}</p> : null}
      <ul className="m-list">
        {list.map((ws) => (
          <li key={ws.id}>
            <Link
              className="m-list__row"
              to={`/crew/${ws.id}`}
              onClick={() => void openWs(ws)}
            >
              <span className="m-list__main">
                <strong>{ws.name}</strong>
                <span className="m-muted">{ws.slug}</span>
              </span>
              {session.wsId === ws.id ? <span className="m-badge m-badge--ok">当前</span> : null}
            </Link>
          </li>
        ))}
      </ul>
      {!loading && !err && list.length === 0 ? (
        <p className="m-muted">暂无工作区。请先在 PC 端创建。</p>
      ) : null}
    </div>
  )
}
