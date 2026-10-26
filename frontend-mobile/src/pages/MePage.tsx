import { useState } from 'react'
import { Link } from 'react-router-dom'
import { logout, me, hasSessionScene } from '@acb/api-client'
import { useSession } from '../context/SessionCtx'
import { pathFromSession } from '../lib/sessionNav'

export default function MePage() {
  const { user: initial, session, hydrateNote, onLogout, saveSession } = useSession()
  const [user, setUser] = useState(initial)
  const [checking, setChecking] = useState(false)
  const [meOk, setMeOk] = useState(true)
  const [msg, setMsg] = useState<string | null>(null)

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

  const scene = hasSessionScene(session)

  return (
    <div className="m-page">
      <header className="m-page__head">
        <h1 className="m-page__title">我的</h1>
        <p className="m-muted">账号 · 续现场</p>
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
        {msg ? <p className={meOk ? 'm-ok' : 'm-error'}>{msg}</p> : null}
      </section>

      <div className="m-me__actions">
        {scene ? (
          <Link className="m-btn m-btn--primary m-btn--link" to={pathFromSession(session)}>
            续当前现场
          </Link>
        ) : null}
        <Link className="m-btn m-btn--ghost m-btn--link" to="/resume">
          现场字段 / PUT
        </Link>
        <button className="m-btn m-btn--ghost" type="button" onClick={recheck} disabled={checking}>
          {checking ? '检查中…' : '再打 /auth/me'}
        </button>
        <button
          className="m-btn m-btn--ghost"
          type="button"
          onClick={() =>
            void saveSession({
              wsId: session.wsId || 'ws_mobile_demo',
              view: session.view || 'home',
            })
          }
        >
          PUT 写一条现场
        </button>
        <button className="m-btn m-btn--danger" type="button" onClick={doLogout}>
          退出登录
        </button>
      </div>
    </div>
  )
}
