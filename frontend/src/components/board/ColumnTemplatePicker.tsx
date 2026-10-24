import { useMemo, useState } from 'react'
import { Button, Input, Modal, Popover, Switch, Tag, Typography, message } from 'antd'
import { MoreOutlined } from '@ant-design/icons'
import type { IssueStatus } from '../../types/issue'
import { BOARD_COLUMNS, STATUS_LABELS } from '../../types/issue'
import {
  BUILTIN_TEMPLATES,
  columnLabel,
  defaultPrefs,
  prefsFromTemplate,
  type BoardColumnDef,
  type BoardColumnLayout,
} from './boardTemplates'
import './board.css'

export type ColumnTemplatePickerProps = {
  prefs: BoardColumnLayout
  onApply: (prefs: BoardColumnLayout) => void | Promise<void>
  /** Compact trigger for empty-board CTA */
  variant?: 'icon' | 'cta'
}

type PendingApply = {
  next: BoardColumnLayout
  risk: string
}

function MiniPreview({ columns }: { columns: BoardColumnDef[] }) {
  const visible = columns.filter((c) => !c.hidden)
  return (
    <div className="acb-tpl-preview" aria-hidden>
      {visible.map((c) => (
        <div key={c.status} className="acb-tpl-preview-col">
          <span className={`acb-col-dot status-${c.status}`} />
          <span className="acb-tpl-preview-bar" />
        </div>
      ))}
    </div>
  )
}

export function ColumnTemplatePicker({
  prefs,
  onApply,
  variant = 'icon',
}: ColumnTemplatePickerProps) {
  const [open, setOpen] = useState(false)
  const [customizing, setCustomizing] = useState(false)
  const [draft, setDraft] = useState<BoardColumnDef[]>([])
  const [pending, setPending] = useState<PendingApply | null>(null)
  const [applying, setApplying] = useState(false)

  const currentName = useMemo(() => {
    if (prefs.templateId === 'custom') return '自定义'
    return BUILTIN_TEMPLATES.find((t) => t.id === prefs.templateId)?.name ?? '自定义'
  }, [prefs.templateId])

  const startCustom = () => {
    // Fork current visible order; include hidden BOARD_COLUMNS not in prefs
    const byStatus = new Map(prefs.columns.map((c) => [c.status, { ...c }]))
    const cols: BoardColumnDef[] = []
    for (const c of prefs.columns) cols.push({ ...c })
    for (const s of BOARD_COLUMNS) {
      if (!byStatus.has(s)) cols.push({ status: s, hidden: true })
    }
    setDraft(cols)
    setCustomizing(true)
  }

  const requestApply = (next: BoardColumnLayout, risk: string) => {
    setPending({ next, risk })
  }

  const confirmApply = async () => {
    if (!pending || applying) return
    setApplying(true)
    try {
      await onApply(pending.next)
      message.success('已应用列模板')
      setPending(null)
      setCustomizing(false)
      setOpen(false)
    } catch {
      // toasted by client
    } finally {
      setApplying(false)
    }
  }

  const applyBuiltin = (templateId: string) => {
    requestApply(
      prefsFromTemplate(templateId),
      '将替换当前项目的列顺序、显隐与显示名偏好。',
    )
  }

  const restoreDefault = () => {
    requestApply(defaultPrefs(), '将恢复为默认五列，并丢弃当前自定义偏好。')
  }

  const applyCustom = () => {
    const visible = draft.filter((c) => !c.hidden)
    if (visible.length === 0) {
      message.warning('至少保留一列可见')
      return
    }
    requestApply(
      { templateId: 'custom', columns: draft.map((c) => ({ ...c })) },
      '将替换当前项目的列顺序、显隐与显示名偏好。',
    )
  }

  const updateDraft = (status: IssueStatus, patch: Partial<BoardColumnDef>) => {
    setDraft((prev) => prev.map((c) => (c.status === status ? { ...c, ...patch } : c)))
  }

  const panel = (
    <div className="acb-tpl-panel">
      {!customizing ? (
        <>
          <div className="acb-tpl-panel-head">
            <Typography.Text strong style={{ fontSize: 13 }}>
              列模板
            </Typography.Text>
            <Typography.Text type="secondary" style={{ fontSize: 11 }}>
              当前：{currentName}
            </Typography.Text>
          </div>
          <div className="acb-tpl-list">
            {BUILTIN_TEMPLATES.map((t) => (
              <button
                key={t.id}
                type="button"
                className={`acb-tpl-item${prefs.templateId === t.id ? ' is-active' : ''}`}
                onClick={() => applyBuiltin(t.id)}
              >
                <div className="acb-tpl-item-main">
                  <span className="acb-tpl-item-name">{t.name}</span>
                  <Tag className="acb-tpl-sys-tag">系统</Tag>
                </div>
                <Typography.Text type="secondary" style={{ fontSize: 11 }}>
                  {t.description}
                </Typography.Text>
                <MiniPreview columns={t.columns} />
              </button>
            ))}
          </div>
          <div className="acb-tpl-actions">
            <Button size="small" type="link" onClick={startCustom}>
              自定义列…
            </Button>
            <Button size="small" type="link" onClick={restoreDefault}>
              恢复默认
            </Button>
          </div>
        </>
      ) : (
        <>
          <div className="acb-tpl-panel-head">
            <Typography.Text strong style={{ fontSize: 13 }}>
              自定义列
            </Typography.Text>
            <Button
              size="small"
              type="text"
              onClick={() => setCustomizing(false)}
              style={{ height: 24, fontSize: 12 }}
            >
              返回模板
            </Button>
          </div>
          <div className="acb-tpl-custom-list">
            {draft.map((c) => (
              <div key={c.status} className="acb-tpl-custom-row">
                <span className={`acb-col-dot status-${c.status}`} aria-hidden />
                <Input
                  size="small"
                  className="acb-tpl-custom-input"
                  value={c.label ?? ''}
                  placeholder={STATUS_LABELS[c.status]}
                  onChange={(e) =>
                    updateDraft(c.status, {
                      label: e.target.value.trim() ? e.target.value : undefined,
                    })
                  }
                />
                <span className="acb-tpl-custom-status">{c.status}</span>
                <Switch
                  size="small"
                  checked={!c.hidden}
                  onChange={(on) => updateDraft(c.status, { hidden: !on })}
                  title={c.hidden ? '显示' : '隐藏'}
                />
              </div>
            ))}
          </div>
          <MiniPreview columns={draft} />
          <div className="acb-tpl-actions">
            <Button size="small" onClick={() => setCustomizing(false)}>
              取消
            </Button>
            <Button size="small" type="primary" onClick={applyCustom}>
              应用
            </Button>
          </div>
        </>
      )}
    </div>
  )

  const trigger =
    variant === 'cta' ? (
      <Button size="small">选择列模板</Button>
    ) : (
      <Button
        type="text"
        size="small"
        icon={<MoreOutlined />}
        aria-label="列模板与更多"
        title="列模板"
      />
    )

  return (
    <>
      <Popover
        open={open}
        onOpenChange={(v) => {
          setOpen(v)
          if (!v) setCustomizing(false)
        }}
        trigger="click"
        placement="bottomRight"
        content={panel}
        arrow={false}
        overlayClassName="acb-tpl-popover"
      >
        {trigger}
      </Popover>
      <Modal
        title="应用列模板？"
        open={!!pending}
        onCancel={() => (!applying ? setPending(null) : undefined)}
        onOk={() => void confirmApply()}
        okText="替换并应用"
        cancelText="取消"
        confirmLoading={applying}
        centered
        width={400}
        destroyOnHidden
      >
        {pending ? (
          <div>
            <p style={{ margin: '0 0 8px', fontSize: 13, color: 'var(--acb-text)' }}>
              {pending.risk}
            </p>
            <MiniPreview columns={pending.next.columns} />
            <Typography.Text type="secondary" style={{ fontSize: 12 }}>
              预览：
              {pending.next.columns
                .filter((c) => !c.hidden)
                .map((c) => columnLabel(c))
                .join(' · ')}
            </Typography.Text>
          </div>
        ) : null}
      </Modal>
    </>
  )
}
