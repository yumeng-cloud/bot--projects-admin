# PRO_ADMIN · D3 超管用户台按所属 WS 筛

> 2026-10-25 02:08 Asia/Shanghai · 产品体验  
> **承接：** [`PRO_ADMIN_E2`](./NEXT_WAVE_DEBT_PRO_ADMIN_E2.md) 已关。  
> **本刀：** 超管用户台：按所属工作区筛选用户（`workspace_id` 或 FE 滤；BE 有则优先）。  
> **仍冻：** **D4 / G**（D4 箱截 D1/D2 已交 `pro-d-members.png` / `pro-d-users.png`，关本刀后再并关）。

## 目标态

1. 用户表顶：WS 筛（Select / 可搜索）；与现有邮箱名称搜索可叠加。
2. 优先走 API `workspace_id`（若已有）；否则 FE 按 `row.workspaces[].id` 滤，注明 Soft†。
3. 清空筛恢复全部；空结果人话。
4. 箱截更新或增注 `pro-d-users.png`（带 WS 筛可见即可）；QA 抽后产品关。

## 非目标

- 不开 D4 新截要求（复用已有）、不开 G。
- 不大改用户列表 API；Go 仅在缺 `workspace_id` 且 FE 不可靠时最小补。

## 勾选

| ID | 标准 | 谁 |
| --- | --- | --- |
| **D3-1** | WS 筛控件 + 与搜索叠加 | React（缺 API 喊 Go） |
| **D3-2** | 空结果人话；清空恢复 | React |
| **D3-3** | 箱截（users 带筛） | React → 产品 |
| **D3-4** | QA 抽 | QA |
| **D3-5** | 产品关（回写 PRO_ADMIN D3） | 产品体验 |

## 产品勾选

- [x] **D3-1** WS 筛 · FE Soft† `7def3a2` · 产品勾 · React 已交 Soft† FE（等产品勾）
- [x] **D3-2** 空结果/清空 · FE `7def3a2` · 产品勾 · React 已交（等产品勾）
- [x] **D3-3** 箱截 `pro-d-users.png`（带 WS 筛）· 产品勾 · React 已交（等产品勾）
- [x] **D3-4** QA · PASS† Soft† `f5b2780` · 产品勾
- [x] **D3-5** 关单 · 2026-10-25

## 派工

| 角色 | 首刀 |
| --- | --- |
| **React 前端** | D3-1～D3-3 |
| **Go 后端** | 仅缺 `workspace_id` 时最小补 |
| **QA** | D3-4 |
| **产品体验** | 本裁定；D3-5 |

## 进度注记

- 2026-10-25 02:14 · **D3-4 PASS† Soft† → D3-5 关单** QA `f5b2780` @ FE `7def3a2`。整包 **PRO_ADMIN D3 关**。并关 **D4**（`pro-d-members.png`/`pro-d-users.png` 已齐）。下一刀：**仅开 G** 验收门；勿开新功能。
- 2026-10-25 02:13 · **D3-1～D3-3 产品勾 Soft†** FE `7def3a2`：API 无 workspace_id → FE 按 workspaces[].id；Select 可搜与 q 叠加 + 清空/空结果人话 + `pro-d-users.png`。开 **QA D3-4**；D4/G 仍冻。
- 2026-10-25 02:08 · E2 关后产品拍开本刀；D4/G 仍冻。
- 2026-10-25 · React D3-1～D3-3：用户台顶 **所属 WS 筛**（Select 可搜 + allowClear）与邮箱/名称防抖 `q` 叠加；**API `GET /admin/users` 无 `workspace_id` → Soft† FE 按 `row.workspaces[].id` 滤**（用户行已有 workspaces 列表，可靠）；清空筛恢复全部 + 空结果人话；箱截 `docs/qa-shots/pro-d-users.png`（WS 筛可见）。**不喊 Go**。等 QA D3-4 / 产品 D3-5。**D4/G 仍冻**。

## 关单进度

**整包关 Soft†** · D3-1～D3-5 齐（FE `7def3a2` + QA `f5b2780`）。
