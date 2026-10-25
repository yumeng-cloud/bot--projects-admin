import { useCallback, useEffect, useState } from 'react'
import {
  Button,
  Empty,
  Popconfirm,
  Space,
  Switch,
  Table,
  Tag,
  Typography,
  message,
} from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import { listProjects, patchProject, type Project } from '../../api/projects'
import { writeStoredProjectId } from '../../utils/currentProject'
import CreateProjectModal from '../../components/common/CreateProjectModal'
import EditProjectModal from '../../components/common/EditProjectModal'
import { formatShanghai } from '../../utils/time'

type Props = {
  workspaceId: string
  isAdmin: boolean
}

/** 设置 · 项目 Tab（PROJECT_ADMIN A1–A4） */
export default function SettingsProjectsPanel({ workspaceId }: Props) {
  const nav = useNavigate()
  const [loading, setLoading] = useState(true)
  const [projects, setProjects] = useState<Project[]>([])
  const [includeArchived, setIncludeArchived] = useState(false)
  const [createOpen, setCreateOpen] = useState(false)
  const [editTarget, setEditTarget] = useState<Project | null>(null)
  const [archivingId, setArchivingId] = useState<string | null>(null)

  const reload = useCallback(async () => {
    if (!workspaceId) return
    setLoading(true)
    try {
      const res = await listProjects(workspaceId, includeArchived)
      setProjects(res.projects || [])
    } catch (e: unknown) {
      message.error(e instanceof Error ? e.message : '加载项目失败')
      setProjects([])
    } finally {
      setLoading(false)
    }
  }, [workspaceId, includeArchived])

  useEffect(() => {
    void reload()
  }, [reload])

  const onArchive = async (row: Project, archive: boolean) => {
    setArchivingId(row.id)
    try {
      const { project } = await patchProject(workspaceId, row.id, { archive })
      message.success(archive ? `已归档「${project.name}」` : `已恢复「${project.name}」`)
      await reload()
    } catch (e: unknown) {
      message.error(e instanceof Error ? e.message : archive ? '归档失败' : '恢复失败')
    } finally {
      setArchivingId(null)
    }
  }

  const enterBoard = (row: Project) => {
    if (row.archived_at) {
      message.info('已归档项目请先恢复，再进入看板')
      return
    }
    writeStoredProjectId(workspaceId, row.id)
    nav(`/w/${workspaceId}/projects/${row.id}/board`)
  }

  return (
    <>
      <div className="acb-settings-section-head" style={{ marginBottom: 12 }}>
        <span className="acb-settings-section-title">项目</span>
        <Space size={12} wrap>
          <span style={{ fontSize: 12, color: 'var(--acb-text-secondary, #8b8b8b)' }}>
            含已归档
            <Switch
              size="small"
              checked={includeArchived}
              onChange={setIncludeArchived}
              style={{ marginLeft: 8 }}
            />
          </span>
          <Button
            type="primary"
            size="small"
            icon={<PlusOutlined />}
            onClick={() => setCreateOpen(true)}
          >
            新建项目
          </Button>
        </Space>
      </div>
      <p className="acb-settings-hint" style={{ marginBottom: 12 }}>
        点行进入该项目看板。新建与看板空态共用同一入口。
      </p>
      {projects.length ? (
        <Table
          className="acb-list-table"
          size="small"
          rowKey="id"
          loading={loading}
          dataSource={projects}
          pagination={false}
          onRow={(row) => ({
            onClick: () => enterBoard(row),
            style: { cursor: row.archived_at ? 'default' : 'pointer' },
          })}
          columns={[
            {
              title: '名称',
              dataIndex: 'name',
              render: (v: string, row: Project) => (
                <span>
                  <Typography.Text strong>{v || '未命名'}</Typography.Text>
                  {row.archived_at ? (
                    <Tag style={{ marginLeft: 8 }} color="default">
                      已归档
                    </Tag>
                  ) : null}
                </span>
              ),
            },
            {
              title: '前缀',
              dataIndex: 'key_prefix',
              width: 100,
              render: (v: string) => <Typography.Text code>{v}</Typography.Text>,
            },
            {
              title: '简介',
              dataIndex: 'description',
              ellipsis: true,
              render: (v?: string) => v || '—',
            },
            {
              title: '更新',
              dataIndex: 'updated_at',
              width: 140,
              render: (v?: string) => (v ? formatShanghai(v) : '—'),
            },
            {
              title: '操作',
              key: 'actions',
              width: 168,
              render: (_: unknown, row: Project) => (
                <Space size={0} onClick={(e) => e.stopPropagation()}>
                  <Button
                    type="link"
                    size="small"
                    disabled={Boolean(row.archived_at)}
                    onClick={() => setEditTarget(row)}
                  >
                    编辑
                  </Button>
                  {row.archived_at ? (
                    <Popconfirm
                      title={`确认恢复「${row.name}」？`}
                      description="恢复后会出现在顶栏项目列表。"
                      okText="恢复"
                      cancelText="取消"
                      onConfirm={() => void onArchive(row, false)}
                    >
                      <Button type="link" size="small" loading={archivingId === row.id}>
                        恢复
                      </Button>
                    </Popconfirm>
                  ) : (
                    <Popconfirm
                      title={`确认归档「${row.name}」？`}
                      description="归档后不再出现在顶栏；可随时恢复。"
                      okText="归档"
                      cancelText="取消"
                      onConfirm={() => void onArchive(row, true)}
                    >
                      <Button type="link" size="small" danger loading={archivingId === row.id}>
                        归档
                      </Button>
                    </Popconfirm>
                  )}
                </Space>
              ),
            },
          ]}
        />
      ) : (
        <div className="acb-empty-panel" style={{ minHeight: 120, padding: '20px 16px' }}>
          <Empty
            image={Empty.PRESENTED_IMAGE_SIMPLE}
            description={includeArchived ? '还没有项目' : '还没有进行中的项目'}
          >
            <Button type="primary" icon={<PlusOutlined />} onClick={() => setCreateOpen(true)}>
              新建项目
            </Button>
          </Empty>
        </div>
      )}

      <CreateProjectModal
        open={createOpen}
        workspaceId={workspaceId}
        onClose={() => setCreateOpen(false)}
        onCreated={(p) => {
          try {
            writeStoredProjectId(workspaceId, p.id)
          } catch {
            /* ignore */
          }
          void reload().then(() => {
            nav(`/w/${workspaceId}/projects/${p.id}/board`)
          })
        }}
      />
      <EditProjectModal
        open={Boolean(editTarget)}
        workspaceId={workspaceId}
        project={editTarget}
        onClose={() => setEditTarget(null)}
        onSaved={() => void reload()}
      />
    </>
  )
}
