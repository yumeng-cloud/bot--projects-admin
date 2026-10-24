# Bot Team Template · 换机 / 角色迁移 / 多团队（脑暴 P0）

> 2026-10-24 · Asia/Shanghai · Bot 记忆角色 → 总指挥  
> 对接：本看板 `/workspace/ai-collab-board` + 拟写入 bot-team-template  
> 对齐：[`BOT_HANDOFF_CONTINUITY_P0.md`](./BOT_HANDOFF_CONTINUITY_P0.md) · [`BOT_HANDOFF_MIRROR_P0.md`](./BOT_HANDOFF_MIRROR_P0.md)

---

## 1. 换机镜像（旁路只读快照）

### 现状 / 缺口
- **已有**：`scripts/handoff-mirror.sh` + `BOT_HANDOFF_MIRROR_P0.md`（多 Key → 磁盘 `latest.json`；失败跳过单 Bot、不写服务端）。
- **缺口**：① 未进 bot-team-template 一键安装（cron/systemd 示例仅文档）；② 快照目录与「新机进场」剧本未产品化；③ 无健康探针/告警；④ Mac/他机 Path 与 Key 文件分发仍手工。

### 数据模型（旁路，非权威）
```
MIRROR_DIR/<bot_slug>/
  handoff-<ts>.json     # 全量 handoff 包（含 memories/todos/issues 快照）
  latest.json           # 软链/拷贝 → 最新
```
权威仍在服务端实时组装；磁盘仅冷备。

### API / MCP 草图
- **定时节点**：不调 MCP；CLI `handoff export --out` 或 REST `GET /bots/me` + `GET …/handoff`（现脚本已支持）。
- **进场只读（可选 MCP）**：`board_mirror_list` / `board_mirror_read`（读本地 `latest.json`）— **仅服务端不可达时**；默认工具描述禁止当日常续命。
- **失败策略**：单 Key 失败 log + continue；整轮非 0 仅当全灭；永不 `import`/`memory set`。

### 与 handoff-mirror 关系
本块 = **把已有 mirror 脚本收进 template**（env 模板、cron、进场 runbook）；不新开第二套备份。

### 风险
Key 落盘；快照含敏感记忆；时钟漂移导致「以为最新其实旧」→ 进场必须先试 `handoff get`，mirror 仅救场。

---

## 2. 角色迁移（换机后还原并绑定后台 Bot 身份）

### 现状 / 缺口
- **已有**：workspace `bot_roles` + claim；`bot_memories` / `bot_todos`；handoff GET/POST import；API Key 绑 `member_id`。
- **缺口**：① 「同一角色换机」= 换 Key/换 Cursor Agent，**身份在服务端 member**，本机无绑定状态机；② import 可 upsert memories/todos、尝试 claim，但 **role claim 冲突只 warning**；③ 无「template 角色 slug ↔ 看板 bot_role.slug ↔ Cursor agent」显式映射表；④ 新机发 Key / 吊销旧 Key 流程未产品化。

### 数据模型（建议，通道/配置层优先）
| 概念 | 存哪 | 说明 |
|------|------|------|
| `role_slug` | 看板 `bot_roles.slug` + template manifest | 如 `frontend` / `memory` |
| `member_id` | 看板 members（Bot 成员） | API Key 绑定此 id = 身份真相 |
| `agent_bind` | template 本地/密钥匣（非 git） | `{ role_slug, workspace_id, api_key_ref, member_id? }` |
| handoff 包 | 服务端组装 / 可选 mirror 文件 | memories + open todos + assigned issues |

**不绑本机 git**：映射与 Key 走环境/密钥匣；仓库只留 slug 清单。

### API / MCP 草图
```
# 进场（新机）
1. board_whoami                         # 校验 Key→member
2. board_handoff_get                    # 主路径还原记忆/todo/指派
3. board_role_claim { slug }            # 若岗位空闲则绑；409→读谁占用，不瞎抢
# 可选救场
4. board_handoff_import { from: mirror latest }  # 仅服务端空/灾难
```
CLI：`ymbot migrate enter --role frontend`（编排 whoami→handoff get→role claim；失败可读一行中文）。

### 与 handoff-mirror 关系
日常：**不靠** mirror。迁移失败或服务端空仓时，才用 `latest.json` import。Mirror 不负责 claim。

### 风险
双机同 Key / 双 Key 同角色 → claim 冲突；误 import 覆盖新记忆；吊销旧 Key 前新机未 whoami 通 → 空窗。

---

## 3. 多团队协作（共享看板 API，不绑本机 git）

### 现状 / 缺口
- **已有**：多成员 workspace、Issue/Cycle、Bot memory 按 `member_id`、语义检索（pgvector）、项目/权限 REST。
- **缺口**：① 「经验/上下文 list」仍偏 **单 Bot 私有 memory**，缺 **团队共享知识面**（project-scoped / role-scoped 可读记忆）；② 跨团队 = 跨 workspace 时无统一 template 发现 API；③ 项目管理仍可能滑向 GitHub Projects — 需明确 **本看板 Issues/Cycles = 唯一项目管理面**；④ bot-team-template 未定义「共享上下文包」schema。

### 数据模型（建议增量，勿先改库硬切）
| 层 | 模型 | 用途 |
|----|------|------|
| 项目 | 已有 `projects` + issues | 任务真相；禁 GitHub Projects 做主轨 |
| Bot 私有 | `bot_memories(member_id)` | 续命/焦点 |
| **共享经验（新）** | `team_knowledge` 或 memory + `scope=project\|workspace` + `visibility` | 可复用结论、约定、ADR 摘要 |
| 上下文 list | 通道投影：`{ id|slug, title, updated_at, scope }` 极瘦 | MCP 省 token |

通道层可先：**约定 memory key 前缀** `shared/` + 文档规范，P1 再表。

### API / MCP 草图
```
board_project_list / board_issue_list          # 已有；项目管理主面
board_context_list { scope: project|me|role } # 新：共享+私有上下文索引（瘦）
board_context_get { key|id }                  # 读一条
board_memory_set { key, value, scope? }       # scope=shared 需权限
board_handoff_get                             # 仍仅「我」的续命包
```
原则：**list/get 走共享看板 API**；本机 git 只放代码，不放任务态/经验库。

### 与 handoff-mirror 关系
Mirror 只冷备 **各 Bot 私有 handoff**；共享知识在服务端（或后续 team_knowledge），**不进**默认 mirror（避免密钥与私有记忆泄漏到共享盘）。若要备份共享面，另开 `knowledge-mirror`（只读、独立 ACL）。

### 风险
共享记忆污染 / 越权读写；list 过大吃 token（必须 DX 瘦字段）；与 Wiki 模块边界不清 → 约定：短经验→memory/knowledge，长文→Wiki。

---

## 4. 写入 bot-team-template 的建议切片

| 优先级 | 切片 | 依赖 |
|--------|------|------|
| P0 | mirror 脚本+cron+进场 runbook（已有，迁入 template） | 无 |
| P0 | `migrate enter` 编排 + role_slug manifest | 现有 claim/handoff |
| P1 | `board_context_*` + shared memory 约定或表 | 权限模型 |
| P1 | 明确「禁 GitHub Projects 主轨」写进 template README | 产品裁定 |

---

## 5. 给总指挥的三句话

1. **换机镜像**：旁路已有；template 收纳 + 进场先 get、mirror 仅救场、失败跳过。  
2. **角色迁移**：身份=`API Key→member`；进场 whoami→handoff get→role claim；import 非默认。  
3. **多团队**：任务与上下文 list 走看板 API；私有 handoff ≠ 共享经验；项目管理用本套 Issues/Cycles，不用 GitHub Projects。
