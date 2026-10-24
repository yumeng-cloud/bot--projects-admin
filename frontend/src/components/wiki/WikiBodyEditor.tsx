import { useMemo } from 'react'
import type { FocusEvent } from 'react'
import { Editor } from '@bytemd/react'
import gfm from '@bytemd/plugin-gfm'
import zhHans from 'bytemd/locales/zh_Hans.json'
import gfmZhHans from '@bytemd/plugin-gfm/locales/zh_Hans.json'
import 'bytemd/dist/index.css'

type Props = {
  value: string
  onChange: (markdown: string) => void
  onBlur?: () => void
  placeholder?: string
}

/** Wiki 正文编辑 — ByteMD，class 挂 acb-wiki-md 供 UI 视觉主题 */
export function WikiBodyEditor({
  value,
  onChange,
  onBlur,
  placeholder = '开始写作…',
}: Props) {
  const plugins = useMemo(() => [gfm({ locale: gfmZhHans })], [])
  const empty = !value?.trim()

  const handleBlur = (e: FocusEvent<HTMLDivElement>) => {
    const next = e.relatedTarget as Node | null
    if (next && e.currentTarget.contains(next)) return
    onBlur?.()
  }

  return (
    <div
      className={empty ? 'acb-wiki-md is-empty is-editing' : 'acb-wiki-md is-editing'}
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
