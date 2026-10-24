# 席位迁绑 · 技术对齐（接产品路径）

> 2026-10-24 · Bot 记忆角色 → 总指挥  
> 产品路径假定：**主航线** `team pull` → `seat enter` → `handoff get`（slug=席位）+ 团队共享记忆墙  
> 与 bot-team-template 同一套说明书。  
> 现成形状见 `models/bot.go` · 续命 H1–H5 已勾 · mirror P0 已落地

---

## 1. 现状与缺口（相对产品路径）

| 块 | 已有 | 缺口 |
|----|------|------|
| 续命 / 进场 | 主航线 team pull→seat enter→handoff get；whoami 可选；export 可选 | 说明书已定序；真服务 T9/T10 待合入证据 |
| mirror | 多 Key 只读冷备、失败跳过 | 未进 template；无「服务端不可达→读 latest」MCP |
| 席位/角色 | `bot_roles` + **单活** claim（`released_at IS NULL`）；handoff 带 `claimed_role` | 无「席位 seat」一等公民；Cursor agent↔member↔Key 映射在仓外；换机发/吊销 Key 流程无 API |
| 记忆 | `bot_memories` **归属 member**（可选 `role_id`）；UNIQUE(ws, member, key) | **无团队共享记忆墙表**；role_id 未当共享面用；跨项目/跨团队 list 未定 |
| 单活 | 角色 claim 409；Issue 叶子 claim 409 | 席位「人机各一」与双 Key 同角色策略未产品化 |

---

## 2. 席位绑定数据模型（建议）

**席位 seat** = 模板角色 **slug**（可被一台 Cursor Agent / 一套 Key 占用；旧称「角色坑」=席位）。

| 字段/关系 | 存哪 | 说明 |
|-----------|------|------|
| `role_slug` | `bot_roles.slug` | 稳定坑名（frontend/memory/…） |
| `member_id` | members（is_bot） | **身份真相**；API Key 绑此 |
| `claim` | `bot_role_claims` 单活 | 谁占坑；换机=新 Key 同 member **或** release+新 member claim |
| `agent_bind` | template 密钥匣（非 git） | `{workspace_id, role_slug, api_key_ref, member_id}` |
| 续命态 | 服务端 memories/todos/assigned issues | 经 `handoff get` 还原；**不**靠本机 |

**推荐迁绑语义（单活）**  
- **同 member 换 Key**（吊销旧、签发新）：handoff 自然续，claim 不动。  
- **换 member**（新 Bot 人）：`release` 旧 claim → 新 Key whoami → `role_claim` → 可选 `handoff import` 从旧 member 导出包（少用；优先共享墙+指派 Issue）。

---

## 3. API / MCP 草图

```
# 进场（主航线唯一；与 bot-team-template 同序）
board_team_pull { team_uuid }      # ① 拉班底；slug=席位
board_seat_enter { role, team_uuid? }  # ② 进席 bind|claim；409 勿盲抢
board_handoff_get                  # ③ 续命；无文件
# board_whoami 可选自检；migrate enter = seat enter

# CLI 同序
ymbot team pull --uuid <team>
ymbot seat enter --role <slug> [--team <uuid>]
ymbot handoff get
ymbot seat leave --role <slug>

# 共享记忆墙（P1，产品「团队墙」）
board_shared_memory_list { project_id? }
board_shared_memory_get/set        # 权限：ws 成员；与 bot_memories 隔离

# 旁路（仅救场）
读 mirror latest → board_handoff_import   # 仅救场
```

通道 DX：默认瘦、YM-n、get 含 version；共享 list 同样禁 UUID 刷屏。

---

## 4. 单活冲突

| 场景 | 行为 |
|------|------|
| 角色已被占 | claim **409** + 占用人 display；迁绑需显式 leave 或管理员 release |
| 双机同 Key | 服务端态共享（可接受）；禁止双机同时写同一 memory key 无协调（后写赢） |
| 双 Key 抢同 slug | 第二 claim 409；mirror 不参与仲裁 |
| Issue 领取 | 仍只叶子；与角色 claim **正交** |

原则：**席位单活在 DB claim**；不靠本机锁文件。

---

## 5. 多团队记忆隔离

| 层 | 归属 | 隔离键 | 用途 |
|----|------|--------|------|
| Bot 私有 | **Bot（member）** | `(workspace_id, member_id, key_name)` | 续命/焦点；进 handoff |
| 角色附注 | member + 可选 `role_id` | 现有字段，未当墙 | 可标记「占坑期间笔记」，release 不自动清 |
| **团队共享墙** | **工作区 / 项目**（拟） | `(workspace_id, project_id?, key)` + ACL | 多团队协作经验；**不进**默认 handoff-mirror |
| 看板 Wiki | 页面 | space/page | 长文；与短记忆墙分工 |

跨团队 = 跨 workspace → **另 Key / 另 seat**；禁止一个 Key 默认可写他队墙。  
list 走共享看板 API，不绑本机 git；项目管理主轨= Issues/Cycles。

---

## 6. 与 H1–H5 / mirror

- 主航线不变：`team pull` → `seat enter` → `handoff get`。  
- mirror：冷备私有 handoff；**共享墙另备或不进默认镜像**。  
- 席位迁绑：**先**同 member 换 Key；import/mirror 最后手段。

## 7. 三句收口

1. 身份=`Key→member`；席位=`slug` 单活（seat enter / claim）；续命=`handoff get`；班底=`team pull`。  
2. 共享记忆墙是 **ws/project 新面**，与 `bot_memories` 隔离，P1 表或先约定。  
3. 冲突用 409 + 显式 leave；mirror 不仲裁、不写回。
