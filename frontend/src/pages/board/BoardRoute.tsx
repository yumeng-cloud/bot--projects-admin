import { useParams } from 'react-router-dom'
import { Spin, Alert, Select, Space, Typography } from 'antd'
import { useEffect, useState } from 'react'
import { listProjects, type Project } from '../../api/projects'
import { preferMock } from '../../api/mockData'
import BoardPage from './BoardPage'
import { useWorkspaceProject } from '../../hooks/useWorkspaceProject'

export default function BoardRoute() {
  const { wsId } = useParams()
  const workspaceId = wsId || ''
  const { projectId, projectName, loading, setProjectId } = useWorkspaceProject(workspaceId)
  const [projects, setProjects] = useState<Project[]>([])

  useEffect(() => {
    if (!workspaceId || preferMock()) return
    listProjects(workspaceId)
      .then((r) => setProjects(r.projects || []))
      .catch(() => setProjects([]))
  }, [workspaceId])

  if (!workspaceId) return <Alert type="error" message="缺少工作区" />
  if (loading || !projectId) {
    return (
      <div style={{ textAlign: 'center', padding: 48 }}>
        <Spin />
      </div>
    )
  }

  return (
    <div>
      <Space style={{ marginBottom: 8 }}>
        <Typography.Text type="secondary">项目</Typography.Text>
        <Select
          size="small"
          style={{ minWidth: 180 }}
          value={projectId}
          onChange={(id) => {
            localStorage.setItem('ai-collab-board:projectId', id)
            setProjectId(id)
          }}
          options={
            projects.length
              ? projects.map((p) => ({ value: p.id, label: `${p.key_prefix} · ${p.name}` }))
              : [{ value: projectId, label: projectName || '演示项目' }]
          }
        />
      </Space>
      <BoardPage workspaceId={workspaceId} projectId={projectId} />
    </div>
  )
}
