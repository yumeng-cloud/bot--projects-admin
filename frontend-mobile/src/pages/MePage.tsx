import { useState } from 'react'
import { Link } from 'react-router-dom'
import {
  logout,
  me,
  putSessionContext,
  hasSessionScene,
  type User,
  type SessionContext,
} from '@acb/api-client'

type Props = {
  user: User
  session: SessionContext
  hydrateNote: string | null
  onLogout: () => void
  onSessionChange: (ctx: SessionContext) => void
}

/** R0：/auth/me；R1：空现场安静默认，可 PUT 写现场 */
export default function MePage({
  user: initial,
  session,
  hydrateNote,
  onLogout,
  onSessionChange,
}: Props) {
  const [user, setUser] = useState(initial)
  const [checking, setChecking] = useState(false)
  const [meOk, setMeOk] = useState(true)
  const [msg, setMsg] = useState<string | null>(null)
  const [saving, setSaving] = useState(false)

  const recheck = async () => {
    setChecking(true)
    setMsg(null)
    try {
      const { user: u } = await me()
      setUser(u)
      setMeOk(true)
      setMsg('/auth/me 正常')
    } catch (e) {
      setMeOk(false)
      setMsg(e instanceof Error ? e.message : '/auth/me 失败')
    } finally {
      setChecking(false)
    }
  }

  const doLogout = async () => {
    try {
      await logout()
    } catch {
      /* token 已清 */
    }
    onLogout()
  }

  /** 演示 PUT：把当前安静默认写成一条可续现场（后写覆盖） */
  const saveDemoScene = async () => {
    setSaving(true)
    setMsg(null)
    try {
      const ctx = await putSessionContext({
        wsId: session.wsId || 'ws_mobile_demo',
        teamId: session.teamId || '',
        projectId: session.projectId || '',
        issueId: session.issueId || '',
        view: session.view || 'home',
      })
      onSessionChange(ctx)
      setMsg('已 PUT 现场（后写覆盖）')
    } catch (e) {
      setMsg(e instanceof Error ? e.message : '保存现场失败')
    } finally {
      setSaving(false)
    }
  }

  const scene = hasSessionScene(session)

  return (
    <div className="m-shell m-me">
      <header className="m-me__head">
        <h1 className="m-me__title">我的</h1>
        <p className="m-muted">MV2-R0/R1 · 登录 · /auth/me · 续现场</p>
      </header>

      {hydrateNote ? <p className="m-warn">{hydrateNote}</p> : null}

      <section className="m-card">
        <div className="m-me__row">
          <span className="m-muted">姓名</span>
          <strong>{user.name || '—'}</strong>
        </div>
        <div className="m-me__row">
          <span className="m-muted">邮箱</span>
          <strong>{user.email}</strong>
        </div>
        <div className="m-me__row">
          <span className="m-muted">用户 ID</span>
          <code className="m-code">{user.id}</code>
        </div>
        <div className="m-me__row">
          <span className="m-muted">/auth/me</span>
          <span className={meOk ? 'm-badge m-badge--ok' : 'm-badge m-badge--bad'}>
            {meOk ? '绿' : '红'}
          </span>
        </div>
        <div className="m-me__row">
          <span className="m-muted">现场</span>
          <span className={scene ? 'm-badge m-badge--ok' : 'm-badge'}>
            {scene ? '有值' : '空·安静'}
          </span>
        </div>
        {msg ? <p className={meOk || msg.includes('PUT') ? 'm-ok' : 'm-error'}>{msg}</p> : null}
      </section>

      <div className="m-me__actions">
        {scene ? (
          <Link className="m-btn m-btn--primary m-btn--link" to="/resume">
            打开续现场
          </Link>
        ) : null}
        <button className="m-btn m-btn--ghost" type="button" onClick={recheck} disabled={checking}>
          {checking ? '检查中…' : '再打 /auth/me'}
        </button>
        <button className="m-btn m-btn--ghost" type="button" onClick={saveDemoScene} disabled={saving}>
          {saving ? '保存中…' : 'PUT 写一条现场'}
        </button>
        <button className="m-btn m-btn--danger" type="button" onClick={doLogout}>
          退出登录
        </button>
      </div>

      {/* R2：底栏 Tab（队/席/看板/我的）— 本刀不做 */}
    </div>
  )
}
