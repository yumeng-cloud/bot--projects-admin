import { useEffect, useState } from 'react'
import { Tabs, Table, Typography, message, Spin, Empty, Tag } from 'antd'
import { useParams } from 'react-router-dom'
import { api } from '../../api/client'
import { roleLabel } from '../../utils/labels'
import { formatShanghai } from '../../utils/time'

type Member = {
  id: string
  display_name: string
  role: string
  is_bot?: boolean | number
  email?: string | null
}
type Label = { id: string; name: string; color: string }
type ApiKey = { id: string; name: string; key_prefix: string; created_at?: string }
type Webhook = { id: string; url: string; enabled: number; events?: string }

function keyTail(prefix?: string | null): string {
  if (!prefix) return '—'
  const clean = prefix.replace(/^ymk_/i, '')
  const tail = clean.slice(-4)
  return tail ? `末四 ·${tail}` : prefix
}

export default function SettingsPage() {
  const { wsId = '' } = useParams()
  const [loading, setLoading] = useState(true)
  const [members, setMembers] = useState<Member[]>([])
  const [labels, setLabels] = useState<Label[]>([])
  const [keys, setKeys] = useState<ApiKey[]>([])
  const [hooks, setHooks] = useState<Webhook[]>([])
  const [err, setErr] = useState<string | null>(null)

  useEffect(() => {
    ;(async () => {
      setLoading(true)
      setErr(null)
      try {
        const [m, l, k, h] = await Promise.allSettled([
          api<{ members: Member[] }>(`/workspaces/${wsId}/members`),
          api<{ labels: Label[] }>(`/workspaces/${wsId}/labels`),
          api<{ api_keys: ApiKey[] }>(`/workspaces/${wsId}/api-keys`),
          api<{ webhooks: Webhook[] }>(`/workspaces/${wsId}/webhooks`),
        ])
        if (m.status === 'fulfilled') setMembers(m.value.members || [])
        if (l.status === 'fulfilled') setLabels(l.value.labels || [])
        if (k.status === 'fulfilled') setKeys(k.value.api_keys || [])
        if (h.status === 'fulfilled') setHooks(h.value.webhooks || [])
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
  }, [wsId])

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

  return (
    <div className="acb-page">
      <div className="acb-page-bar">
        <div>
          <Typography.Title level={4} style={{ margin: 0 }}>
            设置
          </Typography.Title>
          <Typography.Text type="secondary" style={{ fontSize: 12 }}>
            成员、标签与接入凭据
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
              key: 'members',
              label: `成员 (${members.length})`,
              children: (
                <>
                  {members.length ? (
                    <Table
                      className="acb-list-table"
                      size="small"
                      rowKey="id"
                      dataSource={members}
                      pagination={false}
                      columns={[
                        {
                          title: '名称',
                          dataIndex: 'display_name',
                          render: (v: string, row) => (
                            <span>
                              {v || '未命名'}
                              {row.is_bot ? (
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
                          width: 110,
                          render: (v: string) => roleLabel(v),
                        },
                      ]}
                    />
                  ) : (
                    empty('暂无成员或接口未就绪')
                  )}
                  <div className="acb-settings-tab-foot">
                    成员含人类与 Bot。Bot 席位与入场码在「Bot」设置；API Key / Webhook 见其它标签。
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
