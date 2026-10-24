import { useEffect, useState } from 'react'
import { Tabs, Table, Typography, message, Spin, Empty } from 'antd'
import { useParams } from 'react-router-dom'
import { api } from '../../api/client'

type Member = { id: string; display_name: string; role: string; is_bot?: number }
type Label = { id: string; name: string; color: string }
type ApiKey = { id: string; name: string; key_prefix: string; created_at?: string }
type Webhook = { id: string; url: string; enabled: number; events?: string }

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
        const firstErr = [m, l, k, h].find((x) => x.status === 'rejected') as PromiseRejectedResult | undefined
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
        <Typography.Title level={4} style={{ margin: 0 }}>
          设置
        </Typography.Title>
      </div>
      {err && <Typography.Paragraph type="warning">{err}</Typography.Paragraph>}
      <Tabs
        items={[
          {
            key: 'members',
            label: '成员',
            children: members.length ? (
              <Table
                className="acb-list-table"
                size="small"
                rowKey="id"
                dataSource={members}
                pagination={false}
                columns={[
                  { title: '名称', dataIndex: 'display_name' },
                  { title: '角色', dataIndex: 'role' },
                ]}
              />
            ) : (
              empty('暂无成员或接口未就绪')
            ),
          },
          {
            key: 'labels',
            label: '标签',
            children: labels.length ? (
              <Table
                className="acb-list-table"
                size="small"
                rowKey="id"
                dataSource={labels}
                pagination={false}
                columns={[
                  { title: '名称', dataIndex: 'name' },
                  { title: '颜色', dataIndex: 'color' },
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
                  { title: '前缀', dataIndex: 'key_prefix' },
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
                  { title: 'URL', dataIndex: 'url' },
                  { title: '启用', dataIndex: 'enabled', render: (v) => (v ? '是' : '否') },
                ]}
              />
            ) : (
              empty('暂无 Webhook 或接口未就绪')
            ),
          },
        ]}
      />
    </div>
  )
}
