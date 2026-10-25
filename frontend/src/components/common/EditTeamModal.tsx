import { useEffect, useState } from 'react'
import { Form, Input, Modal, message } from 'antd'
import { patchTeam, type Team } from '../../api/teams'

type Props = {
  open: boolean
  workspaceId: string
  team: Team | null
  onClose: () => void
  onSaved: (team: Team) => void
}

/** 团队改名 · patchTeam（SETTINGS_EDIT S-R1） */
export default function EditTeamModal({ open, workspaceId, team, onClose, onSaved }: Props) {
  const [form] = Form.useForm<{ name: string }>()
  const [busy, setBusy] = useState(false)

  useEffect(() => {
    if (!open || !team) return
    form.setFieldsValue({ name: team.name })
  }, [open, team, form])

  const submit = async () => {
    if (!team) return
    let values: { name: string }
    try {
      values = await form.validateFields()
    } catch {
      return
    }
    const name = values.name.trim()
    if (!name) {
      message.warning('请填写团队名称')
      return
    }
    setBusy(true)
    try {
      const { team: next } = await patchTeam(workspaceId, team.id, { name })
      message.success(`已更新「${next.name}」`)
      onSaved(next)
      onClose()
    } catch (e: unknown) {
      message.error(e instanceof Error ? e.message : '保存失败')
    } finally {
      setBusy(false)
    }
  }

  return (
    <Modal
      title="编辑团队"
      open={open}
      onCancel={onClose}
      onOk={() => void submit()}
      okText="保存"
      cancelText="取消"
      confirmLoading={busy}
      destroyOnClose
      width={440}
    >
      <Form form={form} layout="vertical" requiredMark={false} style={{ marginTop: 8 }}>
        <Form.Item
          name="name"
          label="团队名称"
          rules={[{ required: true, message: '请填写团队名称' }]}
        >
          <Input
            placeholder="例如：产品协作队"
            maxLength={64}
            autoFocus
            onPressEnter={() => void submit()}
          />
        </Form.Item>
      </Form>
    </Modal>
  )
}
