# 育蒙「AI 协作看板」验收清单（QA）

> 状态：已与仓库根目录 [`COORDINATION.md`](../COORDINATION.md) 对齐（路径、`/api/v1` 合同、角色所有权）。  
> 用途：功能落地后按条勾选；失败时按「责任归属」回派对应 Bot。  
> 环境约定：`Asia/Shanghai` 展示；默认 SQLite；种子管理员见 `.env.example` / `COORDINATION.md`。  
> **协作入口**：分工与预留路由以仓库根 `COORDINATION.md` 为准（勿改他人目录/路由）。

## 0. 如何跑测

### 0.1 本地一键

```bash
cd /workspace/ai-collab-board
cp .env.example .env   # 可选
make install           # go mod tidy + npm install
make seed              # Makefile 会带 SEED_SAMPLE_DATA=1；或 AUTO_SEED=1+SEED_SAMPLE_DATA=1（二者默认关）
make dev               # API :8080 + Vite :5173
```

种子账号（与 `COORDINATION.md` / `.env.example` 一致）：

- Email: `admin@yumeng.tech`
- Password: `Admin123!`
- Workspace: `育蒙科技`（slug: `yumeng`）

### 0.2 单元 / 包测试

```bash
cd backend && go test ./... -count=1
# 或根目录
make test             # 后端 go test + 前端 npm run build
```

地基包：`internal/auth` 含 `jwt_test.go` / `password_test.go`。

### 0.3 API Smoke（curl，前缀 `/api/v1`）

```bash
BASE=http://127.0.0.1:8080

# 健康检查（两处均已实现）
curl -sS "$BASE/healthz"
curl -sS "$BASE/api/v1/healthz"

# 登录
curl -sS -X POST "$BASE/api/v1/auth/login" \
  -H 'Content-Type: application/json' \
  -d '{"email":"admin@yumeng.tech","password":"Admin123!"}'

# 当前用户 / 工作空间（JWT）
TOKEN=...   # 取登录响应 token
curl -sS "$BASE/api/v1/auth/me" -H "Authorization: Bearer $TOKEN"
curl -sS "$BASE/api/v1/workspaces" -H "Authorization: Bearer $TOKEN"
curl -sS "$BASE/api/v1/workspaces/{workspaceID}" -H "Authorization: Bearer $TOKEN"

# 注册 / 登出
curl -sS -X POST "$BASE/api/v1/auth/register" \
  -H 'Content-Type: application/json' \
  -d '{"email":"qa@example.com","password":"Test1234!","name":"QA"}'
curl -sS -X POST "$BASE/api/v1/auth/logout" -H "Authorization: Bearer $TOKEN"
```

Bot 路径用 `Authorization: Bearer <api_key>`（前缀约定见实现，如 `ymk_`）；写操作带 `Idempotency-Key`。  
错误体约定（顶层）：`{"code":"...","message":"中文"}`（见 `COORDINATION.md`；与实现一致，非嵌套 `error:{}`）。

预留域路由（projects / issues / wiki / time / bots 等）当前多为 **501 stub**，替换实现后删 stub，见 `cmd/server/main.go` 与 `COORDINATION.md`。

### 0.4 责任归属速查（对齐 COORDINATION）

| 代号 | COORDINATION Owner | 范围 |
|------|-------------------|------|
| **TL** | Tech lead（部分冻结） | `db/` schema 核心、`auth/` `middleware/` `config/` `pkg/`；`handlers/auth.go` `health.go` `workspace.go`；FE `Login*` `Home*` `api/client.ts` |
| **BOARD** | Issue/Board Bot（+ Issue/Board FE） | projects / issues / cycles / relations / search / 看板 UI |
| **COLLAB** | Collab Bot | comments / activities / labels / members |
| **INTEG** | Integrations Bot | api-keys / webhooks（HMAC + delivery） |
| **WIKI** | Wiki Bot（+ Wiki FE） | wiki spaces/pages CRUD + 搜索 |
| **TIME** | Time Bot（+ Time FE） | time-entries / billing-rates / 报表 / CSV |
| **BOT** | Bot Continuity Bot（+ Bot/Settings FE） | bot-roles / claim·release / memories / todos / handoff |
| **FE** | 各域 FE 页 | 交互、表单、空态、中文错误展示（跨归属时与上表域 Owner 联查） |

历史速记对照：旧 **BE** ≈ TL+BOARD+COLLAB+INTEG+WIKI+TIME+BOT 后端；旧 **KB** ≈ WIKI+TIME。

---

## 1. 登录与会话

### 1.1 注册成功

- **前置**：干净库或未占用邮箱。
- **步骤**：`POST /api/v1/auth/register`（邮箱+密码+姓名）或前端注册页提交。
- **期望**：返回用户信息与会话（JWT 或 session cookie）；密码不以明文落库（argon2id）。
- **失败归属**：TL（哈希/落库）；FE（表单校验/跳转）。

### 1.2 登录成功

- **前置**：已有用户（或 seed / `AUTO_SEED`）。
- **步骤**：正确邮箱密码登录。
- **期望**：200；拿到 token/cookie；可访问 `GET /api/v1/auth/me`；前端进入工作台。
- **失败归属**：TL / FE。

### 1.3 登录失败

- **前置**：已有用户。
- **步骤**：错误密码或未知邮箱。
- **期望**：401；中文可读错误；不泄露「用户是否存在」细节（可选统一文案）。
- **失败归属**：TL。

### 1.4 登录限流

- **前置**：`LOGIN_RATE_LIMIT` / `LOGIN_RATE_WINDOW_SEC` 生效。
- **步骤**：同一 key（IP/邮箱）短时间连续失败超限。
- **期望**：429 `rate_limited`。
- **失败归属**：TL。

### 1.5 登出与失效

- **前置**：已登录。
- **步骤**：`POST /api/v1/auth/logout` 后再访问受保护路由；或使用过期 JWT。
- **期望**：401「请先登录 / 登录已失效」。
- **失败归属**：TL；FE 需清本地态并跳登录页。

### 1.6 受保护路由

- **前置**：未带 Authorization / session。
- **步骤**：访问 `GET /api/v1/workspaces`、项目、Issue 等 API。
- **期望**：一律 401，无数据泄漏。
- **失败归属**：TL（中间件）。

---

## 2. 权限（Workspace / Role / API Key）

### 2.1 非成员不可见

- **前置**：用户 A 属工作区 W1；用户 B 不属 W1。
- **步骤**：B 用自己的 token 读 W1 的项目/Issue/Wiki。
- **期望**：403 或 404（统一策略）；无跨工作区数据。
- **失败归属**：TL（workspace 地基）；BOARD / WIKI（域资源校验）。

### 2.2 角色：Admin / Member / Bot

- **前置**：同工作区三类成员（schema：`members.role`）。
- **步骤**：Member 尝试创建 API Key、配置 Webhook、删除项目；Admin 执行同样操作；Bot Key 执行写 Issue。
- **期望**：仅 Admin 可管 Key/Webhook/删项目；Member 可日常协作；Bot 可指派范围内读写，越权 403。
- **失败归属**：INTEG（Key/Webhook）；BOARD（删项目）；设置页 UI 为 FE（Bot/Settings FE）。

### 2.3 API Key 安全

- **前置**：Admin 创建 Key（`.../api-keys`，见 COORDINATION 预留）。
- **步骤**：创建时展示明文一次；再次列表只见 prefix；库内为 hash；吊销后立即失效。
- **期望**：明文不落库、不进日志；吊销后 401。
- **失败归属**：INTEG；展示一次明文 UI 为 FE。

### 2.4 Bot Key 越权

- **前置**：Key 绑定某 Bot member / 工作区。
- **步骤**：用该 Key 访问其他工作区或 Admin-only 接口。
- **期望**：403。
- **失败归属**：INTEG / BOT / TL（鉴权中间件）。

### 2.5 活动流 actor 不可伪造

- **前置**：人会话与 Bot Key 各一。
- **步骤**：人写评论；Bot 写评论；尝试在 payload 里伪造 `actor=bot/human`。
- **期望**：actor 由服务端根据鉴权写入（`human|bot|system`）；客户端伪造无效。
- **失败归属**：COLLAB；鉴权上下文 TL。

---

## 3. 看板 / Issue / Cycle

> **成品套装 P0**：见 [`PRODUCT_SUITE_P0.md`](./PRODUCT_SUITE_P0.md)（包装 / 列模板 / 与 Timeline·抽屉一体）。

> **排期方案已裁定**（Roadmap/Timeline P0，含叶子 Bot 领取）：见 [`SCHEDULING_BOARD_P0.md`](./SCHEDULING_BOARD_P0.md)；落地后按该文 §6 增测。

路由前缀：`/api/v1/workspaces/{ws}/projects|issues|...`（见 COORDINATION「预留 · Issue / Board」）。

### 3.1 多项目与标识符

- **前置**：已登录工作区。
- **步骤**：创建项目（含 `key_prefix`）；连续创建多条 Issue。
- **期望**：标识符 `PREFIX-序号` 唯一；并发下不撞号（`issue_counters`）。
- **失败归属**：BOARD；创建表单 FE。

### 3.2 看板拖拽改状态

- **前置**：项目内多状态 Issue（`backlog|todo|in_progress|in_review|done|canceled`）。
- **步骤**：前端拖拽改列；或 `.../issues/{id}/transition`。
- **期望**：状态更新；活动流有记录；`board_order` 合理。
- **失败归属**：FE（交互）；BOARD（持久化/校验）；活动 COLLAB。

### 3.3 乐观锁冲突（409 `conflict`）

- **前置**：同一 Issue `version` 已知（来自 GET/列表）。
- **步骤**：两客户端先后 `PATCH`/`move`；后写仍带**旧** `version`。另测：完全不带 `version`。
- **期望**：
  - 版本落后 → **HTTP 409**，体 `{"code":"conflict","message":"乐观锁冲突：Issue 已被他人更新，请刷新后重试"}`（或等价中文）；FE 提示刷新；**不**静默覆盖。
  - 缺 `version` → **HTTP 400**，`code=bad_request`。
- **失败归属**：BOARD；FE 提示。

### 3.4 列表筛选与批量

- **前置**：多负责人/标签/优先级数据。
- **步骤**：按负责人、标签、状态筛选；批量改状态（若已实现）。
- **期望**：结果正确；空状态有引导。
- **失败归属**：FE 为主；筛选 API BOARD；标签 COLLAB。

### 3.5 Cycle 归属与容量

- **前置**：Cycle 有 capacity；Issue 有 estimate。
- **步骤**：将 Issue 划入 Cycle；超额加入。
- **期望**：Issue 同时只属一个 Cycle；估时合计 vs 容量可见过载提示。
- **失败归属**：BOARD；FE/展示；与工时报表达通则联查 TIME。

### 3.6 Cycle 结束

- **前置**：Cycle 含未完成 Issue。
- **步骤**：将 Cycle 标为 completed。
- **期望**：默认未完成项回 Backlog 并记活动（或按配置留原状，但行为明确不可静默）。
- **失败归属**：BOARD。

### 3.7 Done 需人确认

- **前置**：Bot 推进 Issue 至接近完成。
- **步骤**：Bot 尝试直接 Done；人确认 Done。
- **期望**：默认 Bot **不得**独自落到 `done`（或进入 `in_review`）；人确认后 Done（COORDINATION 约定）。
- **失败归属**：BOARD / BOT。

### 3.8 评论与活动流

- **前置**：Issue 存在。
- **步骤**：人与 Bot 分别评论（`.../issues/{id}/comments` · activities）。
- **期望**：列表可区分 actor；时间按 Asia/Shanghai 展示。
- **失败归属**：COLLAB；FE。

> **附件 P0（评论 + Issue 图片/文档）**：最小可勾选验收见 [`ACCEPTANCE_UX_BOT.md`](./ACCEPTANCE_UX_BOT.md) §C（F1–F10）；本清单 §3 暂不展开逐步用例，关单以 UX 文档为准。

---

## 4. 依赖成环

路由：`POST|DELETE /api/v1/workspaces/{ws}/issues/{id}/relations`（COORDINATION 预留）。

### 4.1 合法依赖

- **前置**：Issue A、B。
- **步骤**：创建 `A blocks B`（或 `blocked_by` / `related` / `parent`）。
- **期望**：关系落库；详情/看板可可视化阻塞。
- **失败归属**：BOARD；可视化 FE。

### 4.2 成环拒绝

- **前置**：已有 A→B→C 阻塞链。
- **步骤**：再加 C→A（或任意回边成环）。
- **期望**：HTTP 4xx（建议 409 conflict）；中文提示「依赖成环」；不落库。
- **失败归属**：BOARD（环检测）；FE 展示错误。

### 4.3 自环与重复

- **前置**：Issue A。
- **步骤**：A blocks A；重复创建同向同类型关系。
- **期望**：拒绝；唯一约束生效。
- **失败归属**：BOARD。

### 4.4 父任务与子任务

- **前置**：parent 关系或 `parent_id`。
- **步骤**：父任务标 Done 时仍有未完成子任务。
- **期望**：默认警告（或可配阻止）；行为与文案明确。
- **失败归属**：BOARD；FE 确认对话框。

---

## 5. Wiki / 知识库

路由：`/api/v1/workspaces/{ws}/wiki/spaces` · `.../wiki/pages` · `.../wiki/search?q=`。

### 5.1 Space 与页面 CRUD

- **前置**：工作区成员。
- **步骤**：创建 wiki space；创建/编辑/移动页面（树形 `parent_id`）；Markdown 正文。
- **期望**：按 space slug + page slug 唯一；更新人/时间正确。
- **失败归属**：WIKI；编辑器 FE。

### 5.2 关联项目 / Issue

- **前置**：已有 project、issue。
- **步骤**：页面绑定 `project_id` / `issue_id`。
- **期望**：从 Issue 侧可跳到相关 Wiki；权限随工作区。
- **失败归属**：WIKI；跳转 FE。

### 5.3 搜索

- **前置**：多页面标题/正文。
- **步骤**：`GET .../wiki/search?q=`。
- **期望**：命中当前工作区页面；无跨租户。
- **失败归属**：WIKI。

### 5.4 非成员不可读

- **前置**：见 2.1。
- **步骤**：非成员读 wiki API。
- **期望**：403/404。
- **失败归属**：WIKI / TL。

---

## 6. 工时与计费

路由：`.../time-entries` · `.../time-entries/report` · `.../time-entries/export.csv` · `.../billing-rates`。

### 6.1 登记工时

- **前置**：Issue + 成员。
- **步骤**：创建 `time_entries`（hours>0、worked_at、note、billable）。
- **期望**：落库；Issue/成员维度可汇总。
- **失败归属**：TIME；录入 UI FE。

### 6.2 非法工时

- **前置**：同上。
- **步骤**：hours≤0 或缺失 worked_at。
- **期望**：400；不落库。
- **失败归属**：TIME。

### 6.3 费率与汇总

- **前置**：`billing_rates` 已配置。
- **步骤**：按成员/项目/时间范围汇总工时与费用（currency 默认 CNY）。
- **期望**：汇总正确；时区按 Asia/Shanghai 解释 worked_at 展示。
- **失败归属**：TIME；报表 UI FE。

### 6.4 权限

- **前置**：Member vs Admin。
- **步骤**：改他人工时、改费率。
- **期望**：仅本人可改自己的条目（或 Admin 可改）；费率 Admin-only（若产品如此约定）。
- **失败归属**：TIME。

---

## 7. Bot 记忆 / 角色 / Handoff

路由（COORDINATION 预留 · Bot continuity）：

- `.../bot-roles` · `.../bot-roles/{id}/claim` · `.../release`
- `.../bots/{memberID}/memories` · search
- `.../bots/{memberID}/todos`
- `GET .../bots/{memberID}/handoff`

### 7.1 Bot 角色 CRUD

- **前置**：Admin 或授权成员。
- **步骤**：创建 `bot_roles`（name/slug/responsibilities/skills/默认项目）。
- **期望**：workspace 内 slug 唯一；可更新。
- **失败归属**：BOT。

### 7.2 角色认领（Claim）与释放

- **前置**：角色存在；Bot member 存在。
- **步骤**：Bot 认领角色；另一 Bot 再认领同一角色；原 Bot 释放后再认领。
- **期望**：同一角色同时仅一个有效 claim（`released_at IS NULL`）；冲突明确错误。
- **失败归属**：BOT。

### 7.3 记忆读写

- **前置**：Bot 已鉴权（API Key）。
- **步骤**：写入 `bot_memories`（key_name/value_text/tags）；读取/更新同 key；按 tags 检索。
- **期望**：`(workspace, member, key)` 唯一；仅所属 Bot/授权方可读写。
- **失败归属**：BOT。

### 7.4 Handoff（换机续工）

- **前置**：Bot A 持有角色 + 记忆 + `bot_todos`。
- **步骤**：A 释放角色并写 handoff 摘要到 memory/todo；Bot B 认领同角色并拉取 `GET .../handoff` / memory/todos。
- **期望**：B 能续上未完成 todo 与关键记忆；活动流记录 handoff；无跨工作区泄漏。
- **失败归属**：BOT。

### 7.5 Bot 待办

- **前置**：Bot member。
- **步骤**：创建/推进 `bot_todos`（todo/doing/done/canceled）；可选关联 issue。
- **期望**：状态机合法；排序 `sort_order` 可用。
- **失败归属**：BOT。

### 7.6 Webhook 与幂等

- **前置**：配置 webhook；Bot 订阅 issue/comment 等事件。
- **步骤**：变更 Issue；Bot 用相同 `Idempotency-Key` 重放创建。
- **期望**：Webhook 签名校验；投递记录可查（`.../webhooks/{id}/deliveries`）；幂等键返回首次响应、不重复建资源。
- **失败归属**：INTEG（webhook）；写幂等各域 Handler + BOT。

---

## 8. 横切：运维与体验

### 8.1 健康检查

- **步骤**：`GET /healthz` 与 `GET /api/v1/healthz`。
- **期望**：200；含 `status` / `db` / `timezone: Asia/Shanghai`。
- **失败归属**：TL。

### 8.2 种子幂等

- **步骤**：连续两次 `make seed`（或两次带 `AUTO_SEED` 启动）。
- **期望**：不报错、不重复脏数据（或明确 upsert）。
- **失败归属**：TL（`internal/seed`；各域追加 seed 须幂等）。

### 8.3 配置与密钥

- **步骤**：检查日志与错误响应。
- **期望**：无密码、无完整 API Key；配置来自环境变量。
- **失败归属**：TL / INTEG。

### 8.4 前端空态与中文错误

- **步骤**：空项目、失败表单、网络错误。
- **期望**：空状态引导；Toast/文案中文可读。
- **失败归属**：FE。

---

## 9. 当前仓库对照（本次 QA 核实）

| 项 | 状态（2026-10-24 核实） |
|----|-------------------------|
| `COORDINATION.md` | **存在于仓库根**；定义目录所有权与 `/api/v1` 已实现/预留表 |
| `cmd/server` / `cmd/seed` | **有** `main.go`；`go build ./cmd/server` **通过** |
| `handlers` | **有** `auth.go` `health.go` `workspace.go`（地基）；域 handlers 多未落地，部分路由 501 stub |
| `services/wiki` · `services/timetrack` | **有**服务层实现（handler 挂载/路由仍待 Wiki/Time Bot） |
| `frontend/` | Vite + React + AntD；关键：`src/pages/LoginPage.tsx` `HomePage.tsx` `api/client.ts`；`layouts/` 目录空；board/wiki 等页多为占位目录 |
| `docs/` | `ACCEPTANCE.md` `API.md` `BOT.md`；分工见根 `COORDINATION.md` |
| `schema.sql` | 已覆盖 users/issues/relations/wiki/time/bot_* 等 |
| auth/middleware/config/db | 已串入 `cmd/server`；JWT + cookie；API Key 中间件已备 |
| `*_test.go` | **有** `internal/auth/jwt_test.go` `password_test.go`；`go test ./... -count=1` **通过** |
| 旧已知问题 `middleware/auth.go` undefined `marshalJSON` | **本次未复现**；build/test 均绿 |

**结论**：

- **§1 登录与会话冒烟：可做**（healthz / login / me / workspaces 实测 200；seed 账号可用）。
- §2–§7 多数域 API 仍为预留/501 或仅有 service 未挂路由 → **端到端全量尚不可**；待各 Bot 按 `COORDINATION.md` 替换 stub 后按 §1→§7 顺序扩测。

---

## 10. QA 执行顺序建议

1. `make install && make seed && make dev`（演示需 `SEED_SAMPLE_DATA=1`；生产 `AUTO_SEED=0`）。
2. API smoke：§1 登录 → §2 权限 → §3 看板 → §4 成环 → §5 Wiki → §6 工时 → §7 Bot。
3. `go test ./...` 与前端 `npm run build` 纳入 CI/发版门禁。
4. 失败单条记录：编号 + 实际结果 + 归属（TL/BOARD/COLLAB/INTEG/WIKI/TIME/BOT/FE）+ 复现 curl/截图。
5. 跨模块改动先同步 `COORDINATION.md` 或 PR 说明。
