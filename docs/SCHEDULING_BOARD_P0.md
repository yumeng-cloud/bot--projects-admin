# 全局排期看板 · P0 方案（已裁定）

> 状态：**已裁定** — 形态锁定 **Roadmap / Timeline（时间轴）**。  
> 范围：只出方案；**本文件不授权改业务代码 / 实现 UI**（开工须总指挥另令）。  
> 仓库：`/workspace/ai-collab-board`  
> 日期：2026-10-24（Asia/Shanghai）

---

## 0. 现状只读结论（方案依据）

| 对象 | 已有 | 缺失 / 缺口 |
|------|------|-------------|
| **Issue** | `estimate`、`cycle_id`、**`parent_id`**（列已存在；PATCH/创建已通；`checkParentIDCycle` 防环） | **无 `start_at` / `due_at`**；前端时间轴页不存在 |
| **Cycle** | `start_date` / `end_date`、`capacity`、Issue 划入；Cycle 页 = 列表 + 容量 | **不是**全局时间轴；仅迭代归属，不能层层拆细看排期 |
| **Relations** | `blocks` / `blocked_by` / `related` / `parent`（防环） | `relation_type=parent` 与列 `parent_id` **双轨**；P0 以列 `parent_id` 为父子树权威 |
| **Bot** | 角色 `claim/release`、handoff、todos | **无**「仅叶子 Issue 可领取」产品闸；易游离发挥 |
| **栈** | React + Ant Design 6 + dayjs + `@dnd-kit`；DB 存 UTC RFC3339，展示 Asia/Shanghai | — |

**推翻点**：此前 `BOT_MCP_CLI.md` 等写「Cycle 页即可、不要求 GitHub Projects 时间线」——用户已明确要**全局排期**，本方案以 Timeline 为 P0，Cycle 降为可选过滤/标注，不再当作排期主视图。

---

## 1. 目标与非目标（P0）

### 目标

1. **全局排期可视**：横轴为日期，Issue 以条带落在时间轴上（Linear / GitHub Projects Roadmap 式）。
2. **层层拆细**：大功能 → 小功能 → 细节操作全部挂在同一棵 **多层父子树**（`parent_id`）；父节点可折叠。
3. **起止可编**：Issue 有 `start_at` / `due_at`；P0 支持拖条改日期（或等价表单改期）。
4. **Bot 只领叶子**：无子节点的 Issue 才可被 AI/Bot claim 执行；父节点只聚合进度/区间，禁止 Bot 自行发明游离任务。

### 非目标（P0 不做）

- 完整 **甘特**：关键路径、依赖箭头墙、资源负荷图、基线对比。
- **日历**月/周格子主视图（日程向，非项目拆解向）。
- 自动从 `estimate` 推算起止、自动排程 / 关键路径算法。
- 替换或废除 Cycle（P0 并存；排期不以 Cycle 为唯一时间来源）。
- 移动端打磨、跨项目资源冲突检测、GitHub Projects 同步。

---

## 2. 三种形态对比

| 维度 | 甘特（Gantt） | **Roadmap / Timeline（裁定）** | 日历（Calendar） |
|------|---------------|-------------------------------|------------------|
| 信息密度 | 高：依赖线+多列属性易挤 | 中：左树 + 右条带，扫读友好 | 低～中：按日格，跨多日条弱 |
| 层层拆细可读性 | 中：深树+依赖易糊 | **高**：折叠子树是主交互 | 低：树与日格难同屏 |
| 起止拖拽 | 强（含依赖约束成本高） | **够用**：拖移/拉两端改 start·due | 拖到某日，跨度编辑弱 |
| 实现成本（Ant+现栈） | 高：需专用库或自研依赖图 | **中**：dayjs + 自研条带 / 轻量时间轴 | 中：可用 Ant Calendar，但产品错位 |
| 移动端 | 差 | 中（横向滚动可接受） | 相对好，但非本需求 |
| 与看板 / Cycle | 平行重系统 | **互补**：看板管状态，Cycle 管迭代，Timeline 管日历排期 | 偏个人日程，与看板弱相关 |

### 各列优缺点（≤3）

**甘特**

- ✅ 依赖与关键路径表达力最强  
- ❌ P0 实现与交互成本过高  
- ❌ 与「先拆细再排期」相比，过早引入依赖墙  

**Roadmap / Timeline（选用）**

- ✅ 横轴日期 + 父子折叠，直接满足「全局排期 + 层层拆细」  
- ✅ 与 Linear / GitHub Roadmap 心智一致，Ant+dayjs 可切片落地  
- ❌ 不做依赖箭头时，阻塞关系仍靠现有 `blocks`（详情/列表），轴上不画  

**日历**

- ✅ 适合「哪天有什么」的日程浏览  
- ❌ 多层拆解与跨周功能条可读性差  
- ❌ 不是项目排期主形态  

---

## 3. 推荐（已锁定）

**锁定：Roadmap / Timeline（时间轴）。**

**理由**：用户要的是「可层层拆细 + 起止时间 + 看全局排期」，不是依赖墙或日程格。Timeline 信息密度与拆树可读性平衡最好，P0 可用现栈切片；甘特重依赖、日历偏日程——**明确不选**。

- **不选甘特**：P0 不需要关键路径/依赖箭头墙，成本与复杂度过高。  
- **不选日历**：无法良好承载多层父子拆解与跨周功能条。

---

## 4. 数据模型（最小）

### 4.1 父子 Issue（树）

| 项 | P0 约定 |
|----|---------|
| 权威字段 | **`issues.parent_id`**（已有；`ON DELETE SET NULL`） |
| 层数 | **不硬限层数**（产品要求拆到细节操作）；实现上建议深度告警阈值（如 ≥6）防误操作，**不**在 P0 拦创建 |
| 与 `issue_relations` | `relation_type='parent'` **P0 不作为树权威**；新建父子只写 `parent_id`。旧 relation-parent 可后续迁移/只读兼容，不阻塞 Timeline |
| 防环 | 沿用现有 `checkParentIDCycle` |
| 叶子定义 | **不存在任何 `parent_id = 本 Issue.id` 的子 Issue** ⇒ **叶子**；否则为**父/聚合节点** |

### 4.2 起止时间

| 字段 | 类型建议 | 说明 |
|------|----------|------|
| `start_at` | 可空 **TEXT**，`YYYY-MM-DD` | Tech lead 已裁定（Go 2026-10-24）；P0 按**日**，不做日内小时 |
| `due_at` | 同 `start_at` | 二者皆有时须 `due_at >= start_at`（Issue 域校验） |
| 时区 | **库内日历日字符串**（非 UTC 时间戳）；**UI Asia/Shanghai** 展示/编辑该日 | 与 schema 增列落地一致；无历史回填 |
| 校验 | 仅一端有值：允许（条带可退化）；两端皆空：见下 |

### 4.3 与 Cycle

- Issue 仍可有 `cycle_id`；Timeline **可读** Cycle 名作条带色点/过滤，**不**用 Cycle 的 `start_date/end_date` 代替 Issue 自己的起止。  
- 无 Cycle 的 Issue 只要有日期仍上轴；有 Cycle 无日期 → 见兼容。

### 4.4 迁移 / 兼容（旧 Issue 无日期）

| 情况 | P0 行为 |
|------|---------|
| `start_at`、`due_at` 皆空 | **不上时间条**；仍出现在左侧树（或「未排期」分组），可一点设置日期 |
| 仅 `due_at` | 条带退化为 due 当日（或 due 向前 1 日宽，实现选定一种并写进 API 注释） |
| 仅 `start_at` | 条带退化为 start 当日 |
| 历史数据 | **不回填**；人工或后续脚本按 Cycle 窗口批量建议（非 P0） |

### 4.5 Bot 叶子领取（产品原则 + 最小字段/API）

**原则**

1. 项目开发须拆到**非常细**：大功能 / 小功能 / 细节操作全挂面板（多层父子树）。  
2. **AI/Bot 只从面板领取叶子任务执行**；禁止自行发挥、禁止漏挂功能对应的叶子。  
3. **禁止 Bot 创建游离任务**：Bot **不得** `POST` 无 `parent_id` 且无明确排期意图的「孤儿」Issue 充当自己的待办；新工作须由人类（或编排 Bot 在**已有父节点下**）拆出叶子后再领。  
4. 父节点：**只聚合**（子完成数、可选 rollup 区间）；**不可 claim 执行**。

**叶子判定（规范）**

```
is_leaf(issue) ⇔  COUNT(children where parent_id = issue.id) = 0
```

响应可增只读字段 `is_leaf: boolean` 与/或 `child_count: int`（列表/详情），避免 Bot 误判。

**P0 最小领取方式（不新表也可）**

| 方式 | 约定 |
|------|------|
| **最小** | Bot 对叶子执行 `PATCH /issues/{id}`：`assignee_id = 自己的 member_id` + 当前 `version`；服务端若 `!is_leaf` → **409** `conflict`（文案含「仅叶子任务可领取」） |
| **可选专用** | `POST /issues/{id}/claim`（或 Bot CLI `ymbot issue claim`）：内部同上校验；与角色 `bot-roles/.../claim` **分离**（角色 claim ≠ Issue claim） |
| 列表 | `GET …/issues?leaf_only=1&unassigned=1`（或 `assignee_id` 空）供 Bot 拉可领池 |

**与现有 Bot 域关系**：`bot_role_claims` 继续管「岗位」；**Issue 叶子 claim** 管「干哪条细活」。两者并存，文档与 MCP 工具须区分命名。

---

## 5. P0 范围切片

### 页面入口

- 工作区 / 项目导航新增 **「排期」/「Timeline」**（与 Board、Cycle 并列）。  
- 路由建议：`/w/:ws/projects/:pid/timeline`（最终以 FE 路由表为准）。

### 视图（最小）

1. **左**：父子树（indent + 折叠），列最小集：`identifier`、标题、状态、`start_at`、`due_at`、（可选）assignee、`is_leaf` 标记。  
2. **右**：横轴日期（日/周缩放 P0 择一，默认周），Issue 条带。  
3. **未排期**桶：无日期 Issue 可拖入轴或点选设期。

### 交互（最小集）

| 交互 | P0 |
|------|----|
| 折叠 / 展开子树 | ✅ |
| 拖条平移改 `start_at`+`due_at` | ✅（保持时长） |
| 拖条端点改一端日期 | ✅（至少一端） |
| 点击开 Issue 详情 / Drawer | ✅（复用现有） |
| 轴上画 `blocks` 箭头 | ❌ |
| 自动排程 | ❌ |
| 创建子 Issue（人类） | ✅ 可用详情/树内「添加子任务」；Bot 创建见 §4.5 |

### 后端切片（Issue 域）

- Schema：`issues.start_at`、`issues.due_at`（TEXT/date，可空）。  
- PATCH/创建暴露字段；乐观锁不变。  
- Claim 闸：非叶子 409；可选 `leaf_only` 查询。  
- **不**改看板状态机枚举。

### 前端切片（Issue/Board FE）

- 新 Timeline 页；Ant + dayjs；拖拽可 `@dnd-kit` 或指针事件（实现自定）。  
- 视觉服从现有 Linear 风（`BOARD_UI.md` / UI bar）。

---

## 6. P0 验收标准（可勾选，≤8）

- [x] **A1** 项目内可打开 Timeline 页：左树右轴，横轴为日期。（React TimelinePage build 过）  
- [x] **A2** Issue 支持多层 `parent_id` 拆解；树可折叠。（FE 树 walk by parent_id；≥3 层现网点验可随截图）  
- [x] **A3** Issue 可读写 `start_at` / `due_at`（API + 抽屉 Property 轨）；库内 `YYYY-MM-DD`；due≥start。  
- [x] **A4** 拖条改期 PATCH start/due；未排期桶 + 设期上轴。（React 已落）  
- [x] **A5** **叶子判定**：`is_leaf`/`child_count`；非叶子 assignee→409「仅叶子任务可领取」（Issue 域测绿）。  
- [x] **A6** Bot **仅能领取叶子**：通道 `issue claim` + 服务端非叶子 409（A5）已齐；面板「已领」待 FE。  
- [x] **A7** 产品闸：通道拒无父 create + 服务端 Bot 无父→403；人类路径不受影响。  
- [x] **A8** Cycle 并存；Timeline 用 Issue 自有日期；无日期进未排期不报错。

---

## 7. 指派（开工用）

| 角色 | 职责 | 路径 / 域 |
|------|------|-----------|
| **Issue/Board FE** | Timeline 时间轴页（树+条带+拖拽改期+未排期） | `frontend/src/pages/` 新排期页；复用 Issue Drawer |
| **Issue/Board Bot（Go）** | `start_at` / `due_at` 字段与 API；`parent_id` 树权威与深度体验；**叶子 claim 校验**；可选 `is_leaf` / `leaf_only` | `backend/internal/db` 增列（须同步 schema）、`services/issueboard`、`handlers/issue*` |
| **Bot Continuity / MCP·CLI** | 工具与文档改为「只领叶子」；禁止教 Bot 自创游离任务；与角色 claim 命名区分 | `docs/BOT*.md`、`cmd/ymbot`、MCP tools |
| **Tech lead** | schema 增列评审（冻结表改列须同步）；路由挂载复核 | `COORDINATION.md` 约定 |
| **QA** | 按 §6 勾选 | `docs/ACCEPTANCE.md` 可另增节（非本文件义务） |

---

## 8. 给总指挥的短报

```
【排期 P0 · 已裁定 · 可开工】
形态：Roadmap/Timeline（横轴日期 + 左树右条带）；不选甘特（依赖墙过重）、不选日历（难承载多层拆解）。
模型：沿用 parent_id 多层树；新增 start_at/due_at（按日 `YYYY-MM-DD` 字符串 / Asia/Shanghai 显）；无日期→未排期桶。
Bot：只领叶子（无子节点）；父节点只聚合；禁 Bot 自创游离任务；P0 可用 assignee=me + 非叶子 409。
验收：Timeline 页；≥3 层树；起止读写与拖改期；叶子 claim；Bot 禁游离；Cycle 并存。
指派：FE=时间轴页｜Go Issue 域=字段+叶子闸｜Bot/MCP=领取叶子工具与原则。
状态：方案已落盘 docs/SCHEDULING_BOARD_P0.md — 总指挥可催开工（本文件仍不含实现）。
```

---

## 附录 · 相关文档

- 分工：[`../COORDINATION.md`](../COORDINATION.md)  
- API 概览：[`API.md`](./API.md)  
- Bot：[`BOT.md`](./BOT.md)、[`BOT_MCP_CLI.md`](./BOT_MCP_CLI.md)（其中「不要求时间线」已被本裁定推翻）  
- Cycle UI：`frontend/src/pages/cycles/CyclesPage.tsx`（容量列表，非 Timeline）


续：划线便捷操作见 [`SCHEDULING_UX_P1.md`](./SCHEDULING_UX_P1.md)。
