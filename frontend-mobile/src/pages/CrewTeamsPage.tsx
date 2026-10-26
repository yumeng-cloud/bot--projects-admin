import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { listTeams, type Team } from '@acb/api-client'
import { useSession } from '../context/SessionCtx'

export default function CrewTeamsPage() {
  const { wsId = '' } = useParams()
  const { saveSession, session } = useSession()
  const [list, setList] = useState<Team[]>([])
  const [err, setErr] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    let cancel = false
    ;(async () => {
      setLoading(true)
      try {
        if (wsId && session.wsId !== wsId) {
          await saveSession({ wsId, view: 'crew' })
        }
        const { teams } = await listTeams(wsId)
        if (!cancel) setList(teams || [])
      } catch (e) {
        if (!cancel) setErr(e instanceof Error ? e.message : '加载团队失败')
      } finally {
        if (!cancel) setLoading(false)
      }
    })()
    return () => {
      cancel = true
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [wsId])

  const openTeam = async (t: Team) => {
    await saveSession({
      wsId,
      teamId: t.id,
      issueId: '',
      view: 'crew',
    })
  }

  return (
    <div className="m-page">
      <header className="m-page__head">
        <Link className="m-back" to="/crew">
          ← 工作区
        </Link>
        <h1 className="m-page__title">团队</h1>
        <p className="m-muted">一支队 · 点进看席位</p>
      </header>
      {loading ? <p className="m-muted">加载中…</p> : null}
      {err ? <p className="m-error">{err}</p> : null}
      <ul className="m-list">
        {list.map((t) => (
          <li key={t.id}>
            <Link
              className="m-list__row"
              to={`/crew/${wsId}/teams/${t.id}`}
              onClick={() => void openTeam(t)}
            >
              <span className="m-list__main">
                <strong>{t.name}</strong>
                <span className="m-muted">{t.slug || t.id.slice(0, 8)}</span>
              </span>
              {session.teamId === t.id ? <span className="m-badge m-badge--ok">当前</span> : null}
            </Link>
          </li>
        ))}
      </ul>
      {!loading && !err && list.length === 0 ? (
        <p className="m-muted">本工作区暂无团队。</p>
      ) : null}
    </div>
  )
}
