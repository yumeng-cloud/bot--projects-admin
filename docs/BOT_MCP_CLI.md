# Bot 工作流 · MCP / CLI 最小能力清单

> 产品体验规格（可执行）。**不改业务代码**；落地由「Bot 记忆角色」+ Go 后端分工。  
> 对照：[`API.md`](./API.md) · [`BOT.md`](./BOT.md) · [`BOT_CONTINUITY_P0.md`](./BOT_CONTINUITY_P0.md) · [`BOT_HANDOFF_CONTINUITY_P0.md`](./BOT_HANDOFF_CONTINUITY_P0.md) · 根目录 [`COORDINATION.md`](../COORDINATION.md)

---

## 1. 目标用户与场景

| 用户 | 场景 | 期望路径 |
|------|------|----------|
| Cursor / IDE 助手 | 在仓库里接看板：拉指派 Issue、写评论、干活中 upsert 记忆；续工用 handoff get（无文件） | **MCP tools**（自然语言 → 结构化 tool call） |
| 终端脚本 / CI / 本地 Agent | 一条命令评论、改状态、续工导入；可管道化 | **CLI**（`ymbot …`） |
| 人类 Admin | 签发/吊销 Bot API Key；不在本清单实现 UI | 现有 REST + 设置页（Integrations / Settings） |

**Bot 典型一天（续命 = 服务端；见 [`BOT_HANDOFF_CONTINUITY_P0.md`](./BOT_HANDOFF_CONTINUITY_P0.md) / [`BOT_CONTINUITY_P0.md`](./BOT_CONTINUITY_P0.md)）：**

1. 用 `ymk_…` 鉴权 → **`whoami`** 确认 `member_id` + workspace  
2. **`handoff get`（无文件）** 恢复 memories + todos + 开放 Issues  
3. `issue list --assignee me` → 读详情 → 评论 / 更新属性（**不得**自动 `done`）；叶子用 `issue claim`  
4. **过程中写回服务端**：`memory set` / `todo` / `comment`（claim 后立刻记进度）；**禁止**攒到会话结束才 export  
5. 换会话 / 新进程：再 `whoami` → `handoff get` 即可。`export`/`import` **仅可选备份/救场**，非主路径

---

## 2. 通道决策（推荐）

**决策：MCP + CLI 都要；同一套「薄客户端」调现有 REST，先交付 MCP P0，CLI 同周跟进或共享库并行。**

| | MCP | CLI |
|--|-----|-----|
| 主用户 | Cursor / 助手连看板 | 脚本、终端、非 MCP 宿主 |
| 形态 | MCP Server（stdio 优先；可选 SSE） | 二进制 `ymbot`（或 `go run ./cmd/ymbot`） |
| 理由 | 助手原生 tool calling，路径最短 | 可脚本、可 CI、可调试；不绑某一 IDE |
| 实现约束 | Tool schema ≈ CLI 子命令；**禁止**在通道层重写业务逻辑 | 共用 `internal/botclient`（或 `pkg/boardclient`）HTTP 客户端 |

**不推荐「只做一个」的原因：** README 将完整 MCP 标为超出范围，但总指挥要求「好用路径」；仅 MCP 则脚本/无人值守弱，仅 CLI 则 Cursor 助手要手写 curl。双通道、单客户端是最小长期成本。

**分期：**

- **P0**：MCP tools（下表 P0）+ CLI 同等子命令（可先薄封装）  
- **P1**：搜索、Cycle、todo 细粒度、角色 claim/release 的独立 tool（handoff 已覆盖续工主路径）

---

## 3. 现状 vs 缺口（扫库结论）

| 域 | 现有 REST（已挂载） | MCP/CLI |
|----|---------------------|---------|
| 鉴权 | `AuthJWTOrAPIKey`；`Bearer ymk_…` → `actor_kind=bot`，绑定 `member_id` + `workspace_id`；Admin `…/api-keys`；**已有** `GET …/bots/me` | **P0 已落地**：`ymbot whoami` / `board_whoami`；仍无 fine-grained scope 列（= 工作区成员权限） |
| Issue | `GET/POST …/issues`；`GET/PATCH/DELETE …/issues/{id}`；filter：`project_id,status,assignee_id,label_id,cycle_id,q`；PATCH 需 `version` | **P0 已落地**：`ymbot issue *` / `board_issue_*` |
| 评论 | `GET/POST/PATCH/DELETE …/issues/{id}/comments` | **P0 已落地**：`ymbot comment *` / `board_comment_*` |
| 记忆 | `GET/POST/PATCH/DELETE …/bots/{memberID}/memories`（POST upsert by `key_name`） | **P0 已落地**：`ymbot memory *` / `board_memory_*` |
| Handoff | `GET/POST …/bots/{memberID}/handoff`（无独立表，实时组装 / 导入） | **P0 已落地**：`ymbot handoff get\|export\|import` / `board_handoff_*`（含 `--out` / `--file`） |
| 搜索 / Cycle | `GET …/search`；`…/cycles` CRUD 子集 | **缺口**：通道封装（**P1**；产品侧用现有 Cycle 页即可，**不**要求 GitHub Projects 时间线） |
| 幂等 | Workspace 路由已套 `Idempotency-Key` | 通道层已支持：CLI `--idempotency-key` / MCP `idempotency_key` / env `YMBOT_IDEMPOTENCY_KEY`，缺省自动 UUID |

> **文档标注（2026-10-24 验收）**：P0 MCP/CLI 已在 `backend/cmd/ymbot` + `backend/internal/botclient` 落地；下文 §7 验收项仍须用真 Key 串联跑通。证据：`cmd/ymbot/{main,cli,mcp}.go`、`internal/botclient/client.go`、`handlers/bots.go`（含 `bots/me`）。

---

## 4. 最小能力清单

命名约定：

- MCP tool：`board_<verb>_<noun>`（短、稳定）  
- CLI：`ymbot <noun> <verb>`（与 tool 一一对应）  
- 默认从环境读：`YMBOT_API_URL`、`YMBOT_API_KEY`、`YMBOT_WORKSPACE_ID`；可选 `YMBOT_MEMBER_ID`（缺省用 Key 绑定成员）

| 能力名 | 通道（MCP tool / CLI） | 对应现有 REST 或缺口 | 输入 / 输出要点 | 优先级 |
|--------|------------------------|----------------------|-----------------|--------|
| 鉴权自检 | `board_whoami` / `ymbot whoami` | **已落地** `GET …/bots/me`（JWT 或 `ymk_`）；不必再调 JWT-only `/auth/me` | 出：`workspace_id, member_id, actor_kind, role, display_name` | **P0** |
| Issue 列表 | `board_issue_list` / `ymbot issue list` | `GET …/issues?…` | 入：`status, assignee_id\|me, project_id, q, cycle_id, label_id, leaf_only?`；出：`issues[]`；`--leaf-only` 拉可领池（依赖后端） | **P0** |
| Issue 读取 | `board_issue_get` / `ymbot issue get` | `GET …/issues/{id}` | 入：`issue_id`（或 identifier 解析，P1）；出：完整 issue + labels | **P0** |
| Issue 创建 | `board_issue_create` / `ymbot issue create` | `POST …/issues` | 入：`project_id, title, **parent_id（必填）**, …`；**禁无父游离**（通道 403 中文）；禁 `done` | **P0** |
| Issue 更新 | `board_issue_update` / `ymbot issue update` | `PATCH …/issues/{id}` | 入：`issue_id, version`（必填）+ 可变字段；409 → 可读冲突；Bot 禁 `done` | **P0** |
| Issue 叶子领取 | `board_issue_claim` / `ymbot issue claim` | `PATCH …/issues/{id}`（`assignee_id=me`） | **只领叶子**；≠ `board_role_claim`（岗位）；非叶子依赖服务端 409 | **P0** |
| 评论读取 | `board_comment_list` / `ymbot comment list` | `GET …/issues/{id}/comments` | 入：`issue_id`；出：`comments[]` | **P0** |
| 评论写入 | `board_comment_create` / `ymbot comment create` | `POST …/issues/{id}/comments` | 入：`issue_id, body`；一条命令完成；幂等 Key | **P0** |
| 记忆读取 | `board_memory_list` / `ymbot memory list` | `GET …/bots/{memberID}/memories?q=` | 入：`q?`；默认当前 Bot member；出：`memories[]` | **P0** |
| 记忆写入 | `board_memory_set` / `ymbot memory set` | `POST …/bots/{memberID}/memories`（upsert） | 入：`key_name, value_text, tags?`；出：`memory` | **P0** |
| Handoff 续工/备份 | `board_handoff_get` / `ymbot handoff get\|export` | `GET …/bots/{memberID}/handoff` | **日常续工**：`get` 无文件；`--out`/`export` 仅换机备份可选 | **P0** |
| Handoff 导入 | `board_handoff_import` / `ymbot handoff import` | `POST …/bots/{memberID}/handoff` | 入：文件或 JSON；出：`imported_* / claimed / warnings` | **P0** |
| 工作区搜索 | `board_search` / `ymbot search` | `GET …/search?q=` | 入：`q`；出：issues/projects 摘要 | P1 |
| Cycle 列表 | `board_cycle_list` / `ymbot cycle list` | `GET …/cycles` 或 `…/projects/{id}/cycles` | 入：`project_id?` | P1 |
| 角色认领（岗位） | `board_role_claim` / `ymbot role claim` | `POST …/bot-roles/{id}/claim` | **≠** `board_issue_claim`；入：`role_id`/`slug`；409 冲突可读 | P1（handoff 可部分覆盖） |
| 角色释放 | `board_role_release` / `ymbot role release` | `POST …/bot-roles/{id}/release` | 入：`role_id` | P1 |
| Todo 列表/写 | `board_todo_*` / `ymbot todo …` | `…/bots/{id}/todos` | 与记忆同模式 | P1 |

P0 合计 **12** 项能力（鉴权自检 + Issue×5含叶子 claim + 评论×2 + 记忆×2 + handoff×2）；排期原则：只领叶子、禁自创游离（§4.5）。

---

## 5. 鉴权模型（Bot API Key 最小安全约定）

| 项 | 约定 |
|----|------|
| 形态 | 明文前缀 `ymk_` + 随机；库内 **仅存** `sha256(key)`（`api_keys.key_hash`） |
| 绑定 | `(workspace_id, member_id)`；请求进入后 `actor_kind=bot` |
| 签发 | 仅 **Admin 人类 JWT**：`POST /api/v1/workspaces/{ws}/api-keys`，body `{name, member_id?}`；**明文仅创建响应返回一次** |
| 吊销 | `DELETE …/api-keys/{keyID}` 或 `POST …/revoke`；吊销后立即 401 |
| Scope（最小） | **无独立 scope 表**：Key = 该 Bot 成员在工作区内的成员权限。禁止用 Bot Key 管理 API Key / 邀请成员（现有 `ensureAdminOf` / Admin 中间件已挡） |
| 存放 | 人类：密码管理器 / CI Secret。Agent：环境变量 `YMBOT_API_KEY` 或 MCP server env；**禁止**写入 git、日志、handoff JSON、记忆 value |
| 请求头 | `Authorization: Bearer ymk_…`；写操作另加 `Idempotency-Key: <uuid>` |
| 示例 | `curl -H "Authorization: Bearer ymk_…" -H "Idempotency-Key: $(uuidgen)" …` |

MCP / CLI **不得**实现登录密码流；只接受 API Key（人类调试可用 JWT，但不作为 Bot 路径文档主推）。

---

## 6. 「好用」体验标准

1. **一条命令完成评论**：`ymbot comment create --issue YM-12 -m "已复现"`（MCP 等价单次 tool call）。  
2. **错误可读**：通道层把 `{"code","message"}` 打成一行中文 + 退出码（401/403/404/409）；409 提示带上当前 `version`。  
3. **写默认幂等**：优先 CLI `--idempotency-key` / MCP `idempotency_key`，其次 `YMBOT_IDEMPOTENCY_KEY`，否则自动 UUID；相同 Key 重放不双写。  
4. **`me` 语义**：`--assignee me` / 缺省 member = Key 绑定成员；handoff/memory 默认操作自己。  
5. **Bot 护栏透传**：尝试 `status=done` 时通道提前拒绝或原样返回后端错误，文案明确「需人类确认」。
6. **只领叶子**：用 `issue claim` / `board_issue_claim`；禁教 Bot 自创无父游离 Issue；角色 claim ≠ Issue claim。
7. **通道 DX（省 token）**：Issue 用 **`YM-n`（identifier）**；列表默认瘦字段；CLI/MCP **默认文本友好**；默认禁 32 位 hex；**`--id` / `ids=true`** 附加裸 id；**`--full` / `full=true`** 才接近 REST 全量。权威验收与勾选：**[`BOT_CHANNEL_DX_P0.md`](./BOT_CHANNEL_DX_P0.md)**（D1–D6 已勾关）。

---

## 7. 验收清单

- [x] 仅用 `YMBOT_API_KEY`（无浏览器 cookie）完成：whoami → handoff get → issue list(me) → comment create → memory set → handoff export（2026-10-24 本机真 Key）  
- [x] 相同 `Idempotency-Key` 重放 `comment create` 不产生第二条（CLI `--idempotency-key` 与 env `YMBOT_IDEMPOTENCY_KEY` 均验过）  
- [x] Bot 将 Issue 设为 `done` 失败（通道护栏中文一行 + exit 11）；人类 JWT 路径未改  
- [x] 吊销 Key 后下一请求 401；CLI 文案含「已吊销」；无效 Key 含「无效」（exit 10）  
- [ ] `handoff import` 后 memories upsert、未完成 todos 出现；claim 冲突仅 warning 不整体失败  
- [x] MCP：Cursor（或 MCP Inspector）接入步骤与冒烟 → **§11.3**（运维用真 Key 自跑；文档缺口已补）  
- [x] CLI：`--help` / `handoff import` 说明 → **§11.1–11.2**；非 TTY 下 JSON 可 `jq`（`--json`）  
- [ ] 密钥未出现在 handoff 文件、命令行历史示例文档用占位符 `ymk_***`  
- [ ] 本文件与 `BOT.md` / `COORDINATION.md` 回链一致；**未**改业务 handler 语义（仅允许为 `bots/me` 等小缺口追加路由，若做）

---

## 8. 给「Bot 记忆角色」的落地拆分

| 归属 | 事项 |
|------|------|
| **Bot 记忆角色** | ① MCP Server 脚手架（stdio）+ P0 tools 映射到 REST；② CLI `ymbot` P0 子命令；③ 共享 HTTP 客户端（鉴权头、幂等、错误包装、`me`）；④ handoff 文件 import/export UX；⑤ 更新 `docs/BOT.md` 使用示例；⑥ 设置页若需展示「MCP/CLI 连接说明」文案 |
| **Go 后端（协作）** | ① 保持现有 Issue/Comment/Bot REST 稳定；② 可选补 `GET /workspaces/{ws}/bots/me`（或 whoami）避免 Bot 调 JWT-only `/auth/me`；③ API Key / Idempotency / Bot 禁 done **行为不变**；④ 不在本迭代做 scope 列除非产品加需求 |
| **不归本规格** | Issue 状态机改动、看板 UI、Webhook 新事件、SSO |

**建议顺序：** client 库 → MCP P0 tools → CLI 同名命令 → 验收脚本（curl 或 `ymbot` 串联）→ 短报关闭。

---

## 9. 给总指挥的短报（≤10 行）

1. **通道**：**MCP + CLI 都要**（共享 REST 客户端）；先 MCP P0 便于 Cursor，CLI 同步或紧随。  
2. **前提**：鉴权 / Issue / 评论 / 记忆 / handoff **REST 已齐**；缺口在通道层，非业务重写。  
3. **P0 能力**：whoami · issue list/get/create/update · comment list/create · memory list/set · handoff get/import。  
4. **鉴权**：`Bearer ymk_…`，绑定 workspace+member；Admin 签发；禁进 git/handoff。  
5. **护栏**：幂等默认开；Bot 不自动 done；错误中文可读。  
6. **落地**：Bot 记忆角色做 MCP/CLI；Go 仅可选补 `bots/me`。  
7. **验收**：无浏览器走通「whoami→handoff get→干活→评论→记记忆（服务端）」；export 仅换机可选；Inspector + CLI help 过关。  
8. **规格全文**：`docs/BOT_MCP_CLI.md`。

---

## 10. 参考命令（示意，非实现）

```bash
export YMBOT_API_URL=http://127.0.0.1:8080
export YMBOT_API_KEY=ymk_***
export YMBOT_WORKSPACE_ID=<ws>
# 可选：YMBOT_IDEMPOTENCY_KEY=<uuid>  # 未传 --idempotency-key 时用于写操作

ymbot whoami
ymbot handoff get                                          # 续命：服务端摘要，无文件
ymbot issue list --assignee me --status in_progress
ymbot comment create --issue YM-12 -m "进展：……" --idempotency-key <uuid>
ymbot memory set --key last_focus --value "YM-12 复现路径" --idempotency-key <uuid>
# 可选：YMBOT_AUTOSAVE=1 或 --autosave → 写成功后 stderr「已写服务端」
# 换机备份可选（日常不需要）：
# ymbot handoff export --out /tmp/handoff.json
```

MCP 写 tool（`board_issue_create|update` / `board_comment_create` / `board_memory_set` / `board_handoff_import`）可传 `idempotency_key`；错误 `isError` content 为一行中文。

MCP：宿主配置 server 命令指向同一二进制（如 `ymbot mcp-serve`），env 注入上述变量。

---

## 11. 尾勾：`--help` / `handoff import` / MCP Inspector

> 通道 P0 代码已关；本节补**操作文档缺口**（help 树、import 救场契约、Inspector 怎么验），便于新人与验收自跑。

### 11.1 CLI `--help`

```bash
# 顶层帮助（无需 Key）
go run ./cmd/ymbot --help

# 子命令帮助：当前实现会先校验环境，请先 export 占位（假 Key 即可看 flag）
export YMBOT_API_URL=http://127.0.0.1:8080
export YMBOT_API_KEY=ymk_help_placeholder
export YMBOT_WORKSPACE_ID=00000000-0000-0000-0000-000000000000
go run ./cmd/ymbot handoff --help          # 若未拆子 help，见顶层 handoff 行
go run ./cmd/ymbot handoff import --help   # --file / - 、--idempotency-key、--full|--id|--json
go run ./cmd/ymbot team --help
go run ./cmd/ymbot seat --help
```

主航线（与顶层 help 一致）：`team pull|join` → `seat enter` → `handoff get` → 干活。  
`export` / `import` 仅换机备份与**救场**，非日常续工。

### 11.2 `handoff import`（救场）

| | |
|--|--|
| CLI | `ymbot handoff import --file path.json` 或 `--file -`（stdin） |
| MCP | `board_handoff_import`（入参：handoff JSON 对象或包；可选 `idempotency_key` / `full` / `ids`） |
| REST | `POST …/bots/{memberID\|me}/handoff` |
| 成功输出 | `imported_*` / `claimed` / `warnings`；claim 冲突 → **warning**，不整单失败 |
| **不要** | 日常换机会话默认 import；用 mirror `latest.json` 覆盖仍健康的服务端记忆 |

```bash
# 灾难/空仓救场示例（先确认服务端确实空或产品同意覆盖）
ymbot handoff export --out /tmp/handoff-backup.json   # 可选冷备
ymbot handoff import --file /tmp/handoff-backup.json
# 或：cat latest.json | ymbot handoff import --file -
```

### 11.3 MCP Inspector / Cursor 接入

**启动（stdio）**：同一二进制。

```bash
export YMBOT_API_URL=http://127.0.0.1:8080
export YMBOT_API_KEY=ymk_***          # 真 Key
export YMBOT_WORKSPACE_ID=<ws-uuid>
# 仓库根或 backend 目录：
go run ./cmd/ymbot mcp-serve
# 或已安装：ymbot mcp-serve
```

**Cursor `mcp.json` 示意**（路径按本机改）：

```json
{
  "mcpServers": {
    "ymbot": {
      "command": "go",
      "args": ["run", "./cmd/ymbot", "mcp-serve"],
      "cwd": "/absolute/path/to/ai-collab-board/backend",
      "env": {
        "YMBOT_API_URL": "http://127.0.0.1:8080",
        "YMBOT_API_KEY": "ymk_***",
        "YMBOT_WORKSPACE_ID": "<ws-uuid>"
      }
    }
  }
}
```

**MCP Inspector**：把 command/args/env 配成同上，连接后：

1. `tools/list` 应含主航线：`board_team_pull`、`board_team_join`、`board_seat_enter`、`board_handoff_get`，以及 issue/comment/memory/whoami、`board_handoff_import`、`board_seat_leave`、`board_seat_expand` 等。  
2. 冒烟：`board_whoami` → `board_issue_list` → `board_comment_create` → `board_memory_set`（写类可带 `idempotency_key`）。  
3. 读类默认瘦文本 / 无 UUID；要全量传 `full=true`，要裸 id 传 `ids=true`（同 CLI `--full` / `--id`）。

**自检一句**：Inspector 能 list tools 且至少跑通 list + comment create + memory set，即本尾勾文档验收通过。

### 11.4 团队上下文 / T6-R 隔离冒烟

> 对接 T6「团队切换隔离」**最小切片**（roster 不串队）。契约全文与本波边界：[`BOT_TEAM_TEMPLATE_T6_T7.md`](./BOT_TEAM_TEMPLATE_T6_T7.md)（T6-R）· 产品刀：[`NEXT_WAVE_DEBT_T6R_VCLOSE.md`](./NEXT_WAVE_DEBT_T6R_VCLOSE.md)。

| | |
|--|--|
| **切队怎么做** | **换入参**，无 `YMBOT_TEAM_ID` sticky。CLI：`team pull --uuid`（队 A 或 B）；`seat enter` / `leave` / `expand --team <uuid>`。MCP：`board_team_pull.team_uuid`；`board_seat_*` / `board_seat_expand` 的 `team_uuid` |
| **环境仍只需** | `YMBOT_API_URL` · `YMBOT_API_KEY` · `YMBOT_WORKSPACE_ID`（同 §11.1/11.3） |
| **REST 作用域** | `/workspaces/{ws}/teams/{teamID}/roster` · `…/seats` · … — path 上的 `teamID` 即隔离边界 |
| **本切片不做** | `board_issue_list` 按 team 过滤；FE 切换器（**pins 通道见 §11.5**，已挂载） |

```bash
# 同 WS 两队（把 A/B 换成真实 team UUID）
ymbot team pull --uuid "$TEAM_A" --json > /tmp/roster-a.json
ymbot team pull --uuid "$TEAM_B" --json > /tmp/roster-b.json
# 验收：两侧 seat id 集合不相交；队 A 的 active claim 不出现在队 B 输出
```

MCP 等价：Inspector 对 `board_team_pull` 各传一次不同 `team_uuid`（可加 `ids=true` 看裸 id）。  
**自检一句**：双 pull 席位组不串 → T6-R 通道文档验收过；**不**因此勾满 `BOT_TEAM_SEAT_P0` T6 全文。

### 11.5 Team pins（`board_pin_*` · 已落地 `63e69c6`）

> 对接 T6 余债 **共享记忆墙**。权威开工单：[`NEXT_WAVE_DEBT_PINS.md`](./NEXT_WAVE_DEBT_PINS.md)。REST：[`API.md`](./API.md)「Team pins」· `6915db4`。  
> **状态：已落地**——通道实现 SHA **`63e69c6`**（botclient + `ymbot pin` + MCP `board_pin_list`/`get`/`set`/`delete`；REST `6915db4`）。箱验 P1–P3 证据：[`qa-shots/pins-team-isolate-20261024.txt`](./qa-shots/pins-team-isolate-20261024.txt)。日常续命仍 `handoff get`，**不**用 pin 替代。

| | |
|--|--|
| **与私有 memory 分界** | pins = 队共享短结论；`board_memory_*` / `bot_memories` = 席位私有。同名 `key` **不**互相覆盖；**不**用 pin 替代 `handoff get` |
| **切队** | 同 §11.4：换入参 `team_uuid`（无 sticky env）。CLI/MCP 每次带本队 UUID |
| **环境仍只需** | `YMBOT_API_URL` · `YMBOT_API_KEY` · `YMBOT_WORKSPACE_ID` |
| **REST** | `GET/PUT/DELETE /workspaces/{ws}/teams/{teamID}/pins`（及 `…/pins/{key}`）；鉴权本队成员或 Admin；错队 → **404** |
| **瘦输出** | 默认文本友好 / 无裸 UUID；`--full`/`full=true`、`--id`/`ids=true` 同 §6 / DX |
| **实现 SHA** | 通道 **`63e69c6`** · REST `6915db4` |

| 能力 | MCP tool | CLI | REST | 入 / 出要点 |
|------|----------|-----|------|-------------|
| 列本队 | `board_pin_list` | `ymbot pin list --team <uuid>` | `GET …/teams/{teamID}/pins` | 入：`team_uuid`；出：`pins[]`（瘦：`key,value` 摘要） |
| 读一条 | `board_pin_get` | `ymbot pin get --team <uuid> --key <k>` | `GET …/pins/{key}` | 入：`team_uuid, key`；无 → 404 一行中文 |
| 写入 | `board_pin_set` | `ymbot pin set --team <uuid> --key <k> --value <v>` | `PUT …/pins` `{key,value}` | upsert；可选 `idempotency_key`；出：`pin` |
| 删除 | `board_pin_delete` | `ymbot pin delete --team <uuid> --key <k>` | `DELETE …/pins/{key}` | 入：`team_uuid, key`；成功 ok；无 → 404 |

```bash
ymbot pin set --team "$TEAM_A" --key decision --value "采用方案 B"
ymbot pin list --team "$TEAM_A"
ymbot pin list --team "$TEAM_B"   # 不应出现队 A 的 decision
ymbot pin get  --team "$TEAM_B" --key decision   # 期望 404
ymbot pin delete --team "$TEAM_A" --key decision
```

MCP：`tools/list` 含 `board_pin_list` / `board_pin_get` / `board_pin_set` / `board_pin_delete`（实现 `63e69c6`）。  
**自检一句**：队 A set → 队 A get 见、队 B list/get 无/404（REST 箱验已 PASS；通道冒烟 list/get/set/delete 不 501）。
