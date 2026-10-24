# 班底模板 · T6 / T7 最小说明

> 2026-10-24 · 种子与文档 · 对齐 [`BOT_TEAM_SEAT_P0.md`](./BOT_TEAM_SEAT_P0.md) 验收表  
> **本文件只定契约与边界**，不实现。总指挥「先写最小说明进 docs」。

| ID | 期 | 一句话 |
|----|-----|--------|
| **T6** | P0.5 | **团队切换隔离** + **共享记忆墙（pins）** 与席位私有 memory 不串 |
| **T7** | P1 | **从模板创建团队**：空席形状 + `playbook_ref`；升版 diff 可选更新，**默认不覆盖** memory |

权威分期仍见 `BOT_TEAM_SEAT_P0` §4；脑暴细节见 [`BOT_TEAM_TEMPLATE_BRAINSTORM_P0.md`](./BOT_TEAM_TEMPLATE_BRAINSTORM_P0.md) §3。

---

## T6 · 团队切换 + 共享记忆墙（P0.5）

### 要什么

1. **团队切换**后：可见 Issue 池、席位组（roster）、补位/扩工范围都落在**当前 `team_id`**；切走不看到他队进行中任务当「我的池」。  
2. **共享 pins（记忆墙）**：团队级短结论（决策、约定、模板版本号）— 全队只读或按角色写；**不是** `bot_memories` 换皮。  
3. **席位私有 memory**：仍按 `member_id`（+ 可选 seat/role）隔离；兼岗不得串读他席私有键。  
4. **不串**：pins ↔ 私有 memory 分表或分 `scope`；list API 必须标明 scope。

### 不做（本阶）

- 跨 workspace 联邦；把私有 handoff 默认同进共享盘；GitHub Projects 当主轨。

### 建议挂载（实现时）

| 面 | 建议 |
|----|------|
| 数据 | `team_pins`（或 memory + `scope=team` + `team_id`）`UNIQUE(team_id, key)` |
| API | `GET/POST …/teams/{team}/pins`；切换团队 = FE/CLI 切 `team_id` 上下文 |
| 通道 | 可选 `board_pin_list/get/set`；日常续命仍 `handoff get`，**不**用 pin 替代 |
| 验收 | 队 A 的 pin 与 Issue 池在队 B 不可见；同 member 兼两席私有 `last_focus` 不覆盖 |

### 关单信号（全文 T6）

箱内：切团队 UI/CLI → 池与 pins 隔离可演示；文档勾 `BOT_TEAM_SEAT_P0` **T6**。

### T6 最小可验 · T6-R（本波切片 · 对齐产品清单）

> 产品裁定见 [`NEXT_WAVE_DEBT_T6R_VCLOSE.md`](./NEXT_WAVE_DEBT_T6R_VCLOSE.md)：**不定整开** pins + Issue 池按队；先验 **双队 roster 隔离**。  
> 通道操作与冒烟步骤对接 [`BOT_MCP_CLI.md`](./BOT_MCP_CLI.md) **§11.4**（`831e806` 尾勾延伸）。

| 项 | 最小标准（可验收） |
|----|-------------------|
| **切换语义** | 无全局 sticky「当前队」env；**切队 = 换入参** `team_uuid` / CLI `--uuid` / `--team` |
| **R1 roster** | 同 WS 两 `team_id`：`team pull` / `board_team_pull` 各拉一次 → 席位组 **seat id 不交叉**；队 A claim **不**出现在队 B roster |
| **R2 作用域** | `seat enter` / `leave` / `expand` 带 `team_uuid` 时，bind/unbind/派叶仅落在该 `team_id` path（REST：`/teams/{teamID}/…`） |
| **错误** | 错/他队 `team_id` 或不存在 → **404**（或通道等价一行中文）；活跃席抢绑仍 **409**（同 T10，与切队无关） |
| **明确不做** | `team_pins`；Issue List 按 `team_id` 过滤；FE 团队切换器；私有 `bot_memories` 改按队（仍 member/seat） |

**箱验证据（记忆主跑）**：`docs/qa-shots/T6R-roster-isolate-*.txt`。  
**勾选口径**：产品勾 T6-R 切片 PASS；`BOT_TEAM_SEAT_P0` **T6 不整勾**（全文仍欠 pins + Issue 池）。

---

## T7 · 从模板创建团队（P1）

### 要什么

1. 选 **bot-team-template（班底作业手册）** → 一键生成 `team` + N 个**空席**（slug/显示名/`playbook_ref`），**尚未**绑 Key。  
2. `playbook_ref` = `{ repo, ref, path }`（或等价）；ymbot / Agent **只执行手册步骤**，不另发明第二套流程。  
3. **模板升版**：UI/CLI 给 diff（更新哪些席位的 playbook / 建议例程）；操作者确认；**默认保留**现场 `bot_memories` / todos / 进行中 Issue。  
4. 同模板开第二队 → 席位**形状**一致，**记忆不串队**（同 T6）。

### 不做（本阶）

- 自动把本机任意 Agent 镜像成席位；模板写回用户 git；静默覆盖 memory。

### 建议挂载（实现时）

| 面 | 建议 |
|----|------|
| 模板清单 | 工作区可读「可用手册」列表（内置或配置 URL） |
| API | `POST …/teams` + `roster/ensure`（已有 T9 幂等）扩展 `playbook_ref`；可选 `POST …/teams/{id}/template/apply`（dry-run + apply） |
| FE | 「从模板创建」向导：预览席位表 → 创建 → 发入场码/迁绑 |
| 验收 | 空席形状与手册一致；升版 dry-run 可见 diff；apply 后 memory 键仍在 |

### 关单信号

箱内走通「选模板 → 空席齐 → 升版不丢记忆」；勾 `BOT_TEAM_SEAT_P0` **T7**。

---

## 与现网关系

| 已有 | T6/T7 增量 |
|------|------------|
| T9 `team pull` 幂等拉班底 | T7 创建时挂 playbook；**T6-R** 双队 pull 不串 roster；全文 T6 再扩池/pins |
| `bot_memories` 私有 | 全文 T6 另立团队 pins（本波不做） |
| `handoff get` 续命 | 不变；pins 不替代 handoff |
| bot-team-template 手册 | T7 只读挂载源 |
| MCP/CLI §11 主航线 | `board_team_pull` / `--team` = 切队入参；冒烟见 **§11.4** |

**派工建议**：实现开子单时 **Bot 记忆角色 + Go**（API）→ **React**（切换/向导）→ **种子与文档**（手册索引与本文勾选）；产品体验关 T6/T7。
