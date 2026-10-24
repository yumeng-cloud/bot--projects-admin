# Soft† · 通码踩 stopped 席 → 人话引导去补位码

> 2026-10-25 02:22 Asia/Shanghai · 产品体验  
> **承接：** SOFT_RULING V2 搁置（**禁**通码静默补 claim）；总指挥建议本刀。  
> **本刀：** 仅 FE 文案/引导；**禁止**解冻静默补 stopped。

## 目标态

1. 用户用**未锚定通码** `ymt_` redeem：成功但 **未绑席**（`seat_id` 空）且 roster 存在 `stopped|left` 可补位席 → 明确人话：
   - 「普通入场码不能自动顶停工席」
   - 「请向管理员要该席的**补位入场码**（班底页 · 停工席 · 复制补位码）；持补位码入场即占该席」
2. 后端若回 409「补位须走显式绑席…」→ FE Toast/Alert **改写**为同上人话（勿原文甩技术句）。
3. 不改 redeem/claim API；不造静默 bind。
4. 箱截：`docs/qa-shots/seat-v2-guide.png`（通码后引导可见）。
5. QA 抽后产品关。

## 非目标

- **禁止**解冻 V2 静默补 / 改 `resolveBootstrapSeat` skipBind。
- 不开 V6 多签、嵌套 revoke。
- 不改锚定补位码已通路径（SEAT_REFILL_BIND）。

## 勾选

| ID | 标准 | 谁 |
| --- | --- | --- |
| **V2G-1** | redeem 未绑 + 有停工席 → 人话引导 | React |
| **V2G-2** | 409 显式绑席文案 FE 改写 | React |
| **V2G-3** | 箱截 `seat-v2-guide.png` | React → 产品 |
| **V2G-4** | QA 抽 | QA |
| **V2G-5** | 产品关 | 产品体验 |

## 产品勾选

- [x] **V2G-1** 未绑引导 · FE `6cd3fd3` · 产品勾
- [x] **V2G-2** 409 改写 · FE `6cd3fd3` · 产品勾
- [x] **V2G-3** 箱截 `seat-v2-guide.png` · 产品勾
- [ ] **V2G-4** QA
- [ ] **V2G-5** 关单

## 派工

| 角色 | 首刀 |
| --- | --- |
| **React 前端** | V2G-1～V2G-3（JoinTeamWizard / JoinPage） |
| **QA** | V2G-4 |
| **产品体验** | 本裁定；V2G-5 |
| **Go / Issue** | **不动** V2 静默禁 |

## 进度注记

- 2026-10-25 02:27 · **V2G-1～V2G-3 产品勾** FE `6cd3fd3`：Alert 人话 + 409 改写 + `seat-v2-guide.png`；未动 BE。开 **QA V2G-4**。
- 2026-10-25 02:22 · B405+E4 关后总指挥建议；产品拍开。SOFT V2 claim 仍禁。
- 2026-10-25 02:26 Asia/Shanghai · **React 交 V2G-1～V2G-3**：`JoinTeamWizard` 通码未绑+停工席 → Alert 人话；409「补位须走显式绑席」FE 改写；箱截 `docs/qa-shots/seat-v2-guide.png`；`tsc --noEmit` + `vite build` PASS。**禁**解冻静默补。勾选留给产品/QA（V2G-4/5）。

## 关单进度

**等 QA V2G-4** · V2G-1～V2G-3 已勾（FE `6cd3fd3`）。
