import { useEffect, useState } from 'react'
import { Link, useNavigate, useParams } from 'react-router-dom'
import { listWikiSpaces, type WikiSpace } from '@acb/api-client'
import { useSession } from '../context/SessionCtx'

/** Wiki 入口占位：列 space；正文阅读留给后续 */
export default function WikiPage() {
  const params = useParams()
  const nav = useNavigate()
  const { session, saveSession } = useSession()
  const wsId = params.wsId || session.wsId
  const [spaces, setSpaces] = useState<WikiSpace[]>([])
  const [err, setErr] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)

  useEffect(() => {
    if (!wsId) return
    if (!params.wsId && session.wsId) {
      nav(`/w/${session.wsId}/wiki`, { replace: true })
      return
    }
    let cancel = false
    ;(async () => {
      setLoading(true)
      try {
        await saveSession({ wsId, view: 'wiki', issueId: '' })
        const { spaces: s } = await listWikiSpaces(wsId)
        if (!cancel) setSpaces(s || [])
      } catch (e) {
        if (!cancel) setErr(e instanceof Error ? e.message : '加载 Wiki 失败')
      } finally {
        if (!cancel) setLoading(false)
      }
    })()
    return () => {
      cancel = true
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [wsId, params.wsId])

  if (!wsId) {
    return (
      <div className="m-page">
        <header className="m-page__head">
          <h1 className="m-page__title">Wiki</h1>
          <p className="m-muted">请先选择工作区</p>
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
        <h1 className="m-page__title">Wiki</h1>
        <p className="m-muted">知识库入口（占位 · 列空间）</p>
      </header>
      {loading ? <p className="m-muted">加载中…</p> : null}
      {err ? <p className="m-error">{err}</p> : null}
      <ul className="m-list">
        {spaces.map((s) => (
          <li key={s.id} className="m-list__row m-list__row--static">
            <span className="m-list__main">
              <strong>{s.name}</strong>
              <span className="m-muted">{s.slug || s.description || '空间'}</span>
            </span>
          </li>
        ))}
      </ul>
      {!loading && !err && spaces.length === 0 ? (
        <p className="m-muted">暂无 Wiki 空间。</p>
      ) : null}
    </div>
  )
}
