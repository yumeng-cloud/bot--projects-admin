/** @acb/api-client — 鉴权 / 错误 / 类型 / session-context / 看板读改最小面 */

export {
  api,
  ApiError,
  getApiBase,
  getToken,
  setToken,
  resolveUrl,
  type ApiRequestInit,
} from './client'

export { login, register, logout, me } from './auth'

export {
  getSessionContext,
  putSessionContext,
  emptySessionContext,
  hasSessionScene,
  type SessionContext,
  type SessionContextPut,
} from './sessionContext'

export { listMyWorkspaces, getWorkspace } from './workspaces'
export { listTeams, getTeam, getRoster } from './teams'
export { listIssues, getIssue, patchIssue, type IssueListQuery } from './issues'
export { listComments, createComment } from './comments'
export { listWikiSpaces } from './wiki'

export type { User, AuthResp, MeResp, ApiErrorBody } from './types'

export type {
  Workspace,
  Team,
  TeamSeat,
  Issue,
  IssueStatus,
  IssuePriority,
  Comment,
  WikiSpace,
} from './boardTypes'

export { ISSUE_STATUS_LABEL, ISSUE_STATUSES } from './boardTypes'
