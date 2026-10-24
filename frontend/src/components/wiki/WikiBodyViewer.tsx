import { useMemo } from 'react'
import { Viewer } from '@bytemd/react'
import gfm from '@bytemd/plugin-gfm'
import gfmZhHans from '@bytemd/plugin-gfm/locales/zh_Hans.json'
import 'bytemd/dist/index.css'

type Props = {
  value: string
}

/** Wiki 正文阅读 — ByteMD Viewer；挂 acb-wiki-md */
export function WikiBodyViewer({ value }: Props) {
  const plugins = useMemo(() => [gfm({ locale: gfmZhHans })], [])
  const empty = !value?.trim()

  if (empty) {
    return (
      <div className="acb-wiki-body is-empty acb-wiki-md is-empty">还没有正文…</div>
    )
  }

  return (
    <div className="acb-wiki-body acb-wiki-md">
      <Viewer value={value} plugins={plugins} />
    </div>
  )
}
