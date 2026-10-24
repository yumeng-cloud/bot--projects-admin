import { useCallback, useRef, useState } from 'react'
import { Alert, Button, Modal, Space, Typography, message } from 'antd'
import { CopyOutlined, FileTextOutlined } from '@ant-design/icons'
import { buildFirstRunPrompt, getPublicApiBase } from '../../utils/oneTapJoin'

export type IdentityKeyRevealModalProps = {
  open: boolean
  token: string | null
  onClose: () => void
  /** 可选：首跑提示词附带入场码 */
  inviteCode?: string | null
  role?: string | null
  apiBase?: string | null
}

/**
 * FOOLPROOF F2/F3/F4：签 Key 成功挡层——完整身份钥匙 + 复制 + 启动提示词。
 * 禁止只 Toast / 埋页内 Alert。
 */
export default function IdentityKeyRevealModal({
  open,
  token,
  onClose,
  inviteCode,
  role,
  apiBase,
}: IdentityKeyRevealModalProps) {
  const [copied, setCopied] = useState(false)
  const copiedRef = useRef(false)

  const markCopied = useCallback(() => {
    copiedRef.current = true
    setCopied(true)
  }, [])

  const onCopyKey = async () => {
    if (!token) return
    try {
      await navigator.clipboard.writeText(token)
      markCopied()
      message.success('已复制身份钥匙')
    } catch {
      message.warning('复制失败，请手动选中复制')
    }
  }

  const onCopyFirstRun = async () => {
    if (!token) return
    try {
      const prompt = buildFirstRunPrompt({
        token,
        apiBase: apiBase || getPublicApiBase(),
        inviteCode,
        role,
      })
      await navigator.clipboard.writeText(prompt)
      markCopied()
      message.success('已复制启动提示词（含身份钥匙，勿转发公开群）')
    } catch (e) {
      message.error(e instanceof Error ? e.message : '复制启动提示词失败')
    }
  }

  const requestClose = () => {
    if (!copiedRef.current && token) {
      Modal.confirm({
        title: '尚未复制身份钥匙',
        content: '关闭后明文不再显示（列表只留末四）。确定关闭？',
        okText: '仍要关闭',
        cancelText: '继续查看',
        okButtonProps: { danger: true },
        onOk: () => {
          copiedRef.current = false
          setCopied(false)
          onClose()
        },
      })
      return
    }
    copiedRef.current = false
    setCopied(false)
    onClose()
  }

  return (
    <Modal
      title="身份钥匙（仅此一次）"
      open={open && Boolean(token)}
      onCancel={requestClose}
      footer={null}
      destroyOnHidden
      centered
      maskClosable={false}
      keyboard={false}
      width={560}
    >
      {token ? (
        <Space direction="vertical" size="middle" style={{ width: '100%' }}>
          <Alert
            type="warning"
            showIcon
            message="含密钥，勿发公开群 / 勿写入可转发入场链接"
            description="身份钥匙（ymk_）证明 Bot 是谁；入场码（ymt_）只决定进哪队。两码分家，勿混用。"
          />
          <div>
            <Typography.Text type="secondary" style={{ fontSize: 12, display: 'block', marginBottom: 6 }}>
              完整身份钥匙（刷新后不可再取）
            </Typography.Text>
            <Typography.Paragraph
              code
              copyable={{ text: token, onCopy: () => markCopied() }}
              style={{
                marginBottom: 0,
                wordBreak: 'break-all',
                whiteSpace: 'pre-wrap',
                fontSize: 13,
                padding: '10px 12px',
                background: 'var(--ant-color-fill-quaternary, #f5f5f5)',
                borderRadius: 6,
              }}
            >
              {token}
            </Typography.Paragraph>
          </div>
          <Space wrap>
            <Button type="primary" icon={<CopyOutlined />} onClick={() => void onCopyKey()}>
              复制身份钥匙
            </Button>
            <Button icon={<FileTextOutlined />} onClick={() => void onCopyFirstRun()}>
              复制启动提示词
            </Button>
            <Button onClick={requestClose}>{copied ? '完成' : '关闭'}</Button>
          </Space>
        </Space>
      ) : null}
    </Modal>
  )
}
