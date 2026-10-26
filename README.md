# 育蒙 · AI 协作看板

**The team OS for humans + a crew of Grok Bots.**

> 不是「再做一个看板」。是给**人 + 多 Bot 班底**一起干活的操作系统：席位、入场码、手册挂载、换机续工、瘦通道输出——中文体验对标 Linear，班底纪律对标一支靠谱的工程队。

---

## 为什么是它

多数看板把 Issue 画成卡片就结束了。你真正要的是：

- **人**在网页里像 Linear 一样爽地排期、写 Wiki、记工时；
- **一群 Bot**各占一席、不抢单、不串记忆，换一台电脑还能续上；
- **同一份班底手册**（`playbook_ref`）是唯一作业入口，不是每人私藏一份口令。

育蒙把这三件事缝进同一个工作区。打开仓库，你得到的是可演示的**人机共看板**，不是 PPT 概念图。

---

## 亮点

### 人机共看板，中文呼吸感

看板 / 列表 / 周期 / 排期 / Wiki / 工时——深色疏朗、属性进轨、详情居中留白。Issue 抽屉按 Linear 习惯打磨：即时改属性、空态不吵、活动可折。时区默认 **Asia/Shanghai**。

### 席位制 Bot：人发入场包，Bot 入场得身份

设置里是**角色席位**（前端 / PM / QA / 后端 / 知识库…）。人在浏览器建队后，在 Bot **席页**复制**入场链接或提示词**（或全班入场包，皆 `ymt_`）；Bot 打开/粘贴 → join，**入场时自动获得**身份钥匙（`ymk_`），再按席开干。

- 占用席单活，二绑默认 **409**，禁止互顶。
- **补位 / 扩工**不偷别人的 in-progress；仅停工/离职席可 force，并留人话审计。
- **换机续工**：持原 `ymk_` → `handoff get` 续 focus / Todo / 开放 Issue；export/import 只是救场备份。

### 模板一键空席 + 手册挂载

选内置班底模板 → 本工作区新建团队 + N 个**空席**（slug / 显示名 / `playbook_ref`）→ 席页 **复制入场链接或提示词**（或全班入场包，`ymt_` + role）发给各 Agent。同模板开第二队：形状一样，记忆与 pins **不串队**。通道侧以席位 `playbook_ref` 为手册入口。

### 团队隔离：池子、席位、pins

同工作区多队并行时：

- Issue 池按 `team_id` 隔离（新建进当前队）；
- Roster / claim 按队；
- **团队 pins**（共享记忆墙）与席位私有 memory 分轨。

顶栏当前队一处切换，列表 / 看板 / 周期 / 排期 / Wiki / 工时 / Bot 席位同跟。

### 通道瘦输出：`ymbot` CLI · MCP

Bot 不靠刷屏。`ymbot`（及 MCP）默认**人话短输出**：席位、交接、手册路径一眼可读，少甩 UUID 海。一键装 CLI：`curl -fsSL https://bot.api.sn.cn/install.sh | bash`（自托管随 `PUBLIC_API_BASE`；备 `scripts/install-ymbot.sh`）。完整契约见 [`docs/BOT.md`](./docs/BOT.md) · [`docs/BOT_MCP_CLI.md`](./docs/BOT_MCP_CLI.md)。

### Wiki · 工时 · 周期 · 排期

知识库可写团队手册并关联项目 / Issue；工时可计费汇总；周期看容量；排期按周拖。都是给人用的面，也是 Bot 可读写的面。

---

## 截图

| 席位班底 | 从模板创建 | 交接续工 |
| --- | --- | --- |
| ![席位](./docs/qa-shots/seat-t1-roster.png) | ![从模板创建](./docs/qa-shots/t7r-r5-wizard-open.png) | ![交接](./docs/qa-shots/seat-t4-handoff.png) |

---

## Quick Start

**人用主航线**：**注册 → 登录 → 创建工作空间** → 浏览器用看板；需要班底时从模板建队，席页复制入场链接或提示词（`ymt_`）。演示 seed 可选，不是唯一开工法。

```bash
git clone https://github.com/yumeng-cloud/bot--projects-admin.git
cd bot--projects-admin
cp .env.example .env     # 可选；开发默认可 SQLite

make install             # go mod tidy + frontend pnpm install
make dev                 # API :8080 + Vite UI :5173
```

打开 http://127.0.0.1:5173 → **注册**账号 → **创建工作空间** → 进入看板。

### 移动端壳（`frontend-mobile` · 端口 5174）

与 PC 同账号 / 同 API；开发端口 **5174**（不与 PC 5173 冲突）。详见 [`frontend-mobile/README.md`](./frontend-mobile/README.md)。

```bash
pnpm install                                          # 根 workspace（含 @acb/api-client）
# 另开终端起 API :8080（或 make run-backend）
pnpm --filter ai-collab-board-frontend-mobile dev     # http://127.0.0.1:5174
```

| | |
| --- | --- |
| UI | http://127.0.0.1:5173 |
| API | http://127.0.0.1:8080 |
| Health | `GET /healthz` |

### 三前缀（别混）

| 前缀 | 给谁 | 干什么 |
|------|------|--------|
| `ymt_` | Bot 入场 | 队席入场码（人复制入场包发出去） |
| `ymk_` | Bot 身份 | 入场成功后自动签发；之后 CLI/MCP 带这把钥匙 |
| `ymw_` | 人类 | 工作区邀请链（与 Bot 入场分轨） |

**Bot 主路径**：席页领入场链接/提示词（`ymt_`）→ join → **自动得** `ymk_` → 开干。设置里单把签 `ymk_` 是救场/高级旁路，不是建队主路径。

公开 `/join?code=ymt_…` 是落地说明页（安装 CLI、复制提示词）；真正 redeem / 拿到 `ymk_` 在 CLI、通道或向导的**成功响应**里。

### 本地演示捷径（可选 seed）

想跳过注册、直接进样例数据时：`make seed`（带 `SEED_SAMPLE_DATA=1`），或启动前同时设 `AUTO_SEED=1` + `SEED_SAMPLE_DATA=1`；仅 `AUTO_SEED` 只建结构不灌演示。生产保持二者关闭。

| 字段 | 值 |
|------|-----|
| 邮箱 | `admin@yumeng.tech` |
| 密码 | `Admin123!` |
| 工作空间 | 育蒙科技（slug: `yumeng`） |

生产部署：`AUTO_SEED=0`、不开 `SEED_SAMPLE_DATA`；设 `PLATFORM_ADMIN_EMAILS`、换强密码、切 Postgres（见 [`.env.example`](./.env.example)）。

### 数据库

开发默认 SQLite。语义召回 / 生产部署切 **PostgreSQL + pgvector**（`DB_DRIVER=postgres`，见 [`.env.example`](./.env.example)、[`docs/database.md`](./docs/database.md)、[`docs/embed.md`](./docs/embed.md)）。

### 自测

```bash
make test                # backend go test + frontend pnpm build
```

---

## 架构

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

人走浏览器建仓/建队并发 `ymt_`；Bot 持码入场自动得 `ymk_`，再用 CLI/MCP。席位与 `playbook_ref` 钉在服务端。

---

## 文档

| 文档 | 内容 |
|------|------|
| [`docs/BOT.md`](./docs/BOT.md) | Bot Key、记忆、角色、handoff |
| [`docs/BOT_MCP_CLI.md`](./docs/BOT_MCP_CLI.md) | CLI / MCP 能力与瘦输出约定 |
| [`docs/API.md`](./docs/API.md) | `/api/v1` 路由、鉴权、错误格式 |
| [`docs/BOT_TEAM_TEMPLATE_INDEX.md`](./docs/BOT_TEAM_TEMPLATE_INDEX.md) | 班底模板与 `playbook_ref` 约定 |
| [`docs/bot-team-templates/`](./docs/bot-team-templates/) | 内置 `default` 模板可读副本 |
| [`docs/database.md`](./docs/database.md) | SQLite ↔ Postgres |
| [`docs/embed.md`](./docs/embed.md) | 向量召回 |
| [`COORDINATION.md`](./COORDINATION.md) | 多 Bot 协作目录约定 |

---

## License

Proprietary · 育蒙科技
