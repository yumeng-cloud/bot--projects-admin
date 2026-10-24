# 全页截图验收清单 · box 交付前硬规矩

> QA · 2026-10-24  
> **硬规矩**：交付前必须在云主机（box）自行把每个主页面截一遍看过；不过关不喊用户验收。  
> 入口（box）：`http://127.0.0.1:5173`（Vite 默认；API 经 proxy → 8080）  
> 产出：`/workspace/qa-shots/`（PNG，文件名 = ID）  
> 触发：前端说可测，或总指挥喊开跑。Wiki 正做 WIKI_UX_P1，Wiki 项等 P1 可测后再截。

## 前置

- [ ] box 后端 `:8080` 健康（`/healthz`）
- [ ] box 前端 Vite 可开（`5173`）
- [ ] 有可用 seed 账号 + 至少一个 workspace / project
- [ ] 登出态可回到登录页（清 token 或无痕）

## 主页面（必截 · 每页至少 1 张全视口）

| ID | 页面 | 路径（相对 `/w/:wsId`） | 看什么（不过关即 Fail） | 截图 | 结果 |
|----|------|-------------------------|-------------------------|------|------|
| **S1** | 登录 | `/`（未登录） | 表单完整、无白屏/错栈、可提交 | | |
| **S2** | 看板 | `…/projects/:pid/board` 或 `/board` | 列+卡片可见，非空壳崩溃 | | |
| **S3** | 列表 | `…/projects/:pid/issues` 或 `/list` | 表格/列表有数据行或空态像样 | | |
| **S4** | Cycles | `…/projects/:pid/cycles` 或 `/cycles` | Cycle 列表/详情区正常 | | |
| **S5** | 排期 | `…/projects/:pid/timeline` 或 `/timeline` | 时间线/排期区渲染，非空白崩 | | |
| **S6** | Wiki | `/wiki`（+ 打开一篇若有） | 目录+正文；**WIKI_UX_P1 后**无漂浮空白/完成紫胶囊/裸 Select 行 | | |
| **S7** | 工时 | `/time` | 工时表/汇总区正常 | | |
| **S8** | 设置 | `/settings` | 设置页可进、无崩 | | |
| **S9** | Bot | `/settings/bots`（有 Bot 则再开详情） | 名册列表；详情记忆/Todo 区不炸 | | |

## 可选加深（有时间再截）

| ID | 场景 | 说明 |
|----|------|------|
| S2′ | 看板开 Issue 抽屉 | 详情主内容+右侧属性轨 |
| S6′ | Wiki 编辑态 | 对标 WIKI_UX_P1 P1-1…P1-6 |
| S9′ | Bot 详情 | `/settings/bots/:memberId` |

## 判定

- **Pass**：9 张主页均已目视，无白屏/未捕获错/明显错布局；Wiki 另满足当时生效的 WIKI_UX 勾选。  
- **Fail**：任一主页崩、空白无空态、或与当前 P0/P1 文档明显冲突 → 缺陷指派对应同事，**不喊用户验收**。  
- 短报总指挥：`S1–S9 PASS/FAIL` + 失败 ID + 截图路径。

## 状态

- 清单已备 · **待开跑**（等 FE 可测 / 总指挥指令）
- Wiki：等 WIKI_UX_P1 可测后再截 S6

## 最近一次跑（box · 2026-10-24 13:01–13:03 Asia/Shanghai）

入口：`http://127.0.0.1:5174` · 账号 seed `admin@yumeng.tech`

| ID | 结果 | 文件 |
|----|------|------|
| S1 | **PASS** | `/workspace/qa-shots/S1-login.png` |
| S2 | **PASS** | `/workspace/qa-shots/S2-board.png` |
| S3 | **PASS** | `/workspace/qa-shots/S3-list.png` |
| S4 | **PASS** | `/workspace/qa-shots/S4-cycles.png` |
| S5 | **PASS** | `/workspace/qa-shots/S5-timeline.png` |
| S6 | **PASS**（苛刻 · WIKI_UX_P1） | `/workspace/qa-shots/S6-wiki.png` + `S6-wiki-edit.png` |
| S7 | **PASS**（`7fbd3d3` · YM-n+人名） | `docs/qa-shots/s7-time-7fbd3d3/01-time-full.png` |
| S8 | **PASS**（a84e0ac · 角色中文） | `docs/qa-shots/fail-fix-20261024/05-settings.png` |
| S9 | **PASS**（a84e0ac · 认领「名」） | `docs/qa-shots/fail-fix-20261024/06-bots.png` |

**汇总（已作废宽松判）**：S1–S5 曾 PASS；**S7–S9 按苛刻标准改记 FAIL 债**；S6 Wiki pending。Cycles/工作区同批苛刻 FAIL，复测前不喊用户。

## 已知缺口（不挡本轮）

- **S7 工时 Issue 列 raw id**：已关 · QA 苛刻复测 PASS（2026-10-24 17:52 Asia/Shanghai · `7fbd3d3`）· Issue=`YM-9 …` 成员=`管理员` · 证据 `docs/qa-shots/s7-time-7fbd3d3/`。

## 苛刻验收标准（总指挥 · 2026-10-24 对齐）

**禁止**以「能打开 / 不白屏」当 PASS。以下任一即 **FAIL**：
- 裸 hash / raw id / 技术字段当主文案（工时 Issue 列、成员 id 等）
- 漂浮空底、大片无信息空白像未做完 demo
- 像后台 CRUD：表单感、技术字段堆砌、无产品态空态

### 记债（同批 computerUse 苛刻复判）

| ID | 原判 | 苛刻复判 | 债 |
|----|------|----------|-----|
| S4 Cycles | PASS | 待 FE 修后按苛刻条复测（总指挥同批点名 Cycles） | 记债 |
| S7 工时 | **PASS**（`7fbd3d3`） | YM-9+标题 / 管理员 | `docs/qa-shots/s7-time-7fbd3d3/` |
| S8 设置 | PASS → **FAIL** | 苛刻：漂浮空底/技术字段（同批 FAIL） | **FAIL 债** |
| S9 Bot | PASS → **FAIL** | 苛刻：同批 FAIL | **FAIL 债** |
| 工作区 | — | 同批 FAIL（裸 hash/漂浮空底） | 记债 |

FE 修完后：按本苛刻条复测 S7–S9（及工作区/Cycles），截图进 `/workspace/qa-shots/`，不过关不喊用户。

## Wiki S6 苛刻跑（2026-10-24 13:09 Asia/Shanghai）

入口：`5174` 硬刷 · `dea0f77` WIKI_UX_P1  
**S6 Wiki：PASS**（P1-1…P1-7 全过；无达 FAIL 条缺陷）  
截图：`/workspace/qa-shots/S6-wiki.png`、`S6-wiki-edit.png`  
Nits（不挡）：树搜索蓝方钮；树宽贴 240 下限；编辑插图小图标簇。

## 苛刻 FAIL 五项关单（2026-10-24 13:10 Asia/Shanghai · `a84e0ac`）

箱 `5174` 自截：`docs/qa-shots/fail-fix-20261024/`

| # | 项 | 判定 |
|---|----|------|
| 1 | Wiki 默认开首篇 | PASS — 进 Wiki 即开「QA page」，无「还没有打开页面」 |
| 2 | 工时 YM-n + 人名 | PASS — Issue=`YM-12 …`，成员=`管理员` |
| 3 | 工作区真导航 | PASS — 表行 看板/列表/Cycles/Wiki/设置，无漂浮小卡 |
| 4 | Cycles 密度+⋯ | PASS — 双栏 panel；Issue 操作仅 ⋯ |
| 5 | 设置/Bot 中文 | PASS — 管理员/机器人；认领「后端」等；无 slug |

看板/列表/排期/抽屉未动。

## S7 苛刻复测 · 2026-10-24 17:52 Asia/Shanghai

| 项 | 结果 |
| --- | --- |
| Commit | `7fbd3d3` |
| API List/Get | **PASS** — `issue_identifier=YM-9` / `issue_title=QA webhook match BOT_A updated` / `member_name=管理员` |
| FE `/w/…/time` | **PASS** — 列见「YM-9 QA webhook match BOT_A updated」+「管理员」，无 raw id |
| 证据 | `docs/qa-shots/s7-time-7fbd3d3/01-time-full.png` · `02-time-issue-member-closeup.png` · `api-list.json` · `api-get.json` |
