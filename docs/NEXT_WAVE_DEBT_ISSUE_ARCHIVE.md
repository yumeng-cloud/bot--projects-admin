# 下一刀 · Issue 批量归档（完成列）

> 2026-10-24 20:52 Asia/Shanghai · 产品体验  
> **用户需求（总指挥转发）：**「完成」列/视图批量归档；归档后默认列表/看板/接口不返回；不进统计。  
> 对标：本仓 **Project** 已有 `archived_at` + `include_archived`；Issue **尚无** 归档字段。

## 一句话

Done（及可选 Canceled）可一键/批量归档；默认视图与计数像从没发生过；只有显式参数或「已归档」视图才能拉回。

## 字段裁定（冻结）

| 项 | 裁定 |
| --- | --- |
| **列** | `issues.archived_at TEXT NULL`（ISO8601 UTC，与 `projects.archived_at` 同形） |
| **语义** | `NULL` = 未归档（默认可见）；非空 = 已归档 |
| **与 status** | **正交**：归档不改 `status`（仍为 `done` / `canceled` 等）；取消归档只清 `archived_at` |
| **谁可归档** | WS `admin` / `member`（与改 Issue 同权）；Bot 可调同一 API |
| **可选后置** | `archived_by` 不做本刀；activity 记 `issue.archived` / `issue.unarchived` 即可 |
| **迁移** | SQLite + Postgres 同步；新文件 `012_issues_archived_at.sql`（或顺位号） |

## 默认过滤规则（冻结）

| 表面 | 默认 | 显式参数 |
| --- | --- | --- |
| **列表 / 看板 / 搜索 / Cycle 内 Issue 列表** | `archived_at IS NULL` | `include_archived=1\|true` → 含归档；`archived_only=1\|true` → **仅**归档（与 include 互斥时 **archived_only 优先**） |
| **GET 单条**（id / identifier） | **仍 200 返回**（深链/通知可开），body 带 `archived_at`；FE 顶栏标「已归档」 | 不要求 query |
| **创建 / 更新其它字段** | 已归档 Issue：**允许** PATCH 解档或改字段；看板拖拽默认列表里看不到故自然不可拖 | `POST …/archive` `POST …/unarchive` 或 PATCH `{ "archived": true\|false }`（二选一，**推荐 PATCH `archived` bool**，对齐 Project） |
| **批量** | `POST …/issues/batch-archive` body `{ "ids": ["…"], "archived": true\|false }`；仅本 WS；部分失败返回逐条结果 | — |

## 统计排除（冻结）

凡「进度 / 周期 / 列计数 / Pack 条」类聚合，**默认排除** `archived_at IS NOT NULL`：

- Board Pack / 列头计数 / `done/total`
- Cycle `estimate_sum`、完成率、未完成扫描
- 任何 workspace/project 级 Issue 计数 API

`include_archived` **不**自动把归档加回统计；统计若需含归档须另参 `stats_include_archived=1`（**本刀可不做**，默认永远剔归档）。

## 含 / 不含

| 做 | 不做 |
| --- | --- |
| Issue `archived_at` + 默认过滤 + 批量归档 API | 自动归档（定时把旧 done 清掉） |
| 「完成」列批量归档 + 列头「归档已完成」 | 归档整个 Project/Cycle（Project 已有） |
| 已归档视图（列表滤镜或子路由）+ 单条解档 | 改 status 语义、删 done 列 |
| 箱截 + API.md | 跨 WS 归档、软删物理删 |

## 验收（A1–A6 · B1–B4 · C1–C6 · D1–D3）

### A · Schema / API（Go · Issue 域）

| ID | 标准 | 谁 |
| --- | --- | --- |
| **A1** | 迁移加 `issues.archived_at`；model/JSON 露出 | Issue 域 / Go |
| **A2** | List/Board/Search/Cycle-issues **默认**剔归档；`include_archived` / `archived_only` 生效 | Issue 域 |
| **A3** | PATCH Issue 支持 `{ "archived": true\|false }`（或等价 archive/unarchive）；写 activity | Issue 域 |
| **A4** | `POST …/issues/batch-archive`（ids + archived）；权限与单条同 | Issue 域 |
| **A5** | GET 单条已归档仍 200 + `archived_at` | Issue 域 |
| **A6** | 单测：默认列表无归档；include/only；批量；解档回列表 | Issue 域 |

### B · 统计（Go）

| ID | 标准 | 谁 |
| --- | --- | --- |
| **B1** | Pack / 列计数 / done/total **不含**归档 | Issue 域 |
| **B2** | Cycle estimate_sum / 完成相关聚合 **不含**归档 | Issue 域 |
| **B3** | 单测：归档 done 后 total/done 下降（或持平，视原是否计入）且不再出现 | Issue 域 |
| **B4** | API.md 写清过滤与统计规则 | Issue 域 + 种子 |

### C · FE（React）

| ID | 标准 | 谁 |
| --- | --- | --- |
| **C1** | 「完成」列：多选 +「归档」；或列头「归档本列已完成」确认后批量 | React |
| **C2** | 归档后卡从看板/默认列表消失；Toast 可带「撤销」→ 解档 | React |
| **C3** | 列表/筛选：「已归档」入口（`archived_only`）；行上「取消归档」 | React |
| **C4** | 深开已归档 Issue：顶栏/Property 显示「已归档」+ 解档按钮 | React |
| **C5** | Pack/列数字与 BE 一致（归档后变） | React |
| **C6** | 箱截 `docs/qa-shots/issue-archive-*.png`（列批量、归档后看板、已归档视图、详情标）→ 产品勾 | React → 产品 |

### D · 就绪

| ID | 标准 | 谁 |
| --- | --- | --- |
| **D1** | BOT.md / API.md 参数表 | 种子 + Go |
| **D2** | QA：默认 API 无归档；显式参数可拉；统计剔除 | QA |
| **D3** | 产品目视箱截勾 C6 | 产品 |

## 派工

| 角色 | 职责 |
| --- | --- |
| **Issue 域后端** | A1–A6 · B1–B4 主责 |
| **Go 后端** | 协迁/路由注册缺口时补；不抢 Issue 域主笔 |
| **React 前端** | C1–C6 |
| **种子与文档** | D1 与 API 对齐 |
| **QA** | D2 |
| **产品体验** | 裁定已冻结；D3 勾选关单 |

## 产品勾选

- [ ] **A1–A6** BE 测绿 + 过滤
- [ ] **B1–B4** 统计剔除
- [ ] **C1–C6** FE + 箱截
- [ ] **D1–D3** 文档 / QA / 产品勾

## 关单进度

**未开干** · 债文已落，待 Issue 域 + React 开工。
