# P0 · 按需入场码 INVITE_ON_DEMAND

> 2026-10-25 10:50 Asia/Shanghai · 产品体验主持拍板  
> **状态：已关**（I-P1 · Wiki `INVITE_ON_DEMAND_CLOSE.md`）。  
> **承接：** 截图怒——建队/全班包一开就墙式堆 `ymt_`（≈模板空席数，叠历史可达≈18）；列表露 `team:uuid:slug` 脏串；Admin 一人吐全班钥。  
> **强令：** 入场码是钥匙不是海报——要一把才开一把。合议→修→QA 自测再交。**不问用户。**  
> **合议：** React / Bot记忆 / QA / Go / 总指挥 **全同意**（记忆反对物理删全班包 → 已采纳为二级+勾选确认）。可提前拍板落债。

## 气质一句

**入场码是钥匙不是海报——要一把才开一把。**

## 主人硬约束（锁死）

1. **按需生成**，禁默认批量堆码。  
2. 能直接指定 Bot 用平台入场券 **K（`ymk_` / 平台侧）** 就走 K，需要时再生成。  
3. **活跃席禁顶替**；新人新席；Admin **不**一人吐全班钥；列表人用可读（**禁裸** `team:uuid:slug`）。

## 合议结论（允许 / 禁止）

| | 规则 |
| --- | --- |
| **禁止** | 建队成功同帧自动 `setCrewPackOpen` → `issueCrewPack` 对每非 active 席循环 `POST …/invites`；无确认一键全班 `ymt_`/`ymk_`；`issueInvitePlain` 静默加堆（已有有效码仍新建）；活跃席 mint 顶替码；列表渲染 raw `team:uuid:slug` / 脏 `role_hint`；新开 `invites/batch`；解冻 Soft V2 通码静默绑 `stopped\|left` |
| **允许** | 指定团队+角色：有可用 K → 查用；无 → 生成一枚；显式「替换」才吊旧发新，不替换可直接发新一枚（勿堆垃圾）；默认 CTA=按席发一把；全班包作二级+勾选空席确认后再签；脏串→「未命名席」 |
| **码轨** | `ymt_`=入场 · `ymk_`=身份 · `ymw_`=人类仓邀 —— 三分轨，勿混 |

## 状态机（指定团队 + 角色）

1. 有可用平台券 **K（`ymk_`）** → 查用（Lookup / 既有 whoami·key 路径）。  
2. **无** → 生成一枚（`ymt_` 入场，或按需另开 `ymk_` 身份；路径分家）。  
3. 显式「**替换**」才吊销旧发新；不替换可直接发新一枚（**勿堆垃圾**）；活跃席顶替 → **409**。

## 根因路径（QA + React 箱证，非猜）

| 凶 | 路径 |
| --- | --- |
| **主凶** | `BotsRosterPage.onTeamCreatedFromTemplate` 建队成功有空席 → 同帧 `setCrewPackOpen(true)` → `CrewPackRevealModal` `useEffect` 一 open 就 `issueCrewPack`（`autoCrewPack.ts`）对每个非 active 席循环 `POST …/invites` → 已签发垃圾墙（≈模板空席数，叠历史可达≈18） |
| **次凶** | 手动「全班入场包」同 Modal **无确认**即 N 签 |
| **叠堆** | `issueInvitePlain` 每次点击新建码，**不查**该席是否已有有效 `ymt_` |
| **丑** | `IssuedInvitesPanel` `seatGroupLabel` 对不上席表时直接渲染 raw `role_hint`/key → 露 `team:uuid:slug` |

**堆码归因（Go @`26564df`）：** `EnsureRoster`/seed **不**插码；堆码 = FE `autoCrewPack` 循环 POST。主修在 **FE**。

## API 现状（Go @`26564df` · 写入债）

| 面 | 现状 |
| --- | --- |
| `POST/GET/revoke …/teams/{id}/invites` | 一次 **1×`ymt_`**；`seat_slug`/`role_hint` **可选**；无 `bot_id`；List 全状态、无席过滤、无明文 |
| 吊销 | 有；未用 GC **无**；**无** `invites/batch` |
| `ymk_` | `api-keys` / redeem bootstrap（与 invite 分家） |
| `ymw_` | 仓邀请隔离 |
| seed | `EnsureRoster`/seed 不插码 |

## Go 本刀目标契约（I-G1 · 锁死）

主修在 FE；Go 本刀若动则按下表（**禁**新开 batch）：

| # | 契约 |
| --- | --- |
| **1 Lookup** | `GET …/teams/{teamID}/entry?seat_slug=` **或** `?role_slug=`（二选一必填）<br>• 优先：席 `active` + Bot 有未吊销 api-key → `{kind:"ymk_", key_id, token_prefix/tail, member_id, display_name}`（**无** token 明文）<br>• 否则 active 未耗尽 `ymt_` → `{kind:"ymt_", invite_id, code_tail, status, seat_slug}`（**无** code 明文）<br>• 都无 → **404** |
| **2 Create** | `POST …/invites` **须** `seat_slug` 或 `role_hint`/`role_slug`（缺 → **400**）；活跃席 `occupancy=active` 再 invite → **409**；`replace=true` 先吊该席/角色 active `ymt_` 再发 1；`replace` 缺省/false 可不吊旧直接再发 1；**一次 1 码**；**禁**开 `invites/batch` |
| **3 ymk_ 明文** | 走现 `POST …/api-keys`；**勿**与 `ymt_` Create 混 |
| **4 List** | 可加 `?seat_slug=&status=active`；响应加 `label`（席显示名 / 角色中文） |

（若 Lookup 可完全文档化既有 whoami/key 路径且 FE 不需新端点，产品勾 I-G1 时可注明「免新路由、文档化既有」——但 Create 强制 seat/role + 活跃 409 + List label **仍建议落地**。）

## Soft V2 / SEAT_REFILL 边界（不解冻 Soft）

| 项 | 本刀 |
| --- | --- |
| Soft V2 通码静默绑 `stopped\|left` | **仍冻** · 本刀不是 V2 |
| Soft 另两件（见 [`SOFT_RULING`](./NEXT_WAVE_DEBT_SOFT_RULING_20261025.md)） | **仍冻** |
| SEAT_REFILL / BIND / SHOT / V2_GUIDE | **不改** 席三态 Tag（在席 / 空闲（从未占用）/ 停工可补）；补位入口不拆 |
| SEAT_PANEL_UX 入场码折叠尺子 | **继承**：已签发降次要；同组有效 ≤1～2；吊销二级 |

## 三分轨

| 前缀 | 用途 |
| --- | --- |
| **`ymt_`** | 入场（队席 invite） |
| **`ymk_`** | 身份（Bot api-key / 平台券 K） |
| **`ymw_`** | 人类仓邀（与上二者隔离） |

## FE 信息架构（锁死）

- **已签发**降次要（折叠摘要；勿一进墙）。  
- 脏串 / 对不上席表 → 显示 **「未命名席」**；**永不**渲染 `team:uuid:slug`。  
- 有效按席 **≤1～2**（继承 SEAT_PANEL 尺子）。  
- **空态勿 null**：CTA 主 = **按席发一把**；全班包 = **二级 + 勾选空席确认**后再签（**不**物理删除入口）。  
- 砍建队自动挡层吐 N；有有效码 → 复用复制或「吊销后再发」，禁静默加堆；在席提示已有人 → 禁 mint 顶替。

## 勾选

| ID | 标准 | 谁 |
| --- | --- | --- |
| **I-G1** | Lookup（`…/entry?seat_slug\|role_slug`）或文档化既有 whoami/key 路径；CreateInvite **须**绑 seat/role（缺 400）；活跃席顶替码 → **409**；`replace` 语义如上；List 可滤席+`label`；**禁**无确认一键全班 `ymk_`；**禁** `invites/batch` | Go |
| **I-R1** | 砍建队自动 `setCrewPackOpen` / 自动 `issueCrewPack` | React |
| **I-R2** | `issueInvitePlain` 有有效码 → 复用复制或「吊销后再发」，禁静默加堆；在席提示已有人禁 mint 顶替 | React |
| **I-R3** | `IssuedInvitesPanel` 人话 label；脏串永不渲染；空态 CTA；已签发降次要 | React |
| **I-R4** | 全班包改二级 + 勾选空席确认后再签（不物理删除） | React |
| **I-M1** | CLI/MCP/`BOT.md` 对齐：有 K 用 / 无则生成 / 在席禁顶 | Bot 记忆 |
| **I-Q1** | 建队默认不弹堆码；按需一码；列表无脏串；K 可用；禁甩主人；箱截 | QA |
| **I-P1** | 产品关 | 产品体验 |

## 产品勾选

- [x] **I-G1** Lookup / Create 强制席 / 活跃 409 / List label
- [x] **I-R1** 砍建队自动堆码
- [x] **I-R2** 按席复用 / 禁顶替
- [x] **I-R3** 人话 label + 空态 CTA
- [x] **I-R4** 全班包二级确认
- [x] **I-M1** 通道对齐
- [x] **I-Q1** QA PASS† @d99e14b / 7f1e576 / 1257830 / b0a4d3d · Go a034a15（纪要 INVITE_ON_DEMAND-I-Q1-qa-20261025.txt）
- [x] **I-P1** 关单 · 2026-10-25 · Wiki [`INVITE_ON_DEMAND_CLOSE.md`](./wiki-dogfood/INVITE_ON_DEMAND_CLOSE.md)

## 派工

| 角色 | 首刀 |
| --- | --- |
| **React 前端** | I-R1～R4（主凶在 FE；优先砍自动堆） |
| **Go 后端** | I-G1（Lookup + Create 强制 + 409；可与 FE 并行） |
| **Bot 记忆** | I-M1（跟 Go/FE 契约） |
| **QA** | I-Q1（FE/Go 可跑后；禁甩主人） |
| **产品体验** | 本裁定；I-P1 |

## 非目标

- Soft 三件仍冻；**不解冻** Soft V2 通码静默补。  
- **不改** SEAT 三态 Tag / SEAT_REFILL 补位主路径。  
- **不**新开 `invites/batch`；不物理删除全班包入口。  
- **角色锚模型档位**（决策/审查=高级，写码=轻量）为后续可扩；**本刀只铺按需入席，不实现档位配置**。勿挡此方向。

## 看板 Issue

- 标题含 **INVITE_ON_DEMAND**  
- 描述链相对路径 `docs/NEXT_WAVE_DEBT_INVITE_ON_DEMAND.md` + 开刀 commit `d7c35c473f77ef8fb891758d01af936a7c72ec39`  
- Issue 链接：_待填_（缺口：公网狗粮 Bot 禁无父游离 Issue；仓内原无父票可挂。已建项目「产品债」待人类建父票后补链）

## 进度注记

- 2026-10-25 10:50 · **开刀拍板**。合议 React/记忆/QA/Go/总指挥全同意；Go API 契约锁死；主修 FE 砍自动堆；Soft 仍冻。
- 2026-10-25 11:00 · **I-M1 通道对齐**（Bot 记忆）：`BOT.md` / `BOT_MCP_CLI.md` + CLI/MCP/join 瘦文案落「有 K 用 / 无则生成 / 显式替换才顶 / 在席禁顶」；`identity_reused` 续队不重发；Soft V2 **未动**。未改 Go Create/Lookup 闸（I-G1）。
- 2026-10-25 · **勾 I-R1～R4** React `b0a4d3d`（砍建队自动堆；复用/禁顶；人话 label；全班包确认）
- 2026-10-25 · **勾 I-G1** Go `a034a15`（LookupEntry / Create 强制席+replace+活跃409 / List label）

- 2026-10-25 11:37 · **I-Q1 PASS†** QA：IQ1-1～8 全绿；IQ1-4 B ymk_ Lookup retest @d99e14b（CTA 查用身份钥匙 · Modal ·66c7 · GET entry 200 · 无 POST）；总纪要 docs/qa-shots/INVITE_ON_DEMAND-I-Q1-qa-20261025.txt。**不勾 I-P1**（产品关）。
- 2026-10-25 · **关单 I-P1** · Wiki `docs/wiki-dogfood/INVITE_ON_DEMAND_CLOSE.md`；锚 I-M1 `22cc85e` / I-R* `b0a4d3d` / I-G1 `a034a15` / Soft† `1257830` / Lookup `7f1e576` / 在席查用 `d99e14b`；I-Q1 PASS†（IQ1-1～6 + API · `5a9fc06`）。状态：**已关**。

## 关单进度

**已关** · I-M1 + I-R1～R4 + I-G1 + I-Q1 PASS† + I-P1 · Wiki [`INVITE_ON_DEMAND_CLOSE.md`](./wiki-dogfood/INVITE_ON_DEMAND_CLOSE.md)。Soft 仍冻。
