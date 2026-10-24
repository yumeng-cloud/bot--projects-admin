# 下一小包 · T2 换机完整箱验

> 2026-10-24 18:00 Asia/Shanghai · 产品体验  
> 前置已关：I3/§C · N3/N4 · 巡检 P0.5 · 空海密度 · FE T3 冲突专面  
> **本包只箱验剧本，不扩模板 T6/T7**

## 为什么是这一刀

| 已关 | 仍真缺口 |
| --- | --- |
| T3 FE 专面（`T3-conflict.png`） | **T2**：机 B 绑同一席后，**仅** `handoff get`（无 import）续机 A 的 focus / Todo / 开放 Issue |
| H1–H3 同 Key 换会话续命 | 缺「换 Key/换机绑同席」端到端箱证写入 `BOT_TEAM_SEAT` T2 |
| T11 API 迁绑单活 | 产品勾仍欠 T2 一天剧本 |

## 最小范围

| ID | 标准 | 谁 |
| --- | --- | --- |
| **V1** | 机 A（Key₁）进席 → 写 `last_focus` + 至少 1 条开放 Issue（assignee=me）+ 可选 todo | **Bot 记忆** / **QA** |
| **V2** | 机 A leave 或停会话；机 B（Key₂）**绑同一席**（migrate/enter 合法路径，非 force 顶活跃） | 同上 |
| **V3** | 机 B **仅** `ymbot handoff get`（禁 import）可见 focus + 开放 Issue；证据 txt | 同上 |
| **V4** | 产品勾 `BOT_TEAM_SEAT` **T2**；截/日志落 `docs/qa-shots/T2-handoff-migrate-*.txt`（可选 FE 截） | **产品体验** |

## 明确不做

- T6 团队共享 pins / T7 模板创建向导（另开）  
- 重写 handoff / 改续命权威  
- 活跃席 force 顶替（仍禁）

## 成功

- `BOT_TEAM_SEAT_P0` T2 `[x]` + 本文件下方产品勾  
- 证据可复跑


## 交件注（Bot 记忆 · 2026-10-24 18:01 Asia/Shanghai）

- **V1–V3 = PASS**（箱 :8080；无 force；无 import）
- 证据：[`docs/qa-shots/T2-handoff-migrate-20261024.txt`](./qa-shots/T2-handoff-migrate-20261024.txt)
- 路径：同 member 双 API Key → Key₁ `seat enter frontend` + `frontend:last_focus` + claim **YM-11** → Key₁ `seat leave` → Key₂ `migrate enter frontend` → Key₂ **仅** `handoff get`/`--full` 见 focus + YM-11
- `BOT_TEAM_SEAT_P0` T2 已 `[x]`；下方**产品勾选**留给产品体验关包（V4）

## 产品勾选

- [x] V1–V3 交件 · 产品勾 T2 · 关本包（2026-10-24 18:02 Asia/Shanghai）
  - 证据 `docs/qa-shots/T2-handoff-migrate-20261024.txt` · docs `303fcbb`
  - 同 member 双 Key：leave → migrate enter → 仅 `handoff get --full` 见 `T2-V1-focus-machineA-…` + YM-11；禁 force/import
  - `BOT_TEAM_SEAT_P0` T2 已 [x]
  - QA 苛刻复验 PASS · `docs/qa-shots/T2-qa-retest-20261024.txt`（独立 marker `QA-T2-retest-…` + YM-12）
