import { useEffect, useState } from 'react'
import { listProjects } from '../api/projects'
import { MOCK_PROJECT, preferMock } from '../api/mockData'

const KEY = 'ai-collab-board:projectId'

export function useWorkspaceProject(workspaceId: string | undefined) {
  const [projectId, setProjectId] = useState<string | null>(() => localStorage.getItem(KEY))
  const [projectName, setProjectName] = useState<string>('')
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!workspaceId) return
    let cancelled = false
    ;(async () => {
      setLoading(true)
      setError(null)
      try {
        if (preferMock()) {
          if (!cancelled) {
            setProjectId(MOCK_PROJECT)
            setProjectName('演示项目')
            localStorage.setItem(KEY, MOCK_PROJECT)
          }
          return
        }
        const res = await listProjects(workspaceId)
        const list = res.projects || []
        if (!list.length) {
          if (!cancelled) {
            setProjectId(null)
            setError('还没有项目')
          }
          return
        }
        const saved = localStorage.getItem(KEY)
        const pick = list.find((p) => p.id === saved) || list[0]
        if (!cancelled) {
          setProjectId(pick.id)
          setProjectName(pick.name)
          localStorage.setItem(KEY, pick.id)
        }
      } catch (e: unknown) {
        if (!cancelled) {
          setProjectId(null)
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
