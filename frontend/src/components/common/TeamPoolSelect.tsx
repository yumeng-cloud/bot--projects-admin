import { useState } from 'react'
import { Button, Select, Space, Typography } from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import { Link } from 'react-router-dom'
import type { Team } from '../../api/teams'
import CreateTeamQuickModal from './CreateTeamQuickModal'

type Props = {
  teams: Team[]
  teamId: string | null
  onChange: (teamId: string) => void
  disabled?: boolean
  className?: string
  /** 有值时：空态可本页简易建队；建成功后回调刷新 */
  workspaceId?: string
  onTeamCreated?: (team: Team) => void
}

/** Lightweight current-team control (T6I I4 / TEAM_SWITCH). Header + page bars. */
export function TeamPoolSelect({
  teams,
  teamId,
  onChange,
  disabled,
  className,
  workspaceId,
  onTeamCreated,
}: Props) {
  const [createOpen, setCreateOpen] = useState(false)

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
    <span className={className} aria-label="当前队">
      <Space size={6} align="center">
        <Typography.Text type="secondary" style={{ fontSize: 12, whiteSpace: 'nowrap' }}>
          暂无团队
        </Typography.Text>
        {workspaceId ? (
          <Button
            type="link"
            size="small"
            icon={<PlusOutlined />}
            disabled={disabled}
            style={{ paddingInline: 0, height: 'auto', fontSize: 12 }}
            onClick={() => setCreateOpen(true)}
          >
            创建团队
          </Button>
        ) : null}
        {workspaceId ? (
          <Link
            to={`/w/${workspaceId}/settings/bots`}
            style={{ fontSize: 12, whiteSpace: 'nowrap' }}
          >
            去 Bot 席
          </Link>
        ) : null}
      </Space>
      {workspaceId ? (
        <CreateTeamQuickModal
          open={createOpen}
          workspaceId={workspaceId}
          onClose={() => setCreateOpen(false)}
          onCreated={(team) => {
            onTeamCreated?.(team)
            onChange(team.id)
          }}
        />
      ) : null}
    </span>
  )
}
