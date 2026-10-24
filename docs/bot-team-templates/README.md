# Bot Team Templates（内置 · 可读副本）

仓内最小班底手册索引副本。权威契约与兄弟仓路径见 [`../BOT_TEAM_TEMPLATE_INDEX.md`](../BOT_TEAM_TEMPLATE_INDEX.md)。  
Go 运行时嵌入 `backend/internal/services/botmem/templates/`（实现待 SHA；**勿以本目录替代清单 API**）。本目录仅供人读 / 对齐索引。

| `template_key`（JSON `id`/`key`） | 说明 |
|-----------------------------------|------|
| `default` | 五席：`frontend` / `pm` / `qa` / `backend` / `wiki`；每席挂 `playbook_ref` → 兄弟仓真实路径 |

REST（与 [`../API.md`](../API.md)「Team templates（T7-R）」一致）：

- `GET /api/v1/workspaces/{ws}/team-templates` — 清单（**实现待 SHA**）
- `POST …/teams` 带 **`template_key`**（非 `template_id`）— 从模板创建空席

`playbook_ref`：对象 `{repo,ref,path}` 或字符串 `repo@ref:path`，见索引。
