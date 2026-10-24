# 补位 Toast 箱截对齐 B3（SEAT_REFILL_SHOT）

> 2026-10-25 01:04 Asia/Shanghai · 产品体验  
> **承接：** [`SEAT_REFILL_BIND`](./NEXT_WAVE_DEBT_SEAT_REFILL_BIND.md) 已关；QA B4 soft†：`seat-refill-after` Toast 旧于 B3 文案。  
> **本刀：** 重截补位成功 Toast / 席卡，对齐「持补位码入场即占该席」；源码已准，只收证据。  
> **明示仍冻：** PRO_ADMIN **C2–D / E2 / G**。

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

- [ ] **S1** 重截
- [ ] **S2** 产品勾
- [ ] **S3** QA（可选）

## 派工

| 角色 | 首刀 |
| --- | --- |
| **React 前端** | S1 |
| **产品体验** | S2；关单 |
| **QA** | S3 可选 |

## 进度注记

- 2026-10-25 01:04 · BIND 关后总指挥重申仍冻 C2–D；产品收回解冻 C2，改拍本软†收口。

## 关单进度

**开干** · 等 React S1。
