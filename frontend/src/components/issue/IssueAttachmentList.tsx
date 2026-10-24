import { useEffect, useState } from 'react'
import { Button, Image, Typography, message } from 'antd'
import {
  DeleteOutlined,
  FileOutlined,
  FilePdfOutlined,
  FileWordOutlined,
  FileExcelOutlined,
  FileZipOutlined,
  FileTextOutlined,
} from '@ant-design/icons'
import type { ReactNode } from 'react'
import type { IssueAttachment } from '../../types/issue'
import { attachmentSizeBytes } from '../../types/issue'
import { downloadAttachment, fetchAttachmentBlob } from '../../api/attachments'

type Props = {
  attachments: IssueAttachment[]
  /** 空态文案；默认「暂无附件」 */
  emptyText?: string
  /** 提供时显示删除按钮 */
  onDelete?: (id: string) => void
  /** 正在删除的附件 id（该行 delete 进 loading） */
  deletingId?: string
}

function formatSize(bytes: number): string {
  if (!Number.isFinite(bytes) || bytes < 0) return '—'
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

function isImage(a: IssueAttachment): boolean {
  return (
    (a.content_type || '').startsWith('image/') ||
    /\.(png|jpe?g|gif|webp|bmp|svg)$/i.test(a.filename)
  )
}

function docIcon(a: IssueAttachment): ReactNode {
  const name = a.filename.toLowerCase()
  const ct = (a.content_type || '').toLowerCase()
  if (ct.includes('pdf') || name.endsWith('.pdf')) return <FilePdfOutlined />
  if (ct.includes('word') || /\.(docx?|rtf)$/i.test(name)) return <FileWordOutlined />
  if (ct.includes('sheet') || ct.includes('excel') || /\.(xlsx?|csv)$/i.test(name))
    return <FileExcelOutlined />
  if (ct.includes('zip') || /\.(zip|rar|7z|tar|gz)$/i.test(name)) return <FileZipOutlined />
  if (ct.startsWith('text/') || /\.(txt|md)$/i.test(name)) return <FileTextOutlined />
  return <FileOutlined />
}

/** 带 Bearer 拉 blob 的缩略图，可点 antd Image 预览。 */
function AuthThumb({ url, alt }: { url: string; alt: string }) {
  const [src, setSrc] = useState<string | null>(null)

  useEffect(() => {
    let revoked: string | null = null
    let cancelled = false
    void fetchAttachmentBlob(url)
      .then((blob) => {
        if (cancelled) return
        const obj = URL.createObjectURL(blob)
        revoked = obj
        setSrc(obj)
      })
      .catch(() => {
        if (!cancelled) setSrc(null)
      })
    return () => {
      cancelled = true
      if (revoked) URL.revokeObjectURL(revoked)
    }
  }, [url])

  if (!src) {
    return <div className="acb-issue-attachment-thumb-ph" title={alt} />
  }
  return (
    <Image
      className="acb-issue-attachment-thumb-img"
      src={src}
      alt={alt}
      width={36}
      height={36}
      style={{ objectFit: 'cover', borderRadius: 4 }}
      preview={{ src }}
    />
  )
}

/**
 * Issue 附件列表（Linear 密度）：统一一行一条。
 * 图片：小缩略图 + 文件名 + 大小；可点 antd Image 预览。
 * 非图：文件图标 + 文件名 + 大小。无上传者/时间噪音。
 */
export function IssueAttachmentList({
  attachments,
  emptyText = '暂无附件',
  onDelete,
  deletingId,
}: Props) {
  if (attachments.length === 0) {
    if (!emptyText) return null
    return <div className="acb-issue-muted-empty">{emptyText}</div>
  }

  const onDownload = async (a: IssueAttachment) => {
    if (!a.url) {
      message.warning('附件地址无效')
      return
    }
    try {
      await downloadAttachment(a.url, a.filename)
    } catch (err) {
      const msg = err instanceof Error ? err.message : '下载失败'
      message.error(msg)
    }
  }

  return (
    <Image.PreviewGroup>
      <ul className="acb-issue-attachment-list">
        {attachments.map((a) => {
          const image = isImage(a)
          const busy = deletingId === a.id
          return (
            <li key={a.id} className="acb-issue-attachment-row">
              <span className="acb-issue-attachment-leading">
                {image && a.url ? (
                  <AuthThumb url={a.url} alt={a.filename} />
                ) : (
                  <span className="acb-issue-attachment-icon" aria-hidden>
                    {docIcon(a)}
                  </span>
                )}
              </span>

              <div className="acb-issue-attachment-meta">
                {image ? (
                  <Typography.Text
                    ellipsis
                    className="acb-issue-attachment-name"
                    title={a.filename}
                  >
                    {a.filename}
                  </Typography.Text>
                ) : (
                  <button
                    type="button"
                    className="acb-issue-attachment-name acb-issue-attachment-name-btn"
                    title={`下载 ${a.filename}`}
                    onClick={() => void onDownload(a)}
                  >
                    {a.filename}
                  </button>
                )}
                <Typography.Text type="secondary" className="acb-issue-attachment-size">
                  {formatSize(attachmentSizeBytes(a))}
                </Typography.Text>
              </div>

              {onDelete ? (
                <Button
                  type="text"
                  size="small"
                  danger
                  className="acb-issue-attachment-delete"
                  icon={<DeleteOutlined />}
                  loading={busy}
                  disabled={busy}
                  title="删除附件"
                  aria-label={`删除 ${a.filename}`}
                  onClick={(e) => {
                    e.stopPropagation()
                    onDelete(a.id)
                  }}
                />
              ) : null}
            </li>
          )
        })}
      </ul>
    </Image.PreviewGroup>
  )
}
