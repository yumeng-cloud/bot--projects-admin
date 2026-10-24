# 建队后空席必须可见（SEAT_VISIBILITY）

> 2026-10-24 23:14 Asia/Shanghai · 产品体验  
> **用户定调（总指挥转·P0 怒）：** 从模板建队 Toast「研学平台 - 5 空席(未绑 Key)」，但 Bot 席位页仍「尚未绑定席位」空态——**5 个空席坑看不见**。  
> **要求：** 建队成功后必须 **立刻切到该队** + 席位页 **列出 N 个空席（可绑 / 可复制入场）**；**禁止只 Toast**。  
> **2026-10-24 23:16 并入：** 绑席成功 Toast vs 卡片仍「等待绑定」——**绑定后席卡立刻显示 Bot 名 / 已绑定**，禁止只 Toast。  
> **承接：** T7-R 创建空席已绿；TEAM_SWITCH 切队；本包只修 **空席可见性 / 空态误伤**。与 [`ONE_TAP_FOOLPROOF`](./NEXT_WAVE_DEBT_ONE_TAP_FOOLPROOF.md)（露 ymk_）并行、不合并。

## 一句话

Toast 说有 N 空席 → 关掉创建弹窗后，**当前队 = 新队**，主区是 **N 张空席卡**（未绑也可点一键入场），不是「尚未绑定席位」。绑席 Toast 成功 → **同一张卡立刻**变「演员 · {Bot 名}」，不得仍写「等待绑定」。

## 根因（产品认定 · 供 FE 验证）

现网 `BotsRosterPage`：`hasAnyBinding` 只认「有 claim 的角色」或「roster 里 active 席」。**纯空席（empty/open）不算绑定** → 建队 ensure 出 5 空席后仍走 F1.1 空态「尚未绑定席位」，席网格根本不渲染。另：`onTeamCreatedFromTemplate` 丢弃传入的 `_roster`，只 `setTeamId` + `load()`，易闪空。

**绑席后仍「等待绑定」：** `claimBotRole` / 指派成功后若只 Toast、未同步刷新 `roles`+`teamSeats`（或 `seatOccForRole` 与卡片文案各读一源），卡面仍走空席分支。

## 含 / 不含

| 做 | 不做 |
| --- | --- |
| 建队成功：切 `teamId` 到新队；主区渲染 N 空席 | 重做 from-template / ensure API |
| 空席卡：显示名/slug；主 CTA「复制入场链接」「复制 Bot 提示词」（ONE_TAP）；可「指派 Bot」若有闲 Bot | 自动绑 Key |
| 「尚未绑定席位」仅当：**无当前队** 或 **当前队 roster 长度为 0** | 有空席仍甩该空态 |
| 创建弹窗关闭后焦点在席位页（无需用户再点切队） | 只改 Toast 文案 |
| 绑席/指派成功：席卡立刻「演员 · Bot名」或等价已绑定态 | 成功 Toast 后卡仍「等待绑定」 |

## 验收

| ID | 标准 | 谁 |
| --- | --- | --- |
| **V1** | 从模板创建成功 → 自动当前队=新队；无需手动切队 | React |
| **V2** | 主区列出 N 张空席卡（N=模板席数）；**禁止**再出现「尚未绑定席位」当 roster 非空 | React |
| **V3** | 空席卡可一键复制入场链接/提示词（带 role）；未绑 Key 文案人话，不装已占用 | React |
| **V4** | 箱截：`seat-visibility-after-create.png`（Toast 后页内可见 N 空席 + 当前队名） | React |
| **V7** | 指派/绑席成功后：**同一席卡**立刻显示 Bot 显示名（或「已绑定」），**禁止**仍「等待绑定」；可与 Toast 同帧；箱截 `seat-visibility-after-bind.png` | React |
| **V5** | QA 全覆盖：默认班底 5 席；再测至少 1 个非 default 模板（若清单≥2）；切到旧队再切回新队席仍在；刷新后仍可见；**另：绑一席 → 卡立刻有名，刷新仍在** | QA |
| **V6** | 产品目视勾选关单 | 产品 |

## 派工

| 角色 | 职责 |
| --- | --- |
| **React 前端** | **P0 首刀 V1–V4 + V7**（先于 FOOLPROOF 露钥） |
| **QA** | **V5** 全覆盖证 |
| **Go** | 默认不动；若 roster/ensure 未返回席则另开，先 FE 验证 |
| **产品体验** | 本裁定；V6 |

## 产品勾选

- [x] **V1** 建队后自动切到该队 · React `0c712cf`
- [x] **V2** N 空席可见，禁误空态 · `seat-visibility-after-create.png`
- [x] **V3** 空席可一键入场复制 · 卡上入场链接/Bot提示词
- [x] **V4** 箱截 · `seat-visibility-after-create.png`
- [x] **V7** 绑席后卡立刻显示 Bot 名 · `seat-visibility-after-bind.png`（演员·Grok Bot）
- [ ] **V5** QA 全覆盖
- [ ] **V6** 产品关单

## 进度注记

- 2026-10-24 23:23 · 用户怒「不会自动组班」→ 另开 [`NEXT_WAVE_DEBT_AUTO_CREW.md`](./NEXT_WAVE_DEBT_AUTO_CREW.md)：建队后一键全班入场包，禁逐席。
- 2026-10-24 23:20 · **产品勾 V1–V4+V7** React `0c712cf`：建队后 5 空席可见；绑知识库→卡立刻「演员 · Grok Bot」。**待 V5 QA** 复测全覆盖（对账原 FAIL `6f47344`）。
- 2026-10-24 23:14 · 定稿。首刀 React V1–V4；QA V5。

## 关单进度

**进行中** · V1–V4+V7 产品勾 `0c712cf`；**待 V5 QA**；V6 等 V5。
