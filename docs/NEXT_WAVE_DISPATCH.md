# 下一波立刻可干 · 派工单

> 2026-10-24 17:22 Asia/Shanghai · HEAD `5b15581` · 箱 5174+8080(PG) 健康  
> 依据：`git log -5`、docs 未勾/P0/REGRESSION/BOT_*_DECISION、QA nits、FE/BE 扫码  
> **禁止**刷进度空活；**优先**用户体感 + 进场主航线收口。本文件只派工，不 push、不叫醒。

## 现状快照

| 已关（勿重复开） | 仍敞口 |
| --- | --- |
| FE F1/T3/S/P `39c4155` 七张 PASS；`allow_expand_seat` 复制码 `5b15581` PASS | QA nits：占用席 `aria-disabled`、title 仍 `CLI·slug`、列模板弹层需滚 |
| Go G2/G3/G6a `0685720` CreateSeat+ymt_+webhook 过滤 PASS | handoff 摘要 `in_review` 未进 `STATUS_CN` |
| 附件 §C F1–F10 代码路径已齐（可关单待产品勾） | Settings 席位「空闲」vs 向导「占用中」数据源不同步 |
| force 活跃→409；真码兑换箱验 PASS | FE `listTeamInvites` 有、BE **无 GET** → 404 |
| | Issue 详情打磨 P0/P0′ 多条仍 `[ ]`；`BOT_FE`「真入场码闭环」勾选滞后；J1–J7 文档未勾 |

`git log -5`：`5b15581` FE allow_expand · `6385eec` J2 箱证 · `0685720` G2/G3/G6a · `8fa4f68` 勾选 FE · `39c4155` F1/T3/P。

---

## 立刻可干清单（按体感优先）

### D1 · handoff `in_review` 中文
- **谁**：React 前端
- **一句话**：`seatHandoffText.ts` 的 `STATUS_CN` 补 `in_review: '评审中'`（与 `types/issue.ts` 对齐），交接卡/抽屉不再漏英文。
- **验收**：席位交接摘要里 Issue 状态为「评审中」；箱 5174 打开已绑席 handoff，无裸 `in_review`。

### D2 · 席位 title 去 `CLI·slug`
- **谁**：React 前端
- **一句话**：`BotsRosterPage` / `JoinTeamWizard` 去掉 `title={\`CLI · ${slug}\`}`；悬停用人话席位名（`roleDisplay` / `seatDisplayName`），slug 仅留高级 Collapse。
- **验收**：悬停无 `CLI ·`；设置班底与向导班底 title 均为中文名；QA 再拍 S-bots / F1-roster。

### D3 · 占用席可访问态
- **谁**：React 前端
- **一句话**：向导占用席：`aria-disabled` 改真正不可聚焦（`disabled` 或 `tabIndex=-1` + 点击/键盘均拦截），灰显「占用中」保留，无 force。
- **验收**：键盘 Tab 跳过占用席；点/Enter 不触发进席；T3 仍仅活跃冲突进专面。

### D4 · 列模板弹层可滚且一屏可用
- **谁**：React 前端（看板 UI 可审密度）
- **一句话**：`ColumnTemplatePicker` Popover/内容区定 `max-height` + `overflow-y:auto`（现 board.css 有 320 线索，确认挂在弹层内容上），小屏不必裁死底部按钮。
- **验收**：箱 5174 看板 ⋯→列模板，包多时内容可滚、套用/自定义入口可见；再截 `P-col-template`。

### D5 · Settings「空闲」与向导「占用」同源
- **谁**：React 前端（必要时 Go 补 roster 字段）
- **一句话**：设置 Bot 班底不要只看 `role.claim.member_id` 标「空闲」；与向导一致读 `team_seats.status`（active→占用/已绑，empty→空闲/空席，left→可补位）。两页同一席不得一空一占。
- **验收**：同一 team 下活跃席：设置≠「空闲」、向导=「占用中」；停工席两边都可补位语义；箱截对比两页。

### D6 · GET 入场码列表（消 404）
- **谁**：Go 后端
- **一句话**：补 `GET /api/v1/workspaces/{ws}/teams/{team}/invites`（末四/状态/过期，**永不回明文**），对齐 FE `listTeamInvites`；Admin only。
- **验收**：复制码后列表 200；无码 200 `[]`；Bot Key 403；旧「invite 404 人话」路径可退役。单测 + 箱 curl。

### D7 · Issue 抽屉打磨收口（用户天天见）
- **谁**：React 前端 + UI 视觉抽检
- **一句话**：按 `ISSUE_DETAIL_POLISH_P0` 未勾项一次收：标题首字不裁（P1）、关联中文+安静添加（P2/P3/P3′）、空描述藏工具栏（P4/P4′）、评论删 hover 显（P6）、活动默认折（P7）。
- **验收**：箱 5174 开任一 Issue 抽屉对照表逐条；产品勾 P1–P7；**不做** P8 用户本机强刷（非箱责）。

### D8 · 进场文档勾选与 J 箱证对齐（非空活）
- **谁**：产品体验 + QA（Go/通道只补缺证据）
- **一句话**：`BOT_FE_ONBOARD`「真入场码闭环」在 `5b15581` 已 PASS→勾上；`BOT_TEAM_JOIN_KEY` J1/J3–J5 用现成 `/workspace/qa-shots/full-0685720` + J2 证据勾或补一条幂等；J7 webhook 过滤已有 B4 则勾，缺则只补指派投递一条。
- **验收**：相关 `[ ]`→`[x]` 带证据路径；无证据不勾；不新开功能分支。

---

## 暂缓（有价值但不「立刻」、防空转）

| 项 | 原因 |
| --- | --- |
| 附件 I3 / §C 再开发 | F1–F10 已标通过；只需产品关单，勿重写上传 |
| Wiki W5 截图勾 | 等空档；不挡进场 |
| T6/T7 团队模板 / 共享记忆墙 | P0.5/P1，决策未催；**最小说明**已落 [`BOT_TEAM_TEMPLATE_T6_T7.md`](./BOT_TEAM_TEMPLATE_T6_T7.md) |
| mirror / 心跳判死 / WS | 决策明确 P0 不做 |
| BOT_MCP_CLI 尾部旧勾选 | 通道 P0 已关；勿当新活 |

---

## 建议席位派发（TOP）

| # | 项 | 席位 |
| --- | --- | --- |
| 1 | D1 `in_review` 译 | **React**（可 5 分钟合进 D2/D3 同 PR） |
| 2 | D2 title 去 CLI·slug | **React** |
| 3 | D3 占用席 a11y | **React** |
| 4 | D4 列模板弹层滚 | **React**（**看板 UI** 审） |
| 5 | D5 空闲/占用同源 | **React**（字段不够再喊 **Go**） |
| 6 | D6 GET invites | **Go** |
| 7 | D7 Issue 抽屉打磨 | **React** + **UI 视觉** |
| 8 | D8 勾选/箱证对齐 | **产品体验** + **QA** |

**并行建议**：React 吃 D1–D5（可同一波 FE）；Go 并行 D6；QA/产品 D8 不挡码；D7 第二波但同日可开。

**成功标准（本文件）**：有可执行派工单 → 本文件已落盘 `docs/NEXT_WAVE_DISPATCH.md`。


---

## 产品体验回执（2026-10-24 17:25 Asia/Shanghai）

| 项 | 结论 |
| --- | --- |
| **D8** | **已勾**。`BOT_FE` 真入场码闭环 → `[x]`（G3 `0685720` + FE `5b15581` · `full-0685720`）；`BOT_TEAM_JOIN_KEY` **J1–J7** 均 `[x]`（J2 原有；J1/J3–J5/J7 对齐箱证；J1 吊销 HTTP/FE、J4 Issue 正文 comment、J7 巡检/下发 FE 仍记债不挡） |
| **D7** | **不重开**。§1/§6 滞后 `[ ]` 已对齐 §7 关单。React 仅顺手 P1 去负字距；主活仍 D1–D5 |
