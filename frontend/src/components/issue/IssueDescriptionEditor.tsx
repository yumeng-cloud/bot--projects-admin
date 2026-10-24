import { useMemo } from 'react'
import { Editor } from '@bytemd/react'
import gfm from '@bytemd/plugin-gfm'
import zhHans from 'bytemd/locales/zh_Hans.json'
import gfmZhHans from '@bytemd/plugin-gfm/locales/zh_Hans.json'
import 'bytemd/dist/index.css'

type Props = {
  value: string
  onChange: (markdown: string) => void
  placeholder?: string
}

/**
 * ByteMD — Issue 描述主舞台（ISSUE_DETAIL_LAYOUT §11）
 * - mode=tab（编辑 | 预览）
 * - locale=zh_Hans
 * - value 为 Markdown 字符串
 * - 空描述：编辑态 placeholder + 预览区「还没有描述…」
 */
export function IssueDescriptionEditor({
  value,
  onChange,
  placeholder = '添加描述…',
}: Props) {
  const plugins = useMemo(() => [gfm({ locale: gfmZhHans })], [])
  const empty = !value?.trim()

  return (
    <div className={empty ? 'acb-issue-md is-empty' : 'acb-issue-md'}>
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
