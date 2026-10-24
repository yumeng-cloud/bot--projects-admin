# 显式补位码可绑停工席（SEAT_REFILL_BIND）

> 2026-10-25 00:54 Asia/Shanghai · 产品体验  
> **承接：** [`SEAT_REFILL`](./NEXT_WAVE_DEBT_SEAT_REFILL.md) / [`SEAT_LEAVE`](./NEXT_WAVE_DEBT_SEAT_LEAVE.md) 已关；HUMAN **V2** 仍禁**静默**自动补位。  
> **本刀：** 管理员签发的 **指定席** 补位 `ymt_`（`seat_slug` / seat 锚定）在席为 `stopped|left|open` 时，redeem **直接绑该席**，不再造 orphan 再 force。  
> **明示仍冻：** PRO_ADMIN **C2–D / E2 / G**。

---

## 目标态

1. 邀请码已锚定 `seat_slug`（或 seat id）且目标席 `stopped|left|open` 无活跃占用 → redeem 成功 **绑定该席**（`status=active`），返回 `seat_id`。
2. 未锚定席的 redeem：**仍**不静默补 `stopped|left`（V2 不变）。
3. `active` 占用 → 继续 **409** 人话禁顶替。
4. FE：补位 Toast/交接可点明「持码入场即占该席」（若文案仍写须再 force，改掉）。
5. 单测 + QA 箱证：停工→复制补位码→redeem→已绑；无 slug 通码仍不静默补。

## 非目标

- 不开放 active force 顶替。
- 不改扩编 `allow_expand_seat` 主路径。
- 不开 PRO_ADMIN 冻项。

## 勾选

| ID | 标准 | 谁 |
| --- | --- | --- |
| **B1** | `resolveBootstrapSeat`：码锚定席且 stopped\|left\|open → 可绑（非 skipBind）；未锚定仍 skip stopped\|left | Go |
| **B2** | 单测：锚定补位码绑停工席；通码不静默补；active 409 | Go |
| **B3** | FE 文案：补位码入场即占席（勿再暗示必须 force） | React |
| **B4** | QA：停工→补位码→redeem 已绑；V2 通码不静默 | QA |
| **B5** | 产品关单 | 产品体验 |

## 产品勾选

- [x] **B1** 锚定可绑 · Go `f7048a4` · 产品勾
- [x] **B2** 单测 · Go `f7048a4` · 产品勾
- [x] **B3** FE 文案 · FE `0c5f67b` · 产品勾
- [x] **B4** QA · PASS† `c2b210b` · 产品勾
- [x] **B5** 关单 · 2026-10-25

## 派工

| 角色 | 首刀 |
| --- | --- |
| **Go 后端** | B1/B2 |
| **React 前端** | B3（可与 Go 并行改文案） |
| **QA** | B4 |
| **产品体验** | 本裁定；B5 |

## 进度注记

- 2026-10-25 01:03 · **B4 PASS† → B5 关单** QA `c2b210b`：锚定停工 redeem 201 直绑；通码 V2；active 409；三测绿。soft† 非挡：R4 `seat-refill-after` Toast 旧于 B3（源码为准）。整包 **SEAT_REFILL_BIND 关**。下一刀：[`PRO_ADMIN_C2`](./NEXT_WAVE_DEBT_PRO_ADMIN_C2.md) ∥ [`SEAT_REFILL_SHOT`](./NEXT_WAVE_DEBT_SEAT_REFILL_SHOT.md)；**C3–D / E2 / G 仍冻**。
- 2026-10-25 00:58 · **产品勾 B1/B2** Go `f7048a4`：锚定停工可绑；通码 V2；active 409。三测绿。开 QA B4。
- 2026-10-25 00:57 · **产品勾 B3** FE `0c5f67b`：持码即占席、无 force 暗示。等 Go B1/B2 → QA B4。
- 2026-10-25 00:56 · **React 交 B3** FE `0c5f67b`：Toast/席卡/交接抽屉「持补位码入场即占该席」；勿暗示须 force；`tsc --noEmit` PASS。Go B1/B2 并行；等 B4。
- 2026-10-25 00:54 · SEAT_REFILL 关后产品拍：显式补位码可绑，收 orphan 迷路；V2 静默禁补不变。

## 关单进度

**整包关** · B1–B5 齐（Go `f7048a4` + FE `0c5f67b` + QA `c2b210b`）。
