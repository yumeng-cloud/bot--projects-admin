import { api } from './client'
import type { WikiSpace } from './boardTypes'

/** Wiki 入口占位：仅列 space */
export function listWikiSpaces(workspaceId: string) {
  return api<{ spaces: WikiSpace[] }>(`/workspaces/${workspaceId}/wiki/spaces`)
}
