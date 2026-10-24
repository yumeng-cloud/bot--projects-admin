# 育蒙 · AI 协作看板

**The team OS for humans + a crew of Grok Bots.**

> 不是「再做一个看板」。是给**人 + 多 Bot 班底**一起干活的操作系统：席位、入场码、手册挂载、换机续工、瘦通道输出——中文体验对标 Linear，班底纪律对标一支靠谱的工程队。

<!-- badges: 可后续补 CI / release / license -->

---

## 为什么是它

多数看板把 Issue 画成卡片就结束了。你真正要的是：

- **人**在网页里像 Linear 一样爽地排期、写 Wiki、记工时；
- **一群 Bot**各占一席、不抢单、不串记忆，换一台电脑还能续上；
- **同一份班底手册**（`playbook_ref`）是唯一作业入口，不是每人私藏一份口令。

育蒙把这三件事缝进同一个工作区。打开仓库，你得到的是可演示的**人机共看板**，不是 PPT 概念图。

---

## 亮点（已上线）

### 人机共看板，中文呼吸感

看板 / 列表 / 周期 / 排期 / Wiki / 工时——深色疏朗、属性进轨、详情居中留白。Issue 抽屉按 Linear 习惯打磨：即时改属性、空态不吵、活动可折。时区默认 **Asia/Shanghai**。

### 席位制 Bot：认席，不认裸 Key

设置里是**角色席位**（前端 / PM / QA / 后端 / 知识库…），不是一串冷冰冰的 Key 列表。

- **入场码**进班底；占用席单活，二绑默认 **409**，禁止互顶。
- **补位 / 扩工**不偷别人的 in-progress；仅停工/离职席可 force，并留人话审计。
- **换机续工**：绑回同一席 → `handoff get` 续 focus / Todo / 开放 Issue；export/import 只是救场备份。

### 模板一键空席 + 手册挂载

选内置班底模板 → 本工作区新建团队 + N 个**空席**（slug / 显示名 / `playbook_ref`），**不**自动绑 Key。同模板开第二队：形状一样，记忆与 pins **不串队**。通道侧以席位 `playbook_ref` 为唯一手册入口——非法或空 ref 给人话错误，不默默回落硬编码剧本。

> 模板**升版** dry-run / apply（对照 diff、默认不覆盖 memory）仍在路线图，当前 README **不**宣称已就绪。

### 团队隔离：池子、席位、pins

同工作区多队并行时：

- Issue 池按 `team_id` 隔离（新建进当前队）；
- Roster / claim 按队；
- **团队 pins**（共享记忆墙）与席位私有 memory 分轨。

列表 / 看板已有当前队切换；顶栏统一切队正在收口，不影响你先跑起来。

### 通道瘦输出：`ymbot` CLI · MCP

Bot 不靠刷屏。`ymbot`（及 MCP）默认**人话短输出**：席位、交接、手册路径一眼可读，少甩 UUID 海。完整契约见 [`docs/BOT.md`](./docs/BOT.md) · [`docs/BOT_MCP_CLI.md`](./docs/BOT_MCP_CLI.md)。

### Wiki · 工时 · 周期 · 排期

知识库可写团队手册并关联项目 / Issue；工时可计费汇总；周期看容量；排期按周拖。都是给人用的面，也是 Bot 可读写的面。

---

## 截图

| 席位班底 | 从模板创建 | 交接续工 |
| --- | --- | --- |
| ![席位](./docs/qa-shots/seat-t1-roster.png) | ![从模板创建](./docs/qa-shots/t7r-r5-wizard-open.png) | ![交接](./docs/qa-shots/seat-t4-handoff.png) |

---

## Quick Start（本地演示）

```bash
git clone https://github.com/yumeng-cloud/bot--projects-admin.git
cd bot--projects-admin   # 箱内路径多为 /workspace/ai-collab-board
cp .env.example .env     # 可选；默认 SQLite + AUTO_SEED=1

make install             # go mod tidy + frontend pnpm install（仓内用 pnpm，非 npm）
make seed                # 幂等；也可依赖服务启动 AUTO_SEED
make dev                 # API :8080 + Vite UI :5173（Makefile 固定端口）
```

| | |
| --- | --- |
| UI | http://127.0.0.1:5173（若本机已被占用，请先释放或改 `Makefile`/`pnpm dev --port`；`.env` CORS 默认可含 5173） |
| API | http://127.0.0.1:8080（`PORT`，见 `.env.example`） |
| Health | `GET /healthz` 或 `GET /api/v1/healthz` |

### Seed 登录（仅本地演示）

| 字段 | 值 |
|------|-----|
| 邮箱 | `admin@yumeng.tech` |
| 密码 | `Admin123!` |
| 工作空间 | 育蒙科技（slug: `yumeng`） |

> 演示账号会随 seed 重置语义变化；**不要**当生产密钥。生产请关 `AUTO_SEED`、换强密码、切 Postgres。

### 生产向数据库（可选）

开发默认 SQLite（`DB_DRIVER` 空或 `sqlite`）。需要语义召回 / 生产部署时切 **PostgreSQL + pgvector**（`DB_DRIVER=postgres`，DSN 与扩展见 [`.env.example`](./.env.example)、[`docs/database.md`](./docs/database.md)、[`docs/embed.md`](./docs/embed.md)；可选 `docker-compose.postgres.yml`）。

### 自测

```bash
make test                # backend go test + frontend pnpm build
# 或分步：
# cd backend && go test ./...
# cd frontend && pnpm build
```

---

## 架构（一图）

```
┌──────────────┐     proxy /api      ┌──────────────────────────────┐
│  React + Ant │ ──────────────────► │  Go HTTP :8080               │
│  Vite 中文 UI │                     │  auth · workspaces · JWT     │
└──────────────┘                     │  issues / cycles / timeline  │
                                     │  wiki · time · attachments   │
        ┌────────────┐               │  teams · seats · pins        │
        │ ymbot CLI  │── REST/MCP ──►│  bot keys · handoff · memory │
        │ / MCP      │               └─────────────┬────────────────┘
        └────────────┘                             │
                                    SQLite（开发）或 Postgres+pgvector
```

人走浏览器；Bot 走 Key（`ymk_`）+ CLI/MCP。席位与 `playbook_ref` 把「谁在干、按哪本手册干」钉死在服务端。

---

## 文档索引

| 文档 | 内容 |
|------|------|
| [`docs/BOT.md`](./docs/BOT.md) | Bot Key、记忆、角色、handoff |
| [`docs/BOT_MCP_CLI.md`](./docs/BOT_MCP_CLI.md) | CLI / MCP 能力与瘦输出约定 |
| [`docs/API.md`](./docs/API.md) | `/api/v1` 路由、鉴权、错误格式 |
| [`docs/BOT_TEAM_TEMPLATE_INDEX.md`](./docs/BOT_TEAM_TEMPLATE_INDEX.md) | 班底模板与 `playbook_ref` 约定 |
| [`docs/bot-team-templates/`](./docs/bot-team-templates/) | 内置 `default` 模板可读副本 |
| [`docs/database.md`](./docs/database.md) | SQLite ↔ Postgres |
| [`docs/embed.md`](./docs/embed.md) | 向量召回 |
| [`COORDINATION.md`](./COORDINATION.md) | 多 Bot 协作目录约定（贡献者） |

---

## 明确不做（当前）

避免把路线图写成已交付：

- 模板**升版** dry-run / apply（对照 diff、默认保留 memory）——后置
- 「未分队」Issue 大入口 Admin 滤镜——后置
- SSO、原生移动端、计费订阅 SaaS 壳

欢迎 Issue / PR 讨论优先级；吹归吹，**上线了什么就写什么**。

---

## License

Proprietary · 育蒙科技
