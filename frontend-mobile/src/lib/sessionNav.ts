import { hasSessionScene, type SessionContext } from '@acb/api-client'

/** 按 session-context 决定进壳落点（R2 主路径） */
export function pathFromSession(ctx: SessionContext): string {
  if (ctx.wsId && ctx.issueId) return `/w/${ctx.wsId}/issues/${ctx.issueId}`
  if (ctx.wsId && (ctx.view === 'board' || ctx.view === 'issues' || ctx.view === 'issue')) {
    return `/w/${ctx.wsId}/issues`
  }
  if (ctx.wsId && ctx.teamId) return `/crew/${ctx.wsId}/teams/${ctx.teamId}`
  if (ctx.wsId && ctx.view === 'wiki') return `/w/${ctx.wsId}/wiki`
  if (ctx.wsId) return `/crew/${ctx.wsId}`
  if (hasSessionScene(ctx)) return '/resume'
  return '/crew'
}

export function issuesPath(wsId: string | undefined | null): string {
  return wsId ? `/w/${wsId}/issues` : '/issues'
}

export function wikiPath(wsId: string | undefined | null): string {
  return wsId ? `/w/${wsId}/wiki` : '/wiki'
}
