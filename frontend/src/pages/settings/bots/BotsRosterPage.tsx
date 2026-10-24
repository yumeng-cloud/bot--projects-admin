import { useEffect, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import { Button, Space, Table, Tag, Typography, message, Spin, Empty } from 'antd'
import { listBots, listBotRoles, claimBotRole, releaseBotRole, type BotRosterItem, type BotRole } from '../../../api/bots'
import { roleLabel } from '../../../utils/labels'

export default function BotsRosterPage() {
  const { wsId = '' } = useParams()
  const [loading, setLoading] = useState(true)
  const [bots, setBots] = useState<BotRosterItem[]>([])
  const [roles, setRoles] = useState<BotRole[]>([])

  const load = async () => {
    setLoading(true)
    try {
      const [b, r] = await Promise.all([listBots(wsId), listBotRoles(wsId)])
      setBots(b.bots || [])
      setRoles(r.bot_roles || [])
    } catch (e: unknown) {
      message.error(e instanceof Error ? e.message : '加载失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void load()
  }, [wsId])

  if (loading) return <Spin />

  return (
    <div className="acb-page">
      <div className="acb-page-bar">
        <Typography.Title level={4} style={{ margin: 0 }}>Bot 名册</Typography.Title>
      </div>
      <Typography.Paragraph type="secondary">查看机器人成员与角色认领；详情页可看记忆、待办与交接。</Typography.Paragraph>

      <Typography.Title level={5}>角色</Typography.Title>
      {roles.length ? (
        <Table
          className="acb-list-table"
          rowKey="id"
          size="small"
          pagination={false}
          style={{ marginBottom: 24 }}
          dataSource={roles}
          columns={[
            { title: '名称', dataIndex: 'name' },
            { title: '职责', dataIndex: 'responsibilities', ellipsis: true },
            {
              title: '认领',
              render: (_, row) =>
                row.claim ? (
                  <Space>
                    <Tag color="blue">
                      {bots.find((b) => b.member_id === row.claim?.member_id)?.display_name || '已认领'}
                    </Tag>
                    <Button
                      size="small"
                      onClick={async () => {
                        await releaseBotRole(wsId, row.id, crypto.randomUUID())
                        message.success('已释放')
                        void load()
                      }}
                    >
                      释放
                    </Button>
                  </Space>
                ) : (
                  <Tag>空闲</Tag>
                ),
            },
          ]}
        />
      ) : (
        <div className="acb-empty-panel" style={{ marginBottom: 24 }}><Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无角色（seed 未跑或未创建）" /></div>
      )}

      <Typography.Title level={5}>Bot 成员</Typography.Title>
      {bots.length ? (
        <Table
          className="acb-list-table"
          rowKey="member_id"
          size="small"
          pagination={false}
          dataSource={bots}
          columns={[
            {
              title: '名称',
              dataIndex: 'display_name',
              render: (v, row) => <Link to={`/w/${wsId}/settings/bots/${row.member_id}`}>{v}</Link>,
            },
            {
              title: '角色',
              dataIndex: 'role',
              width: 100,
              render: (v: string) => roleLabel(v),
            },
            {
              title: '启用',
              dataIndex: 'bot_enabled',
              width: 72,
              render: (v: boolean) => (v ? <Tag color="green">是</Tag> : <Tag>否</Tag>),
            },
            {
              title: '认领角色',
              render: (_, row) => row.claimed_role?.name || '—',
            },
            {
              title: '操作',
              render: (_, row) => (
                <Space wrap size={4}>
                  <Link to={`/w/${wsId}/settings/bots/${row.member_id}`}>详情</Link>
                  {roles
                    .filter((r) => !r.claim)
                    .slice(0, 3)
                    .map((r) => (
                      <Button
                        key={r.id}
                        size="small"
                        onClick={async () => {
                          await claimBotRole(wsId, r.id, row.member_id, crypto.randomUUID())
                          message.success(`已认领「${r.name}」`)
                          void load()
                        }}
                      >
                        认领「{r.name}」
                      </Button>
                    ))}
                </Space>
              ),
            },
          ]}
        />
      ) : (
        <div className="acb-empty-panel"><Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无 Bot 成员" /></div>
      )}
    </div>
  )
}
