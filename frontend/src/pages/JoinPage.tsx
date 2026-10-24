/**
 * ONE_TAP C1 · 同源 FE /join?code=ymt_…&role=…
 * - 已登录：解析 query → 跳转席页并打开预填 JoinWizard
 * - 未登录：短引导（复制提示词 / 安装 / 登录后预填），不消耗码；不改坏 Go API 宿主页
 */
import { useEffect, useMemo, useState } from 'react'
import { Link, useNavigate, useSearchParams } from 'react-router-dom'
import { Alert, Button, Card, Space, Spin, Typography, message } from 'antd'
import { listMyWorkspaces } from '../api/workspaces'
import LoginPage from './LoginPage'
import {
  INVITE_CODE_PREFIX,
  buildBotJoinPrompt,
  getPublicApiBase,
  installCurlLine,
  parseJoinSearchParams,
  validateInviteCode,
} from '../utils/oneTapJoin'

type Props = {
  authed: boolean
  onAuth: () => void
}

function resolveWsId(ids: string[]): string | null {
  const saved = (typeof localStorage !== 'undefined' && localStorage.getItem('workspaceId')) || ''
  if (saved && ids.includes(saved)) return saved
  return ids[0] || null
}

export default function JoinPage({ authed, onAuth }: Props) {
  const [searchParams] = useSearchParams()
  const nav = useNavigate()
  const parsed = useMemo(() => parseJoinSearchParams(searchParams), [searchParams])
  const [busy, setBusy] = useState(false)
  const [showLogin, setShowLogin] = useState(false)

  const codeError = useMemo(() => {
    try {
      validateInviteCode(parsed.code)
      return null
    } catch (e) {
      return e instanceof Error ? e.message : '入场码无效'
    }
  }, [parsed.code])

  // 已登录：带参跳转席页，由 BotsRosterPage 打开预填向导
  useEffect(() => {
    if (!authed || codeError) return
    let cancelled = false
    ;(async () => {
      setBusy(true)
      try {
        const { workspaces } = await listMyWorkspaces()
        if (cancelled) return
        const ids = (workspaces || []).map((w) => w.id)
        const wsId = resolveWsId(ids)
        if (!wsId) {
          message.warning('请先创建或加入工作区，再打开入场链接')
          nav('/', { replace: true })
          return
        }
        const q = new URLSearchParams()
        q.set('join', '1')
        q.set('code', parsed.code)
        if (parsed.role) q.set('role', parsed.role)
        if (parsed.seat) q.set('seat', parsed.seat)
        if (parsed.api) q.set('api', parsed.api)
        localStorage.setItem('workspaceId', wsId)
        nav(`/w/${wsId}/settings/bots?${q.toString()}`, { replace: true })
      } catch (e) {
        if (!cancelled) {
          message.error(e instanceof Error ? e.message : '无法打开入场向导')
        }
      } finally {
        if (!cancelled) setBusy(false)
      }
    })()
    return () => {
      cancelled = true
    }
  }, [authed, codeError, parsed.code, parsed.role, parsed.seat, parsed.api, nav])

  const prompt = useMemo(() => {
    if (codeError) return ''
    try {
      return buildBotJoinPrompt({
        code: parsed.code,
        role: parsed.role || null,
        seat: parsed.seat || null,
        api: parsed.api || null,
      })
    } catch {
      return ''
    }
  }, [codeError, parsed])

  const install = useMemo(
    () => installCurlLine(parsed.api || getPublicApiBase()),
    [parsed.api],
  )

  const copy = async (text: string, ok: string) => {
    try {
      await navigator.clipboard.writeText(text)
      message.success(ok)
    } catch {
      message.error('复制失败，请手动选择文本')
    }
  }

  if (authed) {
    return (
      <div style={{ minHeight: '100%', display: 'grid', placeItems: 'center', padding: 24 }}>
        <Spin tip="正在打开入场向导…" />
      </div>
    )
  }

  if (showLogin) {
    return (
      <div>
        <div style={{ maxWidth: 480, margin: '16px auto 0', padding: '0 16px' }}>
          <Button type="link" onClick={() => setShowLogin(false)} style={{ paddingLeft: 0 }}>
            ← 返回入场引导
          </Button>
        </div>
        <LoginPage onAuth={onAuth} />
      </div>
    )
  }

  return (
    <div style={{ minHeight: '100%', display: 'grid', placeItems: 'center', padding: 24 }}>
      <Card style={{ width: '100%', maxWidth: 520, boxShadow: '0 1px 2px rgba(0,0,0,0.04)' }}>
        <Typography.Title level={3} style={{ marginTop: 0, marginBottom: 4 }}>
          一键入场
        </Typography.Title>
        <Typography.Paragraph type="secondary" style={{ marginBottom: 12 }}>
          登录看板后将预填「加入班底」向导（只贴 {INVITE_CODE_PREFIX}…）。本页不消耗入场码。
        </Typography.Paragraph>

        {codeError ? (
          <Alert type="error" showIcon message={codeError} style={{ marginBottom: 12 }} />
        ) : (
          <Space direction="vertical" size="small" style={{ width: '100%', marginBottom: 12 }}>
            <Typography.Text>
              入场码 <Typography.Text code>{parsed.code}</Typography.Text>
            </Typography.Text>
            {parsed.role ? (
              <Typography.Text>
                角色 <Typography.Text code>{parsed.role}</Typography.Text>
              </Typography.Text>
            ) : null}
            {parsed.seat ? (
              <Typography.Text>
                席位 <Typography.Text code>{parsed.seat}</Typography.Text>
              </Typography.Text>
            ) : null}
          </Space>
        )}

        <Alert
          type="info"
          showIcon
          style={{ marginBottom: 14 }}
          message="人类路径：登录后自动打开向导并预填入场码。Bot 路径：复制下方提示词（含一键安装）。"
        />

        {!codeError ? (
          <Space direction="vertical" size="middle" style={{ width: '100%' }}>
            <div>
              <Typography.Text strong>① 一键安装 ymbot（若无）</Typography.Text>
              <Typography.Paragraph
                code
                copyable={{ text: install }}
                style={{ margin: '6px 0 0', whiteSpace: 'pre-wrap', fontSize: 12 }}
              >
                {install}
              </Typography.Paragraph>
            </div>
            <div>
              <Typography.Text strong>② Bot 提示词</Typography.Text>
              <Typography.Paragraph type="secondary" style={{ margin: '2px 0 6px', fontSize: 12 }}>
                可转发；只含入场码，不含身份钥匙 ymk_。
              </Typography.Paragraph>
              <Button
                type="default"
                block
                disabled={!prompt}
                onClick={() => void copy(prompt, '已复制 Bot 提示词')}
              >
                复制 Bot 提示词
              </Button>
            </div>
          </Space>
        ) : null}

        <Space direction="vertical" size="small" style={{ width: '100%', marginTop: 16 }}>
          <Button type="primary" block loading={busy} disabled={!!codeError} onClick={() => setShowLogin(true)}>
            登录后打开向导
          </Button>
          <Typography.Paragraph type="secondary" style={{ margin: 0, fontSize: 12, textAlign: 'center' }}>
            API 宿主说明页仍在 <Typography.Text code>{getPublicApiBase()}/join</Typography.Text>
            ；本页为看板同源深链。
          </Typography.Paragraph>
          <div style={{ textAlign: 'center' }}>
            <Link to="/">回首页</Link>
          </div>
        </Space>
      </Card>
    </div>
  )
}
