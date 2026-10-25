import { useEffect, useState } from 'react'
import { listProjects } from '../api/projects'
import { MOCK_PROJECT, preferMock } from '../api/mockData'
import {
  readStoredProjectId,
  writeStoredProjectId,
  resolveProjectId,
} from '../utils/currentProject'

export function useWorkspaceProject(workspaceId: string | undefined) {
  const [projectId, setProjectIdState] = useState<string | null>(() =>
    workspaceId ? readStoredProjectId(workspaceId) : null,
  )
  const [projectName, setProjectName] = useState<string>('')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const setProjectId = (id: string | null) => {
    setProjectIdState(id)
    if (workspaceId) writeStoredProjectId(workspaceId, id)
  }

  useEffect(() => {
    if (!workspaceId) return
    let cancelled = false
    // X-R1：切仓先清空再 fetch，零他仓项目残留
    setProjectIdState(null)
    setProjectName('')
    setLoading(true)
    setError(null)
    ;(async () => {
      try {
        if (preferMock()) {
          if (!cancelled) {
            setProjectIdState(MOCK_PROJECT)
            setProjectName('演示项目')
            writeStoredProjectId(workspaceId, MOCK_PROJECT)
          }
          return
        }
        const res = await listProjects(workspaceId)
        const list = res.projects || []
        if (!list.length) {
          if (!cancelled) {
            setProjectIdState(null)
            setError('还没有项目')
          }
          return
        }
        const saved = readStoredProjectId(workspaceId)
        const pickId = resolveProjectId(list, saved)
        const pick = list.find((p) => p.id === pickId) || list[0]
        if (!cancelled) {
          setProjectIdState(pick.id)
          setProjectName(pick.name)
          writeStoredProjectId(workspaceId, pick.id)
        }
      } catch (e: unknown) {
        if (!cancelled) {
          setProjectIdState(null)
          setError(e instanceof Error ? e.message : '加载项目失败')
        }
      } finally {
        if (!cancelled) setLoading(false)
      }
    })()
    return () => {
      cancelled = true
    }
  }, [workspaceId])

  return { projectId, projectName, loading, error, setProjectId }
}
