# P0 · 工作区隔离 WS_ISOLATION（XWS-AUTH）

> 2026-10-25 10:18 Asia/Shanghai · 产品体验主持拍板  
> **主人已述硬约束：** 切工作区可见他仓 Bot 记忆/交接 → 越权；记忆/handoff/席位态必须按当前 workspace 隔离，越权一律拒。禁再当测试；测绿直推 main；不问主人选题。  
> **合议：** QA 探针 `@f169699` + Go / Bot记忆 / Issue / 产品默认拍板 — **无反对即通过**（提前关议）。

## 现象与探针

| 来源 | 结论 |
| --- | --- |
| 主人 UI | 切仓/切「07」仍见他仓记忆·交接·YM-* |
| 双仓 API `@f169699` | 错仓 memories/todos → **200 空**；handoff → **404**。明文跨仓泄漏 **未复现** |
| 根因合议 | ① memories/todos 缺 `member∈ws` 硬门（读 200 空）② **写 IDOR 已实锤** `@0c39d38`：错仓 POST memories/todos → **201** 脏行 ③ **FE 切仓不清 BotDetail/席态**（最贴 UI）④ 同仓跨项目 Issue（另验收行） |

## 气质一句

**当前工作区是唯一真相；错仓路径硬拒，切仓屏幕零残留。**

## 允许 / 禁止

| | 规则 |
| --- | --- |
| **禁止** | 跨 WS 读/写 Bot 记忆·Todo·handoff；错仓 `memberID` 打 path 仍 200 空；切仓后屏上残留他仓 bots/记忆/交接；业务读接口给平台超管旁路 |
| **允许** | 当前 path `workspace_id` 内、且 `member∈该 ws` 的读写；列表无数据时（校验已过）返回 `[]`；跨仓 issue 直链维持 **404**（现网）；超管入席走 `/admin/...` |
| **码表** | 调用方非本仓成员 → **403**；path `memberID`∉path ws → **404**（与 handoff 齐，禁 200 空）；超管 **无例外** |

## 验收矩阵（拆两行）

### A · 跨仓隔离

| # | 动作 | 期望 |
| --- | --- | --- |
| A1 | JWT 双仓成员：切 WS1→WS2，开 Bots/记忆/handoff | 屏上 **零** WS1 残留（空态或本仓数据） |
| A2 | `GET …/{WS2}/bots/{MID_ws1}/memories\|todos` | **404**（非 200 空） |
| A3 | 同上 handoff | **404**（回归） |
| A4 | 错配 mid **POST/PATCH** memories/todos | **404**（禁 201；回归 `@0c39d38`） |
| A5 | 非成员打业务读 | **403** |
| A6 | 写后同路径 GET | 不得因脏行变 200 有内容 |

### B · 跨项目可见性（同仓）

| # | 动作 | 期望 |
| --- | --- | --- |
| B1 | 同仓切到项目「07」看 handoff/Issue 列表 | 默认仅当前项目上下文 Issue；他项目不混进主视图（handoff 组装若带开放 Issue，须按当前 project 滤或人话标明范围） |
| B2 | 整仓 Issue 列表（无 project 上下文） | 可整仓（BE 不强制必带 `project_id`）；与 B1 分开 |

## 勾选

| ID | 标准 | 谁 |
| --- | --- | --- |
| **X-G1** | `assertMemberInWorkspace` → 挂 List/Create/Patch/Delete Memories & Todos（及 Upsert 入口）；与 Handoff 对齐 404 | Go |
| **X-G2** | 单测：错仓 mid 读/写 → 404；同仓 200 | Go |
| **X-G3** | MountBots/Teams/**Wiki/Time** wrap 改为 `authAny(wsMem(next))`（与 chi/测试 mux 齐）；`resolveMember` 有 mid 仍核 `members.workspace_id=path` | Go |
| **X-M1** | 通道旁路：文档/人话不挡；若 CLI/MCP 有错仓软吞一并跟 X-G1 | Bot 记忆 |
| **X-R1** | 切仓零残留：**首选** `AppLayout`/主内容树（含 `CurrentTeamProvider`）`key={wsId}` remount；并行 effect 发请求前 `setX([])/null`；点名清 `BotDetail`/`SeatHandoffDrawer`；顺手 `projectId` → `acb.project.<wsId>` | React |
| **X-R2** | `BotDetailPage`：切仓丢旧 `memberId` 或重导航；queryKey 含 `wsId` | React |
| **X-R3** | 项目上下文 listIssues/handoff 带 `project_id`（B1） | React |
| **X-I1** | 跨仓 Issue Get/List 契约单测 + 双仓 curl 自证（维持 404） | Issue 域 |
| **X-Q1** | 复跑双仓探针 A1–A5 + B1 箱截；纪要 | QA |
| **X-P1** | 产品关 | 产品体验 |

## 产品勾选

- [x] **X-G1** 成员∈ws 硬门
- [x] **X-G2** 单测
- [x] **X-G3** Mount/resolveMember
- [x] **X-M1** 通道
- [x] **X-R1** 切仓清态
- [x] **X-R2** BotDetail
- [x] **X-R3** 项目滤
- [x] **X-I1** Issue 契约
- [ ] **X-Q1** QA
- [ ] **X-P1** 关单

## 非目标

- Soft 三件仍冻；不重做设置 IA；不在本刀解冻 V2 静默补。
- 不强制所有 ListIssues 必带 `project_id`（整仓列表保留）。

## 派工

| 角色 | 首刀 |
| --- | --- |
| **Go 后端** | X-G1～G3（优先） |
| **React 前端** | X-R1～R3（与 Go 并行） |
| **Bot 记忆** | X-M1；跟 Go |
| **Issue 域** | X-I1 |
| **QA** | Go/FE 可跑后 X-Q1；禁甩主人 |
| **产品体验** | 本债；X-P1 |

## 进度注记

- 2026-10-25 10:23 · **勾 X-R1～R3** 同提交 `fae4e48`：key={wsId} remount + BotDetail/Drawer 先清 + acb.project.<wsId> + handoff Issues 按 project 滤。开 **X-Q1**。
- 2026-10-25 10:23 · **勾 X-G1～G3** Go `fae4e48`（与 X-R* 同提交）：assertMemberInWorkspace 写/读 404；MountWiki/Time/Bots/Teams `authAny(wsMem)`；resolveMember mid∈path ws；Bot Key 错仓 403。CLI/MCP 未动（X-M1）。
- 2026-10-25 10:21 · **勾 X-M1** 通道 `47f0c73`：禁软吞 403/404；人话隔离；证据 `WS_ISOLATION-XM1-channel-20261025.txt`。写拒闸仍等 X-G1。
- 2026-10-25 10:20 · **勾 X-I1** Issue `f7474d9`：跨仓 Get 404 + List 隔离契约单测；席报 live curl 齐（维持 404）。
- 2026-10-25 10:20 · Go 三点洞属实确认入 X-G3（含 MountWiki/Time）；写 IDOR 优先堵。
- 2026-10-25 10:19 · **补锤**：写路径 IDOR `@0c39d38`（POST 201）；FE 拍死 `key={wsId}` remount 为首选（React 合议）。
- 2026-10-25 10:18 · **拍板开刀**。探针 `f169699`；码表 403/404 齐 Go 建议。明文泄漏未复现；UI 主修 FE + memories/todos 硬门。
