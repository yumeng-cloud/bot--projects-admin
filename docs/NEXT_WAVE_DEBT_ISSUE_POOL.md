# NEXT_WAVE · Issue 池按队（T6 余债）

> **权威债文**：[`NEXT_WAVE_DEBT_T6I.md`](./NEXT_WAVE_DEBT_T6I.md)（T6-I）。本文仅历史草稿，契约以 T6I / `API.md`「Issue · team_id」为准。


> 2026-10-24 · 总指挥自定开工（用户禁闲等；pins 整单收口后立刻开）  
> **本包：Issue 按 `team_id` 隔离可见。不含 T7、不再动 pins。**

## 目标

同工作区两队：列表/看板/查询默认只见**当前 team** 的 Issue；队 A 的单在队 B 池不可见（或显式「全工作区」才见，本包默认关）。

## 做 / 不做

| 做 | 不做 |
|----|------|
| Issue 表加可空 `team_id`（FK teams）；迁移 PG+SQLite；旧数据 `team_id=NULL` 视为工作区级（本包列表**可隐藏**或标「未分队」，产品裁定：默认列表**仅当前队**，NULL 进「未分队」滤镜可选） | 重做整个 Issue 详情 UX |
| 创建 Issue 时写入当前 `team_id`（FE/API 上下文） | 跨队批量搬家 UI 大工程（可后置 CLI） |
| `GET issues` / board / list 支持 `team_id` 必选或默认当前队过滤 | T7 模板创建 |
| 单测：双队各建单，队 B list 不见队 A | 改 pins；改 handoff |

**默认裁定（总指挥）**：`team_id` 必填于新创建；旧 NULL 仅 Admin「未分队」可见，普通 list 默认 `team_id=<current>`。

## 验收

| ID | 标准 | 主跑 |
|----|------|------|
| **I1** | 队 A 建单 → 队 A list 见、队 B list **不见** | Go + QA |
| **I2** | 无/错 `team_id` 创建 → 4xx 人话（新单） | Go |
| **I3** | FE 列表/看板随当前队换池无残留 | React |
| **I4** | 证据 `docs/qa-shots/issue-pool-team-YYYYMMDD.txt` | QA |

## 主辅

Go 主 · React 辅 · 种子文档 · QA · 产品勾

## 产品勾选

- [ ] I1–I2 API
- [ ] I3 FE
- [ ] I4 QA
