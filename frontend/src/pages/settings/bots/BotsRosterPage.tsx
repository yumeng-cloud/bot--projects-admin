import { useEffect, useMemo, useState } from 'react'
import { Link, useParams } from 'react-router-dom'
import {
  Alert,
  Button,
  Card,
  Collapse,
  Empty,
  Select,
  Space,
  Spin,
  Table,
  Tag,
  Typography,
  message,
} from 'antd'
import {
  claimBotRole,
  listBotRoles,
  listBots,
  releaseBotRole,
  type BotRole,
  type BotRosterItem,
} from '../../../api/bots'
import { api } from '../../../api/client'
import { formatWikiTime } from '../../../utils/time'
import SeatHandoffDrawer from './SeatHandoffDrawer'
import { roleDisplay } from './seatHandoffText'

type ApiKeyRow = {
  id: string
  member_id?: string
  name: string
  key_prefix: string
  last_used_at?: string | null
  revoked_at?: string | null
}

/** 仅展示前缀尾四位，禁止完整 Key */
function keyTailHint(prefix?: string | null): string | undefined {
  if (!prefix) return undefined
  const clean = prefix.replace(/^ymk_/i, '')
  const tail = clean.slice(-4)
  if (!tail) return undefined
  return `Key 末四 ·${tail}`
}

export default function BotsRosterPage() {
  const { wsId = '' } = useParams()
  const [loading, setLoading] = useState(true)
  const [bots, setBots] = useState<BotRosterItem[]>([])
  const [roles, setRoles] = useState<BotRole[]>([])
  const [keys, setKeys] = useState<ApiKeyRow[]>([])
  const [drawerRole, setDrawerRole] = useState<BotRole | null>(null)
  const [claiming, setClaiming] = useState<string | null>(null)

  const load = async () => {
    setLoading(true)
    try {
      const [b, r, k] = await Promise.all([
        listBots(wsId),
        listBotRoles(wsId),
        api<{ api_keys: ApiKeyRow[] }>(`/workspaces/${wsId}/api-keys`).catch(() => ({ api_keys: [] as ApiKeyRow[] })),
      ])
      setBots(b.bots || [])
      setRoles(r.bot_roles || [])
      setKeys((k.api_keys || []).filter((x) => !x.revoked_at))
    } catch (e: unknown) {
      message.error(e instanceof Error ? e.message : '加载失败')
    } finally {
      setLoading(false)
    }
  }

  useEffect(() => {
    void load()
  }, [wsId])

  const botName = (memberId?: string | null) =>
    memberId ? bots.find((b) => b.member_id === memberId)?.display_name || '已绑定' : null

  const keyHintForMember = (memberId?: string | null) => {
    if (!memberId) return undefined
    const row = keys
      .filter((k) => k.member_id === memberId)
      .sort((a, b) => String(b.last_used_at || '').localeCompare(String(a.last_used_at || '')))[0]
    return keyTailHint(row?.key_prefix)
  }

  const freeBots = useMemo(() => bots.filter((b) => b.bot_enabled), [bots])

  const openSeat = (role: BotRole) => setDrawerRole(role)

  const onRelease = async (role: BotRole) => {
    try {
      await releaseBotRole(wsId, role.id, crypto.randomUUID())
      message.success(`已释放席位「${roleDisplay(role)}」`)
      void load()
    } catch {
      /* api client toasts */
    }
  }

  const onClaim = async (role: BotRole, memberId: string) => {
    setClaiming(role.id)
    try {
      await claimBotRole(wsId, role.id, memberId, crypto.randomUUID())
      message.success(`已将「${botName(memberId)}」绑到席位「${roleDisplay(role)}」`)
      void load()
    } catch {
      /* 409 等由 client 提示 */
    } finally {
      setClaiming(null)
    }
  }

  if (loading) {
    return (
      <div className="acb-page" style={{ textAlign: 'center', paddingTop: 48 }}>
        <Spin />
      </div>
    )
  }

  return (
    <div className="acb-page">
      <div className="acb-page-bar" style={{ marginBottom: 4 }}>
        <Typography.Title level={4} style={{ margin: 0 }}>
          Bot 席位
        </Typography.Title>
      </div>
      <Typography.Paragraph type="secondary" style={{ marginBottom: 12 }}>
        席位 = 角色坑（产品 / 前端 / QA…）。换机只换演员，不换席位与剧本；续命靠服务端{' '}
        <Typography.Text code>handoff get</Typography.Text>，不靠本机文件。
      </Typography.Paragraph>

      <Alert
        type="info"
        showIcon
        style={{ marginBottom: 16 }}
        message="换机怎么续"
        description={
          <ol style={{ margin: '4px 0 0', paddingLeft: 18, fontSize: 13 }}>
            <li>新机打开同一工作区</li>
            <li>
              绑定到同一席位（贴 Bot Key，或本机执行{' '}
              <Typography.Text code>ymbot seat enter --role &lt;slug&gt;</Typography.Text>）
            </li>
            <li>
              <Typography.Text code>ymbot handoff get</Typography.Text> 拉记忆 / Todo / 开放 Issue
            </li>
          </ol>
        }
      />

      {roles.length ? (
        <div
          style={{
            display: 'grid',
            gridTemplateColumns: 'repeat(auto-fill, minmax(280px, 1fr))',
            gap: 10,
            marginBottom: 20,
          }}
        >
          {roles.map((role) => {
            const mid = role.claim?.member_id
            const bound = botName(mid)
            const hint = keyHintForMember(mid)
            const busy = Boolean(mid)
            return (
              <Card
                key={role.id}
                size="small"
                hoverable
                styles={{ body: { padding: '12px 14px' } }}
                onClick={() => openSeat(role)}
              >
                <div style={{ display: 'flex', justifyContent: 'space-between', gap: 8, alignItems: 'flex-start' }}>
                  <div style={{ minWidth: 0 }}>
                    <Typography.Text strong style={{ fontSize: 15 }}>
                      {roleDisplay(role)}
                    </Typography.Text>
                    {role.slug ? (
                      <Typography.Text type="secondary" style={{ marginLeft: 8, fontSize: 12 }}>
                        {role.slug}
                      </Typography.Text>
                    ) : null}
                  </div>
                  <Tag color={busy ? 'blue' : 'default'} style={{ margin: 0, flexShrink: 0 }}>
                    {busy ? '已绑' : '空闲'}
                  </Tag>
                </div>

                <div style={{ marginTop: 8, fontSize: 13, lineHeight: 1.5 }}>
                  {busy ? (
                    <>
                      <span>演员 · {bound}</span>
                      {hint ? (
                        <Typography.Text type="secondary" style={{ marginLeft: 8 }}>
                          {hint}
                        </Typography.Text>
                      ) : null}
                      {role.claim?.claimed_at ? (
                        <div>
                          <Typography.Text type="secondary" style={{ fontSize: 12 }}>
                            认领 {formatWikiTime(role.claim.claimed_at)}
                          </Typography.Text>
                        </div>
                      ) : null}
                    </>
                  ) : (
                    <Typography.Text type="secondary">等待绑定本机 Agent / Key</Typography.Text>
                  )}
                </div>

                {role.responsibilities ? (
                  <Typography.Paragraph
                    type="secondary"
                    ellipsis={{ rows: 2 }}
                    style={{ margin: '8px 0 0', fontSize: 12 }}
                  >
                    {role.responsibilities}
                  </Typography.Paragraph>
                ) : null}

                <div
                  style={{ marginTop: 10, display: 'flex', flexWrap: 'wrap', gap: 6 }}
                  onClick={(e) => e.stopPropagation()}
                >
                  <Button size="small" type="link" style={{ padding: 0 }} onClick={() => openSeat(role)}>
                    查看交接
                  </Button>
                  {busy ? (
                    <Button size="small" type="link" danger style={{ padding: 0 }} onClick={() => void onRelease(role)}>
                      释放
                    </Button>
                  ) : (
                    <Select
                      size="small"
                      placeholder="指派 Bot 认领"
                      style={{ minWidth: 140 }}
                      loading={claiming === role.id}
                      options={freeBots.map((b) => ({ value: b.member_id, label: b.display_name }))}
                      onChange={(v) => void onClaim(role, v)}
                      popupMatchSelectWidth={false}
                    />
                  )}
                  {!busy ? (
                    <Typography.Text type="secondary" style={{ fontSize: 12, alignSelf: 'center' }}>
                      或本机 seat enter
                    </Typography.Text>
                  ) : null}
                </div>
              </Card>
            )
          })}
        </div>
      ) : (
        <div className="acb-empty-panel" style={{ marginBottom: 20 }}>
          <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无席位（请先拉班底 / seed）" />
        </div>
      )}

      <Collapse
        size="small"
        style={{ marginBottom: 8 }}
        items={[
          {
            key: 'bind-guide',
            label: '绑定此机怎么做（无设备列表）',
            children: (
              <Typography.Paragraph style={{ marginBottom: 0, fontSize: 13 }}>
                看板侧没有「本机设备」清单。请在新机 Cursor / Agent 环境：
                <br />
                1）打开本工作区；2）配置同一 Bot Key，或执行{' '}
                <Typography.Text code>ymbot seat enter --role &lt;slug&gt;</Typography.Text>；3）再{' '}
                <Typography.Text code>handoff get</Typography.Text>
                。上方「指派 Bot 认领」仅供管理员在看板内把已有 Bot 成员挂到空席。
              </Typography.Paragraph>
            ),
          },
          {
            key: 'advanced',
            label: '高级 · Bot 成员与 Key 管理',
            children: (
              <Space direction="vertical" style={{ width: '100%' }} size="middle">
                <Alert
                  type="warning"
                  showIcon
                  message="export / import 是可选备份与救场"
                  description="默认换机路径是：绑同一席位 → handoff get。不要把 JSON 导出/导入写成日常剧本。"
                />
                <div>
                  <Typography.Text strong>Bot 成员</Typography.Text>
                  {bots.length ? (
                    <Table
                      className="acb-list-table"
                      style={{ marginTop: 8 }}
                      rowKey="member_id"
                      size="small"
                      pagination={false}
                      dataSource={bots}
                      columns={[
                        {
                          title: '名称',
                          dataIndex: 'display_name',
                          render: (v: string, row) => (
                            <Link to={`/w/${wsId}/settings/bots/${row.member_id}`}>{v}</Link>
                          ),
                        },
                        {
                          title: '当前席位',
                          render: (_, row) => row.claimed_role?.name || '—',
                        },
                        {
                          title: 'Key',
                          width: 120,
                          render: (_, row) => keyHintForMember(row.member_id) || '—',
                        },
                        {
                          title: '',
                          width: 72,
                          render: (_, row) => (
                            <Link to={`/w/${wsId}/settings/bots/${row.member_id}`}>记忆</Link>
                          ),
                        },
                      ]}
                    />
                  ) : (
                    <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无 Bot 成员" />
                  )}
                </div>
                <div>
                  <Typography.Text strong>API Key（只读前缀）</Typography.Text>
                  <Typography.Paragraph type="secondary" style={{ margin: '4px 0 8px', fontSize: 12 }}>
                    完整密钥不会在此展示；轮换请走签发流程。设置页「API Key」Tab 亦可查看。
                  </Typography.Paragraph>
                  {keys.length ? (
                    <Table
                      className="acb-list-table"
                      rowKey="id"
                      size="small"
                      pagination={false}
                      dataSource={keys.slice(0, 20)}
                      columns={[
                        { title: '名称', dataIndex: 'name' },
                        {
                          title: '前缀',
                          dataIndex: 'key_prefix',
                          render: (v: string) => (
                            <Typography.Text code style={{ fontSize: 12 }}>
                              {v}
                            </Typography.Text>
                          ),
                        },
                        {
                          title: '演员',
                          render: (_, row) => botName(row.member_id) || '—',
                        },
                        {
                          title: '最近使用',
                          dataIndex: 'last_used_at',
                          render: (v?: string) => (v ? formatWikiTime(v) : '—'),
                        },
                      ]}
                    />
                  ) : (
                    <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description="暂无 Key 或接口未就绪" />
                  )}
                </div>
              </Space>
            ),
          },
        ]}
      />

      <SeatHandoffDrawer
        open={Boolean(drawerRole)}
        onClose={() => setDrawerRole(null)}
        wsId={wsId}
        role={drawerRole}
        bots={bots}
        keyHint={drawerRole ? keyHintForMember(drawerRole.claim?.member_id) : undefined}
      />
    </div>
  )
}
