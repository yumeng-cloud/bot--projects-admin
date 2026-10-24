# 下一小包 · 不定整开 T6/T7 · 两刀立刻干

> 2026-10-24 18:03 Asia/Shanghai · 产品体验  
> **裁定**：不立刻整开 T6 全文（pins）/ T7。  
> **总指挥 18:04 纠偏**：T6 **最小隔离箱验** 权威清单改见 [`NEXT_WAVE_DEBT_T6.md`](./NEXT_WAVE_DEBT_T6.md)（仅切换隔离、**不含 pins**）。本文刀 B = 同范围切片；刀 A 空海仍有效。

## 为什么不整开

| 项 | 期 | 阻塞 |
| --- | --- | --- |
| **T6 全文** | P0.5 | pins 新 API + Issue 池隔离设计未落地 |
| **T7** | P1 | 模板向导+升版 diff，非本波演示刀 |

## 本包两刀（立刻）

### 刀 A · V-CLOSE 空海记债复验关单

| ID | 标准 | 谁 |
| --- | --- | --- |
| **V1** | 箱 5174 重截 Cycles / 工作区 Home / 设置成员：短列表**无**下半屏空罐 | **React**（`87351a9` 已合） |
| **V2** | 视觉总监或产品目视；勾清 [`CYCLES_WS_VISUAL_DEBT.md`](./qa-shots/CYCLES_WS_VISUAL_DEBT.md)（现仍写「非放行」过时） | **产品** + **UI视觉/总监** 可辅 |
| **V3** | C-V3「Cycles」中文「周期」若未齐则顺手热修 | **React** |

### 刀 B · T6-R 双队 roster 隔离箱验（T6 最小切片）

| ID | 标准 | 谁 |
| --- | --- | --- |
| **R1** | 同 WS 两 `team_id`：`team pull` A/B roster **席位组不同或不串 claim**（契约 [`BOT_TEAM_TEMPLATE_T6_T7.md`](./BOT_TEAM_TEMPLATE_T6_T7.md) T6-R · 通道 [`BOT_MCP_CLI.md`](./BOT_MCP_CLI.md) **§11.4**） | **Bot 记忆**（主）· **Go** 缺则补 seed/ensure |
| **R2** | 证据 `docs/qa-shots/T6R-roster-isolate-*.txt`；**不做** pins、**不改** Issue List 按 team | 同上 |
| **R3** | 产品勾本文件；`BOT_TEAM_SEAT` T6 **不**整勾，只注「T6-R 切片 PASS」 | **产品** |

> **Go（辅）**：`EnsureRoster`/`CreateSeat` 现按 `team:`+teamID 建独立 `bot_role`，同 WS 双队同 slug **claim/bound 不串**；测例 `TestTwoTeamsRosterClaimIsolation`。记忆箱验：`POST …/teams` → `…/roster/ensure` → `GET …/roster` → `…/seats/{id}/bind`。

## 明确不做

- `team_pins` / 共享记忆墙  
- Issue 列表按 `team_id` 过滤大改  
- 从模板创建团队向导（T7）

## 产品勾选

- [x] 刀 A V-CLOSE（2026-10-24 18:06 Asia/Shanghai）
  - `81-cycles-recheck.png`：侧栏/标题「周期」；列表 hug+脚注，无空罐
  - `81-workspace-home.png`：入口「周期」；surface hug + 底注
  - `81-settings-recheck.png`：成员 Tab hug+脚注

- [ ] 刀 B T6-R

## 交件注（刀 B · Bot 记忆）

- [x] 执行侧 PASS · [`qa-shots/T6R-roster-isolate-20261024.txt`](./qa-shots/T6R-roster-isolate-20261024.txt) · Go `96df495`
- 产品勾框（刀 B）仍留给产品
