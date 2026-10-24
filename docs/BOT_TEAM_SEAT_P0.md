# Bot 席位 · 换机迁绑 · 多团队共享 · 模板挂载（P0）

> 产品体验 · 2026-10-24 · Asia/Shanghai  
> 前置（已关）：[`BOT_HANDOFF_CONTINUITY_P0.md`](./BOT_HANDOFF_CONTINUITY_P0.md) · [`BOT_CONTINUITY_P0.md`](./BOT_CONTINUITY_P0.md) · [`BOT_HANDOFF_MIRROR_P0.md`](./BOT_HANDOFF_MIRROR_P0.md)  
> 续命权威 = 服务端 `handoff get`；本单补产品面：席位 / 团队 / 交接墙 / 模板挂载。技术对齐 [`BOT_TEAM_TEMPLATE_BRAINSTORM_P0.md`](./BOT_TEAM_TEMPLATE_BRAINSTORM_P0.md)。  
> **P0 定稿**：§1.1 团队 UUID 拉班底 · 席位迁绑 · 补位扩工；验收 **T1–T5、T8–T11**。通道先、FE 席位 UI 随后。进度回总指挥，不 fan-out。

---

## 0. 一句话产品模型

看板 = **多团队共享真相源**（Issue / Wiki / 工时 + Bot 记忆 / Todo / 指派）。  
Bot = **可迁绑的角色席位**（身份挂在工作区·团队下，不钉死某台电脑）。  
本机 Cursor / Agent = **临时演员**；换机只换演员，不换席位与剧本。

### 0.1 模板仓口径

| 件 | 结论 |
|----|------|
| **看板** | 唯一真相源与运行时（Issue / Wiki / 工时 / 记忆） |
| **模板仓（bot-team-template）** | 班底作业手册（席位、进场、审查、决策、补位） |
| **`ymbot` CLI/MCP** | 手册里的手：`team pull` → `seat enter` → `handoff get` |

作业链：**读手册 → ymbot → 写回看板**。

---

## 1. 用户换机一天怎么走（主路径，无文件）

### 早上 · 机 A（老环境）

1. 打开看板同一工作区 → **设置 / Bots（席位）**：确认本机 Agent 已绑到席位（如「产品体验」「React」）。  
2. 开干：`claim` Issue → 过程写 `memory` / `todo` / `comment`（续命 P0 已定）。  
3. （可选）旁路镜像 cron 冷备；**日常不依赖**。

### 中午 · 换到机 B

1. 机 B 打开**同一工作区**入口。  
2. **迁绑**：设置 → Bot 席位 →「绑定此设备 / 此 Agent」（贴同一 Bot Key，或席位邀请码认领）。  
   - 身份跟着**席位**走，不跟着本机文件夹走。  
3. 新会话主航线：`team pull`（若班底未齐）→ `seat enter` → `handoff get` → 恢复记忆 / Todo / 开放 `YM-n` → 接着干（`whoami` 可选自检）。  
4. 默认用 `handoff get` 续工；`export` / `import` 仅救场。

### 晚上 · 多团队

- 同一人可属 Team α / Team β；切团队 = 切换可见 Issue 池与席位组。  
- 记忆默认 **席位 × 工作区**；标「团队共享」的进看板墙（§2.3）；个人草稿不出队。

**一天心智**：换机 = 换座位上的人；看板与角色席位没换。

### 1.1 团队 UUID 拉班底 · 迁绑 · 补位扩工（P0 硬路径）

**班底**：某团队（`team_id` = UUID）下预设的角色席位清单（slug + 显示名 + 可选 `playbook_ref`），来自模板或手工。

| 动作 | 产品行为 | 不打架规则 |
|------|----------|------------|
| **拉班底** | `GET/POST …/teams/{team_id}/roster`（或等价）：按团队 UUID 列出/物化空席位；幂等——同 `team_id`+`role_slug` 不双建 | Idempotency-Key；重复拉 = 返回已有席位，**不**重置 memory |
| **席位迁绑** | 机 B / 新 Agent：绑定 Key→席位（member）；**主航线** `team pull` → `seat enter` → `handoff get`（slug=席位；whoami 可选自检） | **单活**：第二绑定 409/冲突提示；确认后旧绑失效；claim 冲突只读占用者，**禁止瞎抢** |
| **补位** | 仅旧席**离职/停工**后接管；或认领**空席**/新员工新席 | **活跃席禁止顶替**；`--force` **仅**已停/已离职席；新人只加新席，不挤活跃席 |
| **扩工** | 同团队加新角色席位 / 同一人兼第二席位 | 兼岗须**显式**第二席位绑定；默认记忆按席位隔离，禁止两席位写同一 `last_focus` 键互相覆盖（键加 `seat_slug:` 前缀或分 member） |
| **扩团队** | 新 `team_id` 再拉同一模板班底 | 新团队新席位组；**记忆不串队**（member/seat 作用域含 team 或 workspace 隔离） |

通道主航线（唯一；与 bot-team-template 同序）：`board_team_pull` → `board_seat_enter` → `board_handoff_get`。slug=席位。`board_whoami` 可选自检；`migrate enter` = `seat enter`。说明书：`bot-team-template/playbook/看板席位迁绑.md`。

### 1.2 现成技术形状（Bot 记忆角色回帖 · 对齐实现）

| 概念 | 现状 |
|------|------|
| **角色认领** | `BotRoleClaim(workspace_id, role_id, member_id, claimed_at, released_at?)`；同 role **单活**（`released_at IS NULL`）；POST claim 冲突 **409** |
| **BotRole** | `name, slug, responsibilities, skills, default_project_ids`；列表可嵌 `claim` |
| **Handoff** | 无表、实时组装：`claimed_role?` + memories + todos(todo|doing) + 指派开放 issues |
| **记忆** | `bot_memories` UNIQUE`(workspace_id, member_id, key_name)`；可选 `role_id` 未当共享墙；**无**团队/看板级共享记忆表（→ P0.5） |

P0 班底/迁绑/补位：优先复用 **BotRole + Claim 单活 + 409**；`team_id` roster = 按团队 UUID 物化/列出 role 席位，不另发明第二套 claim。


### 1.3 席位 status 与入场码（实现对齐）

- **`team_seats.status`**：`open`（空席）→ `active`（已绑）→ `stopped`（停工）/ `left`（解绑离职）。活跃席禁止 force 顶替。
- **入场码**：`POST/GET …/teams/{teamID}/invites`（Admin）；`POST /team-invites/redeem`。列表不回明文；详见 [`BOT_TEAM_JOIN_KEY_P0.md`](./BOT_TEAM_JOIN_KEY_P0.md)。

---

## 2. 看板里三块面

### 2.1 团队面（Teams）

- 工作区下多团队；成员、默认 Cycle、可见项目。  
- 团队首页：进行中 **Bot 席位条** + 未交接红点（有 open todo / 卡住 `last_focus` memory）。

### 2.2 Bot 身份面（席位，不是「API Key 列表」）

每席位卡片：

| 元素 | 说明 |
|------|------|
| 角色名 | 产品体验 / React / QA … |
| 绑定状态 | 已绑设备数、最后 `handoff get` 时间、Key 末四位 |
| 动作 | **绑定此机** / **解绑** / **轮换 Key** / **查看交接包摘要** |

**迁绑规则（P0）**：新机绑定 → **单活**（旧机冲突时提示并要求确认解绑）；禁止静默多活抢同一席位写回。

### 2.3 交接面（Handoff 墙）

- 席位详情抽屉：实时组装摘要（角色、`last_focus`、todos、开放 Issue）= 现 `handoff get` 的人话版。  
- 「传给同事 / 另一 Bot」：摘要钉到 Issue 评论，或生成**只读交接卡**（**不复制 Key**）。  
- 多团队：**共享记忆墙** = 团队级 pins（决策、约定、模板版本）；席位记忆默认私有于席位。

与现网对齐：续命权威 = 服务端 get；磁盘镜像 = 旁路；本单加 **FE 席位 / 团队 / 交接墙**。

---

## 3. 模板仓深度融合（挂载点先写清）

| 层 | 是什么 | 挂载点（产品契约） |
|----|--------|-------------------|
| **模板仓** | **班底作业手册**（bot-team-template）：角色定义、进场剧本、验收 md、例程建议 | 工作区可「按手册拉班底」；挂载 `team_id` + playbook_ref |
| **Playbook** | 手册内路径，如 `playbook/*.md` / `roles/<slug>` | 席位字段：`seat.playbook_ref` = `{ repo, ref, path }`；**ymbot 只执行手册步骤** |
| **验收单** | 仓内 `docs/*_P0.md` 或 playbook 内嵌勾选 ID | 席位或团队挂 `acceptance_doc_ref`；开干 Issue 可打同标签 |
| **运行时** | 看板 Issue + Wiki + Bot memory 键约定 | **只读升级**：模板升版 → UI diff「更新席位 playbook / 保留现场」；**禁止**静默覆盖 memory |
| **例程** | 仓内 cron / routine 描述 | 团队级「建议例程」列表；一键复制到本工作区例程（人工确认） |

### 3.1 开新团队路径

1. 选模板（如「育蒙看板标准班底」）→ 一键生成 Team + N 个**空席位**（尚未绑 Key）。  
2. 人把各 Cursor Agent / `ymbot` Key **迁绑**到席位。  
3. 开干后经验写回看板；模板仓只读升级，不覆盖现场 memory。  
4. 新团队复用同一模板 → 席位形状一致，**记忆不串队**。

### 3.2 非本轮

跨工作区记忆联邦；自动把本机任意 Agent「镜像成席位」（安全风险）；模板写回用户仓（单向：仓 → 看板）。

---

## 4. 分期

| 期 | 范围 | 谁（建议） |
|----|------|------------|
| **P0（本单）** | 团队 UUID 拉班底 API；席位迁绑单活；补位/扩工不打架；换机 get 续命；handoff 摘要可后置 FE | **Bot 记忆角色 + Go 后端**（先）→ React 席位 UI → 产品体验勾选 |
| **P0.5** | 团队切换 + 团队共享记忆墙（pins） | 同上 |
| **P1** | 模板仓「创建团队」向导 + playbook 挂席位 + 升级 diff | 种子与文档 + React + Bot 记忆角色 |
| **非本轮** | 跨工作区联邦、自动本机镜像成席位 | — |

实现以本文验收为准；细节对齐 `BOT_TEAM_TEMPLATE_BRAINSTORM_P0.md`。

---

## 5. 验收清单（产品勾选 · 箱截 / CLI 证据）

| ID | 标准 | 状态 |
|----|------|------|
| **T1** | 设置中存在 **席位** 叙事（角色名 + 绑定态），不再只有裸 Key 列表 | [x] `73456bc` · `docs/qa-shots/seat-t1-roster.png` |
| **T2** | 机 B：**绑定同一席位** 后，仅 `handoff get`（无 import 文件）能续机 A 的 focus / Todo / 开放 Issue | [x] 2026-10-24 箱验 V1–V3 PASS · `docs/qa-shots/T2-handoff-migrate-20261024.txt`（同 member 双 Key；leave→migrate enter；禁 force/import） |
| **T3** | 迁绑 **单活**：第二机绑定有冲突提示；确认后旧绑失效或只读 | [x] FE 专面 `T3-conflict.png`（禁顶替+换空席/申请新席/返回；无 force）· 见 `BOT_FE_ONBOARD_F1_T3_P0` |
| **T4** | 席位抽屉可见交接摘要（人话版 handoff）；「交接卡」不泄露完整 Key | [x] `seat-t4-handoff.png` · 复制交接卡明示无完整 Key |
| **T5** | 换机引导文案：主路径 = 绑席位 → get；export/import 标为可选备份/救场 | [x] 顶栏「换机怎么续」+ 详情页「可选备份/救场」· `seat-t5-migrate-guide.png` |
| **T6** | （P0.5）团队切换后 Issue 池与席位组隔离；共享 pins 与席位私有 memory 不串 | [ ] 契约见 [`BOT_TEAM_TEMPLATE_T6_T7.md`](./BOT_TEAM_TEMPLATE_T6_T7.md) |
| **T7** | （P1）从模板创建团队：空席位形状与 playbook_ref 挂上；升版 diff 可选更新、默认不覆盖 memory | [ ] 契约见 [`BOT_TEAM_TEMPLATE_T6_T7.md`](./BOT_TEAM_TEMPLATE_T6_T7.md) |
| **T8** | 硬规矩：关键路径箱内自测 / CLI 记录；禁半成品甩用户本机当小白鼠；**禁 fan-out 吵用户** | [x] 2026-10-24 CLI `/tmp/ymbot_t9_t10_evidence.txt` |
| **T9** | **团队 UUID 拉班底**：给定 `team_id`（UUID）ensure/pull roster → 完整角色席位（slug/名/claim 态）；重复调用**幂等**、不双建 role、**不清空** memory/todos | [x] pull×2 五席同 id；`t9_probe` 保留 · HEAD `0968a9a` |
| **T10** | **补位 / 扩工不抢 claim、不抢 in-progress**：① 空席/新席可 claim；② 活跃席已有 `released_at IS NULL` → 默认 **409**，禁止顶替（含 force）；③ 他人 assignee 且进行中 → 不得自动改派；④ **仅停工/离职席**可用 `--force` 接管并留 comment/审计；⑤ 兼岗记忆按 member/seat 隔离；新人只加新席 | [x] 二绑/migrate→409；expand 不偷 YM-1/8/10/15；force 仅停/离职席；④ comment 留痕标 P0.1 补强 |
| **T11** | **席位迁绑续命**：新 Key/新机绑到原 role 后仅 `handoff get` 能续；第二路抢绑同 T3 单活 | [x] 空席可进；活跃占用→409（不可 force 顶替）；`--force` 仅停工/离职席；`frontend:`/`qa:last_focus` 隔离；enter 带回 handoff（与 T2 全文换机剧本可另补 FE） |

### 5.1 通道 P0 必验小节（总指挥点名）

1. **T9 团队 UUID 拉班底** — 入参 team UUID → 班底齐；再拉一次仍同批 id、无重复行、记忆仍在。  
2. **T10 补位扩工不抢** — 场景 A：活跃席已被 claim → 第二 Bot 得 409（不可 force 顶替）；场景 B：Issue 已在他人 in-progress → 不抢走；仅空闲/停工离职席或**新席**可接手。

**关单条件（API/通道 P0）**：**T9 + T10 + T8**（建议连带 T2/T3/T11）全勾即可先关后端通道单。  
**关单条件（含 FE）**：另加 **T1、T4、T5**。T6/T7 另开子单；§3 挂载点已写清。

> **2026-10-24 FE 席位 P0（T1/T3/T4/T5）已关**；**T2** 换机箱验 V1–V3 已交件（产品勾关包见 NEXT_WAVE_DEBT_T2）。

> **2026-10-24 通道 P0 已关**：T8–T11。**FE T1/T3/T4/T5 已勾**。**T2** 箱验已交件（产品关包）。force comment → **N3 已清**。

---

## 6. 状态

- [x] 产品定稿落库  
- [x] 技术对齐附录（BOT_SEAT_BIND_TECH / BRAINSTORM）  
- [x] 总指挥放行实现  
- [x] Bot 记忆角色：通道 CLI/MCP 已按精确契约接线（§7）  
- [x] Go teams API 合入 main `0968a9a`  
- [x] **通道 P0 关单**：T8–T11 已勾（证据 `/tmp/ymbot_t9_t10_evidence.txt` · 2026-10-24）  
- [x] FE 席位 UI **T1/T4/T5**（`73456bc`）  
- [x] **T3** 单活冲突专面（FE 已关 · `T3-conflict.png`）  
- [x] **T2** 换机一天完整箱验 V1–V3 PASS · [`T2-handoff-migrate-20261024.txt`](./qa-shots/T2-handoff-migrate-20261024.txt) → 产品关包见 [`NEXT_WAVE_DEBT_T2.md`](./NEXT_WAVE_DEBT_T2.md)  
- [ ] T6/T7 另开（最小说明已落 [`BOT_TEAM_TEMPLATE_T6_T7.md`](./BOT_TEAM_TEMPLATE_T6_T7.md)）  

**已定稿可开发。** 实现中勿刷屏用户；进度回总指挥 / 产品体验即可。  
**口径**：模板仓 = 班底作业手册；`ymbot` = 手册里的手（与模板 README 同步）。

---

## 7. 通道实现状态（Bot 记忆角色 · 2026-10-24）

| 项 | 状态 | 备注 |
|----|------|------|
| CLI `ymbot team pull --uuid` | 已接线 | 主航线① → `POST …/teams/{id}/roster/ensure` → `{roster}`；slug=席位 |
| CLI `seat enter`（`migrate enter`=别名） | 已接线 | 主航线②：**bind**（`--team`）或 role claim；活跃席 409；`--force` 仅停工/离职席 |
| CLI `handoff get` | 已接线 | 主航线③：服务端摘要无文件 |
| CLI `seat leave` | 已接线 | `--team` → unbind（顺带 release）；否则 role release |
| CLI `seat expand` | 已接线 | → `POST …/expand` `{project_id?,limit?}` → assigned/skipped；`--dry-run` 兜底 |
| MCP `board_team_pull` → `board_seat_enter` → `board_handoff_get`（+ leave/expand） | 已接线 | 与模板说明书同序同术语；默认无 UUID |
| botclient `TeamPull` / `SeatEnter` / `SeatLeave` / `TeamSeatBind` / `TeamSeatExpand` | 已落地 | 记忆键 `SeatMemoryKey(slug,key)`；**未**双写 REST handler |
| Go `teams` / `team_seats` handler + 迁移 | **Go 同事** | 通道已按精确契约接线 |
| T9–T11 真服务证据 | **PASS** · 2026-10-24 · HEAD `0968a9a` · `/tmp/ymbot_t9_t10_evidence.txt` | 产品体验已勾；通道 P0 关单 |

回链：[`BOT_SEAT_BIND_TECH_P0.md`](./BOT_SEAT_BIND_TECH_P0.md) · [`BOT_TEAM_TEMPLATE_BRAINSTORM_P0.md`](./BOT_TEAM_TEMPLATE_BRAINSTORM_P0.md) · **说明书** `bot-team-template`（根 README 唯一作业链 + `playbook/看板席位迁绑.md`）
