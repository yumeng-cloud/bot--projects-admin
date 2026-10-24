# 下一小包 · T6-P · pins 最小（共享记忆墙）

> 2026-10-24 18:54 Asia/Shanghai · 产品体验 · 总指挥催下一刀  
> **本包：仅 `team_pins`。不含 Issue 池按队。T7 不做。**

## 裁定（三选一）

| 候选 | 结论 | 理由 |
| --- | --- | --- |
| **pins** | **开本刀** | 绿地表/API，契约已写清；接 T6-R 后可演示「队级短结论不串」 |
| Issue 池按队 | **后置** | 现网 Issue **无** `team_id`；要迁库 + `ListIssues` 过滤，非最小刀 |
| T7 模板向导 | **不开** | 权威分期 **P1**；非演示刀 |

对齐 [`BOT_TEAM_TEMPLATE_T6_T7.md`](./BOT_TEAM_TEMPLATE_T6_T7.md) T6 §共享 pins；接 [`NEXT_WAVE_DEBT_T6.md`](./NEXT_WAVE_DEBT_T6.md) 脚注「pins 另单」。

## 一句话

同 WS 双 `team_id`：队 A 写的 pin 在队 B **不可见/不可覆写**；与席位私有 `bot_memories` **分表**，list 标明 scope。日常续命仍 `handoff get`，**不用 pin 替代**。

## 含 / 不含

| 做 | 不做 |
| --- | --- |
| `team_pins`（或等价）`UNIQUE(team_id, key)` + 迁移 | Issue 加 `team_id` / List 按队过滤 |
| REST `GET/POST/DELETE …/teams/{team}/pins` | FE 团队切换器 / 记忆墙 UI 大页 |
| 可选 MCP/CLI `board_pin_list|get|set` | T7 模板创建 / 升版 |
| 箱证据：双队 pin 不串；私有 memory 键不被 pin 覆盖 | 用 pin 替代 handoff / 默认同进共享盘 |

## 验收（P1–P4）

| ID | 标准 | 谁 |
| --- | --- | --- |
| **P1** | 迁移 + API：队路径下 CRUD pin；错/他队 `team_id` → **404** | **Go** |
| **P2** | 同 member 兼两队：A 的 pin key 不出现在 B list；B set 同 key 不覆 A | **Go** unit + **Bot 记忆** 箱跑 |
| **P3** | pin 与 `bot_memories` 分表或分 scope；私有 `last_focus` 等键不被 pin API 读写 | **Go** + 证据注 |
| **P4** | 证据 `docs/qa-shots/T6P-pins-isolate-YYYYMMDD.txt`；产品勾 | **Bot 记忆** → **产品** |

## 派工

1. **Go**：表+REST+单测（双队隔离）  
2. **Bot 记忆**：可选通道接线；主跑箱验 P2–P4  
3. **种子与文档**：API.md / BOT_MCP_CLI 补一行（有通道再补）  
4. **React**：本包**不派**（无 FE 必做）  
5. **QA**：Go/记忆交件后苛刻复测  
6. **产品**：勾 P1–P4；`BOT_TEAM_SEAT_P0` T6 注「pins PASS」或整勾（Issue 池仍脚注后置）

## 成功

- 本文产品勾 + 测绿即 `git push origin/main`  
- Issue 池按队 / T7：**另开时再喊**

## 产品勾选

- [ ] P1–P4 交件 · 产品勾 · T6-P 关
