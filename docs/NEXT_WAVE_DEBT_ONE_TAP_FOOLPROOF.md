# 一键入场·傻瓜钥匙（ONE_TAP_FOOLPROOF）

> 2026-10-24 23:11 Asia/Shanghai · 产品体验  
> **用户定调（总指挥转·P0 怒）：** 一键入场仍傻逼——`ymt_` 入场码 vs `ymk_` Bot Key 两套心智；后台「创建 Bot」提示成功但**看不见钥匙/没数据**。要求傻瓜式：**创建成功立刻露出可复制的 `ymk_`**（不做合并「一码通」当 P0）；**复制提示词 / 入场零二次索钥**；**`ymt_` 只用于 join 团队**。  
> **承接：** [`NEXT_WAVE_DEBT_ONE_TAP_JOIN.md`](./NEXT_WAVE_DEBT_ONE_TAP_JOIN.md)（A/B/E 已绿；C/D 仍开）+ [`BOT_TEAM_JOIN_KEY_P0.md`](./BOT_TEAM_JOIN_KEY_P0.md)。本包**修创建露出 + 首跑提示词可带钥**，不重做 redeem/join。  
> **修正 ONE_TAP 旧安全口：** 「默认可转发提示词禁嵌 `ymk_`」仍成立；**仅「创建成功页 / 首跑启动提示词」（Admin→本机 Agent 私贴）允许嵌当次明文 `ymk_`**，并醒目警告勿转发公开群。  
> **2026-10-24 23:13 总指挥/Go 核完：** `POST …/api-keys` **已回明文 `token`（`ymk_` 仅一次）**。洞多半在 FE（成功未挡层展示 / 建成员与签 Key 拆两步未串联）。**优先 FE；能不改 API 就不动 Go。** 仅当「一键建 Bot」必须单响应时再让 Go 绑成员+mint（F1 后置）。

## 一句话

签 Key / 创建 Bot **成功立刻弹层挡住**：完整 `ymk_…` + 一键复制 + 「复制启动提示词」（安装 + API 根 + **本把钥匙** + whoami）。席页入场只带 `ymt_`，零二次索钥。

## 两码不合并（产品冻结）

| 前缀 | 人话名 | 用途 | 出现场景 |
| --- | --- | --- | --- |
| **`ymk_`** | **身份钥匙** | Bot 是谁（Bearer） | **仅**创建成功弹层 / 首跑启动提示词（私贴）；列表只留末四 |
| **`ymt_`** | **入场码** | 进哪个队（join） | 入场链接、席页复制提示词、`/join`；**禁止**当身份用 |

**不做（P0）：** 把两码合成「一码通」。

**现网痛点（须灭）：**

1. `POST /api-keys` 已回 `token`，但 FE 关 Modal 后明文只埋在设置·接入 Alert → **像成功没数据**。  
2. 建 Bot 成员与签 Key 若拆两步，中间无强制露钥。  
3. 席页/Join 仍二次要贴 `ymk_`。  
4. UI「Key/码」混称。

## 含 / 不含

| 做 | 不做 |
| --- | --- |
| **P0 FE**：凡拿到 `token` 的成功路径 → **挡层 Modal** 露完整 ymk_ + 复制 + 启动提示词 | P0 合并 ymk+ymt |
| 设置「新建 API Key」与席页/Bot 创建成功 **同一套成功组件** | 只 Toast「已新建」就关 |
| 首跑启动提示词可嵌当次 `ymk_`（私贴） | `ymk_` 写入可转发入场 URL / 公开 `/join` |
| 席页入场提示词只含 `ymt_` + 安装 + join | 强迫 CLI 再贴一次 ymk |
| **后置 Go F1**：仅当 FE 无法用现有 API 串出「一键建 Bot（成员+钥匙）」时，再加单响应绑成员+mint | 为露钥而改已绿的 api-keys 契约 |

## 契约

### 已有 API（认准 · 勿改）

```
POST /api/v1/workspaces/{ws}/api-keys
→ 201 { "api_key": {…}, "token": "ymk_…" }  // 明文仅此一次
```

FE 必须消费 `token`；**禁止**当响应无钥处理。

### 成功露出（FE 硬标准 · 首刀）

成功后 **Modal 挡住**（推荐，勿只靠页内 Alert）：

1. 标题：「请复制身份钥匙（仅此一次）」  
2. 完整 `ymk_…` + **复制**（主按钮）  
3. 次按钮：「复制启动提示词」  
4. 未复制关窗 → 确认；或本会话缓存到离页  
5. 刷新列表；**禁止**成功后名册/钥匙面空白

### 首跑启动提示词（可嵌 ymk_）

0. 安装（curl\|bash / go install）  
1. API 根  
2. **`export YMBOT_API_KEY=ymk_…`**（仅此模板）  
3. `ymbot whoami`  
4. 可选附已签发的 `ymt_` + join；否则写「入队另发入场链接」  
5. 警告：**含密钥，勿发公开群**

席页日常「复制 Bot 提示词」（ONE_TAP B2）**仍禁**嵌 ymk_。

### F1 · 一键建 Bot（Go · 后置 / 按需）

仅当 FE 确认：现网**没有**可用的「建 is_bot 成员」接口，或两步串联 UX 仍烂，才开：

```
POST …/bots  → 201 { member, token }   // 事务：成员 + mint；或 api-keys 无 bot 时自动建成员
```

产品验收认行为；路由名 Go 定。

### 文案

| 场景 | 正确叫法 | 禁止 |
| --- | --- | --- |
| 成功弹层 / 设置钥匙 | 身份钥匙（`ymk_`） | 「入场码」 |
| 席页 /join | 入场码 / 入场链接（`ymt_`） | 「API Key」当主名 |

## 验收小包

| ID | 标准 | 谁 | 优先级 |
| --- | --- | --- | --- |
| **F2** | 成功弹层：完整钥匙 + 复制；箱截 `foolproof-create-bot-key.png`（设置新建 Key 与/或 Bot 创建成功） | React | **P0 首刀** |
| **F4** | 设置「新建 API Key」消费已有 `token`，同套挡层；禁只 Toast + 埋 Alert | React | **P0 首刀** |
| **F3** | 「复制启动提示词」含安装 + API 根 + 当次 ymk_ + whoami + 勿转发警告 | React | P0 紧跟 |
| **F5** | 席页入场只 `ymt_`，零二次索钥；文案分家 | React | P0 紧跟 |
| **F1** | 一键创建 Bot 单响应 member+token；**仅 FE 要单响应时开** | Go | **后置按需** |
| **F6** | 箱证：露钥 → 启动提示词 → whoami PASS | QA + 产品 | FE 绿后 |

## 派工

| 角色 | 职责 |
| --- | --- |
| **React 前端** | **首刀 F2+F4 → F3 → F5**；吃现有 `token`；箱截；测绿 push |
| **Go 后端** | **默认不动**；等 FE 反馈是否要 F1 单响应再干 |
| **QA** | F6（FE 绿后） |
| **产品体验** | 本裁定；目视勾选 |

## 产品勾选

- [x] **F2** 成功弹层露钥+复制 · React `681de29` · `foolproof-create-bot-key.png`
- [x] **F4** 设置新建 Key 挡层消费 token · IdentityKeyRevealModal
- [x] **F3** 启动提示词嵌钥 · `buildFirstRunPrompt`
- [x] **F5** 入场零二次索钥 + 文案分家 · 席页仍只 ymt_；弹层文案两码分家
- [~] **F1** （按需）一键建 Bot API · **本波不挡关**（后置；现 api-keys+挡层已满足露钥）
- [x] **F6** 箱证 whoami 串联 · QA PASS† `a5299f5` / `FOOLPROOF-F6-qa-retest-20261024.txt`

## 进度注记

- 2026-10-24 23:28 · **P0 关单** QA F6 PASS† `a5299f5`（创建→启动提示词嵌钥→CLI→whoami=同 Bot；Modal 实地点击 soft†）。F1 仍后置不挡关。
- 2026-10-24 23:22 · **产品勾 F2–F5** React `681de29`：挡层露 ymk_+复制+启动提示词；席页入场仍只 ymt_。F1 仍后置；**F6 待 QA**。
- 2026-10-24 23:14 · 并行另开 [`NEXT_WAVE_DEBT_SEAT_VISIBILITY.md`](./NEXT_WAVE_DEBT_SEAT_VISIBILITY.md)：建队 Toast 有 N 空席但页仍「尚未绑定席位」——空席必须可见。

- 2026-10-24 23:13 · 总指挥/Go：api-keys 已回 token；**首刀改 React F2+F4**；Go F1 后置按需。  
- 2026-10-24 23:11 · 初稿。两码不合并；首跑可嵌 ymk_。

## 关单进度

**已关（P0）** · React `681de29` + QA F6 PASS† `a5299f5`；F1 后置不挡。
