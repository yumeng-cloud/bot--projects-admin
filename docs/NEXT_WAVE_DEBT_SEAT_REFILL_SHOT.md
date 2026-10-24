# 补位 Toast 箱截对齐 B3（SEAT_REFILL_SHOT）

> 2026-10-25 01:04 Asia/Shanghai · 产品体验  
> **承接：** [`SEAT_REFILL_BIND`](./NEXT_WAVE_DEBT_SEAT_REFILL_BIND.md) 已关；QA B4 soft†：`seat-refill-after` Toast 旧于 B3 文案。  
> **本刀：** 重截补位成功 Toast / 席卡，对齐「持补位码入场即占该席」；源码已准，只收证据。  
> **并行：** 与 [`PRO_ADMIN_C2`](./NEXT_WAVE_DEBT_PRO_ADMIN_C2.md) **并行**，本刀 soft† **不挡** C2。  
> **仍冻：** PRO_ADMIN **C3–D / E2 / G**（C2 已开）。

## 目标态

1. `docs/qa-shots/seat-refill-after.png`（及必要 before）Toast/席卡文案与 B3 一致：持补位码入场即占该席；无 force 暗示。
2. 产品目视过；可选 QA 对读。

## 非目标

- 不改业务逻辑 / 不重开 BIND。
- 不开 PRO_ADMIN 冻项。

## 勾选

| ID | 标准 | 谁 |
| --- | --- | --- |
| **S1** | 箱内 5174 重截 seat-refill-after（文案对齐 B3） | React |
| **S2** | 产品目视勾 | 产品体验 |
| **S3** | （可选）QA 对读 | QA |

## 产品勾选

- [x] **S1** 重截 · FE `b2c9c9d`
- [x] **S2** 产品勾 · 目视 PASS
- [x] **S3** QA（可选）· **跳过不挡** · 产品关

## 派工

| 角色 | 首刀 |
| --- | --- |
| **React 前端** | S1 |
| **产品体验** | S2；关单 |
| **QA** | S3 可选 |

## 进度注记

- 2026-10-25 01:08 · **S2 目视 PASS → 关单** FE `b2c9c9d`：after Toast「持补位码入场即占该席；非扩编」；席卡脚注同文；无 force。S3 可选跳过。整包 **SEAT_REFILL_SHOT 关**。C2 主航线继续。
- 2026-10-25 01:08 · **React 交 S1** 重截 `seat-refill-{before,after}.png`：Toast/席卡「持补位码入场即占该席」对齐 B3（`0c5f67b`）；无 force 暗示；箱 Chrome headless。等产品 S2。∥ C2 不停。
- 2026-10-25 01:06 · 固化：本刀 soft† ∥ C2；软债其余搁置见 [`SOFT_RULING`](./NEXT_WAVE_DEBT_SOFT_RULING_20261025.md)。
- 2026-10-25 01:05 · **短票统一**：本刀 soft† 与 C2 **并行**；不挡主航线。
- 2026-10-25 01:04 · （已取代）曾独开本刀收 C2；总指挥反对后改并行。

## 关单进度

**开干** · React S1 已交；等产品 S2（∥ C2）。
