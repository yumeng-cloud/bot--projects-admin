/**
 * Silent product UI mocks for SITE_LANDING (CSS-only, fake data).
 * Labels: V-Board / V-Seat / V-Multi / V-Flow / V-Compare / V-Key
 */
import './landing-mocks.css'

/** V-Board — 3-col kanban + fake issue cards + human/agent dots */
export function MockBoard({ compact = false }: { compact?: boolean }) {
  return (
    <div
      className={`acb-mock acb-mock-board${compact ? ' is-compact' : ''}`}
      aria-hidden="true"
    >
      <div className="acb-mock-board-head">
        <span className="acb-mock-board-title">本周迭代</span>
        <span className="acb-mock-board-meta">人 · Agent 同场</span>
      </div>
      <div className="acb-mock-board-cols">
        <div className="acb-mock-col">
          <div className="acb-mock-col-head">待办</div>
          <div className="acb-mock-card">
            <div className="acb-mock-card-title">整理席位入场钥</div>
            <div className="acb-mock-card-foot">
              <span className="acb-mock-dot is-human" title="人" />
              <span className="acb-mock-tag">主持</span>
            </div>
          </div>
          <div className="acb-mock-card">
            <div className="acb-mock-card-title">Wiki 初稿</div>
            <div className="acb-mock-card-foot">
              <span className="acb-mock-dot is-agent" title="Agent" />
              <span className="acb-mock-tag">写码</span>
            </div>
          </div>
        </div>
        <div className="acb-mock-col">
          <div className="acb-mock-col-head">进行</div>
          <div className="acb-mock-card is-active">
            <div className="acb-mock-card-title">落地页图文加厚</div>
            <div className="acb-mock-card-foot">
              <span className="acb-mock-dot is-agent" title="Agent" />
              <span className="acb-mock-dot is-human" title="人" />
              <span className="acb-mock-tag">审图</span>
            </div>
          </div>
          <div className="acb-mock-card">
            <div className="acb-mock-card-title">Issue 验收清单</div>
            <div className="acb-mock-card-foot">
              <span className="acb-mock-dot is-human" title="人" />
              <span className="acb-mock-tag">主持</span>
            </div>
          </div>
        </div>
        <div className="acb-mock-col">
          <div className="acb-mock-col-head">完成</div>
          <div className="acb-mock-card is-done">
            <div className="acb-mock-card-title">登录门同肤</div>
            <div className="acb-mock-card-foot">
              <span className="acb-mock-dot is-agent" title="Agent" />
              <span className="acb-mock-tag">写码</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}

/** V-Seat — seat roster panel */
export function MockSeatPanel() {
  const rows = [
    { role: '主持', who: '人 · 阿蒙', state: '在席', kind: 'human' as const },
    { role: '审图', who: 'Agent · Soft', state: '在席', kind: 'agent' as const },
    { role: '写码', who: 'Agent · FE', state: '在席', kind: 'agent' as const },
    { role: '审查', who: '—', state: '空席', kind: 'empty' as const },
  ]
  return (
    <div className="acb-mock acb-mock-seat" aria-hidden="true">
      <div className="acb-mock-seat-head">
        <span className="acb-mock-seat-title">席位班底</span>
        <span className="acb-mock-seat-meta">3 在席 · 1 空</span>
      </div>
      <ul className="acb-mock-seat-list">
        {rows.map((r) => (
          <li key={r.role} className={`acb-mock-seat-row is-${r.kind}`}>
            <span className="acb-mock-seat-role">{r.role}</span>
            <span className="acb-mock-seat-who">
              {r.kind !== 'empty' && (
                <span className={`acb-mock-dot is-${r.kind === 'human' ? 'human' : 'agent'}`} />
              )}
              {r.who}
            </span>
            <span className={`acb-mock-seat-state is-${r.kind}`}>{r.state}</span>
          </li>
        ))}
      </ul>
    </div>
  )
}

/** V-Multi — 3 machines → unified dispatch */
export function MockMulti() {
  const machines = [
    { name: '机 A', seat: '审图席', host: '笔记本' },
    { name: '机 B', seat: '写码席', host: '台式' },
    { name: '机 C', seat: '主持席', host: '平板旁路' },
  ]
  return (
    <div className="acb-mock acb-mock-multi" aria-hidden="true">
      <div className="acb-mock-multi-machines">
        {machines.map((m) => (
          <div key={m.name} className="acb-mock-machine">
            <div className="acb-mock-machine-chrome">
              <span className="acb-mock-machine-dot" />
              <span className="acb-mock-machine-dot" />
              <span className="acb-mock-machine-dot" />
            </div>
            <div className="acb-mock-machine-body">
              <span className="acb-mock-machine-name">{m.name}</span>
              <span className="acb-mock-machine-seat">{m.seat}</span>
              <span className="acb-mock-machine-host">{m.host}</span>
            </div>
          </div>
        ))}
      </div>
      <div className="acb-mock-multi-arrows" aria-hidden="true">
        <span />
        <span />
        <span />
      </div>
      <div className="acb-mock-multi-hub">统一调度</div>
    </div>
  )
}

/** V-Flow — 4-step connected strip */
export function MockFlow() {
  const steps = ['建班底', '各机认领', '看板推进', '统一调度']
  return (
    <div className="acb-mock acb-mock-flow" aria-hidden="true">
      {steps.map((label, i) => (
        <div key={label} className="acb-mock-flow-item">
          <div className="acb-mock-flow-node">
            <span className="acb-mock-flow-n">{i + 1}</span>
            <span className="acb-mock-flow-label">{label}</span>
          </div>
          {i < steps.length - 1 && <div className="acb-mock-flow-line" />}
        </div>
      ))}
    </div>
  )
}

/** A2 small compare: chat window vs seat roster */
export function MockCompare() {
  return (
    <div className="acb-mock acb-mock-compare" aria-hidden="true">
      <div className="acb-mock-compare-pane is-chat">
        <div className="acb-mock-compare-label">单聊窗</div>
        <div className="acb-mock-chat-bubble">帮我改一下…</div>
        <div className="acb-mock-chat-bubble is-bot">好的，已打开新对话</div>
        <div className="acb-mock-chat-bubble">再开一个审图…</div>
        <div className="acb-mock-compare-note">多窗 · 无席 · 难验收</div>
      </div>
      <div className="acb-mock-compare-vs">vs</div>
      <div className="acb-mock-compare-pane is-seat">
        <div className="acb-mock-compare-label">席位班底</div>
        <div className="acb-mock-compare-seat-line">
          <span className="acb-mock-dot is-human" /> 主持 · 在席
        </div>
        <div className="acb-mock-compare-seat-line">
          <span className="acb-mock-dot is-agent" /> 审图 · 在席
        </div>
        <div className="acb-mock-compare-seat-line">
          <span className="acb-mock-dot is-agent" /> 写码 · 在席
        </div>
        <div className="acb-mock-compare-note is-ok">同场 · 可交付 · 可验收</div>
      </div>
    </div>
  )
}

/** V-Key — on-demand key row */
export function MockKey() {
  return (
    <div className="acb-mock acb-mock-key" aria-hidden="true">
      <div className="acb-mock-key-row">
        <span className="acb-mock-key-icon" />
        <div className="acb-mock-key-copy">
          <span className="acb-mock-key-title">入场钥 · 审图席</span>
          <span className="acb-mock-key-desc">有则用 · 无则生 · 替换才顶</span>
        </div>
        <span className="acb-mock-key-badge">一席一钥</span>
      </div>
    </div>
  )
}
