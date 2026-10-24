import { Select, Typography } from 'antd'
import type { Team } from '../../api/teams'

type Props = {
  teams: Team[]
  teamId: string | null
  onChange: (teamId: string) => void
  disabled?: boolean
}

/** Lightweight team filter for list/board page bars (T6I I4). */
export function TeamPoolSelect({ teams, teamId, onChange, disabled }: Props) {
  if (teams.length > 1) {
    return (
      <Select
        size="small"
        value={teamId || undefined}
        placeholder="团队"
        style={{ minWidth: 140 }}
        options={teams.map((t) => ({ value: t.id, label: t.name }))}
        onChange={(v) => onChange(v)}
        disabled={disabled}
        popupMatchSelectWidth={false}
        aria-label="当前团队"
      />
    )
  }
  if (teams.length === 1) {
    return (
      <Typography.Text type="secondary" style={{ fontSize: 12 }}>
        团队 · {teams[0].name}
      </Typography.Text>
    )
  }
  return null
}
