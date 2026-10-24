# 入队/入席自动身份（BATCH_CREW）· P0

> 2026-10-24 23:35 Asia/Shanghai · 产品体验  
> **用户否决（总指挥转·Go 停 B1）：** **禁止**「一键吐全班 `ymk_`」——Admin 挡层一次复制 N 把身份钥匙 **不做**。  
> **改向：** **入队 / 入席自动身份**——人仍只点一次建队并拿走 **入场包**（[`AUTO_CREW`](./NEXT_WAVE_DEBT_AUTO_CREW.md) 已关）；**身份钥匙在 Agent 入队或入席时由服务端自动签发**，明文 **只回给该次入队方**，不进 Admin 全班钥匙包。  
> **仍生效的人用钉：** 任用规则（空席入座 / 活跃禁顶替 / 新人空席或新开席 / 仅停工离职可补位）；**禁止逐席填表**（禁 Admin 逐席「新建 Bot」「新建 Key」）。  
> **上位梳理：** [`HUMAN_ONBOARD`](./NEXT_WAVE_DEBT_HUMAN_ONBOARD.md)（2026-10-24 全景）；**代码刀冻结**至该债 H0+H1 勾完。本债保留自动身份 API 细案，开干以全景为准。  
> **废止：** `POST …/crew/batch` 回 N×`token`、`BatchCrewRevealModal` 全班 `ymk_` 挡层（`037a6ff` 前口径）。

## 一句话

Admin：建队 → **复制全班入场包（仅 `ymt_`）** 即完事。  
Agent：持入场码 join / 入席 → 若尚无身份，服务端 **自动造 Bot 成员 + 签发 `ymk_` + 按任用规则绑空席（或新席）** → **当次响应**把 `ymk_` 交给该 Agent（可写本机 / 首跑提示）；Admin **永不**收到全班明文钥。

## 与 AUTO_CREW / FOOLPROOF

| 包 | 谁拿什么 | 码 |
| --- | --- | --- |
| **AUTO_CREW（已关）** | Admin 一次拿走全班 **入场材料** | 仅 **`ymt_`** |
| **FOOLPROOF（已关）** | Admin **单把**在设置里主动签 Key 时仍可挡层露钥 | 单把 **`ymk_`**（设置路径保留，非建队主路径） |
| **BATCH_CREW（本债·改向）** | **入队/入席方** 自动得身份；建队主路径 **不** 吐全班钥 | join/bind 当次 **`ymk_`** |

**「一键建队 = 全班就绪可复制包」改读：** 就绪 = **全班入场包可发**（AUTO_CREW）；各席演员与身份在 Agent 入场时自动齐，**不是** Admin 剪贴板里 N 把 `ymk_`。

## 产品冻结

| 做 | 不做 |
| --- | --- |
| join / redeem+bind 主链：**无有效 Bot 身份时自动** `is_bot` 成员 + mint `ymk_` + 入合法席 | Admin「一键生成全班」吐 N×`ymk_` |
| 当次响应（或通道 join 输出）回 **一把** `ymk_` 给入队方；人话「仅此一次，勿转发」 | 把 `ymk_` 打进可转发入场 URL / 公开 `/join` 页 / Admin 入场包 |
| **任用硬规则：** 只进 empty；active → 409/人话禁顶替；无空席 → expand 新席（或显式申请）；补位仅 left/stopped/departed | 逐席 Form；挤 active；静默 force |
| 建队成功主 CTA 仍是 **复制全班入场包**（AUTO_CREW）；文案可补一句「对方入场时自动获得身份钥匙」 | 建队成功优先「一键造班吐钥」 |
| 幂等：已持有效 `ymk_` 的 join 不重发新钥；已绑席再 join 不换演员 | 每次 join 刷一把新钥 |

## 契约（行为）

### Go（首刀·改向后）

在 **Bot 入队主链**上补「自动身份」（名称可调，产品认行为）：

推荐挂在现有 join 封装附近，例如：

```
POST …/team-invites/redeem  或  通道等价 ymbot team join
（调用方：可无 Bearer / 或标记 bootstrap）

成功且需新身份时 → 201/200 增补：
{
  "workspace_id": "…",
  "team_id": "…",
  "seat_id": "…",
  "member_id": "…",
  "role_slug": "…",
  "token": "ymk_…",          // 明文仅此一次；已有身份则省略
  "token_issued": true
}
```

- 造成员 → 按任用选席 bind（empty / 新席 / 合法补位）→ 签发 `api_keys`（同现网 hash 契约）。  
- **禁止**改坏：已登录 Bot 持 `ymk_` 的 redeem/bind；`POST …/api-keys` Admin 单钥。  
- **禁止**实现/合并旧 `POST …/teams/{id}/crew/batch` 批量回 N×token（若已写半截：删或不开路由）。  
- 单测：① 无身份 + 有效 `ymt_` + 有 empty → 得 member+绑席+`ymk_`；② 再 join 同身份不重发 token；③ active 席目标 → 不顶替；④ Admin 入场包路径仍无 `ymk_`。

### FE / 通道

- **Admin：** 维持 AUTO_CREW；**删除/勿做** `BatchCrewRevealModal` 全班钥。设置里单把签 Key 的 FOOLPROOF 挡层可留。  
- **Agent 入场成功页 / 通道 join 输出：** 若响应含 `token`，**当次**展示/写入提示（单把，对齐 IdentityKeyReveal 语气）；勿汇聚多席。  
- 席页文案：空席「复制入场…」旁可注「入场方将自动获得身份钥匙」。

## 验收

| ID | 标准 | 谁 |
| --- | --- | --- |
| **B0** | **否决落实：** 无 Admin 全班 `ymk_` 包；任用规则仍硬；禁逐席填表 | 产品·全员 |
| **B1** | Go：入队/入席无身份时自动造成员+mint `ymk_`+合法绑席；当次回一把 `token`；单测绿（含禁顶替、幂等） | Go |
| **B2** | 建队主路径 CTA = 全班 **入场包**（AUTO_CREW）；无「一键吐全班钥」入口 | React |
| **B3** | Agent/通道 join 成功当次能拿到（或写入）自己的一把 `ymk_`；不进 Admin 剪贴板汇聚 | React / 通道 |
| **B4** | 入座后席卡显示演员名；再次 join 不刷新钥、不顶 active | Go+React |
| **B5** | Admin 入场包抽检：仍仅 `ymt_`、禁嵌 `ymk_` | QA / 产品 |
| **B6** | 箱证：`docs/qa-shots/` —— join 自动身份当次露钥（单把）+ 建队入场包无钥；文件名自定写入勾选 | QA / React |
| **B7** | 产品勾选关单 | 产品 |

## 派工

| 角色 | 职责 |
| --- | --- |
| **Go 后端** | **首刀 B1**（停旧 crew/batch 吐钥；做 join 自动身份） |
| **React 前端** | **停**全班钥挡层；B2 核对建队 CTA；B3 入场成功单把露钥（若 FE 有 join 成功面） |
| **QA** | B5/B6（Go+相关 FE 绿后） |
| **产品体验** | 本改向；B0/B7 |

## 产品勾选

- [x] **B0** 否决全班吐钥 + 任用 + 禁填表 · 产品（HUMAN_ONBOARD 冻）
- [x] **B1** 入队/入席自动身份 API + 单测 · Go · `c4ae0a5`
- [x] **B2** 建队 CTA=入场包，无全班钥入口 · React · V5 `e761305` + AUTO_CREW
- [ ] **B3** 入队方当次单把 `ymk_` · React/通道
- [ ] **B4** 席卡演员 + 幂等/禁顶替 · Go+React
- [ ] **B5** 入场包无 `ymk_` · QA
- [ ] **B6** 箱证
- [ ] **B7** 产品关单

## 进度注记

- 2026-10-25 00:00 · **B3 React V3/V4 冻干**：Join 只贴 ymt_ + token 露钥；提示词两态。V5 未回退。
- 2026-10-24 23:50 · **B1 绿** Go redeem `c4ae0a5`；B0/B2 随 HUMAN 冻+V5 勾。待 React V3/V4（B3）+ QA Q1（B5/B6）。
- 2026-10-24 23:45 · 上位 HUMAN_ONBOARD 已出**可投表决稿**（A–K + V1–V7）；本债 API 细案等投票后再开 B1。
- 2026-10-24 23:40 · **全景冻结**：盲改暂停，等 HUMAN_ONBOARD H0+H1。
- 2026-10-24 23:35 · **改向**：用户否决一键吐全班 `ymk_`；旧 B1 crew/batch 停。新 B1=入队/入席自动身份；Admin 只发 `ymt_` 入场包（AUTO_CREW）。已急停 React 全班钥实现。
- 2026-10-24 23:33 · （废）人用习惯钉入旧吐钥口径。
- 2026-10-24 23:31 · （废）原 P1 升档吐钥定稿。

## 关单进度

**B1 API 已合** `c4ae0a5` · 待 FE 入队露钥 + QA 箱证后 B7 关单。旧吐钥仍废。
