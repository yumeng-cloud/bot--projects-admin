/**
 * LANDING / SITE_LANDING — unauthenticated product gate (thickened A–F + A2).
 * Copy: docs/SITE_LANDING_VISUAL_BASELINE.md @710df9f + product lock.
 * Visual: docs/SITE_LANDING_VISUAL_BASELINE.md
 */
import { Link } from 'react-router-dom'
import { Button } from 'antd'
import './landing.css'

const VALUES = [
  {
    title: '一支队',
    body: '不是多开聊天：人与 Agent 同场共享看板与上下文，各司其席——像一支真团队在场，而不是几个孤立对话窗。',
  },
  {
    title: '多机分工',
    body: '不是换机续聊：不同电脑认领不同角色干活，干完统一调度指挥。换机等于分布式工位，不是把同一会话挪走。',
  },
  {
    title: '席位 · 按需入场 · 模型档',
    body: '班底按席占用；入场钥匙按需发（有 K 用、无则生、替换才顶）。决策与审查用强模型，写码用轻量——档位后续可挂，结构已按席留好。',
  },
] as const

const STEPS = [
  {
    n: '1',
    title: '建班底',
    body: '按角色开席，人与 Agent 各就各位。',
  },
  {
    n: '2',
    title: '各机认领',
    body: '不同电脑认领不同席位角色，并行干活。',
  },
  {
    n: '3',
    title: '看板推进',
    body: '共享 Issue / Wiki / 上下文，进度同场可见。',
  },
  {
    n: '4',
    title: '统一调度',
    body: '干完收束验收；拍板仍在人。',
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
        {/* A Hero */}
        <section className="acb-landing-hero" aria-labelledby="acb-landing-title">
          <div className="acb-landing-hero-copy">
            <p className="acb-landing-eyebrow">人与 Agent，一支团队</p>
            <h1 id="acb-landing-title" className="acb-landing-title">
              育蒙协作看板
            </h1>
            <p className="acb-landing-subtitle">
              把人和 Agent 编进同一支班底，按席干活、多机分工、统一调度。
            </p>
            <div className="acb-landing-cta-row">
              <Link to="/login">
                <Button type="primary" size="middle">
                  登录
                </Button>
              </Link>
              <a className="acb-landing-cta-secondary" href="#acb-landing-why">
                了解入场
              </a>
            </div>
          </div>
          <div className="acb-landing-hero-viz" aria-hidden="true">
            <div className="acb-landing-seat-bar">
              <span className="acb-landing-seat-bar-label">班底席位</span>
              <div className="acb-landing-seat-chips">
                <span className="acb-landing-seat-chip is-human">人 · 主持</span>
                <span className="acb-landing-seat-chip is-agent">Agent · 审图</span>
                <span className="acb-landing-seat-chip is-agent">Agent · 写码</span>
                <span className="acb-landing-seat-chip is-empty">空席</span>
              </div>
            </div>
          </div>
        </section>

        {/* A2 为什么做 */}
        <section
          id="acb-landing-why"
          className="acb-landing-block acb-landing-why"
          aria-labelledby="acb-landing-why-title"
        >
          <h2 id="acb-landing-why-title" className="acb-landing-section-title">
            为什么做
          </h2>
          <p className="acb-landing-section-body">
            单人聊天窗管不住多角色协作；需要像真团队一样占席、席间交付与验收。
          </p>
        </section>

        {/* B 价值三卡 */}
        <section
          id="acb-landing-values"
          className="acb-landing-block acb-landing-values"
          aria-label="产品价值"
        >
          {VALUES.map((v) => (
            <article key={v.title} className="acb-landing-value">
              <h2 className="acb-landing-value-title">{v.title}</h2>
              <p className="acb-landing-value-body">{v.body}</p>
            </article>
          ))}
        </section>

        {/* C 怎么干活 */}
        <section
          className="acb-landing-block acb-landing-steps"
          aria-labelledby="acb-landing-steps-title"
        >
          <h2 id="acb-landing-steps-title" className="acb-landing-section-title">
            怎么干活
          </h2>
          <ol className="acb-landing-steps-list">
            {STEPS.map((s) => (
              <li key={s.n} className="acb-landing-step">
                <span className="acb-landing-step-n" aria-hidden="true">
                  {s.n}
                </span>
                <div className="acb-landing-step-copy">
                  <h3 className="acb-landing-step-title">{s.title}</h3>
                  <p className="acb-landing-step-body">{s.body}</p>
                </div>
              </li>
            ))}
          </ol>
        </section>

        {/* D 分布式换机 */}
        <section
          className="acb-landing-block acb-landing-dist"
          aria-labelledby="acb-landing-dist-title"
        >
          <h2 id="acb-landing-dist-title" className="acb-landing-section-title">
            多机就是一支队的工位
          </h2>
          <p className="acb-landing-section-body">
            不是把同一会话挪到另一台电脑，而是让不同机器上的 Agent
            认领不同角色，活在同一套调度下齐步走。
          </p>
          <div className="acb-landing-dist-viz" aria-hidden="true">
            <div className="acb-landing-machine">
              <span className="acb-landing-machine-name">机 A</span>
              <span className="acb-landing-machine-role">审图席</span>
            </div>
            <div className="acb-landing-machine">
              <span className="acb-landing-machine-name">机 B</span>
              <span className="acb-landing-machine-role">写码席</span>
            </div>
            <div className="acb-landing-machine">
              <span className="acb-landing-machine-name">机 C</span>
              <span className="acb-landing-machine-role">主持席</span>
            </div>
            <div className="acb-landing-dist-hub">统一调度</div>
          </div>
        </section>

        {/* E 席位与入场 */}
        <section
          className="acb-landing-block acb-landing-seats"
          aria-labelledby="acb-landing-seats-title"
        >
          <h2 id="acb-landing-seats-title" className="acb-landing-section-title">
            钥匙按需，不堆墙
          </h2>
          <p className="acb-landing-section-body">
            席位有人占用才算在席；入场钥匙有则用、无则生、替换才顶。决策与审查用强模型，写码席用轻量——档位后续可挂，结构已按席留好。
          </p>
          <p className="acb-landing-how-start">
            <strong>怎么开始：</strong>
            登录进入工作区；按席复制入场链接给 Bot；不要默认堆一墙码。
          </p>
        </section>

        {/* F 底 CTA */}
        <section
          className="acb-landing-block acb-landing-bottom-cta"
          aria-labelledby="acb-landing-close-title"
        >
          <p id="acb-landing-close-title" className="acb-landing-close">
            把 Agent 当同事组班，而不是当插件。
          </p>
          <Link to="/login">
            <Button type="primary" size="middle">
              登录
            </Button>
          </Link>
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
