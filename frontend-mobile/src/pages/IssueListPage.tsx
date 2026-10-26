import { useEffect, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import {
  ISSUE_STATUS_LABEL,
  listIssues,
  type Issue,
} from '@acb/api-client'
import { useSession } from '../context/SessionCtx'

export default function IssueListPage() {
  const params = useParams()
  const nav = useNavigate()
  const { session, saveSession } = useSession()
  const wsId = params.wsId || session.wsId

  const [list, setList] = useState<Issue[]>([])
  const [err, setErr] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    if (!wsId) return
    if (!params.wsId && session.wsId) {
      nav(`/w/${session.wsId}/issues`, { replace: true })
      return
    }
    let cancel = false
    ;(async () => {
      setLoading(true)
      setErr(null)
      try {
        await saveSession({
          wsId,
          view: 'issues',
          issueId: '',
        })
        const { issues } = await listIssues(wsId, {
          team_id: session.teamId || undefined,
        })
        if (!cancel) setList(issues || [])
      } catch (e) {
        if (!cancel) setErr(e instanceof Error ? e.message : '加载 Issue 失败')
      } finally {
        if (!cancel) setLoading(false)
      }
    })()
    return () => {
      cancel = true
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [wsId, session.teamId, params.wsId])

  if (!wsId) {
    return (
      <div className="m-page">
        <header className="m-page__head">
          <h1 className="m-page__title">Issue</h1>
          <p className="m-muted">请先选择工作区 / 队</p>
        </header>
        <Link className="m-btn m-btn--primary m-btn--link" to="/crew">
          去队席选工作区
        </Link>
      </div>
    )
  }

  return (
    <div className="m-page">
      <header className="m-page__head">
        <h1 className="m-page__title">Issue</h1>
        <p className="m-muted">
          {session.teamId ? '当前队过滤' : '工作区全部'} · 点进详情
        </p>
      </header>
      {loading ? <p className="m-muted">加载中…</p> : null}
      {err ? <p className="m-error">{err}</p> : null}
      <ul className="m-list">
        {list.map((i) => (
          <li key={i.id}>
            <Link className="m-list__row" to={`/w/${wsId}/issues/${i.id}`}>
              <span className="m-list__main">
                <strong>
                  <span className="m-id">{i.identifier}</span> {i.title}
                </strong>
                <span className="m-muted">
                  {ISSUE_STATUS_LABEL[i.status] || i.status}
                  {i.assignee?.display_name ? ` · ${i.assignee.display_name}` : ''}
                </span>
              </span>
            </Link>
          </li>
        ))}
      </ul>
      {!loading && !err && list.length === 0 ? (
        <p className="m-muted">暂无 Issue。</p>
      ) : null}
    </div>
  )
}
