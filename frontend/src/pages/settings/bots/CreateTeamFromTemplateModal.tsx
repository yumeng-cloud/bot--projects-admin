import { useEffect, useMemo, useState } from 'react'
import { Alert, Input, Modal, Select, Space, Table, Typography, message } from 'antd'
import { ApiError } from '../../../api/client'
import {
  createTeam,
  ensureRoster,
  listTeamTemplates,
  playbookRefForEnsure,
  type Team,
  type TeamSeat,
  type TeamTemplateCatalogItem,
} from '../../../api/teams'

type Props = {
  open: boolean
  wsId: string
  onClose: () => void
  /** 创建成功：新队已 ensure；调用方写 acb.team 并刷新 */
  onCreated: (team: Team, roster: TeamSeat[]) => void
}

export default function CreateTeamFromTemplateModal({ open, wsId, onClose, onCreated }: Props) {
  const [loadingCatalog, setLoadingCatalog] = useState(false)
  const [templates, setTemplates] = useState<TeamTemplateCatalogItem[]>([])
  const [fromApi, setFromApi] = useState(false)
  const [templateKey, setTemplateKey] = useState<string>('default')
  const [teamName, setTeamName] = useState('')
  const [busy, setBusy] = useState(false)

  useEffect(() => {
    if (!open || !wsId) return
    let cancelled = false
    setLoadingCatalog(true)
    setTeamName('')
    void (async () => {
      try {
        const { templates: list, fromApi: api } = await listTeamTemplates(wsId)
        if (cancelled) return
        setTemplates(list)
        setFromApi(api)
        const first = list[0]?.key || 'default'
        setTemplateKey(first)
      } catch {
        if (cancelled) return
        setTemplates([])
        setFromApi(false)
        setTemplateKey('default')
      } finally {
        if (!cancelled) setLoadingCatalog(false)
      }
    })()
    return () => {
      cancelled = true
    }
  }, [open, wsId])

  const selected = useMemo(
    () => templates.find((t) => t.key === templateKey) || templates[0] || null,
    [templates, templateKey],
  )

  const onOk = async () => {
    const name = teamName.trim()
    if (!name) {
      message.warning('请填写队名')
      return
    }
    const tmpl = selected
    if (!tmpl) {
      message.warning('请选择模板')
      return
    }
    setBusy(true)
    try {
      let team: Team
      let roster: TeamSeat[] = []
      try {
        const created = await createTeam(
          wsId,
          { name, template_key: tmpl.key },
          crypto.randomUUID(),
        )
        team = created.team
        roster = created.roster || []
      } catch (e: unknown) {
        if (e instanceof ApiError && (e.status === 404 || e.status === 501)) {
          message.warning('创建团队接口尚未开通，请稍后再试')
          return
        }
        message.error(e instanceof ApiError ? e.message : '创建团队失败')
        return
      }

      // 若 Go 创建时已挂席则跳过；否则 ensure（带建议 playbook_ref）
      if (!roster.length) {
        const seatsBody = tmpl.seats.map((s) => ({
          slug: s.slug,
          name: s.name,
          playbook_ref: playbookRefForEnsure(s.playbook_ref),
        }))
        try {
          const ens = await ensureRoster(
            wsId,
            team.id,
            { template_key: tmpl.key, seats: seatsBody },
            crypto.randomUUID(),
            { skipToast: true },
          )
          roster = ens.roster || []
        } catch (e: unknown) {
          // 未知模板 / 拒 seats → 只传 template_key: default
          if (
            e instanceof ApiError &&
            (e.status === 400 || e.status === 404 || e.status === 422)
          ) {
            try {
              const ens = await ensureRoster(
                wsId,
                team.id,
                { template_key: 'default' },
                crypto.randomUUID(),
                { skipToast: true },
              )
              roster = ens.roster || []
              message.info('模板席位参数未完全接受，已按默认班底补空席')
            } catch (e2: unknown) {
              if (e2 instanceof ApiError && (e2.status === 404 || e2.status === 501)) {
                message.warning('班底 ensure 尚未开通；团队已创建，可稍后拉席')
              } else {
                message.warning(
                  e2 instanceof ApiError
                    ? `团队已创建，补席失败：${e2.message}`
                    : '团队已创建，补席失败',
                )
              }
            }
          } else if (e instanceof ApiError && (e.status === 404 || e.status === 501)) {
            message.warning('班底 ensure 尚未开通；团队已创建，可稍后拉席')
          } else {
            message.warning(
              e instanceof ApiError ? `团队已创建，补席失败：${e.message}` : '团队已创建，补席失败',
            )
          }
        }
      }

      onCreated(team, roster)
      onClose()
      message.success(`已创建团队「${team.name}」· ${roster.length || tmpl.seats.length} 空席`)
      message.info('复制入场包，对方入场时自动获得身份', 4)
    } finally {
      setBusy(false)
    }
  }

  return (
    <Modal
      title="从模板创建"
      open={open}
      onCancel={onClose}
      onOk={() => void onOk()}
      okText="创建"
      cancelText="取消"
      confirmLoading={busy}
      destroyOnClose
      width={640}
    >
      <Space direction="vertical" size={12} style={{ width: '100%' }}>
        {!fromApi && !loadingCatalog ? (
          <Alert
            type="info"
            showIcon
            message="模板目录接口未就绪，使用内置「默认班底」预览"
          />
        ) : null}
        <div>
          <Typography.Text type="secondary" style={{ fontSize: 12 }}>
            模板
          </Typography.Text>
          <Select
            style={{ width: '100%', marginTop: 4 }}
            loading={loadingCatalog}
            value={selected?.key}
            options={templates.map((t) => ({
              value: t.key,
              label: t.name,
            }))}
            onChange={(v) => setTemplateKey(v)}
            placeholder="选择模板"
          />
          {selected?.description ? (
            <Typography.Paragraph type="secondary" style={{ margin: '6px 0 0', fontSize: 12 }}>
              {selected.description}
            </Typography.Paragraph>
          ) : null}
        </div>
        <div>
          <Typography.Text type="secondary" style={{ fontSize: 12 }}>
            空席预览（创建后复制入场包发给对方）
          </Typography.Text>
          <Table
            size="small"
            style={{ marginTop: 4 }}
            pagination={false}
            rowKey="slug"
            dataSource={selected?.seats || []}
            columns={[
              { title: 'slug', dataIndex: 'slug', width: 120 },
              { title: '显示名', dataIndex: 'name', width: 120 },
              {
                title: 'playbook_ref',
                dataIndex: 'playbook_ref',
                ellipsis: true,
              },
            ]}
            locale={{ emptyText: loadingCatalog ? '加载中…' : '无席位' }}
          />
        </div>
        <div>
          <Typography.Text type="secondary" style={{ fontSize: 12 }}>
            队名
          </Typography.Text>
          <Input
            style={{ marginTop: 4 }}
            value={teamName}
            placeholder="例如：产品一组"
            maxLength={64}
            onChange={(e) => setTeamName(e.target.value)}
            onPressEnter={() => void onOk()}
          />
        </div>
      </Space>
    </Modal>
  )
}
