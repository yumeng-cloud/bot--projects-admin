import { useMemo, useState } from 'react'
import type { FocusEvent } from 'react'
import { Button } from 'antd'
import { EditOutlined } from '@ant-design/icons'
import { Editor } from '@bytemd/react'
import gfm from '@bytemd/plugin-gfm'
import zhHans from 'bytemd/locales/zh_Hans.json'
import gfmZhHans from '@bytemd/plugin-gfm/locales/zh_Hans.json'
import 'bytemd/dist/index.css'
import { WikiBodyViewer } from '../wiki/WikiBodyViewer'

type Props = {
  value: string
  onChange: (markdown: string) => void
  /** 离开编辑器（含 tab 切换后真正失焦）时触发，用于 blur 自动保存 */
  onBlur?: () => void
  placeholder?: string
}

/**
 * Issue 描述 — 默认预览（对齐 Wiki Viewer）；点编辑才进 ByteMD edit；blur/保存回预览
 */
export function IssueDescriptionEditor({
  value,
  onChange,
  onBlur,
  placeholder = '添加描述…',
}: Props) {
  const [editing, setEditing] = useState(false)
  const plugins = useMemo(() => [gfm({ locale: gfmZhHans })], [])
  const empty = !value?.trim()

  const leaveEdit = () => {
    setEditing(false)
    onBlur?.()
  }

  const handleBlur = (e: FocusEvent<HTMLDivElement>) => {
    const next = e.relatedTarget as Node | null
    if (next && e.currentTarget.contains(next)) return
    leaveEdit()
  }

  if (!editing) {
    return (
      <div className={empty ? 'acb-issue-md is-empty is-preview' : 'acb-issue-md is-preview'}>
        <div className="acb-issue-desc-preview-bar">
          <Button
            type="text"
            size="small"
            className="acb-issue-desc-edit-btn"
            icon={<EditOutlined />}
            onClick={() => setEditing(true)}
          >
            编辑
          </Button>
        </div>
        {empty ? (
          <button
            type="button"
            className="acb-issue-desc-empty"
            onClick={() => setEditing(true)}
          >
            {placeholder}
          </button>
        ) : (
          <WikiBodyViewer
            value={value}
            emptyText={placeholder}
            className="acb-issue-md-viewer"
          />
        )}
      </div>
    )
  }

  return (
    <div
      className={empty ? 'acb-issue-md is-empty' : 'acb-issue-md'}
      onBlur={handleBlur}
    >
      <Editor
        value={value || ''}
        plugins={plugins}
        locale={zhHans}
        mode="tab"
        placeholder={placeholder}
        onChange={onChange}
      />
    </div>
  )
}
