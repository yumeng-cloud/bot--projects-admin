# ISSUE_ARCHIVE · 苛刻复测清单（QA · D2）

> 权威：`docs/NEXT_WAVE_DEBT_ISSUE_ARCHIVE.md` @`3238466`（总指挥定稿）  
> 主验：默认 list/board/search **不见**归档；pack/周期计数剔归档；`include_archived=1` / `archived=only` 可拉  
> 等 Issue 域 / React 分段交后立刻跑；**勿盲测**。

## 证据
- 主：`docs/qa-shots/ISSUE_ARCHIVE-qa-retest-YYYYMMDD.txt`
- 本刀：`docs/qa-shots/ISSUE_ARCHIVE-qa-retest-20261024.txt` · **Overall PASS** · HEAD 见证据
- 箱截：`docs/qa-shots/issue-archive-*.png`

## 红线（D2）

| ID | 步骤 | PASS |
|----|------|------|
| Q1 | `POST …/issues/archive` `{ids}`：仅 `status=done`；非 done → 中文 4xx；幂等 | ✅ |
| Q2 | 默认 list / board / search / relations：**不见**已归档 | ✅ |
| Q3 | `include_archived=1` → 含归档；`archived=only` → **仅**归档 | ✅ |
| Q4 | Pack / 列计数 / done/total / 周期聚合：**排除** archived；归档后计数下降且列表无该单 | ✅ |

## 补充
| ID | 步骤 | PASS |
|----|------|------|
| Q5 | GET 单条已归档仍 200 + `archived_at`；status 仍为 done（正交） | ✅ |
| Q6 | Bot/MCP 列表同源默认不返回归档 | ✅ |
| Q7 | FE（有箱截则抽）：完成列批量归档后卡消失；Pack 数字与 BE 一致 | ✅ |

## 关报
短报**产品体验**（+总指挥）：Overall · Q1–Q4 · 证据 · HEAD。测绿可 push main。
