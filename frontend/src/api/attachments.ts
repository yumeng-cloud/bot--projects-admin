/**
 * Issue / Comment 附件 API — 对齐 docs/API.md + router.go
 *
 * POST/GET  …/workspaces/{ws}/issues/{issueID}/attachments
 * POST/GET  …/workspaces/{ws}/issues/{issueID}/comments/{commentID}/attachments  （评论必须带 issueID）
 * GET/DELETE …/workspaces/{ws}/attachments/{id}
 * GET        …/workspaces/{ws}/attachments/{id}/content  （相对 url + Bearer，禁止无头 window.open）
 */
import { api, ApiError, getToken, resolveUrl } from './client'
import type { IssueAttachment } from '../types/issue'

function idem() {
  return crypto.randomUUID()
}

export const ATTACHMENT_MAX_BYTES = 10 * 1024 * 1024

/** 与后端允许类型对齐：png/jpg/jpeg/gif/webp/pdf/docx + txt/md/csv */
export const ATTACHMENT_ALLOWED_EXTS = [
  '.png',
  '.jpg',
  '.jpeg',
  '.gif',
  '.webp',
  '.pdf',
  '.docx',
  '.txt',
  '.md',
  '.csv',
] as const

export const ATTACHMENT_ACCEPT =
  'image/png,image/jpeg,image/gif,image/webp,.pdf,.docx,.txt,.md,.csv'

export type ListAttachmentsResult = { attachments: IssueAttachment[] }

/** Map legacy `size` → `size_bytes` when backend/proxy omits the canonical field. */
function normalizeAttachment(raw: IssueAttachment & { size?: number }): IssueAttachment {
  const size_bytes =
    typeof raw.size_bytes === 'number' && Number.isFinite(raw.size_bytes)
      ? raw.size_bytes
      : typeof raw.size === 'number' && Number.isFinite(raw.size)
        ? raw.size
        : 0
  const { size: _legacy, ...rest } = raw as IssueAttachment & { size?: number }
  return { ...rest, size_bytes }
}

function normalizeList(res: ListAttachmentsResult): ListAttachmentsResult {
  return { attachments: (res.attachments || []).map(normalizeAttachment) }
}


/** 客户端预检：类型 + ≤10MB。通过返回 null，否则返回中文错误文案。 */
export function validateAttachmentFile(file: File): string | null {
  if (file.size > ATTACHMENT_MAX_BYTES) {
    return '文件过大，上限 10MB'
  }
  const name = file.name.toLowerCase()
  const okExt = ATTACHMENT_ALLOWED_EXTS.some((ext) => name.endsWith(ext))
  const okMime =
    file.type === 'image/png' ||
    file.type === 'image/jpeg' ||
    file.type === 'image/gif' ||
    file.type === 'image/webp'
  if (!okExt && !okMime) {
    return '不支持的文件类型'
  }
  return null
}

async function postMultipart<T>(path: string, file: File): Promise<T> {
  const fd = new FormData()
  fd.append('file', file)
  const headers = new Headers()
  const token = getToken()
  if (token) headers.set('Authorization', `Bearer ${token}`)
  headers.set('Idempotency-Key', idem())
  // 勿设 Content-Type，浏览器自动带 multipart boundary
  const res = await fetch(resolveUrl(path), {
    method: 'POST',
    headers,
    body: fd,
    credentials: 'include',
  })
  const text = await res.text()
  let data: unknown = {}
  if (text) {
    try {
      data = JSON.parse(text)
    } catch {
      data = { message: text }
    }
  }
  if (!res.ok) {
    const o = (data && typeof data === 'object' ? data : {}) as Record<string, unknown>
    const nested = o.error
    const nestedMsg =
      nested && typeof nested === 'object' && typeof (nested as { message?: string }).message === 'string'
        ? (nested as { message: string }).message
        : null
    const msg =
      (typeof o.message === 'string' && o.message) ||
      nestedMsg ||
      res.statusText ||
      `HTTP ${res.status}`
    throw new ApiError(msg, typeof o.code === 'string' ? o.code : 'http', res.status)
  }
  return data as T
}

/** POST /workspaces/{ws}/issues/{issueID}/attachments */
export async function uploadIssueAttachment(
  workspaceId: string,
  issueId: string,
  file: File,
): Promise<{ attachment: IssueAttachment }> {
  const res = await postMultipart<{ attachment: IssueAttachment }>(
    `/workspaces/${workspaceId}/issues/${issueId}/attachments`,
    file,
  )
  return { attachment: normalizeAttachment(res.attachment) }
}

/** POST /workspaces/{ws}/issues/{issueID}/comments/{commentID}/attachments */
export async function uploadCommentAttachment(
  workspaceId: string,
  issueId: string,
  commentId: string,
  file: File,
): Promise<{ attachment: IssueAttachment }> {
  const res = await postMultipart<{ attachment: IssueAttachment }>(
    `/workspaces/${workspaceId}/issues/${issueId}/comments/${commentId}/attachments`,
    file,
  )
  return { attachment: normalizeAttachment(res.attachment) }
}

/** GET /workspaces/{ws}/issues/{issueID}/attachments —— Issue 域已通，硬接 */
export async function listIssueAttachments(workspaceId: string, issueId: string) {
  const res = await api<ListAttachmentsResult>(
    `/workspaces/${workspaceId}/issues/${issueId}/attachments`,
    { skipToast: true },
  )
  return normalizeList(res)
}

/** GET /workspaces/{ws}/issues/{issueID}/comments/{commentID}/attachments */
export async function listCommentAttachments(
  workspaceId: string,
  issueId: string,
  commentId: string,
) {
  const res = await api<ListAttachmentsResult>(
    `/workspaces/${workspaceId}/issues/${issueId}/comments/${commentId}/attachments`,
    { skipToast: true },
  )
  return normalizeList(res)
}

/** DELETE /workspaces/{ws}/attachments/{attachmentID} */
export function deleteAttachment(workspaceId: string, attachmentId: string) {
  return api<{ ok: boolean }>(`/workspaces/${workspaceId}/attachments/${attachmentId}`, {
    method: 'DELETE',
    idempotencyKey: idem(),
  })
}

/** 别名：历史调用名 */
export const deleteIssueAttachment = deleteAttachment

/**
 * 带 Bearer 拉取附件 content → Blob。
 * url 为契约相对路径，如 /api/v1/workspaces/{ws}/attachments/{id}/content
 */
export async function fetchAttachmentBlob(contentUrl: string): Promise<Blob> {
  const headers = new Headers()
  const token = getToken()
  if (token) headers.set('Authorization', `Bearer ${token}`)
  const res = await fetch(resolveUrl(contentUrl), {
    headers,
    credentials: 'include',
  })
  if (!res.ok) {
    const text = await res.text().catch(() => '')
    throw new ApiError(text || `下载失败 HTTP ${res.status}`, 'http', res.status)
  }
  return res.blob()
}

/** 带 Token 下载并触发保存（禁止无头 window.open）。 */
export async function downloadAttachment(contentUrl: string, filename: string): Promise<void> {
  const blob = await fetchAttachmentBlob(contentUrl)
  const obj = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = obj
  a.download = filename || 'attachment'
  a.rel = 'noopener'
  document.body.appendChild(a)
  a.click()
  a.remove()
  URL.revokeObjectURL(obj)
}
