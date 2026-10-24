# WORKSPACE_ADMIN · 苛刻复测清单（QA）

> 权威：`docs/NEXT_WAVE_DEBT_WORKSPACE_ADMIN.md`  
> 主验：**C11**（非超管 `/admin/*` 403）+ **人类邀请链与 `ymt_` 不串**  
> 等 Go/React 分段交后立刻跑；勿盲测。产品按 A→B→C→D 分段勾。

## 证据命名
- 主：`docs/qa-shots/WS_ADMIN-qa-retest-YYYYMMDD.txt`
- 箱截对齐债：`ws-admin-a*.png` / `ws-admin-b*.png` / `ws-admin-c*.png`

## 主验（红线）

### C11 · 鉴权双验
| # | 步骤 | PASS |
|---|------|------|
| C11-a | 非超管 JWT 打 `GET/PATCH/POST /api/v1/admin/*` → **403** 中文 | ✓ 2026-10-24 |
| C11-b | 超管 JWT 同路由可达（非 403） | ✓ 2026-10-24 |
| C11-c | FE：非超管**无**「平台管理」入口；直链 `/admin` 或等价路由被挡/藏 | ✓ 2026-10-24 code+shot |
| C11-d | curl 证据写入 retest 文件（请求路径+状态码+body 摘要） | ✓ WS_ADMIN-qa-retest-20261024.txt |

### 人类邀请 ≠ ymt_
| # | 步骤 | PASS |
|---|------|------|
| INV-1 | `POST …/workspaces/{id}/invites` 签发人类邀请；明文一次；列表可见 | ✓ 2026-10-24 |
| INV-2 | `POST /api/v1/workspace-invites/redeem` 入席；吊销后不可再用 | ✓ 2026-10-24 |
| INV-3 | 邀请 token **不是** `ymt_` 前缀；队席入场码路径不接受 WS 邀请 token | ✓ ymw_ 隔离 |
| INV-4 | `ymt_` 绑席路径不接受人类邀请 token；两边错误文案人话可区分 | ✓ 2026-10-24 |
| INV-5 | 未注册邮箱可签发；注册/登录后 redeem 入席（B4） | ✓ 2026-10-24 |

## 分段抽核（交一段验一段）
- **A**：空账号创建 WS → 进看板；无「请运行 seed」；无 WS 引导
- **B**：成员邀请/改角色/移除；末 admin 不可删；Bot 行只读去 Bot 页
- **C**：旗标+me；全站列表；停用→成员 403「工作区已停用」；代开；授撤超管；审计行
- **D**：`go test`；文档路由齐；README 开通路径

## 关报
短报产品+总指挥：Overall · C11 · 邀请/ymt_ · 证据路径 · HEAD  
测绿可 push origin/main（站立令）。

## QA 复测记录
- 2026-10-24 20:48 Asia/Shanghai · 独立苛刻 PASS（C11-a..d + INV-1..5）
- 证据：`docs/qa-shots/WS_ADMIN-qa-retest-20261024.txt`
- Go `05a1d35` · FE `5dbf62a`/`2148f60` · API 重建重启 :8080
