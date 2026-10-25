# SETTINGS_EDIT S-Q1 · 全表自点验收清单（骨架·等码）

> 2026-10-25 08:45 Asia/Shanghai · QA 骨架  
> **活点：** 2026-10-25 09:13 Asia/Shanghai · **Overall FAIL**（纪要 `SETTINGS_EDIT-S-Q1-qa-retest-20261025-0913.md`）  
> **债：** [`docs/NEXT_WAVE_DEBT_SETTINGS_EDIT.md`](../NEXT_WAVE_DEBT_SETTINGS_EDIT.md)  
> **码齐：** Go `bd394ff`（S-G1+S-G2）+ FE `0e15fa0`（S-R1～R5）  
> **硬约束：** 禁甩用户测；不问用户；MOBILE 并行不挡、勿改 MOBILE 证据；只写 QA 文档。

## 目标（债文 §QA 自点矩阵 1–6 + 勾选 S-Q1）

设置下 **工作区 / 项目 / 团队 / 成员 / 标签** 主路径各至少一改可自点通过；团队操作列「设为当前 / 成员 / Bot 席」回归不坏。保存成功 Toast + 列表即时刷新；失败人话；禁白屏/无响应按钮。

## 债文 QA 矩阵摘要

| # | 路径 | 动作 | 期望 |
| --- | --- | --- | --- |
| 1 | 设置·团队 | 编辑改名→保存 | 列表名变；Toast 成功 |
| 2 | 设置·工作区 | 改显示名→保存 | 顶栏/侧栏名变 |
| 3 | 设置·项目 | 编辑改名→保存 | 列表名变 |
| 4 | 设置·成员 | 改角色或确认现入口可用 | 成功或产品已标「无入口」 |
| 5 | 设置·标签 | 改一名→保存 | 成功或页不存在则记缺口 |
| 6 | 回归 | 团队「设为当前/成员/Bot 席」仍可用 | 不回归 |

对应交件：Go **S-G1**；React **S-R1**（团队编辑）· **S-R2**（工作区）· **S-R3**（项目）· **S-R4**（成员）· **S-R5**（标签）；QA **S-Q1**；产品 **S-P1** 关单。

## 前置（交件后填实）

| 项 | 值 / 门禁 |
| --- | --- |
| 债文 | `docs/NEXT_WAVE_DEBT_SETTINGS_EDIT.md` @`b49286c`（或含该文的后续 HEAD） |
| FE 入口 | `/w/{wsId}/settings` · Tab：`?tab=workspace\|projects\|teams\|members\|labels` |
| Admin | `admin@yumeng.tech` / seed → 浏览器登录（须 Admin；团队改名债文要求 Admin 可改） |
| API | 本地 `http://127.0.0.1:8080`；二进制须含 **S-G1** `PATCH /api/v1/workspaces/{ws}/teams/{teamID}`（最少 `name`） |
| 团队编辑 | FE **S-R1**：`SettingsTeamsPanel` 操作列「编辑」+ Modal 改名；保存刷新 |
| 工作区 | FE **S-R2**：工作区 Tab `patchWorkspace` 显示名（Admin 可调 slug） |
| 项目 | FE **S-R3**：`EditProjectModal` + `patchProject` |
| 成员 | FE **S-R4**：成员表「更多」→ 设为管理员/设为成员（`patchMember`）；无入口则产品标缺口另票 |
| 标签 | FE **S-R5**：标签表「编辑」改名/色（后端已有 `LabelHandler.Patch`；FE 缺则补或记缺口） |
| 冻结 / 并行 | Soft 三件仍冻；MOBILE 降速不丢、**勿改** `docs/qa-shots` 下 MOBILE 证据；UNBIND_LOOP / LOGIN_REFRESH **已关勿重开** |

## 步骤矩阵（骨架·等码 · 必须可自点）

| # | 覆盖 | 前置账号 | 路径（设置哪 Tab） | 操作步骤（自点） | 期望 | 证据截图路径 | 状态 |
| --- | --- | --- | --- | --- | --- | --- | --- |
| **SQ1-1** | §QA-1 团队改名 | Admin | 设置 · **团队**（`?tab=teams`） | ① 打开团队表，确认目标队操作列有 **「编辑」** ② 点「编辑」→ Modal 改 `name`（可±简介若有字段）→ 保存 ③ 看列表与 Toast | Toast 成功（人话）；列表该队名即时变；无白屏/按钮无响应；失败有人话 | `docs/qa-shots/settings-edit-team-rename.png`（Modal+保存后列表） | **✅ PASS** · 2026-10-25 09:07 Asia/Shanghai · Toast+列表名变 |
| **SQ1-2** | §QA-2 工作区改名 | Admin | 设置 · **工作区**（`?tab=workspace`） | ① 改「显示名称」为可辨新名（可选：Admin 改 slug）② 点 **保存** ③ 看顶栏/侧栏工作区名与本 Tab 表单回显 | Toast「工作区已更新」或等价成功；顶栏/侧栏名变；表单回显新值；非 Admin 应只读提示 | `docs/qa-shots/settings-edit-workspace-rename.png`（表单+顶栏/侧栏） | **❌ FAIL** · Toast+表单 OK；**侧栏名未变**（AppLayout.ws 未刷新）→ 钉 React S-R2 |
| **SQ1-3** | §QA-3 项目编辑 | Admin（或有权改项目角色） | 设置 · **项目**（`?tab=projects`） | ① 目标项目行点 **「编辑」** ② `EditProjectModal` 改名（可±简介）→ 保存 ③ 看项目列表名 | Toast 成功；列表名变；Modal 关闭；无白屏 | `docs/qa-shots/settings-edit-project-rename.png`（Modal+列表） | **✅ PASS** · 兼证 **S-G2** ws PATCH |
| **SQ1-4** | §QA-4 成员角色 | Admin（≥2 人类成员为宜；勿拆唯一管理员） | 设置 · **成员**（`?tab=members`） | ① 非唯一 Admin 的人类成员行 → **更多** ② 「设为管理员」或「设为成员」（与当前角色相反）③ 看角色芯片与 Toast | Toast「角色已更新」；角色芯片变；唯一管理员降级/移除应拦截人话；**若交件后仍无入口** → 本行记「产品缺口另票」不硬 FAIL 功能码 | `docs/qa-shots/settings-edit-member-role.png`（更多菜单+角色变后） | **✅ PASS** · 「设为管理员」Toast 角色已更新 |
| **SQ1-5** | §QA-5 标签编辑 | Admin | 设置 · **标签**（`?tab=labels`） | ① 已有标签行点 **编辑**（或等价入口）② 改名称（可±色）→ 保存 ③ 看列表 | Toast 成功；列表名/色变。**若 Tab 在但无编辑入口** → 记缺口钉 React **S-R5**（后端 `PATCH …/labels/{id}` 骨架时已存在）；**若无标签页** → 记产品缺口 | `docs/qa-shots/settings-edit-label-rename.png`（编辑态+列表） | **✅ PASS** · **S-R5 覆盖** Toast+列表 |
| **SQ1-6** | §QA-6 回归·团队三按钮 | Admin（或可切队成员） | 设置 · **团队**（`?tab=teams`） | ① 非当前队点 **「设为当前」** → 确认切队/看板池跟队 ② **「成员」** → 跳 `?tab=members` ③ **「Bot 席」** → 进 `/settings/bots?team=…` | 三钮均可点、路径正确；**不**因加「编辑」挤掉/弄坏三钮；无回归白屏 | `docs/qa-shots/settings-edit-team-actions-regression.png`（操作列含编辑+三钮） | **✅ PASS** · 编辑+三钮均在且可点 |

## 证据产出（活点阶段）

开跑后：

1. 按上表自点；截图落入约定 `docs/qa-shots/settings-edit-*.png`（可另附 `SETTINGS_EDIT-S-Q1-qa-retest-YYYYMMDD.txt`）。  
2. 纪要必含：`time_AsiaShanghai` / `HEAD_at_run` / Go **S-G1** SHA / React **S-R\*** SHA。  
3. Overall = PASS | PASS† | FAIL（**有码跑绿前禁止写 PASS**）。  
4. 绿：commit 纪要+截图 → push origin/main → 短报产品体验开 **S-P1**。FAIL：钉 Go（S-G1）/ React（S-R\*）；不问用户。

## 现状笔记（2026-10-25 08:45 Asia/Shanghai · 骨架时探）

| 观察 | 结论 |
| --- | --- |
| HEAD（写清单时） | `b49286c` · 债文在 main；**无** SETTINGS_EDIT 业务 commit（S-G1 / S-R\* 均未入 origin） |
| Go 团队 PATCH | `mount_chi` / `team.go` 注册仅 Get/List/Create + 席/码/钉；**无** `PATCH …/teams/{teamID}` → 与债文根因一致，**等 S-G1** |
| FE 团队操作列 | `SettingsTeamsPanel` 仅「设为当前 / 成员 / Bot 席」；**无「编辑」** → **等 S-R1** |
| FE 工作区 | `SettingsPage` Tab `workspace` + `patchWorkspace` + Toast「工作区已更新」→ S-R2 回归底已有 |
| FE 项目 | `SettingsProjectsPanel`「编辑」+ `EditProjectModal`/`patchProject` → S-R3 回归底已有 |
| FE 成员角色 | 成员表「更多」→ 设为管理员/成员 + `patchMember` → S-R4 底已有（活点须点通） |
| FE 标签 | Tab `labels` 有「新建」+ 表（名称/颜色）；**无行内编辑**；后端 `LabelHandler.Patch` 已在 → **等 S-R5 补入口或产品记缺口** |
| 本地 API | `http://127.0.0.1:8080/healthz` **200**（可作活测底；本阶段不跑全表） |
| origin 分支 | 仅 `origin/main`；未见独立 S-G1/S-R\* 分支 tip |
| 本阶段动作 | **只交本清单骨架**；不改业务/MOBILE 证据；不宣称 PASS |


## 活点笔记（2026-10-25 09:13 Asia/Shanghai）

| 观察 | 结论 |
| --- | --- |
| FE/Go SHA | `0e15fa0` / `bd394ff` 均在 main 祖先；本跑 HEAD `23148da` |
| SQ1-1/3/4/5/6 | Playwright 真点 **PASS**（截图已落） |
| SQ1-2 | Toast+表单 OK；侧栏 `ws?.name` 仍旧 → **FAIL** · 钉 React：`onSaveWorkspace` 后刷新 AppLayout |
| S-G2 | API `PATCH …/workspaces/{ws}/projects/{id}` **200**；UI 项目改名 PASS |
| 标签 S-R5 | UI 编辑入口+保存 PASS |
| Overall | **FAIL** · 勿假 PASS；复测门禁=React 修侧栏刷新 |

## 明确不做（本刀）

- 不重做设置 IA；不另起移动设置页  
- 不删团队/毁档；不碰 Soft 三件；不扩 OAuth  
- 不验收 / 不重开 UNBIND_LOOP、LOGIN_REFRESH  
- 不改 MOBILE 箱截与债文证据  
- **骨架阶段不宣称 PASS**；本文件为 QA 文档骨架，允许单独 commit+push

## 状态

| ID | 状态 |
| --- | --- |
| S-Q1 checklist | **活点已跑** · 见 `SETTINGS_EDIT-S-Q1-qa-retest-20261025-0913.md` |
| S-Q1 活点 | **FAIL** · SQ1-2 侧栏名不刷新（钉 React S-R2）；其余 SQ1-1/3/4/5/6 PASS；S-G2+标签已覆盖 |
| Overall | **FAIL**（勿假 PASS；等 React 修后复测） |

## 下一步门禁（谁交件才能开跑）

1. **Go** 交 **S-G1**（`PATCH …/teams/{id}` 改名 + 校验 + 单测）并 push → SQ1-1 可 API/UI 联测。  
2. **React** 交 **S-R1**（团队「编辑」Modal）必选；**S-R2～S-R4** 回归点通；**S-R5** 标签编辑入口（或产品书面缺口）→ 全表 SQ1-1～6 可开跑。  
3. 码齐后 QA **立刻开跑**本清单 → 出 retest + `settings-edit-*.png` → 产品 **S-P1** 关单。  
4. MOBILE 并行不挡本刀；本刀优先于移动美化（债文派工）。
