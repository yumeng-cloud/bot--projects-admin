import { useEffect, useState } from 'react'
import { Input, Modal, Typography, message } from 'antd'
import { ApiError } from '../../api/client'
import { createTeam, type Team } from '../../api/teams'

type Props = {
  open: boolean
  workspaceId: string
  onClose: () => void
  onCreated: (team: Team) => void
}

/** 顶栏空态用：只填队名的简易建队（完整模板流仍在 Bot 席） */
export default function CreateTeamQuickModal({ open, workspaceId, onClose, onCreated }: Props) {
  const [name, setName] = useState('')
  const [busy, setBusy] = useState(false)

  useEffect(() => {
    if (open) setName('')
  }, [open])

  const submit = async () => {
    const trimmed = name.trim()
    if (!trimmed) {
      message.warning('请填写队名')
      return
    }
    setBusy(true)
    try {
      const { team } = await createTeam(workspaceId, { name: trimmed }, crypto.randomUUID())
      message.success(`已创建团队「${team.name}」`)
      onCreated(team)
      onClose()
    } catch (e: unknown) {
      if (e instanceof ApiError && (e.status === 404 || e.status === 501)) {
        message.warning('创建团队接口尚未开通，请稍后再试或去 Bot 席建队')
        return
      }
      message.error(e instanceof ApiError ? e.message : '创建团队失败')
    } finally {
      setBusy(false)
    }
  }

  return (
    <Modal
      title="创建团队"
      open={open}
      onCancel={onClose}
      onOk={() => void submit()}
      okText="创建"
      cancelText="取消"
      confirmLoading={busy}
      destroyOnClose
      width={400}
    >
      <Typography.Paragraph type="secondary" style={{ marginBottom: 8, fontSize: 13 }}>
        先起个队名即可开干；需要模板班底可稍后到 Bot 席补。
      </Typography.Paragraph>
      <Input
        value={name}
        placeholder="例如：产品一组"
        maxLength={64}
        autoFocus
        onChange={(e) => setName(e.target.value)}
        onPressEnter={() => void submit()}
      />
    </Modal>
  )
}
