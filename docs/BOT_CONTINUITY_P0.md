# Bot 续命 · 短原则

> 权威验收 **H1–H5**：[`BOT_HANDOFF_CONTINUITY_P0.md`](./BOT_HANDOFF_CONTINUITY_P0.md)。流程见 [`BOT.md`](./BOT.md)「续命」「续工流程建议」；通道 [`BOT_MCP_CLI.md`](./BOT_MCP_CLI.md)。

## 四条裁定

1. **真相源 = 服务端**：handoff / memories / todos / 指派 Issue；**不是**本地文件。
2. **干活中增量写服务端**：开干就 `memory set` / todo；关键节点 `comment`；claim 后进度写记忆；**禁止**攒到最后才 `handoff export`。
3. **新 Bot / 新会话续工**：`ymbot whoami` → `ymbot handoff get`（**无文件**）；MCP：`board_whoami` → `board_handoff_get`。
4. **`handoff export`**：仅换机备份可选；日常续工不需要。`import` 为救场，非日常。

## 可选提醒

`--autosave` 或 `YMBOT_AUTOSAVE=1`：claim / comment / memory set **成功后** stderr 提示「已写服务端」。语义是**约定干活中写服务端**，**不是**后台守护进程，也**不是**另写本地文件。

## 旁路（≠主路径）

定时只读磁盘镜像见 [`BOT_HANDOFF_MIRROR_P0.md`](./BOT_HANDOFF_MIRROR_P0.md)：备份旁路，进场才读；**不**替代 `handoff get`，定时动作**禁止**写 memory。
