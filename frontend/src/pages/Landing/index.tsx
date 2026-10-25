/**
 * LANDING / SITE_LANDING — figure-thickened A–F + A2.
 * Visual: docs/SITE_LANDING_VISUAL_BASELINE.md @bc215de §4 / §4.1 / §5 / §6
 */
import { Link } from 'react-router-dom'
import { Button } from 'antd'
import './landing.css'
import {
  MockBoard,
  MockSeatPanel,
  MockMulti,
  MockFlow,
  MockCompare,
  MockKey,
} from './mocks'

const STEPS = [
  { n: '1', title: '建班底', body: '按角色开席，人与 Agent 各就各位。' },
  { n: '2', title: '各机认领', body: '不同电脑认领不同席位角色，并行干活。' },
  { n: '3', title: '看板推进', body: '共享 Issue / Wiki / 上下文，进度同场可见。' },
  { n: '4', title: '统一调度', body: '干完收束验收；拍板仍在人。' },
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
        {/* A Hero — left copy · right V-Board large frame */}
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
          <div className="acb-landing-hero-viz">
            <MockBoard />
          </div>
        </section>

        {/* A2 为什么做 + chat vs seat compare */}
        <section
          id="acb-landing-why"
          className="acb-landing-block acb-landing-why"
          aria-labelledby="acb-landing-why-title"
        >
          <div className="acb-landing-split">
            <div className="acb-landing-split-copy">
              <h2 id="acb-landing-why-title" className="acb-landing-section-title">
                为什么做
              </h2>
              <p className="acb-landing-section-body">
                聊天窗管不住一支队——要占席、席间交付与验收，而不是多开几个对话。
              </p>
            </div>
            <div className="acb-landing-split-viz">
              <MockCompare />
            </div>
          </div>
        </section>

        {/* B 价值 — 左右图文交替（禁三等宽纯字卡墙） */}
        <section
          id="acb-landing-values"
          className="acb-landing-block acb-landing-values"
          aria-label="产品价值"
        >
          {/* 一支队：左文右图 V-Board */}
          <article className="acb-landing-split acb-landing-value-row">
            <div className="acb-landing-split-copy">
              <h2 className="acb-landing-section-title">一支队</h2>
              <p className="acb-landing-section-body">
                人与 Agent 同场协作，共享看板与上下文。不是多开聊天：各司其席，像一支真团队在场。
              </p>
            </div>
            <div className="acb-landing-split-viz">
              <MockBoard compact />
            </div>
          </article>

          {/* 多机：左图右文 V-Multi（交替） */}
          <article className="acb-landing-split is-flip acb-landing-value-row">
            <div className="acb-landing-split-copy">
              <h2 className="acb-landing-section-title">多机分工</h2>
              <p className="acb-landing-section-body">
                不同电脑认领不同角色干活，干完统一调度。换机等于分布式工位，不是把同一会话挪走。
              </p>
            </div>
            <div className="acb-landing-split-viz">
              <MockMulti />
            </div>
          </article>

          {/* 席位：左文右席位面板 */}
          <article className="acb-landing-split acb-landing-value-row">
            <div className="acb-landing-split-copy">
              <h2 className="acb-landing-section-title">席位 · 按需入场 · 模型档</h2>
              <p className="acb-landing-section-body">
                班底按席；钥匙按需；决策审查强模型，写码轻量。档位后续可挂，结构已按席留好。
              </p>
            </div>
            <div className="acb-landing-split-viz">
              <MockSeatPanel />
            </div>
          </article>
        </section>

        {/* C 怎么干活 + V-Flow */}
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
          <div className="acb-landing-flow-wrap">
            <MockFlow />
          </div>
        </section>

        {/* D 分布式换机 — 左文右 V-Multi */}
        <section
          className="acb-landing-block acb-landing-dist"
          aria-labelledby="acb-landing-dist-title"
        >
          <div className="acb-landing-split">
            <div className="acb-landing-split-copy">
              <h2 id="acb-landing-dist-title" className="acb-landing-section-title">
                多机就是一支队的工位
              </h2>
              <p className="acb-landing-section-body">
                不是把同一会话挪到另一台电脑，而是让不同机器上的 Agent
                认领不同角色，活在同一套调度下齐步走。
              </p>
            </div>
            <div className="acb-landing-split-viz">
              <MockMulti />
            </div>
          </div>
        </section>

        {/* E 席位与入场 — 左文右 V-Seat + V-Key */}
        <section
          className="acb-landing-block acb-landing-seats"
          aria-labelledby="acb-landing-seats-title"
        >
          <div className="acb-landing-split is-flip">
            <div className="acb-landing-split-copy">
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
            </div>
            <div className="acb-landing-split-viz acb-landing-seat-stack">
              <MockSeatPanel />
              <MockKey />
            </div>
          </div>
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
