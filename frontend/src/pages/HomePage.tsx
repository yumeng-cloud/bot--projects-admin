import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import {
  Button,
  Form,
  Input,
  Modal,
  Space,
  Spin,
  Table,
  Typography,
  message,
  Empty,
} from 'antd'
import { PlusOutlined, CrownOutlined, GiftOutlined } from '@ant-design/icons'
import { me } from '../api/auth'
import {
  createWorkspace,
  listMyWorkspaces,
  slugifyName,
  type Workspace,
} from '../api/workspaces'
import { ApiError } from '../api/client'
import { redeemWorkspaceInvite } from '../api/workspaceInvites'
import { roleLabel } from '../utils/labels'

export default function HomePage({ onLogout }: { onLogout: () => void }) {
  const nav = useNavigate()
  const [loading, setLoading] = useState(true)
  const [workspaces, setWorkspaces] = useState<Workspace[]>([])
  const [user, setUser] = useState<{
    name: string
    email: string
    is_platform_admin?: boolean
  } | null>(null)
  const [createOpen, setCreateOpen] = useState(false)
  const [creating, setCreating] = useState(false)
  const [slugTouched, setSlugTouched] = useState(false)
  const [form] = Form.useForm<{ name: string; slug?: string }>()
  const [redeemOpen, setRedeemOpen] = useState(false)
  const [redeeming, setRedeeming] = useState(false)
  const [redeemForm] = Form.useForm<{ code: string }>()

  const reload = async () => {
    const meResp = await me()
    setUser(meResp.user)
    const ws = await listMyWorkspaces()
    setWorkspaces(ws.workspaces)
  }

  useEffect(() => {
    ;(async () => {
      try {
        await reload()
      } finally {
        setLoading(false)
      }
    })()
  }, [])

  const enter = (w: Workspace, path: string) => {
    localStorage.setItem('workspaceId', w.id)
    nav(`/w/${w.id}/${path}`)
  }

  const openCreate = () => {
    setSlugTouched(false)
    form.resetFields()
    setCreateOpen(true)
  }

  const onCreate = async () => {
    try {
      const values = await form.validateFields()
      setCreating(true)
      const resp = await createWorkspace({
        name: values.name,
        slug: values.slug?.trim() || undefined,
      })
      message.success('工作空间已创建')
      setCreateOpen(false)
      localStorage.setItem('workspaceId', resp.workspace.id)
      nav(`/w/${resp.workspace.id}/board`)
    } catch (e) {
      if (e && typeof e === 'object' && 'errorFields' in e) return
      const msg =
        e instanceof ApiError
          ? e.message
          : e instanceof Error
            ? e.message
            : '创建失败'
      message.error(msg)
    } finally {
      setCreating(false)
    }
  }


  const openRedeem = () => {
    redeemForm.resetFields()
    setRedeemOpen(true)
  }

  const onRedeem = async () => {
    try {
      const values = await redeemForm.validateFields()
      const code = values.code.trim()
      if (/^ymt_/i.test(code)) {
        message.error('邀请码须以 ymw_ 开头（与队席 ymt_ 隔离）')
        return
      }
      setRedeeming(true)
      const res = await redeemWorkspaceInvite(code, crypto.randomUUID())
      const wsId = res.workspace_id
      message.success('已加入工作空间')
      setRedeemOpen(false)
      localStorage.setItem('workspaceId', wsId)
      nav(`/w/${wsId}/board`)
    } catch (e) {
      if (e && typeof e === 'object' && 'errorFields' in e) return
      message.error(e instanceof ApiError || e instanceof Error ? e.message : '兑换失败')
    } finally {
      setRedeeming(false)
    }
  }

  if (loading) {
    return (
      <div className="acb-page" style={{ textAlign: 'center', paddingTop: 48 }}>
        <Spin />
      </div>
    )
  }

  const empty = workspaces.length === 0

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
        <Space>
          {user?.is_platform_admin ? (
            <Button icon={<CrownOutlined />} onClick={() => nav('/admin')}>
              平台管理
            </Button>
          ) : null}
          <Button icon={<GiftOutlined />} onClick={openRedeem}>
            兑换邀请码
          </Button>
          <Button type="primary" icon={<PlusOutlined />} onClick={openCreate}>
            创建工作空间
          </Button>
          <Button onClick={onLogout}>退出登录</Button>
        </Space>
      </div>

      <div className="acb-surface-panel">
        <Typography.Text strong style={{ display: 'block', marginBottom: 8 }}>
          我的工作区{workspaces.length ? ` · ${workspaces.length}` : ''}
        </Typography.Text>

        {empty ? (
          <div className="acb-empty-panel" style={{ padding: '32px 16px', textAlign: 'center' }}>
            <Empty
              image={Empty.PRESENTED_IMAGE_SIMPLE}
              description={
                <span>
                  还没有工作空间，创建一个开始协作，或兑换邀请码加入
                </span>
              }
            >
              <Space>
                <Button type="primary" size="large" icon={<PlusOutlined />} onClick={openCreate}>
                  创建工作空间
                </Button>
                <Button size="large" icon={<GiftOutlined />} onClick={openRedeem}>
                  兑换邀请码
                </Button>
              </Space>
            </Empty>
          </div>
        ) : (
          <Table
            className="acb-list-table"
            size="small"
            rowKey="id"
            pagination={false}
            dataSource={workspaces}
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
                      周期
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
        )}
        <div className="acb-surface-panel-foot">
          登录后可自助创建工作空间，或兑换 ymw_ 邀请码加入；设置中可邀请成员与签发链接。
        </div>
      </div>

      <Modal
        title="兑换工作区邀请码"
        open={redeemOpen}
        onCancel={() => setRedeemOpen(false)}
        onOk={() => void onRedeem()}
        confirmLoading={redeeming}
        okText="兑换并进入"
        cancelText="取消"
        destroyOnClose
      >
        <Form form={redeemForm} layout="vertical" requiredMark="optional" style={{ marginTop: 8 }}>
          <Form.Item
            name="code"
            label="邀请码"
            extra="须以 ymw_ 开头；队席码 ymt_ 请到 Bot 班底兑换"
            rules={[
              { required: true, message: '请粘贴邀请码' },
              {
                pattern: /^ymw_[A-Za-z0-9_-]+$/i,
                message: '邀请码须以 ymw_ 开头',
              },
            ]}
          >
            <Input placeholder="ymw_…" autoFocus />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title="创建工作空间"
        open={createOpen}
        onCancel={() => setCreateOpen(false)}
        onOk={() => void onCreate()}
        confirmLoading={creating}
        okText="创建"
        cancelText="取消"
        destroyOnClose
      >
        <Form
          form={form}
          layout="vertical"
          requiredMark="optional"
          style={{ marginTop: 8 }}
          onValuesChange={(changed) => {
            if ('name' in changed && !slugTouched) {
              form.setFieldsValue({ slug: slugifyName(changed.name || '') || undefined })
            }
          }}
        >
          <Form.Item
            name="name"
            label="名称"
            rules={[{ required: true, message: '请填写工作空间名称' }]}
          >
            <Input placeholder="例如：产品协作" maxLength={80} autoFocus />
          </Form.Item>
          <Form.Item
            name="slug"
            label="标识（slug）"
            extra="小写字母、数字与连字符；可留空由系统生成"
            rules={[
              {
                pattern: /^([a-z0-9]+(?:-[a-z0-9]+)*)?$/,
                message: '仅允许小写字母、数字与连字符',
              },
            ]}
          >
            <Input
              placeholder="自动生成或手动填写"
              maxLength={64}
              onChange={() => setSlugTouched(true)}
            />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
