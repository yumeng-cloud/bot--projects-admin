import { useMemo } from 'react'
import { Viewer } from '@bytemd/react'
import gfm from '@bytemd/plugin-gfm'
import gfmZhHans from '@bytemd/plugin-gfm/locales/zh_Hans.json'
import 'bytemd/dist/index.css'

type Props = {
  value: string
  /** 空正文占位；Issue 等场景可覆写 */
  emptyText?: string
  className?: string
}

/** Wiki / Issue 共用 Markdown 阅读 — ByteMD Viewer */
export function WikiBodyViewer({
  value,
  emptyText = '还没有正文…',
  className,
}: Props) {
  const plugins = useMemo(() => [gfm({ locale: gfmZhHans })], [])
  const empty = !value?.trim()
  const rootClass = ['acb-wiki-body', 'acb-wiki-md', className].filter(Boolean).join(' ')

  if (empty) {
    return (
      <div className={`${rootClass} is-empty`.trim()}>{emptyText}</div>
    )
  }

  return (
    <div className={rootClass}>
      <Viewer value={value} plugins={plugins} />
    </div>
  )
}
