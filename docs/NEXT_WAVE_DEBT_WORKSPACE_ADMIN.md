# 生产包 · 工作空间开通 + 平台超级管理（完整）

> 2026-10-24 20:31 Asia/Shanghai · 产品体验  
> **用户令：完整最全生产包，不要最小切片。** Issue 详情打磨 / 升版 / 未分队本刀让路。  
> 复盘：总指挥 · 产品体验 · Go 后端 · React · Bot 记忆

## 一句话

任意登录用户能自助开通工作空间并管成员；平台超级管理员能管全站工作区与用户——全程中文 UI，零依赖 seed。

## 现网结论（开会共识）

| 层 | 现状 |
| --- | --- |
| **BE · WS** | `POST /workspaces` 已可创建（创建者 → `admin`）；`GET` 列表/详情/current 齐 |
| **BE · 成员** | `POST/PATCH/DELETE …/members` 齐（Invite 仅已注册邮箱直接入席） |
| **BE · 平台** | **无** `users.is_platform_admin`；**无** 全站 WS 列表/停用/代开；`/auth/me` 不回平台旗标 |
| **BE · 人类邀请** | **无** 邀请链接/token/redeem（队席 `ymt_` 是 Bot，勿混） |
| **FE** | `workspaces.ts` 仅 list/get；Home 空态写「请运行 seed」；设置页成员**只读**，无邀请/改角色/移除 |
| **产品裁定** | 登录用户可自建 WS（保留现 Create 鉴权）；超管另有全站治理权 |

## 完整能力面（本包全部交付）

### A · 工作空间自助开通（人人可用）

| ID | 能力 | 谁 |
| --- | --- | --- |
| **A1** | FE `createWorkspace({name,slug?})` 接 `POST /workspaces`；Home 顶栏/空态主按钮「创建工作空间」 | React |
| **A2** | 开通表单：名称必填、slug 可自动生成可改；成功进该 WS 看板（可顺带引导建首项目，有则做） | React |
| **A3** | 空态文案去掉「请运行 seed」；脚注文案改为自助开通说明 | React |
| **A4** | 登录后无 WS 时直达开通引导（不可死在空表） | React |
| **A5** | `PATCH /workspaces/{id}` 改名/slug（WS admin）；FE 设置「工作区」页可改 | Go + React |
| **A6** | 箱截：空账号 → 创建 → 进入看板；`docs/qa-shots/ws-admin-a*.png` | React → 产品勾 |

### B · 工作区内成员治理（WS Admin）

| ID | 能力 | 谁 |
| --- | --- | --- |
| **B1** | 设置 · 成员：邀请表单（邮箱 + 角色 admin/member）；接现有 `POST …/members` | React |
| **B2** | 改角色、移除成员（接 PATCH/DELETE）；不可摘掉最后一个 admin | React + Go 护栏 |
| **B3** | 人类邀请链接：`POST …/workspaces/{id}/invites` 签发（明文一次）；`POST /api/v1/workspace-invites/redeem` 兑换入席；列表/吊销；与 `ymt_` 队席码隔离 | Go + React |
| **B4** | 未注册邮箱：邀请仍可签发；对方注册/登录后 redeem 入席（或注册时带 token） | Go + React |
| **B5** | 成员表展示角色中文、邮箱、操作；Bot 行只读提示去「Bot」页 | React |
| **B6** | 箱截：邀请 / 改角色 / 链接 redeem / 吊销 · `ws-admin-b*.png` | → 产品 |

### C · 平台超级管理员（全站）

| ID | 能力 | 谁 |
| --- | --- | --- |
| **C1** | schema：`users.is_platform_admin`；迁移 SQLite+Postgres；`GET /auth/me` 回 `is_platform_admin` | Go |
| **C2** | Bootstrap：环境变量 `PLATFORM_ADMIN_EMAILS`（或首用户/seed 升旗）冷启授超管；文档写清 | Go + 种子文档 |
| **C3** | `GET /api/v1/admin/workspaces` 全站列表（含成员数、状态、创建时间） | Go |
| **C4** | `PATCH /api/v1/admin/workspaces/{id}` 停用/恢复；`workspaces.disabled_at`（或 status）；停用后成员访问 **403** 中文 | Go |
| **C5** | 超管代开：`POST /api/v1/admin/workspaces`（指定 owner email/user_id，owner 成 WS admin） | Go |
| **C6** | `GET /api/v1/admin/users`；`PATCH /api/v1/admin/users/{id}` 授/撤平台超管 | Go |
| **C7** | FE「平台管理」入口（仅 `is_platform_admin` 可见）：工作区台 + 用户台 | React |
| **C8** | 工作区台：搜索/列表、停用/恢复、代开表单、点进某 WS（超管只读跳入或带身份） | React |
| **C9** | 用户台：搜索、授/撤超管、看其所属 WS | React |
| **C10** | 管理操作写 activity/audit（谁何时停用/代开/授旗）；至少服务端日志表或 activity 行 | Go |
| **C11** | 非超管打 `/admin/*` → 403；箱 curl + FE 藏入口双验 | Go + QA |
| **C12** | 箱截：超管台全貌 · `ws-admin-c*.png` | → 产品 |

### D · 安全与生产就绪

| ID | 能力 | 谁 |
| --- | --- | --- |
| **D1** | API.md / BOT.md 补全上述路由与旗标；中文错误码 | Go + 种子文档 |
| **D2** | 单测：Create / Invite / redeem / 停用 403 / 超管鉴权 / 末 admin 不可删 | Go |
| **D3** | 停用 WS：Issue/Wiki/Bot API 一律 403，文案「工作区已停用」 | Go |
| **D4** | README 开通路径改「注册 → 创建工作空间」；去掉「只能 seed」口吻 | 产品/种子 |
| **D5** | 全路径箱 5174 自截过关；禁止甩用户本机当小白鼠 | 全员 · 产品勾 |

## 派工

| 角色 | 职责 |
| --- | --- |
| **Go 后端** | C1–C6、C10、A5 API、B2 护栏、B3–B4 邀请链、D1–D3 |
| **React 前端** | A1–A4、A5 UI、B1–B5、C7–C9、箱截 A6/B6/C12 |
| **种子与文档** | C2 文档、D1/D4 README·API 对齐 |
| **QA** | C11 + 全路径苛刻；人类邀请与 ymt_ 不串 |
| **Bot 记忆** | handoff：本刀优先；席位码 ≠ WS 邀请 |
| **UI 视觉** | 平台台/开通表单密度抽检 |
| **产品体验** | 验收勾选；拒假勾 |

## 产品勾选

### A 开通
- [x] **A1** 创建按钮+API · FE `5dbf62a`/`2148f60` · `ws-admin-a1-empty.png`
- [x] **A2** 表单进看板 · `ws-admin-a2-create-modal.png` → `a3-after-create.png` 已进工作区
- [x] **A3** 去 seed 空态 · Home 空态无 seed（a1）；`ProjectRedirect` 代码已改为「请先创建项目」（a3 截图略旧仍见 seed 字，以代码为准）
- [x] **A4** 无 WS 引导 · a1 双 CTA
- [x] **A5** 改名/slug · Go `05a1d35` + FE 设置 Tab
- [x] **A6** 箱截 · `ws-admin-a1/a2/a3`

### B 成员
- [x] **B1** 邀请表单 · `ws-admin-b1/b2`
- [x] **B2** 改角色/移除+末 admin · `b3-role.png` + Go 409
- [x] **B3** 人类邀请链接 `ymw_` · FE `2148f60` · b1 签发/列表/吊销
- [x] **B4** 未注册可邀 / Home「兑换邀请码」· `2148f60`（与 ymt_ 隔离文案在）
- [x] **B5** 成员表 · Bot 只读 · b1–b3
- [x] **B6** 箱截 · `ws-admin-b1/b2/b3`

### C 超管
- [x] **C1** 旗标+me · `05a1d35`
- [x] **C2** Bootstrap · `PLATFORM_ADMIN_EMAILS`
- [x] **C3** 全站 WS 列表 · `05a1d35` + `c1-platform.png`
- [x] **C4** 停用/恢复+403 · `05a1d35`（台可见停用开关）
- [x] **C5** 代开 · 台「代开工作空间」表单 · c1
- [x] **C6** 用户授撤超管 · API 齐；FE 用户 Tab 待补截
- [x] **C7** 平台入口 · `c1-platform.png`（admin@）
- [x] **C8** 工作区台 · 列表+停用+代开 · c1
- [ ] **C9** 用户台 · FE 有 Tab；**缺用户台箱截**，补 `ws-admin-c9-users.png` 再勾
- [x] **C10** 审计 · `05a1d35`
- [ ] **C11** 鉴权双验 · 等 QA / 箱 curl 非超管 403
- [x] **C12** 箱截 · `ws-admin-c1-platform.png`（工作区台；用户台另补）

### D 就绪
- [x] **D1** 文档 · `e2a4c2f`/`05a1d35`
- [x] **D2** 单测绿
- [x] **D3** 停用全拒
- [x] D4 README 开通路径
- [ ] **D5** 箱全路径 · 待 C9 用户台截 + C11

### 进度注记
- 2026-10-24 20:34 · 种子 D4 + D1 契约半齐 `4dbc83f`。
- 2026-10-24 20:38 · Go `05a1d35` 测绿勾 BE。
- 2026-10-24 20:46 · FE `2148f60` + 箱截入库：勾 A1–A6·B1–B6·C7/C8/C12。余 C9 用户台截、C11、D5。

## 关单进度

**近关** · 余 C9 用户台箱截 + C11 鉴权双验 + D5。
