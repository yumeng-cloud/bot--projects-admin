import { useEffect, useRef, useState } from 'react'
import { Link, Navigate, useNavigate } from 'react-router-dom'
import {
  Alert,
  Button,
  Form,
  Input,
  Spin,
  Switch,
  Table,
  Tabs,
  Tag,
  Typography,
  message,
  Empty,
  Popconfirm,
} from 'antd'
import { ArrowLeftOutlined } from '@ant-design/icons'
import { me, type User } from '../../api/auth'
import { ApiError } from '../../api/client'
import {
  createAdminWorkspace,
  listAdminUsers,
  listAdminWorkspaces,
  patchAdminUser,
  patchAdminWorkspace,
  type AdminUser,
  type AdminWorkspace,
} from '../../api/admin'
import { formatShanghai } from '../../utils/time'

/**
 * C7–C9：平台超管台。仅 is_platform_admin 可见；真实 API，无假数据。
 */
export default function PlatformAdminPage() {
  const nav = useNavigate()
  const [boot, setBoot] = useState(true)
  const [user, setUser] = useState<User | null>(null)
  const [workspaces, setWorkspaces] = useState<AdminWorkspace[]>([])
  const [users, setUsers] = useState<AdminUser[]>([])
  const [wsErr, setWsErr] = useState<string | null>(null)
  const [userErr, setUserErr] = useState<string | null>(null)
  const [wsLoading, setWsLoading] = useState(false)
  const [userLoading, setUserLoading] = useState(false)
  const [createForm] = Form.useForm<{ name: string; slug?: string; owner_email: string }>()
  const [creating, setCreating] = useState(false)
  const [userQuery, setUserQuery] = useState('')
  const [userQueryDebounced, setUserQueryDebounced] = useState('')
  const userSearchTimer = useRef<ReturnType<typeof setTimeout> | null>(null)

  useEffect(() => {
    ;(async () => {
      try {
        const r = await me()
        setUser(r.user)
      } catch {
        setUser(null)
      } finally {
        setBoot(false)
      }
    })()
  }, [])

  const loadWorkspaces = async () => {
    setWsLoading(true)
    setWsErr(null)
    try {
      const data = await listAdminWorkspaces()
      setWorkspaces(data.workspaces || [])
    } catch (e) {
      setWorkspaces([])
      const msg =
        e instanceof ApiError
          ? e.status === 403
            ? '需要平台超级管理员权限'
            : e.message
          : e instanceof Error
            ? e.message
            : '加载失败'
      setWsErr(msg)
      if (e instanceof ApiError && e.status !== 404) message.error(msg)
    } finally {
      setWsLoading(false)
    }
  }

  const loadUsers = async (q?: string) => {
    setUserLoading(true)
    setUserErr(null)
    try {
      const data = await listAdminUsers(q?.trim() ? { q: q.trim() } : undefined)
      setUsers(data.users || [])
    } catch (e) {
      setUsers([])
      const msg =
        e instanceof ApiError
          ? e.status === 403
            ? '需要平台超级管理员权限'
            : e.message
          : e instanceof Error
            ? e.message
            : '加载失败'
      setUserErr(msg)
      if (e instanceof ApiError && e.status !== 404) message.error(msg)
    } finally {
      setUserLoading(false)
    }
  }

  useEffect(() => {
    if (userSearchTimer.current) clearTimeout(userSearchTimer.current)
    userSearchTimer.current = setTimeout(() => {
      setUserQueryDebounced(userQuery)
    }, 300)
    return () => {
      if (userSearchTimer.current) clearTimeout(userSearchTimer.current)
    }
  }, [userQuery])

  useEffect(() => {
    if (!user?.is_platform_admin) return
    void loadWorkspaces()
  }, [user?.is_platform_admin])

  useEffect(() => {
    if (!user?.is_platform_admin) return
    void loadUsers(userQueryDebounced)
  }, [user?.is_platform_admin, userQueryDebounced])

  if (boot) {
    return (
      <div className="acb-page" style={{ textAlign: 'center', paddingTop: 48 }}>
        <Spin />
      </div>
    )
  }

  if (!user?.is_platform_admin) {
    return <Navigate to="/" replace />
  }

  const onToggleDisable = async (row: AdminWorkspace, disabled: boolean) => {
    try {
      await patchAdminWorkspace(row.id, { disabled })
      message.success(disabled ? '已停用' : '已恢复')
      await loadWorkspaces()
    } catch (e) {
      message.error(e instanceof ApiError || e instanceof Error ? e.message : '操作失败')
    }
  }

  const onCreateWs = async () => {
    try {
      const values = await createForm.validateFields()
      setCreating(true)
      await createAdminWorkspace({
        name: values.name,
        slug: values.slug?.trim() || undefined,
        owner_email: values.owner_email.trim().toLowerCase(),
      })
      message.success('已代开工作空间')
      createForm.resetFields()
      await loadWorkspaces()
    } catch (e) {
      if (e && typeof e === 'object' && 'errorFields' in e) return
      message.error(e instanceof ApiError || e instanceof Error ? e.message : '代开失败')
    } finally {
      setCreating(false)
    }
  }

  const onTogglePlatformAdmin = async (row: AdminUser, next: boolean) => {
    try {
      await patchAdminUser(row.id, { is_platform_admin: next })
      message.success(next ? '已授予平台超管' : '已撤消平台超管')
      await loadUsers(userQueryDebounced)
    } catch (e) {
      message.error(e instanceof ApiError || e instanceof Error ? e.message : '操作失败')
    }
  }

  const empty = (desc: string) => (
    <div className="acb-empty-panel" style={{ padding: 24 }}>
      <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={desc} />
    </div>
  )

  return (
    <div className="acb-page">
      <div className="acb-page-bar">
        <div>
          <Typography.Title level={4} style={{ margin: 0 }}>
            平台管理
          </Typography.Title>
          <Typography.Text type="secondary" style={{ fontSize: 12 }}>
            全站工作区与用户 · {user.name} · {user.email}
          </Typography.Text>
        </div>
        <Button icon={<ArrowLeftOutlined />} onClick={() => nav('/')}>
          返回工作区
        </Button>
      </div>

      <div className="acb-surface-panel">
        <Tabs
          items={[
            {
              key: 'workspaces',
              label: '工作区',
              children: (
                <>
                  <Typography.Text strong style={{ display: 'block', marginBottom: 8 }}>
                    代开工作空间
                  </Typography.Text>
                  <Form form={createForm} layout="inline" style={{ marginBottom: 16, rowGap: 8 }}>
                    <Form.Item name="name" rules={[{ required: true, message: '请填写名称' }]}>
                      <Input placeholder="名称" style={{ width: 160 }} />
                    </Form.Item>
                    <Form.Item name="slug">
                      <Input placeholder="slug（可选）" style={{ width: 140 }} />
                    </Form.Item>
                    <Form.Item
                      name="owner_email"
                      rules={[
                        { required: true, message: '请填写 owner 邮箱' },
                        { type: 'email', message: '邮箱格式不正确' },
                      ]}
                    >
                      <Input placeholder="owner 邮箱" style={{ width: 200 }} />
                    </Form.Item>
                    <Form.Item>
                      <Button type="primary" loading={creating} onClick={() => void onCreateWs()}>
                        代开
                      </Button>
                    </Form.Item>
                  </Form>

                  {wsErr ? (
                    <Alert
                      type="warning"
                      showIcon
                      style={{ marginBottom: 12 }}
                      message={wsErr}
                      action={
                        <Button size="small" onClick={() => void loadWorkspaces()}>
                          重试
                        </Button>
                      }
                    />
                  ) : null}

                  <Spin spinning={wsLoading}>
                    {workspaces.length ? (
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
                            render: (v: string, row) => (
                              <Link to={`/w/${row.id}/board`}>{v}</Link>
                            ),
                          },
                          { title: 'slug', dataIndex: 'slug', width: 140, ellipsis: true },
                          {
                            title: '成员',
                            dataIndex: 'member_count',
                            width: 72,
                            render: (v?: number) => (v == null ? '—' : v),
                          },
                          {
                            title: '创建',
                            dataIndex: 'created_at',
                            width: 140,
                            render: (v?: string) => (v ? formatShanghai(v) : '—'),
                          },
                          {
                            title: '停用',
                            key: 'disabled',
                            width: 90,
                            render: (_, row) => {
                              const disabled = !!(row.disabled_at || row.status === 'disabled')
                              return (
                                <Popconfirm
                                  title={disabled ? '确认恢复该工作区？' : '确认停用该工作区？'}
                                  okText="确认"
                                  cancelText="取消"
                                  onConfirm={() => void onToggleDisable(row, !disabled)}
                                >
                                  <Switch size="small" checked={disabled} checkedChildren="停" unCheckedChildren="用" />
                                </Popconfirm>
                              )
                            },
                          },
                        ]}
                      />
                    ) : (
                      empty(wsErr || '暂无全站工作区')
                    )}
                  </Spin>
                </>
              ),
            },
            {
              key: 'users',
              label: '用户',
              children: (
                <>
                  <Input.Search
                    allowClear
                    placeholder="搜索邮箱或名称"
                    value={userQuery}
                    onChange={(e) => setUserQuery(e.target.value)}
                    onSearch={(v) => {
                      setUserQuery(v)
                      setUserQueryDebounced(v)
                    }}
                    style={{ maxWidth: 320, marginBottom: 12 }}
                  />
                  {userErr ? (
                    <Alert
                      type="warning"
                      showIcon
                      style={{ marginBottom: 12 }}
                      message={userErr}
                      action={
                        <Button size="small" onClick={() => void loadUsers(userQueryDebounced)}>
                          重试
                        </Button>
                      }
                    />
                  ) : null}
                  <Spin spinning={userLoading}>
                    {users.length ? (
                      <Table
                        className="acb-list-table"
                        size="small"
                        rowKey="id"
                        pagination={false}
                        dataSource={users}
                        columns={[
                          { title: '名称', dataIndex: 'name' },
                          { title: '邮箱', dataIndex: 'email', ellipsis: true },
                          {
                            title: '所属工作区',
                            key: 'workspaces',
                            ellipsis: true,
                            render: (_, row) => {
                              const list = row.workspaces || []
                              if (!list.length) return '—'
                              return (
                                <span style={{ display: 'inline-flex', flexWrap: 'wrap', gap: 4 }}>
                                  {list.map((w) => (
                                    <Tag key={w.id} style={{ margin: 0 }}>
                                      {w.name || w.slug || w.id}
                                    </Tag>
                                  ))}
                                </span>
                              )
                            },
                          },
                          {
                            title: '注册',
                            dataIndex: 'created_at',
                            width: 140,
                            render: (v?: string) => (v ? formatShanghai(v) : '—'),
                          },
                          {
                            title: '平台超管',
                            key: 'is_platform_admin',
                            width: 110,
                            render: (_, row) => (
                              <Switch
                                size="small"
                                checked={!!row.is_platform_admin}
                                onChange={(checked) => void onTogglePlatformAdmin(row, checked)}
                              />
                            ),
                          },
                        ]}
                      />
                    ) : (
                      empty(userErr || (userQueryDebounced.trim() ? '无匹配用户' : '暂无用户'))
                    )}
                  </Spin>
                </>
              ),
            },
          ]}
        />
        <div className="acb-surface-panel-foot">
          平台管理仅对超管可见。可停用/恢复工作区、代开工作空间、授撤平台超管。
        </div>
      </div>
    </div>
  )
}
