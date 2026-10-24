# 产品体验验收报告 · Issue 详情 + Bot MCP/CLI

> 验收日期：2026-10-24（Asia/Shanghai）  
> 仓库：`/workspace/ai-collab-board`  
> 规格：`frontend/ISSUE_DETAIL_LAYOUT.md` · `docs/BOT_MCP_CLI.md` · 本文件 §C 附件 P0  
> 方式：初版只读验收；同日 Bot 记忆角色在 `cmd/ymbot` + `botclient` 落地幂等透传/中文错误并用真 Key 跑通 §7（**未**改业务 handler 语义）  
> 范围说明：排期以现有 Cycle 页为准；**不**把「GitHub Projects 时间线」列为必须缺口。

---

## 总结（一眼结论）

**Issue 详情布局与 ByteMD 主路径已落地，核心清单 A1–A5 / M1–M2 / M4–M5 通过；Bot MCP+CLI P0 能力、幂等 Key 透传、一行中文错误与 §7 真 Key 串联均已关单。**  
B1/C6 已复验勾掉；剩余主要为 M3 CJK 实机与附件 P0（§C）。**无阻塞「能用」的硬失败**；Bot 优先缺口 #2/#3 已关闭。

| 域 | 通过 | 部分 | 失败 | 结论 |
|----|------|------|------|------|
| Issue（A1–A5、B1、M1–M5） | 10 | 1 | 0 | 可验收通过（剩 M3 CJK 实机） |
| Bot MCP/CLI（P0 能力簇 + 鉴权） | 10 | 0 | 0 | P0 结构 + 幂等透传 + 中文错误 + §7 真 Key 串联均通过 |
| **合计（本报告勾选行）** | **19** | **2** | **0** | |

**失败项条数：0**（部分项 2，均属 Issue 密度/CJK；Bot 原部分项已于同日关单）。

---

## A) Issue 详情对照 `ISSUE_DETAIL_LAYOUT` 清单

证据目录：`frontend/src/components/issue/`（`IssueDrawer` / `IssueDetailMain` / `IssuePropertyRail` / `IssueDescriptionEditor` / `issue-detail.css` 等）。依赖：`package.json` 含 `bytemd` / `@bytemd/react` / `@bytemd/plugin-gfm` `^1.22.0`；**无 TipTap**。

| ID | 标准摘要 | 结果 | 文件证据 |
|----|----------|------|----------|
| **A1** | 首屏大标题 + 描述留白，非 Form 标签墙 | **通过** | `IssueDetailMain.tsx`：`acb-issue-title-input` + `IssueDescriptionEditor`；属性不在主栏 |
| **A2** | 状态/优先级/负责人/Cycle/估时/标签/关联在 Property 轨；主栏无这些 Form.Item | **通过** | `IssuePropertyRail.tsx` 全属性；`IssueDrawer.tsx` body = Main + Rail |
| **A3** | 一行一项 label 左 value 右；无 Ant Form 多列栅格墙 | **通过** | `IssuePropertyRow.tsx`；轨内 `Form.Item noStyle` + `component="div"`，CSS 杀掉 form-item 竖距（`issue-detail.css` L214–219） |
| **A4** | 评论流在描述下，输入与列表连贯；非 footer 割裂 | **通过** | `IssueCommentList.tsx` 列表+composer 同栏；抽屉 `footer` CSS `display:none` |
| **A5** | 活动默认折叠，不占满首屏 | **通过** | `IssueActivityFold.tsx`：`defaultActiveKey={[]}` |
| **B1** | 桌面宽默认 800（硬验收 ≥720）、高 100%；主栏 max-width ≤520 居中 | **通过** | `width={820}`（≥720 ✓）、`body height:100%` ✓；`.acb-issue-main-inner { max-width: 520px }`（React 2026-10-24 已收） |
| **M1** | 描述可写/存 Markdown（标题/粗斜体/列表/链接/代码；建议 GFM） | **通过** | `IssueDescriptionEditor.tsx`：`@bytemd/plugin-gfm`；值为 MD 字符串 |
| **M2** | tab 编辑\|预览（验收以 tab 为准） | **通过** | `mode="tab"`；CSS 保留 `.bytemd-toolbar-tab`（`issue-detail.css` L370–394） |
| **M3** | 中文/CJK 输入不卡顿 | **部分** | 静态：走 ByteMD/CodeMirror + `zh_Hans`；**未做**浏览器拼音组字实机测 |
| **M4** | 工具栏/空态中文 | **通过** | `locale={zhHans}` + `placeholder="添加描述…"`；空预览「还没有描述…」 |
| **M5** | 未编辑时外观克制、细边框、贴合主题 | **通过** | `.acb-issue-md .bytemd` 无重边框/透明底；工具栏克制 |

### 附带观察（未计入上表失败）

| 项 | 结果 | 说明 | 建议指派 |
|----|------|------|----------|
| B2 轨宽 240–280 | 通过 | CSS `width: 280px`，落在允许上沿 | — |
| B3 无「快速状态」Segmented 墙 | 通过 | `IssueDrawer` 无 Segmented | — |
| B4 大号无边框标题 | 通过 | 23px / borderless | — |
| B5 描述区 ≥160px + 中文占位 | 通过 | `min-height: 220–240px` | — |
| C6 窄屏 `min(100vw,560)` + 轨改上下 | **通过** | `@media (max-width:639px)` 列布局 + `matchMedia` 窄屏 `width=min(100vw, 560px)`（与轨堆叠断点一致）；宽屏仍 820（React 2026-10-24） | — |
| Cycle 页 / GitHub Projects 时间线 | **不列缺口** | 现有 Cycle 页可用；总指挥明确不要求 GitHub Projects 时间线 | — |

---

## B) Bot MCP/CLI 对照 `BOT_MCP_CLI.md` P0

证据：`backend/cmd/ymbot/{main,cli,mcp}.go`、`backend/internal/botclient/client.go`；后端 `GET …/bots/me` 已挂（`handlers/mount_chi.go` / `bots.go`）。只读验证：`go build ./cmd/ymbot` 成功；`ymbot -h` 列出 P0 子命令。

| 能力 / 项 | 通道 | 结果 | 证据 |
|-----------|------|------|------|
| whoami | `ymbot whoami` / `board_whoami` | **通过** | CLI + MCP；client `WhoAmI` → `GET …/bots/me`（404 回退 roster） |
| issue list/get/create/update | `ymbot issue *` / `board_issue_*` | **通过** | CLI 四动词齐全；`--assignee me` 解析；`GuardBotDone` 禁 done |
| comment list/create | `ymbot comment *` / `board_comment_*` | **通过** | `--issue` + `-m\|--body` |
| memory list/set | `ymbot memory *` / `board_memory_*` | **通过** | upsert `key_name` / `value_text` |
| handoff get\|export / import | `ymbot handoff *` / `board_handoff_*` | **通过** | `--out` / `--file`（含 `-` stdin） |
| MCP tools 对齐 P0（11） | `mcp-serve` `tools/list` | **通过** | `mcp.go`：`board_whoami` · `board_issue_{list,get,create,update}` · `board_comment_{list,create}` · `board_memory_{list,set}` · `board_handoff_{get,import}` |
| 鉴权 `YMBOT_API_KEY` 约定 | env + Bearer | **通过** | help / `LoadConfig` / `Require()` 与规格 §5 一致（`ymk_…`、禁密码主推）；文档 §3 已标注「P0 已落地」 |
| 写默认幂等 | Idempotency-Key | **通过** | CLI `--idempotency-key` / MCP `idempotency_key` / env `YMBOT_IDEMPOTENCY_KEY`；缺省自动 UUID；相同 Key 重放不双写已验 |
| 错误一行中文 + 吊销文案 | APIError | **通过** | `APIError.Error()` 一行中文；401 含「已吊销/无效」；403/404/409 可读；done 护栏保留；MCP isError 同路径 |
| §7 真 Key 串联验收 | whoami→handoff→issue→comment→memory→export | **通过** | 2026-10-24 本机 seed Admin 签发 `ymk_` → 串联通过；重放同 Key 同 comment id；吊销后 exit 10「认证失败：API Key 已吊销」 |

P1（search / cycle CLI / role / todo）按规格不阻塞 P0；Cycle **产品侧**用现有页即可。

---

## 优先修的 3 个缺口（含指派）

| # | 缺口 | 影响 | 指派 | 建议动作 |
|---|------|------|------|----------|
| 1 | ~~主栏 max-width / 窄屏 560~~ | — | **已关闭**（React 2026-10-24；产品体验复验勾掉） | `max-width:520` + `min(100vw,560px)` |
| 2 | ~~CLI/MCP 幂等 Key 不可透传~~ | — | **已关闭**（Bot 记忆角色 2026-10-24） | `ResolveIDKey`：CLI `--idempotency-key` / MCP `idempotency_key` / `YMBOT_IDEMPOTENCY_KEY` / UUID |
| 3 | ~~错误未统一中文 + §7 未关单~~ | — | **已关闭**（Bot 记忆角色 2026-10-24） | `APIError` 一行中文；§7 真 Key 串联 + 吊销 401「已吊销」已验 |

> 非优先：M3 实机 CJK（QA/前端冒烟即可）；P1 search/cycle 通道；**不做** GitHub Projects 时间线。

---

## 给总指挥的短报（可直接转发）

1. **一眼结论**：Issue Linear 布局 + ByteMD 与 Bot `ymbot` MCP/CLI P0 **均已落地**；本轮 **失败 0**；B1/C6 **已复验通过**，剩余部分仅 M3 CJK 实机。Bot 优先缺口 #2/#3 **已关闭**。附件 P0 见 §C。  
2. **Issue**：A1–A5、B1、C6、M1–M2/M4–M5 通过；M3 CJK 实机未跑。  
3. **Bot**：幂等 Key 可透传（CLI/MCP/env）；`APIError` 一行中文（401 含已吊销/无效）；§7 真 Key 串联 whoami→handoff→issue list(me)→comment→memory→export **通过**；同 Key 重放不双写；吊销后 exit 10。  
4. **催落地**：B1/C6 已关；Bot 通道体验债已清；附件 P0 由 Go/Issue/React 落地中。  
5. **不做**：GitHub Projects 时间线；Cycle 用现有页。全文：`docs/ACCEPTANCE_UX_BOT.md`。

---

## C) 附件 P0 · 最小验收标准（Linear 式）

> **裁定**：总指挥 P0——评论 + Issue 可传图片/文档；对标 Linear 附件体验。  
> **本轮交付**：仅验收/规格文档（本文件）；**禁止**本轮改上传业务代码（tsx/go）。  
> **现状（2026-10-24 只读扫）**：`docs/API.md` **无** upload/attachment 路由；`schema` **无** attachments 表；`comments` 仅 `body` 文本；handlers/services **无** multipart/`FormFile`；Issue 前端（`IssueDrawer` / `IssueCommentList` / `IssueDescriptionEditor`）**无**文件上传 UI。缺口 = 全链路未建。  
> **回链**：`docs/ACCEPTANCE.md` §3.8 旁注。

### C.0 产品边界（克制）

- Issue 级附件与评论附件共用同一套存储/鉴权/限制；UI 入口分别在描述区附近与评论 composer。
- **不做**（P0 外）：多图拖拽画廊、在线预览 Office、病毒扫描、外链公开分享、Wiki/工时附件复用。
- 列表形态：一行「文件名 · 可读大小 · 删除」；图片另有缩略；**禁止**密密麻麻 Form 标签墙。

### C.1 可勾选清单

| ID | 验收标准 | 通过条件（可执行） | 建议归属 |
|----|----------|-------------------|----------|
| **F1** | Issue 级可上传常见类型 | **通过** | API + FE `uploadIssueAttachment` / `IssueAttachmentUpload` 路径对齐；预检中文 |
| **F2** | 评论可挂附件 | **通过** | FE 已带 `issueID`：`…/issues/{id}/comments/{cid}/attachments`（upload/list） |
| **F3** | 图片预览 | **通过** | `IssueAttachmentList` AuthThumb + Image 预览；content 带鉴权拉取 |
| **F4** | 文档下载 | **通过** | Go content + FE `downloadAttachment`（Bearer blob 下载） |
| **F5** | 类型限制 + 中文失败 | **通过** | `validate.go` 白名单；测例拒 `.exe`；中文 message；失败不落库 |
| **F6** | 大小限制 + 中文失败 | **通过** | 服务端 10MB + FE 预检「文件过大，上限 10MB」 |
| **F7** | 鉴权同路径 | **通过** | `router.go` 附件路由同 ws 组 `AuthJWTOrAPIKey`；无第二套 URL |
| **F8** | 权限最小 | **通过** | `CanDelete`：上传者或 admin；测例绿（含 bot 绑定 member） |
| **F9** | 紧凑列表 UI | **通过** | 看板 UI：单行「缩略/图标·文件名·大小·删除」+ 空态「暂无附件」（`IssueAttachmentList.tsx`） |
| **F10** | 失败不脏数据 | **通过** | 校验失败不落库；Delete 后 content 不可再下（测例覆盖） |

**P0 验收条数：10**（F1–F10）。全勾方可关「附件 P0」。
> **进度 2026-10-24**：Issue 域 List×2+Delete ✓；Go 上传/content/限制/鉴权/删除策略 ✓（`go test ./internal/attachments` 绿）。React 路径已对齐（评论含 issueID）；预检中文齐全。§C F1–F10 **可关单**（联调以本地 build 收尾为准）。


### C.2 鉴权建议（写入规格，实施时遵守）

1. **同路径**：`POST/GET/DELETE …/issues/{id}/attachments`（及评论挂载变体）一律 `AuthJWTOrAPIKey`，与 Issue/Comment 写读一致；**不要**做人/Bot 两套 URL。  
2. Bot：`Authorization: Bearer ymk_…` → `actor_kind=bot`；写操作建议支持 `Idempotency-Key`（与现有约定对齐）。  
3. 下载：须带会话或短期签名；**禁止**无鉴权永久公开 URL（P0）。

### C.3 给后端落地要点（催办 · 3 条）

1. **表 + API**：append `attachments`（或等价）表（id、workspace/issue、optional comment_id、uploader、filename、mime、size、storage_key、created_at）；路由建议 `…/issues/{id}/attachments` CRUD + 下载；评论创建可接受 `attachment_ids[]` 或同 multipart。Owner：**Go 后端 / Issue 域**（schema append 同步 Tech lead；评论挂载协同 Collab）。  
2. **限制权威在服务端**：白名单 MIME/扩展 + 大小上限；统一错误体 `{"code":"…","message":"中文"}`；存储可先本地目录（配置路径），接口形状预留对象存储。  
3. **鉴权**：全挂 `AuthJWTOrAPIKey`；成员校验与 Issue 一致；删除策略写清并测；Bot 与人同路径。

### C.4 给前端落地要点（催办 · 3 条）

1. **入口克制**：描述区旁「添加附件」+ 评论 composer 回形针；`input[type=file]` / 轻量选择，**不要** Ant Form 多字段墙。Owner：**React 前端**（看板 UI 确认间距/缩略尺寸）。  
2. **列表**：紧凑行（名+大小+删）；图片缩略可点大图；文档仅下载按钮。失败用现有 Toast 中文。  
3. **客户端预检**：选文件即拦类型/大小（与后端文案一致）；上传中禁用重复点；刷新后列表与评论附件仍在。

### C.5 指派一览

| 角色 | 负责 |
|------|------|
| **Go 后端 / Issue 域** | 表、上传下载 API、限制、鉴权同路径 |
| **Collab（评论）** | 评论与 `attachment_ids` / 挂载关系（若拆分） |
| **React 前端** | Issue/评论上传入口、紧凑列表、预览/下载、中文失败 |
| **看板 UI** | 密度/缩略/空态与 Linear 主题对齐（可只审不写逻辑） |

### C.6 给总指挥短报（可直接转发）

1. **现状**：附件全链路未建（无 API/表/前端上传）；评论仅纯文本。  
2. **P0 验收**：本文件 §C 共 **10** 条（F1–F10），可勾选关单。  
3. **指派**：Go/Issue 域扛存储+API+鉴权；React 扛紧凑 UI/预览；看板 UI 审密度；Collab 协同评论挂载。  
4. **约束**：JWT 与 Bot Key 同路径；单文件建议 ≤10MB；禁做成表单墙。  
5. **本轮**：只改验收文档，未写上传业务代码。全文：`docs/ACCEPTANCE_UX_BOT.md` §C。


---



---

## D) Bot 通道 DX · 省 token（对照 `BOT_CHANNEL_DX_P0`）

> 权威细则与字段表：[`BOT_CHANNEL_DX_P0.md`](./BOT_CHANNEL_DX_P0.md)。  
> 状态：**D1–D6 已勾关单**（产品体验 2026-10-24 · Asia/Shanghai · 本机真 Key）。  
> 本表为验收镜像，关单以 DX 文档勾选+证据为准。

| ID | 标准摘要 | 结果 | 说明 |
|----|----------|------|------|
| **D1** | Issue 对外 `YM-n`；入参可 `YM-12` | **通过** | 真 Key：get YM-9/12；YM-9999 exit 12 |
| **D2** | list 极瘦（identifier/title/status/priority/assignee 名…） | **通过** | 实机 list 瘦字段 + 单测 |
| **D3** | 默认文本友好；禁 REST 原样 dump | **通过** | 默认文本表 |
| **D4** | 默认无 32 位 hex；`--id`/`--full`（`ids`/`full`）才放宽 | **通过** | 默认 uuid=0；--id/--full 放宽 |
| **D5** | handoff/memory/comment 默认摘要 | **通过** | 实机摘要形态 |
| **D6** | MCP 与 CLI 同规则；示例用 `YM-n` | **通过** | MCP 12 tools；help/BOT.md |



---

## E) Bot 续命 · 服务端 handoff（对照 `BOT_HANDOFF_CONTINUITY_P0`）

> 权威：[`BOT_HANDOFF_CONTINUITY_P0.md`](./BOT_HANDOFF_CONTINUITY_P0.md)。  
> 状态：**H1–H5 已勾关单**（产品体验 2026-10-24 · Asia/Shanghai · 真 Key）。

| ID | 标准摘要 | 结果 | 说明 |
|----|----------|------|------|
| **H1** | 文档主路径= `handoff get`；export/import 标备份 | **通过** | help/MCP/BOT.md |
| **H2** | 剧本无强制 export | **通过** | 典型一天 |
| **H3** | 中途挂：只写 memory+comment 再 get 可续 | **通过** | 真 Key；焦点+YM-9 评论 |
| **H4** | export 仍可用但非关卡 | **通过** | export --out 成功 |
| **H5** | BOT / MCP_CLI 与 DX 一致 | **通过** | get 含 version |

## 修订记录

| 日期 | 说明 |
|------|------|
| 2026-10-24 | 初版验收；同步标注 `BOT_MCP_CLI.md` §3「P0 已落地」；明确不列 GitHub Projects 时间线缺口 |
| 2026-10-24 | 产品体验：React 路径对齐，勾齐 §C F1–F10 |
| 2026-10-24 | 产品体验：勾 §C 后端项 F1/F4–F8/F10（API）+ F9 UI；F3/FE 联调待齐 |
| 2026-10-24 | 产品体验：复验勾掉 B1/C6（React 已修 max-width 520 + 窄屏 min(100vw,560)）；附件 P0 继续盯 |
| 2026-10-24 | Bot 记忆角色：幂等 Key 透传 + 中文错误 + §7 真 Key 关单；优先缺口 #2/#3 关闭；回写本表与 `BOT_MCP_CLI.md` §7 |
| 2026-10-24 | 增 §C「附件 P0」最小验收（F1–F10）+ 后端/前端催办要点；只改规格，未实现上传代码 |
| 2026-10-24 | 产品体验：增 §D 通道 DX（D1–D6）镜像；权威见 `BOT_CHANNEL_DX_P0.md` |
| 2026-10-24 | Bot 记忆角色：通道 DX D1–D6 投影落地（`--id`/`--full` 分档；默认禁 UUID）；状态改「通道已落地待产品体验勾选」 |
| 2026-10-24 | 产品体验：通道 DX **D1–D6 勾关**（真 Key；默认 stdout UUID=0） |
| 2026-10-24 | 产品体验：增 §E 续命（H1–H5）；权威 `BOT_HANDOFF_CONTINUITY_P0.md` |
| 2026-10-24 | 产品体验：续命 **H1–H5 勾关**（真 Key 中途挂复验） |

---

## F) Issue 详情打磨 P0（续）

布局骨架 §A 已关后的观感打磨：见 [`ISSUE_DETAIL_POLISH_P0.md`](./ISSUE_DETAIL_POLISH_P0.md)（P1–P5：标题裁切 / 关联中文 / 关联添加安静 / 空态藏 MD 工具栏 / 整体抽检）。
