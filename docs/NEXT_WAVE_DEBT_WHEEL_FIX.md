# 全链路车轮修复（最小可修包）

> 2026-10-24 21:11 Asia/Shanghai · 产品体验  
> **用户令（总指挥转）：** 登录→WS/队→看板/Issue→归档→Bot席/入场→pins/顶栏换队→超管；扫软债+冒烟缺口，裁最小包定稿即派，免再问。  
> **不做：** T7-4/5 升版、未分队大入口、跨队搬家大 UI。

## 一句话

演示全路径不卡轮：已知软债能修的本刀修掉；各段冒烟有箱证；坏点当场补，不甩用户本机。

## 现网软债盘点（本刀取舍）

| 来源 | 软债 | 本刀 |
| --- | --- | --- |
| ISSUE_ARCHIVE | 解档 API/UI；归档专面；canceled 归档；ListRelations→已归档边 | **做解档 + Relations 边过滤**；专面/canceled **后置** |
| WORKSPACE_ADMIN | 用户台搜索；所属 WS 列；a3 截仍见 seed 字 | **做搜索 + 所属 WS（BE 补 memberships）+ 重拍 a3** |
| TEAM_SWITCH / pins / T6I | 已关 | **只冒烟回归，不重开** |
| Bot 席 / handoff | 已有通道 | **最小入场冒烟**（绑席→focus 可见） |
| ISSUE_POOL.md 勾选 | 陈旧草稿；以 T6I 已关为准 | **不派** |

## 含 / 不含

| 做 | 不做 |
| --- | --- |
| 解档 `POST …/issues/unarchive` + FE 按钮 | 自动归档、定时清库 |
| Relations 列表默认不展示指向已归档 Issue 的边（或边带 `archived` 且 FE 隐藏） | 物理删边 |
| 超管用户台：搜索 + 所属 WS | 用户细权编辑大页 |
| a3 / 空项目引导去净 seed | 重做开通 IA |
| 全链路 QA 冒烟一条龙证据 | T7 升版 dry-run |

## 验收

### A · 进场 / WS（React 主）

| ID | 标准 | 谁 |
| --- | --- | --- |
| **A1** | `ProjectRedirect` / 空项目文案无「seed」；重拍 `ws-admin-a3-after-create.png` | React |
| **A2** | 箱冒烟：注册/登录 → 无 WS → 创建工作空间 → 进看板（可复用既有截，缺则补 `wheel-a2-onboard.png`） | React → QA |
| **A3** | 人类邀请 redeem 仍通（`ymw_`）；抽核 b5 或补 `wheel-a3-redeem.png` | React / QA |

### B · Issue 归档补轮（Issue 域 + React）

| ID | 标准 | 谁 |
| --- | --- | --- |
| **B1** | `POST /api/v1/workspaces/{ws}/issues/unarchive` `{ids}`；幂等；activity `issue.unarchived`；已归档方可解档 | Issue 域 |
| **B2** | FE：已归档 Drawer / 详情有「取消归档」；成功后回默认列表可见 | React |
| **B3** | ListRelations（及 FE 关联行）默认不展示目标/源已归档的边；单测或 QA 证 | Issue 域 + React |
| **B4** | 箱截 `wheel-b-unarchive-*.png`（归档→解档→卡回看板） | React → 产品 |

### C · 队 / 席 / pins 冒烟（React + 记忆）

| ID | 标准 | 谁 |
| --- | --- | --- |
| **C1** | 顶栏切队 A↔B：看板/列表池变；Bot 席数跟队（回归 TEAM_SWITCH） | React + QA |
| **C2** | pins 跟当前队不串（抽核既有 pins 箱证或补 `wheel-c2-pins.png`） | React / QA |
| **C3** | Bot 席最小入场：绑席或 handoff get 可见 focus/开放 Issue（记忆主证 + QA 抽） | Bot 记忆 + QA |

### D · 超管补齐（Go + React）

| ID | 标准 | 谁 |
| --- | --- | --- |
| **D1** | `GET /admin/users` 支持 `q` 搜索（邮箱/名）；FE 用户台 Input 搜索 | Go + React |
| **D2** | 用户行展示所属 WS（BE 回 `workspaces:[{id,name,slug}]` 或等价；FE 列） | Go + React |
| **D3** | 箱截 `wheel-d-users-search.png`（搜索命中 + 所属 WS 可见） | React → 产品 |

### E · 全链路车轮（QA 主 · 产品勾）

| ID | 标准 | 谁 |
| --- | --- | --- |
| **E1** | 证据 `docs/qa-shots/WHEEL_FIX-smoke-YYYYMMDD.txt` 覆盖：登录→WS→队→看板 Issue→归档→解档→切队→Bot席抽核→超管用户台 | QA |
| **E2** | 无阻塞 Defect；软注可记后置 | QA |
| **E3** | 产品目视关键箱截勾选关单 | 产品 |

## 派工

| 角色 | 职责 |
| --- | --- |
| **Issue 域后端** | B1、B3（BE） |
| **Go 后端** | D1/D2 API；路由缺口 |
| **React 前端** | A1–A3、B2/B4、C1–C2、D1–D3 FE、箱截 |
| **Bot 记忆** | C3 入场/handoff 最小证 |
| **QA** | E1–E2；抽 A/C；B/D 苛刻 |
| **产品体验** | 本裁定；E3 勾选 |

## 产品勾选

- [ ] **A1** a3 去 seed 重拍
- [ ] **A2** 开通进场冒烟
- [ ] **A3** redeem 抽核
- [ ] **B1** unarchive API
- [ ] **B2** FE 解档
- [ ] **B3** Relations 剔已归档边
- [ ] **B4** 解档箱截
- [ ] **C1** 顶栏切队回归
- [ ] **C2** pins 跟队
- [ ] **C3** Bot 席入场
- [ ] **D1** 用户搜索 · BE `aa1a004`（`?q=`）；**FE 待接**
- [ ] **D2** 所属 WS · BE `aa1a004`（`workspaces[]`）；**FE 待接**
- [ ] **D3** 超管箱截
- [ ] **E1** 全链路 smoke txt
- [ ] **E2** QA 无阻塞
- [ ] **E3** 产品关单

## 进度注记

- 2026-10-24 21:13 · Go D1/D2 API `aa1a004` 测绿；等 React FE。

## 关单进度

**进行中** · Go D1/D2 BE `aa1a004` 齐；React A1 码 `9eb025b`（截待）；Issue B / QA E 进行中。
