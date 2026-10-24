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

- [ ] **B1** 锚定可绑
- [ ] **B2** 单测
- [ ] **B3** FE 文案
- [ ] **B4** QA
- [ ] **B5** 关单

## 派工

| 角色 | 首刀 |
| --- | --- |
| **Go 后端** | B1/B2 |
| **React 前端** | B3（可与 Go 并行改文案） |
| **QA** | B4 |
| **产品体验** | 本裁定；B5 |

## 进度注记

- 2026-10-25 00:54 · SEAT_REFILL 关后产品拍：显式补位码可绑，收 orphan 迷路；V2 静默禁补不变。

## 关单进度

**开干** · 等 B1–B4。
