# 人用入场全景（HUMAN_ONBOARD）· 可投表决稿

> 2026-10-24 23:45 Asia/Shanghai · 产品体验主持  
> **用户令（总指挥转）：** 可审稿尽快出；**每项做成可投票清单**（场景 **A–K** + 分歧点）。出稿后总指挥组织投票拍板 → 再派全员开干。  
> **已并入：** 总指挥洞文档简报 · Go redeem/bind 短条 · React FE 断点 · QA X0/E1–E7 · 记忆席 handoff/claim · BATCH_CREW `1f96809` 改向。  
> **刀状态：** 用户 `VOTE_ACCEPT_ALL_RECOMMENDED`（A–K+V1–V7 全认推荐默认）· **H0/H1 已冻** · **已开干**。

---

## 0. 已冻结（不必投 · 违约即打回）

| # | 冻结点 |
| --- | --- |
| F1 | 两码不合并：`ymt_`=入场码，`ymk_`=身份钥匙；禁混称、禁嵌 `ymk_` 进可转发入场包/`/join` URL |
| F2 | 任用硬规则：活跃禁顶替；新人只进空席或新开席；仅停工/离职可补位 |
| F3 | 禁逐席填表心智（Admin 建队主路径不进「高级·Bot」逐行新建/填 Key） |
| F4 | **否决**「一键吐全班 `ymk_`」/ `crew/batch` 回 N×token |
| F5 | Admin 建队主路径 **只发** 全班 `ymt_` 入场包（AUTO_CREW 已绿） |
| F6 | 入队无身份时 **自动签发单把** `ymk_` 仅回入队方（目标态；API 落点见 §3） |
| F7 | 续命权威 = `handoff get`（按 member）；换机换演员不换席池；扩编权在总指挥；Webhook 下发为主 |
| F8 | `ymw_`（人类 WS 邀请）与 Bot `ymt_` **分轨**，勿写进 Bot 续命叙事 |

---

## 0.1 用户票 H0 / H1（总指挥组织 · 本轮先投这两项）

| 用户票 | 含义 | 勾「过」= 认下列推荐默认 |
| --- | --- | --- |
| **H0** | 场景范围认账 | §1 **A–K** 全部按「推荐默认」列（有例外在投票回写写明） |
| **H1** | 主航线 + 关键分歧 | §3 目标态主航线 + 至少 **V1/V2/V3** 推荐默认（无钥重复拒造第二人；补位不自动；Web Join 只贴 ymt_） |

H0+H1 过 → 产品回写 §6 勾选 → 派全员开干。其余 V4–V7 / 场景备选可同轮附带或下轮。

## 1. 可投票 · 场景 A–K

每项：**推荐默认**（产品体验预填）+ **备选**。总指挥组织投「认推荐 / 改备选 / 搁置」。

| 票 | 场景 | 人用一句话 | 推荐默认（投这个即勾） | 备选 |
| --- | --- | --- | --- | --- |
| **A** | 空手上场 | 新人只持入场包行就能进 | redeem **无 Bearer** → 造 is_bot + mint 一把 `ymk_` + 绑 empty（或码上 role 空席）+ 当次回 `token`；再 handoff（空 focus 人话明示「空」，手册=`playbook_ref`） | 仍要求先设置签 Key 再 join（否决倾向，与用户令冲突） |
| **B** | 续队 | 已有钥再进同队 | 持有效 `ymk_` redeem/join → **不重发** token；幂等 ensure+合法席；不新建第二席/第二 member | 每次 join 刷新钥 |
| **C** | 补位 | 停工席换人 | **redeem 默认不自动补位换人**；Admin 标记停工后，新人显式选/带该席码再 bind（可 force 仅 left/stopped，留审计） | redeem 见 left 席自动补上（Go 倾向勿静默） |
| **D** | 分布式派发 | 一包多 Agent 各领一席 | **一席一码**（AUTO_CREW 现状）；并发无身份 join 靠 **per-seat 码 + 事务选席**，禁止同码无身份 N 次造 N 人抢席 | 一码进全队（契约大改，本波不推荐） |
| **E** | 换机 | 同身份新电脑续工 | **必须持原 member 的 `ymk_`**（或 migrate 同 member）；仅 `handoff get`；**禁止**新机无钥再 bootstrap 造第二 member | 允许新机无钥再 mint（记忆席：handoff 断档，不推荐） |
| **F** | 吊销 | 入场码作废 | 已有 J1 吊销；吊销后新 redeem 失败；**已发出的 `ymk_` 不随码吊销自动废**（身份另撤销走 api-keys） | 吊销码连带废该码发出的身份（复杂，搁置） |
| **G** | 误占 | 挤活跃席 | active → **硬 409** 人话「谁占用 / 如何 leave」；FE/通道同文案；错队错席 404/一行中文不串 roster | 提供「确认顶替」FE（否决） |
| **H** | 扩编 | 满员再加人 | 总指挥/Admin **新席 + 新 `ymt_`**；新人只进空席/新席 | 活跃席挤人扩编 |
| **I** | 设置签单钥 | Admin 主动造旁路 Key | **保留** FOOLPROOF 单把挡层；文案降为「旁路 / 救场」；**不是**建队主 CTA | 砍掉设置签 Key（过激，不推荐） |
| **J** | 多机同队 | 多 Agent 同班底 | 各持自己 `ymk_` + 各席码；单活 claim；同 member 双机须 leave→migrate（T2） | 同 member 多机并行无 leave（打架） |
| **K** | 救场 | 续命失败 | `export/import` 仅救场；主航线仍 pull→enter→handoff get；人话给路径不给假 focus | 日常依赖 handoff 文件 |

---

## 2. 可投票 · 分歧点（必须拍）

| 票 | 分歧 | 推荐默认 | 备选 | 依据摘要 |
| --- | --- | --- | --- | --- |
| **V1** | **无钥重复 join**（同 `ymt_` 无 Bearer 点两次） | **第二次起拒绝造人**：若该码已成功 bootstrap 过且明文钥已展示过 → 409/人话「请粘贴已发的身份钥匙续队」；**不做**同码+设备指纹静默认领（易错） | 同码+设备指纹 24h 内返回同一 member（不重发明文则无法续，除非会话缓存） | Go 边角；记忆双 member 坑 |
| **V2** | **补位是否自动** | **否**（=场景 C 推荐）：redeem 不静默换 left 席 | 是：redeem 自动绑合法补位席 | Go 倾向否；产品减意外 |
| **V3** | **Web Join 向导** | Go 自动身份后：**主路径只贴 `ymt_`**（取消强制先贴 `ymk_`）；若响应有 `token` → 单把 IdentityKeyReveal | Web 不做入场，只认 CLI/通道接钥；FE 只改文案 | React 痛点 A/B3 |
| **V4** | **入场提示词话术** | `buildBotJoinPrompt` 改两态：「空手：join 后自动得钥」/「已有钥：带 Bearer」；去掉「必须本机已有 ymk_」断头句 | 维持旧话术等通道全改 | React；与 F6 对齐 |
| **V5** | **建队 Toast/导语** | 热修：去掉「未绑 Key」主暗示；主文案「复制入场包，对方入场时自动获得身份」；设置签 Key 降次要 | 大改 UI 结构 | React 断点；可先于 API 合入 |
| **V6** | **并发同码无身份** | 入场包 **保持一席一码**；服务端选席事务 + 码维度限次；同码第二无身份 → 走 V1 | 允许同码多空席轮询分配 | Go 分布式边角 |
| **V7** | **BATCH B1 开干门禁** | 本表决 **A/B/E/G + V1/V2/V3** 勾完再派 Go 合 redeem 自动身份；此前半截不合 main | 先合 API 再改 FE | 总指挥开会收口 |

---

## 3. 目标态主航线（随票可微调文案，结构认 F1–F8）

### Admin
```
模板建队 → 复制全班入场包（仅 ymt_+role）→ 私发各 Agent → 完事
```

### Agent 空手（A + V3）
```
持 ymt_ → join/redeem（无 Bearer）→ 自动身份当次回 ymk_ → 合法绑席
  → 读 playbook_ref → handoff get（无 focus 明示空）→ 开干
```

### Agent 续队 / 换机（B/E）
```
持原 ymk_ → join 或不需新码 → handoff get；禁无钥 bootstrap 造第二人
```

### API 落点（Go 认 · 待 V7 后开刀）
- **主：** `POST /api/v1/team-invites/redeem`：无 Bearer→bootstrap；有 ymk_→幂等不重发；人类 JWT→只定位不造 Bot  
- **不做：** `crew/batch` 全班吐钥；不改坏 `POST …/api-keys`

---

## 4. 各角分工（投票后）

| 角色 | 投票后首刀 |
| --- | --- |
| **Go** | redeem 自动身份 + V1/V2/V6 单测；任用 409 |
| **React** | V3/V4/V5；Join 成功单把露钥（若投 V3 主路径） |
| **记忆** | 校验造人不踩 claim 单活；换机同 member；空 focus 人话 |
| **QA** | 用 §5 矩阵苛刻验；挂已有证 AUTO_CREW/FOOLPROOF/SEAT |
| **种子** | 文案两码对齐；禁免责腔 |
| **产品** | 维护本表决；勾选关刀 |

---

## 5. QA 矩阵（H3 · QA 加严回表 · 底图并 `e9d9414`）

| 用例 | 前置 | 步骤 | 期望 / PASS | 现况 |
| --- | --- | --- | --- | --- |
| **Q1 新入队自动身份** | 默认班底有 empty；Agent **无** ymk_ | `team join --code ymt_… --role`（可先 install.sh） | 成功；**响应回一把 ymk_**；席 occupied+Bot；whoami=该 Bot | **GAP**（目标态） |
| **Q2 续队不重发钥** | Q1 后持同 ymk_ | 再 join 同队 | **无新 token**；原席/合法席；memory/handoff 不丢 | 部分可验；自动身份后联验 |
| **Q3 活跃禁顶替** | 席 active=BotX | 另一身份/新人指定该席 join 或绑 Key | **409/人话**；席仍 BotX | 已有任用/409；须与 FE 文案一致 |
| **Q4 新人新席** | 无 empty（或满员） | join/扩编允许 expand | **新席**创建并入座；他席绑定不变 | 挂 allow_expand / SEAT_VIS |
| **Q5 补位** | 席 left/stopped/departed | 新人 join 补该席 | 入座成功；**审计 comment**；禁对 active force | GAP 细审计断言 |
| **Q6 Admin 入场包** | 建队成功（默认班底） | 复制全班入场包 | **N×ymt_+role**；正则 **无嵌 ymk_ 密钥** | **COVERED†** AUTO_CREW `37b2228` |
| **Q7 否决全班吐钥** | 建队/席页 | 巡 UI+相关 API | **无**「一键复制全班身份钥匙」；无 batch 回 N×ymk_ | **COVERED†** C7 + 债否决 |
| **Q8 分布式派发** | 席已绑 Bot+webhook | 指派 Issue 到该席 | **仅该席**唤醒；left 不喊 | 挂 Webhook/下发旧证；抽测 |
| **Q9 换机续命** | 同 ymk_ 新机 | handoff get / pull | 得 focus/todos/开放 Issue；**不必 Admin 再发钥** | 挂记忆 handoff；抽测 |
| **Q10 入场包≠身份** | 公开 `/join` 与可转发提示词 | 抽检正文/URL | **无可复制 ymk_**；自动身份只在**私密成功响应**/FOOLPROOF 单把 | 包侧 COVERED†；自动身份待 Q1 |
| **Q11 吊销** | 码已吊销 | redeem | 失败；旧 ymk_ 不随码自动废（认场景 F） | 挂 J1 |
| **Q12 无钥重复 join** | 同 ymt_ 已 bootstrap 过 | 再无 Bearer redeem | 认 **V1** 票（推荐：拒造第二人，人话贴旧钥） | GAP |
| **Q13 救场/双机** | 同 member 双机 | leave→migrate；或 export/import 救场 | 主航线非 import；双机须 leave | 挂 T2 |

**横切：** 两码永不合并；任用=活跃禁顶替+新人新席/空席+停工才补位。  
**旁路：** FOOLPROOF 设置单把 `a5299f5` 不挡主航线。  
**箱证文件（认）：** [`docs/qa-shots/HUMAN_ONBOARD-QA-CASES.md`](./qa-shots/HUMAN_ONBOARD-QA-CASES.md) @ QA `bf8bd8f`（Q1–Q10 补严；债内 Q11–Q13 投票后补进该文件）。

## 6. 产品勾选（总指挥投票结果回写）

### 场景
- [x] **A** 空手上场（推荐 / 备选：____） · **推荐默认** · `VOTE_ACCEPT_ALL_RECOMMENDED`
- [x] **B** 续队 · **推荐默认** · `VOTE_ACCEPT_ALL_RECOMMENDED`
- [x] **C** 补位 · **推荐默认** · `VOTE_ACCEPT_ALL_RECOMMENDED`
- [x] **D** 分布式 · **推荐默认** · `VOTE_ACCEPT_ALL_RECOMMENDED`
- [x] **E** 换机 · **推荐默认** · `VOTE_ACCEPT_ALL_RECOMMENDED`
- [x] **F** 吊销 · **推荐默认** · `VOTE_ACCEPT_ALL_RECOMMENDED`
- [x] **G** 误占 · **推荐默认** · `VOTE_ACCEPT_ALL_RECOMMENDED`
- [x] **H** 扩编 · **推荐默认** · `VOTE_ACCEPT_ALL_RECOMMENDED`
- [x] **I** 设置签单钥 · **推荐默认** · `VOTE_ACCEPT_ALL_RECOMMENDED`
- [x] **J** 多机同队 · **推荐默认** · `VOTE_ACCEPT_ALL_RECOMMENDED`
- [x] **K** 救场 · **推荐默认** · `VOTE_ACCEPT_ALL_RECOMMENDED`

### 分歧
- [x] **V1** 无钥重复 join · **推荐默认** · `VOTE_ACCEPT_ALL_RECOMMENDED`
- [x] **V2** 补位是否自动 · **推荐默认** · `VOTE_ACCEPT_ALL_RECOMMENDED`
- [x] **V3** Web Join 只贴 ymt_ · **推荐默认** · `VOTE_ACCEPT_ALL_RECOMMENDED`
- [x] **V4** 入场提示词两态 · **推荐默认** · `VOTE_ACCEPT_ALL_RECOMMENDED`
- [x] **V5** 建队文案热修（可先于 API） · **推荐默认** · `VOTE_ACCEPT_ALL_RECOMMENDED`
- [x] **V6** 并发同码策略 · **推荐默认** · `VOTE_ACCEPT_ALL_RECOMMENDED`
- [x] **V7** B1 开干门禁 · **推荐默认** · `VOTE_ACCEPT_ALL_RECOMMENDED`

### H4 文案锚点（种子贡 · 投票后修 README/BOT，禁免责腔）

| 锚 | 对外怎么说 |
| --- | --- |
| 人 | 注册 → 建工作区 → 浏览器用看板 |
| Bot | 领席位入场包（`ymt_`）→ **入场时自动得**身份（`ymk_`）→ 开干 |
| 三前缀 | `ymk_` 身份 · `ymt_` 队席入场 · `ymw_` 人类邀请 —— 首页一张表 |
| 旁路 | 设置里仍可单把签 `ymk_`（救场/高级），**不是**建队主路径 |
| `/join` | 落地说明页；真 redeem/得钥在 CLI/通道/向导成功响应 |
| seed | 演示 seed 可选；不写成唯一开工法 |
| 禁 | 「一码打天下」；「明确不做」清单腔；生产告警免责堆砌 |

**开干：** 用户 **H0+H1** 过 + §0 冻结 → 回写 §6 → 派 Go（自动身份）/ React（V3–V5）/ 种子（H4）。V5 文案热修可先于 API。

---
## 7. 附录 · 各角原文摘要（备查）

- **Go：** redeem 现只定位；自动身份主挂 redeem；危险=无钥重复造第二人、补位静默、换机无钥 bootstrap、同码并发。  
- **React：** JoinWizard 仍强制 ymk_；无成功面接 token；Toast「未绑 Key」打架；建议主 CTA=入场包、Web 只贴 ymt_。  
- **记忆：** 换机必须同 member；自动造人只 empty；已鉴权禁再造第二人；空 focus 明示。  
- **QA：** X0/E1–E7 骨架如上。  
- **总指挥洞文档：** 冻结要点=§0；开会收口=自动身份全链 + A–K + 生产 install（install 另票，本波不堵表决）。
- **种子：** 三前缀表缺失；「入场码」易读成一码通；BATCH 与「先 Key 再 join」旧叙事打架；`/join` 不 redeem 落差；人/Bot 顺序未写；设置签钥未标旁路；免责腔点名 README/T6_T7（投票后 H4 修）。
- **Go 附录（再确认）：** 与 §2–§3 一致；半截不合入；刀冻待 H0+H1。

## 8. 进度

- 2026-10-24 23:42 · **用户票通过** `VOTE_ACCEPT_ALL_RECOMMENDED`：A–K+V1–V7 全认推荐默认；H0/H1 冻；派 Go/React/记忆/QA/种子开刀。
- 2026-10-24 23:49 · H3 挂 QA 用例文件 `bf8bd8f`；用户票仍等 H0/H1。
- 2026-10-24 23:48 · 并 QA 加严 Q1–Q13、种子 H4 锚点、§0.1 用户票 H0/H1；待用户拍板。
- 2026-10-24 23:45 · **可投表决稿** 出；待总指挥组织投票。  
- 2026-10-24 23:40 · 首稿骨架 `6eaa4ff`。

## 9. 关单

**已冻已派** · H0/H1 通过；功能刀已按 §4 下发。
