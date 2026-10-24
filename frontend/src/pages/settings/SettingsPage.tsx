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
  Modal,
  Dropdown,
  type MenuProps,
} from 'antd'
import { CopyOutlined, MoreOutlined, PlusOutlined } from '@ant-design/icons'
import { useParams, useSearchParams } from 'react-router-dom'
import SettingsProjectsPanel from './SettingsProjectsPanel'
import SettingsTeamsPanel from './SettingsTeamsPanel'
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
import { createApiKey, listApiKeys } from '../../api/apiKeys'
import IdentityKeyRevealModal from '../../components/common/IdentityKeyRevealModal'
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

/** D1：角色中文芯片（管理员/成员/机器人）；owner 视同管理员 */
function memberRoleChip(m: Member) {
  if (isBotMember(m)) {
    return (
      <Tag color="purple" style={{ marginInlineEnd: 0 }}>
        机器人
      </Tag>
    )
  }
  const r = m.role === 'owner' ? 'admin' : m.role
  if (r === 'admin') {
    return (
      <Tag color="processing" style={{ marginInlineEnd: 0 }}>
        管理员
      </Tag>
    )
  }
  return (
    <Tag style={{ marginInlineEnd: 0 }}>
      成员
    </Tag>
  )
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
  const [searchParams, setSearchParams] = useSearchParams()
  const tabParam = searchParams.get('tab') || 'workspace'
  const activeTab =
    tabParam === 'projects' ||
    tabParam === 'teams' ||
    tabParam === 'members' ||
    tabParam === 'labels' ||
    tabParam === 'integrations' ||
    tabParam === 'workspace'
      ? tabParam
      : 'workspace'
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

  const [labelOpen, setLabelOpen] = useState(false)
  const [labelForm] = Form.useForm<{ name: string; color: string }>()
  const [labelBusy, setLabelBusy] = useState(false)

  const [keyOpen, setKeyOpen] = useState(false)
  const [keyForm] = Form.useForm<{ name: string }>()
  const [keyBusy, setKeyBusy] = useState(false)
  const [lastKeyToken, setLastKeyToken] = useState<string | null>(null)

  const [hookOpen, setHookOpen] = useState(false)
  const [hookForm] = Form.useForm<{ url: string }>()
  const [hookBusy, setHookBusy] = useState(false)

  const isAdmin = myRole === 'admin' || myRole === 'owner'

  const adminCount = useMemo(
    () => members.filter((m) => !isBotMember(m) && (m.role === 'admin' || m.role === 'owner')).length,
    [members],
  )

  /** B2：人类成员池；qaEmpty=1 强制空态供箱截 */
  const forceMembersEmpty = searchParams.get('qaEmpty') === '1'
  const humanMembers = useMemo(
    () => (forceMembersEmpty ? [] : members.filter((m) => !isBotMember(m))),
    [members, forceMembersEmpty],
  )
  const botMembers = useMemo(
    () => (forceMembersEmpty ? [] : members.filter((m) => isBotMember(m))),
    [members, forceMembersEmpty],
  )
  const membersEmpty = humanMembers.length === 0

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
      /* 非管理员 / 暂无列表时静默 */
    }
  }, [wsId, isAdmin])

  const reloadLabels = useCallback(async () => {
    if (!wsId) return
    try {
      const data = await api<{ labels: Label[] }>(`/workspaces/${wsId}/labels`)
      setLabels(data.labels || [])
    } catch {
      /* ignore */
    }
  }, [wsId])

  const reloadKeys = useCallback(async () => {
    if (!wsId) return
    try {
      const data = await listApiKeys(wsId)
      setKeys(data.api_keys || [])
    } catch {
      /* ignore */
    }
  }, [wsId])

  const reloadHooks = useCallback(async () => {
    if (!wsId) return
    try {
      const data = await api<{ webhooks: Webhook[] }>(`/workspaces/${wsId}/webhooks`)
      setHooks(data.webhooks || [])
    } catch {
      /* ignore */
    }
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
      const body: { name: string; slug?: string } = { name: values.name.trim() }
      if (isAdmin && values.slug) body.slug = values.slug.trim().toLowerCase()
      const resp = await patchWorkspace(wsId, body)
      setWsName(resp.workspace.name)
      setWsSlug(resp.workspace.slug)
      wsForm.setFieldsValue({ name: resp.workspace.name, slug: resp.workspace.slug })
      message.success('工作区已更新')
    } catch (e) {
      if (e && typeof e === 'object' && 'errorFields' in e) return
      message.error(e instanceof ApiError || e instanceof Error ? e.message : '保存失败')
    } finally {
      setWsSaving(false)
    }
  }

  const onCreateLabel = async () => {
    try {
      const values = await labelForm.validateFields()
      setLabelBusy(true)
      await api(`/workspaces/${wsId}/labels`, {
        method: 'POST',
        body: JSON.stringify({
          name: values.name.trim(),
          color: values.color || '#5E6AD2',
        }),
      })
      message.success('已新建标签')
      setLabelOpen(false)
      labelForm.resetFields()
      await reloadLabels()
    } catch (e) {
      if (e && typeof e === 'object' && 'errorFields' in e) return
      message.error(e instanceof ApiError || e instanceof Error ? e.message : '新建失败')
    } finally {
      setLabelBusy(false)
    }
  }

  const onCreateKey = async () => {
    try {
      const values = await keyForm.validateFields()
      setKeyBusy(true)
      const res = await createApiKey(wsId, { name: values.name.trim() }, crypto.randomUUID())
      const token = (res.token || '').trim()
      setKeyOpen(false)
      keyForm.resetFields()
      await reloadKeys()
      if (!token) {
        message.error('已创建但服务未返回身份钥匙明文，请重试或联系管理员')
        return
      }
      // FOOLPROOF F2/F4：挡层 Modal 露钥，禁只 Toast / 埋 Alert
      setLastKeyToken(token)
    } catch (e) {
      if (e && typeof e === 'object' && 'errorFields' in e) return
      message.error(e instanceof ApiError || e instanceof Error ? e.message : '新建失败')
    } finally {
      setKeyBusy(false)
    }
  }

  const onCreateHook = async () => {
    try {
      const values = await hookForm.validateFields()
      setHookBusy(true)
      await api(`/workspaces/${wsId}/webhooks`, {
        method: 'POST',
        body: JSON.stringify({ url: values.url.trim() }),
      })
      message.success('已新建 Webhook')
      setHookOpen(false)
      hookForm.resetFields()
      await reloadHooks()
    } catch (e) {
      if (e && typeof e === 'object' && 'errorFields' in e) return
      message.error(e instanceof ApiError || e instanceof Error ? e.message : '新建失败')
    } finally {
      setHookBusy(false)
    }
  }

  if (loading) {
    return (
      <div className="acb-page acb-pro-admin" style={{ textAlign: 'center', paddingTop: 48 }}>
        <Spin />
      </div>
    )
  }

  const empty = (description: string) => (
    <div className="acb-empty-panel" style={{ minHeight: 120, padding: '20px 16px' }}>
      <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={description} />
    </div>
  )

  const memberColumns = [
    {
      title: '名称',
      dataIndex: 'display_name',
      render: (v: string) => v || '未命名',
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
      width: 100,
      render: (_: string, row: Member) => memberRoleChip(row),
    },
    ...(isAdmin
      ? [
          {
            title: '操作',
            key: 'actions',
            width: 56,
            align: 'right' as const,
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
              const cur = row.role === 'owner' ? 'admin' : row.role
              const items: MenuProps['items'] = [
                {
                  key: 'to-admin',
                  label: '设为管理员',
                  disabled: cur === 'admin',
                  onClick: () => void onChangeRole(row, 'admin'),
                },
                {
                  key: 'to-member',
                  label: '设为成员',
                  disabled: cur === 'member' || soleAdmin,
                  title: soleAdmin ? '不能将唯一管理员降级' : undefined,
                  onClick: () => void onChangeRole(row, 'member'),
                },
                { type: 'divider' },
                {
                  key: 'remove',
                  label: '移除',
                  danger: true,
                  disabled: soleAdmin,
                  onClick: () => {
                    if (soleAdmin) {
                      message.warning('不能移除唯一管理员')
                      return
                    }
                    Modal.confirm({
                      title: '确认移除此成员？',
                      okText: '移除',
                      okButtonProps: { danger: true },
                      cancelText: '取消',
                      onOk: () => onRemove(row),
                    })
                  },
                },
              ]
              return (
                <Dropdown menu={{ items }} trigger={['click']} placement="bottomRight">
                  <Button
                    type="text"
                    size="small"
                    icon={<MoreOutlined />}
                    aria-label="更多"
                  />
                </Dropdown>
              )
            },
          },
        ]
      : []),
  ]

  const roleOptions = [
    { value: 'admin' as const, label: '管理员' },
    { value: 'member' as const, label: '成员' },
  ]

  return (
    <div className="acb-page acb-pro-admin">
      <div className="acb-page-bar">
        <div>
          <Typography.Title level={4} className="acb-page-title">
            设置
          </Typography.Title>
          <Typography.Text type="secondary" className="acb-page-sub">
            工作区 · 项目 · 团队 · 成员 · 标签 · 接入
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
          activeKey={activeTab}
          onChange={(key) => {
            const next = new URLSearchParams(searchParams)
            if (key === 'workspace') next.delete('tab')
            else next.set('tab', key)
            setSearchParams(next, { replace: true })
          }}
          items={[
            {
              key: 'workspace',
              label: '工作区',
              children: (
                <>
                  <p className="acb-settings-hint">修改工作区显示名称；管理员还可调整 URL 标识。</p>
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
                        style={{ marginBottom: 12 }}
                      >
                        <Input maxLength={80} />
                      </Form.Item>
                      <Form.Item
                        name="slug"
                        label={
                          <span>
                            URL 标识
                            <span className="acb-settings-slug-extra">用于 URL</span>
                          </span>
                        }
                        rules={[
                          { required: true, message: '请填写 URL 标识' },
                          {
                            pattern: /^[a-z0-9]+(?:-[a-z0-9]+)*$/,
                            message: '仅允许小写字母、数字与连字符',
                          },
                        ]}
                        style={{ marginBottom: 16 }}
                      >
                        <Input maxLength={64} />
                      </Form.Item>
                      <Button type="primary" loading={wsSaving} onClick={() => void onSaveWorkspace()}>
                        保存
                      </Button>
                    </Form>
                  ) : (
                    <div>
                      <Typography.Paragraph style={{ marginBottom: 4 }}>
                        <Typography.Text type="secondary">名称</Typography.Text>
                        <br />
                        {wsName || '—'}
                      </Typography.Paragraph>
                      <Alert type="info" showIcon message="仅管理员可修改工作区名称与 URL 标识" />
                    </div>
                  )}
                </>
              ),
            },
            {
              key: 'projects',
              label: '项目',
              children: <SettingsProjectsPanel workspaceId={wsId} isAdmin={isAdmin} />,
            },
            {
              key: 'teams',
              label: '团队',
              children: <SettingsTeamsPanel workspaceId={wsId} />,
            },
            {
              key: 'members',
              label: `成员 (${humanMembers.length})`,
              children: (
                <>
                  {isAdmin ? (
                    <div className="acb-settings-section">
                      <div className="acb-settings-section-head">
                        <span className="acb-settings-section-title">邀请</span>
                      </div>
                      <Typography.Text
                        id="acb-invite-email"
                        type="secondary"
                        style={{ display: 'block', marginBottom: 8, fontSize: 12 }}
                      >
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
                          <Select style={{ width: 120 }} options={roleOptions} />
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
                      <Typography.Paragraph type="secondary" style={{ marginTop: 8, fontSize: 12, marginBottom: 0 }}>
                        对方须已注册；未注册请用下方「人类邀请链接」。
                      </Typography.Paragraph>

                      <Typography.Text
                        id="acb-invite-ymw"
                        type="secondary"
                        style={{ display: 'block', marginBottom: 8, marginTop: 16, fontSize: 12 }}
                      >
                        人类邀请链接
                      </Typography.Text>
                      <Form
                        form={linkForm}
                        layout="inline"
                        initialValues={{ role: 'member', expires: 'never' }}
                        style={{ rowGap: 8, flexWrap: 'wrap', marginBottom: 8 }}
                      >
                        <Form.Item name="email" rules={[{ type: 'email', message: '邮箱格式不正确' }]}>
                          <Input placeholder="邮箱（可选，未注册亦可）" style={{ width: 220 }} />
                        </Form.Item>
                        <Form.Item name="role" rules={[{ required: true }]}>
                          <Select style={{ width: 120 }} options={roleOptions} />
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
                          <Button type="primary" loading={linkBusy} onClick={() => void onCreateLink()}>
                            签发邀请码
                          </Button>
                        </Form.Item>
                      </Form>
                      <Typography.Paragraph type="secondary" style={{ fontSize: 12, marginBottom: 8 }}>
                        签发码以 <Typography.Text code>ymw_</Typography.Text> 开头，与 Bot 队席码{' '}
                        <Typography.Text code>ymt_</Typography.Text> 隔离。明文仅展示一次。
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
                        <Card size="small" title="已签发邀请" style={{ marginBottom: 0 }}>
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
                        <Typography.Paragraph type="secondary" style={{ fontSize: 12, marginBottom: 0 }}>
                          暂无已签发的人类邀请。
                        </Typography.Paragraph>
                      )}
                    </div>
                  ) : null}

                  <div className="acb-settings-section">
                    <div className="acb-settings-section-head">
                      <span className="acb-settings-section-title">成员</span>
                    </div>

                    {membersEmpty ? (
                      <div
                        className="acb-empty-panel"
                        style={{ minHeight: 120, padding: '20px 16px', textAlign: 'center' }}
                      >
                        <Empty
                          image={Empty.PRESENTED_IMAGE_SIMPLE}
                          description={
                            <span>
                              <Typography.Text
                                strong
                                style={{ display: 'block', fontSize: 15, marginBottom: 4 }}
                              >
                                还没有成员
                              </Typography.Text>
                              <Typography.Text type="secondary" style={{ fontSize: 13 }}>
                                {isAdmin
                                  ? '邀请同事加入，或签发人类邀请码（ymw_ 开头）'
                                  : '请联系管理员邀请成员加入'}
                              </Typography.Text>
                            </span>
                          }
                        >
                          {isAdmin ? (
                            <Space wrap>
                              <Button
                                type="primary"
                                onClick={() =>
                                  document.getElementById('acb-invite-email')?.scrollIntoView({
                                    behavior: 'smooth',
                                    block: 'center',
                                  })
                                }
                              >
                                邀请成员
                              </Button>
                              <Button
                                onClick={() =>
                                  document.getElementById('acb-invite-ymw')?.scrollIntoView({
                                    behavior: 'smooth',
                                    block: 'center',
                                  })
                                }
                              >
                                签发人类邀请
                              </Button>
                            </Space>
                          ) : null}
                        </Empty>
                      </div>
                    ) : null}

                    {humanMembers.length || botMembers.length ? (
                      <Table
                        className="acb-list-table"
                        size="small"
                        rowKey="id"
                        dataSource={[...humanMembers, ...botMembers]}
                        pagination={false}
                        columns={memberColumns}
                        style={membersEmpty ? { marginTop: 16 } : undefined}
                      />
                    ) : null}
                    {/* Bot 席脚注与上方人类邀请空态隔离：不写入空态主文案 */}
                    <div className="acb-settings-tab-foot">
                      兑换人类邀请码可在首页「兑换邀请码」。
                      {!isAdmin ? ' 你当前不是管理员，成员表只读。' : ''}
                      {botMembers.length > 0 || !membersEmpty
                        ? ' 机器人席位请到侧栏「Bot」管理（与人类邀请 ymw_ 分家）。'
                        : ''}
                    </div>
                  </div>
                </>
              ),
            },
            {
              key: 'labels',
              label: `标签 (${labels.length})`,
              children: (
                <>
                  <div className="acb-settings-section-head">
                    <span className="acb-settings-section-title">标签</span>
                    {isAdmin ? (
                      <Button
                        type="primary"
                        size="small"
                        icon={<PlusOutlined />}
                        onClick={() => {
                          labelForm.setFieldsValue({ name: '', color: '#5E6AD2' })
                          setLabelOpen(true)
                        }}
                      >
                        新建
                      </Button>
                    ) : null}
                  </div>
                  {labels.length ? (
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
                    empty('还没有标签')
                  )}
                </>
              ),
            },
            {
              key: 'integrations',
              label: '接入',
              children: (
                <>
                  <div className="acb-settings-section">
                    <div className="acb-settings-section-head">
                      <span className="acb-settings-section-title">身份钥匙</span>
                      {isAdmin ? (
                        <Button
                          type="primary"
                          size="small"
                          icon={<PlusOutlined />}
                          onClick={() => {
                            keyForm.resetFields()
                            setKeyOpen(true)
                          }}
                        >
                          新建身份钥匙
                        </Button>
                      ) : null}
                    </div>
                    <Typography.Paragraph type="secondary" style={{ fontSize: 12, marginBottom: 8 }}>
                      身份钥匙（<Typography.Text code>ymk_</Typography.Text>
                      ）证明 Bot 是谁；与席页入场码（
                      <Typography.Text code>ymt_</Typography.Text>
                      ）分家。明文仅创建成功挡层展示一次，列表只留末四。
                    </Typography.Paragraph>
                    {keys.length ? (
                      <Table
                        className="acb-list-table"
                        size="small"
                        rowKey="id"
                        dataSource={keys}
                        pagination={false}
                        columns={[
                          { title: '名称', dataIndex: 'name' },
                          {
                            title: '身份钥匙',
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
                      empty('还没有身份钥匙')
                    )}
                  </div>

                  <div className="acb-settings-section">
                    <div className="acb-settings-section-head">
                      <span className="acb-settings-section-title">Webhook</span>
                      {isAdmin ? (
                        <Button
                          type="primary"
                          size="small"
                          icon={<PlusOutlined />}
                          onClick={() => {
                            hookForm.resetFields()
                            setHookOpen(true)
                          }}
                        >
                          新建
                        </Button>
                      ) : null}
                    </div>
                    {hooks.length ? (
                      <Table
                        className="acb-list-table"
                        size="small"
                        rowKey="id"
                        dataSource={hooks}
                        pagination={false}
                        columns={[
                          { title: '地址', dataIndex: 'url', ellipsis: true },
                          {
                            title: '状态',
                            dataIndex: 'enabled',
                            width: 88,
                            render: (v) => (
                              <Tag color={v ? 'success' : 'default'}>{v ? '启用' : '停用'}</Tag>
                            ),
                          },
                        ]}
                      />
                    ) : (
                      empty('还没有 Webhook')
                    )}
                  </div>
                </>
              ),
            },
          ]}
        />
      </div>

      <Modal
        title="新建标签"
        open={labelOpen}
        onCancel={() => setLabelOpen(false)}
        onOk={() => void onCreateLabel()}
        confirmLoading={labelBusy}
        okText="创建"
        cancelText="取消"
        destroyOnHidden
      >
        <Form form={labelForm} layout="vertical" style={{ marginTop: 8 }}>
          <Form.Item name="name" label="名称" rules={[{ required: true, message: '请填写名称' }]}>
            <Input maxLength={40} />
          </Form.Item>
          <Form.Item name="color" label="颜色" initialValue="#5E6AD2">
            <Input type="color" style={{ width: 56, padding: 2 }} />
          </Form.Item>
        </Form>
      </Modal>

      <Modal
        title="新建身份钥匙"
        open={keyOpen}
        onCancel={() => setKeyOpen(false)}
        onOk={() => void onCreateKey()}
        confirmLoading={keyBusy}
        okText="创建"
        cancelText="取消"
        destroyOnHidden
      >
        <Form form={keyForm} layout="vertical" style={{ marginTop: 8 }}>
          <Form.Item name="name" label="名称" rules={[{ required: true, message: '请填写名称' }]}>
            <Input placeholder="例如：本机 Agent / 自动化脚本" maxLength={80} />
          </Form.Item>
          <Typography.Paragraph type="secondary" style={{ fontSize: 12, marginBottom: 0 }}>
            创建成功后会弹出挡层展示完整 <Typography.Text code>ymk_</Typography.Text>
            身份钥匙（仅此一次），并可复制启动提示词。
          </Typography.Paragraph>
        </Form>
      </Modal>

      <IdentityKeyRevealModal
        open={Boolean(lastKeyToken)}
        token={lastKeyToken}
        onClose={() => setLastKeyToken(null)}
      />

      <Modal
        title="新建 Webhook"
        open={hookOpen}
        onCancel={() => setHookOpen(false)}
        onOk={() => void onCreateHook()}
        confirmLoading={hookBusy}
        okText="创建"
        cancelText="取消"
        destroyOnHidden
      >
        <Form form={hookForm} layout="vertical" style={{ marginTop: 8 }}>
          <Form.Item
            name="url"
            label="回调地址"
            rules={[
              { required: true, message: '请填写回调地址' },
              { type: 'url', message: '请输入有效的 http(s) 地址' },
            ]}
          >
            <Input placeholder="https://…" />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}
