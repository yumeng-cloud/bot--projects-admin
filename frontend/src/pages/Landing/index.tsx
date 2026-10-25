/**
 * LANDING / SITE_LANDING — L-R3‡ 真图接线（禁 CSS 假框主路径）
 * Visual: docs/SITE_LANDING_VISUAL_BASELINE.md + docs/NEXT_WAVE_DEBT_LANDING.md
 */
import { Link } from 'react-router-dom'
import { Button } from 'antd'
import './landing.css'

import artHero from './assets/landing-hero-board.png'
import artWhy from './assets/landing-why-compare.png'
import artTeam from './assets/landing-team-board.png'
import artMulti from './assets/landing-multi-machines.png'
import artSeats from './assets/landing-seats-keys.png'
import artFlow from './assets/landing-flow-steps.png'

const STEPS = [
  { n: '1', title: '建班底', body: '按角色开席，人与 Agent 各就各位。' },
  { n: '2', title: '按席发一把入场', body: '有钥则用、无则生；按席复制入场链接给 Bot。' },
  { n: '3', title: '各机认领角色', body: '不同电脑认领不同席位角色，并行干活。' },
  { n: '4', title: '看板里统一调度', body: '共享 Issue / Wiki / 上下文；干完收束，拍板仍在人。' },
] as const

function LandingArt({
  src,
  alt,
  className = '',
}: {
  src: string
  alt: string
  className?: string
}) {
  return (
    <img
      className={`acb-landing-art${className ? ` ${className}` : ''}`}
      src={src}
      alt={alt}
      loading="lazy"
      decoding="async"
    />
  )
}

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
        {/* A Hero — 左文 · 右真图大框 */}
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
            <LandingArt
              src={artHero}
              alt="协作看板示意：人与 Agent 同场推进议题"
              className="is-hero"
            />
          </div>
        </section>

        {/* A2 为什么做 — 文左图右 */}
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
                单人聊天窗管不住多角色协作；需要像真团队一样占席、交接、验收。
              </p>
            </div>
            <div className="acb-landing-split-viz">
              <LandingArt
                src={artWhy}
                alt="单聊窗与席位班底对比：同场可交付可验收"
              />
            </div>
          </div>
        </section>

        {/* B 价值 — 左右图文交替（≥3 is-flip） */}
        <section
          id="acb-landing-values"
          className="acb-landing-block acb-landing-values"
          aria-label="产品价值"
        >
          {/* 一支队：图左文右 */}
          <article className="acb-landing-split is-flip acb-landing-value-row">
            <div className="acb-landing-split-copy">
              <h2 className="acb-landing-section-title">一支队，不是多开聊天</h2>
              <p className="acb-landing-section-body">
                人与 Agent 同场共享看板与上下文，各司其席。
              </p>
            </div>
            <div className="acb-landing-split-viz">
              <LandingArt
                src={artTeam}
                alt="一支队看板：人与 Agent 各司其席"
              />
            </div>
          </article>

          {/* 多机：文左图右 */}
          <article className="acb-landing-split acb-landing-value-row">
            <div className="acb-landing-split-copy">
              <h2 className="acb-landing-section-title">多机分工，不是换机续聊</h2>
              <p className="acb-landing-section-body">
                不同电脑认领不同角色干活，干完统一调度指挥。
              </p>
            </div>
            <div className="acb-landing-split-viz">
              <LandingArt
                src={artMulti}
                alt="多机认领不同席位角色，统一调度"
              />
            </div>
          </article>

          {/* 席位：图左文右 */}
          <article className="acb-landing-split is-flip acb-landing-value-row">
            <div className="acb-landing-split-copy">
              <h2 className="acb-landing-section-title">席位 · 按需入场 · 模型档</h2>
              <p className="acb-landing-section-body">
                班底按席占用；入场钥匙按需发（有则用、无则生、替换才顶）；决策与审查用强模型，写码用轻量。
              </p>
            </div>
            <div className="acb-landing-split-viz">
              <LandingArt
                src={artSeats}
                alt="席位班底与按需入场钥"
              />
            </div>
          </article>
        </section>

        {/* C 怎么干活 + 流程真图 — 文左图右 */}
        <section
          className="acb-landing-block acb-landing-steps"
          aria-labelledby="acb-landing-steps-title"
        >
          <div className="acb-landing-split">
            <div className="acb-landing-split-copy">
              <h2 id="acb-landing-steps-title" className="acb-landing-section-title">
                怎么干活
              </h2>
              <ol className="acb-landing-steps-list is-stacked">
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
            </div>
            <div className="acb-landing-split-viz">
              <LandingArt
                src={artFlow}
                alt="四步流程：建班底、按席入场、各机认领、统一调度"
              />
            </div>
          </div>
        </section>

        {/* D 分布式 — 图左文右（第 3 个 is-flip） */}
        <section
          className="acb-landing-block acb-landing-dist"
          aria-labelledby="acb-landing-dist-title"
        >
          <div className="acb-landing-split is-flip">
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
              <LandingArt
                src={artMulti}
                alt="多机工位：不同机器认领不同角色"
              />
            </div>
          </div>
        </section>

        {/* E 怎么开始 — 文左图右 */}
        <section
          className="acb-landing-block acb-landing-seats"
          aria-labelledby="acb-landing-seats-title"
        >
          <div className="acb-landing-split">
            <div className="acb-landing-split-copy">
              <h2 id="acb-landing-seats-title" className="acb-landing-section-title">
                怎么开始
              </h2>
              <p className="acb-landing-section-body">
                登录进入工作区；按席复制入场链接给 Bot；不要默认堆一墙码。
              </p>
              <p className="acb-landing-how-start">
                <strong>钥匙按需，不堆墙：</strong>
                席位有人占用才算在席；入场钥匙有则用、无则生、替换才顶。决策与审查用强模型，写码席用轻量——档位后续可挂，结构已按席留好。
              </p>
            </div>
            <div className="acb-landing-split-viz">
              <LandingArt
                src={artSeats}
                alt="按席入场：一席一钥，按需发放"
              />
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
