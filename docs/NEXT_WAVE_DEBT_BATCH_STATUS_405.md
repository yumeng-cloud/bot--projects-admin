# Soft† · live batch-status 405

> 2026-10-25 02:17 Asia/Shanghai · 产品体验  
> **承接：** PRO_ADMIN 本波已关（`9e5d10e`）；C2 Soft†「改状态串行」遗留。  
> **本刀：** 让 **live** `POST …/issues/batch-status` 真通（非 405），FE 走真批量；405/404/501 仍串行 fallback。

## 背景

- 路由已在 `backend/internal/server/router.go`：`ws.Post("/issues/batch-status", …)`（C2-2 `3b8a0d6`）。
- 单测/集成绿；**live 箱**曾 405 → FE 只对 404/501 fallback，405 会硬失败或被迫串行。
- FE `IssueListPage` 优先 `batchUpdateIssueStatus`，fallback 串行 patch。

## 目标态

1. **Go / Issue 域：** 箱上跑的 server 含 batch-status；`curl -X POST` live → **非 405**（200 + results）。若进程旧则重启对齐 main。
2. **React：** fallback 认 **405**（与 404/501 同）；真批量成功时 Toast 可区分（可选，不挡）。
3. **QA：** live 抽：勾多条改状态 → 网络可见 `batch-status` 200（或证 fallback 仅在 405）；列表刷新正确。
4. 产品关。

## 非目标

- 不开 E4 / 新人摩擦。
- 不重写批量语义；不改 archive/unarchive。

## 勾选

| ID | 标准 | 谁 |
| --- | --- | --- |
| **B405-1** | live POST batch-status ≠405 | Issue 域 / Go |
| **B405-2** | FE fallback 含 405 | React |
| **B405-3** | QA live 抽 | QA |
| **B405-4** | 产品关 | 产品体验 |

## 产品勾选

- [ ] **B405-1** live 真通
- [ ] **B405-2** FE 405 fallback
- [ ] **B405-3** QA
- [ ] **B405-4** 关单

## 派工

| 角色 | 首刀 |
| --- | --- |
| **Issue 域后端** | B405-1（核路由+箱重启+curl） |
| **Go 后端** | 协 Issue；进程/二进制对齐 main |
| **React 前端** | B405-2 |
| **QA** | B405-3（等 1+2） |
| **产品体验** | 本裁定；B405-4 |

## 进度注记

- 2026-10-25 02:17 · PRO_ADMIN 关后产品拍本刀（总指挥建议①）。

## 关单进度

**开干** · 等 Issue B405-1 ∥ React B405-2。
