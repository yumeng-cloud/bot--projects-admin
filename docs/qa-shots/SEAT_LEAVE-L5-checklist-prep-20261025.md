# SEAT_LEAVE L5 · QA 箱验清单（备）

> 2026-10-25 00:55 Asia/Shanghai · QA 备  
> **债：** [`docs/NEXT_WAVE_DEBT_SEAT_LEAVE.md`](../NEXT_WAVE_DEBT_SEAT_LEAVE.md) @`71e9b4a`  
> **门禁：** 等 Go **L1/L2** + React **L3/L4**（+L6 截）推 origin/main 后再跑。  
> **不做：** PRO_ADMIN C–E（含 C1 归档显隐已后置）；不验收「确认顶替活跃」FE。

## 目标（L5 三硬条）

1. **active → Admin 停工 → 新人可补**  
2. **active 仍 409 禁顶替**（含 force）  
3. **扩编新席 + 新 ymt_ 不挤活跃**

## 前置

| 项 | 值 |
| --- | --- |
| API | Admin JWT；`POST …/seats/{id}/stop` · `unbind`；redeem ≥ HUMAN Q1 |
| FE | 席卡停工/离职 Popconfirm；态「停工席 · 可补位」 |
| 对照 | HUMAN Q3 活跃 409 · `29743c3`；Q5 补位仅 left/stopped |

## 步骤矩阵（交件后执行）

| # | 步骤 | 期望 | PASS |
| --- | --- | --- | --- |
| **L5-1** | 建队绑 active 席 BotA | roster `active` + holder | 基线 |
| **L5-2** | Admin `stop`（或 FE 停工） | → `stopped`；Toast/审计；文案可补位 | L1+L2+L4 |
| **L5-3** | 新人（无钥或新身份）持码/bind 补该席 | 成功入座；原 member left/stopped 规则符合债 | 可补 |
| **L5-4** | 另建 active 席；第二人 redeem/bind **同 slug**（含 force） | **409** 禁顶替；holder 不变 | 硬闸 |
| **L5-5** | 满员或 allow_expand：扩编新席 + 新 `ymt_` | roster 增席；既有 active holder 字节级不变 | 不挤 |
| **L5-6** | 顺手：FE 无「强行顶替」主 CTA；成员空态不混 ymt_ Bot 入场 | 文案隔离 | soft† |

## 证据

写 `docs/qa-shots/SEAT_LEAVE-L5-qa-retest-YYYYMMDD.txt`（+ 引用 React `seat-leave-stop-*.png`）。  
绿：push main；短报 **产品体验 + 总指挥**。FAIL：钉 Go（L1/L2）或 React（L3/L4）；反对/口径问题短票回产品。用户睡中不问。

## 状态

| ID | 状态 |
| --- | --- |
| L5 checklist | **PREP**（本文件） |
| L5 活点 | **WAITING** Go L1/L2 + React L3/L4 |
| PRO_ADMIN C–E | **SKIP**（本波勿验） |
