# 班底手册索引（T7-R · 种子）

> 2026-10-24 · 种子与文档 · 对齐 [`NEXT_WAVE_DEBT_T7.md`](./NEXT_WAVE_DEBT_T7.md)（`95e27c6`）  
> **本波只服务「选模板 → 空席 + playbook_ref」**。升版 diff/apply（T7-4/5）**后置**，见 [`BOT_TEAM_TEMPLATE_T6_T7.md`](./BOT_TEAM_TEMPLATE_T6_T7.md)。  
> 手册本体在兄弟仓 **`bot-team-template`**（与本看板仓并列）；下表「最小示例路径」相对该仓根。
> 本仓另有内置 JSON **可读副本**：[`bot-team-templates/`](./bot-team-templates/)（`default.json` + README；与 `template_key=default` 对齐）。

## 内置 / 约定 `template_key`

| `template_key` | 含义 | 空席形状（slug） | 手册锚点 |
|----------------|------|------------------|----------|
| `default` | 看板现网 ensure 默认五席（已有 Go `defaultRosterSeats`） | `frontend` · `pm` · `qa` · `backend` · `wiki` | 业务仓 seed / `roster/ensure` 无 seats 时 |
| `bot-team-template` | 作业手册「标准班底」名册（T7-R 目标键；**清单 API 实现待 SHA**） | 以手册名册为准（见下示例） | 见「最小示例路径」 |

## 最小示例路径（真实 / 约定）

| 用途 | 路径（相对 `bot-team-template/`） |
|------|----------------------------------|
| 仓说明 / 一次跑通 | `README.md` |
| 班底名册与补位 | `playbook/班底与补位.md` |
| 席位迁绑主航线 | `playbook/看板席位迁绑.md` |
| 名册 JSON 样例 | `playbook/templates/班底/roster.example.json` |
| 席位人设目录（slug≠中文名） | `roles/README.md` · 如 `roles/前端研发/` → slug `frontend` |
| ymbot 配手样例 | `playbook/templates/ymbot/` |


## 本仓内置副本（`template_key=default`）

| 本仓路径 | 说明 |
|----------|------|
| [`bot-team-templates/README.md`](./bot-team-templates/README.md) | 内置键说明 · REST 回链（`team-templates` / `template_key`） |
| [`bot-team-templates/default.json`](./bot-team-templates/default.json) | 五席 + 每席 `playbook_ref`（path 须存在于兄弟仓） |

### `default` 席位 → 兄弟仓 path（须存在）

| slug | `playbook_ref.path`（相对 `bot-team-template/`） |
|------|--------------------------------------------------|
| `frontend` | `roles/前端研发/` |
| `pm` | `roles/项目经理/` |
| `qa` | `roles/QA/` |
| `backend` | `roles/平台后端/`（场馆向可改 `roles/场馆后端/`） |
| `wiki` | `playbook/班底与补位.md`（MCP：文档席可无 `roles/` 目录） |

> 勿写不存在的 `playbook/frontend.md` 一类占位；slug↔`roles/` 对照以兄弟仓 `playbook/看板MCP与CLI.md` §4 为准。

## `playbook_ref` 约定（契约）

推荐落库为**字符串**（现网 `team_seats.playbook_ref` TEXT）：

```text
bot-team-template@main:playbook/班底与补位.md
bot-team-template@main:roles/前端研发/
```

等价对象（API 可收，服务端可序列化为上式）：`{ "repo": "bot-team-template", "ref": "main", "path": "playbook/班底与补位.md" }`。

## REST 回链

[`API.md`](./API.md)「Team templates（T7-R）」· 验收行对齐 T6_T7 的 **T7-1 / T7-2 / T7-3 / T7-6 / T7-7**（本波）；**T7-4 / T7-5 不写进本波必验**。
