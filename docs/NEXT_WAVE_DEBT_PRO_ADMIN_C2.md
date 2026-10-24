# PRO_ADMIN · C2 列表多选批量

> 2026-10-25 01:05 Asia/Shanghai · 产品体验  
> **承接：** [`SEAT_REFILL_BIND`](./NEXT_WAVE_DEBT_SEAT_REFILL_BIND.md) 已关。  
> **本刀：** Issue 列表多选 + 批量改状态/归档（解档可选）；操作条不挡表头。  
> **并行：** [`SEAT_REFILL_SHOT`](./NEXT_WAVE_DEBT_SEAT_REFILL_SHOT.md) soft† 箱截，**不挡**本刀。  
> **仍冻：** **C3–D / E2 / G**。解冻条件：C2 关后产品另拍，或人用刚需反对短票统一。

## 目标态

1. Issue 列表可多选行；操作条：批量改状态、批量归档（仅 done，复用 `POST …/issues/archive`）。
2. 批量改状态：若无 batch API → Issue 域补；有则只 FE。
3. 操作条不挡表头；一键清空选择。
4. 箱截 `pro-c-bulk.png`；QA 苛刻后产品勾。

## 非目标

- 不开 C3 列配置 / C4 超管筛 / D 成员台 / E2 看板归档 / G 整门。
- 不重做 archive 契约。

## 勾选

| ID | 标准 | 谁 |
| --- | --- | --- |
| **C2-1** | 列表多选 + 操作条（改状态 / 归档） | React |
| **C2-2** | 批量改状态 API（若缺） | Issue 域 |
| **C2-3** | 箱截 `pro-c-bulk.png` | React → 产品 |
| **C2-4** | QA 抽 | QA |
| **C2-5** | 产品关（回写 PRO_ADMIN C2） | 产品体验 |

## 产品勾选

- [x] **C2-1** 多选操作条 · FE @4534e4c（串行 `patchIssue`；无 batch status API）
- [ ] **C2-2** API（若需）· **缺口**：无 `PATCH …/issues/batch-status`；FE 已串行 `patchIssue`，Issue 域可后补真批量
- [x] **C2-3** 箱截 · `docs/qa-shots/pro-c-bulk.png` @4534e4c
- [ ] **C2-4** QA
- [ ] **C2-5** 关单

## 派工

| 角色 | 首刀 |
| --- | --- |
| **React 前端** | C2-1 / C2-3（∥ SHOT S1） |
| **Issue 域后端** | C2-2 若 FE 要批量改状态 |
| **QA** | C2-4（FE 交后） |
| **产品体验** | 本裁定；C2-5 |

## 进度注记

- 2026-10-25 01:20 Asia/Shanghai · React 交 C2-1+C2-3：`IssueListPage` rowSelection + 操作条（改状态/归档 done-only/解档/清空）；`qaBulk=1` 箱截；archive 复用既有 API；改状态串行 PATCH（C2-2 留给 Issue 域）。
- 2026-10-25 01:06 · **一次拍死**：软债表见 [`SOFT_RULING`](./NEXT_WAVE_DEBT_SOFT_RULING_20261025.md)；React 必须 C2∥SHOT，禁只截图。
- 2026-10-25 01:05 · 短票统一：重开本刀；与 SHOT 并行。Issue 域：archive 已有；batch status 待喊。
- 2026-10-25 01:04 · 曾撤回；01:05 总指挥反对后恢复。

## 关单进度

**FE 已交 C2-1+C2-3** · 等 C2-2（可选）/ QA C2-4 / 产品 C2-5。未碰 seat-refill 截图。
