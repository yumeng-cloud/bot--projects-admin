# SETTINGS_EDIT S-Q1 · QA 活点复测纪要

> **time_AsiaShanghai:** 2026-10-25 09:13 Asia/Shanghai  
> **Overall:** **FAIL**（SQ1-2 顶栏/侧栏未即时刷新）  
> **HEAD_at_run:** `23148da`（含 Go `bd394ff` + FE `0e15fa0`）  
> **Go S-G1/S-G2:** `bd394ffae714674faeec85c514a43f1b912ad027`  
> **React S-R1～R5:** `0e15fa01c1bc71bba1b0152978756420d299670d`  
> **环境:** backend `/tmp/ai-board-server-sq1` → `:8080`；Vite dev `:5175` proxy `/api→8080`  
> **账号:** `admin@yumeng.tech` / Admin123!（JWT 注入 localStorage）  
> **方法:** Playwright Chromium 自点全表（禁甩用户）；辅 curl API 矩阵

## 结论（harsh）

| Overall | **FAIL** |
| --- | --- |
| 可关单？ | **否** — 钉 React：工作区保存后 `AppLayout.ws` 不刷新，侧栏品牌名仍旧 |
| S-G2 覆盖 | **是** — `bd394ff` 补挂 `PATCH /workspaces/{ws}/projects/{id}`；API 200 + UI SQ1-3 PASS |
| 标签 S-R5 | **是** — `0e15fa0` 标签编辑入口；UI SQ1-5 PASS（Toast+列表名变） |

## 矩阵结果

| ID | 结果 | 证据 | 备注 |
| --- | --- | --- | --- |
| **SQ1-1** 团队改名 | **PASS** | `settings-edit-team-rename.png` | Toast `已更新「SQ1-UI-Team-202610250906」`；列表 `listHas=1` |
| **SQ1-2** 工作区改名 | **FAIL** | `settings-edit-workspace-rename.png` | Toast `工作区已更新` ✓；表单回显新名 ✓；**侧栏仍 `育蒙科技`（旧名）** ✗。根因：`SettingsPage.onSaveWorkspace` 只 `setWsName`；`AppLayout` `ws` 仅在 `wsId` effect 拉取，保存后不失效 |
| **SQ1-3** 项目改名 | **PASS** | `settings-edit-project-rename.png` | Toast+列表；同时证明 **S-G2** ws 作用域 PATCH 通 |
| **SQ1-4** 成员角色 | **PASS** | `settings-edit-member-role.png` | 「设为管理员」→ Toast `角色已更新`（测后回滚成员） |
| **SQ1-5** 标签改名 | **PASS** | `settings-edit-label-rename.png` | Toast `已更新「SQ1-UI-Label-202610250906」`；列表刷新 —— **S-R5 覆盖** |
| **SQ1-6** 团队三钮回归 | **PASS** | `settings-edit-team-actions-regression.png` | 编辑+设为当前+成员+Bot 席均在；成员→`?tab=members`；Bot→`/settings/bots?team=`；点通过启用「设为当前」 |

## API 辅证（同会话）

- `PATCH /api/v1/workspaces/{ws}/teams/{id}`（S-G1）— 二进制含路由；UI 已点通  
- `PATCH /api/v1/workspaces/{ws}/projects/{id}`（**S-G2**）— HTTP **200**（探针改名后已恢复）  
- `PATCH /api/v1/workspaces/{ws}` — 200；UI Toast 成功  
- 标签 PATCH — UI 点通  

## 缺陷指派

| 优先级 | 指派 | 缺陷 | 修复建议（给 React，非本刀改码） |
| --- | --- | --- | --- |
| **P0** | **React（S-R2）** | 工作区保存成功后侧栏/顶栏 `ws?.name` 不更新 | `onSaveWorkspace` 成功后通知 `AppLayout` 刷新 `ws`（context/event/`listMyWorkspaces` 重拉，或把 `setWs` 下传）；勿只改 Settings 本地 state |
| — | Go | 无本刀阻塞 | S-G1/S-G2 已齐 |
| — | 产品 | 等 React 修侧栏刷新 → QA 复测 SQ1-2 → 再 S-P1 | |

## Soft† / 非阻塞笔记

- 顶栏右侧文案「工作区」为切换器占位，**真正显示名在侧栏品牌**（`AppLayout` `siderBrand` → `ws?.name`）。验收以侧栏为准。  
- 保存时 Admin 改名可能联动 slug（截图见 `ui-202610250906`）；测后已 API 恢复 `育蒙科技` / `yumeng`。  
- Vite `:5174` preview 曾缺 S-R1 入口；活点用 `:5175` 源码 dev。  
- 未碰 P-Q1 / MOBILE 证据文件。

## 产物清单

- `docs/qa-shots/settings-edit-team-rename.png`
- `docs/qa-shots/settings-edit-team-actions-regression.png`
- `docs/qa-shots/settings-edit-workspace-rename.png`
- `docs/qa-shots/settings-edit-project-rename.png`
- `docs/qa-shots/settings-edit-member-role.png`
- `docs/qa-shots/settings-edit-label-rename.png`
- 本纪要

## 给产品（关单句）

**S-Q1 FAIL：团队/项目/成员/标签/三钮回归均真点 PASS，S-G2+标签已覆盖；工作区保存 Toast+表单 OK 但侧栏名不刷新——钉 React 修 AppLayout 后复测再关 S-P1。**
