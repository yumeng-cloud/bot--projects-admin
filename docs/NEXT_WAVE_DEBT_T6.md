# 下一小包 · T6 最小 · 团队切换隔离箱验

> 2026-10-24 18:04 Asia/Shanghai · 产品体验 · 总指挥调度  
> **本包：仅团队切换隔离。不含 pins。T7 不做。**

## 一句话

同工作区两个 `team_id`：切/拉队 A 与队 B 时，**席位组（roster）不串**；能演示的 Issue「池」也不串（见下边界）。**禁止**本包做 `team_pins` / 共享记忆墙。

## 含 / 不含

| 做 | 不做 |
| --- | --- |
| 双 team 可 `team pull` / ensure | **pins API / 表 / MCP** |
| roster slug/claim 按 `team_id` 隔离可证 | T7 模板创建向导 |
| 证据 txt 箱跑 | 大改 `ListIssues` 加 `team_id`（若现网 Issue 无队字段 → **本包不挡**：用「各队 roster 绑定 member 的开放 Issue / 或分 project」能演示的最小不串即可，并在证据注明口径） |

对齐切片说明亦可参考 [`NEXT_WAVE_DEBT_T6R_VCLOSE.md`](./NEXT_WAVE_DEBT_T6R_VCLOSE.md) 刀 B；**以本文为准**。刀 A 空海复验可并行，不挡本包。

## 验收（V1–V3）

| ID | 标准 | 谁 |
| --- | --- | --- |
| **V1** | 同 WS 建/确保 **TeamA + TeamB**（可 seed） | **Bot 记忆**主跑；缺 API/seed → **Go** |
| **V2** | `ymbot team pull`（或等价）A 与 B：roster **席位集合或 claim 占用互不串**（A 的 bound 不出现在 B 的同 slug 误绑） | **Bot 记忆** |
| **V3** | 证据 `docs/qa-shots/T6-team-isolate-YYYYMMDD.txt`：步骤+断言；**无** pin 相关命令 | **Bot 记忆** → **产品**勾 |

## 成功

- 本文下方产品勾 + `BOT_TEAM_SEAT_P0` **T6** 注「最小隔离 PASS；pins 仍 P0.5 余」或整勾时脚注 pins 后置  
- 测绿即 push

## 产品勾选

- [x] V1–V3 交件 · 产品勾 · T6 最小关（2026-10-24 18:08 Asia/Shanghai）
  - 证据 `docs/qa-shots/T6R-roster-isolate-20261024.txt` · `97ca3b4` + Go `96df495`
  - 双队 seat_id/role_id 不同；A=Grok Bot / B=BotB J2 同 slug frontend claim 不串
  - **pins 另单**；Issue 无 team_id 池隔离另单
  - QA 苛刻复测 PASS · `docs/qa-shots/T6R-qa-retest-20261024.txt`（unit + 双队 claim/pull）


## 交件注（Bot 记忆 · 2026-10-24 18:06 Asia/Shanghai）

- V1–V3 **PASS** · 证据 [`qa-shots/T6R-roster-isolate-20261024.txt`](./qa-shots/T6R-roster-isolate-20261024.txt)（等同 V3 文件名约定）
- Go `96df495` team-scoped bot_role；箱内重建重启 API 后双队同 slug bind claim 不串；旧席 re-ensure 改挂
- **pins 未做**；Issue 无 `team_id` → 池隔离待产品/Go（本包不挡）
- 产品勾框仍留给产品；`BOT_TEAM_SEAT_P0` T6 已注「最小隔离 PASS；pins 仍 P0.5 余」
