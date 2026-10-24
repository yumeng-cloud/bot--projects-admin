import type { ReactNode } from 'react'

type Props = {
  label: string
  children: ReactNode
}

/** 单行 label | value — 对标 ISSUE_DETAIL_LAYOUT §4 */
export function IssuePropertyRow({ label, children }: Props) {
  return (
    <div className="acb-property-row">
      <span className="acb-property-label acb-property-row__label">{label}</span>
      <div className="acb-property-value acb-property-row__value">{children}</div>
    </div>
  )
}
