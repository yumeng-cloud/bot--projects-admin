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
- 不并行 PRO_ADMIN C2–G（C1 已正式关）。

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

- [x] **L1** stop/unbind Admin 路径单测 · Go `c5f5596` · 产品勾
- [x] **L2** 审计留痕 · Go `c5f5596` · 产品勾
- [x] **L3** 席卡一键停工/离职 · FE `4883426`
- [x] **L4** FE client + 文案 · FE `4883426`
- [ ] **L5** QA 箱证
- [x] **L6** 箱截 `docs/qa-shots/seat-leave-stop-{before,after}.png` · FE `4883426`
- [ ] **L7** 产品关单

## 派工

| 角色 | 首刀 |
| --- | --- |
| **Go 后端** | L1/L2：核 `StopSeat`/`UnbindSeat` Admin 路径+审计；缺口补齐 |
| **React 前端** | L3/L4/L6：席卡一键停工/离职 + client + 箱截 |
| **QA** | L5 |
| **产品体验** | 本裁定；L7 |

## 进度注记

- 2026-10-25 00:25 · **React 交 L3/L4/L6** FE `4883426`：对齐 Go `c5f5596` `POST …/seats/{id}/stop|unbind`；席卡+交接抽屉 Popconfirm「停工此席？/离职解绑此席？」；`stopSeat` client + roster 乐观「停工席 · 可补位」；箱截 before/after；`tsc --noEmit` PASS。等 QA L5 → 产品 L7。
- 2026-10-25 00:23 · **产品勾 L1/L2** Go `c5f5596`：单测绿；`seat.stop`/`seat.unbind` activity+comment。软债搁置：无独立 HTTP Admin JWT 集成测；stop 对 open/left 偏宽松（不挡最小包）。等 React L3/L4/L6。
- 2026-10-25 · **Go L1/L2 交件**：Admin `stop`→`stopped` / `unbind`→`left`；错态中文+幂等；成功写 `seat.stop`/`seat.unbind` 人话 activity + Issue comment（force 审计已有，沿用 N3）。
- 2026-10-25 00:19 · **开刀**：PRO_ADMIN C1 产品关 `8c0fe5a`；本包开干。派 Go L1/L2 · React L3/L4/L6 · QA L5。
- 2026-10-25 00:53 · **总指挥对齐**：本波先关 PRO_ADMIN **C1**；本包排 **C1 关后下一刀**。Go/React/QA 已预派可先核底，但勿抢 C1 主航线；产品验收顺序 C1 → SEAT_LEAVE。
- 2026-10-25 00:50 · 总指挥拍 SEAT_LEAVE 为下一 P0；产品落债并派 Go/React/QA。PRO_ADMIN C1 改后置。

## 关单进度

**开干** · L1–L4/L6 已勾；等 QA L5 → 产品 L7。
