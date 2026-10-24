# 下一小包 · T6-I · Issue 池按队（最小）

> 2026-10-24 19:04 Asia/Shanghai · 产品体验 · pins 整关后下一刀  
> **本包：Issue 按 `team_id` 隔离最小可验。不含 T7。不重开 pins。**

## 裁定（三选一）

| 候选 | 结论 | 理由 |
| --- | --- | --- |
| **Issue 池按队** | **开本刀** | T6 全文最后一块；pins/roster 已绿，池隔离才能「切队看自己的活」 |
| T7 模板向导 | **不开** | 仍是 P1；非本波收口刀 |
| 其它 UX 债 | 后置 | 先收完 T6 P0.5 |

## 一句话

同 WS 双队：带 `team_id`（或等价过滤）列 Issue 时，**队 A 的开放池不见队 B 的单**；错队读写 → **404**。现网 Issue 无队字段 → **本包必须迁库**（可空列 + 回填策略见下）。

## 含 / 不含

| 做 | 不做 |
| --- | --- |
| `issues.team_id` 可空 + 索引；迁移 + 回填口径（新建必填当前队；旧单 `NULL`=全 WS 可见或挂默认队，**证据写明**） | 大改看板拖拽/排期语义 |
| `ListIssues?team_id=`（或 path `/teams/{id}/issues`）只返本队（+ 约定 NULL 策略） | FE 全局团队切换器大改（可复用席位页当前 team 入参） |
| Create/Update 可写 `team_id`；跨队改 → 404 | T7；pins 返工；跨 WS |
| 箱证据双队 list 不串 | 历史单批量搬家 UI |

## 验收（I1–I4）

| ID | 标准 | 谁 |
| --- | --- | --- |
| **I1** | 迁移 + List/Create 带 team；单测双队不串 | **Go**（Issue 域或 Go 后端） |
| **I2** | 队 B list 不见队 A 的 `team_id=A` 单；错队 GET/PATCH → 404 | Go + **QA**/记忆箱验 |
| **I3** | 证据 `docs/qa-shots/T6I-issue-pool-YYYYMMDD.txt` | 记忆或 QA → **产品**勾 |
| **I4** | FE 最小：当前 team 上下文下列表换池（可挂看板/列表 filter；无则 API 关半单不挡） | **React** 并行 |

## 派工

1. **Go / Issue 域后端**：迁库 + API + 单测（主）  
2. **QA**：苛刻复测 I1–I3  
3. **React**：I4 并行，不挡 API 半单  
4. **种子**：API.md 一行  
5. **产品**：勾关；T6 全文可整勾（脚注 T7 另开）

## 成功

本文产品勾 + 测绿即 push。**T7 另开时再喊。**


## 文档回链

- REST 契约：[`API.md`](./API.md)「Issue · team_id（T6-I）」

## 产品勾选

- [ ] I1–I3 API/箱验 · 产品勾
- [ ] I4 FE（可后于 API）
