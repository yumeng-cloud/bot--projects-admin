# 呼吸感密度 · 体验验收（P0）

> 状态：**B1–B5 全部已勾**（UI 视觉 + React 2026-10-24）— 呼吸感密度关单。
> 根因：U1 把侧栏压成 180px / 项高 32，像缩水后台。  
> 对标：Linear 精密但有呼吸感；Height 侧栏 ~240 + 宽主区。  
> 仓库：`/workspace/ai-collab-board`  
> 配色气质见 [`UI_CLEAN_P0.md`](./UI_CLEAN_P0.md)。与 [`UI_MODERN_P0.md`](./UI_MODERN_P0.md) 关系：**去老气**（细边框 / 淡品牌选中 / 禁蓝块粗边条 / 近无阴影 / U5 滚动）仍有效；**密度条款以本文为准**，废除「越窄越好」。

---

## 目标尺寸（CSS 变量）

| Token | 值 | 用途 |
|-------|-----|------|
| `--acb-sider-width` | **240px** | 侧栏展开宽（可接受 220–260；**废除 180**） |
| `--acb-sider-collapsed-width` | 64px | 侧栏折叠 |
| `--acb-content-padding-x` | 28px | 主区水平 gutter（≥20） |
| `--acb-content-padding-y` | 20px | 主区垂直 |
| `--acb-issue-main-padding-x` | 40px | Issue 抽屉主栏水平（gutter ≥24） |
| `--acb-issue-main-padding-y` | 32px | Issue 抽屉主栏垂直 |
| `--acb-props-rail-width` | **312px** | 属性轨（总指挥 300–320） |
| `--acb-issue-drawer-wide` | **940px** | Issue 宽抽屉（原 WIDE=820） |

节奏：8px。漂亮靠比例 / 留白 / 对比，不靠厚阴影或彩虹卡。

---

## 验收点（B1–B5）

| ID | 面 | 标准（可勾） | 反例 |
|----|----|--------------|------|
| **B1** | 侧栏 | 宽 **~220–240**（定 **240**）；项高 **36–40**（定 **36**）；字 13–14；图标 ≤16；项内 horizontal padding ~12–14。选中 = 淡品牌底 + 弱字重；**禁** Ant 蓝块 / 粗左边条。 | 写死 180 / 项高 32；蓝块回潮。 |
| **B2** | 主区 | gutter ≥20（`--acb-content-padding-x: 28`）；顶栏 ~48 保持。 | 主区贴边、双层工具条。 |
| **B3** | 包装条 | 宽屏不压扁：`.acb-pack-strip` 保持细条可读；进度轨 `flex-shrink: 0` + 固定宽；窄屏才折叠 secondary。 | 宽屏进度条被挤扁、chip 重叠难读。 |
| **B4** | 抽屉轨 | 属性轨 **~300–320**（定 **312**）；行高 ~36；label/value 呼吸；Issue 抽屉宽 **920–960**（定 **940**）；主栏 padding 用 issue-main 变量。 | 轨 280 挤、抽屉 820  titling 区憋。 |
| **B5** | 卡 / 轴 | 看板列 gap、卡片 padding 略松；仍禁厚阴影 / 彩虹卡；细边框保留。 | 浮卡阴影、过大圆角、强渐变。 |

---

## 勾选（产品体验）

- [x] **B1** 侧栏呼吸（240 / 36）
- [x] **B2** 主区 gutter（content-padding-x 28）
- [x] **B3** 包装条宽屏不压扁
- [x] **B4** 抽屉轨不挤（rail 312 / drawer 940）
- [x] **B5** 卡/轴内边距略松

## 必改文件

1. `docs/UI_BREATHING_P0.md`（本文）
2. `frontend/src/index.css` — `:root` 变量；sider 用 `var(--acb-sider-width)`；项高/字号/padding
3. `frontend/src/theme/linearTheme.ts` — Menu `itemHeight: 36`
4. `frontend/src/layouts/AppLayout.tsx` — `width={240}`
5. `frontend/src/components/issue/issue-detail.css` — rail / main padding 变量
6. `frontend/src/components/issue/IssueDrawer.tsx` — `ISSUE_DRAWER_WIDE = 940`
7. `frontend/src/components/board/board.css` — 列 gap、卡 padding 略松；pack-strip 不挤


---

## 勾选证据（2026-10-24 · Asia/Shanghai · 产品体验）

| ID | 结果 | 证据 |
|----|------|------|
| **B1** | 通过 | sider `--acb-sider-width: 240` + `AppLayout width={240}`；`linearTheme` `itemHeight: 36`；`index.css` 菜单项 height/line-height 36 |
| **B2** | 通过 | `--acb-content-padding-x/y: 28/20`；`.acb-board-page` / `.acb-page` 已用变量 |
| **B3** | 通过 | `.acb-pack-progress` / track `flex-shrink: 0`；≤720 隐 `.acb-pack-secondary` |
| **B4** | 通过 | `--acb-props-rail-width: 312`、`--acb-issue-drawer-wide: 940`、主栏 40×32；`IssueDrawer` WIDE=940 |
| **B5** | 通过 | `.acb-card` padding `11px 12px 12px 14px`；timeline bar `padding: 0 8px`；细边框/无厚阴影保留 |

