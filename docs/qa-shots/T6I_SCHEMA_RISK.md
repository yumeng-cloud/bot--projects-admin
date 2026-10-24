# T6-I · Schema 加列风险（`issues.team_id`）

> 2026-10-24 Asia/Shanghai · Issue 域旁路备忘（**不改** `internal/db/**`）  
> Go / Tech lead 拥有迁移与 schema；本文只列双驱动坑与调用方风险，供落地对照。  
> 契约：`docs/API.md`「Issue · team_id」· 开工：`docs/NEXT_WAVE_DEBT_T6I.md`

## 目标形状

| 项 | 建议 |
|----|------|
| 列 | `issues.team_id TEXT` **可空** |
| FK | `REFERENCES teams(id)` |
| ON DELETE | **`SET NULL`**（删队 → 旧单变未分队；勿 `CASCADE` 误删 Issue） |
| 索引 | `CREATE INDEX … ON issues(workspace_id, team_id)`（list/board 按队） |
| 回填 | **本包默认不回填**；存量保持 `NULL`=未分队；新建 API 必写当前队 |
| 乐观锁 | **不动** `version` / board_order 语义 |

对照 pins：`team_pins.team_id … ON DELETE CASCADE`（钉随队死）。Issue 池相反——队没了单还在 → **SET NULL**。

---

## 本仓 dual-driver 现状（只读）

| 驱动 | 入口 | 加列方式 |
|------|------|----------|
| **SQLite** | `db.Open` → **始终**跑 embed `schema.sql` | 新库：`CREATE TABLE issues` 可带裸 `team_id TEXT`（`teams` 表在文件后部才建 → **CREATE 时不宜 REFERENCES teams**）；旧库：文件尾 `ALTER TABLE issues ADD COLUMN team_id … REFERENCES teams(id) ON DELETE SET NULL`（`teams` 已存在） |
| **Postgres** | `AUTO_MIGRATE=1` → 按名排序跑 `postgres/*.sql`；否则依赖运维 / 回落 `schema_postgres.sql` | 编号迁移例：`010_issues_team_id.sql`：`ADD COLUMN IF NOT EXISTS` + index；**应用角色常无 ALTER 权** → 空库需 `pg-bootstrap.sh` 等（见 `migrateHint`） |

`toleratesMigrateError`：**仅**吞 SQLite `duplicate column name`（及部分 extension 权限）。**不**吞「重复索引」以外的硬错；`CREATE INDEX IF NOT EXISTS` 可幂等。

---

## 风险清单

### 1. CREATE 在 teams 之前挂 FK

`schema.sql` 里 `issues` 远早于 `CREATE TABLE teams`。若在 CREATE issues 写 `team_id … REFERENCES teams(id)` → SQLite/PG 新建库直接炸。

**口径**：CREATE 用可空裸列（或完全不加，只靠尾部 ALTER）；**FK 放在 teams 存在之后的 ALTER / `010_*.sql`**。应用层 `ensureTeamInWorkspace` 仍要拦脏写。

### 2. SQLite 尾部 ALTER 幂等

新库 CREATE 已含 `team_id` → 尾部 `ADD COLUMN team_id` 触发 `duplicate column name` → **依赖** `toleratesMigrateError`。  
若有人关掉容忍或换执行器 → Open 失败。

**口径**：保持与 `start_at`/`due_at` 同一模式；注释写明「旧库 ALTER / 新库 duplicate 可忽略」。

### 3. Postgres：`IF NOT EXISTS` vs 单片 monolith

- 编号迁移：`ADD COLUMN IF NOT EXISTS`（PG 现代版本）较安全。  
- 若有人把 monolith `schema_postgres.sql` 的 CREATE 与 `010` 混跑：列已在 CREATE → `010` 应 IF NOT EXISTS；索引双写靠 IF NOT EXISTS。  
- **无 AUTO_MIGRATE 的现网**：只改 embed 文件**不会**自动加列 → 必须交付可跑的 `010_*.sql` + 运维说明。

### 4. ON DELETE 选错

| 策略 | 后果 |
|------|------|
| `CASCADE` | 删 team 抽干该队全部 Issue → **事故级** |
| `SET NULL` | 单变未分队；符合「NULL=未分队」 |
| 无 FK | 队删后残留幽灵 UUID；list 按队「消失」、详情仍在 |

**建议 SET NULL + 应用层禁「清空 team_id」普通 PATCH**（管理未分队另议）。

### 5. 回填 vs 不回填

| 策略 | 利 | 弊 |
|----------|----|
| 不回填（推荐本包） | 迁移一行；产品「未分队」滤镜有意义 | 默认队 list 变「空池感」直到新单写入 |
| 挂默认队 | 切队立刻有活 | 需定义「默认队」；误挂难撤；双队 WS 无唯一默认 |

T6I：「证据写明」。**Issue 域不擅自 UPDATE 回填**。

### 6. 调用方漏传 `team_id`

| 路径 | 风险 |
|------|------|
| `CreateIssue` 旧单测 / seed / bot expand | 列落地后 **必填** → 大面积 4xx；需同步改 `issue_test.go`、`botmem` expand、handlers |
| `ListIssues` / search / board 不传 `TeamID` | 若服务层「空 = 全 WS」→ **串队回归**；产品要「默认当前队」时 handler 必须注入当前 team |
| `Get`/`Patch` 不带队上下文 | 错队 404 **测不到**；隔离只在 `?team_id=` / `RequireTeamID` 路径生效 |
| 直接 SQL INSERT 省略列 | SQLite/PG 写入 NULL → 变成未分队旧单形态 |

### 7. `version` / 乐观锁

加列、回填、SET NULL **都不应** `version=version+1`（除非产品要把「换队」当一次用户编辑）。  
`UpdateIssue` 换 `team_id` 若走正常 PATCH，会抬 version——那是业务更新，不是迁移。  
**迁移 SQL 禁止**碰 `version`、`board_order`、`updated_at`（除非审计明确要求）。

### 8. 扫描列清单漂移

`issueCols` / `SELECT` 列表增 `team_id` 必须与 `scanIssue` 同步；漏扫 → 错位或 `Scan` 失败。  
Board/search 若手写 SQL 不经 `ListIssues`，同样要带过滤。

### 9. 索引重复定义

`schema.sql` 可能在 CREATE 后早期 `CREATE INDEX idx_issues_ws_team` **又**在尾部 ALTER 后再建一次——`IF NOT EXISTS` 可接受；改名时两处都要改。

### 10. Issue 域边界（硬）

| 允许 | 禁止 |
|------|------|
| 测草、风险文、落地后 `*_test.go` | 编辑 `internal/db/schema*.sql`、`postgres/010_*.sql` |
| 评审 Go 的 filter SQL 是否符合「默认不含 NULL」 | 改 auth / botmem pins / T7 |

---

## 建议验收（给 Go，非本包代做）

1. 新鲜 SQLite `Open` 两次：列存在、FK pragma 下删 team → issue.team_id 变 NULL。  
2. PG `AUTO_MIGRATE=1`：`010` 幂等二跑。  
3. 存量 SQLite（无列）→ Open 后 ALTER 成功。  
4. Create 无 team_id → 4xx；List `team_id=A` **不含** NULL；`unassigned=1` 仅 NULL。  
5. 错队 GET/PATCH → 404。  
6. 并发/乐观锁旧测仍绿（`version` 行为不变）。

---

## 一句话给 Go

> 可空 `team_id` + **SET NULL** + `(workspace_id, team_id)` 索引；SQLite 用「CREATE 裸列 / 尾 ALTER+FK」双段；PG 用 `010 IF NOT EXISTS`；**不回填**；**别动 version**；默认 list **排除 NULL**（与 API.md 对齐），未分队走显式滤镜。

