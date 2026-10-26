import { useState, type FormEvent } from 'react'
import { Link } from 'react-router-dom'
import {
  putSessionContext,
  hasSessionScene,
  type SessionContext,
} from '@acb/api-client'

type Props = {
  session: SessionContext
  hydrateNote: string | null
  onSessionChange: (ctx: SessionContext) => void
  onRehydrate: () => Promise<SessionContext>
}

/** R1：有值续现场页；可再 GET / PUT；失败不踢下线 */
export default function ResumePage({
  session,
  hydrateNote,
  onSessionChange,
  onRehydrate,
}: Props) {
  const [wsId, setWsId] = useState(session.wsId)
  const [teamId, setTeamId] = useState(session.teamId)
  const [projectId, setProjectId] = useState(session.projectId)
  const [issueId, setIssueId] = useState(session.issueId)
  const [view, setView] = useState(session.view || 'home')
  const [busy, setBusy] = useState(false)
  const [msg, setMsg] = useState<string | null>(null)

  const syncFrom = (ctx: SessionContext) => {
    setWsId(ctx.wsId)
    setTeamId(ctx.teamId)
    setProjectId(ctx.projectId)
    setIssueId(ctx.issueId)
    setView(ctx.view || 'home')
    onSessionChange(ctx)
  }

  const reload = async () => {
    setBusy(true)
    setMsg(null)
    try {
      const ctx = await onRehydrate()
      syncFrom(ctx)
      setMsg(hasSessionScene(ctx) ? '已重新 hydrate' : '现场为空（安静默认）')
    } catch (e) {
      setMsg(e instanceof Error ? e.message : 'hydrate 失败（未踢下线）')
    } finally {
      setBusy(false)
    }
  }

  const onSave = async (e: FormEvent) => {
    e.preventDefault()
    setBusy(true)
    setMsg(null)
    try {
      const ctx = await putSessionContext({
        wsId: wsId.trim(),
        teamId: teamId.trim(),
        projectId: projectId.trim(),
        issueId: issueId.trim(),
        view: view.trim(),
      })
      syncFrom(ctx)
      setMsg('已保存（后写覆盖）')
    } catch (err) {
      setMsg(err instanceof Error ? err.message : '保存失败（未踢下线）')
    } finally {
      setBusy(false)
    }
  }

  const clearScene = async () => {
    setBusy(true)
    setMsg(null)
    try {
      const ctx = await putSessionContext({
        wsId: '',
        teamId: '',
        projectId: '',
        issueId: '',
        view: '',
      })
      syncFrom(ctx)
      setMsg('已清空现场')
    } catch (err) {
      setMsg(err instanceof Error ? err.message : '清空失败')
    } finally {
      setBusy(false)
    }
  }

  // keep form in sync if parent session changes externally
  const updatedAt = session.updatedAt

  return (
    <div className="m-shell m-resume">
      <header className="m-me__head">
        <h1 className="m-me__title">续现场</h1>
        <p className="m-muted">MV2-R1 · session-context hydrate（非踢下线）</p>
      </header>

      {hydrateNote ? <p className="m-warn">{hydrateNote}</p> : null}

      <section className="m-card m-resume__summary">
        <div className="m-me__row">
          <span className="m-muted">状态</span>
          <span className={hasSessionScene(session) ? 'm-badge m-badge--ok' : 'm-badge'}>
            {hasSessionScene(session) ? '已续' : '空'}
          </span>
        </div>
        <div className="m-me__row">
          <span className="m-muted">updatedAt</span>
          <code className="m-code">{updatedAt || '—'}</code>
        </div>
      </section>

      <form className="m-card m-login__form" onSubmit={onSave}>
        <label className="m-field">
          <span className="m-field__label">wsId</span>
          <input className="m-input" value={wsId} onChange={(e) => setWsId(e.target.value)} />
        </label>
        <label className="m-field">
          <span className="m-field__label">teamId</span>
          <input className="m-input" value={teamId} onChange={(e) => setTeamId(e.target.value)} />
        </label>
        <label className="m-field">
          <span className="m-field__label">projectId</span>
          <input
            className="m-input"
            value={projectId}
            onChange={(e) => setProjectId(e.target.value)}
          />
        </label>
        <label className="m-field">
          <span className="m-field__label">issueId</span>
          <input className="m-input" value={issueId} onChange={(e) => setIssueId(e.target.value)} />
        </label>
        <label className="m-field">
          <span className="m-field__label">view</span>
          <input className="m-input" value={view} onChange={(e) => setView(e.target.value)} />
        </label>

        {msg ? <p className="m-ok">{msg}</p> : null}

        <button className="m-btn m-btn--primary" type="submit" disabled={busy}>
          {busy ? '处理中…' : 'PUT 更新现场'}
        </button>
      </form>

      <div className="m-me__actions">
        <button className="m-btn m-btn--ghost" type="button" onClick={reload} disabled={busy}>
          再 GET hydrate
        </button>
        <button className="m-btn m-btn--ghost" type="button" onClick={clearScene} disabled={busy}>
          清空现场
        </button>
        <Link className="m-btn m-btn--ghost m-btn--link" to="/">
          回我的
        </Link>
      </div>

      {/* R2：按 view/ws/issue 进列表·详情 — 本刀不做 */}
      <p className="m-muted m-resume__hint">列表 / 详情导航见 MV2-R2。</p>
    </div>
  )
}
