import { useEffect, useState } from 'react'
import { Link, Navigate, useParams, useSearchParams } from 'react-router-dom'
import { Button, Empty, Spin, Typography } from 'antd'
import { PlusOutlined } from '@ant-design/icons'
import { listProjects } from '../api/projects'
import { MOCK_PROJECT, preferMock } from '../api/mockData'
import CreateProjectModal from '../components/common/CreateProjectModal'

/** /w/:wsId/board → 第一个项目的 board；无项目时平静空态 + 本页创建 */
export default function ProjectRedirect({ tab }: { tab: 'board' | 'issues' | 'cycles' | 'timeline' }) {
  const { wsId = '' } = useParams()
  const [searchParams] = useSearchParams()
  const forceEmpty = searchParams.get('qaEmpty') === '1'
  const [projectId, setProjectId] = useState<string | null>(null)
  const [empty, setEmpty] = useState(false)
  const [loadErr, setLoadErr] = useState<string | null>(null)
  const [modalOpen, setModalOpen] = useState(false)

  useEffect(() => {
    let cancelled = false
    setEmpty(false)
    setLoadErr(null)
    setProjectId(null)
    ;(async () => {
      try {
        if (forceEmpty) {
          if (!cancelled) setEmpty(true)
          return
        }
        if (preferMock()) {
          if (!cancelled) setProjectId(MOCK_PROJECT)
          return
        }
        const res = await listProjects(wsId)
        const first = res.projects?.[0]
        if (!cancelled) {
          if (first) setProjectId(first.id)
          else setEmpty(true)
        }
      } catch (e: unknown) {
        if (!cancelled) {
          setLoadErr(e instanceof Error ? e.message : '加载项目失败')
        }
      }
    })()
    return () => {
      cancelled = true
    }
  }, [wsId, forceEmpty])

  if (projectId) {
    return <Navigate to={`/w/${wsId}/projects/${projectId}/${tab}`} replace />
  }

  if (loadErr) {
    return (
      <div className="acb-empty-panel" style={{ padding: '48px 16px', textAlign: 'center' }}>
        <Empty image={Empty.PRESENTED_IMAGE_SIMPLE} description={loadErr}>
          <Button onClick={() => window.location.reload()}>重试</Button>
        </Empty>
      </div>
    )
  }

  if (empty) {
    return (
      <div className="acb-empty-panel" style={{ padding: '48px 16px', textAlign: 'center' }}>
        <Empty
          image={Empty.PRESENTED_IMAGE_SIMPLE}
          description={
            <span>
              <Typography.Text strong style={{ display: 'block', fontSize: 15, marginBottom: 4 }}>
                还没有项目
              </Typography.Text>
              <Typography.Text type="secondary" style={{ fontSize: 13 }}>
                创建一个项目后即可打开看板、列表与周期
              </Typography.Text>
            </span>
          }
        >
          <Button type="primary" icon={<PlusOutlined />} onClick={() => setModalOpen(true)}>
            创建项目
          </Button>
          <div style={{ marginTop: 12 }}>
            <Link to={`/w/${wsId}/settings`} style={{ fontSize: 12 }}>
              去设置
            </Link>
          </div>
        </Empty>
        <CreateProjectModal
          open={modalOpen}
          workspaceId={wsId}
          onClose={() => setModalOpen(false)}
          onCreated={(p) => {
            try {
              localStorage.setItem('ai-collab-board:projectId', p.id)
            } catch {
              /* ignore */
            }
            setProjectId(p.id)
          }}
        />
      </div>
    )
  }

  return (
    <div style={{ padding: 48, textAlign: 'center' }}>
      <Spin />
    </div>
  )
}
