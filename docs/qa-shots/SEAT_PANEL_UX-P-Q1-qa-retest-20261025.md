# SEAT_PANEL_UX P-Q1 · QA 箱验 retest

> time_AsiaShanghai: **2026-10-25 08:56 Asia/Shanghai**（box）  
> Overall: **PASS†**  
> 方式: 源码 Soft† @`aec2966` + 已有 P-V1 箱截 Soft†（禁甩用户；未抢 S-Q1 文件）

## 门禁指针

| 项 | 值 |
| --- | --- |
| HEAD_at_run | `53c85de`（含 P-V1 勾选；本刀纪要提交后另出 SHA） |
| FE P-R1～R3 | `aec2966` `feat(fe): SEAT_PANEL_UX P-R1~R3 席位三态+入场码折叠+未占本队席` |
| 通道对齐 | `a665392` roster 人话三态 + FormatUnboundBotsHint |
| P-M1 | `c433509` 记忆实态纪要（机制=WS Bot≠本队占用） |
| P-V1 | `53c85de` 总监放行双帧 |
| P-G1 | **免刀**（债文：`GET …/bots` − roster 占用集合） |
| 债文 | `docs/NEXT_WAVE_DEBT_SEAT_PANEL_UX.md` |
| 清单 | `docs/qa-shots/SEAT_PANEL_UX-P-Q1-checklist.md` |
| 箱截 Soft† | `docs/qa-shots/SEAT_PANEL-P-V1-states.png` · `…-invites.png`（FE 交件自带；本刀未另截 `seat-panel-ux-*.png`） |
| 冻结 | Soft 三件仍冻；**未改** `SETTINGS_EDIT-S-Q1-checklist.md` / MOBILE 证据；M7 勿重开 |

## 产品定稿三态（本刀验收口径）

| 实态 | 页上人话（必须） | 源码 |
| --- | --- | --- |
| 已绑活跃 | **在席 · {名}** / **在席** | `seatOccupancyTagLabel('active', name)` |
| 从未占用 | **空闲（从未占用）** | `…('empty')` |
| 曾绑已清 | **停工可补** | `…('left')` |

> 清单骨架曾写「占用中 / 可补位 / 未进席」——**产品否决**（见债文 P-V1 注）；本活点按定稿验收。通道「占用中」同义可并存，页上锁「在席」。

## 矩阵结果

| # | 覆盖 | 结果 | 证据 / 判据 |
| --- | --- | --- | --- |
| **PQ1-1** | 三态 · 在席 | **PASS†** | 源：`seatHandoffText.ts` → `在席 · ${name}`；箱截 states：React 前端 Tag「在席 · Grok Bot」+ 演员行 |
| **PQ1-2** | 三态 · 停工可补 | **PASS†** | 源：`kind==='left' → 停工可补`；箱截：产品体验 Tag「停工可补」+「曾绑已清 · 可持补位码入场」；≠空闲 |
| **PQ1-3** | 三态 · 空闲（从未占用） | **PASS†** | 源：默认 `空闲（从未占用）`；箱截：Go 后端 Tag +「从未占用 · 可指派或签发入场码」；禁含糊单字「空闲」当主标 |
| **PQ1-4** | WS≠占席 · 诚实条 | **PASS†** | 源：`unseatedWsBots = bots − (teamSeats active bound/claim)`；有则 Alert「本工作区有 N 个 Bot 尚未占本队席」+ 加入班底/空闲·停工可补引导；无则 `null`。箱截 states：N=2（产品体验、QA）旁路条，**未**画进在席卡。席卡只认 `seatOccForRole`←本队 roster |
| **PQ1-5** | 折叠码 · 默认摘要 | **PASS†** | `IssuedInvitesPanel` Collapse **无** `defaultActiveKey` → 默认折叠；摘要 `已签发 ${total} · 有效 ${activeCount}` / `暂无有效`。箱截 invites 左帧折叠 |
| **PQ1-6** | 按席分组 · ≤2 · ymt/ymw | **PASS†** | 按 `role_hint`/`seat_slug` 分组；无 hint→「未指定席」；`PREVIEW=2`+「还有 k 条」；**队席码 ymt_** / **工作区码 ymw_** 分块；行内仅末四+Tag+吊销，无明文。箱截右帧对齐 |
| **PQ1-7** | 吊销二级藏 | **PASS†** | 有效组外嵌套 Collapse「已吊销 N」；吊销行无吊销钮。箱截：`▶ 已吊销 2` |
| **PQ1-8** | 无码藏卡 | **PASS†** | `total===0 → return null`（整卡隐藏） |
| **PQ1-9** | P-V1 视觉 | **PASS†** | 总监已放行 @`53c85de`；双帧密度跟 bots.css；本刀抽看无末四墙/空卡死白 |
| **PQ1-10** | 禁混 · 口头≠占席 | **PASS†** | 与 PQ1-4 交叉：诚实条列名 ≠ 在席 Tag；P-M1 机制一致；席卡只读本队 roster |

## Soft† 说明

- **未**本刀 live 浏览器点路径（总指挥催并行；箱截已由 FE P-R\* 交件 + P-V1 总监放行）。  
- 源码路径硬对定稿文案与诚实集合差；邀请面板结构硬对折叠/分组/≤2/分块/二级吊销/无码藏。  
- 通道 `a665392`「占用中」为人话同义，**不**挡页上「在席」锁死。  
- Soft 三件（通码静默补 stopped）**未**解冻、未测。

## Overall

**PASS†** — 三态定稿 + 诚实条 + 入场码折叠/分组/ymt·ymw/吊销二级/无码藏卡齐绿。可开产品 **P-P1** 关单。

## 明确未做

- 未改 `SETTINGS_EDIT-S-Q1-checklist.md` / MOBILE 证据  
- 未提交他人 WIP；未动业务源码  
- 未另截 `seat-panel-ux-*.png`（复用 P-V1 双帧 Soft†）
