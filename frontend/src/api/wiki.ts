import { api } from './client'
import type { WikiPage, WikiSpace, WikiTreeNode } from '../types/wiki'

const ws = (id: string) => `/workspaces/${id}/wiki`

function idem() {
  return crypto.randomUUID()
}

export function listWikiSpaces(workspaceId: string) {
  return api<{ spaces: WikiSpace[] }>(`${ws(workspaceId)}/spaces`)
}

export function getWikiSpace(workspaceId: string, spaceId: string) {
  return api<{ space: WikiSpace }>(`${ws(workspaceId)}/spaces/${spaceId}`)
}

export function createWikiSpace(
  workspaceId: string,
  body: { name: string; slug?: string; description?: string },
) {
  return api<{ space: WikiSpace }>(`${ws(workspaceId)}/spaces`, {
    method: 'POST',
    body: JSON.stringify(body),
    idempotencyKey: idem(),
  })
}

export function updateWikiSpace(
  workspaceId: string,
  spaceId: string,
  body: { name?: string; slug?: string; description?: string },
) {
  return api<{ space: WikiSpace }>(`${ws(workspaceId)}/spaces/${spaceId}`, {
    method: 'PATCH',
    body: JSON.stringify(body),
    idempotencyKey: idem(),
  })
}

export function deleteWikiSpace(workspaceId: string, spaceId: string) {
  return api<{ message: string }>(`${ws(workspaceId)}/spaces/${spaceId}`, {
    method: 'DELETE',
    idempotencyKey: idem(),
  })
}

export function listWikiPages(workspaceId: string, spaceId: string) {
  return api<{ pages: WikiPage[] }>(`${ws(workspaceId)}/spaces/${spaceId}/pages`)
}

export function createWikiPage(
  workspaceId: string,
  spaceId: string,
  body: {
    title: string
    body_md: string
    slug?: string
    parent_id?: string | null
    project_id?: string | null
    issue_id?: string | null
  },
) {
  return api<{ page: WikiPage }>(`${ws(workspaceId)}/spaces/${spaceId}/pages`, {
    method: 'POST',
    body: JSON.stringify(body),
    idempotencyKey: idem(),
  })
}

export function getWikiPage(workspaceId: string, pageId: string) {
  return api<{ page: WikiPage }>(`${ws(workspaceId)}/pages/${pageId}`)
}

export function updateWikiPage(
  workspaceId: string,
  pageId: string,
  body: {
    title?: string
    body_md?: string
    slug?: string
    parent_id?: string | null
    project_id?: string | null
    issue_id?: string | null
  },
) {
  return api<{ page: WikiPage }>(`${ws(workspaceId)}/pages/${pageId}`, {
    method: 'PATCH',
    body: JSON.stringify(body),
    idempotencyKey: idem(),
  })
}

export function deleteWikiPage(workspaceId: string, pageId: string) {
  return api<{ message: string }>(`${ws(workspaceId)}/pages/${pageId}`, {
    method: 'DELETE',
    idempotencyKey: idem(),
  })
}

export function searchWiki(workspaceId: string, q: string) {
  const qs = new URLSearchParams({ q })
  return api<{ pages: WikiPage[]; q: string }>(`${ws(workspaceId)}/search?${qs}`)
}

export function wikiTree(workspaceId: string, spaceId: string) {
  const qs = new URLSearchParams({ space_id: spaceId })
  return api<{ tree: WikiTreeNode[]; space_id: string }>(`${ws(workspaceId)}/tree?${qs}`)
}
