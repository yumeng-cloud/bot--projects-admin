import { useCallback, useEffect, useMemo, useState } from 'react'
import {
  Tabs,
  Table,
  Typography,
  message,
  Spin,
  Empty,
  Tag,
  Button,
  Form,
  Input,
  Select,
  Space,
  Popconfirm,
  Alert,
} from 'antd'
import { useParams } from 'react-router-dom'
import { api, ApiError } from '../../api/client'
import { getWorkspace, listMyWorkspaces, patchWorkspace, slugifyName } from '../../api/workspaces'
import {
  inviteMember,
  listMembers,
  patchMember,
  removeMember,
  type WorkspaceMember,
} from '../../api/members'
import { roleLabel } from '../../utils/labels'
import { formatShanghai } from '../../utils/time'

type Member = WorkspaceMember
type Label = { id: string; name: string; color: string }
type ApiKey = { id: string; name: string; key_prefix: string; created_at?: string }
type Webhook = { id: string; url: string; enabled: number; events?: string }

function keyTail(prefix?: string | null): string {
  if (!prefix) return '—'
  const clean = prefix.replace(/^ymk_/i, '')
  const tail = clean.slice(-4)
  return tail ? `末四 ·${tail}` : prefix
}

function isBotMember(m: Member): boolean {
  return m.is_bot === true || m.is_bot === 1 || m.role === 'bot'
}

export default function SettingsPage() {
  const { wsId = '' } = useParams()
  const [loading, setLoading] = useState(true)
  const [members, setMembers] = useState<Member[]>([])
  const [labels, setLabels] = useState<Label[]>([])
  const [keys, setKeys] = useState<ApiKey[]>([])
  const [hooks, setHooks] = useState<Webhook[]>([])
  const [err, setErr] = useState<string | null>(null)
  const [myRole, setMyRole] = useState<string>('')
  const [wsName, setWsName] = useState('')
  const [wsSlug, setWsSlug] = useState('')
  const [wsSaving, setWsSaving] = useState(false)
  const [wsForm] = Form.useForm<{ name: string; slug: string }>()
  const [inviteForm] = Form.useForm<{ email: string; role: 'admin' | 'member'; display_name?: string }>()
  const [inviting, setInviting] = useState(false)

  const isAdmin = myRole === 'admin' || myRole === 'owner'

  const adminCount = useMemo(
    () => members.filter((m) => !isBotMember(m) && (m.role === 'admin' || m.role === 'owner')).length,
    [members],
  )

  const reloadMembers = useCallback(async () => {
    if (!wsId) return
    const data = await listMembers(wsId)
    setMembers(data.members || [])
  }, [wsId])

  useEffect(() => {
    ;(async () => {
      setLoading(true)
      setErr(null)
      try {
        const [m, l, k, h, ws] = await Promise.allSettled([
          listMembers(wsId),
          api<{ labels: Label[] }>(`/workspaces/${wsId}/labels`),
          api<{ api_keys: ApiKey[] }>(`/workspaces/${wsId}/api-keys`),
          api<{ webhooks: Webhook[] }>(`/workspaces/${wsId}/webhooks`),
          getWorkspace(wsId),
        ])
        if (m.status === 'fulfilled') setMembers(m.value.members || [])
        if (l.status === 'fulfilled') setLabels(l.value.labels || [])
        if (k.status === 'fulfilled') setKeys(k.value.api_keys || [])
        if (h.status === 'fulfilled') setHooks(h.value.webhooks || [])
        if (ws.status === 'fulfilled') {
          const w = ws.value.workspace
          setMyRole(w.role || '')
          setWsName(w.name || '')
          setWsSlug(w.slug || '')
          wsForm.setFieldsValue({ name: w.name, slug: w.slug })
        } else {
          // 兜底：listMine 匹配（若 get 失败）
          try {
            const mine = await listMyWorkspaces()
            const hit = mine.workspaces.find((x) => x.id === wsId)
            if (hit) {
              setMyRole(hit.role || '')
              setWsName(hit.name || '')
              setWsSlug(hit.slug || '')
              wsForm.setFieldsValue({ name: hit.name, slug: hit.slug })
            }
          } catch {
            /* ignore */
          }
        }
        const firstErr = [m, l, k, h].find((x) => x.status === 'rejected') as
          | PromiseRejectedResult
          | undefined
        if (firstErr) setErr(firstErr.reason?.message || '部分设置接口尚未就绪')
      } catch (e: unknown) {
        const msg = e instanceof Error ? e.message : '加载失败'
        setErr(msg)
        message.error(msg)
      } finally {
        setLoading(false)
      }
    })()
  }, [wsId, wsForm])

  const onInvite = async () => {
    try {
      const values = await inviteForm.validateFields()
      setInviting(true)
      await inviteMember(wsId, {
        email: values.email,
        role: values.role,
        display_name: values.display_name,
      })
      message.success('已邀请成员')
      inviteForm.resetFields()
      inviteForm.setFieldsValue({ role: 'member' })
      await reloadMembers()
    } catch (e) {
      if (e && typeof e === 'object' && 'errorFields' in e) return
      message.error(e instanceof ApiError || e instanceof Error ? e.message : '邀请失败')
    } finally {
      setInviting(false)
    }
  }

  const onChangeRole = async (row: Member, nextRole: string) => {
    if (nextRole === row.role) return
    const isLastAdmin =
      adminCount <= 1 && (row.role === 'admin' || row.role === 'owner') && nextRole !== 'admin'
    if (isLastAdmin) {
      message.warning('不能将唯一管理员降级')
      return
    }
    try {
      await patchMember(wsId, row.id, { role: nextRole })
      message.success('角色已更新')
      await reloadMembers()
    } catch (e) {
      message.error(e instanceof ApiError || e instanceof Error ? e.message : '更新失败')
    }
  }

  const onRemove = async (row: Member) => {
    if ((row.role === 'admin' || row.role === 'owner') && adminCount <= 1) {
      message.warning('不能移除唯一管理员')
      return
    }
    try {
      await removeMember(wsId, row.id)
      message.success('已移除成员')
      await reloadMembers()
    } catch (e) {
      message.error(e instanceof ApiError || e instanceof Error ? e.message : '移除失败')
    }
  }

  const onSaveWorkspace = async () => {
    try {
      const values = await wsForm.validateFields()
      setWsSaving(true)
      const resp = await patchWorkspace(wsId, {
        name: values.name.trim(),
        slug: values.slug.trim().toLowerCase(),
      })
      setWsName(resp.workspace.name)
      setWsSlug(resp.workspace.slug)
      message.success('工作区已更新')
    } catch (e) {
      if (e && typeof e === 'object' && 'errorFields' in e) return
      const status = e instanceof ApiError ? e.status : 0
      if (status === 404 || status === 405 || status === 501) {
        message.info('改名/slug 即将支持，后端接口尚未就绪')
      } else {
        message.error(e instanceof ApiError || e instanceof Error ? e.message : '保存失败')
      }
    } finally {
      setWsSaving(false)
    }
  }

  if (loading) {
    return (
      <div className="acb-page" style={{ textAlign: 'center', paddingTop: 48 }}>
        <Spin />
      </div>
    )
  }

  const empty = (description: string) => (
    <div className="acb-empty-panel">
      <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={description} />
    </div>
  )

  const memberColumns = [
    {
      title: '名称',
      dataIndex: 'display_name',
      render: (v: string, row: Member) => (
        <span>
          {v || '未命名'}
          {isBotMember(row) ? (
            <Tag color="purple" style={{ marginLeft: 8 }}>
              机器人
            </Tag>
          ) : null}
        </span>
      ),
    },
    {
      title: '邮箱',
      dataIndex: 'email',
      ellipsis: true,
      render: (v?: string | null) => v || '—',
    },
    {
      title: '角色',
      dataIndex: 'role',
      width: isAdmin ? 140 : 110,
      render: (v: string, row: Member) => {
        if (!isAdmin || isBotMember(row)) return roleLabel(v)
        const soleAdmin =
          adminCount <= 1 && (row.role === 'admin' || row.role === 'owner')
        return (
          <Select
            size="small"
            style={{ width: 112 }}
            value={v === 'owner' ? 'admin' : v}
            options={[
              { value: 'admin', label: '管理员' },
              { value: 'member', label: '成员' },
            ]}
            onChange={(next) => void onChangeRole(row, next)}
            disabled={soleAdmin}
            title={soleAdmin ? '不能将唯一管理员降级' : undefined}
          />
        )
      },
    },
    ...(isAdmin
      ? [
          {
            title: '操作',
            key: 'actions',
            width: 88,
            render: (_: unknown, row: Member) => {
              if (isBotMember(row)) {
                return (
                  <Typography.Text type="secondary" style={{ fontSize: 12 }}>
                    只读
                  </Typography.Text>
                )
              }
              const soleAdmin =
                adminCount <= 1 && (row.role === 'admin' || row.role === 'owner')
              return (
                <Popconfirm
                  title="确认移除此成员？"
                  okText="移除"
                  cancelText="取消"
                  disabled={soleAdmin}
                  onConfirm={() => void onRemove(row)}
                >
                  <Button type="link" size="small" danger disabled={soleAdmin}>
                    移除
                  </Button>
                </Popconfirm>
              )
            },
          },
        ]
      : []),
  ]

  return (
    <div className="acb-page">
      <div className="acb-page-bar">
        <div>
          <Typography.Title level={4} style={{ margin: 0 }}>
            设置
          </Typography.Title>
          <Typography.Text type="secondary" style={{ fontSize: 12 }}>
            工作区、成员、标签与接入凭据
            {myRole ? ` · 当前角色 ${roleLabel(myRole)}` : ''}
          </Typography.Text>
        </div>
      </div>
      {err && (
        <Typography.Paragraph type="warning" style={{ marginBottom: 8 }}>
          {err}
        </Typography.Paragraph>
      )}
      <div className="acb-surface-panel acb-settings-panel">
        <Tabs
          items={[
            {
              key: 'workspace',
              label: '工作区',
              children: (
                <>
                  {isAdmin ? (
                    <Form
                      form={wsForm}
                      layout="vertical"
                      style={{ maxWidth: 420 }}
                      initialValues={{ name: wsName, slug: wsSlug }}
                      onValuesChange={(changed) => {
                        if ('name' in changed && !wsForm.isFieldTouched('slug')) {
                          const auto = slugifyName(changed.name || '')
                          if (auto) wsForm.setFieldsValue({ slug: auto })
                        }
                      }}
                    >
                      <Form.Item
                        name="name"
                        label="名称"
                        rules={[{ required: true, message: '请填写名称' }]}
                      >
                        <Input maxLength={80} />
                      </Form.Item>
                      <Form.Item
                        name="slug"
                        label="标识（slug）"
                        rules={[
                          { required: true, message: '请填写 slug' },
                          {
                            pattern: /^[a-z0-9]+(?:-[a-z0-9]+)*$/,
                            message: '仅允许小写字母、数字与连字符',
                          },
                        ]}
                      >
                        <Input maxLength={64} />
                      </Form.Item>
                      <Button type="primary" loading={wsSaving} onClick={() => void onSaveWorkspace()}>
                        保存
                      </Button>
                      <Typography.Paragraph type="secondary" style={{ marginTop: 12, fontSize: 12 }}>
                        改名与 slug 需工作区管理员；若后端尚未开放 PATCH，将提示即将支持。
                      </Typography.Paragraph>
                    </Form>
                  ) : (
                    <div>
                      <Typography.Paragraph style={{ marginBottom: 4 }}>
                        <Typography.Text type="secondary">名称</Typography.Text>
                        <br />
                        {wsName || '—'}
                      </Typography.Paragraph>
                      <Typography.Paragraph>
                        <Typography.Text type="secondary">标识</Typography.Text>
                        <br />
                        {wsSlug || '—'}
                      </Typography.Paragraph>
                      <Alert type="info" showIcon message="仅管理员可修改工作区名称与标识" />
                    </div>
                  )}
                </>
              ),
            },
            {
              key: 'members',
              label: `成员 (${members.length})`,
              children: (
                <>
                  {isAdmin ? (
                    <div style={{ marginBottom: 16 }}>
                      <Typography.Text strong style={{ display: 'block', marginBottom: 8 }}>
                        邀请成员
                      </Typography.Text>
                      <Form
                        form={inviteForm}
                        layout="inline"
                        initialValues={{ role: 'member' }}
                        style={{ rowGap: 8, flexWrap: 'wrap' }}
                      >
                        <Form.Item
                          name="email"
                          rules={[
                            { required: true, message: '请填写邮箱' },
                            { type: 'email', message: '邮箱格式不正确' },
                          ]}
                        >
                          <Input placeholder="已注册邮箱" style={{ width: 220 }} />
                        </Form.Item>
                        <Form.Item name="role" rules={[{ required: true }]}>
                          <Select
                            style={{ width: 120 }}
                            options={[
                              { value: 'admin', label: '管理员' },
                              { value: 'member', label: '成员' },
                            ]}
                          />
                        </Form.Item>
                        <Form.Item name="display_name">
                          <Input placeholder="显示名（可选）" style={{ width: 140 }} />
                        </Form.Item>
                        <Form.Item>
                          <Button type="primary" loading={inviting} onClick={() => void onInvite()}>
                            邀请
                          </Button>
                        </Form.Item>
                      </Form>
                      <Typography.Paragraph type="secondary" style={{ marginTop: 8, fontSize: 12 }}>
                        目前仅支持邀请已注册账号。人类邀请链接即将支持（与 Bot 入场码隔离）。
                      </Typography.Paragraph>
                    </div>
                  ) : null}

                  {members.length ? (
                    <Table
                      className="acb-list-table"
                      size="small"
                      rowKey="id"
                      dataSource={members}
                      pagination={false}
                      columns={memberColumns}
                    />
                  ) : (
                    empty('暂无成员或接口未就绪')
                  )}
                  <div className="acb-settings-tab-foot">
                    成员含人类与 Bot。Bot 行只读；席位与入场码请到侧栏「Bot」管理。
                    {!isAdmin ? ' 你当前不是管理员，成员表只读。' : ''}
                  </div>
                </>
              ),
            },
            {
              key: 'labels',
              label: `标签 (${labels.length})`,
              children: labels.length ? (
                <Table
                  className="acb-list-table"
                  size="small"
                  rowKey="id"
                  dataSource={labels}
                  pagination={false}
                  columns={[
                    { title: '名称', dataIndex: 'name' },
                    {
                      title: '颜色',
                      dataIndex: 'color',
                      width: 120,
                      render: (v: string) => (
                        <span style={{ display: 'inline-flex', alignItems: 'center', gap: 8 }}>
                          <span
                            aria-hidden
                            style={{
                              width: 12,
                              height: 12,
                              borderRadius: 3,
                              background: v || 'var(--acb-border)',
                              border: '1px solid var(--acb-border)',
                            }}
                          />
                          <Typography.Text type="secondary" style={{ fontSize: 12 }}>
                            {v || '—'}
                          </Typography.Text>
                        </span>
                      ),
                    },
                  ]}
                />
              ) : (
                empty('暂无标签或接口未就绪')
              ),
            },
            {
              key: 'keys',
              label: 'API Key',
              children: keys.length ? (
                <Table
                  className="acb-list-table"
                  size="small"
                  rowKey="id"
                  dataSource={keys}
                  pagination={false}
                  columns={[
                    { title: '名称', dataIndex: 'name' },
                    {
                      title: 'Key',
                      dataIndex: 'key_prefix',
                      width: 120,
                      render: (v: string) => keyTail(v),
                    },
                    {
                      title: '创建',
                      dataIndex: 'created_at',
                      width: 140,
                      render: (v?: string) => (v ? formatShanghai(v) : '—'),
                    },
                  ]}
                />
              ) : (
                empty('暂无 API Key 或接口未就绪')
              ),
            },
            {
              key: 'webhooks',
              label: 'Webhook',
              children: hooks.length ? (
                <Table
                  className="acb-list-table"
                  size="small"
                  rowKey="id"
                  dataSource={hooks}
                  pagination={false}
                  columns={[
                    { title: 'URL', dataIndex: 'url', ellipsis: true },
                    {
                      title: '启用',
                      dataIndex: 'enabled',
                      width: 72,
                      render: (v) => (v ? '是' : '否'),
                    },
                  ]}
                />
              ) : (
                empty('暂无 Webhook 或接口未就绪')
              ),
            },
          ]}
        />
      </div>
    </div>
  )
}
