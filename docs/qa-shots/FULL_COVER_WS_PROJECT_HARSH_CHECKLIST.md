# 全覆盖 · 新工作区→项目→列表→顶栏→看板（QA 苛刻）

> 2026-10-24 22:15 Asia/Shanghai · QA  
> 总指挥令：新 WS→建项目→列表见→顶栏切换→再建模→进看板；**禁仅示例/seed 项目**。  
> 含 PRO_ADMIN **B1** 空态（主 CTA 本页创建）。等 FE 交 SHA（含有效箱截）立刻验；有红直怼 React / 对应同事。

## 红线（必须真人路径，禁只测「示例项目」）

| # | 步骤 | PASS |
|---|------|------|
| F1 | **新建**工作区（零项目；`AUTO_SEED=0` / 无自动塞示例） | |
| F2 | 进 `/w/{ws}/board`：空态「还没有项目」+ **主按钮「创建项目」**本页 Modal；「去设置」仅次要 | |
| F3 | Modal 填 name+key_prefix → `POST …/projects` 201 → **直进**该项目看板 URL | |
| F4 | 项目**列表/选择器**可见刚建项目（名称/前缀对得上；非仅「示例项目」） | |
| F5 | **顶栏**切换工作区（或项目选择器切到他项再切回）后，仍见该项目；池不串到别的 WS | |
| F6 | **再建模**第二个项目 → 列表/选择器出现两项；可进第二项看板 | |
| F7 | 从空态/列表进看板后可新建 Issue（轻抽：标题保存可见）；无 seed 技术口吻 | |

## 禁例
- 只打开 seed「育蒙科技 / 示例项目」当绿
- 主 CTA「去设置」踢迷宫
- 箱截纯白 / 无文案（B1 前车：`pro-b-empty-no-project.png`）

## 证据
- 主：`docs/qa-shots/FULL_COVER-ws-project-YYYYMMDD.txt`
- 箱截：`pro-b-empty-no-project.png`（空态 Modal 关）+ 可选 `full-cover-list.png` / `full-cover-board.png`
- API 摘要：新建 WS id、两项目 id、list 条数 0→1→2

## 与 B1 关系
B1 是 F2–F3 子集；B1 箱截无效则 **Overall 不关绿**（产品已暂挂 B1 @`e5fcca8`）。

## 关报
短报总指挥+产品：Overall · F1–F7 · 缺陷责任人 · 证据 · HEAD。测绿 push origin/main。
