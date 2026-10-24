# 建队一键造班（BATCH_CREW）· P0

> 2026-10-24 23:31 Asia/Shanghai · 产品体验  
> **用户定调（总指挥转·P0 怒·升档）：** [`AUTO_CREW`](./NEXT_WAVE_DEBT_AUTO_CREW.md) 原 **P1「自动 mint Bot+ymk_」立刻升 P0**。入场包（`ymt_`）不够——人还要 **逐个新建 Bot / 逐个填 Key**。要傻瓜：**建队后一键生成全班 Bot 成员 + 各席身份钥匙 `ymk_`（仅一次可复制包）+ 入场包**；**禁逐席新建、禁逐席填 Key**；人只点一次。  
> **用户钉（总指挥转·人用习惯）：** 任用规则写进验收——**空席才造人入座**；**活跃席禁顶替**；**新人只进空席或新开席**；**仅停工/离职可补位**（对齐 [`BOT_TEAM_JOIN_DECISION_P0`](./BOT_TEAM_JOIN_DECISION_P0.md) E / [`BOT_TEAM_SEAT_P0`](./BOT_TEAM_SEAT_P0.md) 补位）。**禁止逐席填表心智**。一键建队 = **全班就绪可复制包**（钥匙包±入场包）。  
> **承接：** AUTO_CREW C1–C4（全班 `ymt_` 入场包已绿 `fef5914`）+ [`FOOLPROOF`](./NEXT_WAVE_DEBT_ONE_TAP_FOOLPROOF.md)（单把 `ymk_` 挡层）+ [`SEAT_VISIBILITY`](./NEXT_WAVE_DEBT_SEAT_VISIBILITY.md)。本包补 **造人+绑席+批量露钥**。

## 一句话

从模板建队成功（或席页对当前队空席）→ **一个主按钮「一键生成全班」** → 服务端为每空席 **创建 Bot 成员 + 绑席 + 签发 `ymk_`** → 挡层 **仅一次** 露出全班身份钥匙包（可嵌首跑提示词）并附带/可再取入场包；Admin **禁止**再逐席「新建 Bot」「新建 Key」。

## 与 AUTO_CREW 分工

| 包 | 解决 | 码 |
| --- | --- | --- |
| **AUTO_CREW**（**已关**） | 空席仍在时，一次带走 **入场材料** 给外部 Agent 自己 join | 包内只有 **`ymt_`**，禁 `ymk_` |
| **BATCH_CREW（本债·P0）** | 看板侧 **直接造好全班演员并入座**，一次带走 **身份钥匙** | 包内 **`ymk_`**（私贴）；入场包可同帧附带或次按钮再取 |

**「就绪」本债定义：** 默认班底 N 席卡已显示演员名；每人一把当次明文 `ymk_` 已复制或挡层可见；人未再点 N 次新建。

## 产品冻结

| 做 | 不做（本波） |
| --- | --- |
| 一键：仅 **empty** 席 → Bot 成员（`is_bot`）+ bind + mint `ymk_` | 强迫人先有闲 Bot 再指派（AUTO_CREW C5，本波不挡） |
| **任用规则（硬）：** 活跃席禁顶替；batch 只动 empty；新人=空席或 expand 新席；补位仅 `left`/`stopped`/`departed` | 逐席 Form 填名/填 Key/逐席绑；挤占 active |
| **人用主路径：** 建队成功 → 一键 → 挡层拿走全班可复制包即完事 | 建队后还要进「高级·Bot」逐行操作 |
| 挡层：**全班** 身份钥匙包（每席一节：席名/slug + `ymk_` + 「复制本节启动提示词」）；主按钮「复制全班钥匙包」 | 只 Toast「已创建」；列表只留末四却从不露明文 |
| 同帧保留/附带 AUTO_CREW 入场包入口（`ymt_`，禁嵌 `ymk_`） | 把 `ymk_` 打进公开入场 URL / `/join` |
| 入口：① 建队成功主 CTA（升档后优先「一键生成全班」，入场包降次）② 席页有空席时同权主按钮 | 藏进「高级·Bot 成员与 Key」 |
| 幂等：已绑席跳过；已有钥不重发明文（该席节标「已入座/钥不可再取」） | 重复点刷出第二把静默替换无提示 |
| 失败人话：某席失败标明席名，已成功席仍可复制其钥 | 静默半成功 |

## 契约

### Go（首刀·必做）

推荐单响应（避免 FE 循环 N×建成员×签 Key 半失败）：

```
POST /api/v1/workspaces/{ws}/teams/{teamID}/crew/batch
Body（可空）: { "seat_ids": ["…"] }   // 省略 = 全部 empty 席
→ 201 {
  "crew": [
    {
      "seat_id": "…",
      "role_slug": "frontend",
      "display_name": "前端",
      "member_id": "…",
      "token": "ymk_…",          // 明文仅此一次；跳过席无此字段
      "skipped": false,          // 已绑则 true，无 token
      "skip_reason": ""          // 人话，可选
    }
  ],
  "team_id": "…"
}
```

- Admin only；事务或尽力全成+部分失败明细。  
- **仅 empty 席**（body 显式 seat_ids 若含 active → 该席 skipped，人话「活跃席禁顶替」；**禁止** force）。  
- 每 empty 席：`INSERT members(is_bot=1, role=bot, display_name≈席名)` → `BindSeat` → `api_keys` 签发（同现网 `ymk_`+hash）。  
- 停工/离职席：本波 batch **默认不自动补位造新演员**（避免静默换人）；补位仍走人话路径/force 审计（JOIN_DECISION）。建队首跑只清 empty。  
- **禁止**改坏已有 `POST …/api-keys` 单钥契约。  
- 单测：默认班底 5 空席 → 5×member 绑上 + 5×token 前缀 `ymk_`；再调一次 → 全 skipped、无新 token。

若短期无法上 batch：允许 FE 编排现有 API，但 **产品验收仍认「人只点一次」**；Go 须尽快补 batch（标 B6）。

### FE

- 主 CTA 文案：「一键生成全班」（建队成功挡层 / 席页）。  
- 成功 → `BatchCrewRevealModal`（可复用/扩展 `IdentityKeyRevealModal` 为多节）：每席 `ymk_` + 复制；「复制全班钥匙包」；「复制本节启动提示词」走 `buildFirstRunPrompt`（可嵌该席 `ymk_`，醒目勿公开转发）。  
- 次操作：「复制全班入场包」→ 现有 AUTO_CREW（仅 `ymt_`）。  
- 关窗未复制 → 确认；roster 乐观刷新，席卡须立刻显示演员名（对齐 SEAT_VISIBILITY）。

### 全班钥匙包文案（示意）

```
# 全班身份钥匙包 · {团队名}
（含 N 把 ymk_，仅此一次，勿发公开群）

## {席显示名}（{slug}）
身份钥匙：ymk_…
（下附 FOOLPROOF 首跑启动提示词：安装 + API 根 + 本把 ymk_ + whoami）
```

## 验收

| ID | 标准 | 谁 |
| --- | --- | --- |
| **B0** | **任用/人用：** batch 只填 empty；active 不顶替；无「逐席填表」主路径；一键建队后挡层即全班就绪可复制包 | 产品裁定·全员 |
| **B1** | Go：`POST …/teams/{id}/crew/batch` 对 empty 造 Bot+绑+回 `token`；active→skipped；单测绿（含禁顶替） | Go |
| **B2** | 建队成功主 CTA「一键生成全班」；**一次点击**得全班包；UI **无**逐席新建 Bot/逐席填 Key 必经步骤 | React |
| **B3** | 挡层：每空席一节完整 `ymk_` + 可复制全班包；启动提示词可嵌该席 `ymk_`；醒目勿公开 | React |
| **B4** | 成功后席卡显示演员名（roster 已绑）；再点幂等跳过已绑席、不重发钥 | React+Go |
| **B5** | 同帧可取 AUTO_CREW 入场包（仅 `ymt_`，包内禁 `ymk_`） | React |
| **B6** | 箱截：`batch-crew-after-create.png`（挡层可见 ≥ 模板席数把 `ymk_` 或等价全班包） | React |
| **B7** | QA：默认班底 5 席一键；断言 5×绑定 + 5×`ymk_`；入场包路径仍禁嵌 `ymk_`；证 `docs/qa-shots` | QA |
| **B8** | 产品勾选关单 | 产品 |

## 派工

| 角色 | 职责 |
| --- | --- |
| **Go 后端** | **首刀 B1**（batch API + 单测）；与 Issue 域/Bot 记忆勿抢绑席语义，认现有 `BindSeat` |
| **React 前端** | B1 可联后立刻 **B2–B6**；B1 未合前可先做 Modal 骨架吃假数据，但 **勿宣称绿** |
| **QA** | B7（FE+Go 绿后） |
| **产品体验** | 本裁定；B8；AUTO_CREW C7 不挡本包开工 |

## 产品勾选

- [ ] **B0** 任用规则+禁逐席填表 · 产品
- [ ] **B1** batch API + 单测 · Go
- [ ] **B2** 建队/席页一键生成全班 · React
- [ ] **B3** 全班 `ymk_` 挡层 + 启动提示词 · React
- [ ] **B4** 席卡翻演员 + 幂等 · React+Go
- [ ] **B5** 同帧入场包（`ymt_` only）· React
- [ ] **B6** 箱截 · `batch-crew-after-create.png`
- [ ] **B7** QA
- [ ] **B8** 产品关单

## 进度注记

- 2026-10-24 23:33 · **人用习惯钉入**：任用规则（空席造人 / 活跃禁顶替 / 新人新席或空席 / 停工才补位）写入冻结与 **B0/B1**；禁逐席填表。AUTO_CREW 入场包已关。已派 Go+React 按本修订执行。
- 2026-10-24 23:31 · 定稿。AUTO_CREW 原 P1 升本债 P0；AUTO_CREW 入场包线并行（C7 在验不挡）。首刀 **Go B1**，随即 **React B2–B6**。

## 关单进度

**进行中** · 债文+任用钉入 @ 本修订；AUTO_CREW 已关；待 Go B1 → React B2–B6 → QA B7。
