import { useCallback, useEffect, useState } from 'react'
import { Button, Empty, Space, Table, Tag, Typography, message } from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import { Link, useNavigate } from 'react-router-dom'
import { getRoster, listTeams, type Team } from '../../api/teams'
import CreateTeamQuickModal from '../../components/common/CreateTeamQuickModal'
import { useCurrentTeam } from '../../hooks/useCurrentTeam'
import { formatShanghai } from '../../utils/time'

type Props = {
  workspaceId: string
}

type Row = Team & { seatTotal?: number; seatActive?: number }

/** 设置 · 团队 Tab（TEAM_ADMIN T1/T2） */
export default function SettingsTeamsPanel({ workspaceId }: Props) {
  const nav = useNavigate()
  const { teamId, setTeamId, reload: reloadCurrent } = useCurrentTeam(workspaceId)
  const [loading, setLoading] = useState(true)
  const [rows, setRows] = useState<Row[]>([])
  const [createOpen, setCreateOpen] = useState(false)

  const reload = useCallback(async () => {
    if (!workspaceId) return
    setLoading(true)
    try {
      const res = await listTeams(workspaceId)
      const list = res.teams || []
      const withSeats = await Promise.all(
        list.map(async (t) => {
          try {
            const { roster } = await getRoster(workspaceId, t.id)
            const seats = roster || []
            const seatActive = seats.filter(
              (s) => s.bound_member_id || s.claim?.member_id || (s.status || '').toLowerCase() === 'active',
            ).length
            return { ...t, seatTotal: seats.length, seatActive } as Row
          } catch {
            return { ...t } as Row
          }
        }),
      )
      setRows(withSeats)
    } catch (e: unknown) {
      message.error(e instanceof Error ? e.message : '加载团队失败')
      setRows([])
    } finally {
      setLoading(false)
    }
  }, [workspaceId])

  useEffect(() => {
    void reload()
  }, [reload])

  const onSwitch = (row: Team) => {
    setTeamId(row.id)
    message.success(`已切换到「${row.name}」`)
  }

  return (
    <>
      <div className="acb-settings-section-head" style={{ marginBottom: 12 }}>
        <span className="acb-settings-section-title">团队</span>
        <Space size={8} wrap>
          <Button
            type="primary"
            size="small"
            icon={<PlusOutlined />}
            onClick={() => setCreateOpen(true)}
          >
            新建团队
          </Button>
          <Link to={`/w/${workspaceId}/settings/bots`} style={{ fontSize: 12 }}>
            去 Bot 席
          </Link>
        </Space>
      </div>
      <p className="acb-settings-hint" style={{ marginBottom: 12 }}>
        切换当前队后看板池会跟队；席位与入场码在 Bot 席管理。人类成员见「成员」Tab。
      </p>
      {rows.length ? (
        <Table
          className="acb-list-table"
          size="small"
          rowKey="id"
          loading={loading}
          dataSource={rows}
          pagination={false}
          columns={[
            {
              title: '名称',
              dataIndex: 'name',
              render: (v: string, row: Row) => (
                <span>
                  <Typography.Text strong>{v || '未命名'}</Typography.Text>
                  {teamId === row.id ? (
                    <Tag color="blue" style={{ marginLeft: 8 }}>
                      当前
                    </Tag>
                  ) : null}
                </span>
              ),
            },
            {
              title: '席位',
              key: 'seats',
              width: 120,
              render: (_: unknown, row: Row) => {
                if (row.seatTotal == null) return '—'
                return (
                  <Typography.Text type="secondary" style={{ fontSize: 12 }}>
                    在席 {row.seatActive ?? 0}/{row.seatTotal}
                  </Typography.Text>
                )
              },
            },
            {
              title: '创建',
              dataIndex: 'created_at',
              width: 140,
              render: (v?: string) => (v ? formatShanghai(v) : '—'),
            },
            {
              title: '操作',
              key: 'actions',
              width: 220,
              render: (_: unknown, row: Row) => (
                <Space size={0} wrap>
                  <Button
                    type="link"
                    size="small"
                    disabled={teamId === row.id}
                    onClick={() => onSwitch(row)}
                  >
                    设为当前
                  </Button>
                  <Button
                    type="link"
                    size="small"
                    onClick={() => nav(`/w/${workspaceId}/settings?tab=members`)}
                  >
                    成员
                  </Button>
                  <Button
                    type="link"
                    size="small"
                    onClick={() => nav(`/w/${workspaceId}/settings/bots?team=${row.id}`)}
                  >
                    Bot 席
                  </Button>
                </Space>
              ),
            },
          ]}
        />
      ) : (
        <div className="acb-empty-panel" style={{ minHeight: 120, padding: '20px 16px' }}>
          <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="还没有团队">
            <Button type="primary" icon={<PlusOutlined />} onClick={() => setCreateOpen(true)}>
              新建团队
            </Button>
          </Empty>
        </div>
      )}
      <CreateTeamQuickModal
        open={createOpen}
        workspaceId={workspaceId}
        onClose={() => setCreateOpen(false)}
        onCreated={async (team) => {
          setTeamId(team.id)
          await reloadCurrent()
          await reload()
        }}
      />
    </>
  )
}
