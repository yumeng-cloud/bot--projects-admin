# Wiki 体验 P0 · 头脑风暴 + 验收（拒「三字段垃圾弹窗」）

> 产品体验 · 2026-10-24  
> 触发：用户指「新建页面」Modal——标题 / **父页面 ID** / 正文 TextArea + 取消确定，「不是产品样子」。  
> 入口：`http://127.0.0.1:5174` → Wiki  
> 现实现：`frontend/src/pages/wiki/WikiPage.tsx`（Ant Modal+Form）  
> API：`docs/wiki-and-time.md`（已有 title / body_md / parent_id / project_id / issue_id）

---

## 0. 给总指挥

**原则**：Wiki 新建/编辑 = **文档主舞台**，不是后台 CRUD 表单。对标 Notion / Linear Docs / Confluence 精简版：进页即写、目录点选父级、Markdown 可预览、时间可读。

**本轮不做**：协同光标、实时多人、块编辑器换 TipTap（P0 仍 ByteMD/MD，但要像样）、权限矩阵重做。

**指派**：知识库工时（若 API/树缺口）｜React 前端（交互主责）｜UI 视觉（版式 token）｜产品体验勾选。

---

## 1. 头脑风暴：现在错在哪

| # | 反模式 | 用户感知 | 应对 |
|---|--------|----------|------|
| 1 | Modal 里填「父页面 **ID**」 | 开发者调试台，不是产品 | 父级 = 目录树选择 /「当前页子页」；永不露 UUID |
| 2 | 正文塞进弹窗 TextArea | 写文档像填工单 | **取消新建弹窗主路径**：点「新建」→ 立刻创建空页或进入全页编辑态 |
| 3 | 阅读态时间戳 RFC3339/`Z` | 生硬 | Asia/Shanghai 相对/绝对可读（如 `更新于 今天 12:15`） |
| 4 | 编辑=另一套 vertical Form | 与 Issue 详情已定的 Linear 主栏冲突 | 主栏：大标题 + ByteMD；侧栏可选关联项目/Issue |
| 5 | 目录「+ 新建页」无上下文 | 不知道建在哪 | 在选中节点下建子页；根级「新建」= 顶层 |
| 6 | 空态弱 | 不像知识库 | 空空间引导：模板（空白 / 会议纪要 / 产品约定）可选 |

**产品叙事（一句话）**：Wiki 是「可挂 Issue/项目的团队手册」，创建应像写文档，检索/目录像书，而不是再做一个表单中心。

---

## 2. 目标交互（定稿）

### 2.1 新建（W1–W3）

```
目录「+ 新建页」或顶栏「新建」
        │
        ▼
  立即创建草稿页（title=「无标题」，body 空，parent=当前选中或根）
        │
        ▼
  主区进入编辑态：大标题 inline + ByteMD（与 Issue 描述同级体验）
        │
        ▼
  blur / ⌘S / 自动保存 PATCH（防抖）
```

- **禁止**再出现「父页面 ID」输入框。  
- 可选：创建前 Popover 选模板（空白/会议纪要/产品约定），仍无 Modal 大表单。

### 2.2 阅读 / 编辑（W4–W5）

- 默认**阅读态**：渲染 Markdown（GFM）；顶栏「编辑」。  
- 编辑态：标题 + ByteMD；取消弹窗墙。  
- 侧栏或顶栏 Property 轨（紧凑）：父级（树选择）、关联项目、关联 Issue（搜索 YM-n，禁止手填 UUID）。

### 2.3 目录（W6）

- 树可展开；当前页高亮。  
- 右键/⋯：重命名、新建子页、删除、移到…（父级选择器）。  
- 搜索命中高亮（已有 search API）。

### 2.4 时间与空态（W7–W8）

- `updated_at` → 本地可读，不露裸 `Z`。  
- 无页面：插画/短文案 +「新建第一页」。

---

## 3. 验收清单

| ID | 标准 | 指派 | 状态 |
|----|------|------|------|
| **W1** | 去掉「新建页面」三字段 Modal 主路径；点新建即出可写页（或全页编辑态） | React | [x] |
| **W2** | 任何 UI **不出现**「父页面 ID」字样；父级用目录选择 / 「作为当前页子页面」 | React | [x] |
| **W3** | 正文编辑用 ByteMD（或同级 MD 编辑+预览），禁止弹窗大 TextArea 当主编辑器 | React + UI | [x] |
| **W4** | 阅读/编辑切换清晰；保存自动或 ⌘S，无「确定」关 Modal 才落库的主路径 | React | [x] |
| **W5** | 关联项目/Issue 用搜索选择（identifier），不选手填 UUID | React；（API 已有字段） | [x] |
| **W6** | 目录：在选中节点下建子页；⋯ 至少新建子页/删除 | React | [x] |
| **W7** | 时间展示 Asia/Shanghai 可读，无裸 RFC3339Z | React + UI | [x] |
| **W8** | 空空间/空页引导像样（短文案+主按钮），非仅 Empty | UI + React | [x] |

---

## 4. 非本轮

- 实时协同、评论@、全文权限 ACL  
- 换 TipTap 块编辑器（可开 P1）  
- 向量搜索 UI 大改（API 已有 hybrid，P0 搜索框能用即可）

---

## 5. 与排期/Issue 的关系

- Issue 抽屉 R2、排期 S1–S8 **并行不挡**；Wiki 本单独立。  
- 编辑器气质对齐 Issue 描述（ByteMD + 空态克制），避免两套文档感。

---

## 6. 勾选

只认 5174 Wiki 实机/截图。拒「Modal 里换个 label」假改。


---

## 7. API 就绪（2026-10-24 · 知识库工时）

- Create：空 title →「无标题」；body 可空；parent 省略 = 根；同 space 连续草稿唯一 slug（不 409）
- PATCH：部分字段可 autosave；空 title 亦落「无标题」
- Tree / Search 够用
- 路径：`backend/internal/services/wiki/wiki.go`、`docs/wiki-and-time.md`
- **结论**：React 可直接点新建即草稿，无需再等 API


---

## 8. 落地进度（进行中 · 未勾）

| 项 | 状态 | 备注 |
|----|------|------|
| API | ✅ | 知识库 §7 |
| W3/W7/W8 气质 CSS | 挂上未勾 | `wiki.css` + hooks；等 React 去 Modal + 5174 实机 |
| W1/W2/W4–W6 | React 改写中 | 去新建 Modal、点新建即草稿、ByteMD 主区 |
| 勾选 | 等 | 只认 5174 硬刷截图；拒「只换 class」假关 |


---

## 9. W1 实机勾选（2026-10-24 12:21 · 5174）

- [x] **W1** 点「新建页」无三字段 Modal；草稿「无标题」进主区 — `wiki-w1-draft-edit.png`
- [x] **W2** 画面无「父页面 ID」
- [x] **W4** 「编辑中 · ⌘S 保存」+ 完成；主区可写
- [x] **W7** `更新于 今天 12:20`（Asia/Shanghai 可读）
- [x] **W8** 空主区文案 +「新建第一页」CTA — `wiki-w1-main.png`
- [x] **W3** ByteMD 主编辑器：空态占位已见；聚焦工具条再补一眼（不挡演示）
- [x] **W5** 关联项目/Issue 搜索
- [x] **W6** 目录 ⋯ 新建子页/删除


---

## 10. W3/W6 补图勾选（2026-10-24 12:23）

- [x] **W3** ByteMD 工具条（编辑/预览 + 图标）— `wiki-w3-bytemd-toolbar.png`
- [x] **W6** 目录⋯：新建子页 / 删除 — `wiki-w6-tree-menu.png`
- [ ] **W5** 代码已有 `acb-wiki-props` 可搜索 Select（identifier 标签）；等 5174 截图再勾


---

## 11. W5 关单 · Wiki UX P0 整单收（2026-10-24 12:24）

- [x] **W5** 关联项目/Issue 可搜索 Select；Issue 选项含 identifier（SMK-1、YM-9…），无手填 UUID — `wiki-w5-props.png` / `wiki-w5-issue-search.png`
- **W1–W8 全勾**。垃圾三字段 Modal 主路径已废。演示入口：`http://127.0.0.1:5174` → Wiki
