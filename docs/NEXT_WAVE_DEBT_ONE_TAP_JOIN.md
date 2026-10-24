# 一键入场（ONE_TAP_JOIN）

> 2026-10-24 22:05 Asia/Shanghai · 产品体验  
> **用户定调（总指挥转·怒）：** 停「教人配一堆 env / MCP」路线。只要看板后台 **一条 URL 或一段可粘贴提示词**，新 Bot 打开/粘贴后该有的都带好——**零心智负担**。  
> **生产 API 根（默认写入链接/提示词）：** `https://bot.api.sn.cn`  
> **承接：** 队席入场码 `ymt_` + `team join` / redeem 已有（[`BOT_TEAM_JOIN_KEY_P0.md`](./BOT_TEAM_JOIN_KEY_P0.md)）；本波只做「一键带走」，不重做 claim。

## 一句话

Admin 在 Bot 席页点一下 → 剪贴板里是 **入场链接** 或 **给 Cursor/Agent 的系统提示词**；对方打开/粘贴即知 API 根、入场码、席位角色，能自动 join 更好——**禁止再甩 YMBOT_* / 手改 mcp.json / 长安装文档**。

## 现网可扩（勿重做）

| 已有 | 本波 |
| --- | --- |
| `POST …/teams/{id}/invites` 签发 `ymt_`；明文仅一次 | **签发后直接生成 URL + 提示词**，不只复制裸码 |
| `POST /team-invites/redeem`；`ymbot team join --code` | **深链/提示词内嵌同一 code**；落地仍走既有 join |
| FE「复制入场码」 | **升为「复制入场链接」+「复制 Bot 提示词」**（裸码可留次要） |
| FE JoinTeamWizard（人类贴码） | **可选**：打开 `/join?…` 时预填向导；Bot 侧以提示词/CLI 为主 |

## 含 / 不含

| 做 | 不做 |
| --- | --- |
| 后台一键「复制入场链接」「复制 Bot 提示词」 | 再教人配 `YMBOT_API_URL` / `YMBOT_API_KEY` 清单 |
| URL：`https://bot.api.sn.cn/join?code=ymt_…&role=frontend`（可加 `seat=`；API 根可配置） | 手改 `mcp.json` 教程、长安装 README 当主路径 |
| 提示词内嵌：API 根 + code + role + 主航线（whoami→join→handoff） | 把 `ymk_` 塞进可转发 URL（防泄密；Key 仍本机/Vault） |
| 打开链接：浏览器引导页或前端同源深链，能自动 redeem/join 更好 | 未分队大入口、T7 升版、跨队搬家 |
| 文档改口：主路径 = 一键链接/提示词 | 以 env 教学为主的入场文案（旧文降级附录） |

## 契约（产品冻结）

### 入场链接

```
https://bot.api.sn.cn/join?code={ymt_…}&role={seat_slug}
```

| 参数 | 必填 | 说明 |
| --- | --- | --- |
| `code` | 是 | 明文 `ymt_…`（与签发同一次；链接等同持有码，须短时效/可吊销） |
| `role` | 建议 | 席位 slug（如 `frontend`）；缺省则 join 走新人自动选席 |
| `seat` | 可选 | 指定 `seat_id`（高级）；与 role 二选一优先 seat |
| `api` | 可选 | 覆盖 API 根；默认省略=生产 `https://bot.api.sn.cn` |

**落地页（二选一，本包须定一种主路径）：**

1. **API 宿主静态/轻页** `GET /join`：展示「复制提示词 / 打开看板深链 / CLI 一行」；能调 redeem 则显示班底摘要。  
2. **前端同源** `/join?code=&role=`：进 JoinTeamWizard 预填（人类）；Bot 仍以提示词为主。

产品默认：**链接宿主 = API 根 `/join`**；看板 FE 可另提供「在浏览器打开向导」同源链（同参）。

### Bot 提示词（粘贴即用）

一段中文系统提示（可 Markdown），至少含：

1. API 根：`https://bot.api.sn.cn`（或配置值）  
2. 入场码：`ymt_…`  
3. 目标角色：`role` / 显示名  
4. 主航线硬指令：`board_whoami` / `ymbot whoami` → `team join --code … [--role …]` → `handoff get`；真相源在服务端  
5. **禁止**让用户去改 mcp.json / 导出一长串 env；若通道需 Key：写「用本机已有 Bot Key / Vault，勿把 ymk_ 贴进聊天」

### 安全

- URL/提示词含 **code 明文** → 等同分享入场码：签发 UI 必须提示「勿发公开群」；沿用吊销/过期/次数。  
- **禁止**把 `ymk_` 写入可复制链接或默认可转发提示词模板。  
- 生产 API 根写死默认；自托管用环境变量 `PUBLIC_API_BASE`（名以 Go 落地为准）。

## 验收小包

### A · 链接契约（Go 主）

| ID | 标准 | 谁 |
| --- | --- | --- |
| **A1** | 文档+可选 helper：由 `code/role/api` 拼规范 join URL；单测或 golden | Go |
| **A2** | `GET /join`（或约定宿主）能解析 query；非法 code 人话错误；不做第二套 claim | Go |
| **A3** | 配置项：默认 API 根 `https://bot.api.sn.cn`；可覆盖 | Go + 种子文档 |
| **A4** | curl/箱证：签发 → 拼 URL → redeem/join 仍绿 | Go → QA |

### B · 后台一键（React 主）

| ID | 标准 | 谁 |
| --- | --- | --- |
| **B1** | Bot 席页：「复制入场链接」= 签发（或选用未吊销码）+ 写入剪贴板完整 URL；Toast 含末四 | React |
| **B2** | 「复制 Bot 提示词」= 同码同 role 的完整提示词；一键剪贴板 | React |
| **B3** | 席位卡/空席：可带当前 `role` slug 进链接；无角色时提示先选席或发通用码 | React |
| **B4** | 弱化「只复制裸 ymt_」为次要；主 CTA 改为链接/提示词 | React |
| **B5** | 箱截 `one-tap-copy-link.png` + `one-tap-copy-prompt.png` | React → 产品 |

### C · 打开即用（Go/React）

| ID | 标准 | 谁 |
| --- | --- | --- |
| **C1** | 浏览器打开 join URL：见短引导（复制提示词 / 若已登录 FE 则预填向导） | Go 页或 React `/join` |
| **C2** | 通道：提示词内一行命令可 `team join` 成功（沿用 J2）；不新增必填 env 教学 | 记忆/通道抽 + QA |
| **C3** | 箱证 txt 或截：链接 → join PASS | QA |

### D · 文案收口

| ID | 标准 | 谁 |
| --- | --- | --- |
| **D1** | BOT.md / 席位页：主路径改「复制入场链接或提示词」；env/mcp 降附录并标明非推荐 | 种子 + 产品 |
| **D2** | 产品勾选关单 | 产品体验 |

## 派工

| 角色 | 职责 |
| --- | --- |
| **Go 后端** | **与 React 并行首刀**：A1–A3 链接契约 + `/join` 宿主（或明确委托 FE） |
| **React 前端** | **并行首刀**：B1–B5 两按钮 + 箱截；可选同源 `/join` |
| **Bot 记忆 / 通道** | C2 提示词航线与 `team join` 对齐；勿写回 env 教程 |
| **种子与文档** | A3/D1 |
| **QA** | A4/C3 |
| **产品体验** | 本裁定；D2 |

## 产品勾选

- [ ] **A1** join URL 契约
- [ ] **A2** `/join` 解析页
- [ ] **A3** 默认 API 根可配
- [ ] **A4** 签发→URL→join 箱证
- [ ] **B1** 复制入场链接
- [ ] **B2** 复制 Bot 提示词
- [ ] **B3** role 带入
- [ ] **B4** 主 CTA 切换
- [ ] **B5** 箱截
- [ ] **C1** 打开引导
- [ ] **C2** 提示词航线 join
- [ ] **C3** 打开即用箱证
- [ ] **D1** 文档主路径改口
- [ ] **D2** 产品关单

## 进度注记

- 2026-10-24 22:05 · 债文定稿。首刀：**Go A 契约 + React B 按钮并行**。
- 2026-10-24 22:10+ · **Go A1–A3 已落地**：`internal/joinlink` + `GET /join`（API 根）+ `PUBLIC_API_BASE`（默认 `https://bot.api.sn.cn`）。A4 箱证留给 QA。

## 关单进度

**进行中** · Go A1–A3 已推；待 React B + QA A4/C3。
