import { useMemo } from 'react'
import { Viewer } from '@bytemd/react'
import gfm from '@bytemd/plugin-gfm'
import gfmZhHans from '@bytemd/plugin-gfm/locales/zh_Hans.json'
import 'bytemd/dist/index.css'

type Props = {
  value: string
}

/** Wiki 正文阅读 — ByteMD Viewer，class 挂 acb-wiki-md */
export function WikiBodyViewer({ value }: Props) {
  const plugins = useMemo(() => [gfm({ locale: gfmZhHans })], [])
  const empty = !value?.trim()

  return (
    <div className={empty ? 'acb-wiki-md is-empty is-reading' : 'acb-wiki-md is-reading'}>
      {empty ? (
        <div className="acb-wiki-md-empty-hint">还没有正文…</div>
      ) : (
        <Viewer value={value} plugins={plugins} />
      )}
    </div>
  )
}
