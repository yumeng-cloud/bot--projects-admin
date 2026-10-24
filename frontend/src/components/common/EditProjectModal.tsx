import { useEffect, useState } from 'react'
import { Form, Input, Modal, message } from 'antd'
import { patchProject, type Project } from '../../api/projects'

type Props = {
  open: boolean
  workspaceId: string
  project: Project | null
  onClose: () => void
  onSaved: (project: Project) => void
}

/** 改名 / 简介 · patchProject（PROJECT_ADMIN A3） */
export default function EditProjectModal({ open, workspaceId, project, onClose, onSaved }: Props) {
  const [form] = Form.useForm<{ name: string; description?: string }>()
  const [busy, setBusy] = useState(false)

  useEffect(() => {
    if (!open || !project) return
    form.setFieldsValue({
      name: project.name,
      description: project.description || '',
    })
  }, [open, project, form])

  const submit = async () => {
    if (!project) return
    let values: { name: string; description?: string }
    try {
      values = await form.validateFields()
    } catch {
      return
    }
    const name = values.name.trim()
    if (!name) {
      message.warning('请填写项目名称')
      return
    }
    setBusy(true)
    try {
      const { project: next } = await patchProject(workspaceId, project.id, {
        name,
        description: values.description?.trim() || '',
      })
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
      title="编辑项目"
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
          label="项目名称"
          rules={[{ required: true, message: '请填写项目名称' }]}
        >
          <Input placeholder="例如：产品协作" maxLength={64} autoFocus onPressEnter={() => void submit()} />
        </Form.Item>
        <Form.Item name="description" label="简介（可选）">
          <Input.TextArea rows={2} maxLength={200} placeholder="一句话说明这个项目" />
        </Form.Item>
        {project ? (
          <TypographyHint prefix={project.key_prefix} />
        ) : null}
      </Form>
    </Modal>
  )
}

function TypographyHint({ prefix }: { prefix: string }) {
  return (
    <p style={{ margin: 0, fontSize: 12, color: 'var(--acb-text-secondary, #8b8b8b)' }}>
      编号前缀 <code>{prefix}</code> 创建后不可改
    </p>
  )
}
