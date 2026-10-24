import { Alert } from 'antd'
import { preferMock, setForceMock } from '../../api/mockData'

export function MockBanner() {
  if (!preferMock()) return null
  return (
    <Alert
      type="warning"
      showIcon
      banner
      message="演示数据"
      description="已强制使用 mock（localStorage ai-collab-board:force-mock=1 或 VITE_USE_MOCK=1）。API 失败不再静默回退。"
      action={
        <a
          onClick={() => {
            setForceMock(false)
            location.reload()
          }}
        >
          关闭强制 mock
        </a>
      }
      style={{ marginBottom: 12 }}
    />
  )
}
