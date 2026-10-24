# T6-I / Issue 池 苛刻复测清单（QA）

> 权威优先：`docs/NEXT_WAVE_DEBT_T6I.md`（I1–I4）· 对照 `docs/NEXT_WAVE_DEBT_ISSUE_POOL.md`  
> Go 开工锚 `4177db4` · **等交件/证据后再跑，勿盲关单**

## 口径对齐
- T6I：**I3**=证据文件，**I4**=FE 换池（可后于 API）
- ISSUE_POOL 旧表：**I4**=QA 证据 —— 本清单以 **T6I** 为准；QA 主跑 **I1–I3 API 苛刻复测**，FE I4 交件后另验

## 环境
- API `:8080` 含 `issues.team_id` 迁移；FE `:5174`（I4）
- WS `ba1481…`；双队新鲜 CreateTeam + 成员/Key
- 旧单 NULL 策略：证据写明（默认 list 仅当前队 / 未分队 Admin）

## API（I1–I2 → 证据 I3）
1. 队 A Create Issue（带 team_id=A）→ A list/`?team_id=A` 见；B list **不见**
2. 队 B 同理互不串
3. 错队 GET/PATCH → **404**
4. 新单无/错 team_id → 4xx 人话（按实现）
5. 单测绿；写 `docs/qa-shots/T6I-issue-pool-YYYYMMDD.txt`（或 `issue-pool-team-*`）

## FE（I4，并行不挡 API 半单）
1. 当前 team 上下文下列表/看板换池
2. 切队无跨队残留单

## 红线
- 双队 list 串单 → FAIL
- 错队可读写 → FAIL
- 不做 T7 / 不重开 pins

## 关报
短报产品+总指挥：Overall · I1–I4 · 证据路径 · HEAD
