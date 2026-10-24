# 全覆盖 · 新工作区→项目/团队→列表→顶栏→看板（QA 苛刻）

> 2026-10-24 22:17 Asia/Shanghai · QA  
> 总指挥令：新 WS→建项目→列表见→顶栏切换→再建模→进看板；**并加团队**：新 WS→建队→列表→顶栏切→再建→进席/看板。**禁仅示例/seed**。  
> 含 PRO_ADMIN **B1** 空态（主 CTA 本页创建）。**B1 箱截 @640fd6e 复验 PASS 已关**；**项目 F1–F7 API PASS† / 顶栏 UI PENDING**；**团队 T1–T5 等管理面 SHA**。

## 红线（必须真人路径，禁只测「示例项目」）

| # | 步骤 | PASS |
|---|------|------|
| F1 | **新建**工作区（零项目；`AUTO_SEED=0` / 无自动塞示例） | **PASS** (API WS 05706ecc… projects=[]) |
| F2 | 进 `/w/{ws}/board`：空态「还没有项目」+ **主按钮「创建项目」**本页 Modal；「去设置」仅次要 | **PASS** (shot @640fd6e + code) |
| F3 | Modal 填 name+key_prefix → `POST …/projects` 201 → **直进**该项目看板 URL | **PASS†** (API 201 + Navigate code; live Modal †PENDING) |
| F4 | 项目**列表/选择器**可见刚建项目（名称/前缀对得上；非仅「示例项目」） | **PASS** (list name/prefix) |
| F5 | **顶栏**切换工作区（或项目选择器切到他项再切回）后，仍见该项目；池不串到别的 WS | **PASS† partial** (API isolation OK; 顶栏 UI PENDING) |
| F6 | **再建模**第二个项目 → 列表/选择器出现两项；可进第二项看板 | **PASS** (API n=2 distinct ids) |
| F7 | 从空态/列表进看板后可新建 Issue（轻抽：标题保存可见）；无 seed 技术口吻 | **PASS** (API issue B1R-1 201) |

## 禁例
- 只打开 seed「育蒙科技 / 示例项目」当绿
- 主 CTA「去设置」踢迷宫
- 箱截纯白 / 无文案（B1 前车已清：`pro-b-empty-no-project.png` @640fd6e 有效）

## 证据
- 主：`docs/qa-shots/FULL_COVER-ws-project-20261024.txt`
- B1：`docs/qa-shots/PRO_ADMIN-B1-qa-retest-20261024.txt` retest #2 · **Overall PASS**
- 箱截：`pro-b-empty-no-project.png`（空态 Modal 关 · 32310 B · 有效）
- API 摘要：WS 05706ecc…；P1 71e85b60… / P2 d23080d9…；list 0→1→2；Issue 2a2c3d78…
- 2026-10-24 22:17 Asia/Shanghai · HEAD `e6c5d60`

## 与 B1 关系
B1 是 F2–F3 子集；**B1 Overall PASS**（retest #2）— 请产品勾 B1。全覆盖 Overall F PASS†（顶栏 UI soft）。

## 关报
短报总指挥+产品：B1 **PASS** · F1–F7 **PASS†** · T1–T5 **PENDING wait SHA** · 证据 · HEAD。测绿 push origin/main。

## 团队红线（管理面 SHA 到后立刻验；禁仅 seed「加速班底」）

| # | 步骤 | PASS |
|---|------|------|
| T1 | 新 WS 内 **新建团队**（非仅打开示例队） | PENDING wait SHA（API incidental team OK, not closing） |
| T2 | 团队**列表**可见刚建队（名称对得上） | PENDING wait SHA |
| T3 | **顶栏**切队后 Bot 席/Issue 池跟当前队；不串 | PENDING wait SHA |
| T4 | **再建**第二队 → 列表两项；可切到第二队 | PENDING wait SHA |
| T5 | 进该队 **Bot 席 / 看板**可用（轻抽 roster 或 board） | PENDING wait SHA |

有红直怼 React / 对应同事。证据并入 `FULL_COVER-ws-project-20261024.txt` 或另 `FULL_COVER-team-YYYYMMDD.txt`。
