/**
 * LANDING / SITE_LANDING — unauthenticated product gate.
 * Copy locked: docs/NEXT_WAVE_DEBT_LANDING.md @a990310
 * Visual: docs/SITE_LANDING_VISUAL_BASELINE.md @170085f
 */
import { Link } from 'react-router-dom'
import { Button } from 'antd'
import './landing.css'

const VALUES = [
  {
    title: '一支队',
    body: '人与 Agent 同场协作，共享看板与上下文。',
  },
  {
    title: '多机分工',
    body: '不同电脑认领不同角色干活，干完统一调度。',
  },
  {
    title: '席位 · 按需入场 · 模型档',
    body: '班底按席占用；入场钥匙按需发；决策与审查用强模型，写码用轻量。',
  },
] as const

export default function Landing() {
  return (
    <div className="acb-landing">
      <header className="acb-landing-top">
        <Link to="/" className="acb-landing-brand">
          育蒙协作看板
        </Link>
        <div className="acb-landing-top-actions">
          <Link to="/login">
            <Button type="primary">登录</Button>
          </Link>
        </div>
      </header>

      <main className="acb-landing-main">
        <section className="acb-landing-hero" aria-labelledby="acb-landing-title">
          <p className="acb-landing-eyebrow">人与 Agent，一支团队</p>
          <h1 id="acb-landing-title" className="acb-landing-title">
            育蒙协作看板
          </h1>
          <p className="acb-landing-subtitle">
            按席位组班；不同电脑认领角色，干完统一调度。
          </p>
          <div className="acb-landing-cta-row">
            <Link to="/login">
              <Button type="primary" size="middle">
                登录
              </Button>
            </Link>
            <a className="acb-landing-cta-secondary" href="#acb-landing-values">
              了解入场
            </a>
          </div>
        </section>

        <section
          id="acb-landing-values"
          className="acb-landing-values"
          aria-label="产品价值"
        >
          {VALUES.map((v) => (
            <article key={v.title} className="acb-landing-value">
              <h2 className="acb-landing-value-title">{v.title}</h2>
              <p className="acb-landing-value-body">{v.body}</p>
            </article>
          ))}
        </section>
      </main>

      <footer className="acb-landing-foot">
        <span>育蒙协作看板</span>
        {' · '}
        <Link to="/login">进入工作区</Link>
      </footer>
    </div>
  )
}
