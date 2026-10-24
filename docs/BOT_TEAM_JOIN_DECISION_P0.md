# 班底入场 · 分布式开工 — 拍板稿

> 产品体验定稿 · 2026-10-24 · Asia/Shanghai  
> 供用户拍板。对齐现网：`roster/ensure` · `seats/bind|unbind` · `handoff` · workspace `webhooks`（T9–T11 已关）。  
> 详案附录：[`BOT_TEAM_JOIN_KEY_P0.md`](./BOT_TEAM_JOIN_KEY_P0.md)。

---

## 1. 目标场景（拍板后即此）

| # | 场景 | 结果 |
|---|------|------|
| A | **空 Bot 进队** | 贴入场码 → 班底齐、进合法席、handoff 还原，开干 |
| B | **总指挥扩编** | 默认由其发码/定席；后台可自签备用 |
| C | **远程机** | 与本机同一 join；分布式同队同真相源 |
| D | **下发开工** | 总指挥任务进看板 → 推到对应席位 Bot **被动**开工 |
| E | **不打架** | 活跃席禁顶替；新人只进空席或 **新开席**；仅停工/离职可补位 |

---

## 2. 对象模型

| 对象 | 是什么 | 标识 |
|------|--------|------|
| **工作区** | 看板租户 | `workspace_id` |
| **团队** | 一班角色席位的容器 | `team_id`（UUID） |
| **席位** | 角色坑位（slug + 显示名 + 可选 playbook） | `team_seats`；可绑 `member` |
| **Bot 身份** | 演员（Key 持有者） | `ymk_…` → `member_id` |
| **入场码** | 定位「进哪个工作区·团队」的邀请 | **`ymt_…`** opaque；兑换 → `{workspace_id, team_id}`；可过期/吊销/限次 |
| **停工/离职** | 席位可补位态 | P0：**已 unbind**（`bound_member_id` 空）且对应 role claim 已 `released`；P0.1 可加显式 `duty_status` |
| **活跃** | 已绑 member 且 claim 未 release | **禁止**他人 bind/force |

**两把钥匙**：`ymk_` = 我是谁；`ymt_` = 我进哪队。禁止混用、禁止用裸 UUID 当主输入。

---

## 3. 主路径交互

### 3.1 空手上场（丝滑一条链）

```
装班底手册
  → 配置本机 ymk_ + API
  → 粘贴 ymt_ 入场码（可选目标 slug）
  → team join
       ① redeem 入场码
       ② roster/ensure（幂等，不清 memory）
       ③ 选席：空席 bind / 无空席则「加新席」再 bind / 停工席可补
       ④ handoff get → 恢复 focus / todo / 开放 Issue
```

**FE**：设置 · 席位空态「加入班底」→ 单框贴码 → 班底列表（空席可选 / 活跃灰显占用中 / 无空席「申请新席」）→ 成功进交接摘要。  
**409**：只说明谁占用、如何 leave；**无「强行顶替」**。

### 3.2 扩编（权在总指挥）

- 总指挥（本机或后台）：签发/转发 `ymt_`、指定 slug 或允许加新席。  
- Admin 团队页「复制入场码」= 自助备用，不替代调配默认。  
- 加新人 = **空席或新席**；不是挤掉活跃演员。

### 3.3 下发唤醒（被动开工）

| 优先级 | 机制 | 行为 |
|--------|------|------|
| **主** | 看板 **Webhook** | `issue.created` / `issue.updated`（含 assignee 变更）→ 席位例程 URL → Bot 醒 → `handoff get` → 领指派叶子 |
| **辅** | 轻量 **巡检**（5–15min） | 拉「指派给我且未读」；**无变化安静** |
| **禁** | 对已 leave 旧 Bot / 猜角色名乱叫醒 | 只打名册 active 席绑定 |

远程机：Webhook 打到该席登记的唤醒 URL；离线靠巡检补漏，单不丢。

---

## 4. API 切片（相对现网）

### 已有（复用，不另起炉灶）

| 能力 | 路径 |
|------|------|
| 拉班底 | `POST …/teams/{teamID}/roster/ensure` |
| 进/离席 | `POST …/seats/{seatID}/bind` · `unbind` |
| 续命 | `GET …/bots/{me}/handoff` |
| Webhook CRUD | `…/webhooks` + `DispatchEvent` |

### 本单净增

| 能力 | 建议 | 说明 |
|------|------|------|
| 签发入场码 | `POST …/teams/{teamID}/invites` → 一次明文 `ymt_` | 总指挥/Admin；hash 入库 |
| 兑换 | `POST …/team-invites/redeem` `{code}` → `{workspace_id,team_id,roster?}` | Bot 持 `ymk_` 调用 |
| 一键 join | 通道 `ymbot team join --code`（= redeem+ensure+bind+handoff get） | 产品封装，非第二套 claim |
| **加新席** | `ensure` 带新 `role_slug`，或 `POST …/seats` | **≠** 现网 `expand`（expand=给空闲席派叶子 Issue） |
| force 收窄 | `bind(force)`：**仅**目标席已 unbind/停工；活跃 → 硬 409 | 现网 force 可顶活跃，**产品废该行为** |
| Webhook 接通 | Issue 创建/更新/指派 **必须** `DispatchEvent` | 现网有投递器，缺事件接线则本单补 |
| 事件名 | 至少 `issue.created`、`issue.updated`（payload 含 `assignee_id`） | 席位例程按 assignee=本席 member 过滤 |

---

## 5. 不做清单

- 不用裸 `team_id` UUID 当用户主输入  
- 不把 `ymk_` 叫成「团队 Key」  
- 不做活跃席「确认顶替」FE/CLI  
- 不静默多活同一席写回  
- 不把 `expand`（派活）当成「加新员工席」  
- join **不**清空 memory/todos  
- 日常续工不依赖 handoff 文件；export/import 仅救场  
- 过程中不 fan-out 吵用户  

---

## 6. 验收（勾选即关）

| ID | 标准 |
|----|------|
| **D1** | Admin/总指挥可签发 `ymt_`；redeem 得正确 workspace+team；可吊销 |
| **D2** | 空 Bot：`ymk_`+`ymt_` → join 完成 ensure+合法 bind+handoff；箱内证据 |
| **D3** | 活跃席第二 bind → 409；无顶替入口；force 对活跃硬拒 |
| **D4** | 无空席时「加新席」成功且不改他人绑定；停工（已 unbind）可补位 |
| **D5** | 指派 Issue → Webhook 投递成功（或巡检补到）；无变化巡检安静 |
| **D6** | 远程机同一入场码进同队；记忆按席位隔离不串 |
| **D7** | 硬规矩：箱测过关；禁半成品甩用户本机 |

---

## 7. 风险与收口

| 风险 | 收口 |
|------|------|
| 现网 `bind(force=true)` 仍可顶活跃 | Go：**改语义**为仅空/已解绑；单测改写；通道去掉活跃 force |
| `DispatchEvent` 未挂 Issue 写路径 | Go：创建/更新/指派处接线；补投递测试 |
| 「expand」词义混乱 | 文案：**加新席** vs **派活**；CLI 保持 `expand`=派活，加席走 ensure/seats |
| 停工无显式字段 | P0 用 unbind=可补；P0.1 `duty_status` |
| 入场码泄露 | 短 TTL、限次、可吊销；日志只留末四位 |
| 双 Key 认知成本 | 向导文案固定：「身份钥匙 / 团队入场码」 |

---

## 8. 拍板问句（请用户选）

1. **入场码形态**：确认 `ymt_` opaque（推荐）— 还是坚持只贴 team UUID？  
2. **扩编默认**：确认总指挥调配为主、Admin 自签为辅？  
3. **唤醒**：确认 Webhook 主 + 巡检辅？  
4. **force**：确认活跃席彻底禁止顶替（含管理员 API）？

**推荐全选「确认」**。确认后派：Go（invite + force 收窄 + webhook 接线）→ 通道 `team join` → React 空态向导 → 产品体验按 D1–D7 箱验。

---

*本页为唯一拍板面；实施以本节为准，JOIN_KEY 为展开说明。*
