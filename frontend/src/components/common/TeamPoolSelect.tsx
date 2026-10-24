import { Select, Typography } from 'antd'
import type { Team } from '../../api/teams'

type Props = {
  teams: Team[]
  teamId: string | null
  onChange: (teamId: string) => void
  disabled?: boolean
  className?: string
}

/** Lightweight current-team control (T6I I4 / TEAM_SWITCH). Header + page bars. */
export function TeamPoolSelect({ teams, teamId, onChange, disabled, className }: Props) {
  if (teams.length > 1) {
    return (
      <Select
        className={className}
        size="small"
        variant="borderless"
        value={teamId || undefined}
        placeholder="当前队"
        style={{ minWidth: 140 }}
        options={teams.map((t) => ({ value: t.id, label: t.name }))}
        onChange={(v) => onChange(v)}
        disabled={disabled}
        popupMatchSelectWidth={false}
        aria-label="当前队"
      />
    )
  }
  if (teams.length === 1) {
    return (
      <Typography.Text
        className={className}
        type="secondary"
        style={{ fontSize: 12, whiteSpace: 'nowrap' }}
      >
        队 · {teams[0].name}
      </Typography.Text>
    )
  }
  return (
    <Typography.Text
      className={className}
      type="secondary"
      style={{ fontSize: 12, whiteSpace: 'nowrap' }}
      aria-label="当前队"
    >
      暂无团队
    </Typography.Text>
  )
}
