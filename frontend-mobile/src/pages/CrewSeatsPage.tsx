import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { getRoster, getTeam, type Team, type TeamSeat } from '@acb/api-client'
import { useSession } from '../context/SessionCtx'
import { issuesPath } from '../lib/sessionNav'

function seatLabel(s: TeamSeat): string {
  if (s.bound_display_name) return s.bound_display_name
  if (s.claim) return '已占用'
  return '空席'
}

export default function CrewSeatsPage() {
  const { wsId = '', teamId = '' } = useParams()
  const { saveSession } = useSession()
  const [team, setTeam] = useState<Team | null>(null)
  const [roster, setRoster] = useState<TeamSeat[]>([])
  const [err, setErr] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    let cancel = false
    ;(async () => {
      setLoading(true)
      try {
        await saveSession({ wsId, teamId, view: 'crew' })
        const [{ team: t }, { roster: r }] = await Promise.all([
          getTeam(wsId, teamId),
          getRoster(wsId, teamId),
        ])
        if (!cancel) {
          setTeam(t)
          setRoster(r || [])
        }
      } catch (e) {
        if (!cancel) setErr(e instanceof Error ? e.message : '加载席位失败')
      } finally {
        if (!cancel) setLoading(false)
      }
    })()
    return () => {
      cancel = true
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [wsId, teamId])

  return (
    <div className="m-page">
      <header className="m-page__head">
        <Link className="m-back" to={`/crew/${wsId}`}>
          ← 团队
        </Link>
        <h1 className="m-page__title">{team?.name || '席位'}</h1>
        <p className="m-muted">人 + Agent 一支队</p>
      </header>
      {loading ? <p className="m-muted">加载中…</p> : null}
      {err ? <p className="m-error">{err}</p> : null}
      <ul className="m-list">
        {roster.map((s) => (
          <li key={s.id} className="m-list__row m-list__row--static">
            <span className="m-list__main">
              <strong>{s.display_name || s.role_slug}</strong>
              <span className="m-muted">
                {s.role_slug} · {seatLabel(s)}
              </span>
            </span>
          </li>
        ))}
      </ul>
      {!loading && !err && roster.length === 0 ? (
        <p className="m-muted">暂无席位。</p>
      ) : null}
      <div className="m-me__actions">
        <Link className="m-btn m-btn--primary m-btn--link" to={issuesPath(wsId)}>
          去 Issue 列表
        </Link>
      </div>
    </div>
  )
}
