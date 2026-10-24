# 席位停工/补位 Admin 最小包（SEAT_LEAVE）

> 2026-10-25 00:50 Asia/Shanghai · 产品体验  
> **总指挥拍 P0**（人用任用 S3 / HUMAN F2·C·V2）：Admin 对活跃席一键标停工/离职 → 才允许新人补位；活跃继续 409；扩编=新席+新 `ymt_`。  
> **承接：** API 已有 `POST …/seats/{id}/stop` · `unbind`（[`API.md`](./API.md)）；CLI `seat leave`；HUMAN 冻 `b6614ff`。本包补 **Admin 可见一键 + 审计人话 + FE 空态/文案**，禁填表心智。  
> **PRO_ADMIN C–E/G 仍冻**（PRO_ADMIN **C1 先关**（本包排其后）；C2–E/G 仍冻）。

---

## 目标态（人话）

1. Admin 在 Bot 席页对 **占用中** 席：一键 **停工**（`stopped`）或 **离职解绑**（`left`），Popconfirm 中文；成功 Toast + roster 即时变「可补位」。
2. **活跃席** 继续硬 409 禁顶替（含 force）；仅 `stopped|left` 可 force/显式补位。
3. 新人补位：走空席/新席码，或停工席显式 bind（不静默 redeem 自动补 — HUMAN V2）。
4. 扩编：开新席 + 新 `ymt_`；禁止挤活跃。
5. 审计：stop/unbind/force 留人对人可见痕迹（activity 或 Issue comment，沿用 N3 口径）。

## 非目标

- 不重做 claim/roster；不改 HUMAN 自动身份主航线。
- 不做「确认顶替活跃」FE（否决）。
- 不并行 PRO_ADMIN C2–G；C1 归档显隐 **后置**。

## 勾选

| ID | 标准 | 谁 |
| --- | --- | --- |
| **L1** | Admin JWT：`stop` / `unbind` 对 active 席成功 → `stopped`/`left`；幂等/错态人话；单测 | Go |
| **L2** | stop/unbind/force 成功有审计（activity 或 comment）；缺则最小补 | Go |
| **L3** | 席卡/交接抽屉：占用中露出 **停工**、**离职解绑**（Popconfirm）；禁塞进高级 Form 栅格 | React |
| **L4** | `stopSeat` FE client + roster 乐观/刷新；态文案「停工席 · 可补位」与 HUMAN 一致；与 Bot/`ymt_` 入场文案不混 | React |
| **L5** | 箱证：active→停工→新人可补；active 仍 409；扩编新席不挤 | QA |
| **L6** | 箱截 `seat-leave-stop-*.png`（停工前/后） | React → 产品 |
| **L7** | 产品勾选关单 | 产品体验 |

## 产品勾选

- [ ] **L1** stop/unbind Admin 路径单测
- [ ] **L2** 审计留痕
- [ ] **L3** 席卡一键停工/离职
- [ ] **L4** FE client + 文案
- [ ] **L5** QA 箱证
- [ ] **L6** 箱截
- [ ] **L7** 产品关单

## 派工

| 角色 | 首刀 |
| --- | --- |
| **Go 后端** | L1/L2：核 `StopSeat`/`UnbindSeat` Admin 路径+审计；缺口补齐 |
| **React 前端** | L3/L4/L6：**C1 关后**再干本包 |
| **QA** | L5 |
| **产品体验** | 本裁定；L7 |

## 进度注记

- 2026-10-25 00:53 · **总指挥对齐**：本波先关 PRO_ADMIN **C1**；本包排 **C1 关后下一刀**。Go/React/QA 已预派可先核底，但勿抢 C1 主航线；产品验收顺序 C1 → SEAT_LEAVE。
- 2026-10-25 00:50 · 总指挥拍 SEAT_LEAVE 为下一 P0；产品落债并派 Go/React/QA。PRO_ADMIN C1 改后置。

## 关单进度

**排队（C1 关后）** · 债已落；等 PRO_ADMIN C1 关后再收 L1–L6。
