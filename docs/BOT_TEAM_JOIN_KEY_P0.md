# Bot 空手上场 · 团队入场码（产品方案 P0）

> 产品体验 · 2026-10-24 · Asia/Shanghai  
> 领导层决策落地：空 Bot 装班底 → 粘贴后台「团队 Key」→ 自动 `pull` + `enter` + `handoff`；活跃席禁顶替。  
> 前置已关：[`BOT_TEAM_SEAT_P0.md`](./BOT_TEAM_SEAT_P0.md)（T8–T11）。本单补 **入场码产品契约 + FE 向导 + force 收窄**。

---

## 0. 一句话

新人只贴 **一串团队入场码**，系统定位工作区+团队、拉齐班底、进合法席位、还原 handoff；**绝不**挤掉活跃席。

---

## 1. 「团队 Key」是什么（结论）

| 名称（人话） | 技术形态 | 不是 |
|--------------|----------|------|
| **团队入场码** | 后台签发的 **opaque invite token**（建议前缀 `ymt_`），服务端兑换 → `{ workspace_id, team_id, template_ref?, expires_at?, max_uses? }` | ❌ 裸 `team` UUID（难记、易抄错、不含 workspace） |
| | | ❌ Bot API Key `ymk_…`（身份钥匙，一人一席） |
| | | ❌ 含完整 Key 的「交接卡」 |

**签发**：工作区 Admin · 团队页「复制入场码」→ 一次明文；库内存 hash；可吊销/过期。  
**权限**：入场码只授权 **定位团队 + pull roster + 申请 enter**；干活仍靠该 Bot 自己的 `ymk_`（向导内「本机 Bot Key」另贴或从 vault/env 读）。

**空手上场最小输入**：

1. 本机 Bot Key（`ymk_`，身份）  
2. 团队入场码（`ymt_`，定位班底）  
3. （可选）目标席位 slug；不填则走「新人规则」自动选席  

---

## 2. 安装向导步骤与 FE 面

### 2.1 通道（ymbot / MCP）— 主自动化

```
装班底手册（bot-team-template）
  → 配置 YMBOT_API_KEY + API_URL
  → ymbot team join --code <ymt_…> [--role <slug>]
       = resolve code → team pull → seat enter（合法席）→ handoff get
```

`team join` = 产品封装；底层仍是已有 `pull` / `enter` / `handoff get`，**不另造第二套 claim**。

### 2.2 FE 面（设置 · 席位）

| 步 | 面 | 行为 |
|----|----|------|
| 1 | **空态引导卡**（尚无席位绑定） | 「加入团队班底」主按钮 |
| 2 | **粘贴入场码** | 单输入框；校验格式；错误人话（过期/已吊销/非本工作区） |
| 3 | **选席 / 自动** | 展示班底：空席可选；活跃席灰显「占用中」；停工/离职席标「可补位」；无空席时「加新员工席」 |
| 4 | **一键进场** | 调 join；成功进席位抽屉 + handoff 摘要（同 T4） |
| 5 | **失败 409** | 专面：占用者末活时间 + 「请等对方 leave / 管理员标停工」；**无「强行顶替」按钮**（见 §3） |

换机已有席：仍走 T5「绑席位 → get」；入场码向导给 **空手上场 / 新人**，不替代迁绑。

---

## 3. 活跃 / 停工判定与 force

| 态 | 判定（产品） | enter 行为 |
|----|--------------|------------|
| **空席** | 无 `released_at IS NULL` 的 claim | 默认可 enter |
| **活跃** | claim 存在且 `released_at IS NULL`，且非管理员标「停工/离职」 | **一律 409**；UI/CLI **禁止顶替** |
| **停工 / 离职** | Admin 或本人 `leave`；或席位状态 `stopped`/`left`（显式标记） | 可 enter；CLI `--force` **仅此态**；须留 comment/审计 |
| **新人** | 无匹配空席且无合法补位席 | **只** `expand` 新席，禁止挤活跃 |

**force 收窄（相对现状 T10）**：

- **废除**：对活跃席的任何 force / 「确认顶替」产品入口（含 FE）。  
- **保留**：仅 `stopped|left` 席的 `--force`（通道）+ 管理端「标记停工」后的补位。  
- **P0.1**：force 成功必须写 Issue/席位 comment 审计（T10④ 已标补强）。

---

## 4. 与现有 seat API 的差

| 已有（T9–T11） | 本单要补 |
|----------------|----------|
| `POST …/teams/{id}/roster/ensure`（UUID 路径） | **`POST …/team-invites` 签发** + **`POST …/team-invites/redeem`（入场码→workspace+team）** |
| `seat enter` / claim 单活 409；force 仅停/离职 | **产品封装 `team join`**；FE 向导；**活跃席 force 入口清零** |
| `handoff get` | join 成功后 **自动** get（通道一步完成） |
| FE：席位条 / 交接摘要 / 换机文案（T1/T4/T5） | **空态「加入班底」向导**；入场码输入；选席/加新席；409 专面（可并 T3） |
| Admin 复制的是 UUID + 口头交代 | Admin **一键复制入场码**（可设过期/次数） |

**不改**：续命权威仍是 `handoff get`；模板仓仍是作业手册；`ymk_` 仍是身份。

---

## 5. 验收条（勾选）

| ID | 标准 | 勾选 |
|----|------|------|
| **J1** | 入场码 ≠ UUID ≠ `ymk_`；Admin 可签发/吊销；兑换得正确 workspace+team | [ ] |
| **J2** | 空 Bot：贴 `ymk_` + 入场码 → 一次 `join` 完成 pull+合法 enter+handoff；箱内自测有记录 | [ ] |
| **J3** | 活跃席第二路 enter → 409；FE/CLI **无**顶替按钮；新人无空席 → 只加新席 | [ ] |
| **J4** | 仅停工/离职席可 force/补位；成功留审计 comment | [ ] |
| **J5** | 重复 join / 重复 pull **幂等**，不清 memory/todos | [ ] |
| **J6** | 硬规矩：箱 5174/CLI 证据；禁甩用户本机；进度只回总指挥 | [ ] |

---

## 6. 分工建议（回总指挥调度）

| 谁 | 做什么 |
|----|--------|
| **Go / Bot 记忆** | invite 表+签发/兑换 API；`team join` 封装；force 活跃路径硬拒 |
| **通道** | `ymbot team join --code` + MCP `board_team_join` |
| **React** | 空态向导 + 入场码框 + 选席/新席 + 409 专面 |
| **产品体验** | 方案本页；交件后按 J1–J6 箱验收勾选 |

**开干条件**：总指挥确认 §1 入场码形态（`ymt_` opaque）后派工；未确认前 FE/Go 勿另发明「第三种 Key」。
