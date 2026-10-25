# SEAT_PANEL_UX P-Q1 · 席位实态 + 入场码展示验收清单（活点·PASS†）

> 2026-10-25 08:56 Asia/Shanghai · QA 活点 Soft†  
> **债：** [`docs/NEXT_WAVE_DEBT_SEAT_PANEL_UX.md`](../NEXT_WAVE_DEBT_SEAT_PANEL_UX.md)  
> **FE：** `aec2966` P-R1～R3 · **P-V1** 总监放行 @`53c85de` · **P-M1** @`c433509` · 通道 `a665392`  
> **纪要：** [`SEAT_PANEL_UX-P-Q1-qa-retest-20261025.md`](./SEAT_PANEL_UX-P-Q1-qa-retest-20261025.md) @`6e66ac7`  
> **硬约束：** 禁甩用户；不问用户；未改 SETTINGS_EDIT / MOBILE；Soft 仍冻；M7 勿重开。  
> **三态定稿（产品放行 · 本清单验收口径）：** **在席** / **空闲（从未占用）** / **停工可补**（骨架「占用中/可补位/未进席」已否决对齐）。

## 目标

Bot 席位页：人话三态对齐本队 roster；禁把 WS 有 Bot / Key / 口头加入画成在席。有数据出诚实条。入场码默认折叠摘要 + 按席分组 + 同组≤2 末四 + ymt_/ymw_ 分块 + 吊销二级；无码整卡隐藏。

## 债文 QA / 勾选摘要

| ID | 标准 | 谁 | 本清单覆盖 |
| --- | --- | --- | --- |
| **P-M1** | 实态纪要 | Bot 记忆 | 前置 · @`c433509` |
| **P-G1** | 免刀（bots − roster） | Go | PQ1-4 |
| **P-R1** | 席卡三态：在席 / 空闲（从未占用）/ 停工可补 | React | **PQ1-1～3** |
| **P-R2** | 未占席诚实条 | React | **PQ1-4** |
| **P-R3** | 折叠+按席+≤2+ymt/ymw+吊销二级 | React | **PQ1-5～8** |
| **P-V1** | 视觉放行 | 审图→总监 | **PQ1-9** |
| **P-Q1** | 箱自测 | QA | **本文件 · PASS†** |
| **P-P1** | 产品关 | 产品体验 | 待关 |

### 三态对照（产品定稿 · 活点硬对）

| 实态 | 页上人话（必须） | 判定 | 禁 |
| --- | --- | --- | --- |
| 已绑活跃 | **在席 · {Bot 显示名}** | `active` / 有 `bound_member_id`\|claim | 把 WS 成员当占席 |
| 从未占用 | **空闲（从未占用）** | `empty`/`open`/无绑 | 含糊单字「空闲」当主标；写成「停工可补」 |
| 曾绑已清 | **停工可补** | `stopped`\|`left`\|`departed` | 写成「空闲（从未占用）」或含糊「可补位」当空态 |

### 入场码拍死

| 拍死 | 不做 |
| --- | --- |
| **默认折叠摘要**：`已签发 N · 有效 M`（或「暂无有效」） | 一进来整墙末四 |
| **按席/role_hint 分组**；无 hint →「未指定席」 | 平铺无结构 |
| 同组默认可展 **≤2** 末四，其余「还有 k 条」 | 同组整墙 |
| **ymt_ / ymw_ 分块** | 混一列 |
| **默认只展开有效**；已吊销进二级「已吊销」折叠 | 有效与吊销混排刷屏 |
| 行内：末四 + 状态 Tag + 吊销（有效行） | 展示明文 `ymt_` |
| 无任何码 → **整卡隐藏** | 空卡占版 |

## 前置（交件实填）

| 项 | 值 |
| --- | --- |
| FE | `aec2966` 已入 main |
| 箱截 Soft† | `SEAT_PANEL-P-V1-states.png` / `…-invites.png` |
| Admin | 未本刀 live 登录（源码+箱截 Soft†） |
| 冻结 / 并行 | Soft 仍冻；S-Q1 文件未碰；M7 勿重开 |

## 步骤矩阵（活点 Soft†）

| # | 覆盖 | 期望 | 证据 | 状态 |
| --- | --- | --- | --- | --- |
| **PQ1-1** | 三态 · 在席 | Tag「在席 · {名}」；只认本队 roster | `SEAT_PANEL-P-V1-states.png` + `seatOccupancyTagLabel` | **PASS†** |
| **PQ1-2** | 三态 · 停工可补 | Tag「停工可补」；≠空闲 | 同上 | **PASS†** |
| **PQ1-3** | 三态 · 空闲（从未占用） | Tag「空闲（从未占用）」 | 同上 | **PASS†** |
| **PQ1-4** | WS≠占席 · 诚实条 | 有数据显「本工作区有 N 个 Bot 尚未占本队席」+引导；禁画成在席；无则不显 | states 帧诚实条 + `unseatedWsBots` | **PASS†** |
| **PQ1-5** | 折叠码 · 默认摘要 | 默认折叠；`已签发 N · 有效 M` | `SEAT_PANEL-P-V1-invites.png` 左帧 + Collapse 无 defaultActiveKey | **PASS†** |
| **PQ1-6** | 按席分组 · ≤2 · ymt/ymw | 分组+未指定席；PREVIEW=2；分块；仅末四 | invites 右帧 + `IssuedInvitesPanel` | **PASS†** |
| **PQ1-7** | 吊销二级藏 | 嵌套「已吊销 N」；不混排 | invites 右帧 | **PASS†** |
| **PQ1-8** | 无码藏卡 | `total===0 → null` | 源码硬对 | **PASS†** |
| **PQ1-9** | P-V1 视觉 | 总监已放行；密度跟 bots.css | P-V1 双帧 @`53c85de` | **PASS†** |
| **PQ1-10** | 禁混 · 口头≠占席 | 只认本队 roster；与 PQ1-4 交叉 | P-M1 + 诚实条/席卡 | **PASS†** |

## 证据产出

1. Soft†：复用 FE 箱截 `SEAT_PANEL-P-V1-*.png`（未另截 `seat-panel-ux-*.png`）。  
2. 纪要：`SEAT_PANEL_UX-P-Q1-qa-retest-20261025.md` @`6e66ac7`。  
3. Overall = **PASS†**。  
4. 绿 → commit 纪要+清单+债勾 → push origin/main → 产品开 **P-P1**。

## 现状笔记（2026-10-25 08:56 · 活点）

| 观察 | 结论 |
| --- | --- |
| FE | `aec2966` 三态+诚实条+IssuedInvitesPanel 齐 |
| 箱截 | P-V1 双帧 Soft† 对齐定稿文案与折叠/分组 |
| 骨架旧文案 | 已按产品定稿改「在席 / 空闲（从未占用）/ 停工可补」 |
| Overall | **PASS†** |

## 明确不做

- 不解冻 Soft† 通码静默补 stopped  
- 不重做入场兑换主航线；不改 ymt_/ymk_ 分工  
- 不另起移动席页；不重开 M7  
- 不改 `SETTINGS_EDIT-S-Q1-checklist.md`  
- 本刀未 live 浏览器全点（Soft† 源码+已有箱截）

## 状态

| ID | 状态 |
| --- | --- |
| P-Q1 checklist | **活点完成** |
| P-Q1 活点 | **PASS†** |
| Overall | **PASS†** → 产品 **P-P1** |
