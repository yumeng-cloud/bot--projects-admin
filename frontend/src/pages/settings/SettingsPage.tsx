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
  Card,
} from 'antd'
import { CopyOutlined } from '@ant-design/icons'
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
import {
  createWorkspaceInvite,
  listWorkspaceInvites,
  revokeWorkspaceInvite,
  workspaceInviteStatusCn,
  workspaceInviteTail,
  type WorkspaceInvite,
} from '../../api/workspaceInvites'
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

function expiresPresetToRfc3339(preset: string): string | undefined {
  if (!preset || preset === 'never') return undefined
  const days = Number(preset)
  if (!Number.isFinite(days) || days <= 0) return undefined
  const d = new Date(Date.now() + days * 24 * 60 * 60 * 1000)
  return d.toISOString()
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

  const [linkForm] = Form.useForm<{
    email?: string
    role: 'admin' | 'member'
    expires: string
  }>()
  const [linkBusy, setLinkBusy] = useState(false)
  const [wsInvites, setWsInvites] = useState<WorkspaceInvite[]>([])
  const [lastPlainCode, setLastPlainCode] = useState<string | null>(null)

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

  const reloadInvites = useCallback(async () => {
    if (!wsId || !isAdmin) return
    try {
      const data = await listWorkspaceInvites(wsId)
      setWsInvites(data.invites || [])
    } catch {
      /* 非 admin / 暂无列表时静默 */
    }
  }, [wsId, isAdmin])

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
        let role = ''
        if (ws.status === 'fulfilled') {
          const data = ws.value
          const w = data.workspace
          role = data.role || w?.role || ''
          setMyRole(role)
          setWsName(w?.name || '')
          setWsSlug(w?.slug || '')
          wsForm.setFieldsValue({ name: w?.name, slug: w?.slug })
        } else {
          try {
            const mine = await listMyWorkspaces()
            const hit = mine.workspaces.find((x) => x.id === wsId)
            if (hit) {
              role = hit.role || ''
              setMyRole(role)
              setWsName(hit.name || '')
              setWsSlug(hit.slug || '')
              wsForm.setFieldsValue({ name: hit.name, slug: hit.slug })
            }
          } catch {
            /* ignore */
          }
        }
        if (role === 'admin' || role === 'owner') {
          try {
            const inv = await listWorkspaceInvites(wsId)
            setWsInvites(inv.invites || [])
          } catch {
            setWsInvites([])
          }
        }
        const firstErr = [m, l, k, h].find((x) => x.status === 'rejected') as
          | PromiseRejectedResult
          | undefined
        if (firstErr) setErr(firstErr.reason?.message || '部分设置接口加载失败')
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

  const onCreateLink = async () => {
    try {
      const values = await linkForm.validateFields()
      setLinkBusy(true)
      const { invite, code } = await createWorkspaceInvite(
        wsId,
        {
          email: values.email?.trim() || undefined,
          role: values.role,
          expires_at: expiresPresetToRfc3339(values.expires),
        },
        crypto.randomUUID(),
      )
      const plain = code || invite.code || ''
      if (plain) {
        setLastPlainCode(plain)
        try {
          await navigator.clipboard.writeText(plain)
          message.success(`已签发并复制邀请码（末四 ·${workspaceInviteTail(invite)}）`)
        } catch {
          message.success(`已签发邀请码（末四 ·${workspaceInviteTail(invite)}），请手动复制`)
        }
      } else {
        message.success('已签发邀请')
      }
      setWsInvites((prev) => [
        {
          ...invite,
          code: undefined,
          code_tail: workspaceInviteTail(invite),
          status: invite.status || 'active',
        },
        ...prev.filter((x) => x.id !== invite.id),
      ])
      linkForm.resetFields()
      linkForm.setFieldsValue({ role: 'member', expires: 'never' })
      await reloadInvites()
    } catch (e) {
      if (e && typeof e === 'object' && 'errorFields' in e) return
      message.error(e instanceof ApiError || e instanceof Error ? e.message : '签发失败')
    } finally {
      setLinkBusy(false)
    }
  }

  const onRevokeLink = async (inv: WorkspaceInvite) => {
    try {
      const { invite } = await revokeWorkspaceInvite(wsId, inv.id, crypto.randomUUID())
      setWsInvites((prev) =>
        prev.map((x) =>
          x.id === inv.id
            ? {
                ...x,
                ...invite,
                status: invite.status || 'revoked',
                revoked_at: invite.revoked_at || x.revoked_at || new Date().toISOString(),
              }
            : x,
        ),
      )
      message.success(`已吊销邀请（末四 ·${workspaceInviteTail(invite.status ? invite : inv)}）`)
    } catch (e) {
      message.error(e instanceof ApiError || e instanceof Error ? e.message : '吊销失败')
    }
  }

  const onCopyPlain = async () => {
    if (!lastPlainCode) return
    try {
      await navigator.clipboard.writeText(lastPlainCode)
      message.success('已复制邀请码')
    } catch {
      message.warning('复制失败，请手动选中复制')
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
      message.error(e instanceof ApiError || e instanceof Error ? e.message : '保存失败')
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
                        仅工作区管理员可改名与标识。保存后立即生效。
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
                        邀请已注册成员
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
                        对方须已注册；未注册请用下方「人类邀请链接」。
                      </Typography.Paragraph>

                      <Typography.Text strong style={{ display: 'block', marginBottom: 8, marginTop: 16 }}>
                        人类邀请链接
                      </Typography.Text>
                      <Form
                        form={linkForm}
                        layout="inline"
                        initialValues={{ role: 'member', expires: 'never' }}
                        style={{ rowGap: 8, flexWrap: 'wrap', marginBottom: 8 }}
                      >
                        <Form.Item
                          name="email"
                          rules={[{ type: 'email', message: '邮箱格式不正确' }]}
                        >
                          <Input placeholder="邮箱（可选，未注册亦可）" style={{ width: 220 }} />
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
                        <Form.Item name="expires">
                          <Select
                            style={{ width: 140 }}
                            options={[
                              { value: 'never', label: '不过期' },
                              { value: '7', label: '7 天后' },
                              { value: '30', label: '30 天后' },
                            ]}
                          />
                        </Form.Item>
                        <Form.Item>
                          <Button
                            type="primary"
                            loading={linkBusy}
                            onClick={() => void onCreateLink()}
                          >
                            签发邀请码
                          </Button>
                        </Form.Item>
                      </Form>
                      <Typography.Paragraph type="secondary" style={{ fontSize: 12, marginBottom: 8 }}>
                        签发码以 <Typography.Text code>ymw_</Typography.Text> 开头，与 Bot
                        队席码 <Typography.Text code>ymt_</Typography.Text> 隔离。明文仅展示一次。
                      </Typography.Paragraph>

                      {lastPlainCode ? (
                        <Alert
                          type="success"
                          showIcon
                          style={{ marginBottom: 12 }}
                          message="邀请码（仅此一次，请妥善转发）"
                          description={
                            <Space wrap>
                              <Typography.Text code copyable={{ text: lastPlainCode }}>
                                {lastPlainCode}
                              </Typography.Text>
                              <Button size="small" icon={<CopyOutlined />} onClick={() => void onCopyPlain()}>
                                复制
                              </Button>
                            </Space>
                          }
                        />
                      ) : null}

                      {wsInvites.length ? (
                        <Card size="small" title="已签发邀请" style={{ marginBottom: 12 }}>
                          <Space direction="vertical" style={{ width: '100%' }} size={8}>
                            {wsInvites.slice(0, 20).map((inv) => {
                              const st = (inv.status || 'active').toLowerCase()
                              const canRevoke = st === 'active'
                              return (
                                <div
                                  key={inv.id}
                                  style={{
                                    display: 'flex',
                                    flexWrap: 'wrap',
                                    gap: 8,
                                    alignItems: 'center',
                                    justifyContent: 'space-between',
                                  }}
                                >
                                  <Space wrap size={8}>
                                    <Typography.Text type="secondary" style={{ fontSize: 12 }}>
                                      末四 ·{workspaceInviteTail(inv)}
                                    </Typography.Text>
                                    <Tag>{workspaceInviteStatusCn(st)}</Tag>
                                    <Tag>{roleLabel(inv.role)}</Tag>
                                    {inv.email ? (
                                      <Typography.Text style={{ fontSize: 12 }}>{inv.email}</Typography.Text>
                                    ) : (
                                      <Typography.Text type="secondary" style={{ fontSize: 12 }}>
                                        不限邮箱
                                      </Typography.Text>
                                    )}
                                    {inv.expires_at ? (
                                      <Typography.Text type="secondary" style={{ fontSize: 12 }}>
                                        过期 {formatShanghai(inv.expires_at)}
                                      </Typography.Text>
                                    ) : null}
                                    <Typography.Text type="secondary" style={{ fontSize: 12 }}>
                                      使用 {inv.use_count ?? 0}
                                      {inv.max_uses != null ? `/${inv.max_uses}` : ''}
                                    </Typography.Text>
                                  </Space>
                                  {canRevoke ? (
                                    <Popconfirm
                                      title="确认吊销此邀请？"
                                      okText="吊销"
                                      cancelText="取消"
                                      onConfirm={() => void onRevokeLink(inv)}
                                    >
                                      <Button size="small" danger type="link">
                                        吊销
                                      </Button>
                                    </Popconfirm>
                                  ) : null}
                                </div>
                              )
                            })}
                          </Space>
                        </Card>
                      ) : (
                        <Typography.Paragraph type="secondary" style={{ fontSize: 12 }}>
                          暂无已签发的人类邀请。
                        </Typography.Paragraph>
                      )}
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
                    empty('暂无成员')
                  )}
                  <div className="acb-settings-tab-foot">
                    成员含人类与 Bot。Bot 行只读；席位与入场码请到侧栏「Bot」管理。
                    兑换人类邀请码可在首页「兑换邀请码」。
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
                empty('暂无标签')
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
                empty('暂无 API Key')
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
                empty('暂无 Webhook')
              ),
            },
          ]}
        />
      </div>
    </div>
  )
}
