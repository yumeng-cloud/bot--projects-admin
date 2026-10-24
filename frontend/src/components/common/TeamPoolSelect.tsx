import { useState } from 'react'
import { Button, Divider, Empty, Select, Typography } from 'antd'
import { PlusOutlined, SettingOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import type { Team } from '../../api/teams'
import CreateTeamQuickModal from './CreateTeamQuickModal'

type Props = {
  teams: Team[]
  teamId: string | null
  onChange: (teamId: string) => void
  disabled?: boolean
  className?: string
  /** 有值时：可本页简易建队；建成功后回调刷新 */
  workspaceId?: string
  onTeamCreated?: (team: Team) => void
}

/** 顶栏队选择器：0 队主 CTA 建队；≥1 可切换；底「管理团队」（TEAM_ADMIN T3/T4） */
export function TeamPoolSelect({
  teams,
  teamId,
  onChange,
  disabled,
  className,
  workspaceId,
  onTeamCreated,
}: Props) {
  const nav = useNavigate()
  const [createOpen, setCreateOpen] = useState(false)
  const [open, setOpen] = useState(false)

  const valueInList = teamId && teams.some((t) => t.id === teamId) ? teamId : undefined

  const openCreate = () => {
    setOpen(false)
    setCreateOpen(true)
  }

  const goManage = () => {
    setOpen(false)
    if (workspaceId) nav(`/w/${workspaceId}/settings?tab=teams`)
  }

  const footer = workspaceId ? (
    <>
      <Divider style={{ margin: '4px 0' }} />
      <div style={{ padding: '2px 4px 6px', display: 'flex', flexDirection: 'column', gap: 0 }}>
        <Button
          type="text"
          size="small"
          icon={<PlusOutlined />}
          block
          style={{ textAlign: 'left', fontSize: 12, height: 28 }}
          onClick={openCreate}
        >
          创建团队
        </Button>
        <Button
          type="text"
          size="small"
          icon={<SettingOutlined />}
          block
          style={{ textAlign: 'left', fontSize: 12, height: 28 }}
          onClick={goManage}
        >
          管理团队
        </Button>
      </div>
    </>
  ) : null

  // 0 队：主 CTA，禁只剩灰字（T4）
  if (teams.length === 0) {
    return (
      <span className={className} aria-label="当前队">
        <span style={{ display: 'inline-flex', alignItems: 'center', gap: 6, flexWrap: 'wrap' }}>
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
              onClick={openCreate}
            >
              创建团队
            </Button>
          ) : null}
          {workspaceId ? (
            <Button
              type="link"
              size="small"
              disabled={disabled}
              style={{ paddingInline: 0, height: 'auto', fontSize: 12 }}
              onClick={goManage}
            >
              管理团队
            </Button>
          ) : null}
        </span>
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

  return (
    <>
      <Select
        className={className}
        size="small"
        variant="borderless"
        value={valueInList}
        placeholder="当前队"
        style={{ minWidth: 140 }}
        options={teams.map((t) => ({ value: t.id, label: t.name }))}
        onChange={(v) => {
          onChange(v)
          setOpen(false)
        }}
        disabled={disabled}
        open={open}
        onOpenChange={setOpen}
        popupMatchSelectWidth={false}
        aria-label="当前队"
        notFoundContent={
          <Empty
            image={Empty.PRESENTED_IMAGE_SIMPLE}
            description={
              <Typography.Text type="secondary" style={{ fontSize: 12 }}>
                还没有团队
              </Typography.Text>
            }
            styles={{ image: { height: 36 } }}
          />
        }
        popupRender={(menu) => (
          <div>
            {menu}
            {footer}
          </div>
        )}
      />
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
    </>
  )
}
