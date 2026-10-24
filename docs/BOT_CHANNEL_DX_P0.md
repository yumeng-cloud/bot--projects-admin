# Bot 通道 DX · 省 token / 少 ID / 多可读（P0）

> 状态：**D1–D6 全部已勾关单**（产品体验验收 2026-10-24 · Asia/Shanghai）  
> 原则：**省 token、少 UUID、多可读文案**；默认给人（和模型）看，`--full` 才给机器裸 id。  
> 仓库：`/workspace/ai-collab-board`  
> 权威验收：本文；通道总规格仍见 [`BOT_MCP_CLI.md`](./BOT_MCP_CLI.md)（§6 已回链）。  
> **不改业务 handler 语义**；通道层（`cmd/ymbot` + MCP 输出整形 + 可选 client 解析）由 **Bot 记忆角色**牵头；**产品体验**勾选关单。

---

## 0. 用户裁定（一句话）

Issue 对外用 **`YM-n`（`identifier`）**；列表极瘦；CLI/MCP **默认文本友好**；**禁止**默认把 REST 原样 dump。  
默认输出中 **不得出现** 32 位 hex UUID（可带连字符的 UUID 形态）。要裸 id / 全量时显式 **`--full`** 或 **`--id`**（MCP：`full=true` / `ids=true`）。

---

## 1. 与既有文档

| 文档 | 职责 | 本改是否动 |
|------|------|------------|
| [`BOT_MCP_CLI.md`](./BOT_MCP_CLI.md) | MCP+CLI 能力清单、鉴权、§7 串联 | **保留**；§6 增 DX 条款 + 链本文 |
| [`BOT.md`](./BOT.md) | 记忆/角色/handoff REST | 不改表语义；示例可改用 `YM-n` |
| [`API.md`](./API.md) | REST 仍返回完整 `id` + `identifier` | **API 体不变**；瘦身在通道展示层 |
| **本文** | 通道输出 / 入参可读性验收 | 权威勾选 |

REST 继续带 UUID（乐观锁、关联、前端不变）。DX 只约束 **ymbot CLI + MCP tool 返回给人看的那一层**。

---

## 2. 规则（可执行）

### 2.1 Issue 标识：`YM-n`

| 场景 | 默认 | `--full` / `full=true` |
|------|------|------------------------|
| 列表 / get / 评论结果里的 Issue | 展示 **`identifier`**（如 `YM-12`） | 可同时带 `id`（UUID） |
| CLI `--issue` / MCP `issue_id` 入参 | **接受 `YM-12` 或 UUID**；通道解析后打 REST | 同左 |
| 错误文案 | 优先写 `YM-12`，勿只甩 UUID | 可附 `id=` |

解析失败 → 一行中文（如「找不到 Issue YM-99」）+ 非 0 退出 / MCP `isError`。  
`identifier` 已由后端按 `key_prefix-number` 生成（seed 示例项目为 `YM`）；通道 **不得**自造另一套编号。

### 2.2 列表瘦字段（默认）

`issue list` / `board_issue_list` 默认每行（或等价 JSON 瘦对象）**仅**：

| 字段 | 说明 |
|------|------|
| `identifier` | 如 `YM-12`（主钥匙） |
| `title` | 标题 |
| `status` | 状态 |
| `priority` | 优先级 |
| `assignee` | **可读名**（无则 `-` / 空）；默认 **不**出 `assignee_id` |
| `updated_at` | 可选；若出则短 ISO 即可 |

**默认禁止**出现在列表行：`id`、`project_id`、`workspace_id`、`description`、`version`、整段 labels 数组、关系图等大块。  
需要时：`--full`（或 MCP `full=true`）再吐接近 REST 的完整对象；仅要钥匙时用 `--id`（见 §2.3）。

**`issue get` / `board_issue_get`（及 create/update/claim 写回结果）默认瘦输出须含 `version`**（乐观锁整数），以便 `issue update` **不必** `get --full`。`issue list` 默认仍**不**带 `version`。冲突 409 时应提示重新 `issue get` 取 version（无需 `--full`）。

#### handoff / memory / comment（默认摘要）

| 命令 | 默认 | `--full` |
|------|------|----------|
| `memory list` | `key` + 值摘要（截断）+ 可选 tags 短串；**无** memory UUID | 含 `id` 与全文 |
| `handoff get`（终端打印） | 角色名、记忆条数、todo 条数、开放 Issue 的 **`YM-n`+title** 列表；**禁止**整包 REST JSON | 接近导出包；文件 `--out` 仍可含 id（导出给换机，不算「默认打印」） |
| `comment list` | 作者显示名 + 时间 + body 摘要；Issue 用 `YM-n`；**无** comment/issue UUID | 含裸 id |

**硬禁（默认路径）**：通道输出 **禁止 REST 原样转发**；**禁止**出现 32 位 hex（UUID）。验收用 `rg`/`grep` 扫默认 stdout 即可判。

### 2.3 默认文本友好

| 通道 | 默认 | `--full` / `full=true` |
|------|------|------------------------|
| **CLI**（TTY 或未指定格式） | **人类可读文本**（多行表或 `YM-12  title  status  …`），不是整树 indent JSON | 缩进 JSON，含裸 `id` |
| **CLI**（管道 / 明确要机读） | 允许默认 **瘦 JSON 数组**（仅 §2.2 字段、无 UUID）；与「少 token」一致 | 完整 JSON |
| **MCP** | tool 返回的 `content[].text` 为 **短文本或瘦 JSON 字符串**（优先短文本表）；禁止默认 dump 全 Issue | `full=true` 时完整 JSON 文本 |

建议 CLI 增加（名称可微调，语义不变）：

| 旗标 | 作用 |
|------|------|
| **`--full`** | 接近 REST 全量 JSON（含裸 UUID、description、version…） |
| **`--id`** | 仍瘦/可读，但**额外**放出相关裸 `id`（Issue/member/memory）；比 `--full` 轻 |
| 可选 `--json` | 强制瘦 JSON（非 TTY）；`--json --full` = 完整 JSON |

MCP 对应：`full=true` ≈ `--full`；`ids=true` ≈ `--id`。

未传 `--full` / `--id`（及 MCP 等价）时，stdout / tool text **不得**出现 UUID，**不得**以 UUID 为主键展示 Issue。

### 2.4 入参与示例（对照）

```bash
# 默认：可读、瘦
ymbot issue list --assignee me
# YM-12  修好登录态  in_progress  high  前端Bot  2026-10-24T…

ymbot issue get YM-12
# YM-12  修好登录态
# status=in_progress  priority=high  assignee=前端Bot
# version=3
ymbot comment create --issue YM-12 -m "已复现"

# 只要钥匙（仍瘦）
ymbot issue list --assignee me --id

# 全字段 / REST 级
ymbot issue list --assignee me --full
ymbot issue get YM-12 --full
```

MCP：`board_issue_list` 默认瘦；`ids: true` 放裸 id；`full: true` 全量。

---

## 3. 非目标（本 P0）

- 改 REST 字段名或去掉响应里的 `id`
- 改乐观锁语义（`version` 仍在 update 路径；**通道默认 get 须露出 version**，list 仍可藏）
- 新编号体系、跨项目全局序号
- 重做 handoff 包 schema（导出文件可仍含 id；**终端默认打印**要瘦）

---

## 4. 落地归属

| 角色 | 事项 |
|------|------|
| **Bot 记忆角色** | ① CLI 默认文本 / 瘦 JSON + `--full` / `--id` ② 入参解析 `YM-n`↔id（list/get 缓存或按 identifier 查）③ MCP 默认瘦文本 ④ 更新 `BOT.md` / help 示例 ⑤ 自测后喊产品体验勾选 |
| **Go 后端** | 仅当「按 identifier 取 Issue」REST 不便时补小缺口（如已有 get-by-id 则通道可先 list/`q=` 解析）；**不**改 Issue 业务语义 |
| **产品体验** | 按 §5 **D1–D6** 勾选关单；假勾拒绝 |

---

## 5. 验收（产品体验 · D1–D6）

> 勾选前对 **未带** `--full`/`--id` 的默认命令扫 stdout：不得匹配 UUID（8-4-4-4-12 hex）。  
> 对照报告也可记入 [`ACCEPTANCE_UX_BOT.md`](./ACCEPTANCE_UX_BOT.md) §D。

- [x] **D1** Issue 对外主钥匙为 **`YM-n`（identifier）**；`--issue YM-12` / MCP 等价可跑通 get/comment/claim/update  
- [x] **D2** `issue list` **极瘦**（§2.2 字段集）；无 description/version/workspace 等大块  
- [x] **D3** CLI/MCP **默认文本友好**；**禁止 REST 原样 dump**  
- [x] **D4** 默认输出 **不出现 32 位 hex UUID**；**`--id`（`ids=true`）** 才附加裸 id；**`--full`（`full=true`）** 才接近 REST 全量  
- [x] **D5** `handoff get` / `memory list` / `comment list` **默认摘要**（§2.2 表）；全量要 `--full`  
- [x] **D6** MCP P0 列表/读类与 CLI 同一套标识规则；help / `BOT.md` 示例已改用 `YM-n`

### 勾选证据（2026-10-24 · Asia/Shanghai · 产品体验 · 本机真 Key）

> 复验通过。通道自测笔记保留作参考。

| ID | 结果 | 证据 |
|----|------|------|
| **D1** | 通过 | `ymbot issue get YM-9` / `YM-12` 可读；`YM-9999` →「找不到 Issue YM-9999」exit **12** |
| **D2** | 通过 | `issue list --assignee me` 仅 identifier/title/status/priority/assignee/updated_at；单测 `TestProjectIssuesSlimNoUUID` 绿 |
| **D3** | 通过 | 默认文本表（非 REST dump）；whoami=`bot  Grok Bot`；`go test ./internal/botclient` 绿 |
| **D4** | 通过 | 默认 whoami/list/get/handoff/memory/comment **uuid_hits=0**；`issue get` 默认含 `version`（无 UUID）；`--id` 附 `id=`；`--full` 含 UUID |
| **D5** | 通过 | handoff=`role… memories=N` + YM-n 行；memory=`key	时间	摘要`；comment=`YM-n 作者 时间 body` |
| **D6** | 通过 | MCP `tools/list`=12 且 schema 含 `full`/`ids`；`--help` + `BOT.md` 示例为 YM-n |


---

## 6. 给总指挥的短报（≤8 行）

1. **裁定已落文档**：`docs/BOT_CHANNEL_DX_P0.md`（并回链 `BOT_MCP_CLI.md` §6）。  
2. **规则**：Issue 用 `YM-n`；list 极瘦；默认禁 UUID/禁 REST dump；`--id` 放钥匙、`--full` 全量。  
3. **API 不变**：UUID 仍在 REST；DX 只在 CLI/MCP 展示与入参解析。  
4. **落地**：Bot 记忆角色已落地；产品体验 **D1–D6 已勾关单**（2026-10-24）。  
5. **非目标**：不改编号体系、不改乐观锁、不瘦 REST。
