# PRO_ADMIN · C2 列表多选批量（本波仅开此项）

> 2026-10-25 01:03 Asia/Shanghai · 产品体验  
> **承接：** [`SEAT_REFILL_BIND`](./NEXT_WAVE_DEBT_SEAT_REFILL_BIND.md) 已关；席链齐。  
> **本刀：** 仅 [`PRO_ADMIN`](./NEXT_WAVE_DEBT_PRO_ADMIN.md) **C2** — Issue 列表多选 + 批量改状态/归档（解档可选）；操作条不挡表头。  
> **仍冻：** **C3–D / E2 / G**（明示）。

## 目标态

1. Issue 列表可多选行；顶/底操作条：批量改状态、批量归档（仅 done，对齐既有 archive API）。
2. 操作条不挡表头；取消选择一键清空。
3. 无批量 API 则 Issue 域补；有则只 FE。
4. 箱截 `pro-c-bulk.png`；QA 抽苛刻后产品勾。

## 非目标

- 不开列配置 C3、超管筛 C4、成员台 D、看板归档 E2、G 整门。
- 不重做 archive 契约（沿用 ISSUE_ARCHIVE）。

## 勾选

| ID | 标准 | 谁 |
| --- | --- | --- |
| **C2-1** | 列表多选 + 操作条（改状态 / 归档） | React |
| **C2-2** | 批量 API 缺口补齐（若缺） | Issue 域 |
| **C2-3** | 箱截 `pro-c-bulk.png` | React → 产品 |
| **C2-4** | QA 抽 | QA |
| **C2-5** | 产品关（回写 PRO_ADMIN C2） | 产品体验 |

## 产品勾选

- [ ] **C2-1** 多选操作条
- [ ] **C2-2** API（若需）
- [ ] **C2-3** 箱截
- [ ] **C2-4** QA
- [ ] **C2-5** 关单

## 派工

| 角色 | 首刀 |
| --- | --- |
| **React 前端** | C2-1 / C2-3 |
| **Issue 域后端** | C2-2 若 FE 发现缺口 |
| **QA** | C2-4 |
| **产品体验** | 本裁定；C2-5 |

## 进度注记

- 2026-10-25 01:03 · SEAT_REFILL_BIND 关后产品拍开本刀；C3–D/E2/G 仍冻。

## 关单进度

**开干** · 等 React C2-1。
