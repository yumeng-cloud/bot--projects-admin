# ISSUE_ARCHIVE · 苛刻复测清单（QA）

> 权威：`docs/NEXT_WAVE_DEBT_ISSUE_ARCHIVE.md`  
> 总指挥主验四点：①完成批量归档 ②默认 list/board 不见 ③`include_archived` 可拉 ④归档前后 pack 统计 done/total（归档的 done 不计入）  
> 等 Issue 域 / React 分段交后立刻跑；勿盲测。

## 证据
- 主：`docs/qa-shots/ISSUE_ARCHIVE-qa-retest-YYYYMMDD.txt`
- 箱截：`docs/qa-shots/issue-archive-*.png`

## 红线（总指挥四点）

| ID | 步骤 | PASS |
|----|------|------|
| Q1 | 完成列/批量：`POST …/issues/batch-archive` 或等价；Done（及可选 Canceled）可批量归档 | |
| Q2 | 默认 list / board / search / cycle-issues：**不见**已归档（`archived_at IS NULL`） | |
| Q3 | `include_archived=1` 可拉回；`archived_only=1` 仅归档（only 优先） | |
| Q4 | Pack / 列计数 / done/total：归档前后对比——**已归档的 done 不计入**；归档后数字下降或剔净 | |

## 补充苛刻
| ID | 步骤 | PASS |
|----|------|------|
| Q5 | GET 单条已归档仍 **200** + `archived_at`；解档回默认列表 | |
| Q6 | 归档不改 `status`（仍 done/canceled）；与 status 正交 | |
| Q7 | 人类/Bot 同权可调；跨 WS id 拒 | |
| Q8 | FE 箱截：列批量、归档后看板空位、已归档视图、详情「已归档」标（有则抽） | |

## 关报
短报产品+总指挥：Overall · Q1–Q4 · 证据 · HEAD。测绿可 push main。
