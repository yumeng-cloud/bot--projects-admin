# 停工席补位引导（SEAT_REFILL）最小包

> 2026-10-25 00:47 Asia/Shanghai · 产品体验  
> **承接：** [`SEAT_LEAVE`](./NEXT_WAVE_DEBT_SEAT_LEAVE.md) 已关；HUMAN V2 不静默补位仍冻。  
> **本刀：** Admin 对 **停工/可补位** 席的人话补位入口（复制入场码 / 指派），减少 redeem orphan 迷路。  
> **明示仍冻：** PRO_ADMIN **C2–D / E2 / G**；不改 HUMAN redeem 主航线（V2 orphan 硬禁另议）。

---

## 目标态

1. 席卡 `stopped|left`（可补位）：主路径清晰——**复制补位入场码**（该席 `ymt_`）+ 既有指派；短说明「新人持码入场；活跃席仍禁顶替」。
2. 与 Bot / `ymt_` 入场文案分家：补位语境写「补位」，不写「新开席扩编」。
3. 禁 Form 栅格；Popconfirm/短 Toast 即可。
4. 箱截 before（可补位卡）/ 复制或指派后态。

## 非目标

- 不禁 V2 orphan redeem（另刀）。
- 不改 active 409 / force 仅 stopped|left。
- 不开 PRO_ADMIN C2–D/E2/G。

## 勾选

| ID | 标准 | 谁 |
| --- | --- | --- |
| **R1** | 可补位席卡：补位入场码入口人话 + 与占用中停工按钮不打架 | React |
| **R2** | 文案分家：补位 ≠ 扩编新席；禁混 ymk_ | React |
| **R3** | 箱截 `seat-refill-*.png` | React → 产品 |
| **R4** | QA 抽：停工→见补位入口→码可用（或对读截） | QA |
| **R5** | 产品关单 | 产品体验 |

## 产品勾选

- [x] **R1** 可补位入口 · FE `d7e0396`
- [x] **R2** 文案分家 · FE `d7e0396`
- [x] **R3** 箱截 · FE `d7e0396` · `docs/qa-shots/seat-refill-{before,after}.png`
- [x] **R4** QA · PASS† `7e2c81a`
- [x] **R5** 关单 · 2026-10-25

## 派工

| 角色 | 首刀 |
| --- | --- |
| **React 前端** | R1–R3 |
| **QA** | R4（FE 交后） |
| **产品体验** | 本裁定；R5 |

## 进度注记

- 2026-10-25 00:54 · **整包已关**；下一刀 [`NEXT_WAVE_DEBT_SEAT_REFILL_BIND.md`](./NEXT_WAVE_DEBT_SEAT_REFILL_BIND.md)（显式补位码可绑停工席）。
- 2026-10-25 00:53 · **R4 PASS† → R5 关单** QA `7e2c81a`：可补位入口/文案/不打架/箱截对读 OK。整包 **SEAT_REFILL 关**。仍冻 PRO_ADMIN C2–D/E2/G。
- 2026-10-25 00:51 · **产品勾 R1–R3** FE `d7e0396`：可补位「复制补位入场码」+ Toast 末四；补位≠扩编；与停工/离职不打架。等 QA R4。
- 2026-10-25 00:51 · **React 交 R1–R3** FE `d7e0396`：可补位席卡「复制补位入场码」+ 指派；Toast/短说明写补位禁扩编/ymk_；占用中仍停工/离职；箱截 seat-refill-{before,after}.png；`tsc --noEmit` PASS。等 QA R4 → 产品 R5。PRO_ADMIN C2–D/E2/G 仍冻。
- 2026-10-25 00:47 · E3 正式关后产品拍本包为下一刀（PRO_ADMIN 可验项已尽且 C2–D/E2/G 仍冻）。

## 关单进度

**整包关** · R1–R5 齐（FE `d7e0396` + QA `7e2c81a`）。
