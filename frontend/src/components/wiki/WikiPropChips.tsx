import { useState } from 'react'
import { Popover, Select } from 'antd'

export type WikiPropOption = { label: string; value: string }

type ChipKind = 'project' | 'issue' | 'parent'

export type ChipDef = {
  kind: ChipKind
  label: string
  value?: string | null
  display: string
  emptyText: string
  options: WikiPropOption[]
  onChange: (value: string | null) => void
  allowClear?: boolean
}

type Props = {
  chips: ChipDef[]
}

function PropChip({
  label,
  display,
  emptyText,
  options,
  value,
  onChange,
  allowClear = true,
}: Omit<ChipDef, 'kind'>) {
  const [open, setOpen] = useState(false)
  const hasValue = !!value

  const content = (
    <div className="acb-wiki-prop-chip-popover" style={{ width: 280 }}>
      <Select
        autoFocus
        allowClear={allowClear}
        showSearch
        optionFilterProp="label"
        placeholder={`选择${label}`}
        style={{ width: '100%' }}
        size="middle"
        value={value || undefined}
        options={options}
        defaultOpen
        listHeight={240}
        onChange={(v) => {
          onChange((v as string) ?? null)
          setOpen(false)
        }}
      />
    </div>
  )

  return (
    <Popover
      trigger="click"
      open={open}
      onOpenChange={setOpen}
      placement="bottomLeft"
      content={content}
      overlayClassName="acb-wiki-prop-chip-overlay"
      destroyTooltipOnHide
    >
      <button
        type="button"
        className={
          hasValue
            ? 'acb-wiki-prop-chip acb-wiki-prop-chip--set'
            : 'acb-wiki-prop-chip acb-wiki-prop-chip--empty'
        }
        aria-label={`${label}：${hasValue ? display : emptyText}`}
      >
        <span className="acb-wiki-props-label">{label}</span>
        <span className="acb-wiki-prop-chip-value">{hasValue ? display : emptyText}</span>
      </button>
    </Popover>
  )
}

/** P1-3 Property 芯片轨：默认不露大 Select，点击芯片再弹出选择 */
export function WikiPropChips({ chips }: Props) {
  return (
    <div className="acb-wiki-props acb-wiki-prop-rail" role="group" aria-label="页面属性">
      {chips.map((c) => (
        <PropChip key={c.kind} {...c} />
      ))}
    </div>
  )
}
