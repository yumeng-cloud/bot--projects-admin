import { useState, type FormEvent } from 'react'
import { Link } from 'react-router-dom'
import { hasSessionScene, type SessionContext } from '@acb/api-client'
import { useSession } from '../context/SessionCtx'
import { pathFromSession } from '../lib/sessionNav'

export default function ResumePage() {
  const { session, hydrateNote, saveSession, onRehydrate } = useSession()
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
    const ctx = await saveSession({
      wsId: wsId.trim(),
      teamId: teamId.trim(),
      projectId: projectId.trim(),
      issueId: issueId.trim(),
      view: view.trim(),
    })
    if (ctx) {
      syncFrom(ctx)
      setMsg('已保存（后写覆盖）')
    } else {
      setMsg('保存失败（未踢下线）')
    }
    setBusy(false)
  }

  const clearScene = async () => {
    setBusy(true)
    setMsg(null)
    const ctx = await saveSession({
      wsId: '',
      teamId: '',
      projectId: '',
      issueId: '',
      view: '',
    })
    if (ctx) {
      syncFrom(ctx)
      setMsg('已清空现场')
    } else {
      setMsg('清空失败')
    }
    setBusy(false)
  }

  return (
    <div className="m-page">
      <header className="m-page__head">
        <Link className="m-back" to="/me">
          ← 我的
        </Link>
        <h1 className="m-page__title">续现场</h1>
        <p className="m-muted">session-context · 非踢下线</p>
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
          <code className="m-code">{session.updatedAt || '—'}</code>
        </div>
      </section>

      {hasSessionScene(session) ? (
        <Link className="m-btn m-btn--primary m-btn--link" to={pathFromSession(session)}>
          按现场打开
        </Link>
      ) : null}

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
      </div>
    </div>
  )
}
