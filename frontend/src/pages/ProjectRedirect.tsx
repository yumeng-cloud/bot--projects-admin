import { useEffect, useState } from 'react'
import { Navigate, useParams } from 'react-router-dom'
import { Spin, Result, Button } from 'antd'
import { listProjects } from '../api/projects'
import { MOCK_PROJECT, preferMock } from '../api/mockData'

/** /w/:wsId/board → 第一个项目的 board */
export default function ProjectRedirect({ tab }: { tab: 'board' | 'issues' | 'cycles' | 'timeline' }) {
  const { wsId = '' } = useParams()
  const [projectId, setProjectId] = useState<string | null>(null)
  const [err, setErr] = useState<string | null>(null)

  useEffect(() => {
    let cancelled = false
    ;(async () => {
      try {
        if (preferMock()) {
          if (!cancelled) setProjectId(MOCK_PROJECT)
          return
        }
        const res = await listProjects(wsId)
        const first = res.projects?.[0]
        if (!cancelled) {
          if (first) setProjectId(first.id)
          else setErr('暂无项目，请先创建项目')
        }
      } catch (e: unknown) {
        if (!cancelled) {
          setErr(e instanceof Error ? e.message : '加载项目失败')
        }
      }
    })()
    return () => {
      cancelled = true
    }
  }, [wsId])

  if (err) {
    return (
      <Result
        status="warning"
        title="无法进入项目视图"
        subTitle={err}
        extra={
          <Button type="primary" href={`/w/${wsId}/settings`}>
            去设置
          </Button>
        }
      />
    )
  }
  if (!projectId) {
    return (
      <div style={{ padding: 48, textAlign: 'center' }}>
        <Spin />
      </div>
    )
  }
  return <Navigate to={`/w/${wsId}/projects/${projectId}/${tab}`} replace />
}
