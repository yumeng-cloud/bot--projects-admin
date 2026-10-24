# 下一小包 · T7-3 · 执行源=playbook_ref（最小）

> 2026-10-24 19:45 Asia/Shanghai · 产品体验 · T7-R 创建整单关后下一刀  
> **本包：席位/通道只认 `playbook_ref` 手册，不另起第二套流程文。升版 T7-4/5 仍后置。**  
> **整单 PASS** · 箱证 [`T73-playbook-source-20261024.txt`](./qa-shots/T73-playbook-source-20261024.txt) · API `d9e5a3c` · 证据 push `d5436e8`。升版 T7-4/5 仍后置。

## 裁定

| 候选 | 结论 |
| --- | --- |
| **T7-3 执行源** | **开本刀** |
| T7-4/5 升版 dry-run/apply | **仍后置** |
| FE 顶栏统一切队 | 后置（list/board 已有 `TeamPoolSelect`；旁路页只读 `useCurrentTeam` 可活） |
| 未分队大入口 | 后置（T6-I 已裁定本包不做） |

对齐 [`BOT_TEAM_TEMPLATE_T6_T7.md`](./BOT_TEAM_TEMPLATE_T6_T7.md) **T7-3**；创建面已绿见 [`NEXT_WAVE_DEBT_T7.md`](./NEXT_WAVE_DEBT_T7.md)。

## 一句话

Claim/enter/handoff/CLI 展示与指引：**以席位 `playbook_ref` 为唯一作业手册入口**；禁止再塞一份并行「流程.md / 口令剧本」当权威。

## 含 / 不含

| 做 | 不做 |
| --- | --- |
| roster/enter/handoff 输出含可读 `playbook_ref`（或解析后的 repo@ref:path） | 升版 diff/apply（T7-4/5） |
| ymbot/MCP：跟手册前解析 ref；非法/空 → 一行中文，不默默换别的文 | 自动 clone 模板仓；Agent 自动跑完手册 |
| 箱证：同席两次 enter 指同一 ref；改 DB ref 后出口跟着变 | FE 多步向导大改；全局顶栏切队 |

## 验收（E1–E4）

| ID | 标准 | 谁 |
| --- | --- | --- |
| **E1** | `roster` / `seat enter`（或等价）回包可见该席 `playbook_ref` | **Bot 记忆** + Go |
| **E2** | CLI/MCP 默认人话含手册路径（截断可）；无第二套「流程说明」字段抢权威 | 记忆 |
| **E3** | 非法/空 ref → 4xx 或一行中文警示，不回落硬编码别的 playbook | Go/记忆 |
| **E4** | 证据 `docs/qa-shots/T73-playbook-source-YYYYMMDD.txt` → **产品**勾；勾后回写 T6_T7 **T7-3 [x]** | → 产品 |

## 派工

**Bot 记忆** 主 · **Go** 契约缺口 · **种子** 一行 API/BOT.md · **QA** 抽苛刻 · **产品**勾  
**React** 本包不挡（席位卡已有 ref 展示则顺手对齐，无则后置）

## 产品勾选

- [x] **E1–E4** PASS · 2026-10-24 19:52 Asia/Shanghai
  - API `d9e5a3c` · 箱证 [`qa-shots/T73-playbook-source-20261024.txt`](./qa-shots/T73-playbook-source-20261024.txt) · `d5436e8`
  - E1：roster/enter/handoff 可见席位 `playbook_ref`；同席两次同 ref；ensure 改 ref 后出口跟着变
  - E2：CLI 人话「手册 playbook_ref=…」；无「流程说明」第二套字段
  - E3：非法/空 ensure → 400 中文；DB 空 → roster 警示 + enter 400；不回落硬编码
  - 产品抽跑 `go test … botmem/botclient -run Playbook|T73|ParsePlaybook` ok
  - 已回写 [`BOT_TEAM_TEMPLATE_T6_T7.md`](./BOT_TEAM_TEMPLATE_T6_T7.md) **T7-3 [x]**
  - 升版 T7-4/5 仍后置

## 关单进度

**整单关**：E1–E4 全勾 · 2026-10-24 19:52 Asia/Shanghai。升版 T7-4/5 仍后置。
