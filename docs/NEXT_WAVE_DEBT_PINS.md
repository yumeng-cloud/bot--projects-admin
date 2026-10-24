# NEXT_WAVE · pins 最小可验（T6 余债）

> 2026-10-24 · 总指挥自定开工（用户禁闲等）  
> **本包：仅团队共享 pins（记忆墙）。不含 Issue 池按队、不含 T7、不改私有 bot_memories。**

## 目标

同工作区两队：队 A 写入的 pin 在队 B **不可见 / 不可改**；与席位私有 `bot_memories` **分表不串**。

## 做 / 不做

| 做 | 不做 |
|----|------|
| 表 `team_pins`：`UNIQUE(team_id, key)`；字段至少 `id, workspace_id, team_id, key, value, updated_by, updated_at` | Issue List 按 `team_id` 过滤 |
| REST：`GET/PUT/DELETE /workspaces/{ws}/teams/{teamID}/pins`（PUT upsert by key；GET 列本队；DELETE by key） | FE 全局团队切换器大改 |
| 鉴权：须为本队成员（或 Admin）；错队 / 无队 → **404** | 用 pin 替代 handoff |
| MCP/CLI（已落地 `63e69c6`）：`board_pin_list` / `get` / `set` / `delete`；瘦输出 | 跨 workspace；共享墙吞私有 memory |
| FE 最小：Bot/团队设置页「记忆墙」列表+新增/编辑一行 key/value（当前 team 上下文） | 漂亮大屏墙、评论、附件 |

## 验收（箱验）

| ID | 标准 | 主跑 |
|----|------|------|
| **P1** | 队 A `PUT pin` → 队 A `GET` 见；队 B `GET` **无**该 key | Go 测 + 记忆箱验 |
| **P2** | 队 B 对队 A 的 key `PUT/DELETE` → **404** | 同上 |
| **P3** | 私有 `bot_memories` 键与 pin key 同名不互相覆盖 | 记忆 |
| **P4** | 证据 `docs/qa-shots/pins-team-isolate-YYYYMMDD.txt` | QA 复验 |
| **P5** | FE：当前队可见本队 pins；切/换 team 入参后列表换池（无跨队残留） | React → 产品勾 |

## 主辅席

| 席 | 职责 |
|----|------|
| **Go 后端** | 迁移 + handler + 单测；测绿直接 push main |
| **Bot 记忆** | MCP/CLI 已落地 `63e69c6` + 箱验 P1–P3 主跑 |
| **React** | FE 最小记忆墙（并行，不挡 API 关单） |
| **种子** | API.md / BOT_MCP_CLI 补 pins 段 |
| **产品** | 清单对齐 + 勾关 |
| **QA** | P4 苛刻复测 |

## 文档回链（种子）

- REST 契约：[`API.md`](./API.md)「Team pins」
- 通道：[`BOT_MCP_CLI.md`](./BOT_MCP_CLI.md) §11.5 `board_pin_*`
- 实现：Go `6915db4` · FE `4459d01`/`8c2097a` · API 箱证 `1a38ce4` · FE 箱截 `54d5527` · 通道 `63e69c6`

## 产品勾选

- [x] **P1–P3** 箱验 PASS · 2026-10-24 19:00 Asia/Shanghai
  - 证据 [`qa-shots/pins-team-isolate-20261024.txt`](./qa-shots/pins-team-isolate-20261024.txt) · `1a38ce4` + API `6915db4`
  - unit `TestPinsTeamIsolate` / `TestPinsDoNotClobberBotMemories` 产品抽跑 ok
  - 双队 GET 不串；跨队 PUT/DELETE → 404；同名 key 与 bot_memories 分表不串
- [x] **P4** QA 苛刻复测 PASS · 2026-10-24 19:04 Asia/Shanghai
  - [`qa-shots/pins-qa-retest-20261024.txt`](./qa-shots/pins-qa-retest-20261024.txt) Overall PASS
  - FE 墙 [`pins-fe-qa-teamA-wall.png`](./qa-shots/pins-fe-qa-teamA-wall.png) / [`pins-fe-qa-teamB-wall.png`](./qa-shots/pins-fe-qa-teamB-wall.png)
  - 产品目视：A 仅 A 键；B 仅 `qa.fe.wall.B`；无跨队残留
- [x] **P5** FE 最小记忆墙 · 2026-10-24 19:03 Asia/Shanghai
  - FE `4459d01` + 热修 `8c2097a` · 箱截 `54d5527`
  - [`pins-p5-list.png`](./qa-shots/pins-p5-list.png) / [`pins-p5-upsert.png`](./qa-shots/pins-p5-upsert.png) / [`pins-p5-switch-team.png`](./qa-shots/pins-p5-switch-team.png)
  - 产品目视：本队 list 见 pin；切/换队后墙为空或换池，无跨队残留；upsert 表单可用
  - 通道已落地 `63e69c6` `board_pin_*`（不挡 FE 勾）

## 关单

**整单关**：P1–P5 全勾 · 2026-10-24 19:04 Asia/Shanghai。`BOT_TEAM_SEAT_P0` T6：**pins PASS**；Issue 池按队另开 [`NEXT_WAVE_DEBT_T6I.md`](./NEXT_WAVE_DEBT_T6I.md)。
