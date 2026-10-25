# MOBILE_RESPONSIVE M7 · QA spotcheck

| 项 | 值 |
| --- | --- |
| **time** | 2026-10-25 08:46 Asia/Shanghai |
| **result** | **PASS†**（harsh Soft†：FE 源码 + 三帧 375×812 箱截对照，未强制起 Vite / 未 computerUse） |
| **debt** | `docs/NEXT_WAVE_DEBT_MOBILE_RESPONSIVE.md` |
| **FE commit** | `871bdf5ae1e0918c60dbf7bb7f71522d823af48c` · `feat(fe): MOBILE_RESPONSIVE M1–M5（≤639 抽屉壳/登录/看板/Issue + 375×812 箱截）` |
| **HEAD at run** | `33835114fa8dbd1ea6b4dd9868748dc22adb1b53` · `docs(settings): SETTINGS_EDIT 扩拍 — S-G2 项目 ws PATCH + 标签本刀` |
| **ancestor** | `871bdf5` ∈ ancestry of HEAD · YES |
| **截图复用** | YES · `docs/qa-shots/mobile-{shell,board,issue}.png`（同 FE `871bdf5`，未重截） |
| **viewport** | 375×812（债文箱截视口；三帧 `identify`/`PIL` 均 (375, 812) RGB） |
| **live Vite** | 未起（源码与箱截一致且 M1–M6 已勾/放行；Soft† 可接受） |
| **禁区** | 未改业务功能代码；未碰 SETTINGS_EDIT 清单 / Soft† 静默补；禁甩用户测 |

## Method

1. `git cat-file` / `merge-base --is-ancestor 871bdf5 HEAD` → FE 在 main 祖先链
2. `git show 871bdf5` → `AppLayout` / `KanbanBoard` / `IssueDrawer` + `board.css` / `issue-detail.css` / `login.css` / `page.css` + 三帧箱截 + debt 注记
3. Read(PNG) 三帧对照壳抽屉 / 看板横滑 / Issue 关钮
4. `rg` 断点 `max-width: 639`、热区 `44`、抽屉 / 汉堡 / `touch-action: manipulation`
5. 登录：`login.css` M2 窄轨 + `App.tsx` Login 路由 + 三帧已入站（登录后态）交叉证明

## Checklist（主路径 4 点）

- [x] **1. 壳抽屉（导航开合）** — `AppLayout.tsx`：`useIsMobileNarrow` ≤639 → 桌面 `Sider` 换 `Drawer` + 顶栏 `MenuOutlined`（`acb-nav-trigger` 44×44）；路由变关抽屉。箱截 `mobile-shell.png` 抽屉开（看板/列表/…/设置/Bot/平台管理）+ 半透明遮罩；`mobile-board.png` 汉堡可见。`page.css` 抽屉关钮/菜单项 ≥44 → Soft† PASS
- [x] **2. 登录可进** — `login.css` M2：卡宽 `min(400px, calc(100% - 32px))`、输入/主钮/Tab ≥44；`index.html` viewport；`App.tsx` `path="*"` → `LoginPage`。三帧均为登录后工作区态（无白屏/错栈）。LOGIN_REFRESH L7 已 Soft† PASS。无独立 `mobile-login.png`（债文三帧仅为 shell/board/issue）→ Soft† PASS（源码窄轨 + 入站态）
- [x] **3. 看板横滑开卡** — `board.css` ≤639：列 `flex: 0 0 min(280px, calc(100vw - 48px))` + `scroll-snap`；卡 `touch-action: manipulation`、`min-height: 44`。`KanbanBoard` 窄屏拖拽激活距 14（点按优先）。箱截 `mobile-board.png`：待办池满列 + 右侧「待处理」露出条 + 底横滑条 → Soft† PASS
- [x] **4. Issue 关（关闭/返回）** — `IssueDrawer` 窄宽 `100vw`；`issue-detail.css` 关钮 44×44、属性行 `min-height: 44`。箱截 `mobile-issue.png`：左上 X + `YM-24 v1`、主栏可滚、底属性可读 → Soft† PASS

## Evidence

| 路径 | 说明 |
| --- | --- |
| `docs/qa-shots/mobile-shell.png` | 25251 B · mtime 2026-10-25 08:42 Asia/Shanghai · 375×812 · 抽屉开 |
| `docs/qa-shots/mobile-board.png` | 26236 B · 同 · 汉堡 + 列横滑 |
| `docs/qa-shots/mobile-issue.png` | 24969 B · 同 · 近全屏 Issue + X 关 |
| `frontend/src/layouts/AppLayout.tsx` | 抽屉壳 / 汉堡 / 路由关抽屉 |
| `frontend/src/components/board/KanbanBoard.tsx` + `board.css` | 点按优先 + 横滑列 |
| `frontend/src/components/issue/IssueDrawer.tsx` + `issue-detail.css` | 100vw + 关钮≥44 |
| `frontend/src/pages/login.css` | M2 窄轨触控 |
| `frontend/src/styles/page.css` | M1 壳热区 |

## Gaps（非挡）

- 无独立窄屏登录帧：债文 M5 只锁三帧；登录以源码 M2 + 入站态 Soft† 覆盖。
- 未 live 点按横滑手势：箱截 + 激活距/touch-action 源码对照；executor 无 browser computerUse，未阻塞等人。

## FAIL redlines hit?

NONE

- 整页打不开 / 固定 Sider 挤出视口：NO（抽屉化）
- 横滑地狱 / 主钮点不到：NO（列 snap + 热区≥44）
- Issue 关不到：NO（X 44×44 + 100vw）
- 本刀改业务 / SETTINGS_EDIT / Soft 静默补：NO（仅本纪要 + debt M7 勾）

## SUMMARY

| # | 项 | 结果 |
| --- | --- | --- |
| 1 | 壳抽屉开合 | PASS† Soft† |
| 2 | 登录可进 | PASS† Soft† |
| 3 | 看板横滑开卡 | PASS† Soft† |
| 4 | Issue 关 | PASS† Soft† |

**Overall = PASS†**

给产品：M7 绿，请关 **M8**。
