import { useState, type FormEvent } from 'react'
import { ApiError, login } from '@acb/api-client'

type Props = { onAuthed: () => void | Promise<void> }

export default function LoginPage({ onAuthed }: Props) {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const onSubmit = async (e: FormEvent) => {
    e.preventDefault()
    setError(null)
    setLoading(true)
    try {
      await login({ email: email.trim(), password })
      await onAuthed()
    } catch (err) {
      const msg =
        err instanceof ApiError
          ? err.message
          : err instanceof Error
            ? err.message
            : '登录失败'
      setError(msg)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="m-shell m-login">
      <header className="m-login__brand">
        <h1 className="m-login__title">育蒙协作</h1>
        <p className="m-login__sub">移动端 · 同账号续现场</p>
      </header>

      <form className="m-card m-login__form" onSubmit={onSubmit} noValidate>
        <label className="m-field">
          <span className="m-field__label">邮箱</span>
          <input
            className="m-input"
            type="email"
            name="email"
            autoComplete="username"
            inputMode="email"
            placeholder="请输入邮箱"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            required
          />
        </label>

        <label className="m-field">
          <span className="m-field__label">密码</span>
          <input
            className="m-input"
            type="password"
            name="password"
            autoComplete="current-password"
            placeholder="请输入密码"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            required
          />
        </label>

        {error ? <p className="m-error" role="alert">{error}</p> : null}

        <button className="m-btn m-btn--primary" type="submit" disabled={loading}>
          {loading ? '登录中…' : '登录'}
        </button>
      </form>
    </div>
  )
}
