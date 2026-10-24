# 成品套装 · P0 可执行清单（已裁定）

> 状态：**已裁定** — 用户要的是「成品套装」，不是散装半成品。  
> 范围：方案 + 缺口 → 优先级 → 指派；**本文件不授权改业务代码**（开工由总指挥令各域）。  
> 仓库：`/workspace/ai-collab-board` · 日期：2026-10-24（Asia/Shanghai）  
> 视觉契约：UI 视觉头脑风暴（列模板 + 进度包装）已并入 §1.1 / §1.2。

---

## 0. 成品定义与非目标

### 成品 = 一条闭环

用户打开工作区就能完成：

**看全局包装 → 看板推进 → 时间轴细拆/排期 → 开抽屉改详情/挂附件 → Bot 领叶子执行 → 回到包装看卡点与进度**

| 合格 | 半成品反例（禁止当交付） |
|------|--------------------------|
| 项目内一眼见进度 / 周期 / 协作人 / 卡点 | 只有 Issue 抽屉打磨、无包装层 |
| 看板列可套模板、可自定义，密度对齐 Linear | 硬编码五列、无模板入口 |
| Timeline 多层树 + start/due + Bot 只领叶子 | 只有 Cycle 列表冒充排期 |
| 抽屉 / 附件 / Cycle 与上述同导航、同数据 | 各页互不跳转、字段各说各话 |

### 非目标（P0）

- 独立「进度 Dashboard / Canvas / 报告页」（进度是环境信息，挂头栏）
- 甘特依赖箭头墙、自动排程、跨项目资源冲突
- 为模板另起粗边框卡片墙或全屏向导
- 替换 Wiki / 工时主流程（可链出，不重做）

### 对标

Linear：主内容留白 + Property 轨；克制灰阶；`--acb-*`、细边框、controlHeight ~28。

---

## 1. 四块能力清单

### 1.1 全局包装视图（项目进度 / 周期 / 协作人 / 卡点）

| | |
|--|--|
| **目标** | 进入项目（或看板头）即见「环境信息」：完成度、当前 Cycle、活跃协作人、阻塞卡点；不新开半成品大屏。 |
| **现状** | 有 Board / List / Cycles / Timeline 路由与导航；Cycle 页有 Progress；Issue 有 `blocks`/`blocked_by`；**无**项目级包装条带，卡点未聚合到头。 |
| **缺口** | ① 看板头/项目头「进度条带」组件 ② 当前 Cycle 摘要 chip ③ 协作人头像叠（assignee 去重）④ 卡点条：`blocked_by` 未解或过期 due 的叶子计数+跳转 ⑤ 禁止独立 Dashboard 页 |

**视觉契约（UI 视觉）**

- 进度挂**看板头 / Cycle 头**；只露一层主指标：细进度条（120–160px）或 `已完成/总数`（可切 story point）；填色品牌 `#5E6AD2`；完成极淡绿；忌彩虹环。
- Hover 再展开细分（优先级/负责人）；默认折叠。
- 列上维持计数徽章；Done 列不加重进度条。
- 禁止：同屏叠环+条+%大字；半成品报告页。

**P0 切片**

1. 复用「进度条带」挂 `BoardPage` 头（+ Cycle 页头对齐）。
2. 卡点：聚合 `blocked_by` 未完成 + （有 due 且 overdue）→ 头栏数字，点击进列表过滤。
3. 协作人：当前项目 Issue 上非空 assignee 去重头像（最多 5 +N）。
4. 当前 Cycle：若项目有 active Cycle，chip 链到 Cycles 页。

**验收（≤6）**

- [x] **P1** 看板头可见进度条带（FE `BoardPackStrip` + `useBoardPackStats`；Board/Cycles 同挂）（完成数/总数或 %），无独立进度页。
- [x] **P2** 有 active Cycle 时头栏可见并可跳转 Cycles。
- [x] **P3** 协作人叠层展示（≥2 人项目可辨）。
- [x] **P4** 卡点跳转列表 `?blocked=1`/`?overdue=1`（列表页已接过滤）。
- [x] **P5** 视觉：细条 + 品牌色，无彩虹环/粗卡片墙。
- [x] **P6** 窄屏不挤爆头栏（可折叠次要 chip；`board.css` ≤720 隐 `.acb-pack-secondary`）。

---

### 1.2 看板列：内置模板 + 可自定义

| | |
|--|--|
| **目标** | 列 = 工作流预设；空看板/看板头可套模板；支持自定义列名与显隐/顺序（状态枚举仍权威）。 |
| **现状** | `BOARD_COLUMNS` 硬编码 `backlog→todo→in_progress→in_review→done`；无模板、无自定义、无「恢复默认」。 |
| **缺口** | 模板模型；内置包；轻量入口；预览迷你竖条；应用确认（替换/合并）；持久化（项目级）。 |

**视觉契约（UI 视觉）**

- 入口轻：空看板 CTA 或看板头「⋯」；禁大卡片墙。
- 内置：次级「系统」标签、不可删（可恢复默认）；色点复用状态色，不新开色盘。
- 自定义：同密度编辑（细边框、行高 ~28–32）；勿全屏向导。
- 预览 = 迷你竖条列条；应用单次确认 + 一句风险文案。
- 应用后列头视觉不变（左缘色点/细顶边 + 计数）。
- 建议内置包：**默认五列（现网）**；**精简三列** Todo→进行中→Done；**缺陷分诊**；**Sprint 流**。自定义从内置 fork。

**P0 切片**

1. 产品定稿 3–4 个内置模板（列顺序 ⊆ 现有 `IssueStatus`，P0 **不新增状态枚举**）。
2. 项目级 `board_column_prefs`（或等价）：顺序、隐藏、显示名覆盖。
3. UI：模板选择器 + 自定义面板（抽屉或 Popover，非向导）。
4. 拖卡跨列仍走现有 status PATCH。

**验收（≤6）**

- [x] **C1** 空看板或「⋯」可打开模板选择（迷你竖条预览）。
- [x] **C2** 套用内置模板（已接 `boardColumnPrefs` API，localStorage 已撤）后列顺序/显隐立即生效且刷新保持。
- [x] **C3** 可「恢复默认」；系统模板不可删。
- [x] **C4** 自定义显示名/隐藏列；编辑态同密度。
- [x] **C5** 应用时替换/合并有确认 + 风险一句。
- [x] **C6** 不引入新状态枚举；拖拽改状态行为与现网一致。

---

### 1.3 时间轴细拆 + Bot 领叶子

> **权威方案**：[`SCHEDULING_BOARD_P0.md`](./SCHEDULING_BOARD_P0.md)（已裁定）。此处只写衔接，不重写 A1–A8。

| | |
|--|--|
| **目标** | Timeline：左树右条带；`start_at`/`due_at` = `YYYY-MM-DD`；Bot 只领叶子。 |
| **现状** | 路由/导航「排期」已有；Timeline 页占位；Go 增列/`is_leaf` 闸在代码侧；**Bot/MCP/CLI（S7）已落地**：`issue claim` / `board_issue_claim`、禁无父游离、claim 命名区分。 |
| **与包装/看板衔接** | 包装头「逾期」链 Timeline 或列表 due 过滤；看板卡可显示 due 点；抽屉 Property 轨读写 start/due；父节点进度可 rollup 到包装条。 |

**验收**：沿用 SCHEDULING §6 **A1–A8**（不重复列表）。

---

### 1.4 已有抽屉 / 附件 / Cycle · 如何一体

| 能力 | 现状 | 一体约定（P0） |
|------|------|----------------|
| **Issue 抽屉** | 主栏+Property 轨、ByteMD、附件 §C 已关 | 所有视图（看板/列表/Timeline/卡点跳转）**只开同一抽屉**；Property 轨含 start/due、parent、Cycle、附件入口 |
| **附件** | upload/list/content + FE 路径对齐 | 抽屉内完成；Bot API Key 同路径；不在包装页另做附件墙 |
| **Cycle** | 列表+容量+Progress | 包装头 chip → Cycle 页；Timeline **不**用 Cycle 日期替代 Issue 起止；看板可按 Cycle 过滤（若已有则保持） |
| **导航** | board / list / cycles / timeline / wiki / time | 项目级 Tab 保持；默认进 board（包装条在板上）；禁止再加第五个「总览」Tab |

**一体验收（≤6）**

- [x] **I1** 看板 / 列表 / Timeline 点击 Issue → 同一 IssueDrawer。
- [x] **I2** 抽屉 Property 可改 start/due；Timeline 拖改期同源字段。
- [x] **I3** 抽屉附件上传/预览/下载仍可用（§C）· 箱截 [`ATTACH_I3_BOX_CHECKLIST`](./ATTACH_I3_BOX_CHECKLIST.md) A1–A5 · `c749195`/`I3-*.png`（§C 全表仍等 QA）。
- [x] **I4** 包装头 Cycle chip ↔ Cycles（BoardPackStrip）。
- [x] **I5** 无独立 Dashboard；进度挂看板头。
- [x] **I6** Bot 领叶子后，看板/Timeline/抽屉 assignee 一致。（通道：`issue claim` 写共享字段 `assignee_id`；claim 响应 / GET / `list --assignee me` 本机一致；FE 三视图读同一字段即可。）

---

## 2. 一体信息架构

```
/w/:ws/projects/:pid/board     ← 默认落地；头栏 = 包装条（进度/Cycle/人/卡点）+ 列模板入口
/w/:ws/projects/:pid/issues    ← 列表；承接卡点过滤
/w/:ws/projects/:pid/cycles    ← 迭代容量；头栏同款进度条带
/w/:ws/projects/:pid/timeline  ← 排期树+条带（SCHEDULING P0）
         └─ 任意 Issue → IssueDrawer（描述/评论/附件/Property 轨）
```

**跳转**

| 从 | 到 |
|----|-----|
| 包装·卡点 | 列表 `?blocked=1` 或 `?overdue=1` |
| 包装·Cycle chip | Cycles 页当前 Cycle |
| 看板 / Timeline 卡片 | Drawer |
| Drawer · 在时间轴查看 | Timeline 并滚动定位（P1 可降级为仅打开 Timeline） |

**共享字段**：`status`、`assignee_id`、`cycle_id`、`parent_id`、`start_at`/`due_at`、`estimate`、relations(`blocks`)、attachments。

---

## 3. 缺口总表（优先级 → 指派）

| ID | 缺口 | 优先级 | 指派 | 依赖 |
|----|------|--------|------|------|
| S1 | 看板头进度条带 + Cycle chip + 协作人叠 | **P0 ✓** | React `useBoardPackStats` + Board/Cycles 挂 `BoardPackStrip` | 关 |
| S2 | 卡点聚合（blocked / overdue）+ 列表过滤 | **P0 ✓** | hook 聚合 + 列表 `?blocked`/`?overdue`；卡点 chip 跳转 | 关 |
| S3 | 列模板内置包 + 选择器 UI | **P0** | 看板 UI + React 前端；视觉契约 **UI 视觉** | 产品模板名单（本文 §1.2） |
| S4 | 列偏好持久化（顺序/显隐/显示名） | **P0 ✓** | Go API + 看板 UI 已切真 prefs | 关 |
| S5 | Timeline 页实装（树+条带+拖期） | **P0** | React 前端 | `start_at`/`due_at` API |
| S6 | `start_at`/`due_at` 增列+API + is_leaf + claim 409 | **P0** | Go 后端（schema）+ Issue 域后端 | Tech lead 已裁定 YYYY-MM-DD |
| S7 | Bot/MCP/CLI 只领叶子、禁游离创建 | **P0 · 通道已关** | Bot 记忆角色 | S6（服务端非叶子 409 / `is_leaf` 列表字段需运行中的 Issue 域二进制） |
| S8 | 抽屉 Property 轨接 start/due；包装/列/Timeline 跳转抽屉统一 | **P0** | React 前端 | S5/S6 |
| S9 | 验收勾选 + 缺陷回流 | **P0** | QA 测试 | S1–S8 可跑版本 |
| S10 | 包装 Hover 细分、分享图、跨项目包装 | P1 | React / UI 视觉 | S1 |
| S11 | 列模板「合并」策略精细化、自定义状态枚举 | P1 | Issue 域 + 产品体验 | S3/S4 |
| S12 | Wiki/工时链入包装（只读摘要） | P1 | 知识库工时 + React | 非阻塞 P0 |

---

## 4. 给总指挥的短报

```
【成品套装 P0 · 已裁定 · 可开工】
闭环：包装头（进度/Cycle/人/卡点）→ 看板（列模板）→ Timeline（细拆+Bot叶）→ 同一抽屉（附件/属性）→ 回包装。
禁止：独立进度 Dashboard；只抠抽屉；用 Cycle 冒充排期。
四块：①包装挂头栏 ②列模板内置+自定义（同密度）③排期见 SCHEDULING_BOARD_P0 ④抽屉/附件/Cycle 一体导航。
P0 指派：看板UI+React=包装头/列模板/Timeline/抽屉衔接｜Go+Issue域=日期+叶子闸+列偏好｜Bot记忆=只领叶｜UI视觉=token｜QA=勾选。
文档：docs/PRODUCT_SUITE_P0.md（本文）+ docs/SCHEDULING_BOARD_P0.md。
```

---

## 5. 开工催办话术（一行/域）

| 域 | 催办 |
|----|------|
| **看板 UI** | 读 PRODUCT_SUITE §1.1/§1.2：包装条挂看板头；列模板入口+迷你预览；与 React 分文件。 |
| **React 前端** | 包装头数据聚合、卡点过滤、Timeline 实装、抽屉接 start/due、列模板壳；统一 Drawer。 |
| **UI 视觉** | 进度条带 + 模板选择器 token/CSS（`--acb-*`）；不开新页。 |
| **Issue 域后端** | 续 S6；评估 S4 列偏好存储；列表支持 blocked/overdue 过滤若缺。 |
| **Go 后端** | 续 schema 增列；评审列偏好表/JSON；路由挂载复核。 |
| **Bot 记忆角色** | **S7 已关**：只领叶子 / 禁游离 / claim 命名；I6 通道侧 `assignee_id` 一致已验。 |
| **API 接线** | 列偏好 + start/due 类型进 client；错误 Toast。 |
| **QA 测试** | 按 §1 各块勾选 + SCHEDULING A1–A8；可跑后开测。 |
| **知识库工时** | P1：包装只读链 Wiki/工时，P0 不挡。 |

---


### S1/S2 FE 收口（2026-10-24）

> **状态：FE 已绿** — `useBoardPackStats` 聚合进度/Cycle/协作人/卡点（blocked∪overdue）；Board + Cycles 挂 `BoardPackStrip`；P6 ≤720 折叠次要 chip。列表过滤与卡点跳转已接。

### S4 裁定（2026-10-24 · **最终锁定 · 不再翻案**）

> **状态：API 已绿**（Go 测通；现网 :8080 已热更）。FE 请从 localStorage 切到 `boardColumnPrefs` client。

- **存储**：`projects.board_column_prefs` 可空 TEXT JSON。独立表 DDL **作废**（由 Go 清除 006 / CREATE TABLE）。
- **形状**：`{ "template_id"?: string, "columns": [{ "status", "hidden"?, "label"? }] }`；status ⊆ 现网枚举；label 空=FE 默认；null/DELETE=恢复默认五列。
- **分工**：**Go** schema 列 + `GET/PUT/DELETE …/projects/{id}/board-column-prefs`。**Issue 域不挂。**
- **FE**：`api/boardColumnPrefs.ts` 已接。
- P0 不做：新状态枚举、合并策略、工作区模板库。

## 附录 · 相关文档

- 排期权威：[`SCHEDULING_BOARD_P0.md`](./SCHEDULING_BOARD_P0.md)
- UX/Bot 验收：[`ACCEPTANCE_UX_BOT.md`](./ACCEPTANCE_UX_BOT.md)
- 去老气/滚动：[`UI_MODERN_P0.md`](./UI_MODERN_P0.md)
- 呼吸感密度（现行）：[`UI_BREATHING_P0.md`](./UI_BREATHING_P0.md)
- 清爽气质/配色：[`UI_CLEAN_P0.md`](./UI_CLEAN_P0.md)
- 总验收：[`ACCEPTANCE.md`](./ACCEPTANCE.md)
- 分工：[`../COORDINATION.md`](../COORDINATION.md)
