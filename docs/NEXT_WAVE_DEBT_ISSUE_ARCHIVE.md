# 下一刀 · Issue 批量归档（完成列）

> 2026-10-24 20:53 Asia/Shanghai · 产品体验  
> **总指挥定稿 · 用户免再问 · 测绿直推。** 对齐 Project `archived_at` + `include_archived` 模式。

## 目标

1. 「完成(done)」列可多选批量归档  
2. 归档后默认一切列表/看板/搜索/Bot API/MCP **不返回**；仅 `include_archived=1` 或 `archived=only` 才拉  
3. 归档后 **不进统计**（pack 进度 done/total、周期计数、协作看板计数等）

## 契约（冻结）

| 项 | 定稿 |
| --- | --- |
| **DB** | `issues.archived_at TEXT NULL` + idx `(workspace_id, archived_at)`（表现无此列；对齐 Project） |
| **归档 API** | `POST /api/v1/workspaces/{ws}/issues/archive` body `{ "ids": ["uuid",…] }`：**仅允许当前 `status=done`**；幂等；写 activity `issue.archived` |
| **解档** | `POST …/issues/unarchive` **本包可选 / 可后置**；本包至少 archive |
| **List/board/search/relations** | 默认 `archived_at IS NULL`；`include_archived=1` → 全量；`archived=only` → 仅归档 |
| **GET 单条** | 仍可读（深链）；列表默认不见 |
| **统计** | SQL + FE `useBoardPackStats`：**排除** archived |
| **与 status** | 正交：归档不改 `status`（保持 `done`） |

## 软债（后置，不挡本包）

- 归档专面 / 已归档浏览 UI  
- `canceled` 态归档  
- unarchive 若本包未做则下一刀补

## 验收（A1–A6 · B1–B3 · C1–C5 · D1–D3）

### A · Schema / API（Issue 域主 · Go 补路由）

| ID | 标准 | 谁 |
| --- | --- | --- |
| **A1** | 迁移 `issues.archived_at` + idx `(workspace_id, archived_at)`；JSON 露出 | Issue 域 |
| **A2** | List / board / search / relations 默认剔归档；`include_archived=1` 全量；`archived=only` 仅归档 | Issue 域 |
| **A3** | `POST …/issues/archive` `{ids}`：仅 `status=done`；非 done → 中文 4xx；幂等；activity `issue.archived` | Issue 域 |
| **A4** | GET 单条已归档仍 200 + `archived_at` | Issue 域 |
| **A5** | Bot API / MCP 列表默认不返回归档（同源过滤） | Issue 域 + Go |
| **A6** | 单测：默认不见 / include / only / 非 done 拒 / 幂等 | Issue 域 |

### B · 统计

| ID | 标准 | 谁 |
| --- | --- | --- |
| **B1** | Pack / 列计数 / done/total 排除 archived | Issue 域 + React |
| **B2** | Cycle 等周期聚合排除 archived | Issue 域 |
| **B3** | 单测或箱 curl：归档后计数下降且列表无该单 | Issue 域 / QA |

### C · FE（React）

| ID | 标准 | 谁 |
| --- | --- | --- |
| **C1** | 完成列多选 + 批量归档（调 archive API） | React |
| **C2** | 默认 list/board **不传** include；归档后卡消失 | React |
| **C3** | Pack/`useBoardPackStats` 与 BE 一致（剔归档） | React |
| **C4** | 深开已归档：可见 `archived_at` 态（解档若未做则只读提示） | React |
| **C5** | 箱截 `docs/qa-shots/issue-archive-*.png`（多选归档、归档后看板、统计）→ 产品勾 | React → 产品 |

### D · 文档 / QA / 关单

| ID | 标准 | 谁 |
| --- | --- | --- |
| **D1** | API.md / BOT.md 一句：archive + 查询参数 | 种子 |
| **D2** | QA 苛刻：默认不见 / 统计不变（相对归档前剔掉）/ include 可拉 | QA |
| **D3** | 产品目视箱截勾 C5 | 产品 |

## 派工

| 角色 | 职责 |
| --- | --- |
| **Issue 域后端** | A1–A6 · B1–B2 主笔；测绿直推 |
| **Go 后端** | 路由注册 / 缺口 |
| **React 前端** | C1–C5 |
| **种子与文档** | D1 |
| **QA** | D2 |
| **产品体验** | 验收点本表；D3 勾选 |

## 产品勾选

- [ ] **A1** 迁移 + 字段
- [ ] **A2** 默认过滤 + include / only
- [ ] **A3** batch archive（仅 done）
- [ ] **A4** GET 单条可读
- [ ] **A5** Bot/MCP 同源默认剔
- [ ] **A6** 单测绿
- [ ] **B1** Pack/列计数剔归档
- [ ] **B2** Cycle 聚合剔归档
- [ ] **B3** 计数箱验
- [ ] **C1** 完成列多选归档
- [ ] **C2** 默认不拉归档
- [ ] **C3** FE 统计一致
- [ ] **C4** 深链已归档态
- [ ] **C5** 箱截入库
- [ ] **D1** API/BOT 一句
- [ ] **D2** QA 苛刻 PASS
- [ ] **D3** 产品关单

## 关单进度

**定稿已落** · 待 Issue 域 / React / 种子 / QA 按上表开工。测绿直推 origin/main。
