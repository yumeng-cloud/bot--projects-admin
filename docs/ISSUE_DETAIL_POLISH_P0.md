# Issue 详情打磨 P0（对标 Linear · 拒假勾）

> 产品体验 · 2026-10-24  
> 触发：用户实机截图「Webhook 重试策略」抽屉——标题裁字、关联英文、空态 MD 工具栏吵。  
> 前置：`frontend/ISSUE_DETAIL_LAYOUT.md` 主栏+Property 轨骨架 / `ACCEPTANCE_UX_BOT.md` §A 已关；**本单只打磨观感与关联 DX，不重开布局大改。**  
> 仓库：`/workspace/ai-collab-board/frontend`

---

## 0. 给总指挥（短报）

**原则**：主栏继续大标题+描述留白；属性仍在右轨。本轮修「一眼土/破」的三处：标题不被裁、关联中文且不像表单、空描述不露 IDE 工具栏。

**验收标准（可勾）**：见下方 P1–P5；产品体验真机/截图复验后勾，拒「只改文档」。

**指派**：React 前端（标题裁切 / 关联文案+添加交互 / ByteMD 空态）｜UI 视觉（工具栏克制 token、轨内关联密度）｜看板 UI 若改抽屉壳则协同。

---

## 1. 截图缺陷 → 验收项

| ID | 缺陷（截图） | 标准 | 指派 | 状态 |
|----|--------------|------|------|------|
| **P1** | 大标题左侧首字被裁（`Webhook`→`ebhook`） | 标题全文可见；首字不被 drawer/overflow/负字距裁切；左右与主栏 padding 对齐，无贴边 | React 前端 | [x] §7 已关 |
| **P2** | 关联类型下拉为英文 `blocks` / `blocked_by` / `related` / `parent` | `RELATION_TYPE_LABELS`（及轨内展示）改为中文：**阻塞 / 被阻塞 / 相关 / 父任务**；API value 仍用英文枚举 | React 前端 | [x] §7 已关 |
| **P3** | 关联添加区常驻双 Select +「+」，打开即 Form 感 | 默认安静：无关联时一行「添加关联」；点开再选类型+搜 Issue；有关联时列表+轻量添加。禁止轨内常驻 Compact 表单墙 | React 前端 + UI 视觉 | [x] §7 已关 |
| **P4** | 空描述仍露出 ByteMD 整排图标工具栏（含碍眼图标），不像 Linear | **空且未聚焦**：只见「添加描述…」占位，**隐藏**左侧格式图标条；聚焦或已有正文后再显示工具栏。编辑\|预览 tab 可保留但克制 | React 前端 + UI 视觉 | [x] §7 已关 |
| **P5** | 整体仍偏「编辑器 demo」 | 空态主栏安静：标题清晰、描述占位、附件一行、评论输入；Property 轨行高/字号与 `UI_CLEAN` / `UI_BREATHING` token 一致，无新增粗边框 | UI 视觉抽检 + 产品体验勾 | [x] §7 已关 |

---

## 2. 非本单（勿扩 scope）

- 不重做主栏/轨双栏骨架（A1–A5 已关）
- 不换掉 ByteMD（除非 P4 无法用 CSS/空态隐藏解决，再另开单议 TipTap）
- 不做 GitHub 时间线、不做独立关联全页
- Webhook 业务本身（重试策略实现）不归本单

---

## 3. 实现提示（给前端，非产品体验改业务 tsx）

1. **P1**：查 `.acb-issue-title-input` 父级 `overflow`、drawer body 左 padding、`letter-spacing: -0.025em` 是否吃掉首字；必要时 title-wrap 加 `padding-inline-start: 2px` 或去掉负字距。
2. **P2**：改 `frontend/src/types/issue.ts` 的 `RELATION_TYPE_LABELS` 即可（`IssueRelationsBlock` 已读该表）。
3. **P3**：`IssueRelationsBlock.tsx` 默认折叠添加器；类型用中文 label；搜索 placeholder「搜索 Issue…」。
4. **P4**：`.acb-issue-md.is-empty:not(:focus-within) .bytemd-toolbar-left { display: none }`（或等价）；空态 toolbar 仅留 tab 或整条更淡。

---

## 4. 勾选规则

- 产品体验用**实机或新截图**勾；只改 label 字符串可代码 diff 辅助，但 P1/P3/P4 必须见 UI。
- 假勾打回；同事自报「已关」不代替勾选。

---

## 5. 回链

- 布局母规：`frontend/ISSUE_DETAIL_LAYOUT.md`
- 已关骨架：`docs/ACCEPTANCE_UX_BOT.md` §A
- 清爽/呼吸：`docs/UI_CLEAN_P0.md` · `docs/UI_BREATHING_P0.md`

---

## 6. 第二轮实机（2026-10-24 · YM-1「看板拖拽改状态」）

用户机截图仍见下列问题——**以用户所见为准**，box 上已改代码不算勾：

| ID | 缺陷 | 标准 | 指派 | 状态 |
|----|------|------|------|------|
| **P3′** | 关联区仍常驻 Compact：英文截断 `relat…` + Issue + 蓝「+」，并残留「暂无」 | 默认仅「添加关联」文字按钮；类型中文；无「暂无」+表单双显 | React | [x] §7 已关 |
| **P4′** | 空描述仍整排 ByteMD 图标（含碍眼图标） | 空且未聚焦：只见「添加描述…」，藏左侧格式条 | React + UI | [x] §7 已关 |
| **P6** | 每条评论常驻红垃圾桶 | 删除仅 hover/聚焦行可见；默认安静 | React + UI | [x] §7 已关 |
| **P7** | 活动区展开占半屏（5 条时间线） | 打开抽屉默认折叠；展开后行高克制 | React（确认 default）+ UI | [x] §7 已关 |
| **P8** | 用户机与仓库 diff 不一致（箱内已有中文 label / 安静添加，用户仍见旧 UI） | 用户打开的前端必须热更/重建到含 P1–P4 的构建；QA 在**用户同入口**复验 | React + QA | [x] §7 已关 |

**硬规则**：产品体验勾选只认用户机或同入口截图；同事自报「已改」无效。


---

## 7. 产品体验勾选（2026-10-24 · 入口 `http://127.0.0.1:5174`）

证据：QA Playwright + 截图 `qa-out/Webhook_重试策略.png`、`看板拖拽改状态.png`、`title_crop.png`；产品体验已目视同图。

| ID | 结果 | 说明 |
|----|------|------|
| **P1** | [x] 视觉通过 | 标题「Webhook…」首字完整；残留 `letter-spacing:-0.01em` → React 再去负字距防复现（不挡关单） |
| **P2** | [x] | 类型中文 阻塞/被阻塞/相关/父任务 |
| **P3 / P3′** | [x] | 默认仅「添加关联」；无常驻 Compact / relat… |
| **P4 / P4′** | [x] | 空态无 MD 工具栏，仅「添加描述…」 |
| **P5** | [x] | 5174 观感安静；未做 token 毫米量测 |
| **P6** | [x] | YM-1 三评论红删默认不可见 |
| **P7** | [x] | 「活动 (n)」默认折叠 |
| **P8** | [x] | 用户确认入口 5174；同机已验 |

**关单**：Issue 详情打磨 P0 **已关**（5174）。P1 残留字距收口交 React 小补丁，不重开大单。

### D7 文档对齐（2026-10-24 · Asia/Shanghai）

§1/§6 表曾滞后仍 `[ ]`，已与 §7 5174 关单对齐。**不重开大单**。
派 React 优先级：**低于** `NEXT_WAVE` D1–D5；仅顺手去掉标题残留 `letter-spacing:-0.01em`（防 P1 复现）。


续：样式第二轮见 [`ISSUE_DETAIL_POLISH_R2.md`](./ISSUE_DETAIL_POLISH_R2.md)。
