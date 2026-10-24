import { useEffect, useRef, useState } from 'react'
import { Link, Navigate, useNavigate, useSearchParams } from 'react-router-dom'
import {
  Alert,
  Button,
  Form,
  Input,
  Modal,
  Select,
  Spin,
  Table,
  Tabs,
  Tag,
  Tooltip,
  Typography,
  message,
  Empty,
  Popconfirm,
  Space,
} from 'antd'
import { ArrowLeftOutlined, PlusOutlined } from '@ant-design/icons'
import { me, type User } from '../../api/auth'
import { ApiError } from '../../api/client'
import {
  createAdminWorkspace,
  listAdminUsers,
  listAdminWorkspaces,
  patchAdminUser,
  patchAdminWorkspace,
  type AdminUser,
  type AdminUserWorkspace,
  type AdminWorkspace,
} from '../../api/admin'
import { listMyWorkspaces } from '../../api/workspaces'
import { formatShanghai } from '../../utils/time'

/** C4：工作区状态筛 */
type WsStatusFilter = 'all' | 'enabled' | 'disabled'

/**
 * 平台超管台（PRO_ADMIN A2）：顶栏标题+副文+主操作一排；工作区|用户对等；表区 hug。
 * C4：工作区表顶名称搜索 + 状态筛（FE 滤）。
 * D2：用户台所属 WS 芯片可点（有成员权进 WS / 否则复制名）；搜索防抖 + 空搜人话提示。
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
  const [createOpen, setCreateOpen] = useState(false)
  const [createForm] = Form.useForm<{ name: string; slug?: string; owner_email: string }>()
  const [creating, setCreating] = useState(false)
  const [userQuery, setUserQuery] = useState('')
  const [userQueryDebounced, setUserQueryDebounced] = useState('')
  const userSearchTimer = useRef<ReturnType<typeof setTimeout> | null>(null)
  /** D2：当前超管本人可进入的 WS（listMine）；用于芯片优先跳转 vs 复制 */
  const [myWsIds, setMyWsIds] = useState<Set<string>>(() => new Set())
  /** C4：工作区名称搜索 + 状态筛（FE 滤；API 无 q/status） */
  const [searchParams, setSearchParams] = useSearchParams()
  const forceWsFilter = searchParams.get('qaWsFilter') === '1'
  const tabParam = searchParams.get('tab')
  const initialTab = tabParam === 'users' || searchParams.get('qaUsers') === '1' ? 'users' : 'workspaces'
  const [activeTab, setActiveTab] = useState(initialTab)
  const [wsQuery, setWsQuery] = useState(forceWsFilter ? 'QA' : '')
  const [wsQueryDebounced, setWsQueryDebounced] = useState(forceWsFilter ? 'QA' : '')
  const wsSearchTimer = useRef<ReturnType<typeof setTimeout> | null>(null)
  const [wsStatusFilter, setWsStatusFilter] = useState<WsStatusFilter>(
    forceWsFilter ? 'enabled' : 'all',
  )

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
    if (wsSearchTimer.current) clearTimeout(wsSearchTimer.current)
    wsSearchTimer.current = setTimeout(() => {
      setWsQueryDebounced(wsQuery)
    }, 300)
    return () => {
      if (wsSearchTimer.current) clearTimeout(wsSearchTimer.current)
    }
  }, [wsQuery])

  useEffect(() => {
    if (!user?.is_platform_admin) return
    void loadWorkspaces()
    void listMyWorkspaces()
      .then((r) => setMyWsIds(new Set((r.workspaces || []).map((w) => w.id))))
      .catch(() => setMyWsIds(new Set()))
  }, [user?.is_platform_admin])

  useEffect(() => {
    if (!user?.is_platform_admin) return
    void loadUsers(userQueryDebounced)
  }, [user?.is_platform_admin, userQueryDebounced])

  if (boot) {
    return (
      <div className="acb-page acb-pro-admin" style={{ textAlign: 'center', paddingTop: 48 }}>
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
      message.success('已新建工作区')
      createForm.resetFields()
      setCreateOpen(false)
      await loadWorkspaces()
    } catch (e) {
      if (e && typeof e === 'object' && 'errorFields' in e) return
      message.error(e instanceof ApiError || e instanceof Error ? e.message : '新建失败')
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

  /** D2-1：所属 WS 芯片——本人有成员权则只读进 WS；否则复制名称 + Toast */
  const onUserWsChipClick = async (w: AdminUserWorkspace) => {
    const name = (w.name || '').trim() || '未命名工作区'
    if (myWsIds.has(w.id)) {
      nav(`/w/${w.id}/board`)
      return
    }
    try {
      await navigator.clipboard.writeText(name)
      message.success(`已复制工作区名称「${name}」`)
    } catch {
      message.warning(`无权限进入「${name}」，请手动复制名称`)
    }
  }

  const isWsDisabled = (row: AdminWorkspace) => !!(row.disabled_at || row.status === 'disabled')

  const filteredWorkspaces = workspaces.filter((row) => {
    const q = wsQueryDebounced.trim().toLowerCase()
    if (q && !row.name.toLowerCase().includes(q)) return false
    if (wsStatusFilter === 'enabled' && isWsDisabled(row)) return false
    if (wsStatusFilter === 'disabled' && !isWsDisabled(row)) return false
    return true
  })

  const wsFilterActive = !!(wsQueryDebounced.trim() || wsStatusFilter !== 'all')

  const empty = (desc: string) => (
    <div className="acb-empty-panel" style={{ padding: 24, minHeight: 120 }}>
      <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={desc} />
    </div>
  )

  return (
    <div className="acb-page acb-pro-admin">
      <div className="acb-page-bar">
        <div>
          <Typography.Title level={4} className="acb-page-title">
            平台管理
          </Typography.Title>
          <Typography.Text type="secondary" className="acb-page-sub">
            全站工作区与用户 · {user.name || user.email}
          </Typography.Text>
        </div>
        <Space>
          <Button icon={<ArrowLeftOutlined />} onClick={() => nav('/')}>
            返回工作区
          </Button>
          <Button
            type="primary"
            icon={<PlusOutlined />}
            onClick={() => {
              createForm.resetFields()
              setCreateOpen(true)
            }}
          >
            新建工作区
          </Button>
        </Space>
      </div>

      <div className="acb-surface-panel">
        <Tabs
          activeKey={activeTab}
          onChange={(key) => {
            setActiveTab(key)
            const next = new URLSearchParams(searchParams)
            if (key === 'users') next.set('tab', 'users')
            else next.delete('tab')
            next.delete('qaUsers')
            setSearchParams(next, { replace: true })
          }}
          items={[
            {
              key: 'workspaces',
              label: `工作区${filteredWorkspaces.length ? ` (${filteredWorkspaces.length})` : ''}`,
              children: (
                <>
                  <Space wrap style={{ marginBottom: 12 }} size={8}>
                    <Input.Search
                      allowClear
                      placeholder="搜索工作区名称"
                      value={wsQuery}
                      onChange={(e) => setWsQuery(e.target.value)}
                      onSearch={(v) => {
                        setWsQuery(v)
                        setWsQueryDebounced(v)
                      }}
                      style={{ width: 240 }}
                      aria-label="搜索工作区名称"
                    />
                    <Select
                      value={wsStatusFilter}
                      onChange={(v: WsStatusFilter) => setWsStatusFilter(v)}
                      style={{ width: 120 }}
                      aria-label="状态筛选"
                      options={[
                        { value: 'all', label: '全部' },
                        { value: 'enabled', label: '启用' },
                        { value: 'disabled', label: '停用' },
                      ]}
                    />
                  </Space>
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
                    {filteredWorkspaces.length ? (
                      <Table
                        className="acb-list-table"
                        size="small"
                        rowKey="id"
                        pagination={false}
                        dataSource={filteredWorkspaces}
                        columns={[
                          {
                            title: '名称',
                            dataIndex: 'name',
                            render: (v: string, row) => (
                              <Link to={`/w/${row.id}/board`}>{v}</Link>
                            ),
                          },
                          {
                            title: 'URL 标识',
                            dataIndex: 'slug',
                            width: 140,
                            ellipsis: true,
                            render: (v?: string) => v || '—',
                          },
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
                            title: '状态',
                            key: 'status',
                            width: 120,
                            render: (_, row) => {
                              const disabled = isWsDisabled(row)
                              return (
                                <Space size={8}>
                                  <Tag color={disabled ? 'default' : 'success'}>
                                    {disabled ? '已停用' : '启用'}
                                  </Tag>
                                  <Popconfirm
                                    title={disabled ? '确认恢复该工作区？' : '确认停用该工作区？'}
                                    okText="确认"
                                    cancelText="取消"
                                    onConfirm={() => void onToggleDisable(row, !disabled)}
                                  >
                                    <Button type="link" size="small" style={{ padding: 0 }}>
                                      {disabled ? '恢复' : '停用'}
                                    </Button>
                                  </Popconfirm>
                                </Space>
                              )
                            },
                          },
                        ]}
                      />
                    ) : (
                      empty(
                        wsErr ||
                          (wsFilterActive ? '没有符合条件的工作区' : '还没有工作区'),
                      )
                    )}
                  </Spin>
                </>
              ),
            },
            {
              key: 'users',
              label: `用户${users.length ? ` (${users.length})` : ''}`,
              children: (
                <>
                  <Input.Search
                    allowClear
                    placeholder="输入邮箱或名称筛选用户"
                    value={userQuery}
                    onChange={(e) => setUserQuery(e.target.value)}
                    onSearch={(v) => {
                      setUserQuery(v)
                      setUserQueryDebounced(v)
                    }}
                    style={{ maxWidth: 320, marginBottom: 8 }}
                    aria-label="输入邮箱或名称筛选用户"
                  />
                  {!userQueryDebounced.trim() ? (
                    <Typography.Paragraph type="secondary" style={{ fontSize: 12, marginTop: 0, marginBottom: 12 }}>
                      输入邮箱或名称筛选用户；留空显示全部
                    </Typography.Paragraph>
                  ) : (
                    <Typography.Paragraph type="secondary" style={{ fontSize: 12, marginTop: 0, marginBottom: 12 }}>
                      正在筛选「{userQueryDebounced.trim()}」
                    </Typography.Paragraph>
                  )}
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
                                <span style={{ display: 'inline-flex', flexWrap: 'wrap', gap: 6 }}>
                                  {list.map((w) => {
                                    const label = w.name || '未命名工作区'
                                    const canEnter = myWsIds.has(w.id)
                                    return (
                                      <Tooltip
                                        key={w.id}
                                        title={canEnter ? '点击进入该工作区' : '无成员权：点击复制名称'}
                                      >
                                        <Tag
                                          color={canEnter ? 'processing' : 'default'}
                                          style={{ margin: 0, cursor: 'pointer' }}
                                          onClick={() => void onUserWsChipClick(w)}
                                        >
                                          {label}
                                        </Tag>
                                      </Tooltip>
                                    )
                                  })}
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
                            key: 'platform_admin',
                            width: 120,
                            render: (_, row) => (
                              <Space size={8}>
                                <Tag color={row.is_platform_admin ? 'processing' : 'default'}>
                                  {row.is_platform_admin ? '是' : '否'}
                                </Tag>
                                <Popconfirm
                                  title={
                                    row.is_platform_admin
                                      ? '确认撤消该用户的平台超管？'
                                      : '确认授予平台超管？'
                                  }
                                  okText="确认"
                                  cancelText="取消"
                                  onConfirm={() => void onTogglePlatformAdmin(row, !row.is_platform_admin)}
                                >
                                  <Button type="link" size="small" style={{ padding: 0 }}>
                                    {row.is_platform_admin ? '撤消' : '授予'}
                                  </Button>
                                </Popconfirm>
                              </Space>
                            ),
                          },
                        ]}
                      />
                    ) : (
                      empty(
                        userErr ||
                          (userQueryDebounced.trim()
                            ? '无匹配用户，试试其他邮箱或名称'
                            : '还没有用户 · 输入邮箱或名称筛选用户'),
                      )
                    )}
                  </Spin>
                </>
              ),
            },
          ]}
        />
      </div>

      <Modal
        title="新建工作区"
        open={createOpen}
        onCancel={() => setCreateOpen(false)}
        onOk={() => void onCreateWs()}
        confirmLoading={creating}
        okText="创建"
        cancelText="取消"
        destroyOnHidden
      >
        <Form form={createForm} layout="vertical" style={{ marginTop: 8 }}>
          <Form.Item name="name" label="名称" rules={[{ required: true, message: '请填写名称' }]}>
            <Input placeholder="工作区名称" maxLength={80} />
          </Form.Item>
          <Form.Item
            name="slug"
            label={
              <span>
                URL 标识
                <span className="acb-settings-slug-extra">用于 URL</span>
              </span>
            }
          >
            <Input placeholder="可选，留空自动生成" maxLength={64} />
          </Form.Item>
          <Form.Item
            name="owner_email"
            label="所有者邮箱"
            rules={[
              { required: true, message: '请填写所有者邮箱' },
              { type: 'email', message: '邮箱格式不正确' },
            ]}
          >
            <Input placeholder="已注册用户邮箱" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
