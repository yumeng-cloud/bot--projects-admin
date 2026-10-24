import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Button, Space, Spin, Table, Typography } from 'antd'
import { me } from '../api/auth'
import { listMyWorkspaces, type Workspace } from '../api/workspaces'
import { roleLabel } from '../utils/labels'

export default function HomePage({ onLogout }: { onLogout: () => void }) {
  const nav = useNavigate()
  const [loading, setLoading] = useState(true)
  const [workspaces, setWorkspaces] = useState<Workspace[]>([])
  const [user, setUser] = useState<{ name: string; email: string } | null>(null)

  useEffect(() => {
    ;(async () => {
      try {
        const meResp = await me()
        setUser(meResp.user)
        const ws = await listMyWorkspaces()
        setWorkspaces(ws.workspaces)
      } finally {
        setLoading(false)
      }
    })()
  }, [])

  if (loading) {
    return (
      <div className="acb-page" style={{ textAlign: 'center', paddingTop: 48 }}>
        <Spin />
      </div>
    )
  }

  const enter = (w: Workspace, path: string) => {
    localStorage.setItem('workspaceId', w.id)
    nav(`/w/${w.id}/${path}`)
  }

  return (
    <div className="acb-page">
      <div className="acb-page-bar">
        <div>
          <Typography.Title level={4} style={{ margin: 0 }}>
            工作空间
          </Typography.Title>
          <Typography.Text type="secondary">
            {user?.name} · {user?.email}
          </Typography.Text>
        </div>
        <Button onClick={onLogout}>退出登录</Button>
      </div>

      <Table
        className="acb-list-table"
        size="small"
        rowKey="id"
        pagination={false}
        dataSource={workspaces}
        locale={{ emptyText: '暂无工作空间，请运行 seed' }}
        columns={[
          {
            title: '名称',
            dataIndex: 'name',
            render: (v: string, w) => (
              <Button type="link" style={{ padding: 0 }} onClick={() => enter(w, 'board')}>
                {v}
              </Button>
            ),
          },
          {
            title: '角色',
            dataIndex: 'role',
            width: 100,
            render: (v: string) => roleLabel(v),
          },
          {
            title: '进入',
            key: 'nav',
            render: (_, w) => (
              <Space size={4} wrap>
                <Button size="small" type="link" onClick={() => enter(w, 'board')}>
                  看板
                </Button>
                <Button size="small" type="link" onClick={() => enter(w, 'list')}>
                  列表
                </Button>
                <Button size="small" type="link" onClick={() => enter(w, 'cycles')}>
                  Cycles
                </Button>
                <Button size="small" type="link" onClick={() => enter(w, 'wiki')}>
                  Wiki
                </Button>
                <Button size="small" type="link" onClick={() => enter(w, 'settings')}>
                  设置
                </Button>
              </Space>
            ),
          },
        ]}
      />
    </div>
  )
}
