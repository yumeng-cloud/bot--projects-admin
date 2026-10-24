import { useState } from 'react'
import { Button, Divider, Empty, Select, Typography } from 'antd'
import { PlusOutlined, SettingOutlined } from '@ant-design/icons'
import { useNavigate } from 'react-router-dom'
import type { Project } from '../../api/projects'
import CreateProjectModal from './CreateProjectModal'

type Props = {
  workspaceId: string
  projects: Project[]
  projectId: string | undefined
  onSelect: (projectId: string) => void
  /** 创建成功后：刷新列表并进入该项目 */
  onCreated: (project: Project) => void
  disabled?: boolean
  className?: string
}

/** 顶栏项目选择器：真实未归档列表 + 新建 / 管理入口（PROJECT_ADMIN B1–B2） */
export function ProjectPoolSelect({
  workspaceId,
  projects,
  projectId,
  onSelect,
  onCreated,
  disabled,
  className,
}: Props) {
  const nav = useNavigate()
  const [createOpen, setCreateOpen] = useState(false)
  const [open, setOpen] = useState(false)

  const valueInList = projectId && projects.some((p) => p.id === projectId) ? projectId : undefined

  return (
    <>
      <Select
        className={className}
        size="small"
        variant="borderless"
        style={{ minWidth: 160 }}
        placeholder="项目"
        value={valueInList}
        options={projects.map((p) => ({
          value: p.id,
          label: `${p.key_prefix} · ${p.name}`,
        }))}
        onChange={(id) => {
          onSelect(id)
          setOpen(false)
        }}
        disabled={disabled || !workspaceId}
        open={open}
        onOpenChange={setOpen}
        popupMatchSelectWidth={false}
        listHeight={240}
        aria-label="项目"
        notFoundContent={
          <div style={{ padding: '8px 4px', textAlign: 'center' }}>
            <Empty
              image={Empty.PRESENTED_IMAGE_SIMPLE}
              description={
                <Typography.Text type="secondary" style={{ fontSize: 12 }}>
                  还没有项目
                </Typography.Text>
              }
              styles={{ image: { height: 36 } }}
            />
            <Button
              type="link"
              size="small"
              icon={<PlusOutlined />}
              style={{ paddingInline: 0, fontSize: 12 }}
              onClick={() => {
                setOpen(false)
                setCreateOpen(true)
              }}
            >
              新建项目
            </Button>
          </div>
        }
        popupRender={(menu) => (
          <div>
            {projects.length > 0 ? menu : null}
            {projects.length === 0 ? (
              <div style={{ padding: '8px 4px', textAlign: 'center' }}>
                <Empty
                  image={Empty.PRESENTED_IMAGE_SIMPLE}
                  description={
                    <Typography.Text type="secondary" style={{ fontSize: 12 }}>
                      还没有项目
                    </Typography.Text>
                  }
                  styles={{ image: { height: 36 } }}
                />
              </div>
            ) : null}
            <Divider style={{ margin: '4px 0' }} />
            <div style={{ padding: '2px 4px 6px', display: 'flex', flexDirection: 'column', gap: 0 }}>
              <Button
                type="text"
                size="small"
                icon={<PlusOutlined />}
                block
                style={{ textAlign: 'left', fontSize: 12, height: 28 }}
                onClick={() => {
                  setOpen(false)
                  setCreateOpen(true)
                }}
              >
                新建项目
              </Button>
              <Button
                type="text"
                size="small"
                icon={<SettingOutlined />}
                block
                style={{ textAlign: 'left', fontSize: 12, height: 28 }}
                onClick={() => {
                  setOpen(false)
                  nav(`/w/${workspaceId}/settings?tab=projects`)
                }}
              >
                管理项目
              </Button>
            </div>
          </div>
        )}
      />
      <CreateProjectModal
        open={createOpen}
        workspaceId={workspaceId}
        onClose={() => setCreateOpen(false)}
        onCreated={(project) => {
          onCreated(project)
        }}
      />
    </>
  )
}
