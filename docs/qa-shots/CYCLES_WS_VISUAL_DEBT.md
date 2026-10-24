# 工作区 / Cycles 苛刻视觉记债（2026-10-24 Asia/Shanghai）

对照 `docs/CN_VISUAL_QA_PLAYBOOK.md`：禁下半屏无功能白海、有序密度。证据：`80-cycles-debt.png` · `80-workspace-home.png` · `80-settings-debt.png`（同源 debt-retest；箱截硬刷限流未刷新，结构债仍成立）。

## Cycles — 交 React

| ID | 债 | 建议 |
|----|----|------|
| **C-V1** | 左「周期列表」仅 2 项时 `.acb-cycles-panel` 纵向拉满 → 下半屏空罐白海 | 短列表时勿空 stretch：底接容量/历史摘要，或内容高+底对齐次要块；禁空罐拉高 |
| **C-V2** | 右栏 Issue 行偏疏，首屏可显条数偏少 | 行高压到 ~36–40；Meta 标题/描述更贴；⋯ 保持 |
| **C-V3** | 页标题/侧栏仍英文「Cycles」，与设置/Bot 中文不一致 | 统一「周期」或产品拍板保留代号 |

## 工作区 Home — 交 React

| ID | 债 | 建议 |
|----|----|------|
| **W-V1** | 「我的工作区」1 行表浮在 `.acb-surface-panel` 拉满容器上 → 大片无功能白海 FAIL | 表下接最近项目/快捷；或 panel 随内容收缩并底栏接地，禁空罐拉高 |

## 设置（工作区内）— 交 React

| ID | 债 | 建议 |
|----|----|------|
| **ST-V1** | 成员表 3 行 + Tabs 内容区 ~70% 空海（同 Home） | 同 W-V1：内容区勿空 stretch；或 Tab 内补接入说明/空态有功能块 |

## CSS 侧（UI视觉可跟，不挡 React）

`page.css` 已做 panel 满高防「短面板悬空」；满高 + 少数据 = 空海。根因是**结构缺次要内容/短列表策略**，优先 React；CSS 仅能再压行距（C-V2）辅助。

## 结论

**记债 · 交 React**（C-V1/V2/V3 · W-V1 · ST-V1）。非放行。
