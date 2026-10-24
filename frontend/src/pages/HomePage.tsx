import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { Card, Col, Row, Typography, Spin, Button, Space, Tag } from 'antd'
import { me } from '../api/auth'
import { listMyWorkspaces, type Workspace } from '../api/workspaces'

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

  if (loading) return <div style={{ padding: 48, textAlign: 'center' }}><Spin /></div>

  return (
    <div style={{ maxWidth: 960, margin: '0 auto', padding: 24 }}>
      <Space style={{ width: '100%', justifyContent: 'space-between', marginBottom: 24 }}>
        <div>
          <Typography.Title level={3} style={{ margin: 0 }}>工作空间</Typography.Title>
          <Typography.Text type="secondary">{user?.name} · {user?.email}</Typography.Text>
        </div>
        <Button onClick={onLogout}>退出登录</Button>
      </Space>
      <Row gutter={[16, 16]}>
        {workspaces.map((w) => (
          <Col span={12} key={w.id}>
            <Card
              title={w.name}
              extra={<Tag color="blue">{w.role}</Tag>}
              hoverable
              onClick={() => {
                localStorage.setItem('workspaceId', w.id)
                nav(`/w/${w.id}/board`)
              }}
            >
              <Typography.Paragraph type="secondary">slug: {w.slug}</Typography.Paragraph>
              <Typography.Text type="secondary">进入看板 / 列表 / Cycles / 设置</Typography.Text>
            </Card>
          </Col>
        ))}
        {workspaces.length === 0 && (
          <Col span={24}>
            <Card><Typography.Text type="secondary">暂无工作空间，请运行 seed</Typography.Text></Card>
          </Col>
        )}
      </Row>
    </div>
  )
}
