# Bot 续命 · 服务端 handoff 为准（P0）

> 状态：**H1–H5 全部已勾关单**（产品体验验收 2026-10-24 · Asia/Shanghai）
> 痛点：导入/导出麻烦，怕中途挂掉丢上下文。  
> 原则：**续命以服务端 `handoff get` 为准**；干活过程中**自动**写 memory / todo / comment；**`export` 仅可选备份**，不是续工主路径。  
> 仓库：`/workspace/ai-collab-board`  
> 对照：[`BOT.md`](./BOT.md) · [`BOT_MCP_CLI.md`](./BOT_MCP_CLI.md) · [`BOT_CHANNEL_DX_P0.md`](./BOT_CHANNEL_DX_P0.md) · 短原则 [`BOT_CONTINUITY_P0.md`](./BOT_CONTINUITY_P0.md)  · 备份旁路（≠主路径）[`BOT_HANDOFF_MIRROR_P0.md`](./BOT_HANDOFF_MIRROR_P0.md)  
> **落地：Bot 记忆角色**（通道指引 + 文档/示例 + 可选 MCP/CLI 文案）；**产品体验**按 §4 **H1–H5** 勾选。不改 Issue 业务 handler 语义；handoff 仍无独立表（实时组装）。

---

## 0. 用户裁定（一句话）

换会话 / 换机 / 中途挂掉 → **再 `handoff get`（服务端）就能续**；不要依赖「先 export 文件再 import」。过程中该记的记到服务端；文件导出只是保险柜。

---

## 1. 目标模型 vs 旧习惯

| | **新（本 P0）** | **旧（降级）** |
|--|----------------|----------------|
| 续命主路径 | `GET …/bots/{me}/handoff` / `ymbot handoff get` / `board_handoff_get` | export JSON → 拷贝 → import |
| 上下文来源 | 服务端已持久的 memories + todos + 指派开放 Issues（实时组装） | 某次导出的快照文件 |
| 干活中 | **自动/及时** `memory set`、todo 维护、`comment create` | 会话结束才一次性倒进文件 |
| `handoff export --out` | **可选备份**（审计、冷备、合规）；**不**写进「必须换机步骤」 | 换机必经 |
| `handoff import` | **可选恢复**（从备份救场 / 迁入外来包）；**不是**日常续工 | 换机必经 |

一句话教 Bot：**先 get 再干；干就写回服务端；文件可有可无。**

可选提醒：CLI `--autosave` 或 `YMBOT_AUTOSAVE=1` → claim / comment / memory set 成功后 stderr「已写服务端」（约定，非 daemon，非本地文件）。

---

## 2. 可执行约定

### 2.1 会话启动（续命）

1. `whoami`  
2. **`handoff get`**（默认摘要即可；需要细节再 `--full`）→ 恢复角色 / 记忆要点 / todos / 开放 `YM-n`  
3. 按开放 Issue 继续；**禁止**把「必须先 import 某个 json」写成默认剧本  

断线 / 杀进程 / 换 Cursor 会话：**同一 Bot Key** 再 `handoff get` 即续，无需文件。

### 2.2 过程中自动写回（防中途挂）

| 时机 | 应落服务端 | 通道 |
|------|------------|------|
| 认清当前焦点 / 卡点 | `memory set`（如 `last_focus`、`blocker`） | `ymbot memory set` / `board_memory_set` |
| 步骤拆解 / 进行中项 | todo 创建或改状态（P1 通道若未齐，可用 memory 键暂代，但文档须标明） | `board_todo_*` 或等价 |
| 对人类可见进展 | `comment create` 挂在对应 `YM-n` | `ymbot comment create` |
| Issue 属性变化 | `issue update`（禁 Bot 自 `done`） | 既有 |

**验收语义**：任选一「中途挂」模拟——只完成 memory+comment、**不** export——新会话仅 `handoff get` 仍能看到焦点与评论（todo 若 P1 未齐则 memory 键可见即可，须在证据注明）。

### 2.3 export / import 定位

| 命令 | 定位 | 文档/help 须写明 |
|------|------|------------------|
| `handoff export --out` | 可选备份 | 「备份用；续工请用 `handoff get`」 |
| `handoff import --file` | 可选从备份/外来包恢复 | 「非常用；日常续工不依赖」 |
| `handoff get` | **续命权威** | 主路径第一公民 |

不得在 README / BOT.md / MCP 描述里把 export→import 写成唯一换机流程。

### 2.4 非目标

- 新建 handoff 持久表或改组装字段契约（除非缺口阻塞 get）  
- 强制前端设置页大改（文案提示可做）  
- 取消 import/export 能力（保留，降级为备份）  
- 自动把每次 tool call 无差别全量打进 memory（要**有意义**的焦点/卡点/进展，避免刷屏）
- 把定时磁盘镜像当成日常续命（镜像是**备份旁路**，见 [`BOT_HANDOFF_MIRROR_P0.md`](./BOT_HANDOFF_MIRROR_P0.md)；进场才读）

---

## 3. 落地归属

| 角色 | 事项 |
|------|------|
| **Bot 记忆角色** | ① 改 `BOT.md` / `BOT_MCP_CLI` / help / MCP tool description：主路径= get + 过程写回 ② 示例剧本去掉「必须 export」 ③ 可选：CLI `handoff export` help 标注「可选备份」 ④ 自测 H1–H5 后喊产品体验勾选 |
| **Go 后端** | handoff GET 组装已够用则**不动**；仅当 get 缺关键上下文时补小缺口 |
| **产品体验** | §4 **H1–H5** 勾选；假勾拒绝 |

---

## 4. 验收（产品体验 · H1–H5）

- [x] **H1** 文档/help/MCP 描述明确：**续命 = 服务端 `handoff get`**；export/import 标为可选备份/救场  
- [x] **H2** 推荐剧本：`whoami` → `handoff get` → 干活 → 过程中 `memory`/`comment`（及 todo）；**无**强制 export 步骤  
- [x] **H3** 「中途挂」复验：写入 memory + comment 后**不** export；新进程仅 `handoff get` 能还原焦点（及开放 Issue）；评论在对应 `YM-n` 可见  
- [x] **H4** `handoff export --out` 仍可用，但 help/文档称备份；**不**作为验收必经关卡  
- [x] **H5** `BOT.md` §续工与 `BOT_MCP_CLI`「典型一天」已与本文件一致；与通道 DX（YM-n / 瘦输出）不冲突  

### 勾选证据（2026-10-24 · Asia/Shanghai · 产品体验 · 本机真 Key）

| ID | 结果 | 证据 |
|----|------|------|
| **H1** | 通过 | `ymbot --help` / handoff 子命令文案：续命=`handoff get`；export=换机备份可选；MCP `board_handoff_import` 描述「救场；非日常」；`BOT.md` / `BOT_CONTINUITY_P0` 对齐 |
| **H2** | 通过 | 典型一天与 BOT 续工：`whoami`→`handoff get`→过程 memory/comment；**无**强制 export |
| **H3** | 通过 | 真 Key：`memory set last_focus=H3-focus-…` + `comment` 于 YM-9 → **不** export；新进程 `handoff get`/`memory list` 见焦点；`comment list --issue YM-9` 见评论；`--autosave` stderr「已写服务端」 |
| **H4** | 通过 | `handoff export --out` 仍成功（备份文件可读）；help 标明可选；非关卡 |
| **H5** | 通过 | `BOT.md` / `BOT_MCP_CLI` / DX 不冲突；`issue get` 默认含 `version`（list 仍无）便于乐观锁 |


---

## 5. 给总指挥的短报（≤6 行）

1. **裁定落文档**：`docs/BOT_HANDOFF_CONTINUITY_P0.md`。  
2. **续命**：服务端 `handoff get` 为准；不怕中途挂——上下文在服务端。  
3. **过程**：自动/及时写 memory、todo、comment。  
4. **export**：仅可选备份；import 非日常。  
5. **落地**：Bot 记忆角色已落地；产品体验 **H1–H5 已勾关单**（2026-10-24）。
