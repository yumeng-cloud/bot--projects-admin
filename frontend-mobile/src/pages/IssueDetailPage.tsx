import { useEffect, useState, type FormEvent } from 'react'
import { Link, useParams } from 'react-router-dom'
import {
  ISSUE_STATUS_LABEL,
  ISSUE_STATUSES,
  createComment,
  getIssue,
  listComments,
  patchIssue,
  type Comment,
  type Issue,
  type IssueStatus,
} from '@acb/api-client'
import { useSession } from '../context/SessionCtx'

export default function IssueDetailPage() {
  const { wsId = '', issueId = '' } = useParams()
  const { saveSession } = useSession()
  const [issue, setIssue] = useState<Issue | null>(null)
  const [comments, setComments] = useState<Comment[]>([])
  const [err, setErr] = useState<string | null>(null)
  const [loading, setLoading] = useState(true)
  const [saving, setSaving] = useState(false)
  const [note, setNote] = useState<string | null>(null)
  const [draft, setDraft] = useState('')

  const load = async () => {
    setLoading(true)
    setErr(null)
    try {
      const [{ issue: i }, { comments: c }] = await Promise.all([
        getIssue(wsId, issueId),
        listComments(wsId, issueId),
      ])
      setIssue(i)
      setComments(c || [])
      await saveSession({
        wsId,
        projectId: i.project_id,
        issueId: i.id,
        teamId: i.team_id || undefined,
        view: 'issue',
      })
    } catch (e) {
      setErr(e instanceof Error ? e.message : '加载详情失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void load()
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [wsId, issueId])

  const changeStatus = async (status: IssueStatus) => {
    if (!issue) return
    setSaving(true)
    setNote(null)
    try {
      const { issue: next } = await patchIssue(wsId, issue.id, {
        version: issue.version,
        status,
      })
      setIssue(next)
      setNote('状态已更新')
      await saveSession({
        wsId,
        projectId: next.project_id,
        issueId: next.id,
        teamId: next.team_id || undefined,
        view: 'issue',
      })
    } catch (e) {
      setNote(e instanceof Error ? e.message : '改状态失败')
    } finally {
      setSaving(false)
    }
  }

  const sendComment = async (e: FormEvent) => {
    e.preventDefault()
    const body = draft.trim()
    if (!body || !issue) return
    setSaving(true)
    setNote(null)
    try {
      const { comment } = await createComment(wsId, issue.id, body)
      setComments((prev) => [...prev, comment])
      setDraft('')
      setNote('评论已发送')
    } catch (err) {
      setNote(err instanceof Error ? err.message : '评论失败')
    } finally {
      setSaving(false)
    }
  }

  return (
    <div className="m-page">
      <header className="m-page__head">
        <Link className="m-back" to={`/w/${wsId}/issues`}>
          ← Issue 列表
        </Link>
        <h1 className="m-page__title">{issue?.identifier || '详情'}</h1>
      </header>

      {loading ? <p className="m-muted">加载中…</p> : null}
      {err ? <p className="m-error">{err}</p> : null}

      {issue ? (
        <>
          <section className="m-card">
            <h2 className="m-detail__title">{issue.title}</h2>
            <p className="m-muted m-detail__meta">
              {ISSUE_STATUS_LABEL[issue.status]} · v{issue.version}
              {issue.assignee?.display_name ? ` · ${issue.assignee.display_name}` : ''}
            </p>
            <div className="m-detail__desc">
              {issue.description?.trim() ? issue.description : '（无描述）'}
            </div>
          </section>

          <section className="m-card m-detail__status">
            <p className="m-field__label">改状态</p>
            <div className="m-chip-row">
              {ISSUE_STATUSES.map((s) => (
                <button
                  key={s}
                  type="button"
                  className={`m-chip${issue.status === s ? ' is-on' : ''}`}
                  disabled={saving || issue.status === s}
                  onClick={() => void changeStatus(s)}
                >
                  {ISSUE_STATUS_LABEL[s]}
                </button>
              ))}
            </div>
            {note ? <p className="m-ok">{note}</p> : null}
          </section>

          <section className="m-card">
            <h3 className="m-section-title">评论</h3>
            <ul className="m-comments">
              {comments.map((c) => (
                <li key={c.id} className="m-comments__item">
                  <strong>{c.author?.display_name || '成员'}</strong>
                  <p>{c.body}</p>
                </li>
              ))}
            </ul>
            {comments.length === 0 ? <p className="m-muted">暂无评论</p> : null}
            <form className="m-comment-form" onSubmit={sendComment}>
              <textarea
                className="m-textarea"
                rows={3}
                placeholder="写评论…"
                value={draft}
                onChange={(e) => setDraft(e.target.value)}
              />
              <button className="m-btn m-btn--primary" type="submit" disabled={saving || !draft.trim()}>
                发送评论
              </button>
            </form>
          </section>
        </>
      ) : null}
    </div>
  )
}
